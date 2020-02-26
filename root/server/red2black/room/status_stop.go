package room

import (
	"root/common"
	"root/common/config"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/red2black/account"
	"root/server/red2black/algorithm"
	"root/server/red2black/send_tools"
	"sort"
)

type (
	stop struct {
		*Room
		s         ERoomStatus
		timestamp int64
	}
)

func (self *stop) Enter(now int64) {
	duration := self.status_duration[int(self.s)] // 持续时间 秒
	self.timestamp = now + int64(duration)
	bets, bets_robot := self.total_player_bet()

	randCount := 51 // 17个随机的牌组
	cards := algorithm.GetRandom_Card(randCount)
	cardwin := cards[:3]
	cardlose := cards[3:6]
	//cardwin = []algorithm.Card_info{
	//	{common.ECardType_HEITAO.UInt8(), 1},
	//	{common.ECardType_HONGTAO.UInt8(), 1},
	//	{common.ECardType_HEITAO.UInt8(), 1},
	//}
	//cardlose = []algorithm.Card_info{
	//	{common.ECardType_HEITAO.UInt8(), 11},
	//	{common.ECardType_HONGTAO.UInt8(), 11},
	//	{common.ECardType_FANGKUAI.UInt8(), 11},
	//}
	//RoomMgr.Water_line = -249606

	//var badCards1 []algorithm.Card_info
	//var badCards2 []algorithm.Card_info
	//for i := 6; i <= randCount-3; i++ {
	//	c := cards[i : i+3]
	//	if algorithm.JudgeCardType(c) == common.ECardType_SANPAI {
	//		if badCards1 == nil {
	//			badCards1 = c
	//		} else {
	//			badCards2 = c
	//			break
	//		}
	//	}
	//}

	result, twin, tlose := algorithm.Compare(cardwin, cardlose)
	if !result {
		cardwin, cardlose = cardlose, cardwin
		twin, tlose = tlose, twin

	}

	s := &algorithm.Card_sorte{}
	s.A = true

	s.S = cardwin
	sort.Sort(s)
	s.S = cardlose
	sort.Sort(s)

	self.red_cards = cardwin
	tred := twin
	self.black_cards = cardlose
	tblack := tlose

	bet_r := bets[1]
	bet_b := bets[2]
	//bet_s := bets[3]
	// 水位线高于高位让玩家赢, 低于低位让玩家输
	if RoomMgr.Water_line > config.GetPublicConfig_Int64("R2B_CEILING_LINE") {
		if bet_r > bet_b {
			self.red_cards = cardwin
			tred = twin
			self.black_cards = cardlose
			tblack = tlose
		} else if bet_r < bet_b {
			self.red_cards = cardlose
			tred = tlose
			self.black_cards = cardwin
			tblack = twin
		} else if bet_r == bet_b {
			blackwin := utils.Randx_y(0, 2)
			if blackwin == 1 {
				self.red_cards = cardlose
				tred = tlose
				self.black_cards = cardwin
				tblack = twin
			}
		}
		log.Debugf("Water_line:%v > %v", RoomMgr.Water_line, config.GetPublicConfig_Int64("R2B_CEILING_LINE"))
	} else if RoomMgr.Water_line < config.GetPublicConfig_Int64("R2B_FLOOR_LINE") {
		self.red_cards = cardwin
		tred = twin
		self.black_cards = cardlose
		tblack = tlose
		ret1 := self.prep_settlement()

		self.red_cards = cardlose
		tred = tlose
		self.black_cards = cardwin
		tblack = twin
		ret2 := self.prep_settlement()

		if ret1 == ret2 || ret1 == 0 {
			blackwin := utils.Randx_y(0, 2)
			if blackwin == 1 {
				self.red_cards = cardlose
				tred = tlose
				self.black_cards = cardwin
				tblack = twin
			} else {
				self.red_cards = cardwin
				tred = twin
				self.black_cards = cardlose
				tblack = tlose
			}
		} else {
			if ret1 < 0 && ret2 > 0 {
				self.red_cards = cardlose
				tred = tlose
				self.black_cards = cardwin
				tblack = twin
			} else if ret1 > 0 && ret2 < 0 {
				self.red_cards = cardwin
				tred = twin
				self.black_cards = cardlose
				tblack = tlose
			}
		}

		//player_master_rate := self.total_master_player_val() * 100 / self.total_master_val()
		//
		//// 如果庄家中，玩家份额比例超过30% 优先
		//if player_master_rate > 30 {
		//	bets_robot_r := bets_robot[1]
		//	bets_robot_b := bets_robot[2]
		//	if bets_robot_r < bets_robot_b {
		//		self.red_cards = cardlose
		//		tred = tlose
		//		self.black_cards = cardwin
		//		tblack = twin
		//
		//	} else if bets_robot_r > bets_robot_b {
		//		self.red_cards = cardwin
		//		tred = twin
		//		self.black_cards = cardlose
		//		tblack = tlose
		//	} else if bets_robot_r == bets_robot_b {
		//		blackwin := utils.Randx_y(0, 2)
		//		if blackwin == 1 {
		//			self.red_cards = cardlose
		//			tred = tlose
		//			self.black_cards = cardwin
		//			tblack = twin
		//		}
		//	}
		//} else {
		//	if bet_r < bet_b {
		//		self.red_cards = cardwin
		//		tred = twin
		//		self.black_cards = cardlose
		//		tblack = tlose
		//	} else if bet_r > bet_b {
		//		self.red_cards = cardlose
		//		tred = tlose
		//		self.black_cards = cardwin
		//		tblack = twin
		//	} else if bet_r == bet_b {
		//		blackwin := utils.Randx_y(0, 2)
		//		if blackwin == 1 {
		//			self.red_cards = cardlose
		//			tred = tlose
		//			self.black_cards = cardwin
		//			tblack = twin
		//		}
		//	}
		//}

		log.Debugf("Water_line:%v < %v", RoomMgr.Water_line, config.GetPublicConfig_Int64("R2B_FLOOR_LINE"))
	} else {
		blackwin := utils.Randx_y(0, 2)
		if blackwin == 1 {
			self.red_cards = cardlose
			tred = tlose
			self.black_cards = cardwin
			tblack = twin
		}
	}

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_R2B_NEXT_STATE.UInt16())
	send.WriteUInt8(uint8(ERoomStatus_STOP_BETTING))
	send.WriteUInt32(uint32(duration * 1000))

	self.pack = packet.NewPacket(nil)
	self.pack.WriteUInt8(uint8(tred))
	self.pack.WriteUInt16(uint16(len(self.red_cards)))
	for _, v := range self.red_cards {
		self.pack.WriteUInt8(v[0])
		self.pack.WriteUInt8(v[1])
	}
	self.pack.WriteUInt8(uint8(tblack))
	self.pack.WriteUInt16(uint16(len(self.black_cards)))
	for _, v := range self.black_cards {
		self.pack.WriteUInt8(v[0])
		self.pack.WriteUInt8(v[1])
	}
	send.CatBody(self.pack)
	self.SendBroadcast(send.GetData())
	log.Debugf(colorized.Green("stop enter duration:%v"), duration)
	log.Debugf("三方押注:%v %v 红方牌:%v %v  黑方牌:%v %v", bets, bets_robot, self.red_cards, twin.String(), self.black_cards, tlose.String())
}

func (self *stop) Tick(now int64) {
	if now >= self.timestamp {
		self.switchStatus(now, ERoomStatus_SETTLEMENT)
		return
	}
}

func (self *stop) prep_settlement() int64 {
	before := self.total_bet_player_val
	after := before
	reddWin, tred, tblack := algorithm.Compare(self.red_cards, self.black_cards)
	var t common.EJinHuaType
	var c []algorithm.Card_info
	if reddWin {
		t = tred
		c = self.red_cards
	} else {
		t = tblack
		c = self.black_cards
	}

	if t == common.ECardType_DUIZI {
		if c[1][1] < 9 && c[1][1] > 1 {
			t = common.ECardType_SANPAI
		}
	}

	specific_rate := algorithm.Rate_type(t)

	total_share_val := self.total_master_val()
	total_win_val := int64(0)
	total_lose_val := int64(0)
	for _, acc := range self.accounts {
		bets := acc.BetVal
		win := int64(0)
		lose := int64(0)
		come_back := int64(0)

		// 1红、2黑、3特
		if reddWin {
			win += int64(bets[1])
			lose += int64(bets[2])
			come_back += int64(bets[1])

		} else {
			win += int64(bets[2])
			lose += int64(bets[1])
			come_back += int64(bets[2])
		}

		// 特殊牌型处理
		if specific_rate > 0 {
			win += int64(specific_rate) * int64(bets[3])
			come_back += int64(bets[3])
		} else {
			lose += int64(bets[3])
		}

		total_win_val += win   // 所有玩家赢得钱
		total_lose_val += lose // 所有玩家输得钱
		if win+come_back > 0 && acc.Robot == 0 {
			after -= win + come_back
		}

	}

	// 计算庄家得输赢 /////////////////////////////////////////////////////////////////////////////////////////////////////
	total_master_profit := total_lose_val * config.GetPublicConfig_Int64("R2B_SYSTEM_FEE") / 100
	total_master_profit -= total_win_val

	for _, master := range self.master_seats {
		if master == nil || master.Robot != 0 {
			continue
		}
		master_profit := (total_master_profit * master.Share) / total_share_val
		after -= master_profit
	}

	log.Infof(colorized.Gray("---------------before :%v after :%v------------------- "), before, after)
	return after
}

func (self *stop) Leave(now int64) {
	log.Debugf(colorized.Green("stop leave\n"))
}
func (self *stop) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_ENTER_GAME.UInt16(): // 客户端链接进入游戏
		self.Old_MSGID_R2B_ENTER_GAME(actor, msg, session)

	case protomsg.Old_MSGID_R2B_GAME_ENTER_GAME.UInt16(): // 客户端链接进入游戏
		self.Old_MSGID_R2B_GAME_ENTER_GAME(actor, msg, session)
	default:
		log.Warnf("stop 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}
	return true
}

func (self *stop) Old_MSGID_R2B_ENTER_GAME(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	_ = pack.ReadUInt32()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_ENTER_GAME.UInt16())
	if ret := self.canEnterRoom(accountId); ret > 0 {
		send.WriteUInt8(uint8(ret))
		send_tools.Send2Account(send.GetData(), session)
		return
	}
	self.enterRoom(accountId)

	now := utils.SecondTimeSince1970()
	// 通知客户端，进入游戏成功
	acc := account.AccountMgr.GetAccountByID(accountId)
	send2c := packet.NewPacket(nil)
	send2c.SetMsgID(protomsg.Old_MSGID_ENTER_GAME.UInt16())
	send2c.WriteUInt8(0)
	send2c.WriteUInt32(self.roomId)
	send_tools.Send2Account(send2c.GetData(), acc.SessionId)
	send2acc := self.sendGameData(acc, uint32(self.timestamp-now))
	send2acc.CatBody(self.pack)
	//_, r, b := algorithm.Compare(self.red_cards, self.black_cards)
	//send2acc.WriteUInt8(uint8(r))
	//send2acc.WriteUInt16(3)
	//send2acc.WriteUInt8(uint8(self.red_cards[0][0]))
	//send2acc.WriteUInt8(uint8(self.red_cards[0][1]))
	//send2acc.WriteUInt8(uint8(self.red_cards[1][0]))
	//send2acc.WriteUInt8(uint8(self.red_cards[1][1]))
	//send2acc.WriteUInt8(uint8(self.red_cards[2][0]))
	//send2acc.WriteUInt8(uint8(self.red_cards[2][1]))
	//
	//send2acc.WriteUInt8(uint8(b))
	//send2acc.WriteUInt16(3)
	//send2acc.WriteUInt8(uint8(self.black_cards[0][0]))
	//send2acc.WriteUInt8(uint8(self.black_cards[0][1]))
	//send2acc.WriteUInt8(uint8(self.black_cards[1][0]))
	//send2acc.WriteUInt8(uint8(self.black_cards[1][1]))
	//send2acc.WriteUInt8(uint8(self.black_cards[2][0]))
	//send2acc.WriteUInt8(uint8(self.black_cards[2][1]))
	send_tools.Send2Account(send2acc.GetData(), acc.SessionId)
}

func (self *stop) Old_MSGID_R2B_GAME_ENTER_GAME(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	_ = pack.ReadUInt32()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_R2B_GAME_ENTER_GAME.UInt16())
	if ret := self.canEnterRoom(accountId); ret > 0 {
		send.WriteUInt8(uint8(ret))
		send_tools.Send2Account(send.GetData(), session)
		return
	}
	self.enterRoom(accountId)

	now := utils.SecondTimeSince1970()
	// 通知客户端，进入游戏成功
	acc := account.AccountMgr.GetAccountByID(accountId)
	send2c := packet.NewPacket(nil)
	send2c.SetMsgID(protomsg.Old_MSGID_R2B_GAME_ENTER_GAME.UInt16())
	send2c.WriteUInt8(0)
	send2c.WriteUInt32(self.roomId)
	send_tools.Send2Account(send2c.GetData(), acc.SessionId)
	send2acc := self.sendGameData(acc, uint32(self.timestamp-now))
	send2acc.CatBody(self.pack)
	//_, r, b := algorithm.Compare(self.red_cards, self.black_cards)
	//send2acc.WriteUInt8(uint8(r))
	//send2acc.WriteUInt16(3)
	//send2acc.WriteUInt8(uint8(self.red_cards[0][0]))
	//send2acc.WriteUInt8(uint8(self.red_cards[0][1]))
	//send2acc.WriteUInt8(uint8(self.red_cards[1][0]))
	//send2acc.WriteUInt8(uint8(self.red_cards[1][1]))
	//send2acc.WriteUInt8(uint8(self.red_cards[2][0]))
	//send2acc.WriteUInt8(uint8(self.red_cards[2][1]))
	//
	//send2acc.WriteUInt8(uint8(b))
	//send2acc.WriteUInt16(3)
	//send2acc.WriteUInt8(uint8(self.black_cards[0][0]))
	//send2acc.WriteUInt8(uint8(self.black_cards[0][1]))
	//send2acc.WriteUInt8(uint8(self.black_cards[1][0]))
	//send2acc.WriteUInt8(uint8(self.black_cards[1][1]))
	//send2acc.WriteUInt8(uint8(self.black_cards[2][0]))
	//send2acc.WriteUInt8(uint8(self.black_cards[2][1]))
	send_tools.Send2Account(send2acc.GetData(), acc.SessionId)
}
