package room

import (
	"github.com/golang/protobuf/proto"
	"root/common/config"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/game_lhd/send_tools"
)

type (
	stop struct {
		*Room
		s               ERoomStatus
		start_timestamp int64
		end_timestamp   int64
		enterMsg        *protomsg.StatusMsgLHD
	}
)

func (self *stop) Enter(now int64) {
	duration := self.status_duration[self.s]
	self.start_timestamp = utils.MilliSecondTimeSince1970()
	self.end_timestamp = self.start_timestamp + duration
	self.log(colorized.Yellow("stop enter duration:%v"), duration)

	kill := utils.Probability10000(config.Get_configInt("lhd_room", int(self.roomId), "Kill_Point"))
	if self.profit < -int64(config.Get_configInt("lhd_room", int(self.roomId), "Lose_Gold")) {
		kill = true
	}
	if kill {
		killbetVal := self.areaBetVal(false)
		self.kill(killbetVal)
	}

	// 组装消息
	stop, err := proto.Marshal(&protomsg.Status_Stop{
		//todo .....................................................
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
		Status_Data:      stop,
	}

	for accid, acc := range self.accounts {
		if acc.SessionId == 0 {
			continue
		}
		betval_own := self.playerAreaBetVal(accid)
		self.enterMsg.AreaBetVal_Own = betval_own
		send_tools.Send2Account(protomsg.LHDMSG_SC_SWITCH_GAME_STATUS_BROADCAST_LHD.UInt16(), &protomsg.SWITCH_GAME_STATUS_BROADCAST_LHD{self.enterMsg}, acc.SessionId)
	}
	self.log("结果牌:%v 三方押注:%v", self.GameCards[:2], betval)
}

func (self *stop) Tick(now int64) {
	if now >= self.end_timestamp {
		self.switchStatus(now, ERoomStatus_SETTLEMENT)
		return
	}
}

func (self *stop) leave(accid uint32) bool {
	_, exist := self.betPlayers[accid]
	// 如果 玩家有押注，不能退出游戏
	if exist {
		return false
	}
	return true
}

func (self *stop) enterData(accountId uint32) *protomsg.StatusMsgLHD {
	betval_own := self.playerAreaBetVal(accountId)
	self.enterMsg.AreaBetVal_Own = betval_own
	return self.enterMsg
}

//三方押注
func (self *stop) kill(betVal map[int32]int64) {
	if len(betVal) == 0 {
		// 没有一个真实玩家下注，不需要处理吃大配小
		return
	}

	var (
		win protomsg.LHDAREA
	)
	cards := self.GameCards[:2]
	if cards[0].Number > cards[1].Number {
		win = protomsg.LHDAREA_LHD_AREA_DRAGON
	} else if cards[0].Number == cards[1].Number {
		win = protomsg.LHDAREA_LHD_AREA_TIGER
	} else {
		win = protomsg.LHDAREA_LHD_AREA_PEACE
	}

	self.log("杀分处理 牌组:%v  win：%v ", cards, win)
	val := self.prep_settlement(betVal, win)
	if win == protomsg.LHDAREA_LHD_AREA_PEACE {
		self.log("平 不处理杀分")
		return
	}
	self.log("系统盈利:%v ", val)

	if val < 0 {
		self.GameCards[0], self.GameCards[1] = self.GameCards[1], self.GameCards[0]
		cards = self.GameCards[:2]
		val := self.prep_settlement(betVal, win)
		self.log("换牌后 牌组:%v 系统盈利:%v ", cards, val)
	}
}

// betVal  真实玩家总押注 win 赢的一方, 牌型赔率
func (self *stop) prep_settlement(betVal map[int32]int64, win protomsg.LHDAREA) int64 {
	sysWinVal := int64(0)
	for a, v := range betVal {
		if a != int32(win) {
			sysWinVal += v
		}
	}

	sysWinVal -= betVal[int32(win)] * self.odds_conf[win] * (10000 - self.pump_conf[win]) / 10000
	return sysWinVal
}

func (self *stop) Leave(now int64) {
	self.log(colorized.Green("stop leave\n"))
	self.log("")
}
func (self *stop) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	default:
		log.Warnf("stop 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}
	return true
}
