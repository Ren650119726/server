package room

import (
	"root/common"
	"root/common/algorithm"
	"root/common/config"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"math"
	"root/protomsg"
	"root/server/paodekuai/account"
	"root/server/paodekuai/send_tools"
	"root/server/paodekuai/types"
)

type (
	playing struct {
		*Room
		next_op_index  uint8  // 下一个操作的玩家下标
		next_op_time   int64  // 下次操作时间戳, 单位: 毫秒
		next_op_tion   uint8  // 下一个操作方式
		need_guanpai   uint8  // 关牌标记
		last_big_index uint8  // 最后一个最大的玩家
		out_card_count uint16 // 出牌次数计数
	}
)

func (self *playing) BulidPacket(tPacket packet.IPacket, tAccount *account.Account) {
	tPacket.WriteInt64(self.next_op_time)
	tPacket.WriteUInt8(self.next_op_index + 1)
	tPacket.WriteUInt8(self.next_op_tion)
}

func (self *playing) Enter(now int64) {
	self.all_cards = algorithm.PaoDeKuai_DY_ShuffleCard(2)
	self.track_log(colorized.Yellow("RoomID:%v playing enter"), self.roomId)

	self.banker_index = math.MaxUint8
	self.next_op_index = math.MaxUint8
	self.last_big_index = math.MaxUint8
	self.next_op_time = 0
	self.need_guanpai = 0
	self.out_card_count = 0

	// 每个人发10张牌
	for i := uint8(0); i < self.max_count; i++ {
		tPlayer := self.seats[i]
		tPlayer.hand_cards = self.all_cards[:algorithm.PDK_DY_MAX_CARDS]
		algorithm.Poker_SortCard(tPlayer.hand_cards)

		self.all_cards = self.all_cards[algorithm.PDK_DY_MAX_CARDS:]
		self.track_log(colorized.Cyan("房间ID:%v, 第%v局 发牌, 玩家:%v 下标:%v 第%v把 手牌:%v"), self.roomId, self.games, tPlayer.acc.AccountId, i, tPlayer.acc.Games, tPlayer.hand_cards)

		// 确定庄家下标
		if self.banker_index == math.MaxUint8 {
			isHaveFirstCard := algorithm.Poker_IsHaveCard(tPlayer.hand_cards, algorithm.PDK_DY_FIRST_CARD)
			if isHaveFirstCard == true {
				self.banker_index = i
				self.track_log(colorized.Cyan("房间ID:%v, 第%v局 庄家, 玩家:%v 下标:%v"), self.roomId, self.games, tPlayer.acc.AccountId, i)
			}
		}

		// 判断是否有可关牌的玩家
		if self.next_op_index == math.MaxUint8 {
			isGuanPai := algorithm.PaoDeKuai_DY_CanGuanPai(tPlayer.hand_cards)
			if isGuanPai == true {
				self.need_guanpai = 1
				self.next_op_index = i
				self.next_op_tion = OP_GUAN_CARD
				self.track_log(colorized.Cyan("房间ID:%v, 第%v局 可关牌, 玩家:%v 下标:%v"), self.roomId, self.games, tPlayer.acc.AccountId, i)
			}
		}
	}

	// 如果没有可关牌的玩家, 下个操作者是庄家
	if self.next_op_index == math.MaxUint8 {
		self.next_op_index = self.banker_index
		self.last_big_index = self.banker_index
		self.track_log(colorized.Cyan("房间ID:%v, 第%v局 确定第一手出牌, 下标:%v"), self.roomId, self.games, self.next_op_index)
	} else {
		self.track_log(colorized.Cyan("房间ID:%v, 第%v局 确定第一手关牌操作, 下标:%v"), self.roomId, self.games, self.next_op_index)
	}

	// 广播发牌
	_, nChouShui, nServerFee := self.calc_fee()

	tFee := packet.NewPacket(nil)
	tFee.SetMsgID(protomsg.Old_MSGID_UPDATE_SERVICE_FEE.UInt16())
	tFee.WriteUInt8(uint8(self.gameType))
	tFee.WriteUInt32(uint32(self.roomId))
	tFee.WriteUInt16(uint16(self.max_count))
	for _, tPlayer := range self.seats {
		tFee.WriteUInt32(tPlayer.acc.AccountId)
		tFee.WriteUInt32(uint32(nServerFee))
		if tPlayer.acc.GetMoney() >= uint64(nChouShui) {
			tPlayer.acc.AddMoney(-nChouShui, 0, common.EOperateType_SERVICE_FEE)
			self.track_log("房间ID:%v, 第%v局 扣抽水, 玩家:%v 抽水:%v 服务费:%v", self.roomId, self.games, tPlayer.acc.AccountId, nChouShui, nServerFee)
		} else {
			log.Warnf("Error: !玩家ID:%v 开局扣除抽水费用:%v不足, 身上金额:%v", tPlayer.acc.AccountId, nChouShui, tPlayer.acc.GetMoney())
		}
	}
	send_tools.Send2Hall(tFee.GetData())

	nFiringTime := config.GetPublicConfig_Int64("PDK_FIRING_CARD_TIME") * 1000
	self.next_op_time = utils.MilliSecondTimeSince1970() + nFiringTime
	mSeat := make(map[uint32]bool)
	for _, tPlayer := range self.seats {
		tSitDownFiring := packet.NewPacket(nil)
		tSitDownFiring.SetMsgID(protomsg.Old_MSGID_PDK_GAME_FIRING_CARD.UInt16())
		tSitDownFiring.WriteInt64(self.next_op_time)
		tSitDownFiring.WriteUInt32(uint32(nChouShui))
		tSitDownFiring.WriteUInt8(self.banker_index + 1)
		tSitDownFiring.WriteUInt16(algorithm.PDK_DY_MAX_CARDS)
		for i := 0; i < algorithm.PDK_DY_MAX_CARDS; i++ {
			card := tPlayer.hand_cards[i]
			tSitDownFiring.WriteUInt8(card[0])
			tSitDownFiring.WriteUInt8(card[1])
		}
		send_tools.Send2Account(tSitDownFiring.GetData(), tPlayer.acc.SessionId)
		mSeat[tPlayer.acc.AccountId] = true
	}

	tWatchFiring := packet.NewPacket(nil)
	tWatchFiring.SetMsgID(protomsg.Old_MSGID_PDK_GAME_FIRING_CARD.UInt16())
	tWatchFiring.WriteInt64(self.next_op_time)
	tWatchFiring.WriteUInt32(uint32(nChouShui))
	tWatchFiring.WriteUInt8(self.banker_index + 1)
	tWatchFiring.WriteUInt16(0)
	for nAccountID, tAccount := range self.accounts {
		if _, isExist := mSeat[nAccountID]; isExist == false {
			send_tools.Send2Account(tWatchFiring.GetData(), tAccount.SessionId)
		}
	}
}

func (self *playing) Tick(now int64) {

	nNowTime := utils.MilliSecondTimeSince1970()

	if self.next_op_time > 0 && nNowTime >= self.next_op_time {
		if self.need_guanpai == 1 {
			// 广播开始关牌操作
			nGuanCardTime := config.GetPublicConfig_Int64("PDK_GUAN_CARD_TIME") * 1000
			self.next_op_time = nNowTime + nGuanCardTime
			self.need_guanpai = 2
			self.next_op_tion = OP_GUAN_CARD
			self.broadcast_next_op(self.next_op_time, 0, self.next_op_index, OP_GUAN_CARD)
			self.track_log(colorized.Cyan("房间ID:%v, 第%v局 通知开始关牌操作"), self.roomId, self.games)

		} else if self.need_guanpai == 2 {
			// 关牌操作超时, 默认不关牌
			nOutCardTime := config.GetPublicConfig_Int64("PDK_OUT_CARD_TIME") + 1000
			self.next_op_time = nNowTime + nOutCardTime
			self.next_op_index = self.banker_index
			self.last_big_index = self.banker_index
			self.need_guanpai = 0
			self.next_op_tion = OP_OUT_CARD
			self.broadcast_next_op(self.next_op_time, 0, self.next_op_index, OP_OUT_CARD)
			self.track_log(colorized.Cyan("房间ID:%v, 第%v局 通知关牌操作超时, 默认不关牌"), self.roomId, self.games)

		} else if self.need_guanpai == 0 {
			self.need_guanpai = 4
			nOutCardTime := config.GetPublicConfig_Int64("PDK_OUT_CARD_TIME") * 1000
			self.next_op_time = nNowTime + nOutCardTime
			self.next_op_index = self.banker_index
			self.last_big_index = self.banker_index
			self.next_op_tion = OP_OUT_CARD
			self.broadcast_next_op(self.next_op_time, 0, self.next_op_index, OP_OUT_CARD)
			self.track_log(colorized.Cyan("房间ID:%v, 第%v局 预判第一手出牌, 下次操作下标:%v"), self.roomId, self.games, self.next_op_index)

		} else {

			tPlayer := self.seats[self.next_op_index]
			nHandCardLen := len(tPlayer.hand_cards)
			next_idx := self.next_index(self.next_op_index)

			if tPlayer.op == OP_NIL && nHandCardLen > 0 {
				// 未结束本局, 当前操作者超时默认操作
				if self.out_card_count == 0 {
					// 第一次出牌, 必须包含首张牌
					sRemove := []common.Card_info{algorithm.PDK_DY_FIRST_CARD}
					self.track_log(colorized.Cyan("房间ID:%v, 第%v局 第一手出牌超时, 玩家:%v, 手牌:%v 默认出牌:%v"), self.roomId, self.games, tPlayer.acc.AccountId, tPlayer.hand_cards, sRemove)
					tPlayer.hand_cards = algorithm.Poker_RemoveCard(tPlayer.hand_cards, algorithm.PDK_DY_FIRST_CARD)
					tPlayer.last_out_card = sRemove
					self.last_big_index = self.next_op_index
					self.last_out = sRemove
					self.out_card_count++
					self.broadcast_out_card(self.next_op_index, sRemove)

				} else {

					// 找到比上次出牌大的牌出
					var sRemove []common.Card_info
					if self.last_out == nil {
						// 先判断下家是否报单
						tNextPlayer := self.seats[next_idx]
						if len(tNextPlayer.hand_cards) == 1 {
							// 下家保单, 取最大单张出
							sRemove = []common.Card_info{tPlayer.hand_cards[0]}
						} else {
							// 下家未保单, 取最小单张出
							sRemove = []common.Card_info{tPlayer.hand_cards[nHandCardLen-1]}
						}
					} else {
						// 先判断下家是否报单
						isNeedBigToSamll := false
						if len(self.last_out) == 1 {
							tNextPlayer := self.seats[next_idx]
							if len(tNextPlayer.hand_cards) == 1 {
								isNeedBigToSamll = true
							}
						}
						sRemove = algorithm.PaoDeKuai_DY_GetBigCard(tPlayer.hand_cards, self.last_out, isNeedBigToSamll)
					}
					if sRemove != nil {
						self.track_log(colorized.Cyan("房间ID:%v, 第%v局 出牌超时, 玩家:%v, 手牌:%v, 默认出牌:%v"), self.roomId, self.games, tPlayer.acc.AccountId, tPlayer.hand_cards, sRemove)
						tPlayer.hand_cards = algorithm.Poker_RemoveCard(tPlayer.hand_cards, sRemove)
						tPlayer.last_out_card = sRemove
						self.last_big_index = self.next_op_index
						self.last_out = sRemove
						self.out_card_count++
						self.broadcast_out_card(self.next_op_index, sRemove)

						nRemoveType := algorithm.PaoDeKuai_DY_CalcCardType(sRemove)
						if nRemoveType == common.PDK_ZHA_DAN {
							tPlayer.bomb_out_count++
						}
					} else {
						log.Warnf("!找不到合适的牌, 玩家ID:%v 手牌:%v, 最后出牌:%v", tPlayer.acc.AccountId, tPlayer.hand_cards, self.last_out)
					}
				}
			}

			// 判断是否结束本局
			nHandCardLen = len(tPlayer.hand_cards)
			if nHandCardLen <= 0 {
				self.next_op_time = 0
				nWaitSettlementTime := config.GetPublicConfig_Int64("PDK_WAIT_SETTLEMENT_TIME") * 1000
				self.owner.AddTimer(nWaitSettlementTime, 1, func(dt int64) {
					self.switch_room_status(now, types.ERoomStatus_SETTLEMENT)
				})
				return
			}

			// 判断是否都要不起
			nPassCount := uint8(0)
			for _, tPlayer := range self.seats {
				if tPlayer.op == OP_PASS {
					nPassCount++
				}
			}
			if nPassCount >= (self.max_count - 1) {
				// 都要不起, 该玩家继续操作
				for _, tPlayer := range self.seats {
					tPlayer.op = OP_NIL
					tPlayer.last_out_card = nil
				}
				self.last_out = nil

				nOutCardTime := config.GetPublicConfig_Int64("PDK_OUT_CARD_TIME") * 1000
				self.next_op_time = nNowTime + nOutCardTime
				self.next_op_tion = OP_OUT_CARD
				self.next_op_index = self.last_big_index
				self.broadcast_next_op(self.next_op_time, 1, self.next_op_index, OP_OUT_CARD)
				self.track_log(colorized.Cyan("房间ID:%v, 第%v局 预判 %v家要不起, 新一轮出牌, 下次出牌下标:%v"), self.roomId, self.games, nPassCount, self.next_op_index)

			} else {
				// 继续轮询
				self.next_op_index = next_idx
				tNext := self.seats[next_idx]
				isHaveBigger := algorithm.PaoDeKuai_DY_IsHaveBiggerCard(tNext.hand_cards, self.last_out)
				if isHaveBigger == false {
					tNext.op = OP_PASS
					tNext.last_out_card = nil
					self.next_op_time = nNowTime + 1000
					self.next_op_tion = OP_PASS
					self.broadcast_next_op(self.next_op_time, 0, self.next_op_index, OP_PASS)
					self.track_log(colorized.Cyan("房间ID:%v, 第%v局 预判 下家%v要不起, 手牌:%v, 下次操作下标:%v"), self.roomId, self.games, tNext.acc.AccountId, tNext.hand_cards, self.next_op_index)

				} else {
					tNext.op = OP_NIL
					nOutCardTime := config.GetPublicConfig_Int64("PDK_OUT_CARD_TIME") * 1000
					self.next_op_time = nNowTime + nOutCardTime
					self.next_op_tion = OP_OUT_CARD
					self.broadcast_next_op(self.next_op_time, 0, self.next_op_index, OP_OUT_CARD)
					self.track_log(colorized.Cyan("房间ID:%v, 第%v局 预判 下家%v要得起, 下次操作下标:%v"), self.roomId, self.games, tNext.acc.AccountId, self.next_op_index)
				}
			}
		}
	}
}

func (self *playing) Leave(now int64) {
	self.track_log(colorized.Yellow("RoomID:%v playing leave\n"), self.roomId)
}

///////////////////////////////// handler ///////////////////////////////////////////////////
func (self *playing) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_PDK_DO_OPTION.UInt16(): // 操作
		self.Old_MSGID_PDK_DO_OPTION(actor, msg, session)
	default:
		log.Warnf("playing 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}
	return true
}

// 客户端响应操作
func (self *playing) Old_MSGID_PDK_DO_OPTION(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountID := pack.ReadUInt32()
	nOpType := pack.ReadUInt8()

	account.CheckSession(accountID, session)

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_PDK_DO_OPTION.UInt16())
	acc := account.AccountMgr.GetAccountByID(accountID)
	if acc == nil {
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		log.Warnf("!玩家没在游戏中 :%v!", accountID)
		return
	}

	acc = self.accounts[accountID]
	if acc == nil {
		send.WriteUInt8(11)
		send_tools.Send2Account(send.GetData(), session)
		log.Warnf("!玩家没在房间内中 :%v!", accountID)
		return
	}

	index := self.get_seat_index(accountID)
	if index > self.max_count {
		send.WriteUInt8(3)
		send_tools.Send2Account(send.GetData(), session)
		log.Warnf("!玩家没在座位上 :%v!", accountID)
		return
	}

	tPlayer := self.seats[index]
	if tPlayer.status != types.EGameStatus_READY {
		send.WriteUInt8(4)
		send_tools.Send2Account(send.GetData(), session)
		log.Warnf("!玩家状态不对 :%v 玩家状态:%v!", accountID, tPlayer.status)
		return
	}

	switch nOpType {
	case OP_OUT_CARD:
		nRemoveLen := pack.ReadUInt16()
		sRemove := make([]common.Card_info, 0, nRemoveLen)
		for i := uint16(0); i < nRemoveLen; i++ {
			nColor := pack.ReadUInt8()
			nPoint := pack.ReadUInt8()
			sRemove = append(sRemove, common.Card_info{nColor, nPoint})
		}

		if len(sRemove) <= 0 {
			send.WriteUInt8(10)
			send_tools.Send2Account(send.GetData(), session)
			log.Warnf("!玩家ID:%v 没有找到准备要出的牌:%v!", accountID, sRemove)
			return
		}

		if self.out_card_count == 0 {
			isFirstCard := algorithm.Poker_IsHaveCard(sRemove, algorithm.PDK_DY_FIRST_CARD)
			if isFirstCard == false {
				send.WriteUInt8(12)
				send_tools.Send2Account(send.GetData(), session)
				log.Warnf("!玩家ID:%v 第一手出牌没包含%v, 准备出牌:%v!", accountID, algorithm.PDK_DY_FIRST_CARD, sRemove)
				return
			}
		}

		nRemoveType := algorithm.PaoDeKuai_DY_CalcCardType(sRemove)
		if common.PDK_NIL == nRemoveType {
			send.WriteUInt8(13)
			send_tools.Send2Account(send.GetData(), session)
			log.Warnf("!玩家ID:%v 出牌类型异常, 准备出牌:%v!", accountID, sRemove)
			return
		}

		isBigger := true
		if self.last_out != nil {
			isBigger = algorithm.PaoDeKuai_DY_OneIsBiggerTwo(sRemove, self.last_out)
		}
		if isBigger == false {
			send.WriteUInt8(14)
			send_tools.Send2Account(send.GetData(), session)
			log.Warnf("!玩家ID:%v 准备出的牌:%v 比最后出的牌小:%v!", accountID, sRemove, self.last_out)
			return
		}

		// 判断下家是否报单
		if nRemoveLen == 1 {
			nNextIndex := self.next_index(self.next_op_index)
			tNextPlayer := self.seats[nNextIndex]
			if len(tNextPlayer.hand_cards) == 1 {
				isMaxBigger := algorithm.PaoDeKuai_IsMaxSingleCard(tPlayer.hand_cards, sRemove[0])
				if isMaxBigger == false {
					send.WriteUInt8(15)
					send_tools.Send2Account(send.GetData(), session)
					log.Warnf("!玩家ID:%v 下家报单, 准备出的牌:%v 不是最大的单张:%v!", accountID, sRemove, tPlayer.hand_cards)
					return
				}
			}
		}

		if nRemoveType == common.PDK_ZHA_DAN {
			tPlayer.bomb_out_count++
		}

		self.track_log(colorized.Magenta("房间ID:%v, 第%v局 玩家:[%v], 座位号:[%v], 手牌:%v, 本次出牌:%v"), self.roomId, self.games, tPlayer.acc.AccountId, index, tPlayer.hand_cards, sRemove)
		tPlayer.hand_cards = algorithm.Poker_RemoveCard(tPlayer.hand_cards, sRemove)
		tPlayer.last_out_card = sRemove
		tPlayer.op = OP_OUT_CARD
		self.last_out = sRemove
		self.last_big_index = self.next_op_index
		self.out_card_count++
		self.broadcast_out_card(self.next_op_index, sRemove)
		self.next_op_time = utils.MilliSecondTimeSince1970()

	case OP_GUAN_CARD:
		nGuanPai := pack.ReadUInt8()
		self.track_log(colorized.Magenta("房间ID:%v, 第%v局 玩家:[%v], 座位号:[%v], 关牌:[%v]"), self.roomId, self.games, tPlayer.acc.AccountId, index, nGuanPai)
		if nGuanPai == 1 {
			tSend := packet.NewPacket(nil)
			tSend.SetMsgID(protomsg.Old_MSGID_PDK_DO_OPTION.UInt16())
			tSend.WriteUInt8(0)
			tSend.WriteUInt8(OP_GUAN_CARD)
			tSend.WriteUInt8(index + 1)
			tSend.WriteUInt8(nGuanPai)
			self.SendBroadcast(tSend.GetData())

			tPlayer.op = OP_GUAN_CARD
			self.next_op_time = 0
			nWaitSettlementTime := config.GetPublicConfig_Int64("PDK_WAIT_SETTLEMENT_TIME") * 1000
			self.owner.AddTimer(nWaitSettlementTime, 1, func(dt int64) {
				self.switch_room_status(0, types.ERoomStatus_SETTLEMENT)
			})
			return
		}
		self.next_op_time = utils.MilliSecondTimeSince1970()
		self.need_guanpai = 0

	default:
		send.WriteUInt8(5)
		send_tools.Send2Account(send.GetData(), session)
		log.Warnf("!帐号:%v 操作类型异常 :%v!", accountID, nOpType)
		return
	}
}
