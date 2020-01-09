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
	"root/server/dehgame/account"
	"root/server/dehgame/algorithm"
	"root/server/dehgame/event"
	"root/server/dehgame/send_tools"
	"root/server/dehgame/types"
	"sort"
	"strconv"
)

type (
	playing struct {
		*Room
		s         types.ERoomStatus
		deal_bout int // 发牌回合数

		speech_index   int   // 当前喊话座位
		speech_timeout int64 // 喊话倒计时 （到期时间戳 秒）

		bout_max_bet int64 // 当次回合最大押注

		bridgeVal int64 // 上一次搭桥的值 首次为芒果

		timerid      int64
		special_pack packet.IPacket
	}
)

func (self *playing) Enter(now int64) {
	self.track_log(colorized.Yellow("playing enter"))
	self.speech_index = 0
	self.deal_bout = 1
	self.continues = make([]*GamePlayer, 0)
	self.diu = make([]*GamePlayer, 0)
	self.qiao = make([]*GamePlayer, 0)
	mango := self.mango()
	for _, player := range self.seats {
		if player != nil && player.status == types.EGameStatus_PREPARE {
			player.status = types.EGameStatus_PLAYING
			if player.bobo < int64(mango) {
				log.Errorf("检查enter，这里的bobo值小于需要扣的值 bobo:%v deduct:%v", player.bobo, mango)
				return
			}
			player.timeout_count = 0
			player.bobo -= int64(mango)
			player.mangoVal += int64(mango) // 所有人都准备好了，押扣芒果分

			// 所有人一开始先加入可以喊话的人当中
			self.continues = append(self.continues, player)
		}
	}

	// 先选一个庄
	if self.lastBanker_index == -1 {
		index := utils.Randx_y(0, len(self.continues)-1)
		accid := self.continues[index]

		self.lastBanker_index = self.seatIndex(uint32(accid.acc.AccountId))
	} else {
		banker := self.nextIndex(self.lastBanker_index)
		player := self.seats[banker]
		for player.status != types.EGameStatus_PLAYING {
			banker = self.nextIndex(banker)
			player = self.seats[banker]
		}
		self.lastBanker_index = banker
	}
	self.track_log(colorized.Yellow("本轮庄家:[%v]号位"), self.lastBanker_index)

	// 随机 4*N 张牌
	number := self.playerCount()
	if number > 6 {
		log.Errorf("最多6人，不然强制显示下3张牌的计算有问题!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		return
	}
	cards := algorithm.GetRandom_Card((number * 4) + 3)
	index := 0
	for _, player := range self.continues {
		player.cards = make([]common.Card_info, 4)
		player.cards = cards[index : index+4]
		index += 4
		self.update_bet_bobo_mango(player.acc.AccountId) // 开始游戏，同步簸簸
		self.track_log(colorized.Yellow("玩家:[%v] 牌:[%v]"), player.acc.AccountId, player.cards)
	}
	self.next3cards = cards[index:]
	self.track_log(colorized.Yellow("最后三张牌:[%v]"), self.next3cards)
	//self.test_Deals()
	// 第一回合发牌
	self.bout(1)

	// 事件处理
	//event.Dispatcher.AddEventListener(event.EventType_hanhua, self)
	self.specific_account()
}
func (self *playing) test_Deals() {
	cards1 := []common.Card_info{
		{common.ECardType_HONGTAO.UInt8(), 10},
		{common.ECardType_FANGKUAI.UInt8(), 10},
		{common.ECardType_HONGTAO.UInt8(), 2},
		{common.ECardType_FANGKUAI.UInt8(), 2},
	}

	cards2 := []common.Card_info{
		{common.ECardType_JKEOR.UInt8(), 6},
		{common.ECardType_HONGTAO.UInt8(), 3},
		{common.ECardType_HEITAO.UInt8(), 5},
		{common.ECardType_MEIHUA.UInt8(), 5},
	}

	//cards3 := []common.Card_info{
	//	{common.ECardType_JKEOR.UInt8(), 6},
	//	{common.ECardType_HONGTAO.UInt8(), 3},
	//	{common.ECardType_FANGKUAI.UInt8(), 12},
	//	{common.ECardType_HONGTAO.UInt8(), 12},
	//}

	self.seats[0].cards = cards1
	self.seats[1].cards = cards2
	//self.seats[3].cards = cards4

}

func (self *playing) specific_account() {
	acccounts := make([]*account.Account, 0)
	for _, acc := range self.accounts {
		// 特殊账号处理
		if common.IsHaveSpecialType(acc.Special, common.SPECIAL_TEST) == true {
			acccounts = append(acccounts, acc)
		}
	}

	pk := packet.NewPacket(nil)
	pk.SetMsgID(protomsg.Old_MSGID_CX_SHOW_CARD_SPECIAL.UInt16())
	pk.WriteUInt16(uint16(len(self.continues)))

	for _, p := range self.continues {
		if p.status == types.EGameStatus_PLAYING {
			index := uint8(self.seatIndex(uint32(p.acc.AccountId)) + 1)
			log.Debugf("index:%v", index)
			pk.WriteUInt8(index)

			pk.WriteUInt8(p.cards[0][0])
			pk.WriteUInt8(p.cards[0][1])

			pk.WriteUInt8(p.cards[1][0])
			pk.WriteUInt8(p.cards[1][1])

			pk.WriteUInt8(p.cards[2][0])
			pk.WriteUInt8(p.cards[2][1])

			pk.WriteUInt8(p.cards[3][0])
			pk.WriteUInt8(p.cards[3][1])
		}
	}
	self.special_pack = pk

	for _, acc := range acccounts {
		send_tools.Send2Account(pk.GetData(), acc.SessionId)
	}
}

func (self *playing) Tick(now int64) {
	if now > self.speech_timeout && self.speech_timeout > 0 {
		accid := self.seats[self.speech_index].acc.AccountId
		pack := packet.NewPacket(nil)
		if self.CanSpeech(accid, types.DIU) == 1 {
			// 喊话倒计时结束，自动让玩家 丢
			pack.SetMsgID(protomsg.Old_MSGID_CX_HANHUA_DIU.UInt16())
			self.track_log(colorized.Yellow("玩家:[%v] 座位号:[%v]没有喊话 自动丢"), accid, self.speech_index)
		} else if self.CanSpeech(accid, types.XIU) == 1 {
			// 喊话倒计时结束，自动让玩家 丢
			pack.SetMsgID(protomsg.Old_MSGID_CX_HANHUA_XIU.UInt16())
			self.track_log(colorized.Yellow("玩家:[%v] 座位号:[%v]没有喊话 自动休"), accid, self.speech_index)
		} else if self.CanSpeech(accid, types.QIAO) == 1 {
			// 喊话倒计时结束，自动让玩家 丢
			pack.SetMsgID(protomsg.Old_MSGID_CX_HANHUA_QIAO.UInt16())
			self.track_log(colorized.Yellow("玩家:[%v] 座位号:[%v]没有喊话 自动敲"), accid, self.speech_index)
		}

		pack.WriteUInt32(accid)
		core.CoreSend(self.owner.Id, self.owner.Id, pack.GetData(), 0)
	}
}

func (self *playing) Leave(now int64) {
	event.Dispatcher.RemoveListener(self)
	self.owner.CancelTimer(self.timerid)
	self.track_log(colorized.Yellow("playing leave\n"))
}

// 当前状态下，玩家是否可以退出
func (self *playing) CanQuit(accId uint32) bool {
	return self.canQuit(accId)
}

func (self *playing) ShowCard(player *GamePlayer, show_self bool) packet.IPacket {
	pack := packet.NewPacket(nil)
	temp := packet.NewPacket(nil)
	tempcount := uint16(0)
	for i := 0; i < self.show_count; i++ {
		if i >= int(player.showcards) {
			break
		}
		tempcount++
		if i <= 1 && !show_self {
			temp.WriteUInt8(0)
			temp.WriteUInt8(0)
		} else {
			temp.WriteUInt8(player.cards[i][0])
			temp.WriteUInt8(player.cards[i][1])
		}
	}
	pack.WriteUInt16(tempcount)
	pack.CatBody(temp)
	return pack
}

//
func (self *playing) CombineMSG(pack packet.IPacket, acc *account.Account) {
	pack.WriteInt64(self.speech_timeout * 1000) // 到期时间
	speech_index := uint8(self.speech_index + 1)
	pack.WriteUInt8(speech_index) // 当前喊话的人

	if speech_index != 0 {
		speech_accid := self.seats[self.speech_index].acc.AccountId
		pack.WriteUInt8(self.CanSpeech(speech_accid, types.XIU))
		pack.WriteUInt8(self.CanSpeech(speech_accid, types.DIU))
		pack.WriteUInt8(self.CanSpeech(speech_accid, types.DA))
		pack.WriteUInt8(self.CanSpeech(speech_accid, types.QIAO))
		pack.WriteUInt8(self.CanSpeech(speech_accid, types.GEN))
		pack.WriteUInt32(uint32(self.max_bet))
		pack.WriteUInt32(self.Da_Val(speech_accid))
	}

	// 延迟发送
	self.owner.AddTimer(1000, 1, func(dt int64) {
		acccounts := make([]*account.Account, 0)
		for _, acc := range self.accounts {
			// 特殊账号处理
			if common.IsHaveSpecialType(acc.Special, common.SPECIAL_TEST) == true {
				acccounts = append(acccounts, acc)
			}
		}

		for _, acc := range acccounts {
			send_tools.Send2Account(self.special_pack.GetData(), acc.SessionId)
		}
	})

}

// 判断某个玩家能否喊某种操作
func (self *playing) CanSpeech(accid uint32, speech types.ESpeechStatus) uint8 {
	index := self.seatIndex(accid)
	if index == -1 {
		log.Errorf("错误的玩家:%v", accid)
		return 0
	}
	player := self.seats[index]
	if player == nil {
		log.Errorf("玩家不再房间内:%v", accid)
		return 0
	}

	if player.status != types.EGameStatus_PLAYING {
		log.Errorf("玩家状态不是游戏中  :%v", player.status.String())
		return 0
	}

	tobridge := func() bool {
		if player.last_speech == types.NIL {
			return true
		}

		if player.bet == 0 {
			return true
		}

		v := player.bet
		if v == 0 {
			v = self.bridgeVal
		}

		if self.bout_max_bet/v >= 2 {
			return true
		}

		return false
	}

	switch speech {
	case types.XIU:
		if player.bobo == 0 {
			return 0
		}
		// 每一轮押注为0，才能休
		if self.bout_max_bet == 0 {
			return 1
		} else {
			return 0
		}
	case types.DIU:
		if player.bobo == 0 {
			return 0
		}
		// 最大押注为0 并且第一轮发牌 不能丢
		if self.bout_max_bet == 0 {
			return 0
		} else {
			return 1
		}
	case types.DA:
		// 没有搭桥，不能喊大
		if !tobridge() {
			return 0
		}

		ex := int64(self.Da_Val(accid)) - player.bet
		if player.bobo >= ex {
			return 1
		} else {
			return 0
		}
	case types.QIAO:
		if player.bobo+player.bet >= self.bout_max_bet*2 {
			// 没有搭桥，不能喊大
			if !tobridge() {
				return 0
			}
		}
		return 1
	case types.GEN:
		if player.bobo == 0 {
			return 0
		}
		//簸簸的钱 >= 最大押注 并且 当前最大押注能>0
		if self.bout_max_bet > 0 && (player.bobo >= self.bout_max_bet-player.bet) {
			return 1
		} else {
			return 0
		}
	default:
		log.Errorf("错误的喊话操作!!!!%v", speech)
		return 0
	}
}

// 进入n回合发牌
func (self *playing) bout(b int) {

	if b != 1 && b != 2 && b != 3 {
		log.Errorf("error logic !!!!!!!:%v", b)
		return
	}
	self.bout_max_bet = 0
	self.deal_bout = b

	if self.max_bet == 0 {
		self.bridgeVal = int64(self.mango() * uint64(self.playerCount()))
	} else {
		self.bridgeVal = self.max_bet
	}

	// 每次发牌，清空玩家喊话
	for _, player := range self.seats {
		if player != nil {
			if player.status == types.EGameStatus_PLAYING {
				player.last_speech = types.NIL
			}

			if player.last_speech_c == types.DA || player.last_speech_c == types.GEN {
				player.last_speech_c = types.NIL
			}
		}

	}

	// 发牌 通知客户端
	self.ShowCards()

	// 发牌表现时间
	strConf := "DEAL_CARDS" + strconv.Itoa(b)
	animation := config.GetPublicConfig_Int64(strConf) // 第一次发牌的动画时长 秒
	self.speech_timeout = -1
	self.speech_index = -1
	self.timerid = self.owner.AddTimer(animation*1000, 1, func(dt int64) {
		if self.status.State() != self.s.Int32() {
			// 如果切换到其他状态，这里不需要往下执行
			return
		}

		if b == 1 {
			self.speech_index = self.lastBanker_index
		} else {
			// 牌最大的，第一个喊话
			index_card := b
			var maxCard *common.Card_info
			index_max := -1
			for _, player := range self.continues {
				if player == nil || player.status != types.EGameStatus_PLAYING {
					continue
				}

				if maxCard == nil {
					index_max = self.seatIndex(player.acc.AccountId)
					maxCard = &player.cards[index_card]
				} else {
					compare := algorithm.CompareOneCardMainType(player.cards[index_card], *maxCard)
					if compare == 1 {
						index_max = self.seatIndex(player.acc.AccountId)
						maxCard = &player.cards[index_card]
					} else if compare == 0 {
						index_max = int(algorithm.CalcFromBankerRecently(uint8(self.lastBanker_index), uint8(self.seatIndex(player.acc.AccountId)), uint8(index_max)))
						if index_max >= len(self.seats) {
							self.track_log("函数CalcFromBankerRecently，返回 值越界 bankIndex:%v 1:%v 2:%v!!", uint8(self.lastBanker_index), uint8(self.seatIndex(player.acc.AccountId)), uint8(index_max))
							return
						}
						maxCard = &self.seats[index_max].cards[index_card]
					}
				}
			}

			self.speech_index = index_max
		}

		// 检查事件
		if index := self.check_event(); index != -1 {
			self.speech_timeout = config.GetPublicConfig_Int64("SPEECH_TIME") + utils.SecondTimeSince1970()
			self.broadcast_next_speecher() // 当前喊话结束，广播一下次喊话
		}
	})

	self.track_log(colorized.Cyan("发牌:%v"), b)
}

// 判断某个玩家能否喊某种操作
func (self *playing) broadcast_next_speecher() {
	if self.speech_index == -1 {
		return // 没有人可以喊话，所以speech_index 为-1
	}
	self.speech_timeout = config.GetPublicConfig_Int64("SPEECH_TIME") + utils.SecondTimeSince1970()

	player := self.seats[self.speech_index]
	// 通知客户端喊话
	speech_broadcast := packet.NewPacket(nil)
	speech_broadcast.SetMsgID(protomsg.Old_MSGID_CX_HANHUA_PLAYER.UInt16())
	speech_broadcast.WriteInt64(self.speech_timeout * 1000)
	speech_broadcast.WriteUInt8(uint8(self.speech_index + 1))
	speech_broadcast.WriteUInt8(uint8(self.CanSpeech(player.acc.AccountId, types.XIU)))
	speech_broadcast.WriteUInt8(uint8(self.CanSpeech(player.acc.AccountId, types.DIU)))
	speech_broadcast.WriteUInt8(uint8(self.CanSpeech(player.acc.AccountId, types.DA)))
	speech_broadcast.WriteUInt8(uint8(self.CanSpeech(player.acc.AccountId, types.QIAO)))
	speech_broadcast.WriteUInt8(uint8(self.CanSpeech(player.acc.AccountId, types.GEN)))

	speech_broadcast.WriteUInt32(uint32(self.bout_max_bet)) // 可以跟的值
	v := self.max_bet
	if self.max_bet == 0 {
		v = self.bridgeVal
	} else {
		v *= 2
	}

	speech_broadcast.WriteUInt32(uint32(v)) // 可以喊大的值
	self.SendBroadcast(speech_broadcast.GetData())
	event.Dispatcher.Dispatch(&event.Hanhua{AccountId: player.acc.AccountId}, event.EventType_hanhua)
	self.track_log(colorized.Yellow("通知客户端，下一个喊话的人是:[%v] 座位号:[%v]"), player.acc.AccountId, self.speech_index)
}

// 发牌
func (self *playing) ShowCards() {
	pack := packet.NewPacket(nil)
	pack.SetMsgID(protomsg.Old_MSGID_CX_SHOW_CARDS.UInt16())
	pack.WriteInt64(self.speech_timeout)
	pack.WriteUInt8(uint8(self.deal_bout))
	pack.WriteUInt8(uint8(self.lastBanker_index + 1))

	if self.deal_bout == 1 { // 第一次发牌
		for accid, acc := range self.accounts {
			body := packet.NewPacket(nil)
			player_index := self.seatIndex(accid)
			body.WriteUInt16(uint16(self.playerCount()))
			for index, player := range self.seats {
				if player != nil && player.status == types.EGameStatus_PLAYING {
					player.showcards = 2
					body.WriteUInt32(uint32(player.acc.AccountId))
					body.WriteUInt8(uint8(index + 1))

					body.WriteUInt16(2)
					if index == player_index {
						for i := 0; i < 2; i++ {
							body.WriteUInt8(player.cards[i][0])
							body.WriteUInt8(player.cards[i][1])
						}
					} else {
						for i := 0; i < 2; i++ {
							body.WriteUInt8(0)
							body.WriteUInt8(0)
						}
					}
				}
			}
			send := packet.PacketMakeup(pack, body)
			send_tools.Send2Account(send.GetData(), acc.SessionId)
		}
		self.show_count = 2
	} else if self.deal_bout == 2 { // 第二次发牌
		body := packet.NewPacket(nil)
		temp := packet.NewPacket(nil)
		tempcount := 0
		for index, player := range self.seats {
			if player != nil && player.status == types.EGameStatus_PLAYING &&
				!self.isInDIU(player.acc.AccountId) &&
				!self.isInXIU(player.acc.AccountId) {
				player.showcards = 3
				tempcount++
				temp.WriteUInt32(uint32(player.acc.AccountId))
				temp.WriteUInt8(uint8(index + 1))
				temp.WriteUInt16(1)
				temp.WriteUInt8(player.cards[2][0])
				temp.WriteUInt8(player.cards[2][1])
				self.track_log(colorized.Yellow("--给玩家:[%v] 座位号:[%v] 发第三张牌:%v"), player.acc.AccountId, index, player.cards[2])
			}
		}
		body.WriteUInt16(uint16(tempcount))
		body.CatBody(temp)
		self.SendBroadcast(packet.PacketMakeup(pack, body).GetData())
		self.show_count = 3
	} else if self.deal_bout == 3 { // 第三次发牌
		if self.show_count == 2 {
			body := packet.NewPacket(nil)
			temp := packet.NewPacket(nil)
			tempcount := 0
			for index, player := range self.seats {
				if player != nil && player.status == types.EGameStatus_PLAYING &&
					!self.isInDIU(player.acc.AccountId) &&
					!self.isInXIU(player.acc.AccountId) {
					player.showcards = 4
					tempcount++
					temp.WriteUInt32(uint32(player.acc.AccountId))
					temp.WriteUInt8(uint8(index + 1))
					temp.WriteUInt16(2)
					for i := 2; i < 4; i++ {
						temp.WriteUInt8(player.cards[i][0])
						temp.WriteUInt8(player.cards[i][1])
						self.track_log(colorized.Yellow("给玩家:[%v] 座位号:[%v] 发第%v 张牌:%v"), player.acc.AccountId, index, i+1, player.cards[i])
					}
				}
			}
			body.WriteUInt16(uint16(tempcount))
			body.CatBody(temp)
			self.SendBroadcast(packet.PacketMakeup(pack, body).GetData())
		} else if self.show_count == 3 {
			body := packet.NewPacket(nil)
			temp := packet.NewPacket(nil)
			tempcount := 0
			for index, player := range self.seats {
				if player != nil && player.status == types.EGameStatus_PLAYING &&
					!self.isInDIU(player.acc.AccountId) &&
					!self.isInXIU(player.acc.AccountId) {
					player.showcards = 4
					tempcount++
					temp.WriteUInt32(uint32(player.acc.AccountId))
					temp.WriteUInt8(uint8(index + 1))
					temp.WriteUInt16(1)
					temp.WriteUInt8(player.cards[3][0])
					temp.WriteUInt8(player.cards[3][1])
					self.track_log(colorized.Yellow("--给玩家:[%v] 座位号:[%v] 发第四张牌:%v"), player.acc.AccountId, index, player.cards[3])
				}
			}
			body.WriteUInt16(uint16(tempcount))
			body.CatBody(temp)
			self.SendBroadcast(packet.PacketMakeup(pack, body).GetData())
		}
		self.show_count = 4
	}
}

// 喊话操作统一处理
func (self *playing) operation(accid uint32, t types.ESpeechStatus, next_bet int64) {
	index := self.seatIndex(accid)
	if index == -1 {
		log.Warnf("玩家不再房间内 :%v", accid)
		return
	}
	if self.speech_index != index {
		log.Warnf("还未轮到 %v 号位玩家喊话  当前喊话座位:%v ", index, self.speech_index)
		return
	}

	// 判断玩家状态是否在游戏中
	player := self.seats[index]
	if player.status != types.EGameStatus_PLAYING {
		log.Warnf("玩家:[%v]号位 状态不是游戏中 当前玩家状态:[%v]", index, player.status.String())
		return
	}

	// 判断玩家能否操作
	if self.CanSpeech(accid, t) == 0 {
		log.Warnf("玩家:[%v]号位 不能操作[%v]", index, t.String())
		return
	}

	if player.bobo == 0 {
		log.Warnf("玩家:[%v]号位 簸簸已经没有钱了", index, t.String())
		return
	}

	// 如果有人敲了，从喊话列表删除，加入敲列表
	qiaoFun := func(p *GamePlayer) {
		for index, cp := range self.continues {
			if cp.acc.AccountId == p.acc.AccountId {
				self.continues = append(self.continues[:index], self.continues[index+1:]...)
				break
			}
		}
		self.qiao = append(self.qiao, p)
	}

	decrease := int64(0)
	send := packet.NewPacket(nil)
	switch t {
	case types.XIU:
		send.SetMsgID(protomsg.Old_MSGID_CX_HANHUA_XIU.UInt16())
		player.last_bet = 0

	case types.DIU:
		send.SetMsgID(protomsg.Old_MSGID_CX_HANHUA_DIU.UInt16())
		for index, cp := range self.continues {
			if cp.acc.AccountId == player.acc.AccountId {
				self.continues = append(self.continues[:index], self.continues[index+1:]...)
				break
			}
		}
		if player.bet == 0 {
			xiaoP := int64(self.GetParamInt(0)) // 小皮
			player.last_bet = xiaoP
			player.bet += xiaoP
			player.bobo -= xiaoP
		}
		self.diu = append(self.diu, player)
		self.pipool += player.bet

	case types.DA:
		if self.max_bet == 0 {
			if next_bet < int64(self.mango())*int64(self.playerCount()) {
				log.Warnf("第一个喊话，分数小于总的芒果分", next_bet, self.mango())
				return
			}
		} else if next_bet < self.max_bet*2 {
			log.Warnf("大的值不够，当前下注最大:[%v] 喊话回合数:[%v]  芒果总分:[%v]", self.max_bet, self.deal_bout, self.mango())
			return
		}
		ex := next_bet - player.bet
		if ex == player.bobo {
			t = types.QIAO
			qiaoFun(player)

			send.SetMsgID(protomsg.Old_MSGID_CX_HANHUA_QIAO.UInt16())
		} else if ex > player.bobo {
			log.Warnf("身上钱不够喊大：[%v] bet:[%v] bobo:[%v]", next_bet, player.bet, player.bobo)
			return
		} else {
			send.SetMsgID(protomsg.Old_MSGID_CX_HANHUA_DA.UInt16())
		}
		player.last_bet = next_bet
		decrease = ex

	case types.QIAO:
		send.SetMsgID(protomsg.Old_MSGID_CX_HANHUA_QIAO.UInt16())
		player.last_bet = player.bobo + player.bet
		decrease = player.bobo
		qiaoFun(player)

	case types.GEN:
		ex := self.max_bet - player.bet
		if ex == player.bobo {
			t = types.QIAO
			qiaoFun(player)
			send.SetMsgID(protomsg.Old_MSGID_CX_HANHUA_QIAO.UInt16())
		} else {
			send.SetMsgID(protomsg.Old_MSGID_CX_HANHUA_GEN.UInt16())
		}
		player.last_bet = self.max_bet
		decrease = ex
	}

	player.bet += decrease
	player.bobo -= decrease
	player.last_speech = t
	player.last_speech_c = t

	if player.bet > self.max_bet {
		self.max_bet = player.bet
		self.bout_max_bet = player.bet
	}

	send.WriteUInt8(0)                // 结果
	send.WriteUInt8(uint8(index + 1)) // 座位号
	send.WriteInt64(player.bet)       // 总下注
	send.WriteUInt8(uint8(t))         // 新状态
	self.SendBroadcast(send.GetData())
	self.update_bet_bobo_mango(player.acc.AccountId) // 喊话，更新

	self.track_log(colorized.Cyan("玩家:[%v] 座位号:[%v] 喊话:[%v], 喊的钱:[%v], 簸簸:[%v]"), player.acc.AccountId, index, t.String(), player.bet, player.bobo)
	// 检查事件
	if index := self.check_event(); index != -1 {
		// ???
		self.speech_index = index
		self.speech_timeout = config.GetPublicConfig_Int64("SPEECH_TIME") + utils.SecondTimeSince1970()
		self.broadcast_next_speecher() // 当前喊话结束，广播一下次喊话
	}
}

func (self *playing) isInDIU(accid uint32) bool {
	for _, player := range self.diu {
		if player.acc.AccountId == accid {
			return true
		}
	}
	return false
}

func (self *playing) isInQIAO(accid uint32) bool {
	for _, player := range self.qiao {
		if player.acc.AccountId == accid {
			return true
		}
	}

	return false
}

// 判断当前是否满足特定条件 返回是还有一下次喊话
func (self *playing) check_event() int {
	lastSpeech := types.NIL
	if self.speech_index != -1 {
		lastSpeech = self.seats[self.speech_index].last_speech
	}

	// 大小皮
	if len(self.diu) == self.playerCount()-1 {
		self.overType = types.ESettlementStatus_daxiaoP
		self.switchStatus(utils.SecondTimeSince1970(), types.ERoomStatus_SETTLEMENT)
		return -1
	}

	xius := []*GamePlayer{}
	// 流局判断
	allxiu := false
	if len(self.continues) > 1 {
		allxiu = true
		for _, player := range self.continues {
			xius = append(xius, player)
			if player != nil && player.last_speech != types.XIU {
				allxiu = false
			}
		}
	}

	if allxiu && lastSpeech == types.XIU {
		// 判断是否死皮
		if l := len(self.qiao); l != 0 {
			// 死皮
			self.overType = types.ESettlementStatus_sipi
			if l >= 2 {
				// 死皮特殊发牌，一次发完所有牌
				self.xiu = xius
				self.bout(3)
				self.switchStatus(utils.SecondTimeSince1970(), types.ERoomStatus_ARRANGEMENT)
				return -1
			} else {
				self.switchStatus(utils.SecondTimeSince1970(), types.ERoomStatus_SETTLEMENT)
				return -1
			}
		} else {
			// 休芒
			self.overType = types.ESettlementStatus_xiumang
			self.switchStatus(utils.SecondTimeSince1970(), types.ERoomStatus_SETTLEMENT)
			return -1
		}
	}

	next := self.nextSpeecher()

	deal_card := false
	// 判断理牌和发牌 /////////////////////////////////////////////////////////////////////////////
	conclutionVal := int64(-1)
	if self.bout_max_bet > 0 {
		for _, player := range self.continues {
			if conclutionVal == -1 {
				conclutionVal = player.bet
			} else {
				if conclutionVal != player.bet {
					conclutionVal = 0
					break
				}
			}
		}
		if conclutionVal > 0 {
			if len(self.qiao) == 0 {
				deal_card = true
			} else {
				// 选出敲最大的
				s := &property_sorte{S: self.qiao}
				sort.Sort(s)
				if conclutionVal >= s.S[0].bet {
					deal_card = true
				}
			}
		}
	}

	// 判断理牌和发牌 /////////////////////////////////////////////////////////////////////////////

	// 人都丢了，或者敲了
	if len(self.continues) == 0 || next == -1 {
		deal_card = true
	}

	if deal_card {
		if self.deal_bout >= 3 {
			// 理牌
			self.overType = types.ESettlementStatus_compareCard
			self.switchStatus(utils.SecondTimeSince1970(), types.ERoomStatus_ARRANGEMENT)
			return -1
		} else {
			// 继续发牌
			self.bout(self.deal_bout + 1)
			return -1
		}
	}
	return next
}

// 判断当前是否满足特定条件
func (self *playing) nextSpeecher() int {
	next := self.speech_index
	if next == -1 {
		return next
	}
	// 决定下一个说话者
	for {
		next = self.nextIndex(next)

		if next == self.speech_index {
			//log.Errorf("下一位 喊话者 重复轮询到自己，逻辑有错误！！！！next:[%v]", next)
			return -1
		}

		player := self.seats[next]
		if player == nil {
			log.Errorf("获得 玩家是nil 座位号:[%v] ", next)
			return -1
		}
		if player.status != types.EGameStatus_PLAYING {
			continue
		}

		// 在丢和敲里的玩家不能喊话
		if self.isInDIU(player.acc.AccountId) || self.isInQIAO(player.acc.AccountId) {
			continue
		}

		break
	}
	return next
}

///////////////////////////////// handler ///////////////////////////////////////////////////
func (self *playing) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_CX_HANHUA_DIU.UInt16(): // 丢
		pack := packet.NewPacket(msg)
		accountId := pack.ReadUInt32()
		self.operation(accountId, types.DIU, 0)
	case protomsg.Old_MSGID_CX_HANHUA_XIU.UInt16(): // 休
		pack := packet.NewPacket(msg)
		accountId := pack.ReadUInt32()
		self.operation(accountId, types.XIU, 0)
	case protomsg.Old_MSGID_CX_HANHUA_QIAO.UInt16(): // 敲
		pack := packet.NewPacket(msg)
		accountId := pack.ReadUInt32()
		self.operation(accountId, types.QIAO, 0)
	case protomsg.Old_MSGID_CX_HANHUA_DA.UInt16(): // 大
		pack := packet.NewPacket(msg)
		accountId := pack.ReadUInt32()
		val := pack.ReadInt64()
		self.operation(accountId, types.DA, val)
	case protomsg.Old_MSGID_CX_HANHUA_GEN.UInt16(): // 跟
		pack := packet.NewPacket(msg)
		accountId := pack.ReadUInt32()
		self.operation(accountId, types.GEN, 0)
	default:
		log.Warnf("playing 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}
	return true
}
