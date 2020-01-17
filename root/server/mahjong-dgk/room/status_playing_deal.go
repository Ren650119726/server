package room

import (
	"root/common"
	ca "root/common/algorithm"
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/mahjong-dgk/account"
	"root/server/mahjong-dgk/algorithm"
	"root/server/mahjong-dgk/event"
	"sort"
)

type (
	deal struct {
		*playing
		s         int32
		timestamp int64 // 结算倒计时 时间戳 豪秒
		gangs     []int
		bguo      int8

		gangjiao_zimo map[int][]algorithm.Jiao_Card // 杠叫

		send packet.IPacket
		bhu  bool
		qhu  bool

		animation_ bool
		delay      bool
		delay2     bool
		zq         uint8
	}
)

func (self *deal) Enter(now int64) {
	self.animation_ = false
	duration := config.GetPublicConfig_Int64("DGK_DEAL_TIME") // 持续时间 秒
	self.timestamp = utils.SecondTimeSince1970() + int64(duration)
	self.track_log(colorized.Yellow("--- deal enter duration:%v"), duration)
	self.delay2 = true
	self.owner.AddTimer(500, 1, func(dt int64) {
		self.delay2 = false
		self.bguo = 0
		self.gangs = []int{}

		gamePlayer := self.seats[self.deal_player]
		if gamePlayer == nil {
			log.Warnf("gamePlayer == nil  index:%v", self.deal_player)
			return
		}
		self.track_log(colorized.Yellow("--- 当前打牌的玩家:%v 牌:%v"), gamePlayer.acc.AccountId, gamePlayer.cards.String())

		self.all_hu = []*ca.Majiang_Hu{}
		// 判断当前杠、胡
		if !self.deal_peng {
			self.all_hu, self.zhua_bao, self.zq, _ = self.calcExtra(gamePlayer, 0, true)
			so := &Majiang_fan_Sort{Room: self.Room, All: self.all_hu}
			sort.Sort(so)
			self.all_hu = so.All
			// 胡

			// 打完牌，不能杠
			if len(self.cards) > 0 {
				self.gangs = algorithm.AllGang(gamePlayer.cards.hand, gamePlayer.cards.peng)
			}
		}

		if self.qingfu[self.deal_player] {
			self.delay = true
			self.owner.AddTimer(500, 1, func(dt int64) {
				self.delay = false
				self.track_log(colorized.Yellow("--- 自动请胡"))
				qing_msg := packet.NewPacket(nil)
				qing_msg.SetMsgID(protomsg.Old_MSGID_DGK_GAME_HU_REQ.UInt16())
				qing_msg.WriteUInt32(gamePlayer.acc.AccountId)
				qing_msg.WriteUInt8(2)
				core.CoreSend(0, int32(self.roomId), qing_msg.GetData(), 0)
			})
			return
		}
		// 如果玩家报叫，需要额外判断是否能杠
		if (gamePlayer.jiao != nil || self.master_bj) && len(self.gangs) > 0 {
			self.gangjiao_zimo = make(map[int][]algorithm.Jiao_Card)
			for i := len(self.gangs) - 1; i >= 0; i-- {
				v := self.gangs[i]
				card := gamePlayer.cards.hand[v]
				hand := []common.EMaJiangType{}
				gang := [][]common.EMaJiangType{}
				hand = append(hand, gamePlayer.cards.hand[:v]...)
				hand = append(hand, gamePlayer.cards.hand[v+4:]...)
				gang = append(gang, []common.EMaJiangType{card, card, card, card})
				j := algorithm.Jiao_(hand, gamePlayer.cards.peng, gang)
				if len(j) == 0 {
					self.gangs = append(self.gangs[:i], self.gangs[i+1:]...)
				} else {
					self.gangjiao_zimo[v] = j
				}
			}
		}
		// 通知客户端打牌
		self.send = packet.NewPacket(nil)
		self.send.SetMsgID(protomsg.Old_MSGID_DGK_GAME_PUSH_CARD_NOTICE.UInt16())
		self.send.WriteInt8(int8(self.deal_player + 1))
		self.send.WriteInt64(self.timestamp * 1000)
		l := len(self.all_hu)
		self.bhu = false
		for _, v := range self.all_hu {
			if v.HuType != common.HU_QING_WU_DUI && v.HuType != common.HU_WU_DUI {
				self.bhu = true
				break
			}
		}
		if self.bhu {
			self.send.WriteInt8(1)
		} else {
			self.send.WriteInt8(0)
		}

		self.qhu = false
		for _, v := range self.all_hu {
			if v.HuType == common.HU_QING_WU_DUI || v.HuType == common.HU_WU_DUI {
				self.qhu = true
				break
			}
		}

		if self.qhu {
			self.send.WriteInt8(1)
		} else {
			self.send.WriteInt8(0)
		}

		if self.gangs == nil {
			self.send.WriteInt16(0)

		} else {
			self.send.WriteInt16(int16(len(self.gangs)))
			for _, v := range self.gangs {
				self.send.WriteInt8(int8(v + 1))
			}
		}

		self.track_log(colorized.Yellow("--- 当前打牌的玩家:%v hu:%v gang:%v"), gamePlayer.acc.AccountId, self.all_hu, self.gangs)

		// 不能胡，也不能杠
		if l == 0 && len(self.gangs) == 0 {
			// 如果玩家报叫 不能胡也不能杠，自动打牌
			if gamePlayer.jiao != nil && len(gamePlayer.jiao) != 0 {
				auto_push := packet.NewPacket(nil)
				auto_push.SetMsgID(protomsg.Old_MSGID_DGK_GAME_PUSH_CARD_REQ.UInt16())
				auto_push.WriteInt8(int8(self.deal_player) + 1)
				auto_push.WriteInt8(int8(gamePlayer.cards.last_index) + 1)
				core.CoreSend(0, self.owner.Id, auto_push.GetData(), 0)
				return
			}
		}

		self.SendBroadcast(self.send.GetData())

		self.dispatcher.Dispatch(&event.EnterDeal{
			Index: self.deal_player,
			Bhu:   self.bhu,
			Qhu:   self.qhu,
			Gangs: self.gangs,
		}, event.EventType_Deal)
	})

}

func (self *deal) Tick(now int64) {
	if self.animation_ || self.delay || self.delay2 {
		return
	}
	gamePlayer := self.seats[self.deal_player]
	if gamePlayer == nil {
		log.Warnf("gamePlayer == nil,  deal_player:%v ", self.deal_player)
		return
	}
	if gamePlayer.trusteeship == 1 {
		if self.bhu {
			auto_send := packet.NewPacket(nil)
			auto_send.SetMsgID(protomsg.Old_MSGID_DGK_GAME_HU_REQ.UInt16())
			auto_send.WriteUInt32(gamePlayer.acc.AccountId)
			auto_send.WriteUInt8(1)
			core.CoreSend(self.owner.Id, self.owner.Id, auto_send.GetData(), 0)
		} else if self.qhu {
			auto_send := packet.NewPacket(nil)
			auto_send.SetMsgID(protomsg.Old_MSGID_DGK_GAME_HU_REQ.UInt16())
			auto_send.WriteUInt32(gamePlayer.acc.AccountId)
			auto_send.WriteUInt8(2)
			core.CoreSend(self.owner.Id, self.owner.Id, auto_send.GetData(), 0)
		} else {
			index := gamePlayer.cards.last_index
			auto_push := packet.NewPacket(nil)
			auto_push.SetMsgID(protomsg.Old_MSGID_DGK_GAME_PUSH_CARD_REQ.UInt16())
			auto_push.WriteInt8(int8(self.deal_player) + 1)
			auto_push.WriteInt8(int8(index) + 1)
			core.CoreSend(0, self.owner.Id, auto_push.GetData(), 0)
		}

		return
	}
	if now >= self.timestamp {
		// 时间到了，当前玩家还没有打牌，就随机打一张
		self.trusateeship(uint8(self.deal_player))
		self.track_log(colorized.Yellow("--- 超时 自动帮玩家:%v 打牌"), gamePlayer.acc.AccountId)
		index := gamePlayer.cards.last_index
		if len(self.master_bjs) != 0 {
			index = self.master_bjs[0]
			self.track_log("庄家报叫可以打牌:%v超时自动帮打 :%v ", self.master_bjs, index)
		}

		auto_push := packet.NewPacket(nil)
		auto_push.SetMsgID(protomsg.Old_MSGID_DGK_GAME_PUSH_CARD_REQ.UInt16())
		auto_push.WriteInt8(int8(self.deal_player) + 1)
		auto_push.WriteInt8(int8(index) + 1)
		core.CoreSend(0, self.owner.Id, auto_push.GetData(), 0)
		return
	}
}

// 打出一张牌 t:0 普通牌 1 弯杠 2 请胡
func (self *deal) push_card(index int8, carIndex int, t int8) {
	if index < 0 || int(index) >= len(self.seats) {
		log.Errorf("越界 :%v", index)
		return
	}

	gamePlayer := self.seats[index]
	if carIndex < 0 || int(carIndex) >= len(gamePlayer.cards.hand) {
		log.Warnf("牌越界 index:%v len:%v", carIndex, len(gamePlayer.cards.hand))
		return
	}
	card := gamePlayer.cards.hand[carIndex]

	if gamePlayer.jiao == nil {
		gamePlayer.exclude_hu = 0 // 清除胡牌限制
	}
	all_hu, _, zq, _ := self.calcExtra(gamePlayer, 0, true)
	so := &Majiang_fan_Sort{Room: self.Room, All: all_hu}
	sort.Sort(so)
	all_hu = so.All
	if len(all_hu) != 0 {
		gamePlayer.exclude_hu = int(self.hu_fan[int32(all_hu[0].HuType)])
		if all_hu[0].Extra != nil {
			for t, v := range all_hu[0].Extra {
				gamePlayer.exclude_hu += int(self.extra_fan[int32(t)] * int32(v))
			}
		}

		if zq != 0 {
			gamePlayer.exclude_hu += int(zq - 1)
		}
	}

	gamePlayer.trash_cards = append(gamePlayer.trash_cards, card)
	gamePlayer.cards.hand = append(gamePlayer.cards.hand[:carIndex], gamePlayer.cards.hand[carIndex+1:]...)
	self.track_log(colorized.Yellow("--- 打出牌:（%v） 玩家:%v 位置:%v 牌位:%v 垃圾牌:%v"), card, gamePlayer.acc.AccountId, index, carIndex, gamePlayer.trash_cards)

	self.wanGang_qinHu = int(t)
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_DGK_GAME_PUSH_CARD_REQ.UInt16())
	send.WriteInt8(index + 1)
	send.WriteInt8(int8(carIndex) + 1)
	send.WriteInt8(int8(card))
	send.WriteInt8(int8(t))
	self.SendBroadcast(send.GetData())

	self.push_count++
	self.last_push_index = index
	self.last_push_cardIndex = int8(carIndex)

	if carIndex < gamePlayer.cards.last_index {
		gamePlayer.cards.last_index--
	} else if carIndex == gamePlayer.cards.last_index {
		gamePlayer.cards.last_index = len(gamePlayer.cards.hand) - 1
	}
	// 进入喊话状态
	self.game_state.Swtich(0, BREAKIN_STATE)
}

func (self *deal) Combine_Game_MSG(pack packet.IPacket, acc *account.Account) {
	if !self.delay2 {
		pack.CatBody(self.send)
	} else {
		s := packet.NewPacket(nil)
		s.WriteInt8(int8(self.deal_player + 1))
		s.WriteInt64(self.timestamp * 1000)
		s.WriteInt8(0)
		s.WriteInt8(0)
	}

	pack.WriteInt8(int8(self.seats[self.deal_player].cards.last_index + 1))
	pack.WriteInt8(int8(self.bguo))
}
func (self *deal) Leave(now int64) {
	self.deal_peng = false
	self.track_log(colorized.Yellow("--- deal leave\n"))
}

///////////////////////////////// handler ///////////////////////////////////////////////////
func (self *deal) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_DGK_GAME_GUO_REQ.UInt16(): // 过
		self.Old_MSGID_DGK_GAME_GUO_REQ(actor, msg, session)
	case protomsg.Old_MSGID_DGK_GAME_PUSH_CARD_REQ.UInt16(): // 请求打牌
		self.Old_MSGID_DGK_GAME_PUSH_CARD_REQ(actor, msg, session)
	case protomsg.Old_MSGID_DGK_GAME_HU_REQ.UInt16(): // 请求胡
		self.Old_MSGID_DGK_GAME_HU_REQ(actor, msg, session)
	case protomsg.Old_MSGID_DGK_GAME_GANG_REQ.UInt16(): // 请求杠
		self.Old_MSGID_DGK_GAME_GANG_REQ(actor, msg, session)
	default:
		log.Warnf("deal 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}
	return true
}

// 过
func (self *deal) Old_MSGID_DGK_GAME_GUO_REQ(actor int32, msg []byte, session int64) {
	if self.delay2 {
		log.Warnf("消息发早了 过")
		return
	}
	pack := packet.NewPacket(msg)
	index := int(pack.ReadInt8()) - 1
	if index == -1 {
		log.Warnf("出错！%v", index)
		return
	}

	if index != self.deal_player {
		log.Warnf("出错！!!!   %v %v", index, self.deal_player)
		return
	}
	gamePlayer := self.seats[index]
	self.bguo = 1
	self.track_log(colorized.Yellow("--- 座位:%v 选择【过】"), index)

	if gamePlayer.jiao != nil {
		self.push_card(int8(self.deal_player), gamePlayer.cards.last_index, 0)
		return
	}
}

// 玩家请求打牌
func (self *deal) Old_MSGID_DGK_GAME_PUSH_CARD_REQ(actor int32, msg []byte, session int64) {
	if self.delay2 {
		log.Warnf("消息发早了 打牌")
		return
	}
	pack := packet.NewPacket(msg)
	index := pack.ReadInt8() - 1
	cardIndex := pack.ReadInt8() - 1

	if index != int8(self.deal_player) {
		log.Warnf("还未轮到:%v 号位打牌， 当前打牌:%v 号位", index, self.deal_player)
		return
	}
	self.track_log(colorized.Yellow("--- 座位:%v 请求打出牌位:%v "), index, cardIndex)

	self.push_card(index, int(cardIndex), 0) // 玩家主动打牌
	// 庄家的情况，特殊处理，因为庄家，首发11张牌，报叫必须打一张
	if self.master_bj {
		// 校验一下 如果庄家报叫，那么这次打牌的肯定是庄家
		if self.master != int(index) {
			log.Errorf("bug!！ 打牌的不是庄家!! index :%v cardIndex:%v", index, cardIndex)
			return
		}

		self.master_bjs = []int{}
		self.master_bj = false
		gamePlayer := self.seats[index]
		j := algorithm.Jiao_(gamePlayer.cards.hand, gamePlayer.cards.peng, gamePlayer.cards.gang)
		gamePlayer.jiao = j
		self.track_log(colorized.Yellow("--- 庄家报叫【打出牌】后，叫牌:%v"), j)

		// 帮助客户端做校验 如果庄家报叫，那么打出去的牌后，一定会报叫
		if len(j) == 0 {
			log.Warnf("bug!!!! 房间:%v 客户端:%v 请求打牌 位置:%v ,牌:%v 不能报叫 打出去后的牌:%v ",
				self.roomId, gamePlayer.acc.AccountId, cardIndex, gamePlayer.trash_cards[len(gamePlayer.trash_cards)-1], gamePlayer.cards.String())
			return
		}
	}

}

// 玩家胡牌
func (self *deal) Old_MSGID_DGK_GAME_HU_REQ(actor int32, msg []byte, session int64) {
	if self.delay2 {
		log.Warnf("消息发早了 胡牌")
		return
	}
	pack := packet.NewPacket(msg)
	accid := pack.ReadUInt32()
	hutype := pack.ReadUInt8() // 1胡   2 请胡

	if self.deal_peng {
		log.Warnf("roomId:%v 当前是碰进来的，不能胡牌 accid:%v", self.roomId, accid)
		return
	}

	if self.bguo == 1 {
		log.Warnf("已经过了 accid:%v", accid)
		return
	}

	index := self.seatIndex(accid)
	if index == -1 {
		log.Warnf("错误胡牌 玩家不在座位上 ：%v", index)
		return
	}
	gamePlayer := self.seats[index]
	if len(self.all_hu) == 0 {
		log.Warnf("错误胡牌 玩家:%v 根本没有牌可以胡 %v", accid, index)
		return
	}

	if hutype == 1 {
		if !self.bhu {
			log.Warnf("错误胡牌 玩家:%v 不能胡:%v", accid, self.bhu)
			return
		}
		// 胡
		for i, v := range self.all_hu {
			if v.HuType == common.HU_QING_WU_DUI || v.HuType == common.HU_WU_DUI {
				continue
			}
			self.all_hu[i], self.all_hu[0] = self.all_hu[0], self.all_hu[i]
			break
		}
	} else {
		if !self.qhu {
			log.Warnf("错误胡牌 玩家:%v 不能请胡:%v", accid, self.qhu)
			return
		}
		// 胡
		for i, v := range self.all_hu {
			if v.HuType == common.HU_QING_WU_DUI || v.HuType == common.HU_WU_DUI {
				self.all_hu[i], self.all_hu[0] = self.all_hu[0], self.all_hu[i]
				break
			}
		}
	}
	h := self.all_hu[0].HuType
	if h == common.HU_NIL {
		log.Warnf("roomId:%v 错误胡牌 玩家:%v 不能胡牌 ：%v", self.roomId, accid, gamePlayer.cards.String())
		return
	}

	// 请胡判断
	if h == common.HU_QING_WU_DUI || h == common.HU_WU_DUI {
		// 先找出不一样的那张牌
		singleindex := -1
		cards := gamePlayer.cards.hand
		l := len(cards)
		for i := 0; i < l; {
			j := i + 1
			if j >= l {
				singleindex = i
				break
			} else {
				if cards[i] != cards[j] {
					singleindex = i
					break
				} else {
					i += 2
				}
			}
		}

		self.qingfu[index] = true
		// 打出去
		self.track_log(colorized.Yellow("--- 玩家:%v 请胡 牌:%v 请胡的牌位置:%v"), gamePlayer.acc.AccountId, gamePlayer.cards.String(), singleindex)
		self.push_card(int8(index), int(singleindex), 2)
		return
	}

	if self.deal_count == 0 && gamePlayer.acc.AccountId == self.seats[self.master].acc.AccountId {
		h = common.HU_TIAN
		self.all_hu[0].HuType = h
	}

	gamePlayer.hu = h
	gamePlayer.hut = 1                                                     // 自摸
	gamePlayer.huCard = gamePlayer.cards.hand[gamePlayer.cards.last_index] // 自摸
	gamePlayer.cards.hand = append(gamePlayer.cards.hand[:gamePlayer.cards.last_index], gamePlayer.cards.hand[gamePlayer.cards.last_index+1:]...)

	self.track_log(colorized.Yellow("--- 座位:%v 胡牌:%v "), index, accid)

	// 自摸胡牌积分逻辑
	total_money, packt := self.zimo(gamePlayer, self.all_hu, self.zhua_bao)

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_DGK_GAME_HU_REQ.UInt16())
	send.WriteInt8(int8(self.seatIndex(gamePlayer.acc.AccountId)) + 1)
	send.WriteInt8(1)
	send.WriteUInt8(gamePlayer.hu.Value())
	send.WriteInt8(int8(gamePlayer.huCard))
	send.WriteInt64(total_money)
	send.CatBody(packt)
	send.WriteUInt8(0)
	send.WriteUInt8(0)
	self.SendBroadcast(send.GetData())

	self.deal_player = index

	// 结算数据
	self.settle_hu_count++
	self.settle_hu.WriteInt8(int8(index + 1))
	self.settle_hu.WriteInt8(1)
	self.settle_hu.WriteInt8(int8(h.Value()))
	self.settle_hu.WriteInt8(int8(self.hu_fan[gamePlayer.hu]))
	self.settle_hu.CatBody(packt)
	self.settle_hu.WriteUInt8(0)
	self.settle_hu.WriteUInt8(0)

	// 奖金池
	c := config.GetPublicConfig_String("DGK_REWARD_POOL_RATE")
	arr_rate := utils.SplitConf2ArrInt32(c, ",")

	if int(gamePlayer.hu.Value()) >= len(arr_rate) {
		log.Errorf("奖金池错误:%v :%v", gamePlayer.hu.Value(), arr_rate)
		return
	}

	if arr_rate[int32(gamePlayer.hu)] != 0 {
		bet := uint32(self.GetParamInt(0))
		val := RoomMgr.Bonus[bet] * uint64(arr_rate[int32(gamePlayer.hu)]) / 100
		gamePlayer.acc.AddMoney(int64(val), 0, common.EOperateType_DGK_REWARD)
		gamePlayer.acc.ExtractBoun += int64(val)
		if self.clubID == 0 {
			RoomMgr.Add_bonus(bet, -val)
			RoomMgr.Add_award_hisotry(gamePlayer.acc.AccountId, gamePlayer.acc.Name, uint32(val), gamePlayer.hu.String(), bet)
		}

		self.track_log(colorized.Yellow("天胡 中奖 :%v "), val)
		// 通知中奖
		self.reward_pool_pack.WriteUInt32(uint32(gamePlayer.acc.AccountId))
		self.reward_pool_pack.WriteUInt8(uint8(self.seatIndex(gamePlayer.acc.AccountId) + 1))
		self.reward_pool_pack.WriteUInt32(uint32(val))
		self.reward_pool_pack.WriteString(gamePlayer.hu.String())
		self.reward_pool_pack_count++

		send := packet.NewPacket(nil)
		send.SetMsgID(protomsg.Old_MSGID_DGK_GAME_REWARD.UInt16())
		send.WriteUInt16(self.reward_pool_pack_count)
		send.CatBody(self.reward_pool_pack)
		self.SendBroadcast(send.GetData())
		self.animation_ = true
		self.owner.AddTimer(reward_animation_time, 1, func(dt int64) {
			self.huCount(index)
			self.assignCard(-1)
		})
	} else {
		self.huCount(index)
		self.assignCard(-1)
	}
}

// 玩家请求杠牌
func (self *deal) Old_MSGID_DGK_GAME_GANG_REQ(actor int32, msg []byte, session int64) {
	if self.delay2 {
		log.Warnf("消息发早了 杠牌")
		return
	}
	pack := packet.NewPacket(msg)
	index := pack.ReadInt8() - 1
	gangCard := pack.ReadInt8()

	if index < 0 || int(index) >= len(self.seats) {
		log.Errorf("客户端发送错误数据")
		return
	}
	gamePlayer := self.seats[index]
	if self.deal_peng {
		log.Warnf("roomId:%v 当前是碰进来的，不能杠牌 accid:%v", self.roomId, gamePlayer.acc.AccountId)
		return
	}

	if self.bguo == 1 {
		log.Warnf("已经过了 accid:%v", gamePlayer.acc.AccountId)
		return
	}

	if gamePlayer.hu != common.HU_NIL {
		log.Warnf("roomId:%v 玩家:%v 之前已经胡了:%v，不能杠", self.roomId, gamePlayer.hu, gamePlayer.cards.String())
		return
	}

	gangIndex := -1
	// 是否能杠这张牌
	all_gang := algorithm.AllGang(gamePlayer.cards.hand, gamePlayer.cards.peng)
	for _, v := range all_gang {
		if uint8(gangCard) == gamePlayer.cards.hand[v].Value() {
			gangIndex = v

			break
		}
	}

	if gangIndex == -1 {
		log.Warnf("roomId:%v 玩家:%v 牌:%v 不能杠这张牌:%v", self.roomId, gamePlayer.acc.AccountId, gamePlayer.cards.String(), gangIndex)
		return
	}

	card := gamePlayer.cards.hand[gangIndex]
	angang := []common.EMaJiangType{}
	count := 0
	for i := int8(gangIndex); i < int8(len(gamePlayer.cards.hand)); i++ {
		if gamePlayer.cards.hand[i] == card {
			count++
			angang = append(angang, card)
			if count == 4 {
				break
			}
		}
	}

	if count == 0 {
		log.Warnf("玩家:%v 打牌状态 杠的时候 牌：%v 中没有该牌:%v 位置:%v", gamePlayer.acc.AccountId, gamePlayer.cards.String(), card, gangIndex)
		return
	}

	// 暗杠
	if count == 4 {
		gangMsg := packet.NewPacket(nil)
		gangMsg.SetMsgID(protomsg.Old_MSGID_DGK_GAME_GANG_REQ.UInt16())
		gangMsg.WriteInt8(index + 1)
		gangMsg.WriteInt8(2)
		gangMsg.WriteInt8(gangCard)

		self.track_log(colorized.Yellow("--- 玩家:%v 座位:%v 当前牌:%v 请求暗杠牌:%v 位置:%v "),
			gamePlayer.acc.AccountId, index, gamePlayer.cards.String(), card, gangIndex)

		gamePlayer.cards.gang = append(gamePlayer.cards.gang, angang)
		gamePlayer.cards.hand = append(gamePlayer.cards.hand[:gangIndex], gamePlayer.cards.hand[gangIndex+4:]...)

		if gamePlayer.jiao != nil {
			gamePlayer.jiao = self.gangjiao_zimo[gangIndex]
			self.track_log(colorized.Yellow("--- 玩家:%v 暗杠 换叫 可叫:%v"), gamePlayer.acc.AccountId, gamePlayer.jiao)
		}

		gamePlayer.show_card = append(gamePlayer.show_card, Showcard{card: angang[0], t: 2})

		// 庄家的情况，特殊处理，因为庄家，首发11张牌，报叫必须打一张
		if self.master_bj {
			// 校验一下 如果庄家报叫，那么这次打牌的肯定是庄家
			if self.master != int(index) {
				log.Errorf("bug!！ 打牌的不是庄家!! index :%v gangCard:%v", index, gangCard)
				return
			}

			self.master_bj = false
			gamePlayer := self.seats[index]
			j := algorithm.Jiao_(gamePlayer.cards.hand, gamePlayer.cards.peng, gamePlayer.cards.gang)
			gamePlayer.jiao = j
			self.track_log(colorized.Yellow("--- 庄家报叫【杠牌】后，叫牌:%v"), j)

			// 帮助客户端做校验 如果庄家报叫，那么打出去的牌后，一定会报叫
			if len(j) == 0 {
				log.Errorf("bug!!!! 房间:%v 客户端:%v 请求打牌 位置:%v ,牌:%v 不能报叫 打出去后的牌:%v ",
					self.roomId, gamePlayer.acc.AccountId, gangCard, gamePlayer.trash_cards[len(gamePlayer.trash_cards)-1], gamePlayer.cards.String())
				return
			}
		}

		// 暗杠积分逻辑
		self.settle_gang_count++
		self.settle_gang.WriteInt8(2)
		self.settle_gang.WriteInt8(index + 1)
		pack := self.gang_score(gamePlayer, -1, self.settle_gang, 2) // 暗杠
		gangMsg.CatBody(pack)
		self.SendBroadcast(gangMsg.GetData())
		self.assignCard(int(index))

	} else {
		// 弯杠
		if count == 1 {
			check := false
			for _, v := range gamePlayer.cards.peng {
				if v[0] == card {
					check = true
					break
				}
			}
			if !check {
				log.Warnf("玩家:%v 打牌状态 杠的时候 不能是弯杠牌：%v 碰牌里没有该牌 牌:%v 位置:%v", gamePlayer.acc.AccountId, gamePlayer.cards.String(), card, gangIndex)
				return
			}
		} else {
			log.Warnf("玩家:%v 打牌状态 杠的时候 不能是弯杠牌：%v count:%v 牌:%v 位置:%v", gamePlayer.acc.AccountId, gamePlayer.cards.String(), count, card, gangIndex)
			return
		}
		self.track_log(colorized.Yellow("--- 玩家:%v 请求弯杠牌:%v 杠牌位置:%v "), gamePlayer.acc.AccountId, card, gangIndex)

		self.push_card(int8(self.deal_player), int(gangIndex), 1)
	}

}
