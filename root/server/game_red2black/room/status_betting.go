package room

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"root/common"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/game_red2black/account"
	"root/server/game_red2black/algorithm"
	"root/server/game_red2black/send_tools"
)

type (
	betting struct {
		*Room
		s                        ERoomStatus
		start_timestamp          int64
		end_timestamp            int64
		enterMsg                 *protomsg.StatusMsg
		interval_broadcast_timer int64 // 间隔广播下注缓存

		bets_cache      []*protomsg.BET_RED2BLACK_RES_BetPlayer // 下注缓存
		cd              map[uint32]int64                        // 玩家最后一次下注时间戳
		forbidBetplayer map[uint32]bool                         // 禁止下注的玩家
	}
)

func (self *betting) Enter(now int64) {
	duration := self.status_duration[self.s]
	self.start_timestamp = utils.MilliSecondTimeSince1970()
	self.end_timestamp = self.start_timestamp + duration
	log.Debugf(colorized.Yellow("betting enter duration:%v"), duration)

	self.cd = make(map[uint32]int64)
	self.forbidBetplayer = make(map[uint32]bool)
	// 随机获得6张牌
	self.GameCards = algorithm.GetRandom_Card(self.RoomCards, 6)
	log.Infof("开始下注显示:%v 张 本局牌:%+v ", self.showNum, self.GameCards)

	self.bets_cache = make([]*protomsg.BET_RED2BLACK_RES_BetPlayer, 0)
	// 广播房间玩家，切换状态
	bet, err := proto.Marshal(&protomsg.Status_Bet{
		//todo .....................................................
	})
	if err != nil {
		log.Panicf("错误:%v ", err.Error())
	}

	for accid, acc := range self.accounts {
		if acc.GetMoney() < uint64(self.betlimit) {
			self.forbidBetplayer[acc.AccountId] = true
		}
		if acc.SessionId == 0 {
			continue
		}
		betval, betval_own := self.areaBetVal(true, accid)
		self.enterMsg = &protomsg.StatusMsg{
			Status:           protomsg.RED2BLACKGAMESTATUS(self.s),
			Status_StartTime: uint64(self.start_timestamp),
			Status_EndTime:   uint64(self.end_timestamp),
			RedCards:         self.GameCards[0:self.showNum],
			BlackCards:       self.GameCards[3 : 3+self.showNum],
			AreaBetVal:       betval,
			AreaBetVal_Own:   betval_own,
			Status_Data:      bet,
		}
		send_tools.Send2Account(protomsg.RED2BLACKMSG_SC_SWITCH_GAME_STATUS_BROADCAST.UInt16(), &protomsg.SWITCH_GAME_STATUS_BROADCAST{self.enterMsg}, acc.SessionId)
	}

	self.interval_broadcast_timer = self.owner.AddTimer(500, -1, self.updateBetPlayers)
}

func (self *betting) updateBetPlayers(now int64) {
	if len(self.bets_cache) != 0 {
		betval, _ := self.areaBetVal(true, 0)
		self.SendBroadcast(protomsg.RED2BLACKMSG_SC_BET_RED2BLACK_RES.UInt16(), &protomsg.BET_RED2BLACK_RES{
			Players:    self.bets_cache,
			AreaBetVal: betval,
		})
		self.bets_cache = self.bets_cache[:0]
	}
}

func (self *betting) Tick(now int64) {
	if now >= self.end_timestamp {
		self.owner.CancelTimer(self.interval_broadcast_timer)
		self.updateBetPlayers(now) // 把剩余的发给客户端
		self.switchStatus(now, ERoomStatus_STOP_BETTING)
		return
	}
}

func (self *betting) leave(accid uint32) bool {
	_, exist := self.betPlayers[accid]
	// 如果 玩家有押注，不能退出游戏
	if exist {
		return false
	}
	return true
}

func (self *betting) enterData(accountId uint32) *protomsg.StatusMsg {
	return self.enterMsg
}

func (self *betting) Leave(now int64) {
	log.Debugf(colorized.Yellow("betting leave\n"))
	log.Debugf(colorized.Blue(""))
}

func (self *betting) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.RED2BLACKMSG_CS_BET_RED2BLACK_REQ.UInt16():
		self.RED2BLACKMSG_CS_BET_RED2BLACK_REQ(actor, msg, session)
	default:
		log.Warnf("betting 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}

	return true
}

func (self *betting) RED2BLACKMSG_CS_BET_RED2BLACK_REQ(actor int32, msg []byte, session int64) {
	betdata := packet.PBUnmarshal(msg, &protomsg.BET_RED2BLACK_REQ{}).(*protomsg.BET_RED2BLACK_REQ)
	var acc *account.Account
	now := utils.MilliSecondTimeSince1970()
	if session == 0 {
		acc = account.AccountMgr.GetAccountByIDAssert(betdata.GetAccountID())
	} else {
		acc = account.AccountMgr.GetAccountBySessionIDAssert(session)
	}

	if acc.GetMoney() < betdata.GetBet() {
		log.Warnf("acc:%v room:%v 钱不够下注 身上钱:%v 请求下注:%v ", acc.AccountId, self.roomId, acc.GetMoney(), betdata.GetBet())
		return
	}
	if self.forbidBetplayer[acc.AccountId] {
		log.Warnf("acc:%v room:%v 钱不够下注 身上钱:%v 低于bet_limit 请求下注失败:%v ", acc.AccountId, self.roomId, acc.GetMoney(), betdata.GetBet())
		return
	}
	if last := self.cd[acc.GetAccountId()]; now-last < self.interval_conf {
		return
	}

	check := false
	for _, betVal := range self.bets_conf {
		if uint64(betVal) == betdata.Bet {
			check = true
			break
		}
	}
	if !check {
		log.Warnf("acc:%v room:%v 钱不够下注 请求下注不在下注区域内:%v ", acc.AccountId, self.roomId, betdata.GetBet())
		return
	}

	back := func(backunique string, backmoney int64) {
		if acc.GetMoney()-betdata.GetBet() != uint64(backmoney) {
			log.Warnf("数据错误  ->>>>>> userID:%v money:%v Bet:%v gold:%v", acc.GetUnDevice(), acc.GetMoney(), betdata.GetBet(), backmoney)
			acc.AddMoney(backmoney-int64(acc.GetMoney()), common.EOperateType_INIT)
		} else {
			acc.AddMoney(int64(-(betdata.GetBet())), common.EOperateType_RED2BLACK_BET)
		}

		self.cd[acc.GetAccountId()] = now
		self.betPlayers[acc.AccountId][betdata.Area] += int64(betdata.Bet)
		self.bets_cache = append(self.bets_cache, &protomsg.BET_RED2BLACK_RES_BetPlayer{
			AccountID: acc.GetAccountId(),
			Area:      betdata.GetArea(),
			Bet:       betdata.Bet,
		})
	}

	if acc.Robot == 0 {
		back(acc.UnDevice, int64(acc.GetMoney()-betdata.GetBet()))
	} else {
		// 错误返回
		errback := func() {
			log.Panicf("http请求报错 玩家:%v roomID:%v  下注:%v 失败", acc.GetAccountId(), self.roomId, betdata.GetBet())
		}
		asyn_addMoney(acc.UnDevice, -int64(betdata.GetBet()), int32(self.roomId), fmt.Sprintf("红黑大战请求下注:%v", betdata.GetBet()), back, errback)
	}

}
