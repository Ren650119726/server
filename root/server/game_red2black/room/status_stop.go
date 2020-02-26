package room

import (
	"root/common"
	"root/common/config"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/game_red2black/algorithm"
)

type (
	stop struct {
		*Room
		s               ERoomStatus
		start_timestamp int64
		end_timestamp   int64
		enterMsg        *protomsg.StatusMsg
	}
)

func (self *stop) Enter(now int64) {
	duration := self.status_duration[self.s]
	self.start_timestamp = utils.MilliSecondTimeSince1970()
	self.end_timestamp = self.start_timestamp + duration
	log.Debugf(colorized.Yellow("stop enter duration:%v"), duration)

	cardred := self.GameCards[:3]
	cardblack := self.GameCards[3:6]

	result, tred, tblack := algorithm.Compare(cardred, cardblack)

	s := &algorithm.Card_sorte{}
	s.A = true

	//s.S = cardwin
	//sort.Sort(s)
	//s.S = cardlose
	//sort.Sort(s)

	log.Debugf(colorized.Green("stop enter duration:%v"), duration)
	log.Debugf("三方押注:%v %v 红方牌:%v %v  黑方牌:%v %v", bets, bets_robot, self.red_cards, twin.String(), self.black_cards, tlose.String())
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

func (self *stop) enterData(accountId uint32) *protomsg.StatusMsg {
	return nil
}

//win 赢的一方, 牌型赔率
func (self *stop) prep_settlement(win protomsg.RED2BLACKAREA, special_odds int32) int64 {
	var playerWinVal int64
	// 先计算出玩家总押注
	betVal := self.areaBetVal(false)
	// 赢的一方，统计玩家总盈利
	playerWinVal += betVal[int32(win)] * self.odds_conf[win] * (10000 - self.pump_conf[win]) / 10000
	if special_odds == 0 { // 如果赢方 牌型赔率为0，特殊区域下注为系统盈利
		playerWinVal += betVal[3]
	} else {
		sysWinVal -= betVal[3]
	}

}

func (self *stop) Leave(now int64) {
	log.Debugf(colorized.Green("stop leave\n"))
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
