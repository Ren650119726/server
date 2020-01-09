package room

import (
	"root/common"
	"root/common/algorithm"
	"root/common/config"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/paodekuai/account"
	"root/server/paodekuai/send_tools"
	"root/server/paodekuai/types"
)

type (
	settlement struct {
		*Room
		timestamp  int64 // 单位: 毫秒
		settlement packet.IPacket
	}
)

func (self *settlement) BulidPacket(tPacket packet.IPacket, tAccount *account.Account) {
	tPacket.CatBody(self.settlement)
}

func (self *settlement) Enter(now int64) {
	self.track_log(colorized.Yellow("RoomID:%v settlement enter"), self.roomId)

	mLoserValue := make(map[uint32]int64)
	var tWiner *GamePlayer
	var nWinerIndex uint8
	nTotalBombOutCount := uint8(0)
	nTotalBombBei := uint8(1)
	nTotalLoserCardLen := uint8(0)
	nWinTotal := int64(0)

	for i := uint8(0); i < self.max_count; i++ {
		tPlayer := self.seats[i]
		nTotalBombOutCount += tPlayer.bomb_out_count

		nCardLen := uint8(len(tPlayer.hand_cards))
		if nCardLen == 0 || tPlayer.op == OP_GUAN_CARD {
			tWiner = tPlayer
			nWinerIndex = i
		} else {
			nTotalLoserCardLen += nCardLen
		}
	}

	// 正常结算
	nSettlementType := uint8(1)
	if tWiner.op == OP_GUAN_CARD {
		// 关牌结算
		nSettlementType = 2
	}

	if nTotalBombOutCount > self.bomb_limit {
		nTotalBombOutCount = self.bomb_limit
	}
	if nTotalBombOutCount == 1 {
		nTotalBombBei = 2
	} else if nTotalBombOutCount == 2 {
		nTotalBombBei = 4
	} else if nTotalBombOutCount == 3 {
		nTotalBombBei = 8
	} else {
		nTotalBombBei = 1
	}

	nLoser := int64(0)
	for i := uint8(0); i < self.max_count; i++ {
		tPlayer := self.seats[i]
		nLoser = 0
		if nWinerIndex != i {
			nCardLen := uint8(len(tPlayer.hand_cards))
			if nCardLen >= algorithm.PDK_DY_MAX_CARDS {
				nLoser = self.bet * int64(nCardLen) * 2 * int64(nTotalBombBei)
				mLoserValue[tPlayer.acc.AccountId] = nLoser
			} else if nCardLen > 1 {
				nLoser = self.bet * int64(nCardLen) * int64(nTotalBombBei)
				mLoserValue[tPlayer.acc.AccountId] = nLoser
			}
			nWinTotal += nLoser
		}
	}

	nSettlementTime := config.GetPublicConfig_Int64("PDK_SETTLEMENT_TIME") * 1000
	self.timestamp = utils.MilliSecondTimeSince1970() + nSettlementTime

	// 判断是否中奖; 赢家出完牌+出了炸弹; 输家全部都没出牌
	nWinReward := int64(0)
	var tAward packet.IPacket
	AWARD_NAME := config.GetPublicConfig_String("PDK_AWARD_NAME")
	if self.clubID == 0 && tWiner != nil && tWiner.bomb_out_count > 0 && nTotalLoserCardLen >= algorithm.PDK_DY_MAX_CARDS*2 {
		// 如果有玩家中奖, 则额外增加几秒钟给客户端播放动画
		nRewardDuration := config.GetPublicConfig_Int64("PDK_REWARD_DURATION") * 1000
		self.timestamp += nRewardDuration

		nBonusPoolScale := config.GetPublicConfig_Int64("PDK_BONUS_POOL_SCALE")
		nBonusPoolValue := RoomMgr.Bonus[uint32(self.bet)]
		nWinReward = nBonusPoolValue * nBonusPoolScale / 100
		RoomMgr.AddBonusPool(uint32(self.bet), -nWinReward)
		self.track_log("房间ID:%v, 第%v局, 玩家:%v中奖, 中奖金额:%v", self.roomId, self.games, tWiner.acc.AccountId, nWinReward)

		RoomMgr.AddAwardHisotry(tWiner.acc.AccountId, tWiner.acc.Name, uint32(nWinReward), AWARD_NAME, uint32(self.bet))

		tAward = packet.NewPacket(nil)
		tAward.WriteUInt16(1)
		tAward.WriteUInt8(nWinerIndex + 1)
		tAward.WriteUInt32(uint32(nWinReward))
		tAward.WriteString(AWARD_NAME)
	}

	// 计算每局增加的奖金池金额
	if self.clubID == 0 {
		nAddBonusValue, _, _ := self.calc_fee()
		nAddBonusValue *= 3
		RoomMgr.AddBonusPool(uint32(self.bet), nAddBonusValue)
		self.track_log("房间ID:%v, 第%v局, 奖金池增加:%v", self.roomId, self.games, nAddBonusValue)
	}

	self.settlement = packet.NewPacket(nil)
	self.settlement.SetMsgID(protomsg.Old_MSGID_PDK_GAME_SETTLEMENT.UInt16())
	self.settlement.WriteInt64(self.timestamp)
	self.settlement.WriteUInt8(nSettlementType)
	self.settlement.WriteUInt16(uint16(self.max_count))
	for i := uint8(0); i < self.max_count; i++ {
		tPlayer := self.seats[i]

		self.settlement.WriteUInt8(i + 1)
		self.settlement.WriteUInt8(tPlayer.bomb_out_count)

		if nWinerIndex == i {
			if tPlayer.op == OP_GUAN_CARD {
				// 关牌玩家, 显示手牌
				nCardLen := uint8(len(tPlayer.hand_cards))
				self.settlement.WriteUInt16(uint16(nCardLen))
				for k := uint8(0); k < nCardLen; k++ {
					card := tPlayer.hand_cards[k]
					self.settlement.WriteUInt8(card[0])
					self.settlement.WriteUInt8(card[1])
				}
			} else {
				// 赢家显示最后一手出的牌
				nLastOutLen := uint16(len(self.last_out))
				self.settlement.WriteUInt16(nLastOutLen)
				for k := uint16(0); k < nLastOutLen; k++ {
					card := self.last_out[k]
					self.settlement.WriteUInt8(card[0])
					self.settlement.WriteUInt8(card[1])
				}
			}

			self.settlement.WriteInt64(nWinTotal)
			self.settlement.WriteUInt8(1) // 赢家标记, 赢金额可能是0, 输家都只剩余1张牌
			tPlayer.acc.Profit += nWinTotal
			tPlayer.acc.AddMoney(nWinTotal+nWinReward, 0, common.EOperateType_SETTLEMENT)
			self.track_log("房间ID:%v, 第%v局结算, 玩家:%v, 手牌数量:%v 炸弹:%v 赢:%v 中奖:%v; 累积盈利:%v", self.roomId, self.games, tPlayer.acc.AccountId, len(tPlayer.hand_cards), tPlayer.bomb_out_count, nWinTotal, nWinReward, tPlayer.acc.Profit)
		} else {
			// 输家显示剩余手牌
			nCardLen := uint8(len(tPlayer.hand_cards))
			self.settlement.WriteUInt16(uint16(nCardLen))
			for k := uint8(0); k < nCardLen; k++ {
				card := tPlayer.hand_cards[k]
				self.settlement.WriteUInt8(card[0])
				self.settlement.WriteUInt8(card[1])
			}

			iChange := mLoserValue[tPlayer.acc.AccountId]
			self.settlement.WriteInt64(-iChange)
			self.settlement.WriteUInt8(0) // 输家标记
			tPlayer.acc.Profit -= iChange
			if tPlayer.acc.GetMoney() >= uint64(iChange) {
				tPlayer.acc.AddMoney(-iChange, 0, common.EOperateType_SETTLEMENT)
			} else {
				log.Warnf("!玩家ID:%v 结算扣钱:%v不足, 身上金额:%v", tPlayer.acc.AccountId, iChange, tPlayer.acc.GetMoney())
			}
			self.track_log("房间ID:%v, 第%v局结算, 玩家:%v, 手牌数量:%v 炸弹:%v 输:%v; 累积盈利:%v", self.roomId, self.games, tPlayer.acc.AccountId, len(tPlayer.hand_cards), tPlayer.bomb_out_count, iChange, tPlayer.acc.Profit)
		}
	}

	// 组装中奖信息
	if tAward == nil {
		self.settlement.WriteUInt16(0)
	} else {
		self.settlement.CatBody(tAward)
	}

	// 广播结算消息
	self.SendBroadcast(self.settlement.GetData())

	updateAccount := packet.NewPacket(nil)
	updateAccount.SetMsgID(protomsg.Old_MSGID_UPDATE_ACCOUNT.UInt16())
	updateAccount.WriteUInt32(self.roomId)
	updateAccount.WriteUInt8(0)
	updateAccount.WriteUInt16(uint16(self.max_count))
	for i := uint8(0); i < self.max_count; i++ {
		tPlayer := self.seats[i]
		updateAccount.WriteUInt32(tPlayer.acc.AccountId)
		updateAccount.WriteInt64(int64(tPlayer.acc.GetMoney()))
		if tWiner.acc.AccountId == tPlayer.acc.AccountId {
			updateAccount.WriteInt64(nWinTotal + nWinReward)
			updateAccount.WriteString(AWARD_NAME)
		} else {
			iChange := mLoserValue[tPlayer.acc.AccountId]
			updateAccount.WriteInt64(-iChange)
			updateAccount.WriteString("")
		}
	}
	send_tools.Send2Hall(updateAccount.GetData())
}

func (self *settlement) Tick(now int64) {

	nNowTime := utils.MilliSecondTimeSince1970()
	if self.timestamp > 0 && nNowTime >= self.timestamp {
		self.switch_room_status(now, types.ERoomStatus_WAITING)
		return
	}
}

func (self *settlement) Leave(now int64) {

	for i := uint8(0); i < self.max_count; i++ {
		tPlayer := self.seats[i]
		if tPlayer.acc.GetMoney() < self.situp_limit {
			self.leave_seat(tPlayer.acc, i)
			self.track_log(colorized.Yellow("RoomID:%v settlement ID:%v 金额不足:%v, 离开座位:%v"), self.roomId, tPlayer.acc.AccountId, tPlayer.acc.GetMoney(), i)
		}
	}
	self.track_log(colorized.Yellow("RoomID:%v settlement leave\n"), self.roomId)
}

///////////////////////////////// handler ///////////////////////////////////////////////////
func (self *settlement) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	default:
		log.Warnf("settlement 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}
	return true
}
