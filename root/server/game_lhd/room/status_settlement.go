package room

import (
	"github.com/golang/protobuf/proto"
	"root/common"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/game_lhd/send_tools"
	"strconv"
)

type (
	settlement struct {
		*Room
		s               ERoomStatus
		start_timestamp int64
		end_timestamp   int64
		enterMsg        *protomsg.StatusMsgLHD
	}
)

func (self *settlement) Enter(now int64) {
	duration := self.status_duration[self.s]
	self.start_timestamp = utils.MilliSecondTimeSince1970()
	self.end_timestamp = self.start_timestamp + duration
	self.log(colorized.Gray("settlement enter duration:%v"), duration)

	var (
		win protomsg.LHDAREA
	)
	cards := self.GameCards[:2]
	if cards[0].Number > cards[1].Number {
		win = protomsg.LHDAREA_LHD_AREA_DRAGON
	} else if cards[0].Number < cards[1].Number {
		win = protomsg.LHDAREA_LHD_AREA_TIGER
	} else {
		win = protomsg.LHDAREA_LHD_AREA_PEACE
	}

	allprofit := map[int32]int64{}
	for accid, bets := range self.betPlayers {
		loss_val := int64(0) // 输的钱
		// 先统计三方押注总额
		loss_val += bets[1] + bets[2] + bets[3]
		// 扣掉退还的钱
		loss_val -= bets[int32(win)]
		if win == protomsg.LHDAREA_LHD_AREA_PEACE {
			loss_val -= (bets[1] + bets[2]) * self.peaceBack_conf / 10000
		}

		principal_val := int64(0) // 本金
		principal_val += bets[int32(win)]

		// 计算利润
		winArea_profit := bets[int32(win)] * self.odds_conf[win] * (10000 - self.pump_conf[win]) / 10000

		acc := self.accounts[accid]
		if acc == nil {
			continue
		}
		val := winArea_profit + principal_val
		acc.AddMoney(val, common.EOperateType_LHD_WIN)
		if acc.Robot == 0 && acc.OSType == 4 {
			asyn_addMoney(self.addr_url, acc.UnDevice, val, int32(self.roomId), "龙虎斗盈利", nil, nil) //中奖
		}
		allprofit[int32(accid)] = winArea_profit
		if acc.Robot == 0 {
			self.profit -= winArea_profit - loss_val
		}
		self.log("玩家:%v 押注:%v 输掉的钱:%v 归还本金:%v 赢方区域盈利:%v 总输赢(不算本金):%v ", accid, bets, loss_val, principal_val, winArea_profit, winArea_profit-loss_val)
	}
	self.history = append(self.history, &protomsg.ENTER_GAME_LHD_RES_Winner{
		WinArea: win,
	})

	// 组装消息
	settle, err := proto.Marshal(&protomsg.Status_Settle_LHD{
		WinArea:     win,
		DragonCards: self.GameCards[0],
		TigerCards:  self.GameCards[1],
		Players:     allprofit,
	})
	if err != nil {
		log.Panicf("错误:%v ", err.Error())
	}

	betval := self.areaBetVal(true)
	self.enterMsg = &protomsg.StatusMsgLHD{
		Status:           protomsg.LHDGAMESTATUS(self.s),
		Status_StartTime: uint64(self.start_timestamp),
		Status_EndTime:   uint64(self.end_timestamp),
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
		send_tools.Send2Account(protomsg.LHDMSG_SC_SWITCH_GAME_STATUS_BROADCAST_LHD.UInt16(), &protomsg.SWITCH_GAME_STATUS_BROADCAST_LHD{self.enterMsg}, acc.SessionId)
	}

	send_tools.Send2Hall(inner.SERVERMSG_GH_ROOM_PROFIT_SAVE.UInt16(), &inner.ROOM_PROFIT_SAVE{
		RoomID: self.roomId,
		Value:  strconv.Itoa(int(self.profit)),
	})

	self.robotQuit()
	self.log("win:%v 龙牌:%v  虎牌:%v 房间盈利:%v", win, self.GameCards[0], self.GameCards[1], self.profit)
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

func (self *settlement) enterData(accountId uint32) *protomsg.StatusMsgLHD {
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
