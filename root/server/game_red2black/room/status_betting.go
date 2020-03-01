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

	betval, _ := self.areaBetVal(true, 0)
	for _, acc := range self.accounts {
		acc.Betcount = 0
		if acc.GetMoney() < uint64(self.betlimit) {
			self.forbidBetplayer[acc.AccountId] = true
		}
	}

	self.enterMsg = &protomsg.StatusMsg{
		Status:           protomsg.RED2BLACKGAMESTATUS(self.s),
		Status_StartTime: uint64(self.start_timestamp),
		Status_EndTime:   uint64(self.end_timestamp),
		RedCards:         self.GameCards[0:self.showNum],
		BlackCards:       self.GameCards[3 : 3+self.showNum],
		AreaBetVal:       betval,
		AreaBetVal_Own:   map[int32]int64{},
		Status_Data:      bet,
	}
	self.SendBroadcast(protomsg.RED2BLACKMSG_SC_SWITCH_GAME_STATUS_BROADCAST.UInt16(), &protomsg.SWITCH_GAME_STATUS_BROADCAST{self.enterMsg})

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
	_, betval_own := self.areaBetVal(true, accountId)
	self.enterMsg.AreaBetVal_Own = betval_own
	return self.enterMsg
}

func (self *betting) Leave(now int64) {
	log.Debugf(colorized.Yellow("betting leave\n"))
	log.Debugf(colorized.Blue(""))
}

func (self *betting) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.RED2BLACKMSG_CS_BET_RED2BLACK_REQ.UInt16(): // 请求下注
		self.RED2BLACKMSG_CS_BET_RED2BLACK_REQ(actor, pack.ReadBytes(), session)
	case protomsg.RED2BLACKMSG_CS_CLEAN_BET_RED2BLACK_REQ.UInt16(): // 请求清空下注
		self.RED2BLACKMSG_CS_CLEAN_BET_RED2BLACK_REQ(actor, pack.ReadBytes(), session)
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
	if now-acc.CLeanTime < 3000 {
		log.Warnf("acc:%v 还未收到三方的清除消息，不能押注:%v ", acc.GetAccountId(), now-acc.CLeanTime)
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

	self.cd[acc.GetAccountId()] = now
	back := func(backunique string, backmoney int64) {
		if acc.GetMoney()-betdata.GetBet() != uint64(backmoney) {
			log.Warnf("数据错误  ->>>>>> userID:%v money:%v Bet:%v gold:%v", acc.GetUnDevice(), acc.GetMoney(), betdata.GetBet(), backmoney)
			acc.AddMoney(backmoney-int64(acc.GetMoney()), common.EOperateType_INIT)
		} else {
			acc.AddMoney(int64(-(betdata.GetBet())), common.EOperateType_RED2BLACK_BET)
		}
		playerBets, e := self.betPlayers[acc.AccountId]
		if !e {
			self.betPlayers[acc.AccountId] = make(map[protomsg.RED2BLACKAREA]int64)
			playerBets, _ = self.betPlayers[acc.AccountId]
		}
		playerBets[betdata.Area] += int64(betdata.Bet)
		self.bets_cache = append(self.bets_cache, &protomsg.BET_RED2BLACK_RES_BetPlayer{
			AccountID: acc.GetAccountId(),
			Area:      betdata.GetArea(),
			Bet:       betdata.Bet,
		})
		log.Infof("acc:%v下注成功,下注区域:%v 金额:%v", acc.GetAccountId(), betdata.Area, betdata.Bet)
		acc.Betcount--
	}

	acc.Betcount++
	if acc.Robot == 0 || acc.GetOSType() != 4 {
		back(acc.UnDevice, int64(acc.GetMoney()-betdata.GetBet()))
	} else {
		// 错误返回
		errback := func() {
			log.Panicf("http请求报错 玩家:%v roomID:%v  下注:%v 失败", acc.GetAccountId(), self.roomId, betdata.GetBet())
		}
		log.Infof("acc:%v unique:%v 请求下注,下注区域:%v 金额:%v", acc.GetAccountId(), acc.UnDevice, betdata.Area, betdata.Bet)
		asyn_addMoney(acc.UnDevice, -int64(betdata.GetBet()), int32(self.roomId), fmt.Sprintf("红黑大战请求下注:%v", betdata.GetBet()), back, errback)
	}
}
func (self *betting) RED2BLACKMSG_CS_CLEAN_BET_RED2BLACK_REQ(actor int32, msg []byte, session int64) {
	betdata := packet.PBUnmarshal(msg, &protomsg.CLEAN_BET_RED2BLACK_REQ{}).(*protomsg.CLEAN_BET_RED2BLACK_REQ)
	var acc *account.Account
	now := utils.MilliSecondTimeSince1970()
	if session == 0 {
		acc = account.AccountMgr.GetAccountByIDAssert(betdata.GetAccountID())
	} else {
		acc = account.AccountMgr.GetAccountBySessionIDAssert(session)
	}
	if acc.Betcount > 0 {
		log.Warnf("玩家:%v 下注计数为:%v 需要等到所有下注都成功，才能清除", acc.GetAccountId(), acc.Betcount)
		return
	}

	// 先统计一下玩家的总下注
	totalBets := self.betPlayers[acc.GetAccountId()]
	totalVal := uint64(0)
	for _, v := range totalBets {
		totalVal += uint64(v)
	}

	if totalVal == 0 {
		log.Warnf("玩家:%v 下注 为0 区域下注为:%v 不需要清除", acc.GetAccountId(), totalBets)
		return
	}

	acc.CLeanTime = now
	self.cd[acc.GetAccountId()] = now
	back := func(backunique string, backmoney int64) {
		if acc.GetMoney()+totalVal != uint64(backmoney) {
			log.Warnf("数据错误  ->>>>>> userID:%v money:%v Bet:%v gold:%v", acc.GetUnDevice(), acc.GetMoney(), totalVal, backmoney)
			acc.AddMoney(backmoney-int64(acc.GetMoney()), common.EOperateType_INIT)
		} else {
			acc.AddMoney(int64(totalVal), common.EOperateType_RED2BLACK_BET_CLEAN)
		}

		delete(self.betPlayers, acc.AccountId)
		acc.CLeanTime = 0
		// 通知玩家更新下注区域
		total, _ := self.areaBetVal(true, 0)
		msg := &protomsg.CLEAN_BET_RED2BLACK_RES{
			AccountID:  acc.AccountId,
			AreaBetVal: total,
		}
		self.SendBroadcast(protomsg.RED2BLACKMSG_SC_CLEAN_BET_RED2BLACK_RES.UInt16(), msg)
	}

	if acc.Robot == 0 {
		back(acc.UnDevice, int64(acc.GetMoney()+totalVal))
	} else {
		// 错误返回
		errback := func() {
			log.Panicf("http请求报错 玩家:%v roomID:%v  下注:%v 失败", acc.GetAccountId(), self.roomId, totalVal)
		}
		asyn_addMoney(acc.UnDevice, int64(totalVal), int32(self.roomId), fmt.Sprintf("红黑大战请求下注:%v ", totalVal), back, errback)
	}
}
