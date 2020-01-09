package room

import (
	"root/common"
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"fmt"
	"math"
	"root/protomsg"
	"root/server/dehgame/account"
	"root/server/dehgame/algorithm"
	"root/server/dehgame/send_tools"
	"root/server/dehgame/types"
	"sort"
	"strconv"
)

type (
	Settlement_sort struct {
		S  []*GamePlayer
		Ht int8 // 0 头  1 尾
		R  *Room
	}
	settlement struct {
		*Room
		s               types.ERoomStatus
		timestamp       int64 // 结算倒计时 时间戳 秒
		settlement_pack packet.IPacket
		bouns_pack      packet.IPacket
		clear_xiumang   bool
		next_card       int
		force_watch     map[uint32]bool
		bouns_time      int64
	}

	Settle struct {
		bobo  int64
		accid uint32
	}
	Win_settle struct {
		S []*Settle
	}
)

func (self *settlement) Enter(now int64) {
	self.force_watch = make(map[uint32]bool)
	self.next_card = 0
	duration := config.GetPublicConfig_Int64("SETTLEMENT_TIME") // 持续时间 秒
	self.track_log(colorized.White("settlement enter duration:%v"), duration)
	self.timestamp = now + int64(duration)
	self.bouns_pack = packet.NewPacket(nil)

	self.deal_bouns()

}

func (self *settlement) deal_bouns() {
	scales := []int32{}
	accids := []uint32{}
	cardTypes := []string{}

	deals := make([]*GamePlayer, 0)
	// 先找出需要分牌的玩家
	deals = append(deals, self.continues...)
	deals = append(deals, self.qiao...)

	in := func(id uint32) bool {
		for _, p := range deals {
			if p.acc.AccountId == id {
				return true
			}
		}
		return false
	}

	max_accountId := uint32(0)
	max_t := types.SPECIAL_CARD_NIL
	max_tail := uint8(0)
	for _, player := range self.seats {
		if player != nil && player.status == types.EGameStatus_PLAYING && in(player.acc.AccountId) {
			self.track_log(colorized.Green("分牌后: Accid:[%v] cards:[%v]"), player.acc.AccountId, player.cards)
			head, tail, _ := algorithm.CalcOnePlayerCardType(player.cards, 0, false)
			t := algorithm.CalcReceiveAward(head, tail)
			if t != types.SPECIAL_CARD_NIL {
				ret := algorithm.CompareCardSet(tail, max_tail)
				if ret == 1 || max_accountId == 0 {
					max_accountId = player.acc.AccountId
					max_t = t
					max_tail = tail
				}
			}
		}
	}

	if max_accountId != 0 && self.show_card {
		scale := config.GetPublicConfig_Int64("CARD_TYPE_SCALE_" + strconv.Itoa(int(max_t)))
		scales = append(scales, int32(scale))
		accids = append(accids, max_accountId)
		cardTypes = append(cardTypes, max_t.String())
	}
	count := uint16(len(accids))
	self.bouns_pack.WriteUInt16(count)
	if count != 0 && self.clubID == 0 {
		self.bouns_time = 3 // 中奖额外时间
		self.Get_bonus(scales, accids, cardTypes, self.bouns_pack, self.conclusion)
	} else {
		self.conclusion()
	}
}
func (self *settlement) conclusion() {
	self.timestamp += self.bouns_time
	self.settlement_pack = packet.NewPacket(nil)
	self.settlement_pack.WriteInt64(int64(self.timestamp + 1000))
	self.settlement_pack.WriteUInt8(uint8(self.overType))

	///////////////////////// 结算前的金额 //////////////////////////////////////////////////////
	self.settlement_pack.CatBody(self.after_playing_pack)
	///////////////////////// 结算前的金额 //////////////////////////////////////////////////////

	BIG_PI := int64(self.GetParamInt(3))
	SMALL_PI := int64(self.GetParamInt(0)) // 小p

	fee_calcuate := false
	var (
		head_max int8 = -1
		tail_max int8 = -1
		shows    []*GamePlayer

		extra_val int64 = 0 // 尾最大的，额外显示增加自己的芒果分
	)
	switch self.overType {
	case types.ESettlementStatus_xiumang:
		self.clear_xiumang = false
		if len(self.qiao) > 0 {
			self.track_log("结算异常, 休芒结算中敲列表人数:%v", len(self.qiao))
		}

		nMaxCount := self.playerCount()
		if nMaxCount != len(self.diu)+len(self.continues) {
			self.track_log("结算异常, 休芒结算中参与游戏人数不匹配:%v, 丢人数:%v, 喊话人数:%v", nMaxCount, len(self.diu), len(self.continues))
		}

		nMaxMangoCount := int8(config.GetPublicConfig_Int64("DEH_MAX_MANGO"))
		if self.mangoCount < nMaxMangoCount {
			self.mangoCount++
		}

		self.track_log(colorized.White("-------------------   RoomID:%v Games:%v, 休芒结算, 小皮:%v, 大皮:%v, 芒果次数:%v"), self.roomId, self.games, SMALL_PI, BIG_PI, self.mangoCount)
		self.continues = append(self.continues, self.qiao...)
		for _, tPlayer := range self.continues {
			self.track_log(colorized.White("下标:%v %v %v, ID:%v %v, Money:%v, BoBo:%v"), self.seatIndex(tPlayer.acc.AccountId), tPlayer.cards, tPlayer.last_speech, tPlayer.acc.AccountId, tPlayer.acc.Name, tPlayer.acc.GetRMB(), tPlayer.bobo)
		}
		self.timestamp -= 1

	case types.ESettlementStatus_daxiaoP:
		self.timestamp -= 2
		fee_calcuate = true
		self.clear_xiumang = true
		var tWiner *GamePlayer
		var sLoser []*GamePlayer
		lose := BIG_PI - SMALL_PI
		var nWinTotal int64

		arr := append(self.continues, self.qiao...)
		if len(arr) != 1 {
			self.track_log("数据异常!!!!!!!!!!!!!!，大小皮结算，len(self.continues)：%v", len(self.continues))
			return
		}

		tWiner = arr[0]
		tail_max = int8(self.seatIndex(tWiner.acc.AccountId))
		extra_val = tWiner.mangoVal
		sLoser = self.diu
		for _, tPlayer := range sLoser {
			if tPlayer.bet > SMALL_PI {
				lose = 0
			}
		}

		self.track_log(colorized.White("-------------------   RoomID:%v  %v, Games:%v, 大小皮结算, 小皮:%v, 大皮:%v, 芒果次数:%v"), self.roomId, common.EGameType(self.gameType).String(), self.games, SMALL_PI, BIG_PI, self.mangoCount)
		// 输家先输下注和芒果分
		for _, tPlayer := range sLoser {
			nWinTotal += tPlayer.bet + tPlayer.mangoVal + lose
			tPlayer.bobo -= lose
			tPlayer.bet = 0
			tPlayer.mangoVal = 0
			self.track_log(colorized.White("下标:%v %v, ID:%v %v, Money:%v, BoBo:%v; bet:%v 输芒果:%v"), self.seatIndex(tPlayer.acc.AccountId), tPlayer.cards, tPlayer.acc.AccountId, tPlayer.acc.Name, tPlayer.acc.GetRMB(), tPlayer.bobo, tPlayer.bet, tPlayer.mangoVal)

		}
		tWiner.bobo += tWiner.bet + tWiner.mangoVal + nWinTotal
		tWiner.bet = 0
		tWiner.mangoVal = 0
		self.track_log(colorized.White("下标:%v %v %v, ID:%v %v, Money:%v, BoBo:%v; 赢:%v"), self.seatIndex(tWiner.acc.AccountId), tWiner.cards, tWiner.last_speech, tWiner.acc.AccountId, tWiner.acc.Name, tWiner.acc.GetRMB(), tWiner.bobo, nWinTotal)

	case types.ESettlementStatus_sipi, types.ESettlementStatus_compareCard:
		fee_calcuate = true
		self.clear_xiumang = true
		if len(self.qiao) == 1 && types.ESettlementStatus_sipi == self.overType {
			self.timestamp -= 2
		}
		head_max, tail_max, shows, extra_val = self.settlement(self.overType == types.ESettlementStatus_sipi)
		log.Debugf("尾巴最大:%v", tail_max)
	default:
		self.track_log(colorized.White("settlement Enter 异常结束状态:%v"), self.overType.String())
	}

	// 所有参与玩家退换下注金额和芒果分
	for _, player := range self.seats {
		if player != nil && player.status == types.EGameStatus_PLAYING {
			player.bobo += player.bet + player.mangoVal
			player.bet = 0
			player.mangoVal = 0
		}
	}

	arr := &Win_settle{S: make([]*Settle, 0)}
	for index, player := range self.seats {
		if player != nil && player.status != types.EGameStatus_GIVE_UP {
			profit_val := player.bobo - self.after_playing_bobo[index]
			player.profit += profit_val
			arr.S = append(arr.S, &Settle{bobo: profit_val, accid: player.acc.AccountId})
		}
	}

	// 抽水、服务费相关 //////////////////////////////////////////////////////////////////////////////////
	if fee_calcuate == true {
		sort.Sort(arr)

		fee := self.mango() - uint64(self.GetParamInt(3))
		count := config.GetPublicConfig_Int64("FEE_PLAYER_COUNT")
		if self.playerCount() < int(count) {
			fee += uint64(self.GetParamInt(0)) // 小p
		} else {
			fee += uint64(self.GetParamInt(3))
		}

		max := arr.S[0].bobo
		div := 0
		for _, s := range arr.S {
			if s.bobo == max {
				div++
			}
		}

		///////////////////////////////////// 只有两个人，抽水只抽小P //////////////////////////////////////////////////////
		if self.playerCount() == 2 {
			fee = uint64(self.GetParamInt(0)) // 小p
		}
		///////////////////////////////////// 只有两个人，抽水只抽小P //////////////////////////////////////////////////////
		tax_scale := uint64(config.GetPublicConfig_Int64("TAX"))
		average := float64(fee) / float64(div)
		average = math.Floor(average/100) * 100
		fee = uint64(average) * uint64(div)
		substract_val := fee / uint64(div)
		for i := 0; i < div; i++ {
			accid := arr.S[i].accid
			player := self.seats[self.seatIndex(accid)]
			player.bobo -= int64(substract_val)
			player.profit -= int64(substract_val)
			player.extractDec += int64(substract_val)
			self.track_log(colorized.White("大赢家 Accid:[%v] 座位号:[%v] 扣除抽水:[%v]"), accid, self.seatIndex(accid), average)
		}

		// 服务费需要扣除官税
		fee = fee * tax_scale / 100

		// 每人平分服务
		serviceFee := fee / uint64(self.playerCount())

		// 服务费还需要扣除掉奖金池增加部分
		reward_scale := uint64(config.GetPublicConfig_Int64("REWARD_POOL_SCALE"))
		serviceFee = serviceFee - serviceFee*reward_scale/100
		self.track_log(colorized.White("本局服务费 每人:%v"), serviceFee)
		servicepack := packet.NewPacket(nil)

		playercount := uint16(self.playerCount())
		updateAccount := packet.NewPacket(nil)
		updateAccount.SetMsgID(protomsg.Old_MSGID_UPDATE_ACCOUNT.UInt16())
		updateAccount.WriteUInt32(self.roomId)
		updateAccount.WriteUInt8(0)
		updateAccount.WriteUInt16(playercount)
		for index, player := range self.seats {
			if player != nil && player.status == types.EGameStatus_PLAYING {
				updateAccount.WriteUInt32(uint32(player.acc.AccountId))
				updateAccount.WriteInt64(int64(player.acc.GetMoney()))
				updateAccount.WriteInt64(int64(player.bobo - self.after_playing_bobo[index]))
				updateAccount.WriteString("")

				servicepack.WriteUInt32(uint32(player.acc.AccountId))
				servicepack.WriteUInt32(uint32(serviceFee))
			}
		}
		send_tools.Send2Hall(updateAccount.GetData())

		if serviceFee > 0 {
			ser_fee := packet.NewPacket(nil)
			ser_fee.SetMsgID(protomsg.Old_MSGID_UPDATE_SERVICE_FEE.UInt16())
			ser_fee.WriteUInt8(uint8(self.gameType))
			ser_fee.WriteUInt32(uint32(self.roomId))
			ser_fee.WriteUInt16(playercount)
			ser_fee.CatBody(servicepack)
			send_tools.Send2Hall(ser_fee.GetData())
		}

		// 奖金池计算  增加金额 = 服务费 * 官税后 * 奖金池比例
		award_val := fee * reward_scale / 100
		//self.track_log(colorized.White("奖金池增加:%v 现有金额:%v"), award_val, RoomMgr.Bonus[uint32(self.GetParamInt(0))])
		if self.clubID == 0 {
			RoomMgr.Add_bonus(uint32(self.GetParamInt(0)), award_val)
		}

		core.LocalCoreSend(0, common.EActorType_MAIN.Int32(), func() {
			RoomMgr.Fee += int32(fee)
		})
	}

	///////////////////////// 结算后的金额 //////////////////////////////////////////////////////
	temppack := packet.NewPacket(nil)
	count := uint16(0)
	for index, player := range self.seats {
		if player != nil && player.status != types.EGameStatus_GIVE_UP {
			count++
			temppack.WriteUInt8(uint8(index + 1))
			if int8(index) == tail_max {
				log.Debugf("结算，尾巴最大:%v   %v", player.acc.Name, extra_val)
				temppack.WriteInt64(int64(player.bobo + extra_val))
			} else {
				temppack.WriteInt64(int64(player.bobo))
			}

		}
	}

	self.settlement_pack.WriteUInt16(count)
	self.settlement_pack.CatBody(temppack)

	///////////////////////// 需要亮牌的玩家//////////////////////////////////////////////////////
	if !self.show_card {
		self.settlement_pack.WriteUInt16(0)
		self.settlement_pack.WriteUInt8(uint8(head_max + 1))
		self.settlement_pack.WriteUInt8(uint8(tail_max + 1))
	} else {
		self.settlement_pack.WriteUInt16(uint16(len(shows)))
		for _, player := range shows {
			self.settlement_pack.WriteUInt8(uint8(self.seatIndex(player.acc.AccountId)) + 1)
			self.settlement_pack.WriteUInt8(uint8(player.cards[2][0]))
			self.settlement_pack.WriteUInt8(uint8(player.cards[2][1]))
			self.settlement_pack.WriteUInt8(uint8(player.cards[3][0]))
			self.settlement_pack.WriteUInt8(uint8(player.cards[3][1]))
			self.settlement_pack.WriteUInt8(uint8(player.cards[0][0]))
			self.settlement_pack.WriteUInt8(uint8(player.cards[0][1]))
			self.settlement_pack.WriteUInt8(uint8(player.cards[1][0]))
			self.settlement_pack.WriteUInt8(uint8(player.cards[1][1]))
		}

		self.settlement_pack.WriteUInt8(uint8(head_max + 1))
		self.settlement_pack.WriteUInt8(uint8(tail_max + 1))
	}

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_CX_SETTLEMENT.UInt16())
	send.CatBody(self.settlement_pack)
	send.CatBody(self.bouns_pack)
	self.SendBroadcast(send.GetData())
}

func (self *Win_settle) Len() int {
	return len(self.S)
}
func (self *Win_settle) Less(i, j int) bool {
	return self.S[i].bobo > self.S[j].bobo
}
func (self *Win_settle) Swap(i, j int) {
	self.S[i], self.S[j] = self.S[j], self.S[i]
}

func (self *settlement) settlement(s bool) (max_head, max_tail int8, shows []*GamePlayer, tail_extra int64) {
	var (
		A []*GamePlayer // 参与分牌的人
		B []*GamePlayer // 休的人
		C []*GamePlayer // 丢的人
	)
	/////////////////////// print //////////////////////////////////////////////
	self.track_log(colorized.White("结算settlement"))
	for _, player := range self.continues {
		self.track_log(colorized.White("continues->>>>座位号:[%v] Accid:[%v] bet:%v"), self.seatIndex(player.acc.AccountId), player.acc.AccountId, player.bet)
	}
	for _, player := range self.qiao {
		self.track_log(colorized.White("qiao->>>>座位号:[%v] Accid:[%v] bet:%v "), self.seatIndex(player.acc.AccountId), player.acc.AccountId, player.bet)
	}
	for _, player := range self.diu {
		self.track_log(colorized.White("diu->>>>座位号:[%v] Accid:[%v] bet:%v"), self.seatIndex(player.acc.AccountId), player.acc.AccountId, player.bet)
	}
	/////////////////////// print //////////////////////////////////////////////

	C = self.diu
	// 如果有死皮
	if s {
		B = self.continues
		A = self.qiao
	} else {
		A = append(self.continues, self.qiao...)
	}
	shows = A

	// 找出A中最大下注金额
	max_bet_A := int64(0)
	for _, tPlayer := range A {
		if tPlayer.bet > max_bet_A {
			max_bet_A = tPlayer.bet
		}
	}
	// A中的人排序
	so := &Settlement_sort{R: self.Room, S: A, Ht: 0}
	sort.Sort(so)
	max_head = int8(self.seatIndex(so.S[0].acc.AccountId))
	self.track_log(colorized.White("最大 头 的座位号:[%v]"), max_head)

	so.Ht = 1
	sort.Sort(so)
	A = so.S

	// 尾牌最大的玩家座位号
	max_tail = int8(self.seatIndex(so.S[0].acc.AccountId))
	self.track_log(colorized.White("最大 尾 的座位号:[%v]"), max_tail)

	// B中的人 每人输max_bet_A
	for _, playerB := range B {
		decrease := max_bet_A
		for _, playerA := range A {
			increaseA := playerA.bet
			if decrease >= increaseA {
				decrease -= increaseA
				playerB.bet -= increaseA
				playerA.bobo += increaseA
			} else {
				playerB.bet -= decrease
				playerA.bobo += decrease
				decrease = 0
				break
			}
		}
	}
	// C中的人 每人输max_bet_A
	for _, playerC := range C {
		decrease := playerC.bet
		if decrease > max_bet_A {
			decrease = max_bet_A
		}

		for _, playerA := range A {
			increaseA := playerA.bet

			if decrease >= increaseA {
				decrease -= increaseA
				playerC.bet -= increaseA
				playerA.bobo += increaseA
			} else {
				playerC.bet -= decrease
				playerA.bobo += decrease
				decrease = 0
				break
			}
		}
	}

	// 参与分牌的人两两比
	if len(A) > 1 {
		m := make(map[uint8][]*GamePlayer)
		for _, player := range A {
			m[uint8(self.seatIndex(player.acc.AccountId))] = make([]*GamePlayer, 0)
		}

		// 找出每个人的所有赢家
		j := 1
		for _, player := range A {
			for i := j; i < len(A); i++ {
				compareObj := A[i]
				pi := uint8(self.seatIndex(player.acc.AccountId))
				oi := uint8(self.seatIndex(compareObj.acc.AccountId))
				ret := algorithm.CompareTouWei(
					player.cards,
					algorithm.CalcFromBankerPositionWeight(uint8(self.lastBanker_index), pi),
					compareObj.cards,
					algorithm.CalcFromBankerPositionWeight(uint8(self.lastBanker_index), oi),
					0,
					true)

				if ret == 1 {
					m[oi] = append(m[oi], player)
					self.track_log(colorized.White("%v号位 大于 %v号位"), pi, oi)
				} else if ret == 2 {
					m[pi] = append(m[pi], compareObj)
					self.track_log(colorized.White("%v号位 大于 %v号位"), oi, pi)
				} else {
					self.track_log(colorized.White("%v号位 和 %v号位 打走"), oi, pi)
				}
			}
			j++
		}

		temp_calculation := make(map[uint8]int64)
		// 增对每个人的赢家，排序
		for index, set := range m {
			so.S = set
			so.Ht = 1
			sort.Sort(so)

			lose_player := self.seats[index]
			lose_bet := lose_player.bet

			// 找出赢家里，喊话最多的值
			max_bet_S := int64(0)
			for _, tPlayer := range so.S {
				if tPlayer.bet > max_bet_S {
					max_bet_S = tPlayer.bet
				}
			}
			if lose_bet > max_bet_S {
				lose_bet = max_bet_S
			}
			save_val := lose_player.bet - lose_bet
			// 赢家，按照大->小的优先顺序瓜分奖励
			for _, win_player := range so.S {
				award_val := win_player.bet
				if lose_bet > award_val {
					win_player.bobo += award_val
					lose_bet -= award_val
				} else {
					win_player.bobo += lose_bet
					lose_bet = 0
					break
				}
			}
			temp_calculation[index] = save_val + lose_bet
		}

		for index, val := range temp_calculation {
			self.seats[index].bet = val
		}
	}

	// 吃芒果分 尾最大的全部赢
	mango_winer := A[0]
	total := int64(0)
	for _, player := range self.seats {
		if player != nil && player.status == types.EGameStatus_PLAYING {
			tail_extra = player.mangoVal
			total += player.mangoVal
			player.mangoVal = 0
		}
	}
	mango_winer.bobo += total
	self.track_log(colorized.White("赢得芒果分的人 Accid:[%v] 座位号:[%v], total:[%v]"), mango_winer.acc.AccountId, self.seatIndex(mango_winer.acc.AccountId), total)
	return
}
func (self *settlement) Tick(now int64) {
	if now >= self.timestamp {
		self.switchStatus(now, types.ERoomStatus_WAITING)
		return
	}
}

func (self *settlement) Leave(now int64) {
	if self.clear_xiumang == true {
		self.mangoCount = 0
	}
	for index, player := range self.seats {
		if player != nil {
			self.track_log(colorized.White(" 玩家:[%v] 座位号:[%v] 簸簸里的钱:[%v] "), player.acc.AccountId, index, player.bobo)
		}
	}
	self.track_log(colorized.White("settlement leave\n"))
}

// 当前状态下，玩家是否可以退出
func (self *settlement) CanQuit(accId uint32) bool {
	return self.canQuit(accId)
}

func (self *settlement) ShowCard(player *GamePlayer, show_self bool) packet.IPacket {
	pack := packet.NewPacket(nil)
	tempcount := uint16(0)
	temp := packet.NewPacket(nil)
	for i := 0; i < self.show_count; i++ {
		if i >= int(player.showcards) {
			break
		}
		tempcount++
		if i <= 1 && !show_self {
			temp.WriteUInt8(0)
			temp.WriteUInt8(0)
		} else {
			if i >= len(player.cards) {
				log.Error("越界 i:%v len(player.cards):%v self.show_count:%v player.showcards:%v", i, len(player.cards), self.show_count, player.showcards)
				break
			}
			temp.WriteUInt8(player.cards[i][0])
			temp.WriteUInt8(player.cards[i][1])
		}
	}
	pack.WriteUInt16(tempcount)
	pack.CatBody(temp)
	return pack
}

//
func (self *settlement) CombineMSG(pack packet.IPacket, acc *account.Account) {
	pack.CatBody(self.settlement_pack)
	pack.CatBody(self.bouns_pack)
}

func (self *settlement) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_CX_FORCE_WATCH_CARDS_REQ.UInt16(): // 请求强制看牌
		self.Old_MSGID_CX_FORCE_WATCH_CARDS_REQ(actor, msg, session)

	case protomsg.Old_MSGID_CX_FORCE_WATCH_NEXT_CARDS_REQ.UInt16(): // 请求强制查看下张牌
		self.Old_MSGID_CX_FORCE_WATCH_NEXT_CARDS_REQ(actor, msg, session)

	default:
		self.track_log(colorized.White("settlement 状态 没有处理消息msgId:%v"), pack.GetMsgID())
		return false
	}

	return true
}

// 请求强制看牌
func (self *settlement) Old_MSGID_CX_FORCE_WATCH_CARDS_REQ(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountID := pack.ReadUInt32()

	if self.force_watch[accountID] {
		return
	}
	index := self.seatIndex(accountID)
	if index == -1 {
		log.Errorf("玩家不在座位上")
		return
	}

	player := self.seats[index]
	if player.status != types.EGameStatus_PLAYING {
		log.Warnf("玩家没有参与游戏 ：accid:%v index%v status:%v", player.acc.AccountId, index, player.status)
		return
	}

	result := packet.NewPacket(nil)
	result.SetMsgID(protomsg.Old_MSGID_CX_FORCE_WATCH_CARDS_REQ.UInt16())

	confstr := config.GetPublicConfig_String(fmt.Sprintf("EXPEND_") + strconv.Itoa(int(self.matchType)))
	arrInt := utils.SplitConf2ArrInt32(confstr, ",")
	spend_val := arrInt[0]
	if player.acc.GetMoney() < uint64(spend_val) {
		result.WriteUInt8(1)
		send_tools.Send2Account(result.GetData(), session)
		return
	}

	if player.acc.GetMoney() >= uint64(spend_val) {
		player.acc.AddMoney(-int64(spend_val), 0, common.EOperateType_DIVVIDEND)
	}

	result.WriteUInt8(0)
	result.WriteUInt8(uint8(index + 1))
	self.SendBroadcast(result.GetData())
	self.update_bet_bobo_mango(player.acc.AccountId)

	self.force_watch[player.acc.AccountId] = true

	// 强制亮牌
	send2other_all := packet.NewPacket(nil)
	send2other_all.SetMsgID(protomsg.Old_MSGID_CX_FORCE_WATCH_CARDS.UInt16())
	send2other_all.WriteUInt16(1)
	send2other_all.WriteUInt8(uint8(index + 1))
	send2other_all.WriteUInt8(player.cards[0][0])
	send2other_all.WriteUInt8(player.cards[0][1])
	send2other_all.WriteUInt8(player.cards[1][0])
	send2other_all.WriteUInt8(player.cards[1][1])
	for _, acc := range self.accounts {
		if acc.Robot == 0 && acc.SessionId > 0 && acc.AccountId != player.acc.AccountId {
			send_tools.Send2Account(send2other_all.GetData(), acc.SessionId)
		}
	}

	send2dest := packet.NewPacket(nil)
	send2dest.SetMsgID(protomsg.Old_MSGID_CX_FORCE_WATCH_CARDS.UInt16())

	temp := packet.NewPacket(nil)
	count := uint16(0)
	for _, player := range self.seats {
		if player != nil && player.status == types.EGameStatus_PLAYING {
			count++
			temp.WriteUInt8(uint8(self.seatIndex(player.acc.AccountId) + 1))
			temp.WriteUInt8(player.cards[0][0])
			temp.WriteUInt8(player.cards[0][1])
			temp.WriteUInt8(player.cards[1][0])
			temp.WriteUInt8(player.cards[1][1])
		}
	}
	send2dest.WriteUInt16(count)
	send2dest.CatBody(temp)
	send_tools.Send2Account(send2dest.GetData(), session)

	// 延长结算时间
	self.timestamp += config.GetPublicConfig_Int64("SETTLEMENT_TIME_FORCE_WATCH")
}

// 请求强制看下一张
func (self *settlement) Old_MSGID_CX_FORCE_WATCH_NEXT_CARDS_REQ(actor int32, msg []byte, session int64) {
	if self.next_card >= 3 {
		return
	}
	pack := packet.NewPacket(msg)
	accountID := pack.ReadUInt32()

	index := self.seatIndex(accountID)
	if index == -1 {
		log.Errorf("玩家不在座位上")
		return
	}

	player := self.seats[index]
	if player.status != types.EGameStatus_PLAYING {
		log.Warnf("玩家没有参与游戏 ：accid:%v index%v status:%v", player.acc.AccountId, index, player.status)
		return
	}

	result := packet.NewPacket(nil)
	result.SetMsgID(protomsg.Old_MSGID_CX_FORCE_WATCH_NEXT_CARDS_REQ.UInt16())

	confstr := config.GetPublicConfig_String(fmt.Sprintf("EXPEND_") + strconv.Itoa(int(self.matchType)))
	arrInt := utils.SplitConf2ArrInt32(confstr, ",")
	spend_val := arrInt[1]
	if player.acc.GetMoney() < uint64(spend_val) {
		result.WriteUInt8(1)
		send_tools.Send2Account(result.GetData(), session)
		return
	}

	if player.acc.GetMoney() >= uint64(spend_val) {
		player.acc.AddMoney(-int64(spend_val), 0, common.EOperateType_DIVVIDEND)
	}

	result.WriteUInt8(0)
	result.WriteUInt8(uint8(index + 1))
	self.SendBroadcast(result.GetData())
	self.update_bet_bobo_mango(player.acc.AccountId)

	// 看下一张牌
	send2other_all := packet.NewPacket(nil)
	send2other_all.SetMsgID(protomsg.Old_MSGID_CX_FORCE_WATCH_NEXT_CARDS.UInt16())
	send2other_all.WriteUInt8(self.next3cards[self.next_card][0])
	send2other_all.WriteUInt8(self.next3cards[self.next_card][1])

	self.SendBroadcast(send2other_all.GetData())
	self.next_card++

	// 延长结算时间
	self.timestamp += config.GetPublicConfig_Int64("SETTLEMENT_TIME_FORCE_NEXT")
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (self *Settlement_sort) Len() int {
	return len(self.S)
}
func (self *Settlement_sort) Less(i, j int) bool {
	headi, taili, _ := algorithm.CalcOnePlayerCardType(self.S[i].cards, 0, true)
	headj, tailj, _ := algorithm.CalcOnePlayerCardType(self.S[j].cards, 0, true)
	ii := headi
	jj := headj
	if self.Ht == 1 {
		ii = taili
		jj = tailj
	}
	ret := algorithm.CompareCardSet(ii, jj)

	if ret == 0 {
		iw := algorithm.CalcFromBankerPositionWeight(uint8(self.R.lastBanker_index), uint8(self.R.seatIndex(self.S[i].acc.AccountId)))
		jw := algorithm.CalcFromBankerPositionWeight(uint8(self.R.lastBanker_index), uint8(self.R.seatIndex(self.S[j].acc.AccountId)))
		return iw < jw
	} else {
		return ret == 1
	}

}
func (self *Settlement_sort) Swap(i, j int) {
	self.S[i], self.S[j] = self.S[j], self.S[i]
}
