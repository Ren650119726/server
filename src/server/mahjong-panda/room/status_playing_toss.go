package room

import (
	"root/common"
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/mahjong-panda/account"
	"root/server/mahjong-panda/algorithm"
	"root/server/mahjong-panda/event"
	"root/server/mahjong-panda/send_tools"
	"sort"
)

const (
	PENG = 1
	GANG = 2
	HU   = 3
	GUO  = 4
)

type (
	option struct {
		peng int8
		gang int8
		hu   int8
	}
	toss struct {
		*playing
		s         int32
		timestamp int64          // 倒计时 时间戳 秒
		players   map[int]option // 有操作的玩家
		players_  map[int]int8   // 玩家操作

		gangjiao map[int][]algorithm.Jiao_Card // 杠叫

		send       map[int]packet.IPacket
		animation_ bool

		max_t int
	}
)

func (self *toss) Enter(now int64) {
	self.animation_ = false
	duration := config.GetPublicConfig_Int64("PANDA_AFTER_DEAL_TIME") // 持续时间 秒
	self.timestamp = utils.SecondTimeSince1970() + int64(duration)
	self.players = make(map[int]option)
	self.players_ = make(map[int]int8)
	self.gangjiao = make(map[int][]algorithm.Jiao_Card)
	self.send = make(map[int]packet.IPacket)
	self.track_log(colorized.Blue("---- toss enter duration:%v"), duration)
	lastPlayer := self.seats[self.deal_player]
	l := len(lastPlayer.trash_cards) - 1
	if l == -1 {
		log.Warnf("玩家没有废牌:%v ", lastPlayer.acc.AccountId)
		return
	}
	card := lastPlayer.trash_cards[l]
	self.track_log(colorized.Blue("---- 打牌的人:%v 位置%v 打出的牌 %v"), lastPlayer.acc.AccountId, self.deal_player, card)

	all_card_count := len(self.cards)
	for index, player := range self.seats {
		if index == self.deal_player {
			continue
		}

		if player.hu != common.HU_NIL {
			continue
		}

		if player.decide_t == int8(card/10) {
			continue
		}

		self.track_log(colorized.Blue("---- 玩家:%v 位置:%v 牌:%v"), player.acc.AccountId, index, player.cards.String())
		opt := option{}

		// 能不能碰
		if p := algorithm.CheckPeng(player.cards.hand, card); p != -1 && self.wanGang_qinHu != 2 && !player.exclude_peng[card] {
			self.track_log(colorized.Blue("---- 能 碰:%v"), p)
			opt.peng = 1
		}

		// 能不能杠
		if g := algorithm.CheckGang(player.cards.hand, card); g != -1 && (all_card_count > 0) && self.wanGang_qinHu != 2 {
			opt.gang = 1
		}

		// 能不能胡
		t, h := self.point_pao_check(lastPlayer, card, player)
		if et := player.exclude_hu; (et <= 0 || t > uint8(et)) && h > 0 {
			if len(self.seats) == 3 || t > 0 {
				self.track_log(colorized.Blue("---- 能 胡:%v"), h)
				opt.hu = 1
				self.max_t = int(t)

				// 如果手牌里还有定缺的牌，不能胡
				for _, card := range player.cards.hand {
					if int8(card/10) == player.decide_t {
						opt.hu = 0
						break
					}
				}
			}
		}

		if opt.peng != 0 || opt.gang != 0 || opt.hu != 0 {
			self.players[index] = opt

			self.dispatcher.Dispatch(&event.EnterToss{
				Index: index,
				Bhu:   opt.hu == 1,
				Peng:  opt.peng == 1,
				Gangs: opt.gang == 1,
			}, event.EventType_Toss)
		}

		send := packet.NewPacket(nil)
		send.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_OPTION.UInt16())
		send.WriteInt64(self.timestamp * 1000)
		send.WriteInt8(opt.peng)
		send.WriteInt8(opt.gang)
		send.WriteInt8(opt.hu)
		self.send[index] = send
		send_tools.Send2Account(send.GetData(), player.acc.SessionId)
	}
}

func (self *toss) Tick(now int64) {
	if self.animation_ {
		return
	}
	if now >= self.timestamp {
		// 如果有玩家没操作，通通都过
		for index := range self.players {
			self.trusateeship(uint8(index))
			data := packet.NewPacket(nil)
			data.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_GUO_REQ.UInt16())
			data.WriteInt8(int8(index) + 1)
			core.CoreSend(self.owner.Id, self.owner.Id, data.GetData(), 0)
		}
	}

	for index, op := range self.players {
		p := self.seats[index]
		if p.trusteeship == 1 {
			if op.hu == 1 {
				auto_send := packet.NewPacket(nil)
				auto_send.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_HU_REQ.UInt16())
				auto_send.WriteUInt32(p.acc.AccountId)
				core.CoreSend(self.owner.Id, self.owner.Id, auto_send.GetData(), 0)
			} else {

				auto_send := packet.NewPacket(nil)
				auto_send.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_GUO_REQ.UInt16())
				auto_send.WriteUInt8(uint8(index + 1))
				core.CoreSend(self.owner.Id, self.owner.Id, auto_send.GetData(), 0)
			}
		}

	}
	self.checklogic()
}

func (self *toss) checklogic() {
	if len(self.players) == 0 {
		// 处理玩家操作
		self.optdeal()

	}
}

func (self *toss) point_pao_check(fangpao_er *GamePlayer, pao common.EMaJiangType, hu_er *GamePlayer) (uint8, common.EMaJiangHu) {
	//all_hu, zhua_bao, zq, _ := self.calcExtra(hu_er, pao, true)
	all_hu := self.calcExtra(hu_er, pao, 0, true)
	if len(all_hu) == 0 {
		//log.Errorf("放炮的时候，胡的人，数据错误:%v pao:%v 牌:%v", all_hu, pao, hu_er.cards.String())
		return 0, 0
	}
	so := &Majiang_fan_Sort{All: all_hu, Room: self.Room}
	sort.Sort(so)
	all_hu = so.All
	if self.push_count == 1 && self.deal_player == self.master {
		all_hu[0].HuType = common.HU_DI
	}
	hu := all_hu[0].HuType

	bet := self.GetParamInt(0)
	total_fan := uint8(self.hu_fan[hu])

	if all_hu[0].Extra != nil {
		for t, v := range all_hu[0].Extra {
			if int(t) >= len(self.extra_fan) {
				log.Errorf("数组越界:%v len:%v", t, len(self.extra_fan))
				return 0, 0
			}
			total_fan += uint8(self.extra_fan[int32(t)] * int32(v))
		}
	}

	//if zq > 0 {
	//	total_fan += zq - 1
	//}

	fan := total_fan
	if total_fan > MAX_FAN {
		fan = MAX_FAN
	}
	rate := FAN_RATIO[fan]
	total_score := int64(rate * bet)
	if fangpao_er.acc.GetMoney() < uint64(total_score) {
		log.Errorf("玩家:%v身上的钱:%v 不够赔:%v", fangpao_er.acc.AccountId, fangpao_er.acc.GetMoney(), total_score)
		total_score = int64(fangpao_er.acc.GetMoney())
	}
	return total_fan, hu
}

func (self *toss) point_pao(fangpao_er *GamePlayer, pao common.EMaJiangType, hu_er *GamePlayer, mutilp bool, rew packet.IPacket) {
	//all_hu, zhua_bao, zq, zqi := self.calcExtra(hu_er, pao, true)
	all_hu := self.calcExtra(hu_er, pao, 0, true)
	if len(all_hu) == 0 {
		log.Errorf("放炮的时候，胡的人，数据错误:%v pao:%v 牌:%v", all_hu, pao, hu_er.cards.String())
		return
	}

	packt := packet.NewPacket(nil)
	so := &Majiang_fan_Sort{All: all_hu, Room: self.Room}
	sort.Sort(so)
	all_hu = so.All
	if self.push_count == 1 && self.deal_player == self.master {
		all_hu[0].HuType = common.HU_DI
	}
	/////// 中奖  ///////////////////////////////////////////////////////////////////////////////////////////////////////
	c := config.GetPublicConfig_String("PANDA_REWARD_POOL_RATE")
	arr_rate := utils.SplitConf2ArrInt32(c, ",")

	if int(all_hu[0].HuType.Value()) >= len(arr_rate) {
		log.Errorf("奖金池错误:%v :%v", all_hu[0].HuType.Value(), arr_rate)
		return
	}

	if arr_rate[int32(all_hu[0].HuType)] != 0 {
		hu_str := all_hu[0].HuType.String()
		t := uint32(self.GetParamInt(0))
		val := RoomMgr.Bonus[t] * uint64(arr_rate[int32(all_hu[0].HuType)]) / 100
		hu_er.acc.AddMoney(int64(val), 0, common.EOperateType_PANDA_REWARD)
		hu_er.acc.ExtractBoun += int64(val)
		if self.clubID == 0 {
			RoomMgr.Add_bonus(t, -val)
			RoomMgr.Add_award_hisotry(hu_er.acc.AccountId, hu_er.acc.Name, uint32(val), hu_str, t)
		}

		self.reward_pool_pack.WriteUInt32(uint32(hu_er.acc.AccountId))
		self.reward_pool_pack.WriteUInt8(uint8(self.seatIndex(hu_er.acc.AccountId) + 1))
		self.reward_pool_pack.WriteUInt32(uint32(val))
		self.reward_pool_pack.WriteString(hu_str)

		rew.WriteUInt32(uint32(hu_er.acc.AccountId))
		rew.WriteUInt8(uint8(self.seatIndex(hu_er.acc.AccountId) + 1))
		rew.WriteUInt32(uint32(val))
		rew.WriteString(hu_str)
		self.reward_pool_pack_count++
	}
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	hu_er.hu = all_hu[0].HuType
	hu_er.hut = 2      // 点炮
	hu_er.huCard = pao // 点炮

	bet := self.GetParamInt(0)
	total_fan := uint8(self.hu_fan[hu_er.hu])

	if all_hu[0].Extra != nil {
		packt.WriteUInt16(uint16(len(all_hu[0].Extra)))
		for t, v := range all_hu[0].Extra {
			packt.WriteInt8(int8(t))
			fan := uint8(self.extra_fan[int32(t)] * int32(v))
			packt.WriteUInt8(fan)
			total_fan += fan
		}
	} else {
		packt.WriteUInt16(0)
	}

	//z := uint8(0)
	//if zq > 0 {
	//	zq -= 1
	//	total_fan += zq
	//}

	fan := total_fan
	if total_fan > MAX_FAN {
		fan = MAX_FAN
	}
	rate := FAN_RATIO[fan]
	total_score := int64(rate * bet)
	if fangpao_er.acc.GetMoney() < uint64(total_score) {
		log.Errorf("玩家:%v身上的钱:%v 不够赔:%v", fangpao_er.acc.AccountId, fangpao_er.acc.GetMoney(), total_score)
		total_score = int64(fangpao_er.acc.GetMoney())
	}
	fangpao_er.acc.AddMoney(-total_score, 0, common.EOperateType_PANDA_HU)
	hu_er.acc.AddMoney(total_score, 0, common.EOperateType_PANDA_HU)

	// 如果有杠上炮，转雨//////////////////////////////////////////////////////////////////////////////////////
	temp := packet.NewPacket(nil)
	for t, c := range all_hu[0].Extra {
		if t == common.EXTRA_GANGSHANGPAO && c == 1 && !mutilp {
			self.track_log(colorized.Blue("---- 杠上炮触发 雨水钱:%v"), fangpao_er.gang_score_z)
			total := int64(0)
			for i, v := range fangpao_er.gang_score_z {
				fangpao_er.gang_score[i] -= v
				total += v
			}
			fangpao_er.acc.AddMoney(-total, 0, common.EOperateType_PANDA_GANG)
			hu_er.acc.AddMoney(total, 0, common.EOperateType_PANDA_GANG)

			fangpao_er.gang_score_z = map[int]int64{}
			temp.WriteInt8(int8(self.seatIndex(fangpao_er.acc.AccountId)) + 1)
			temp.WriteInt8(int8(self.seatIndex(hu_er.acc.AccountId)) + 1)
			temp.WriteInt64(total)
			self.settle_zy_count++
			break
		}
	}

	self.settle_zy.CatBody(temp)

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_HU_REQ.UInt16())
	send.WriteInt8(int8(self.seatIndex(hu_er.acc.AccountId)) + 1)
	send.WriteInt8(2)
	send.WriteUInt8(hu_er.hu.Value())
	send.WriteInt8(int8(pao))
	send.WriteInt64(total_score)
	l := len(all_hu[0].Extra)
	send.WriteUInt16(uint16(l))
	for t, v := range all_hu[0].Extra {
		send.WriteInt8(int8(t))
		send.WriteUInt8(uint8(self.extra_fan[int32(t)] * int32(v)))
	}

	send.WriteUInt16(1)
	//send.WriteUInt8(z)
	send.WriteUInt8(uint8(self.seatIndex(fangpao_er.acc.AccountId) + 1))
	send.WriteInt64(total_score)
	//send.WriteUInt8(zq)
	//send.WriteUInt8(uint8(zqi + 1))

	self.SendBroadcast(send.GetData())

	self.track_log(colorized.Blue("---- 放炮的人:%v 玩家:%v 胡牌:%v 总番:%v 额外番:%v,  座位:%v 赔付总金额:%v"),
		fangpao_er.acc.AccountId, hu_er.acc.AccountId, hu_er.hu, total_fan, all_hu[0].Extra, self.deal_player, total_score)

	packt.WriteUInt16(1)
	//packt.WriteUInt8(uint8(z))
	packt.WriteInt8(int8(self.seatIndex(fangpao_er.acc.AccountId) + 1))
	packt.WriteInt64(total_score)

	// 结算数据
	self.settle_hu_count++
	self.settle_hu.WriteInt8(int8(self.seatIndex(hu_er.acc.AccountId) + 1))
	self.settle_hu.WriteInt8(2)
	self.settle_hu.WriteInt8(int8(hu_er.hu))
	self.settle_hu.WriteInt8(int8(self.hu_fan[hu_er.hu]))
	self.settle_hu.CatBody(packt)

	//self.settle_hu.WriteUInt8(zq)
	//self.settle_hu.WriteUInt8(uint8(zqi + 1))
}
func (self *toss) optdeal() {
	// 优先找出胡的所有人
	all_hu := []int{}
	gang_peng := int(-1)
	for index, opt := range self.players_ {
		if opt == PENG || opt == GANG {
			gang_peng = index
		} else if opt == HU {
			all_hu = append(all_hu, index)
		}
	}

	lastPlayer := self.seats[self.deal_player]
	l := len(lastPlayer.trash_cards)
	card := lastPlayer.trash_cards[l-1]
	lhu := len(all_hu)

	if lhu > 1 {
		multip_ := packet.NewPacket(nil)
		multip_.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_MULTIP_HU.UInt16())
		multip_.WriteUInt8(uint8(card))
		multip_.WriteUInt16(uint16(len(all_hu)))
		for _, v := range all_hu {
			multip_.WriteUInt8(uint8(v + 1))
		}
		self.multip = multip_
		self.SendBroadcast(multip_.GetData())

		self.settle_ty = packet.NewPacket(nil)
		self.settle_ty.WriteInt8(int8(self.deal_player) + 1)             // 杠的人下标
		self.settle_ty.WriteUInt16(uint16(len(lastPlayer.gang_score_z))) // 数量
		for i, v := range lastPlayer.gang_score_z {
			self.settle_ty.WriteInt8(int8(i + 1)) // 退雨的人
			self.settle_ty.WriteInt64(v)          // 退雨的钱

			tp := self.seats[i]
			tp.acc.AddMoney(v, 0, common.EOperateType_PANDA_GANG)
			lastPlayer.acc.AddMoney(-v, 0, common.EOperateType_PANDA_GANG)
		}

	}
	// 没有胡的人，就碰\杠 否则就胡
	if lhu != 0 {
		temp := packet.NewPacket(nil)
		count := uint16(0)
		for _, v := range all_hu {
			rew := packet.NewPacket(nil)
			self.point_pao(lastPlayer, card, self.seats[v], lhu > 1, rew)
			self.huCount(v)

			if rew.GetDataSize() > packet.PACKET_HEAD_LEN {
				count++
				temp.CatBody(rew)
			}
		}

		if count != 0 {
			send := packet.NewPacket(nil)
			send.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_REWARD.UInt16())
			send.WriteUInt16(count)
			send.CatBody(temp)
			self.SendBroadcast(send.GetData())

			self.animation_ = true
			self.owner.AddTimer(reward_animation_time, 1, func(dt int64) {
				if lhu > 1 {
					self.next_master = self.deal_player
				}
				self.deal_player = all_hu[0]
				lastPlayer.trash_cards = lastPlayer.trash_cards[:l-1]
				self.assignCard(-1)
			})
		} else {
			if lhu > 1 {
				self.next_master = self.deal_player
			}
			self.deal_player = all_hu[0]
			lastPlayer.trash_cards = lastPlayer.trash_cards[:l-1]
			self.assignCard(-1)
		}

	} else if gang_peng != -1 {
		gamePlayer := self.seats[gang_peng]
		opt_man := self.players_[gang_peng]
		lastPlayer.trash_cards = lastPlayer.trash_cards[:l-1]
		if opt_man == PENG { // 碰操作!!!!!!!!!!!!!!!!!
			self.track_log(colorized.Blue("---- 碰成功了! 玩家:%v 碰 :%v 的牌:%v"), gang_peng, self.deal_player, card)
			sendpeng := packet.NewPacket(nil)
			sendpeng.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_PENG_REQ.UInt16())
			sendpeng.WriteInt8(int8(gang_peng + 1))
			sendpeng.WriteInt8(int8(self.deal_player + 1))
			self.SendBroadcast(sendpeng.GetData())
			// 找出要碰出去的下标
			index := -1
			for i, v := range gamePlayer.cards.hand {
				if v == card {
					index = i
					break
				}
			}
			gamePlayer.cards.hand = append(gamePlayer.cards.hand[:index], gamePlayer.cards.hand[index+2:]...)
			gamePlayer.cards.peng = append(gamePlayer.cards.peng, []common.EMaJiangType{card, card, card})
			gamePlayer.show_card = append(gamePlayer.show_card, Showcard{card: card, t: 4})
			self.deal_player = gang_peng
			self.deal_peng = true

			card := int8(-1)
			max_card := common.EMaJiangType(0)
			lastIndex := 0

			for k, v := range gamePlayer.cards.hand {
				if int8(v.Value()/10) == gamePlayer.decide_t {
					card = int8(v.Value())
				} else if card != -1 {
					break
				}

				if k == 0 {
					max_card = v
					lastIndex = k
				} else if v > max_card {
					max_card = v
					lastIndex = k
				}

			}
			gamePlayer.cards.last_index = lastIndex

			self.game_state.Swtich(0, DEAL_STATE)
		} else if opt_man == GANG { // 杠操作!!!!!!!!!!!!!!!!!
			self.track_log(colorized.Blue("---- 杠成功了! 玩家:%v 杠 :%v 的牌:%v"), gang_peng, self.deal_player, card)
			sendgang := packet.NewPacket(nil)
			sendgang.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_GANG_REQ.UInt16())
			sendgang.WriteInt8(int8(gang_peng + 1))
			sendgang.WriteInt8(int8(1))
			sendgang.WriteInt8(int8(card))

			// 找出要杠出去的下标
			index := -1
			for i, v := range gamePlayer.cards.hand {
				if v == card {
					index = i
					break
				}
			}
			// 只可能是直杠
			gamePlayer.cards.hand = append(gamePlayer.cards.hand[:index], gamePlayer.cards.hand[index+3:]...)
			gamePlayer.cards.gang = append(gamePlayer.cards.gang, []common.EMaJiangType{card, card, card, card})
			gamePlayer.show_card = append(gamePlayer.show_card, Showcard{card: card, t: 1})
			gamePlayer.cards.last_index = int(len(gamePlayer.cards.hand) - 1)
			// 如果报叫了，执行换叫
			//if gamePlayer.jiao != nil {
			//	gamePlayer.jiao = self.gangjiao[gang_peng]
			//}

			self.settle_gang_count++
			self.settle_gang.WriteInt8(1)
			self.settle_gang.WriteInt8(int8(gang_peng) + 1)
			pack := self.gang_score(gamePlayer, self.deal_player, self.settle_gang, 2) // 直杠
			sendgang.CatBody(pack)
			self.SendBroadcast(sendgang.GetData())

			self.deal_player = gang_peng
			self.assignCard(gang_peng) // 发牌
		} else {
			log.Errorf("错啦！:%v", opt_man)
		}
	} else if self.wanGang_qinHu != 0 {
		// 发牌给下一个玩家
		self.wangang_qinghu_check()
	} else {
		self.assignCard(-1) // 没有人有操作，直接发牌
	}
}

func (self *toss) Combine_Game_MSG(packet packet.IPacket, acc *account.Account) {
	index := self.seatIndex(acc.AccountId)
	s := self.send[index]
	if s != nil {
		packet.CatBody(s)
	} else {
		packet.WriteInt64(0)
		packet.WriteInt8(0)
		packet.WriteInt8(0)
		packet.WriteInt8(0)
	}
}

func (self *toss) Leave(now int64) {

	self.track_log(colorized.Blue("---- toss leave\n"))
}

// 所有人都操作完了，判断一次
func (self *toss) wangang_qinghu_check() {
	gamePlayer := self.seats[self.deal_player]
	l := len(gamePlayer.trash_cards)
	card := gamePlayer.trash_cards[l-1]
	gamePlayer.trash_cards = gamePlayer.trash_cards[:l-1]
	// 弯杠
	if self.wanGang_qinHu == 1 {
		gangMsg := packet.NewPacket(nil)
		gangMsg.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_GANG_REQ.UInt16())
		gangMsg.WriteInt8(int8(self.deal_player) + 1)
		gangMsg.WriteInt8(3)
		gangMsg.WriteInt8(int8(card))

		self.track_log(colorized.Blue("---- 玩家:%v 座位:%v 当前牌:%v 触发弯杠牌:%v"), gamePlayer.acc.AccountId, self.deal_player, gamePlayer.cards.String(), card)

		gamePlayer.cards.gang = append(gamePlayer.cards.gang, []common.EMaJiangType{card, card, card, card})
		for index, v := range gamePlayer.cards.peng {
			if v[0] == card {
				gamePlayer.cards.peng = append(gamePlayer.cards.peng[:index], gamePlayer.cards.peng[index+1:]...)
				break
			}
		}

		for index, v := range gamePlayer.show_card {
			if v.card == card && v.t == 4 {
				gamePlayer.show_card[index].t = 3
				break
			}
		}

		self.settle_gang_count++
		self.settle_gang.WriteInt8(3)
		self.settle_gang.WriteInt8(int8(self.deal_player) + 1)
		pack := self.gang_score(gamePlayer, -1, self.settle_gang, 1) // 弯杠
		gangMsg.CatBody(pack)
		self.SendBroadcast(gangMsg.GetData())
		// 为杠的人分牌
		self.assignCard(self.deal_player)
	} else if self.wanGang_qinHu == 2 { // 请胡
		log.Error("熊猫麻将，走到请胡来了！！！！！！！！！！！！！！！！！！！！！！！")
		return
		//if len(self.all_hu) == 0 {
		//	log.Errorf("请胡错误!! hu:%v 玩家:%v", self.all_hu, gamePlayer.acc.AccountId)
		//	return
		//}
		//
		//if self.deal_count == 0 && gamePlayer.acc.AccountId == self.seats[self.master].acc.AccountId {
		//	self.all_hu[0].HuType = common.HU_TIAN
		//}
		//gamePlayer.hu = self.all_hu[0].HuType
		//gamePlayer.hut = 1       // 请胡
		//gamePlayer.huCard = card // 请胡
		//
		//self.track_log(colorized.Blue("---- 请胡成功! 胡的人:%v 胡牌:%v"), gamePlayer.acc.AccountId, self.all_hu)
		//total_money, packt := self.zimo(gamePlayer, self.all_hu, self.zhua_bao)
		//
		//send := packet.NewPacket(nil)
		//send.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_HU_REQ.UInt16())
		//send.WriteInt8(int8(self.seatIndex(gamePlayer.acc.AccountId)) + 1)
		//send.WriteInt8(1)
		//send.WriteUInt8(gamePlayer.hu.Value())
		//send.WriteInt8(int8(card.Value()))
		//send.WriteInt64(total_money)
		//send.CatBody(packt)
		//send.WriteUInt8(0)
		//send.WriteUInt8(0)
		//self.SendBroadcast(send.GetData())
		//// 为下一个人分牌
		//
		//// 结算数据
		//self.settle_hu_count++
		//self.settle_hu.WriteInt8(int8(self.seatIndex(gamePlayer.acc.AccountId) + 1))
		//self.settle_hu.WriteInt8(1)
		//self.settle_hu.WriteInt8(int8(gamePlayer.hu.Value()))
		//self.settle_hu.WriteInt8(int8(self.hu_fan[gamePlayer.hu]))
		//self.settle_hu.CatBody(packt)
		//self.settle_hu.WriteUInt8(0)
		//self.settle_hu.WriteUInt8(0)
		//
		//// 奖金池
		//c := config.GetPublicConfig_String("PANDA_REWARD_POOL_RATE")
		//arr_rate := utils.SplitConf2ArrInt32(c, ",")
		//
		//if int(gamePlayer.hu.Value()) >= len(arr_rate) {
		//	log.Error("奖金池错误:%v :%v", gamePlayer.hu.Value(), arr_rate)
		//	return
		//}
		//
		//if arr_rate[int32(gamePlayer.hu)] != 0 {
		//	t := uint32(self.GetParamInt(0))
		//	val := RoomMgr.Bonus[t] * uint64(arr_rate[int32(gamePlayer.hu)]) / 100
		//	gamePlayer.acc.AddMoney(int64(val), 0, common.EOperateType_PANDA_REWARD)
		//	RoomMgr.Add_bonus(t, -val)
		//	RoomMgr.Add_award_hisotry(gamePlayer.acc.AccountId, gamePlayer.acc.Name, uint32(val), gamePlayer.hu.String(), t)
		//
		//	self.track_log(colorized.Yellow("天胡 中奖 :%v "), val)
		//	// 通知中奖
		//	self.reward_pool_pack.WriteUInt32(uint32(gamePlayer.acc.AccountId))
		//	self.reward_pool_pack.WriteUInt8(uint8(self.seatIndex(gamePlayer.acc.AccountId) + 1))
		//	self.reward_pool_pack.WriteUInt32(uint32(val))
		//	self.reward_pool_pack.WriteString(gamePlayer.hu.String())
		//	self.reward_pool_pack_count++
		//
		//	send := packet.NewPacket(nil)
		//	send.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_REWARD.UInt16())
		//	send.WriteUInt16(self.reward_pool_pack_count)
		//	send.CatBody(self.reward_pool_pack)
		//	self.SendBroadcast(send.GetData())
		//	self.animation_ = true
		//	self.owner.AddTimer(reward_animation_time, 1, func(dt int64) {
		//		self.huCount(self.deal_player)
		//		self.assignCard(-1)
		//	})
		//} else {
		//	self.huCount(self.deal_player)
		//	self.assignCard(-1)
		//}
	} else {
		log.Warnf("出错:%v", self.wanGang_qinHu)
	}
}

///////////////////////////////// handler ///////////////////////////////////////////////////
func (self *toss) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_PANDA_GAME_GUO_REQ.UInt16(): // 过
		self.Old_MSGID_PANDA_GAME_GUO_REQ(actor, msg, session)
	case protomsg.Old_MSGID_PANDA_GAME_PENG_REQ.UInt16(): // 碰
		self.Old_MSGID_PANDA_GAME_PENG_REQ(actor, msg, session)
	case protomsg.Old_MSGID_PANDA_GAME_GANG_REQ.UInt16(): // 杠
		self.Old_MSGID_PANDA_GAME_GANG_REQ(actor, msg, session)
	case protomsg.Old_MSGID_PANDA_GAME_HU_REQ.UInt16(): // 胡
		self.Old_MSGID_PANDA_GAME_HU_REQ(actor, msg, session)
	default:
		log.Warnf("toss 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}
	return true
}

// 过
func (self *toss) Old_MSGID_PANDA_GAME_GUO_REQ(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	index := int(pack.ReadInt8()) - 1
	if index == -1 {
		log.Warnf("出错！%v", index)
		return
	}

	if _, exit := self.players[index]; exit == false {
		return
	}

	gamePlayer := self.seats[index]
	if self.players[index].hu == 1 {

		gamePlayer.exclude_hu = int(self.max_t)
	}
	if self.players[index].peng == 1 {
		lastPlayer := self.seats[self.deal_player]
		len := len(lastPlayer.trash_cards) - 1
		card := lastPlayer.trash_cards[len]
		gamePlayer.exclude_peng[card] = true
	}
	delete(self.players, index)
	self.players_[index] = GUO

	self.track_log(colorized.Blue("---- 座位:%v 选择【过】"), index)
}

// 碰
func (self *toss) Old_MSGID_PANDA_GAME_PENG_REQ(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	index := int(pack.ReadInt8()) - 1
	if index == -1 {
		log.Warnf("出错！%v", index)
		return
	}

	if _, exit := self.players[index]; exit == false {
		return
	}

	if self.wanGang_qinHu == 2 {
		// 别人请胡了，不能杠
		log.Errorf("客户端校验一下，别人请胡，不能杠")
		return
	}

	lastPlayer := self.seats[self.deal_player]
	l := len(lastPlayer.trash_cards) - 1
	card := lastPlayer.trash_cards[l]

	gamePlayer := self.seats[index]
	if self.players[index].peng != 1 {
		log.Warnf("玩家%v 牌%v 单牌 %v 不能的碰", gamePlayer.acc.AccountId, gamePlayer.cards.String(), card)
		return
	}

	if gamePlayer.exclude_peng[card] {
		log.Warnf("玩家%v 牌%v 单牌 %v 不能的碰 ,被禁止了:%v", gamePlayer.acc.AccountId, gamePlayer.cards.String(), card, gamePlayer.exclude_peng)
		return
	}
	delete(self.players, index)
	self.players_[index] = PENG
	self.track_log(colorized.Blue("---- 座位:%v 选择【碰】"), index)
}

// 杠
func (self *toss) Old_MSGID_PANDA_GAME_GANG_REQ(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	index := int(pack.ReadInt8()) - 1
	if index == -1 {
		log.Warnf("出错！%v", index)
		return
	}

	if _, exit := self.players[index]; exit == false {
		return
	}

	if self.wanGang_qinHu == 2 {
		// 别人请胡了，不能杠
		log.Errorf("客户端校验一下，别人请胡，不能杠")
		return
	}

	if index < 0 || index >= len(self.seats) {
		log.Errorf("-----")
		return
	}

	lastPlayer := self.seats[self.deal_player]
	len := len(lastPlayer.trash_cards) - 1
	card := lastPlayer.trash_cards[len]

	gamePlayer := self.seats[index]
	if self.players[index].gang != 1 {
		log.Warnf("玩家%v 牌%v 单牌 %v 不能杠", gamePlayer.acc.AccountId, gamePlayer.cards.String(), card)
		return
	}
	delete(self.players, index)
	self.players_[index] = GANG

	self.track_log(colorized.Blue("---- 座位:%v 选择【杠】"), index)
}

// 胡
func (self *toss) Old_MSGID_PANDA_GAME_HU_REQ(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accid := pack.ReadUInt32()
	index := self.seatIndex(accid)
	if index == -1 {
		log.Warnf("出错！%v", index)
		return
	}

	if _, exit := self.players[index]; exit == false {
		return
	}

	lastPlayer := self.seats[self.deal_player]
	l := len(lastPlayer.trash_cards) - 1
	card := lastPlayer.trash_cards[l]

	if index >= len(self.seats) {
		log.Warnf("出错!!!！%v  len:%v", index, len(self.seats))
		return
	}
	gamePlayer := self.seats[index]
	if self.players[index].hu != 1 {
		log.Warnf("玩家%v 牌 %v 单牌 %v 不能胡", gamePlayer.acc.AccountId, gamePlayer.cards.String(), card)
		return
	}

	delete(self.players, index)
	self.players_[index] = HU

	// 检测一下，还有没有人可以胡的没有胡
	hu := false
	for _, option := range self.players {
		if option.hu == 1 {
			hu = true
			break
		}
	}

	// 没有可以胡的人了，还有操作的人，也可以不用操作了
	if !hu {
		self.players = make(map[int]option)
	}

	self.track_log(colorized.Blue("---- 玩家:%v 座位:%v 选择【胡】"), accid, index)
}
