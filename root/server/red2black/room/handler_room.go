package room

import (
	"root/common/config"
	"root/core/log"
	"root/core/packet"
	"root/protomsg"
	"root/server/red2black/account"
	"root/server/red2black/event"
	"root/server/red2black/send_tools"
)

func (self *Room) Old_MSGID_R2B_LEAVE_GAME(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	t := pack.ReadUInt32() // ????

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_LEAVE_GAME.UInt16())

	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	if t > 3 {
		send.WriteUInt8(2)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	//// 判断是否押注了，押注的不能离开
	if acc.GetTotalBetVal() > 0 {
		send.WriteUInt8(3)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	// 庄家不能离开游戏
	if self.SeatMasterIndex(accountId) != -1 {
		send.WriteUInt8(4)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	send.WriteUInt8(0)
	send_tools.Send2Account(send.GetData(), session)
}
func (self *Room) Old_MSGID_R2B_GAME_LEAVE_GAME(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	t := pack.ReadUInt32() // ????

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_R2B_GAME_LEAVE_GAME.UInt16())

	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	if t > 3 {
		send.WriteUInt8(2)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	if acc.GetTotalBetVal() > 0 {
		send.WriteUInt8(3)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	// 庄家不能离开游戏
	if self.SeatMasterIndex(accountId) != -1 {
		send.WriteUInt8(4)
		send_tools.Send2Account(send.GetData(), session)
		return
	}
	send.WriteUInt8(0)
	send_tools.Send2Account(send.GetData(), session)
}

func (self *Room) Old_MSGID_R2B_UP_SEAT(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	seatIndex := pack.ReadUInt8() - 1 // 座位号

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_R2B_UP_SEAT.UInt16())

	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return
	}
	// 庄家不能坐下
	if self.SeatMasterIndex(accountId) != -1 {
		send.WriteUInt8(17)
		send_tools.Send2Account(send.GetData(), session)
		return
	}
	ret := self.UpSeat(accountId, seatIndex, send)
	log.Debugf("玩家:%v 请求坐下:%v  ret :%v", accountId, seatIndex, ret)
	if ret != 0 {
		send.WriteUInt8(uint8(ret))
		send_tools.Send2Account(send.GetData(), session)
		return
	}
}

func (self *Room) Old_MSGID_R2B_PLAYER_LIST(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_R2B_PLAYER_LIST.UInt16())

	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	if self.accounts[accountId] == nil {
		send.WriteUInt8(11)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	send.WriteUInt8(0)
	send.WriteUInt16(uint16(len(self.accounts)))
	for id, v := range self.accounts {
		send.WriteUInt32(id)
		send.WriteString(v.Name)
		send.WriteString(v.HeadURL)
		send.WriteInt64(int64(v.GetMoney()))
		send.WriteString(v.Signature)
	}
	send_tools.Send2Account(send.GetData(), session)
}

func (self *Room) Old_MSGID_R2B_STATISTICS_LIST(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_R2B_STATISTICS_LIST.UInt16())

	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	if self.accounts[accountId] == nil {
		send.WriteUInt8(11)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	send.WriteUInt8(0)
	send.WriteUInt16(uint16(len(self.statis)))
	for _, result := range self.statis {
		send.WriteUInt8(uint8(result.ret))
		send.WriteUInt8(uint8(result.t))
	}

	send_tools.Send2Account(send.GetData(), session)
}

func (self *Room) Old_MSGID_SEND_EMOJI(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	sendID := pack.ReadUInt32()
	targetID := pack.ReadUInt32()
	emojiType := pack.ReadUInt8()
	emojiID := pack.ReadUInt8()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_SEND_EMOJI.UInt16())
	send.WriteUInt8(0)
	send.WriteUInt32(sendID)
	send.WriteUInt32(targetID)
	send.WriteUInt8(emojiType)
	send.WriteUInt8(emojiID)
	self.SendBroadcast(send.GetData())

	acc := self.accounts[targetID]
	if self.seatIndex(sendID) != -1 && acc != nil && acc.Robot > 0 && emojiType == 2 {
		event.Dispatcher.Dispatch(&event.Emotion{
			RoomID:   self.roomId,
			SendID:   sendID,
			TargetID: targetID,
		}, event.EventType_Emotion)
	}

}

func (self *Room) Old_MSGID_SEND_TEXT_SHORTCUTS(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountID := pack.ReadUInt32()
	textID := pack.ReadUInt8()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_SEND_TEXT_SHORTCUTS.UInt16())
	acc := account.AccountMgr.GetAccountByID(accountID)
	if acc == nil {
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	send.WriteUInt8(0)
	send.WriteUInt32(accountID)
	send.WriteUInt8(textID)
	self.SendBroadcast(send.GetData())
}

func (self *Room) Old_MSGID_R2B_UP_MASTER(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountID := pack.ReadUInt32()
	op := pack.ReadUInt8() // 1 上庄 0 下庄
	share := pack.ReadUInt64()

	acc := self.accounts[accountID]
	if acc == nil {
		log.Warnf("上庄失败，玩家不再房间:%v 玩家ID：%v op:%v", self.roomId, accountID, op)
		return
	}

	master_index := self.SeatMasterIndex(accountID)
	check := self.check_apply_list(accountID)

	// 如果已经在庄家位，或者 在申请列表 都不能上庄
	if op == 1 && (master_index != -1 || check) {
		err := packet.NewPacket(nil)
		err.SetMsgID(protomsg.Old_MSGID_R2B_UP_MASTER.UInt16())
		err.WriteUInt8(1)
		err.WriteUInt8(op)
		send_tools.Send2Account(err.GetData(), acc.SessionId)
		return
	}

	need_val := share * uint64(config.GetPublicConfig_Int64("R2B_DOMINATE_MONEY"))
	if op == 1 && acc.GetMoney() < need_val {
		err := packet.NewPacket(nil)
		err.SetMsgID(protomsg.Old_MSGID_R2B_UP_MASTER.UInt16())
		err.WriteUInt8(2)
		err.WriteUInt8(op)
		send_tools.Send2Account(err.GetData(), acc.SessionId)
		return
	}

	if op == 0 && master_index == -1 && !check {
		err := packet.NewPacket(nil)
		err.SetMsgID(protomsg.Old_MSGID_R2B_UP_MASTER.UInt16())
		err.WriteUInt8(3)
		err.WriteUInt8(op)
		send_tools.Send2Account(err.GetData(), acc.SessionId)
		return
	}

	if op == 1 {
		self.apply_list = append(self.apply_list, &account.Master{Account: acc, Share: int64(share)})

	} else if op == 0 {
		if check {
			for i, v := range self.apply_list {
				if v.AccountId == accountID {
					self.apply_list = append(self.apply_list[:i], self.apply_list[i+1:]...)
					break
				}
			}
		} else if master_index != -1 {
			// 如果当前状态不是等待，就先不忙删
			if self.status.State() == int32(ERoomStatus_WAITING_TO_START) {
				for i, v := range self.master_seats {
					if v == nil {
						continue
					}
					if v.AccountId == accountID {
						self.master_seats[i] = nil
						self.dominated_times = -1
						break
					}
				}
				self.update_master_list()

			} else {
				if self.downMasterMSG[accountID] != nil {
					err := packet.NewPacket(nil)
					err.SetMsgID(protomsg.Old_MSGID_R2B_UP_MASTER.UInt16())
					err.WriteUInt8(4)
					err.WriteUInt8(op)
					send_tools.Send2Account(err.GetData(), acc.SessionId)
					return
				} else {
					self.downMasterMSG[accountID] = pack
				}
			}
		}
	} else {
		log.Warnf("error :%v", op)
	}

	if session != 0 {
		err := packet.NewPacket(nil)
		err.SetMsgID(protomsg.Old_MSGID_R2B_UP_MASTER.UInt16())
		err.WriteUInt8(0)
		err.WriteUInt8(op)
		send_tools.Send2Account(err.GetData(), acc.SessionId)
	}

	self.update_applist_sort()  // 新的上\下庄
	self.update_applist_count() // 新的上\下庄
}

func (self *Room) Old_MSGID_R2B_UPMASTER_LIST(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(nil)
	pack.SetMsgID(protomsg.Old_MSGID_R2B_UPMASTER_LIST.UInt16())

	temp := packet.NewPacket(nil)
	count := uint16(0)
	for _, v := range self.apply_list {
		count++
		temp.WriteUInt32(v.AccountId)
		temp.WriteString(v.Name)
		temp.WriteString(v.HeadURL)
		temp.WriteInt64(int64(v.GetMoney()))
		temp.WriteString(v.Signature)
		temp.WriteUInt64(uint64(v.Share))
	}
	pack.WriteUInt16(count)
	pack.CatBody(temp)
	send_tools.Send2Account(pack.GetData(), session)
}
