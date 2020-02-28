package room

import (
	"github.com/golang/protobuf/proto"
	"root/common/config"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/game_red2black/algorithm"
	"root/server/game_red2black/send_tools"
	"sort"
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

	kill := utils.Probability10000(config.Get_configInt("red2black_room", int(self.roomId), "Kill_Point"))
	if self.profit < 0 {
		kill = true
	}
	if kill {
		killbetVal, _ := self.areaBetVal(false, 0)
		self.kill(killbetVal)
	}

	// 组装消息
	stop, err := proto.Marshal(&protomsg.Status_Stop{
		//todo .....................................................
	})
	if err != nil {
		log.Panicf("错误:%v ", err.Error())
	}

	var logbetVal map[int32]int64
	for accid, acc := range self.accounts {
		if acc.SessionId == 0 {
			continue
		}
		betval, betval_own := self.areaBetVal(true, accid)
		logbetVal = betval
		self.enterMsg = &protomsg.StatusMsg{
			Status:           protomsg.RED2BLACKGAMESTATUS(self.s),
			Status_StartTime: uint64(self.start_timestamp),
			Status_EndTime:   uint64(self.end_timestamp),
			RedCards:         self.GameCards[0:3],
			BlackCards:       self.GameCards[3:6],
			AreaBetVal:       betval,
			AreaBetVal_Own:   betval_own,
			Status_Data:      stop,
		}
		send_tools.Send2Account(protomsg.RED2BLACKMSG_SC_SWITCH_GAME_STATUS_BROADCAST.UInt16(), &protomsg.SWITCH_GAME_STATUS_BROADCAST{self.enterMsg}, acc.SessionId)
	}
	log.Debugf("结果牌:%v 三方押注:%v", self.GameCards, logbetVal)
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
	return self.enterMsg
}

//三方押注
func (self *stop) kill(betVal map[int32]int64) {
	if len(betVal) == 0 {
		// 没有一个真实玩家下注，不需要处理吃大配小逻辑
		return
	}

	availableCard := self.RoomCards[:46] // 剩下的46张牌可以用来进行配牌
	type temp struct {
		win       int32
		odds      int64
		syswinVal int64
	}

	duizi_odds := int64(config.Get_configInt("red2black_card", int(protomsg.RED2BLACKCARDTYPE_RED2BLACK_CARDTYPE_2), "Card_Odds"))
	arr := []*temp{
		{win: 1, odds: 0, syswinVal: 0},
		{win: 1, odds: duizi_odds, syswinVal: 0},
		{win: 2, odds: 0, syswinVal: 0},
		{win: 2, odds: duizi_odds, syswinVal: 0},
	}

	for _, v := range arr {
		v.syswinVal = self.prep_settlement(betVal, protomsg.RED2BLACKAREA(v.win), v.odds)
	}

	sort.Slice(arr, func(i, j int) bool {
		return arr[i].syswinVal > arr[j].syswinVal
	})
	log.Infof("三方押注触发 吃大赔小 对子赔率:%v 演算结果:%+v ", duizi_odds, arr)
	for i := 0; i < len(arr)-1; i++ {
		if arr[i].syswinVal < 0 {
			if i == 0 {
				log.Errorf("计算不出吃大赔小的情况，请检查 初始随机牌组:%v 押注：%v", self.GameCards, betVal)
				break
			}
			// 去掉所有系统亏钱的情况
			arr = arr[:i]
			break
		}
	}

	// 做一个随机，最终取用下标0的值
	if l := len(arr); l >= 2 {
		i := utils.Randx_y(1, l)
		arr[0], arr[i] = arr[i], arr[0]
	}
	log.Infof("决定取用的杀分结果:%v", arr[0])

	// 如果当前结果满足吃大赔小，不需要配牌
	var (
		win protomsg.RED2BLACKAREA
		t   protomsg.RED2BLACKCARDTYPE
	)
	result, tred, tblack := algorithm.Compare(self.GameCards[:3], self.GameCards[3:6])
	if result {
		win = protomsg.RED2BLACKAREA_RED2BLACK_AREA_RED
		t = tred
	} else {
		win = protomsg.RED2BLACKAREA_RED2BLACK_AREA_BLACK
		t = tblack
	}
	odd := int64(config.Get_configInt("red2black_card", int(t), "Card_Odds"))

	if int32(win) == arr[0].win && odd <= arr[0].odds {
		log.Infof("结果和杀分结果一致，不需要配牌 win:%v t:%v", win.String(), t.String())
	} else {
		// 先给输家配牌
		num := (3 - self.showNum) * 2 // 1边需要随3-show张牌
		for i := 0; i < 5000; i++ {
			cards := algorithm.GetRandom_Card(availableCard, num)
			red := append([]*protomsg.Card{}, self.GameCards[:self.showNum]...)
			red = append(red, cards[:num/2]...)
			black := append([]*protomsg.Card{}, self.GameCards[self.showNum:3+self.showNum]...)
			black = append(black, cards[num/2:num]...)
			if len(red) != 3 || len(black) != 3 {
				log.Errorf("逻辑错误 red:%v black:%v show:%v len(availableCard):%v", red, black, self.showNum, len(availableCard))
				break
			}
			result, tred, tblack = algorithm.Compare(red, black)
			if result {
				win = protomsg.RED2BLACKAREA_RED2BLACK_AREA_RED
				t = tred
			} else {
				win = protomsg.RED2BLACKAREA_RED2BLACK_AREA_BLACK
				t = tblack
			}
			odd = int64(config.Get_configInt("red2black_card", int(t), "Card_Odds"))
			if int32(win) == arr[0].win && odd <= arr[0].odds {
				log.Infof("随机配牌结果成功，result:%v red:%v %v black:%v %v ", result, red, tred.String(), black, tblack.String())
				self.GameCards = self.GameCards[:0]
				self.GameCards = append(red, black...)
				break
			} else {
				log.Infof("随机配牌结果失败，result:%v red:%v %v black:%v %v ", result, red, tred.String(), black, tblack.String())
			}
		}
	}
}

// totalBetVal 总押注 win 赢的一方, 牌型赔率
func (self *stop) prep_settlement(betVal map[int32]int64, win protomsg.RED2BLACKAREA, special_odds int64) int64 {
	sysWinVal := betVal[3-int32(win)]
	if special_odds == 0 {
		sysWinVal += betVal[3]
	}

	// 获胜区域，玩家总盈利
	sysWinVal -= betVal[int32(win)] * self.odds_conf[win] * (10000 - self.pump_conf[win]) / 10000
	sysWinVal -= betVal[3] * special_odds * (10000 - self.pump_conf[3]) / 10000
	return sysWinVal
}

func (self *stop) Leave(now int64) {
	log.Debugf(colorized.Green("stop leave\n"))
	log.Debugf("")
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
