package room

import (
	"github.com/golang/protobuf/proto"
	"root/common"
	"root/common/config"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/game_red2black/algorithm"
	"root/server/game_red2black/send_tools"
	"sort"
	"strconv"
)

type (
	settlement struct {
		*Room
		s               ERoomStatus
		start_timestamp int64
		end_timestamp   int64
		enterMsg        *protomsg.StatusMsg
	}
)

func (self *settlement) Enter(now int64) {
	duration := self.status_duration[self.s]
	self.start_timestamp = utils.MilliSecondTimeSince1970()
	self.end_timestamp = self.start_timestamp + duration
	self.log(colorized.Gray("settlement enter duration:%v"), duration)

	var (
		win   protomsg.RED2BLACKAREA
		t     protomsg.RED2BLACKCARDTYPE
		losst protomsg.RED2BLACKCARDTYPE
	)
	redcard := append([]*protomsg.Card{}, self.GameCards[:3]...)
	blackcard := append([]*protomsg.Card{}, self.GameCards[3:6]...)
	result, tred, tblack := algorithm.Compare(redcard, blackcard)
	var wincard []*protomsg.Card
	if result {
		win = protomsg.RED2BLACKAREA_RED2BLACK_AREA_RED
		t = tred
		wincard = redcard
		losst = tblack
	} else {
		win = protomsg.RED2BLACKAREA_RED2BLACK_AREA_BLACK
		t = tblack
		wincard = blackcard
		losst = tred
	}
	specialArea_odd := int64(config.Get_configInt("red2black_card", int(t), "Card_Odds"))
	if t == protomsg.RED2BLACKCARDTYPE_RED2BLACK_CARDTYPE_2 {
		sort.Slice(wincard, func(i, j int) bool {
			return wincard[i].Number > wincard[j].Number
		})
		self.log("特殊对子判断:%v ", wincard)
		// 特殊对子，没有赔率
		if 2 <= wincard[1].Number && wincard[1].Number <= 8 {
			specialArea_odd = 0
		}
	}

	allprofit := map[int32]int64{}
	for accid, bets := range self.betPlayers {
		loss_val := int64(0) // 输的钱
		loss_val += bets[int32(3-win)]

		principal_val := int64(0) // 本金
		principal_val += bets[int32(win)]
		if bets[3] != 0 && specialArea_odd != 0 {
			principal_val += bets[3]
		} else {
			loss_val += bets[3]
		}

		// 计算利润
		winArea_profit := bets[int32(win)] * self.odds_conf[win] * (10000 - self.pump_conf[win]) / 10000
		specialArea_profit := bets[3] * specialArea_odd * (10000 - self.pump_conf[3]) / 10000

		acc := self.accounts[accid]
		if acc == nil {
			continue
		}
		val := winArea_profit + specialArea_profit + principal_val
		acc.AddMoney(val, common.EOperateType_RED2BLACK_WIN)
		if acc.Robot == 0 && acc.OSType == 4 {
			asyn_addMoney(self.addr_url, acc.UnDevice, val, int32(self.roomId), "红黑大战盈利", nil, nil) //中奖
		}
		allprofit[int32(accid)] = winArea_profit + specialArea_profit
		if acc.Robot == 0 {
			self.profit -= winArea_profit + specialArea_profit - loss_val
		}
		self.log("玩家:%v 押注:%v 输掉的钱:%v 归还本金:%v 赢方区域盈利:%v 特殊区域盈利:%v 总输赢(不算本金):%v ", accid, bets, loss_val, principal_val, winArea_profit, specialArea_profit, winArea_profit+specialArea_profit-loss_val)
	}
	self.history = append(self.history, &protomsg.ENTER_GAME_RED2BLACK_RES_Winner{
		WinArea:     win,
		WinCardType: t,
	})
	if len(self.history) > 70 {
		self.history = self.history[1:]
	}

	// 组装消息
	settle, err := proto.Marshal(&protomsg.Status_Settle{
		WinArea:      win,
		WinCardType:  t,
		LossCardType: losst,
		WinOdds:      uint64(specialArea_odd),
		Players:      allprofit,
	})
	if err != nil {
		log.Panicf("错误:%v ", err.Error())
	}

	betval := self.areaBetVal(true)
	self.enterMsg = &protomsg.StatusMsg{
		Status:           protomsg.RED2BLACKGAMESTATUS(self.s),
		Status_StartTime: uint64(self.start_timestamp),
		Status_EndTime:   uint64(self.end_timestamp),
		RedCards:         self.GameCards[0:3],
		BlackCards:       self.GameCards[3:6],
		AreaBetVal:       betval,
		AreaBetVal_Own:   nil,
		Status_Data:      settle,
	}

	for accid, acc := range self.accounts {
		if acc.SessionId == 0 {
			continue
		}
		betval_own := self.playerAreaBetVal(accid)
		self.enterMsg.AreaBetVal_Own = betval_own
		send_tools.Send2Account(protomsg.RED2BLACKMSG_SC_SWITCH_GAME_STATUS_BROADCAST.UInt16(), &protomsg.SWITCH_GAME_STATUS_BROADCAST{self.enterMsg}, acc.SessionId)
	}

	send_tools.Send2Hall(inner.SERVERMSG_GH_ROOM_PROFIT_SAVE.UInt16(), &inner.ROOM_PROFIT_SAVE{
		RoomID: self.roomId,
		Value:  strconv.Itoa(int(self.profit)),
	})
	self.log("win:%v 红方牌:%v  黑方牌:%v 房间盈利:%v", win, tred, tblack, self.profit)
}

func (self *settlement) Tick(now int64) {
	if now >= self.end_timestamp {
		self.switchStatus(now, ERoomStatus_WAITING_TO_START)
		return
	}
}
func (self *settlement) leave(accid uint32) bool {
	return true
}

func (self *settlement) enterData(accountId uint32) *protomsg.StatusMsg {
	betval_own := self.playerAreaBetVal(accountId)
	self.enterMsg.AreaBetVal_Own = betval_own
	return self.enterMsg
}

func (self *settlement) Leave(now int64) {
	self.log(colorized.Gray("settlement leave\n"))
	self.log("")
}

func (self *settlement) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	default:
		log.Warnf("settlement 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}

	return true
}
