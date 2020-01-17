package room

import (
	"root/common"
	ca "root/common/algorithm"
	"root/common/config"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"fmt"
	"root/protomsg"
	"root/server/mahjong-panda/account"
	"root/server/mahjong-panda/algorithm"
	"root/server/mahjong-panda/send_tools"
	"root/server/mahjong-panda/types"
	"sort"
)

const (
	CHANGE_STATE  = 1 // 换牌状态
	DECIDE_STATE  = 2 // 定缺状态
	DEAL_STATE    = 3 // 打牌状态
	BREAKIN_STATE = 4 // 断牌状态

)

const DKG_START_TIME = 1000

type (
	playing struct {
		*Room
		s     types.ERoomStatus
		cards []common.EMaJiangType

		dice          []uint8 // 骰子
		deal_player   int     // 当前打牌的玩家
		deal_count    int     // 发牌次数
		push_count    int     // 打牌次数
		deal_peng     bool    // 是否碰 碰的时候，至为true
		game_state    *utils.FSM
		wanGang_qinHu int //  1 弯杠 2 请胡用于判断弯杠,请胡 抢杠标记

		last_gang int // 最后一个杠的人 下次分牌就清除
		hu_count  int

		// 自摸用
		all_hu []*ca.Majiang_Hu

		// 显示数据需要
		last_push_index     int8 // 最后一个打牌的人位置
		last_push_cardIndex int8 // 最后一个打牌的人牌型 0 普通牌、1 弯杠、2 请胡

		//qingfu map[int]bool // 自动请服

		multip packet.IPacket
	}

	Majiang_fan_Sort struct {
		*Room
		All []*ca.Majiang_Hu // 所有的胡牌类型  和 额外加番
	}
)

func (self *playing) initData_() {
	servicepack := packet.NewPacket(nil)
	playerCount := uint16(0)
	reward_ := 0
	// 所有人，都进入游戏状态
	conf_str := config.GetPublicConfig_String("PANDA_FEE_RATIO")
	conf_val := utils.SplitConf2Arr_ArrInt64(conf_str)
	fee_ratio := 0
	for _, v := range conf_val {
		if int64(self.GetParamInt(0)) <= v[0] {
			fee_ratio = int(v[1])
			break
		}
	}

	tax_ratio := config.GetPublicConfig_Int64("TAX")
	reward_ratio := config.GetPublicConfig_Int64("PANDA_REWARD_RATIO")
	self.track_log(colorized.Green("fee_ratio:%v tax_ratio:%v reward_ratio:%v "), fee_ratio, tax_ratio, reward_ratio)

	for index, player := range self.seats {
		// 玩家抽水
		fee := fee_ratio
		player.acc.Fee += int64(fee)
		player.acc.AddMoney(-int64(fee), 0, common.EOperateType_PANDA_FEE)

		service_ := fee * int(tax_ratio) / 100
		val := service_ * int(reward_ratio) / 100
		reward_ += val
		service_ = service_ - val

		playerCount++
		servicepack.WriteUInt32(uint32(player.acc.AccountId))
		servicepack.WriteUInt32(uint32(service_))

		change := int64(player.money_after) - int64(player.money_before)
		change -= int64(fee)

		// 初始化数据
		player.status = types.EGameStatus_PLAYING
		player.show_card = []Showcard{}
		player.money_before = player.acc.GetMoney()
		player.money_after = player.acc.GetMoney()

		//player.jiao = nil
		player.trash_cards = make([]common.EMaJiangType, 0, 0)
		player.gang_score = make(map[int]int64)
		player.gang_score_z = make(map[int]int64)
		player.exclude_hu = 0
		player.card_time = 0
		player.exclude_peng = make(map[common.EMaJiangType]bool)
		player.acc.Games++

		up_games := packet.NewPacket(nil)
		up_games.SetMsgID(protomsg.Old_MSGID_PANDA_UPDATE_GAME_TIME.UInt16())
		up_games.WriteUInt32(uint32(player.acc.Games))
		send_tools.Send2Account(up_games.GetData(), player.acc.SessionId)
		self.track_log(colorized.Green("玩家:%v 座位:%v 抽水:%v 服务费:%v "), player.acc.AccountId, index, fee, service_)
	}

	if playerCount > 0 {
		ser_fee := packet.NewPacket(nil)
		ser_fee.SetMsgID(protomsg.Old_MSGID_UPDATE_SERVICE_FEE.UInt16())
		ser_fee.WriteUInt8(uint8(self.gameType))
		ser_fee.WriteUInt32(uint32(self.roomId))
		ser_fee.WriteUInt16(playerCount)
		ser_fee.CatBody(servicepack)
		send_tools.Send2Hall(ser_fee.GetData())
	}

	// 奖金池增加 reward_
	if self.clubID == 0 {
		RoomMgr.Add_bonus(uint32(self.GetParamInt(0)), uint64(reward_))
	}

	self.track_log(colorized.Green("-------------------------------------------------------"))
}
func (self *playing) Enter(now int64) {
	self.cards = algorithm.GetRandom_Card(72)
	self.last_gang = -1

	isTestServer, _, _ := config.IsTestServer()
	if isTestServer == true {
		//self.cards = []common.EMaJiangType{28, 34, 23, 37, 15, 32, 27, 13, 32, 24, 14, 15, 37, 21, 19, 25, 25, 28, 36, 26, 15, 11, 33, 32, 24, 12, 12, 23, 36, 28, 25, 18, 16, 13, 37, 38, 31, 13, 37, 23, 12, 11, 28, 39, 29, 17, 26, 18, 35, 27, 16, 32, 17, 31, 11, 21, 33, 19, 31, 13, 22, 35, 34, 14, 18, 38, 36, 18, 19, 27, 21, 24}
		//
		//self.cards = append(self.cards,
		//	[]common.EMaJiangType{24, 24, 25, 25, 26, 26, 27, 21, 13, 22, 23, 23, 15, 23, 23, 22, 14, 22, 27, 17, 17, 28, 16, 14, 28, 18, 22, 16, 15, 15, 18, 27, 12, 11, 28, 14, 28, 13, 27, 13, 18}...)

		//if self.GetParamInt(0) == 1000 {
		//	//	////////////////////////////// test ////////////////////////////////////////////////
		//	self.cards = []common.EMaJiangType{12, 13, 14, 14, 14, 15, 15, 15, 16, 16, 17, 17, 17, 11, 11, 11, 11, 12, 13, 14, 15, 16, 17, 18, 19, 19, 19}
		//
		//	self.cards = append(self.cards,
		//		[]common.EMaJiangType{12, 13, 22, 22, 23, 24, 25, 25, 25, 26, 27, 28, 28, 29, 13, 25, 32, 34, 34, 34, 34, 24, 31, 27, 21, 31, 21, 23, 19, 22, 24, 37, 29, 22, 36, 31, 38, 39, 37, 35, 32, 12, 37, 33, 16, 26, 24, 26, 38, 38, 27, 21, 28, 32, 36, 23, 26, 33, 37, 21, 39, 18, 35, 39, 32, 28, 35, 33, 27, 31, 35, 18, 36, 29, 23, 33, 29, 18, 36, 39, 38}...)
		//	//	////////////////////////////////////////////////////////////////////////////////////
		//}
		//	self.cards = append(self.cards,
		//		[]common.EMaJiangType{17, 11, 11, 14, 27, 15, 15, 15, 17, 21, 11, 28, 21, 12, 14, 17, 27, 27, 12, 15, 14, 25, 27, 12, 18, 21, 19, 21, 17, 12, 22, 29, 28, 11, 16, 19, 28, 13, 14, 28, 24}...)
		//	//	////////////////////////////////////////////////////////////////////////////////////
		//} else if self.GetParamInt(0) == 1000 {
		//	//	////////////////////////////// test ////////////////////////////////////////////////
		//	self.cards = []common.EMaJiangType{23, 14, 27, 16, 23, 13, 27, 21, 28, 19, 22, 15, 18, 18, 19, 26, 25, 14, 13, 17, 22, 22, 23, 18, 29, 11, 16, 24, 11, 25, 25, 24, 14, 24, 15, 27, 23, 12, 11, 28, 22, 19, 25, 27, 26, 12, 12, 16, 26, 17, 21, 14, 28, 17, 16, 13, 15, 15, 19, 26, 28, 12, 24, 21, 29, 21, 11, 29, 13, 29, 17, 18}
		//	//
		//	//	self.cards = append(self.cards,
		//	//		[]common.EMaJiangType{27, 24, 24, 23, 29, 23, 27, 27, 21, 26, 29, 24, 23, 13, 19, 21, 23, 25, 26, 24, 14, 22, 26, 21, 27, 19, 14, 22, 21, 29, 16, 29, 18, 12, 11, 28, 14, 26, 13, 13, 18}...)
		//	//	////////////////////////////////////////////////////////////////////////////////////
		//if self.GetParamInt(0) == 500 {
		//	//	////////////////////////////// test ////////////////////////////////////////////////
		//	self.cards = []common.EMaJiangType{17, 17, 17, 17, 23, 23, 24, 25, 25, 27, 27, 29, 29, 29, 28, 28, 28, 13, 14, 32, 34, 34, 34, 34, 31, 21, 31}
		//
		//	self.cards = append(self.cards,
		//		[]common.EMaJiangType{21, 19, 22, 24, 11, 37, 22, 15, 36, 31, 11, 38, 39, 15, 37, 35, 32, 11, 19, 12, 37, 33, 16, 26, 24, 26, 38, 38, 11, 14, 21, 32, 36, 26, 33, 37, 21, 39, 12, 22, 12, 15, 18, 35, 39, 32, 35, 33, 16, 16, 12, 19, 14, 14, 26, 27, 24, 31, 13, 35, 18, 36, 28, 23, 22, 16, 33, 15, 29, 18, 19, 18, 25, 25, 13, 23, 13, 36, 27, 39, 38}...)
		//	//	////////////////////////////////////////////////////////////////////////////////////
		//} // else if self.GetParamInt(0) == 300 {
		//	self.cards = []common.EMaJiangType{12, 12, 12, 14, 15, 16, 17, 18, 18, 18, 21, 21, 21, 23, 24, 25, 26, 27, 28, 29, 29, 22, 22, 22, 22, 23, 23, 25, 25, 26, 26}
		//	self.cards = append(self.cards,
		//		[]common.EMaJiangType{13, 11, 11, 19, 15, 15, 17, 11, 24, 14, 17, 19, 27, 27, 16, 15, 13, 14, 25, 27, 18, 13, 19, 29, 26, 21, 17, 12, 29, 28, 16, 11, 16, 19, 28, 13, 14, 28, 24, 24, 23}...)
		//} else if self.GetParamInt(0) == 400 {
		//	self.cards = []common.EMaJiangType{11, 12, 13, 17, 17, 18, 18, 19, 19, 21, 22, 16, 11, 14, 27, 25, 26, 15, 15, 15, 11, 22, 24, 26, 25, 28, 25, 26, 21, 24, 14}
		//	self.cards = append(self.cards,
		//		[]common.EMaJiangType{17, 29, 27, 23, 27, 12, 16, 15, 13, 14, 25, 18, 23, 27, 29, 12, 18, 13, 21, 19, 29, 22, 26, 21, 23, 17, 12, 22, 29, 28, 16, 11, 16, 19, 28, 13, 14, 28, 24, 24, 23}...)
		//}
	}

	self.track_log(colorized.Green("playing enter"))
	log_str := ""
	for _, v := range self.cards {
		log_str += fmt.Sprintf("%v", v.Value()) + ", "
	}
	self.track_log(colorized.Green("开局总牌:%v"), log_str)
	self.initData_()

	self.last_push_index = -1
	self.last_push_cardIndex = -1
	self.hu_count = 0
	self.push_count = 0
	//self.qingfu = make(map[int]bool, 0)
	self.multip = nil

	self.game_state = utils.NewFSM()
	self.game_state.Add(CHANGE_STATE, &changing{playing: self, s: CHANGE_STATE})
	//self.game_state.Add(DECIDE_STATE, &deciding{playing: self, s: DECIDE_STATE})
	self.game_state.Add(DEAL_STATE, &deal{playing: self, s: DEAL_STATE})
	self.game_state.Add(BREAKIN_STATE, &toss{playing: self, s: BREAKIN_STATE})

	self.games++
	if self.games == 1 {
		self.master = utils.Randx_y(0, self.sitDownCount())
		self.next_master = self.master
	} else {
		self.master = self.next_master
	}

	self.deal_player = self.master
	self.deal_count = 0
	self.dice = []uint8{
		uint8(utils.Randx_y(1, 7)),
		uint8(utils.Randx_y(1, 7)),
	}

	// 每个人发13张牌ca
	for i, p := range self.seats {
		if i == self.master { // 庄家多发一张牌
			p.cards.hand = append(p.cards.hand, self.cards[:14]...)
			self.cards = self.cards[14:]
		} else {
			p.cards.hand = append(p.cards.hand, self.cards[:13]...)
			self.cards = self.cards[13:]
		}

		p.cards.last_index = len(p.cards.hand) - 1
	}

	// 所有牌排序
	for i, player := range self.seats {
		s := &ca.Majiang_Sort{Cards: player.cards.hand}
		sort.Sort(s)

		if i == self.master {
			self.track_log(colorized.Green("庄家:%v accid:%v 名字:%v 钱:%v 牌:%v"), i, player.acc.AccountId, player.acc.GetName(), player.acc.GetMoney(), player.cards.String())
		} else {
			self.track_log(colorized.Green("玩家:%v accid:%v 名字:%v 钱:%v 牌:%v"), i, player.acc.AccountId, player.acc.GetName(), player.acc.GetMoney(), player.cards.String())
		}

	}

	//  广播消息，通知开始游戏
	sendHead := packet.NewPacket(nil)
	sendHead.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_START_DATA.UInt16())
	sendHead.WriteUInt8(uint8(self.master + 1))
	sendHead.WriteUInt8(self.dice[0]) // 骰子1
	sendHead.WriteUInt8(self.dice[1]) // 骰子2
	for _, player := range self.accounts {
		cardData := packet.NewPacket(nil)
		if index := self.seatIndex(player.AccountId); index == -1 { // 观战的人
			cardData.WriteUInt16(0)
		} else {
			cards := self.seats[index].cards.hand
			cardData.WriteUInt16(uint16(len(cards)))
			for _, card := range cards {
				cardData.WriteUInt8(card.Value())
			}
		}
		send_tools.Send2Account(packet.PacketMakeup(sendHead, cardData).GetData(), player.SessionId)
	}

	if self.GetParamInt(5) == 1 {
		self.game_state.Swtich(0, CHANGE_STATE)
	} else {
		self.game_state.Swtich(0, DEAL_STATE)
	}
}

func (self *playing) SaveQuit(accid uint32) bool {
	return false
}

func (self *playing) Tick(now int64) {
	self.game_state.Update(now)
}

func (self *playing) huCount(index int) {
	self.hu_count++
	if self.hu_count == 1 {
		self.next_master = index
	}
}

//  index != -1 杠牌的玩家自己分牌
func (self *playing) assignCard(Index int) {
	// 如果所有人都胡了，或者桌面没有牌了，进入结算
	settlement := false

	hu_count := 0
	for _, player := range self.seats {
		if player.hu != common.HU_NIL {
			hu_count++
		}
	}

	if hu_count == len(self.seats)-1 {
		settlement = true
	} else if len(self.cards) == 0 {
		settlement = true
		self.liuju = true
	}

	if settlement {
		self.status.Swtich(0, types.ERoomStatus_SETTLEMENT.Int32())
	} else {
		if Index == -1 {
			self.deal_player = self.nextIndex(self.deal_player)
			// 还需要判断当前分牌的玩家是否已经胡了，如果胡了，再换下一个
			if self.deal_player == -1 {
				log.Errorf("错误 的坐标 当前：%v", self.deal_player)
				return
			}
			for i := 0; i < 999; i++ {
				gamePlayer := self.seats[self.deal_player]
				if gamePlayer.hu != common.HU_NIL {
					self.deal_player = self.nextIndex(self.deal_player)
				} else {
					break
				}
				if i > 10 {
					log.Errorf("死循环了！ ")
					return
				}
			}
		} else {
			self.deal_player = Index
		}

		self.last_gang = Index // 记录这次发牌是不是杠来的

		// 选出发牌的玩家
		player := self.seats[self.deal_player]
		player.exclude_peng = make(map[common.EMaJiangType]bool)
		player.exclude_hu = 0 // 清除胡牌限制

		card := self.cards[0]
		player.cards.hand, player.cards.last_index = algorithm.InsertCard(player.cards.hand, card)
		player.card_time++
		self.cards = self.cards[1:]

		//通知客户端，发牌
		assign := packet.NewPacket(nil)
		assign.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_ASSISGN_CARD.UInt16())
		assign.WriteInt8(int8(self.deal_player + 1))
		assign.WriteUInt8(uint8(player.cards.last_index + 1))
		assign.WriteInt8(int8(card))
		self.SendBroadcast(assign.GetData())
		self.deal_count++
		self.track_log(colorized.Green("发牌 ***（%v）*** 玩家:%v 名字:%v 插入位置倒数:%v 剩余:%v张 玩家牌:%v"),
			card, player.acc.AccountId, player.acc.GetName(), player.cards.last_index, len(self.cards),
			player.cards.String())

		self.game_state.Swtich(0, DEAL_STATE)
	}
}

// 杠的分数计算
func (self *playing) gang_score(gamePlayer *GamePlayer, index int, pack packet.IPacket, rate int) packet.IPacket {
	ret := packet.NewPacket(nil)
	score := rate * self.GetParamInt(0)

	if self.last_gang != self.seatIndex(gamePlayer.acc.AccountId) {
		gamePlayer.gang_score_z = map[int]int64{}
	}
	amount_ := 0
	temp := packet.NewPacket(nil)
	if index == -1 {
		count := 0
		temppack := packet.NewPacket(nil)
		for index, player := range self.seats {
			if player.acc.AccountId == gamePlayer.acc.AccountId || player.hu != common.HU_NIL {
				continue
			}
			count++
			player.acc.AddMoney(-int64(score), 0, common.EOperateType_PANDA_GANG)
			gamePlayer.acc.AddMoney(int64(score), 0, common.EOperateType_PANDA_GANG)
			amount_ += score
			gamePlayer.gang_score[index] += int64(score)
			gamePlayer.gang_score_z[index] += int64(score)
			self.track_log(colorized.Green("玩家:%v 杠  赔付者:%v 座位:%v 赔钱:%v"),
				gamePlayer.acc.AccountId, player.acc.AccountId, index, score)

			temppack.WriteInt8(int8(index + 1))
			temppack.WriteInt64(int64(score))
		}
		pack.WriteUInt16(uint16(count))
		pack.CatBody(temppack)

		temp.WriteUInt16(uint16(count))
		temp.CatBody(temppack)
	} else {
		player := self.seats[index]
		player.acc.AddMoney(-int64(score), 0, common.EOperateType_PANDA_GANG)
		gamePlayer.acc.AddMoney(int64(score), 0, common.EOperateType_PANDA_GANG)
		amount_ += score
		gamePlayer.gang_score[index] += int64(score)
		gamePlayer.gang_score_z[index] += int64(score)
		self.track_log(colorized.Green("玩家:%v 直杠  赔付者:%v 座位:%v 赔钱:%v"),
			gamePlayer.acc.AccountId, player.acc.AccountId, index, score)

		pack.WriteUInt16(1)
		pack.WriteInt8(int8(index) + 1)
		pack.WriteInt64(int64(score))

		temp.WriteUInt16(uint16(1))
		temp.WriteInt8(int8(index + 1))
		temp.WriteInt64(int64(score))
	}

	ret.WriteInt64(int64(amount_))
	ret.CatBody(temp)
	return ret
}

// 额外番计算
func (self *playing) calcExtra(player *GamePlayer, card, card_zimo common.EMaJiangType, z bool) (ret []*ca.Majiang_Hu) {
	var (
		hand []common.EMaJiangType
		peng [][]common.EMaJiangType
		gang [][]common.EMaJiangType
	)

	if card == 0 {
		hand = player.cards.hand
	} else {
		hand, _ = algorithm.InsertCard(player.cards.hand, card)
	}
	peng = player.cards.peng
	gang = player.cards.gang

	ret = ca.XMMJ_CalcHuAndExtra(hand, peng, gang, card)
	if ret == nil {
		ret = []*ca.Majiang_Hu{}
	}

	menq := true
	for _, v := range player.show_card {
		if v.t == 4 || v.t == 1 || v.t == 3 {
			menq = false
			break
		}
	}

	if menq {
		for _, v := range ret {
			if v.Extra == nil {
				v.Extra = map[common.EMaJiangExtra]uint8{}
			}
			v.Extra[common.EXTRA_MENQ] = 1
		}
	}

	// 夹心五
	if len(self.seats) == 3 && (card%10 == 5 || card_zimo%10 == 5) {
		e4 := false
		e6 := false
		for _, c := range player.cards.hand {
			if c == card-1 {
				e4 = true
			}
			if c == card+1 {
				e6 = true
			}
		}

		if e4 && e6 {
			for _, v := range ret {
				if v.Extra == nil {
					v.Extra = map[common.EMaJiangExtra]uint8{}
				}
				v.Extra[common.EXTRA_JIAXINWU] = 1
			}
		}
	}

	if !z {
		return ret
	}

	if card == 0 {
		for _, v := range ret {
			if v.Extra == nil {
				v.Extra = map[common.EMaJiangExtra]uint8{}
			}
			v.Extra[common.EXTRA_ZIMO] = 1
		}
	}

	// 计算额外番数
	// 杠上花 杠&自摸
	if card == 0 && self.last_gang == self.seatIndex(player.acc.AccountId) { // card == 0 自摸
		for _, v := range ret {
			if v.Extra == nil {
				v.Extra = map[common.EMaJiangExtra]uint8{}
			}
			v.Extra[common.EXTRA_GANGSHANGHUA] = 1
		}
	}

	// 杠上炮 点炮是杠打出来的
	if card != 0 && self.deal_player == self.last_gang {
		for _, v := range ret {
			if v.Extra == nil {
				v.Extra = map[common.EMaJiangExtra]uint8{}
			}
			v.Extra[common.EXTRA_GANGSHANGPAO] = 1
		}
	}

	// 抢杠胡 点炮是弯杠
	if card != 0 && self.wanGang_qinHu == 1 {
		for _, v := range ret {
			if v.Extra == nil {
				v.Extra = map[common.EMaJiangExtra]uint8{}
			}
			v.Extra[common.EXTRA_QIANGGANGHU] = 1
		}
	}

	// 海底花 自摸 & 牌堆里没牌了
	if card == 0 && len(self.cards) == 0 {
		for _, v := range ret {
			if v.Extra == nil {
				v.Extra = map[common.EMaJiangExtra]uint8{}
			}
			v.Extra[common.EXTRA_HAIDIHUA] = 1
		}
	}

	// 海底炮 点炮 & 牌堆里没牌了
	if card != 0 && len(self.cards) == 0 {
		for _, v := range ret {
			if v.Extra == nil {
				v.Extra = map[common.EMaJiangExtra]uint8{}
			}
			v.Extra[common.EXTRA_HAIDIPAO] = 1
		}
	}
	return ret
}

func (self *playing) settlement() {
	self.all_ting = map[int]ca.Majiang_Hu{} // 听牌
	self.all_no_ting = []int{}              // 未听牌
	self.pigs = []int{}                     // 花猪
	// 计算 听牌 s和 未听牌 的玩家
	if len(self.cards) == 0 {
		for index, player := range self.seats {
			if player.hu != common.HU_NIL {
				continue
			}
			// 如果是花猪，肯定没有叫
			count := 0
			for _, card := range player.cards.hand {
				if int(card/10) == int(player.decide_t) {
					count = 1
					break
				}
			}
			if count == 1 {
				self.pigs = append(self.pigs, index)
			} else {
				all := []*ca.Majiang_Hu{}
				alljiao := algorithm.Jiao_(player.cards.hand, player.cards.peng, player.cards.gang)
				for _, v := range alljiao {
					var allhu []*ca.Majiang_Hu
					if v.T == 1 {
						var tempCard CardGroup
						m := player.cards
						tempCard = *player.cards
						tempCard.hand, _ = algorithm.InsertCard(tempCard.hand, v.Card)
						player.cards = &tempCard
						allhu = self.calcExtra(player, 0, v.Card, false) // 如果是自摸，替换参数，带入计算
						player.cards = m
					} else {
						allhu = self.calcExtra(player, v.Card, 0, false)
					}

					if allhu != nil {
						all = append(all, allhu...)
					}
				}
				so := &Majiang_fan_Sort{Room: self.Room, All: all}
				sort.Sort(so)

				if len(all) != 0 {
					self.all_ting[index] = *all[0]
				} else {
					self.all_no_ting = append(self.all_no_ting, index)
				}
			}
		}
	}

	self.track_log(colorized.Green("听牌的玩家:%v 未听牌的玩家:%v 花猪:%v"), self.all_ting, self.all_no_ting, self.pigs)

	ting_count := len(self.all_ting)
	no_ting_count := len(self.all_no_ting)
	bet := self.GetParamInt(0)

	if no_ting_count != 0 {
		self.settle_pig.WriteUInt16(uint16(len(self.pigs)))
		for _, index := range self.pigs {
			self.settle_pig.WriteInt8(int8(index + 1))
			pig := self.seats[index]
			// 花猪给未听牌的赔钱
			self.settle_pig.WriteUInt16(uint16(no_ting_count))
			for _, index_t := range self.all_no_ting {
				ting_player := self.seats[index_t]

				rate := FAN_RATIO[MAX_FAN]
				total_score := int64(rate * bet)
				pig.acc.AddMoney(-total_score, 0, common.EOperateType_PANDA_PIG)
				ting_player.acc.AddMoney(total_score, 0, common.EOperateType_PANDA_PIG)

				self.settle_pig.WriteInt8(int8(index_t + 1))
				self.settle_pig.WriteInt64(int64(total_score))
				self.track_log(colorized.Green("花猪:%v 给index:%v 赔钱:%v"), index, index_t, total_score)
			}
		}
	} else {
		self.settle_pig.WriteUInt16(uint16(0))
	}

	// 未听牌的给听牌的赔钱

	if ting_count == 0 || no_ting_count == 0 {
		self.settle_ting.WriteUInt16(0)
		// 未听牌的人，退杠收到的钱
		new := append(self.all_no_ting, self.pigs...)
		self.settle_ting.WriteUInt16(uint16(len(new)))
		for _, int := range new {
			self.settle_ting.WriteInt8(int8(int + 1))
			p := self.seats[int]
			self.settle_ting.WriteUInt16(uint16(len(p.gang_score)))
			for index, money := range p.gang_score {
				gamePlayer := self.seats[index]
				gamePlayer.acc.AddMoney(money, 0, common.EOperateType_PANDA_GANG)
				p.acc.AddMoney(-money, 0, common.EOperateType_PANDA_GANG)

				self.settle_ting.WriteInt8(int8(index + 1))
				self.settle_ting.WriteInt64(int64(money))
				self.track_log(colorized.Green("index:%v 给index:%v 退杠:%v"), int, index, money)
			}
		}
	} else {
		if !self.liuju {
			self.settle_ting.WriteUInt16(0)
			self.settle_ting.WriteUInt16(0)
		} else {
			self.settle_ting.WriteUInt16(uint16(ting_count))
			for it, hu := range self.all_ting {
				total_fan := uint8(self.hu_fan[hu.HuType])
				if hu.Extra != nil {
					for t, v := range hu.Extra {
						f := uint8(self.extra_fan[int32(t)] * int32(v))
						total_fan += f
					}
				}
				fan := total_fan

				if total_fan > MAX_FAN {
					fan = MAX_FAN
				}
				rate := FAN_RATIO[fan]
				total_score := int64(rate * bet)
				gamePlayer := self.seats[it]

				self.settle_ting.WriteInt8(int8(it + 1))
				self.settle_ting.WriteInt8(int8(total_fan))
				self.settle_ting.WriteInt64(int64(total_score))

				self.settle_ting.WriteUInt16(uint16(no_ting_count))
				for _, int := range self.all_no_ting {
					p := self.seats[int]
					gamePlayer.acc.AddMoney(total_score, 0, common.EOperateType_PANDA_HU)
					p.acc.AddMoney(-total_score, 0, common.EOperateType_PANDA_HU)

					self.settle_ting.WriteInt8(int8(int) + 1)
					self.settle_ting.WriteInt64(int64(total_score))
					self.track_log(colorized.Green("index:%v 给index:%v 赔钱:%v，胡:%v "), int, it, total_score, hu.HuType)
				}
			}

			// 未听牌的人，退杠收到的钱
			new := append(self.all_no_ting, self.pigs...)
			self.settle_ting.WriteUInt16(uint16(len(new)))
			for _, int := range new {
				self.settle_ting.WriteInt8(int8(int + 1))
				p := self.seats[int]
				self.settle_ting.WriteUInt16(uint16(len(p.gang_score)))
				for index, money := range p.gang_score {
					gamePlayer := self.seats[index]
					gamePlayer.acc.AddMoney(money, 0, common.EOperateType_PANDA_GANG)
					p.acc.AddMoney(-money, 0, common.EOperateType_PANDA_GANG)

					self.settle_ting.WriteInt8(int8(index + 1))
					self.settle_ting.WriteInt64(int64(money))
					self.track_log(colorized.Green("index:%v 给index:%v 退杠:%v"), int, index, money)
				}
			}
		}
	}

}

func (self *playing) zimo(gamePlayer *GamePlayer, all_hu []*ca.Majiang_Hu) (int64, packet.IPacket) {
	bet := self.GetParamInt(0)
	pack := packet.NewPacket(nil)
	total_fan := uint8(self.hu_fan[all_hu[0].HuType])

	if all_hu[0].Extra == nil {
		pack.WriteUInt16(0)
	} else {
		pack.WriteUInt16(uint16(len(all_hu[0].Extra)))
	}

	if all_hu[0].Extra != nil {
		for t, v := range all_hu[0].Extra {
			pack.WriteInt8(int8(t))
			f := uint8(self.extra_fan[int32(t)] * int32(v))
			pack.WriteUInt8(f)
			total_fan += f
		}
	}

	total_money := int64(0)
	cout := 0
	temp := packet.NewPacket(nil)
	for i, player := range self.seats {
		if player.acc.AccountId == gamePlayer.acc.AccountId {
			continue
		}
		//胡的人不用赔
		if player.hu != common.HU_NIL {
			continue
		}
		fan_ := total_fan
		//temp.WriteUInt8(0) // 没用的！！！抓报叫番数 0 表示没有抓报叫

		fan__ := fan_
		if fan__ > MAX_FAN {
			fan__ = MAX_FAN
		}
		rate := FAN_RATIO[fan__]
		total_score := int64(rate * bet)
		if player.acc.GetMoney() < uint64(total_score) {
			log.Errorf("玩家:%v身上的钱:%v 不够赔:%v", player.acc.AccountId, player.acc.GetMoney(), total_score)
			total_score = int64(player.acc.GetMoney())
		}
		player.acc.AddMoney(-total_score, 0, common.EOperateType_PANDA_HU)
		total_money += total_score
		gamePlayer.acc.AddMoney(total_score, 0, common.EOperateType_PANDA_HU)
		self.track_log(colorized.Green("自摸 玩家:%v 胡牌:%v 总番:%v 赔钱的人:%v 座位:%v 赔付总金额:%v"),
			gamePlayer.acc.AccountId, gamePlayer.hu, fan_, player.acc.AccountId, i, total_score)

		cout++
		temp.WriteInt8(int8(i + 1))
		temp.WriteInt64(total_score)
	}

	pack.WriteUInt16(uint16(cout))
	pack.CatBody(temp)

	return total_money, pack
}

func (self *playing) game_status() IPANDAStatus_Game_universal {
	status := self.game_state.Current()
	return status.(IPANDAStatus_Game_universal)
}

func (self *playing) decided_MAX(player *GamePlayer) int {
	max := -1
	for i, card := range player.cards.hand {
		if int8(card/10) == player.decide_t {
			max = i
		} else if max != -1 {
			break
		}
	}
	return max
}
func (self *playing) CombineMSG(packet packet.IPacket, acc *account.Account) {
	packet.WriteInt8(self.last_push_index + 1)
	packet.WriteInt8(self.last_push_cardIndex)
	packet.WriteUInt8(uint8(self.game_state.State()))

	packet.WriteUInt16(uint16(self.sitDownCount()))
	for index, player := range self.seats {
		if player != nil {
			packet.WriteUInt8(uint8(index + 1))
			packet.WriteUInt8(uint8(player.decide_t))
			packet.WriteInt8(player.hut)
			packet.WriteUInt8(player.hu.Value())
			packet.WriteUInt8(player.huCard.Value())

			// 手牌
			packet.WriteUInt16(uint16(len(player.cards.hand)))
			for _, card := range player.cards.hand {
				packet.WriteUInt8(card.Value())
			}
			// 织牌
			packet.WriteUInt16(uint16(len(player.show_card)))
			for _, card := range player.show_card {
				packet.WriteUInt8(card.card.Value())
				packet.WriteUInt8(card.t)
			}

			// 废牌
			packet.WriteUInt16(uint16(len(player.trash_cards)))
			for _, card := range player.trash_cards {
				packet.WriteUInt8(card.Value())
			}
		}
	}

	if self.multip == nil {
		packet.WriteUInt8(0)
		packet.WriteUInt16(0)
	} else {
		packet.CatBody(self.multip)
	}

	// 庄家可打手牌
	//packet.WriteUInt16(uint16(len(self.master_bjs)))
	//for _, v := range self.master_bjs {
	//	packet.WriteInt8(int8(v + 1))
	//}

	// 游戏状态数据
	self.game_status().Combine_Game_MSG(packet, acc)
}

func (self *playing) Leave(now int64) {
	self.track_log(colorized.Green("playing leave\n"))
	self.settlement()
}

///////////////////////////////// handler ///////////////////////////////////////////////////
func (self *playing) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)

	switch pack.GetMsgID() {
	default:
		self.game_state.Handle(actor, msg, session)
		return true
	}
	return true
}

func (self *Majiang_fan_Sort) Len() int {
	return len(self.All)
}
func (self *Majiang_fan_Sort) Less(i, j int) bool {
	iobj := self.All[i]
	jobj := self.All[j]
	ifan := self.hu_fan[int(iobj.HuType)]
	jfan := self.hu_fan[int(jobj.HuType)]

	if iobj.Extra != nil {
		for ft, c := range iobj.Extra {
			ifan += self.extra_fan[ft] * int32(c)
		}
	}

	if jobj.Extra != nil {
		for ft, c := range jobj.Extra {
			jfan += self.extra_fan[ft] * int32(c)
		}
	}

	if ifan > jfan {
		return true
	} else if ifan < jfan {
		return false
	} else {
		return self.All[i].HuType > self.All[j].HuType
	}

}
func (self *Majiang_fan_Sort) Swap(i, j int) {
	self.All[i], self.All[j] = self.All[j], self.All[i]
}
