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
	"root/server/red2black/account"
	"root/server/red2black/algorithm"
	"root/server/red2black/event"
	"root/server/red2black/send_tools"
)

type (
	settlement struct {
		*Room
		s         ERoomStatus
		timestamp int64
		result    int8

		sendData   packet.IPacket
		masterData packet.IPacket
		changeVal  map[uint32]int64
		conf_fee   int64
	}
)

func (self *settlement) Enter(now int64) {
	self.changeVal = make(map[uint32]int64)
	self.masterData = packet.NewPacket(nil)
	self.conf_fee = config.GetPublicConfig_Int64("R2B_SERVER_FEE")
	duration := self.status_duration[int(self.s)] // 持续时间 秒
	self.timestamp = now + int64(duration)
	log.Debugf(colorized.Gray("settlement enter duration:%v"), duration)
	log.Debugf(colorized.White("tred:【%v】 tblack:【%v】"), self.red_cards, self.black_cards)
	reddWin, tred, tblack := algorithm.Compare(self.red_cards, self.black_cards)
	var winner_type common.EJinHuaType
	var winner_cards []algorithm.Card_info
	if reddWin {
		winner_type = tred
		winner_cards = self.red_cards
	} else {
		winner_type = tblack
		winner_cards = self.black_cards
	}

	if winner_type == common.ECardType_DUIZI {
		if winner_cards[1][1] < 9 && winner_cards[1][1] > 1 {
			winner_type = common.ECardType_SANPAI
		}
	}
	log.Debugf(colorized.White("reddWin:%v tred:%v【%v】 tblack:%v【%v】 winner_type:%v"), reddWin, tred, self.red_cards, tblack, self.black_cards, winner_type.String())

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_R2B_NEXT_STATE.UInt16())
	send.WriteUInt8(uint8(ERoomStatus_SETTLEMENT))
	send.WriteUInt32(uint32(duration * 1000))

	pack := packet.NewPacket(nil)

	if reddWin {
		pack.WriteUInt8(1)
		self.addStatList(1, int8(winner_type))
	} else {
		pack.WriteUInt8(2)
		self.addStatList(2, int8(winner_type))
	}
	pack.WriteUInt8(winner_type.UInt8())

	pack.WriteUInt16(uint16(self.count_seat()))

	specific_rate := algorithm.Rate_type(winner_type)

	servicepack := packet.NewPacket(nil)
	servicepack2 := packet.NewPacket(nil)
	playerCount := uint16(0)
	playerCount2 := uint16(0)
	tax_scale := config.GetPublicConfig_Int64("TAX")

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

		// 玩家服务费
		acc_service_fee := win * self.conf_fee / 1000

		// 系统服务费
		master_service_fee := lose * self.conf_fee / 1000

		total_win_val += win   // 所有玩家赢得钱
		total_lose_val += lose // 所有玩家输得钱

		win = win * config.GetPublicConfig_Int64("R2B_SYSTEM_FEE") / 100
		if win+come_back > 0 {
			acc.AddMoney(win+come_back, 0, common.EOperateType_SETTLEMENT)
		}

		if acc.Robot == 0 {
			// 计算水位线
			RoomMgr.Water_line -= win + acc_service_fee
			RoomMgr.Water_line += lose - master_service_fee
		}

		// 当前这把总输赢
		change := int64(win - lose)
		self.changeVal[acc.AccountId] = change

		if total_fee := acc_service_fee + master_service_fee; total_fee > 0 {
			playerCount++
			servicepack.WriteUInt32(uint32(acc.AccountId))
			servicepack.WriteUInt32(uint32(total_fee * tax_scale / 100))
		}

		if change != 0 {
			playerCount2++
			servicepack2.WriteUInt32(acc.AccountId)
			servicepack2.WriteInt64(int64(acc.GetMoney()))
			servicepack2.WriteInt64(int64(change))
			servicepack2.WriteString("")

			event.Dispatcher.Dispatch(&event.WinOrLoss{
				RoomID:      self.roomId,
				Acc:         acc,
				Change:      change,
				Seats:       self.seats,
				MasterSeats: self.master_seats,
			}, event.EventType_WinOrLoss)
		}

		if acc.Robot == 0 {
			log.Debugf("reddWin:%v 玩家：%v acc_service_fee:%v master_service_fee:%v change:%v game_count:%v", reddWin, acc.AccountId, acc_service_fee, master_service_fee, change, self.game_count)
		}

		// 组装消息
		index := self.seatIndex(acc.AccountId)
		if index != -1 {
			pack.WriteUInt8(uint8(index + 1))
			pack.WriteInt64(int64(change))
			pack.WriteInt64(int64(acc.GetMoney()))

			temc := 0
			for _, betV := range acc.BetVal {
				if betV > 0 {
					temc++
				}
			}
			pack.WriteUInt16(uint16(temc))
			for k, betV := range acc.BetVal {
				if betV > 0 {
					pack.WriteUInt8(uint8(k))
					pack.WriteUInt32(uint32(betV))
				}
			}
		}
	}

	// 计算庄家得输赢 /////////////////////////////////////////////////////////////////////////////////////////////////////
	total_master_server_fee := ((total_win_val + total_lose_val) * self.conf_fee / 1000)
	total_master_profit := total_lose_val * config.GetPublicConfig_Int64("R2B_SYSTEM_FEE") / 100
	total_master_profit -= total_win_val
	log.Infof("total_win_val:%v total_lose_val:%v self.conf_fee:%v tax_scale:%v total_master_server_fee:%v total_master_profit:%v",
		total_win_val, total_lose_val, self.conf_fee, tax_scale, total_master_server_fee, total_master_profit)

	count := uint16(0)
	tempMaster := packet.NewPacket(nil)
	conf_val := config.GetPublicConfig_Int64("R2B_DOMINATE_MONEY")
	for index, master := range self.master_seats {
		if master == nil {
			continue
		}
		count++
		master_fee := (total_master_server_fee * master.Share) / total_share_val
		if master_fee > 0 {
			playerCount++
			servicepack.WriteUInt32(uint32(master.AccountId))
			servicepack.WriteUInt32(uint32(master_fee * tax_scale / 100))
		}

		master_profit := (total_master_profit * master.Share) / total_share_val

		if master.Robot == 0 {
			// 计算水位线
			RoomMgr.Water_line -= master_profit + master_fee
		}

		if master_profit != 0 {
			log.Infof("庄家:%v 身上钱:%v 份额:%v 盈利:%v ", master.AccountId, master.GetMoney(), master.Share, master_profit)
			self.changeVal[master.AccountId] = master_profit
			master.AddMoney(master_profit, 0, common.EOperateType_SETTLEMENT)
			playerCount2++
			servicepack2.WriteUInt32(master.AccountId)
			servicepack2.WriteInt64(int64(master.GetMoney()))
			servicepack2.WriteInt64(int64(master_profit))
			servicepack2.WriteString("")

			if mon := int64(master.GetMoney()); mon < master.Share*conf_val {
				cur_share := mon / conf_val
				log.Infof(colorized.Yellow("份额下降 玩家:%v 钱:%v < 份额:%v  cur_share:%v "), master.AccountId, mon, master.Share*conf_val, cur_share)

				if cur_share == 0 || (self.dominated_times != -1 && cur_share < config.GetPublicConfig_Int64("R2B_DOMINATE_QUIT")) {
					pack := packet.NewPacket(nil)
					pack.SetMsgID(protomsg.Old_MSGID_R2B_UP_MASTER.UInt16())
					pack.WriteUInt32(master.AccountId)
					pack.WriteUInt8(0)
					pack.WriteUInt64(0)
					core.CoreSend(0, self.owner.Id, pack.GetData(), 0)

					log.Infof("玩家:%v 钱:%v cur_share:%v 份额不够，踢出庄家", master.AccountId, master.GetMoney(), cur_share)
				} else {
					master.Share = cur_share
				}

			}

			event.Dispatcher.Dispatch(&event.WinOrLoss{
				RoomID:      self.roomId,
				Acc:         master.Account,
				Change:      master_profit,
				Seats:       self.seats,
				MasterSeats: self.master_seats,
			}, event.EventType_WinOrLoss)
		}

		if master.Robot == 0 {
			log.Debugf("master reddWin:%v 玩家：%v  master_service_fee:%v master_profit:%v game_count:%v", reddWin, master.AccountId, master_fee, master_profit, self.game_count)
		}
		tempMaster.WriteUInt8(uint8(index) + 1)
		tempMaster.WriteInt64(int64(master_profit))
		tempMaster.WriteInt64(int64(master.GetMoney()))

	}
	self.masterData.WriteUInt16(count)
	self.masterData.CatBody(tempMaster)

	if playerCount > 0 {
		updateAccount := packet.NewPacket(nil)
		updateAccount.SetMsgID(protomsg.Old_MSGID_UPDATE_ACCOUNT.UInt16())
		updateAccount.WriteUInt32(self.roomId)
		updateAccount.WriteUInt8(0)
		updateAccount.WriteUInt16(playerCount)
		updateAccount.CatBody(servicepack2)
		send_tools.Send2Hall(updateAccount.GetData())
	}

	if playerCount2 > 0 {
		ser_fee := packet.NewPacket(nil)
		ser_fee.SetMsgID(protomsg.Old_MSGID_UPDATE_SERVICE_FEE.UInt16())
		ser_fee.WriteUInt8(uint8(self.gameType))
		ser_fee.WriteUInt32(uint32(self.roomId))
		ser_fee.WriteUInt16(playerCount2)
		ser_fee.CatBody(servicepack)
		send_tools.Send2Hall(ser_fee.GetData())
	}

	self.sendData = pack

	for _, acc := range self.accounts {
		newSend := packet.PacketMakeup(send, self.sendData)
		newSend.WriteInt64(self.changeVal[acc.AccountId])
		newSend.WriteInt64(int64(acc.GetMoney()))

		newSend.CatBody(self.masterData)
		send_tools.Send2Account(newSend.GetData(), acc.SessionId)
	}

	if self.dominated_times > 0 {
		self.dominated_times--
	}
	log.Infof("water line:[%v]", RoomMgr.Water_line)
}

func (self *settlement) Tick(now int64) {
	if now >= self.timestamp {
		self.switchStatus(now, ERoomStatus_WAITING_TO_START)
		return
	}
}

func (self *settlement) Leave(now int64) {
	RoomMgr.SaveWaterLine()
	log.Debugf(colorized.Gray("settlement leave\n"))
}

func (self *settlement) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_ENTER_GAME.UInt16(): // 客户端链接进入游戏
		self.Old_MSGID_R2B_ENTER_GAME(actor, msg, session)
	case protomsg.Old_MSGID_R2B_GAME_ENTER_GAME.UInt16(): // 客户端链接进入游戏
		self.Old_MSGID_R2B_GAME_ENTER_GAME(actor, msg, session)
	default:
		log.Warnf("settlement 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}

	return true
}

func (self *settlement) Old_MSGID_R2B_ENTER_GAME(actor int32, msg []byte, session int64) {
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

	newsend := packet.PacketMakeup(send2acc, self.sendData)
	newsend.WriteInt64(int64(self.changeVal[acc.AccountId]))
	newsend.CatBody(self.masterData)
	send_tools.Send2Account(newsend.GetData(), acc.SessionId)
}

func (self *settlement) Old_MSGID_R2B_GAME_ENTER_GAME(actor int32, msg []byte, session int64) {
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

	newsend := packet.PacketMakeup(send2acc, self.sendData)
	newsend.WriteInt64(int64(self.changeVal[acc.AccountId]))
	newsend.CatBody(self.masterData)
	send_tools.Send2Account(newsend.GetData(), acc.SessionId)
}
