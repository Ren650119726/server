package room

import (
	"root/common"
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/mahjong-panda/event"
)

type (
	Robot_Deal struct {
		Room *Room
	}
)

func New_Behavior(room *Room) {
	obj := &Robot_Deal{Room: room}
	room.dispatcher.AddEventListener(event.EventType_ThreeChange, obj)
	room.dispatcher.AddEventListener(event.EventType_Deciding, obj)
	room.dispatcher.AddEventListener(event.EventType_Deal, obj)
	room.dispatcher.AddEventListener(event.EventType_Toss, obj)
	room.dispatcher.AddEventListener(event.EventType_Watting, obj)
}

func (self *Robot_Deal) OnEvent(e core.Event, t core.EventType) {
	switch t {
	case event.EventType_ThreeChange:
		wrap := e.(core.WrapEvent)
		data := wrap.Event.(*event.ThreeChange)
		self.tree_logic(data)

	case event.EventType_Deciding:
		wrap := e.(core.WrapEvent)
		data := wrap.Event.(*event.Deciding)
		self.decide_logic(data)

	case event.EventType_Deal:
		wrap := e.(core.WrapEvent)
		data := wrap.Event.(*event.EnterDeal)
		self.deal_logic(data)

	case event.EventType_Toss:
		wrap := e.(core.WrapEvent)
		data := wrap.Event.(*event.EnterToss)
		self.toss_logic(data)

	case event.EventType_Watting:
		wrap := e.(core.WrapEvent)
		data := wrap.Event.(*event.EnterWatting)
		self.watting_logic(data)
	}
}

func (self *Robot_Deal) watting_logic(ev *event.EnterWatting) {
	count_conf := 20
	if utils.Probability(30) {
		min := int32(-1)
		for _, player := range self.Room.seats {
			if player != nil {
				if min == -1 || (player.acc.Robot != 0 && min < player.acc.Games) {
					min = player.acc.Games
				}
			}

		}
		if min > int32(count_conf) {
			for _, player := range self.Room.seats {
				if player != nil && player.acc.Robot != 0 {
					accid := player.acc.AccountId
					self.Room.owner.AddTimer(int64(utils.Randx_y(100, 5000)), 1, func(dt int64) {
						msg := packet.NewPacket(nil)
						msg.SetMsgID(protomsg.Old_MSGID_LEAVE_GAME.UInt16())
						msg.WriteUInt32(accid)
						core.CoreSend(0, self.Room.owner.Id, msg.GetData(), 0)
					})
				}

			}

			return
		}

	}

	for _, player := range self.Room.seats {

		if player == nil || player.acc.Robot == 0 {
			continue
		}

		if player.acc.Games >= 16 && utils.Probability(40) {
			// 退出游戏
			robot := self.Room.getRobot()
			if robot != nil {
				self.Room.leaveRoom(player.acc.AccountId, true)
				msg := packet.NewPacket(nil)
				msg.SetMsgID(protomsg.Old_MSGID_PANDA_SIT_DOWN.UInt16())
				msg.WriteUInt32(robot.AccountId)
				core.CoreSend(0, self.Room.owner.Id, msg.GetData(), 0)
				return
			}
		} else {
			p := player
			self.Room.owner.AddTimer(int64(utils.Randx_y(1000, 4000)), 1, func(dt int64) {
				msg := packet.NewPacket(nil)
				msg.SetMsgID(protomsg.Old_MSGID_PANDA_PREPARE.UInt16())
				msg.WriteUInt32(p.acc.AccountId)
				core.CoreSend(0, self.Room.owner.Id, msg.GetData(), 0)
			})
		}
	}
}

func (self *Robot_Deal) tree_logic(ev *event.ThreeChange) {
	player := self.Room.seats[ev.Index]
	if player == nil {
		log.Warnf("player == nil inded:%v ", ev.Index)
		return
	}
	if player.acc.Robot == 0 {
		return
	}

	interval := int64(0)
	if utils.Probability(70) {
		interval = int64(utils.Randx_y(int(5500), int(9000)))
	} else if utils.Probability(70) {
		interval = int64(utils.Randx_y(int(10000), int(15000)))
	} else {
		interval = int64(utils.Randx_y(int(15000), int(35000)))
	}
	self.Room.owner.AddTimer(interval, 1, func(dt int64) {
		send := packet.NewPacket(nil)
		send.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_CHANGE_CARDS_CONFIRM.UInt16())
		send.WriteUInt32(player.acc.AccountId)
		send.WriteUInt16(3)
		send.WriteUInt8(ev.CardsIndex[0] + 1)
		send.WriteUInt8(ev.CardsIndex[1] + 1)
		send.WriteUInt8(ev.CardsIndex[2] + 1)
		core.CoreSend(0, self.Room.owner.Id, send.GetData(), 0)
	})
}

func (self *Robot_Deal) decide_logic(ev *event.Deciding) {
	player := self.Room.seats[ev.Index]
	if player == nil {
		log.Warnf("player == nil inded:%v ", ev.Index)
		return
	}
	if player.acc.Robot == 0 {
		return
	}

	interval := int64(0)
	if utils.Probability(70) {
		interval = int64(utils.Randx_y(int(500), int(4000)))
	} else if utils.Probability(70) {
		interval = int64(utils.Randx_y(int(5000), int(10000)))
	} else {
		interval = int64(utils.Randx_y(int(10000), int(30000)))
	}

	self.Room.owner.AddTimer(interval, 1, func(dt int64) {
		send := packet.NewPacket(nil)
		send.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_REQUEST_DECIDE.UInt16())
		send.WriteUInt32(player.acc.AccountId)
		send.WriteInt8(int8(ev.Type))
		core.CoreSend(0, self.Room.owner.Id, send.GetData(), 0)
	})
}

func (self *Robot_Deal) deal_logic(ev *event.EnterDeal) {
	player := self.Room.seats[ev.Index]
	if player == nil {
		log.Warnf("player == nil inded:%v ", ev.Index)
		return
	}
	if player.acc.Robot == 0 {
		return
	}

	interval := int64(0)
	conf1 := config.GetPublicConfig_Slice("DGK_ROBOT_CONFIG1")
	conf2 := config.GetPublicConfig_Slice("DGK_ROBOT_CONFIG2")
	conf3 := config.GetPublicConfig_Slice("DGK_ROBOT_CONFIG3")
	probability1 := conf1[0]
	probability2 := conf2[0]
	min1, max1 := conf1[1], conf1[2]
	min2, max2 := conf2[1], conf2[2]
	if utils.Probability(probability1) {
		interval = int64(utils.Randx_y(int(min1), int(max1)))
	} else if utils.Probability(probability2) {
		interval = int64(utils.Randx_y(int(min2), int(max2)))
	} else {
		interval = int64(utils.Randx_y(int(conf3[0]), int(conf3[1])))
	}
	self.Room.owner.AddTimer(interval, 1, func(dt int64) {
		if ev.Bhu {
			msg := packet.NewPacket(nil)
			msg.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_HU_REQ.UInt16())
			msg.WriteUInt32(player.acc.AccountId)
			msg.WriteUInt8(1)
			core.CoreSend(0, self.Room.owner.Id, msg.GetData(), 0)
			return
		}

		// 能杠就杠
		if l := len(ev.Gangs); l != 0 {
			ri := utils.Randx_y(0, l)
			msg := packet.NewPacket(nil)
			msg.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_GANG_REQ.UInt16())
			msg.WriteUInt8(uint8(ev.Index + 1))
			msg.WriteUInt8(player.cards.hand[ev.Gangs[ri]].Value())
			core.CoreSend(0, self.Room.owner.Id, msg.GetData(), 0)
			return
		}

		var statis map[int]*statisics
		// 判断打一张牌
		statis = player.cards.Classification()

		d := []common.EMaJiangType{}
		for _, card := range player.cards.hand {
			if int8(card/10) == player.decide_t {
				d = append(d, card)
			}
		}
		// 找到散牌，就随机打一张////////////////////////////////////////////////////////////////////////////////////////
		single_len := 0
		st := 0
		for t, data := range statis {
			l := len(data.single)
			if single_len == 0 || single_len < l {
				single_len = l
				st = t
			}
		}
		if single_len != 0 {
			card := statis[st].single[utils.Randx_y(0, len(statis[st].single))]
			for i, cardh := range player.cards.hand {
				if card == cardh {
					if len(d) == 0 || int8(card/10) == player.decide_t {
						msg := packet.NewPacket(nil)
						msg.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_PUSH_CARD_REQ.UInt16())
						msg.WriteUInt8(uint8(ev.Index + 1))
						msg.WriteUInt8(uint8(i + 1))
						core.CoreSend(0, self.Room.owner.Id, msg.GetData(), 0)
						return
					}

				}
			}

		}

		// 如果没有找到散牌，牌少的打 ///////////////////////////////////////////////////////////////////////////////////
		//ct := 0
		//if len(statis) == 2 {
		//	if statis[1].len < statis[2].len {
		//		ct = 1
		//	} else if statis[1].len > statis[2].len {
		//		ct = 2
		//	}
		//
		//	if ct != 0 {
		//		for i, card := range player.cards.hand {
		//			if int(card/10) == ct {
		//				msg := packet.NewPacket(nil)
		//				msg.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_PUSH_CARD_REQ.UInt16())
		//				msg.WriteUInt8(uint8(ev.Index + 1))
		//				msg.WriteUInt8(uint8(i + 1))
		//				core.CoreSend(0, self.Room.owner.Id, msg.GetData(), 0)
		//				return
		//			}
		//		}
		//	}
		//}

		//////// 随机打一张 /////////////////////////////////////////////////////////////////////////////////////////////
		ri := utils.Randx_y(0, len(d))
		if len(d) != 0 {
			for i, card := range player.cards.hand {
				if card == d[ri] {
					msg := packet.NewPacket(nil)
					msg.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_PUSH_CARD_REQ.UInt16())
					msg.WriteUInt8(uint8(ev.Index + 1))
					msg.WriteUInt8(uint8(i + 1))
					core.CoreSend(0, self.Room.owner.Id, msg.GetData(), 0)
					return
				}
			}
		} else {
			ri := utils.Randx_y(0, len(player.cards.hand))
			msg := packet.NewPacket(nil)
			msg.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_PUSH_CARD_REQ.UInt16())
			msg.WriteUInt8(uint8(ev.Index + 1))
			msg.WriteUInt8(uint8(ri + 1))
			core.CoreSend(0, self.Room.owner.Id, msg.GetData(), 0)
			return
		}

	})
}

func (self *Robot_Deal) toss_logic(ev *event.EnterToss) {
	player := self.Room.seats[ev.Index]
	if player == nil {
		log.Warnf("player == nil inded:%v ", ev.Index)
		return
	}
	if player.acc.Robot == 0 {
		return
	}

	interval := int64(0)
	conf1 := config.GetPublicConfig_Slice("DGK_ROBOT_CONFIG1")
	conf2 := config.GetPublicConfig_Slice("DGK_ROBOT_CONFIG2")
	conf3 := config.GetPublicConfig_Slice("DGK_ROBOT_CONFIG3")
	probability1 := conf1[0]
	probability2 := conf2[0]
	min1, max1 := conf1[1], conf1[2]
	min2, max2 := conf2[1], conf2[2]
	if utils.Probability(probability1) {
		interval = int64(utils.Randx_y(int(min1), int(max1)))
	} else if utils.Probability(probability2) {
		interval = int64(utils.Randx_y(int(min2), int(max2)))
	} else {
		interval = int64(utils.Randx_y(int(conf3[0]), int(conf3[1])))
	}
	self.Room.owner.AddTimer(interval, 1, func(dt int64) {
		if ev.Bhu {
			msg := packet.NewPacket(nil)
			msg.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_HU_REQ.UInt16())
			msg.WriteUInt32(player.acc.AccountId)
			msg.WriteUInt8(1)
			core.CoreSend(0, self.Room.owner.Id, msg.GetData(), 0)
			return
		}

		// 能杠就杠
		if ev.Gangs {
			msg := packet.NewPacket(nil)
			msg.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_GANG_REQ.UInt16())
			msg.WriteUInt8(uint8(ev.Index + 1))
			core.CoreSend(0, self.Room.owner.Id, msg.GetData(), 0)
			return
		}

		// 能碰
		if ev.Peng {
			if utils.Probability(70) {
				msg := packet.NewPacket(nil)
				msg.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_PENG_REQ.UInt16())
				msg.WriteUInt8(uint8(ev.Index + 1))
				core.CoreSend(0, self.Room.owner.Id, msg.GetData(), 0)
			} else {
				msg := packet.NewPacket(nil)
				msg.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_GUO_REQ.UInt16())
				msg.WriteUInt8(uint8(ev.Index + 1))
				core.CoreSend(0, self.Room.owner.Id, msg.GetData(), 0)
			}
			return
		}
	})
}
