package room

import (
	"root/common"
	"root/common/config"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/mahjong-panda/account"
	"root/server/mahjong-panda/algorithm"
	"root/server/mahjong-panda/event"
	"root/server/mahjong-panda/send_tools"
)

type (
	changing struct {
		*playing
		s         int32
		timestamp int64 // 结算倒计时 时间戳 豪秒

		change_cards map[uint32][]uint8
		opt          map[uint32]uint8
		delay        bool
	}
)

const ANIMITION_TIME = 3000

func (self *changing) Enter(now int64) {
	duration := config.GetPublicConfig_Int64("PANDA_CHANGE_STATE_TIME") // 持续时间 秒
	self.timestamp = utils.SecondTimeSince1970() + int64(duration)
	self.change_cards = make(map[uint32][]uint8)
	self.opt = make(map[uint32]uint8)
	self.track_log(colorized.Gray("--- changing enter duration:%v"), duration)

	// 从散牌里筛选
	single_fun := func(player *GamePlayer, single []common.EMaJiangType) {
		for _, card := range single {
			for i, hc := range player.cards.hand {
				if card == hc {
					inc := false
					for _, iii := range self.change_cards[player.acc.AccountId] {
						if iii == uint8(i) {
							self.change_cards[player.acc.AccountId] = append(self.change_cards[player.acc.AccountId], uint8(i+1))
							inc = true
							break
						}
					}
					if !inc {
						self.change_cards[player.acc.AccountId] = append(self.change_cards[player.acc.AccountId], uint8(i))
					}
					break
				}
			}
		}
	}
	for _, player := range self.seats {
		self.opt[player.acc.AccountId] = 0
		jk := player.cards.Classification()

		min := 0
		equalt := 0
		t := 0
		for i, v := range jk {
			if v.len <= 2 {
				delete(jk, i)
				continue
			}

			if min > v.len || min == 0 {
				min = v.len
				t = i
			} else if min == v.len {
				equalt = 1
			}
		}
		if equalt == 0 {
			self.change_cards[player.acc.AccountId] = make([]uint8, 0)
			l := len(jk[t].single)
			if l >= 3 {
				single_fun(player, jk[t].single[:3])
			} else {
				single_fun(player, jk[t].single)
				remain := 3 - l
				i := 0
				for _, card := range player.cards.hand {
					if int(card/10) == t {
						e := false
						for _, cs := range jk[t].single {
							if card == cs {
								e = true
								break
							}
						}
						if !e {
							i++
							single_fun(player, []common.EMaJiangType{card})
							if i == remain {
								break
							}
						}
					}

				}
			}
			continue
		}

		for _, v := range jk {
			if len(v.single) >= 3 {
				self.change_cards[player.acc.AccountId] = make([]uint8, 0)
				single_fun(player, v.single[:3])
				break
			}
		}

		if self.change_cards[player.acc.AccountId] == nil {
			// 判断杠
			for t, v := range jk {
				if len(v.single) == 2 && v.coincide[4] == 0 {
					self.change_cards[player.acc.AccountId] = make([]uint8, 0)
					single_fun(player, v.single)
					if l := len(self.change_cards[player.acc.AccountId]); l != 2 {
						log.Warnf("错误！！！！！的计算长度:%v cards:%v", l, player.cards.hand)
						continue
					}
					// 再额外选一张
					for i, card := range player.cards.hand {
						if int(card/10) == t {
							exist := false
							for _, cc := range self.change_cards[player.acc.AccountId] {
								if uint8(i) == cc {
									exist = true
									break
								}
							}
							if !exist {
								self.change_cards[player.acc.AccountId] = append(self.change_cards[player.acc.AccountId], uint8(i))
								break
							}
						}
					}
					break
				}

			}
		}
		// 如果还没选出来，就找1张单牌的选了
		if self.change_cards[player.acc.AccountId] == nil {
			// 判断杠
			for t, v := range jk {
				if len(v.single) == 1 && v.coincide[4] == 0 {
					self.change_cards[player.acc.AccountId] = make([]uint8, 0)
					single_fun(player, v.single)
					if l := len(self.change_cards[player.acc.AccountId]); l != 1 {
						log.Warnf("错误！！！！！的计算长度:%v cards:%v", l, player.cards.hand)
						continue
					}
					// 再额外选两张
					count := 0
					for i, card := range player.cards.hand {
						if int(card/10) == t {
							exist := false
							for _, cc := range self.change_cards[player.acc.AccountId] {
								if uint8(i) == cc {
									exist = true
									break
								}
							}
							if !exist {
								self.change_cards[player.acc.AccountId] = append(self.change_cards[player.acc.AccountId], uint8(i))
								count++
								if count == 2 {
									break
								}
							}
						}
					}
					break
				}

			}

		}
		// 如果还没选出来.....那就随鸡巴
		if self.change_cards[player.acc.AccountId] == nil {
			self.change_cards[player.acc.AccountId] = make([]uint8, 0)
			count := 0
			for t := range jk {
				for i, card := range player.cards.hand {
					if count == 3 {
						break
					}
					if int(card/10) == t {
						self.change_cards[player.acc.AccountId] = append(self.change_cards[player.acc.AccountId], uint8(i))
						count++

					}
				}
			}
		}

	}

	for id, acc := range self.accounts {
		if index := self.seatIndex(id); index != -1 {
			player := self.seats[index]
			msg := packet.NewPacket(nil)
			msg.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_CHANGE_STATE.UInt16())
			msg.WriteInt64(self.timestamp * 1000)
			msg.WriteUInt16(uint16(len(self.change_cards[player.acc.AccountId])))
			for _, index := range self.change_cards[player.acc.AccountId] {
				msg.WriteUInt8(uint8(player.cards.hand[index]))
				log.Debugf("player:%v card:%v ", player.acc.AccountId, uint8(player.cards.hand[index]))
			}
			send_tools.Send2Account(msg.GetData(), player.acc.SessionId)

			if len(self.change_cards[player.acc.AccountId]) != 3 {
				log.Errorf("room:%v accid:%v 数据错误，默认选出来的牌不是3张:%v ", self.roomId, acc.AccountId, self.change_cards[player.acc.AccountId])
				return
			}
			self.dispatcher.Dispatch(&event.ThreeChange{Index: index, CardsIndex: self.change_cards[player.acc.AccountId]}, event.EventType_ThreeChange)
			self.track_log(colorized.Gray("推荐给玩家:%v 的牌:%v "), player.acc.AccountId, self.change_cards[player.acc.AccountId])
		} else {
			msg := packet.NewPacket(nil)
			msg.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_CHANGE_STATE.UInt16())
			msg.WriteInt64(self.timestamp * 1000)
			msg.WriteUInt16(0)
			send_tools.Send2Account(msg.GetData(), acc.SessionId)
		}
	}
	self.delay = false
}

func (self *changing) Tick(now int64) {
	if now >= self.timestamp {
		self.default_choice()
	}
}

func (self *changing) default_choice() {
	for accID := range self.change_cards {
		if self.opt[accID] == 0 {
			self.opt[accID] = 1
			broadcast := packet.NewPacket(nil)
			broadcast.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_CHANGE_CARDS_CONFIRM.UInt16())
			broadcast.WriteUInt8(0)
			broadcast.WriteUInt8(uint8(self.seatIndex(accID)) + 1)
			self.SendBroadcast(broadcast.GetData())
		}

	}

	self.over()
}

func (self *changing) over() {
	if self.delay {
		return
	}
	self.delay = true
	z := utils.Randx_y(0, 2) // 0 顺时针 1 逆时针

	// 先把要换的牌存下来，并且从手牌中删除
	all_change_cards := make(map[uint32][]common.EMaJiangType)
	for _, p := range self.seats {
		// 存
		cards := []common.EMaJiangType{}
		for _, i := range self.change_cards[p.acc.AccountId] {
			cards = append(cards, p.cards.hand[i])
		}
		all_change_cards[p.acc.AccountId] = cards

		// 手牌中删除
		for _, card := range cards {
			for i, hcard := range p.cards.hand {
				if hcard == card {
					p.cards.hand = append(p.cards.hand[:i], p.cards.hand[i+1:]...)
					break
				}
			}
		}
	}

	change := func(prev, next *GamePlayer) {
		if all_change_cards[prev.acc.AccountId] != nil {
			sendMsg := packet.NewPacket(nil)
			sendMsg.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_CHANGE_STATE_OVER.UInt16())
			sendMsg.WriteUInt8(uint8(self.seatIndex(prev.acc.AccountId) + 1))
			sendMsg.WriteUInt16(3)

			for _, card := range all_change_cards[prev.acc.AccountId] {
				next.cards.hand, _ = algorithm.InsertCard(next.cards.hand, card)
				sendMsg.WriteUInt8(card.Value())
			}

			sendMsg.WriteUInt16(uint16(len(next.cards.hand)))
			for _, card := range next.cards.hand {
				sendMsg.WriteUInt8(card.Value())
			}
			self.track_log("玩家:%v 三张:%v 换牌后的手牌:%v ", next.acc.AccountId, all_change_cards[prev.acc.AccountId], next.cards.hand)

			send_tools.Send2Account(sendMsg.GetData(), next.acc.SessionId)
		}
	}

	l := len(self.seats)
	if z == 0 {
		for i := 0; i < l; i++ {
			if i == l-1 {
				change(self.seats[i], self.seats[0])
			} else {
				change(self.seats[i], self.seats[i+1])
			}
		}
	} else if z == 1 {
		for i := l - 1; i >= 0; i-- {
			if i == 0 {
				change(self.seats[i], self.seats[l-1])
			} else {
				change(self.seats[i], self.seats[i-1])
			}
		}
	}

	self.owner.AddTimer(ANIMITION_TIME, 1, func(dt int64) {
		self.game_state.Swtich(0, DEAL_STATE)
	})

}
func (self *changing) Combine_Game_MSG(pack packet.IPacket, acc *account.Account) {
	pack.WriteInt64(self.timestamp * 1000)
	cards := self.change_cards[acc.AccountId]
	if cards == nil {
		pack.WriteUInt8(0)
		pack.WriteUInt16(0)
	} else {
		player := self.seats[self.seatIndex(acc.AccountId)]
		pack.WriteUInt8(self.opt[player.acc.AccountId])
		pack.WriteUInt16(uint16(len(cards)))
		for _, i := range cards {

			pack.WriteUInt8(uint8(player.cards.hand[i]))
		}
	}

	pack.WriteUInt16(uint16(len(self.seats)))
	for i, v := range self.seats {
		pack.WriteUInt8(uint8(i + 1))
		pack.WriteUInt8(self.opt[v.acc.AccountId])
	}
}
func (self *changing) Leave(now int64) {
	self.track_log(colorized.Gray("--- changing leave\n"))
}

///////////////////////////////// handler ///////////////////////////////////////////////////
func (self *changing) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_PANDA_GAME_CHANGE_CARDS_CONFIRM.UInt16(): // 确定要换的牌
		self.Old_MSGID_PANDA_GAME_CHANGE_CARDS_CONFIRM(actor, msg, session)
	default:
		log.Warnf("changing 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}
	return true
}

// 玩家确定换牌
func (self *changing) Old_MSGID_PANDA_GAME_CHANGE_CARDS_CONFIRM(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accID := pack.ReadUInt32()
	size := pack.ReadUInt16()
	if self.change_cards[accID] == nil {
		return
	}

	if self.delay {
		return
	}

	if self.opt[accID] == 1 {
		log.Warnf("重复发 确定换牌消息")
		return
	}
	i := self.seatIndex(accID)
	if i == -1 {
		log.Warnf("找不到座位:%v ", accID)
		return
	}
	player := self.seats[i]
	arr_i := []uint8{}
	hand_len := len(player.cards.hand)
	hua := 0
	for i := uint16(0); i < size; i++ {
		index := pack.ReadUInt8()
		index--
		if index < 0 || index > 100 {
			log.Warnf("玩家:%v 客户端发来的确定换牌数据错误:%v ", accID, index)
			return
		}

		if int(index) >= hand_len {
			log.Warnf("玩家:%v 数组越界:%v %v", accID, index, len(player.cards.hand))
			return
		}
		for _, v := range arr_i {
			if v == index {
				log.Warnf("玩家:%v 有重复的索引:%v ", arr_i, index)
				return
			}
		}
		arr_i = append(arr_i, index)

		if hua == 0 {
			hua = int(player.cards.hand[int(index)]) / 10
		} else if int(player.cards.hand[int(index)])/10 != hua {
			msg := packet.NewPacket(nil)
			msg.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_CHANGE_CARDS_CONFIRM.UInt16())
			msg.WriteUInt8(1)
			msg.WriteUInt8(uint8(self.seatIndex(accID)) + 1)
			send_tools.Send2Account(msg.GetData(), session)
			return
		}
	}

	self.change_cards[accID] = arr_i
	self.opt[player.acc.AccountId] = 1

	broadcast := packet.NewPacket(nil)
	broadcast.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_CHANGE_CARDS_CONFIRM.UInt16())
	broadcast.WriteUInt8(0)
	broadcast.WriteUInt8(uint8(self.seatIndex(accID)) + 1)
	self.SendBroadcast(broadcast.GetData())
	self.track_log("玩家请求换三张:%v", arr_i)

	ov := true
	for _, v := range self.opt {
		if v == 0 {
			ov = false
			break
		}
	}
	if ov {
		self.over()
	}
}
