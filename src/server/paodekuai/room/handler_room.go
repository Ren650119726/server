package room

import (
	"root/common"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"encoding/json"
	"root/protomsg"
	"root/server/paodekuai/account"
	"root/server/paodekuai/send_tools"
)

func (self *Room) Old_MSGID_PDK_WATCH_LIST(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()

	tSend := packet.NewPacket(nil)
	tSend.SetMsgID(protomsg.Old_MSGID_PDK_WATCH_LIST.UInt16())

	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		tSend.WriteUInt8(1)
		send_tools.Send2Account(tSend.GetData(), session)
		return
	}

	if self.accounts[accountId] == nil {
		tSend.WriteUInt8(11)
		send_tools.Send2Account(tSend.GetData(), session)
		return
	}

	nCount := uint16(0)
	tSend.WriteUInt8(0)
	nWPos := tSend.GetWritePos()
	tSend.WriteUInt16(nCount)
	for id, v := range self.accounts {
		index := self.get_seat_index(v.AccountId)
		if index > self.max_count {
			nCount++
			tSend.WriteUInt32(id)
			tSend.WriteString(v.Name)
			tSend.WriteString(v.HeadURL)
			tSend.WriteInt64(int64(v.GetMoney()))
			tSend.WriteString(v.Signature)
		}
	}
	tSend.Rrevise(nWPos, nCount)
	send_tools.Send2Account(tSend.GetData(), session)
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
}

func (self *Room) Old_MSGID_SEND_TEXT_SHORTCUTS(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountID := pack.ReadUInt32()
	textID := pack.ReadUInt8()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_SEND_TEXT_SHORTCUTS.UInt16())
	send.WriteUInt8(0)
	send.WriteUInt32(accountID)
	send.WriteUInt8(textID)
	self.SendBroadcast(send.GetData())
}

// 请求坐下
func (self *Room) Old_MSGID_PDK_SIT_DOWN(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountID := pack.ReadUInt32()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_PDK_SIT_DOWN.UInt16())
	acc := account.AccountMgr.GetAccountByID(accountID)
	if acc == nil {
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	acc = self.accounts[accountID]
	if acc == nil {
		send.WriteUInt8(2)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	money := acc.GetMoney()
	if money < self.sitdown_limit {
		send.WriteUInt8(4)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	if check_index := self.get_seat_index(accountID); check_index < self.max_count {
		send.WriteUInt8(4)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	index := self.sit_down(accountID)
	if index > self.max_count {
		send.WriteUInt8(3)
		send_tools.Send2Account(send.GetData(), session)
		return
	} else {
		send.WriteUInt8(0)
		send.WriteUInt8(uint8(index + 1))
		send_tools.Send2Account(send.GetData(), session)
	}

	addplayer := packet.NewPacket(nil)
	addplayer.SetMsgID(protomsg.Old_MSGID_PDK_ADD_PLAYER.UInt16())
	addplayer.WriteUInt8(uint8(index + 1))
	addplayer.WriteUInt32(accountID)
	addplayer.WriteString(acc.Name)
	addplayer.WriteString(acc.HeadURL)
	addplayer.WriteInt64(int64(acc.GetMoney()))
	addplayer.WriteString(acc.Signature)
	addplayer.WriteUInt8(acc.IsOnline())
	self.send_broadcast_excludeid(addplayer.GetData(), acc.AccountId)
	self.broadcast_watch_count()
	self.send_game_data(acc)

	// 发送离开消息到大厅
	tLeave := packet.NewPacket(nil)
	tLeave.SetMsgID(protomsg.Old_MSGID_UPDATE_LEAVE.UInt16())
	tLeave.WriteUInt32(acc.AccountId)
	tLeave.WriteUInt32(self.roomId)
	tLeave.WriteUInt16(uint16(self.get_sit_down_count()))
	tLeave.WriteUInt8(1) // 1观战
	send_tools.Send2Hall(tLeave.GetData())

	// 2 hall
	tEnter := packet.NewPacket(nil)
	tEnter.SetMsgID(protomsg.Old_MSGID_UPDATE_ENTER.UInt16())
	tEnter.WriteUInt32(acc.AccountId)
	tEnter.WriteUInt32(self.roomId)
	tEnter.WriteUInt16(uint16(self.get_sit_down_count()))
	tEnter.WriteUInt8(0) // 0坐下
	tEnter.WriteUInt8(index + 1)
	send_tools.Send2Hall(tEnter.GetData())
	log.Infof(colorized.Yellow("玩家:[%v], 座位号:[%v], 身上余额:[%v] 坐下"), acc.AccountId, index, acc.GetMoney())
}

func (self *Room) Old_MSGID_ENTER_GAME(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	nAccountID := pack.ReadUInt32()
	_ = pack.ReadUInt32()
	nEnterType := pack.ReadUInt8()

	send2c := packet.NewPacket(nil)
	send2c.SetMsgID(protomsg.Old_MSGID_ENTER_GAME.UInt16())

	// 通知客户端，进入游戏成功
	acc := account.AccountMgr.GetAccountByID(nAccountID)
	if acc == nil {
		send2c.WriteUInt8(1)
		send_tools.Send2Account(send2c.GetData(), session)
		return
	}

	ret, index := self.can_enter_room(nAccountID)
	if ret > 0 {
		send2c.WriteUInt8(uint8(ret))
		send_tools.Send2Account(send2c.GetData(), session)
		return
	}

	if _, exist := self.accounts[nAccountID]; !exist {
		send2hall := packet.NewPacket(nil)
		send2hall.SetMsgID(protomsg.Old_MSGID_UPDATE_ENTER.UInt16())
		send2hall.WriteUInt32(nAccountID)
		send2hall.WriteUInt32(self.roomId)
		send2hall.WriteUInt16(uint16(self.get_sit_down_count()))
		send2hall.WriteUInt8(uint8(1)) // 1观战标记
		send2hall.WriteUInt8(uint8(0))
		send_tools.Send2Hall(send2hall.GetData())
	}

	send2c.WriteUInt8(0)
	send2c.WriteUInt32(self.roomId)
	send_tools.Send2Account(send2c.GetData(), acc.SessionId)

	self.enter_room(acc)
	self.set_need_passwd(nAccountID, nEnterType)
	self.send_game_data(acc)

	// 已在座位上, 通知其他人
	if index < self.max_count {
		addplayer := packet.NewPacket(nil)
		addplayer.SetMsgID(protomsg.Old_MSGID_PDK_ADD_PLAYER.UInt16())
		addplayer.WriteUInt8(uint8(index + 1))
		addplayer.WriteUInt32(nAccountID)
		addplayer.WriteString(acc.Name)
		addplayer.WriteString(acc.HeadURL)
		addplayer.WriteInt64(int64(acc.GetMoney()))
		addplayer.WriteString(acc.Signature)
		addplayer.WriteUInt8(acc.IsOnline())
		self.send_broadcast_excludeid(addplayer.GetData(), acc.AccountId)
	}
}

func (self *Room) Old_MSGID_LEAVE_GAME(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	nType := pack.ReadUInt32() // ????

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_LEAVE_GAME.UInt16())
	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	if nType > 3 {
		send.WriteUInt8(2)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	nRet := self.leave_room(acc, true)
	if nRet > 0 {
		send.WriteUInt8(nRet)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	send.WriteUInt8(0)
	send_tools.Send2Account(send.GetData(), session)
}

// 连接断开处理
func (self *Room) Disconnect(session int64) {
	acc := account.AccountMgr.GetAccountBySessionID(session)
	if acc == nil {
		log.Warnf("找不到玩家:%v", session)
		return
	}

	acc = self.accounts[acc.AccountId]
	if acc == nil {
		return
	}
	acc.State = common.STATUS_OFFLINE.UInt32()
	index := self.get_seat_index(acc.AccountId)
	if index < self.max_count {
		offline := packet.NewPacket(nil)
		offline.SetMsgID(protomsg.Old_MSGID_PDK_OFFLINE.UInt16())
		offline.WriteUInt8(uint8(index + 1))
		self.SendBroadcast(offline.GetData())
	}
}

// 请求盈利
func (self *Room) Old_MSGID_PDK_PROFIT(actor int32, msg []byte, session int64) {
	tPack := packet.NewPacket(msg)
	nAccountID := tPack.ReadUInt32()

	tSend := packet.NewPacket(nil)
	tSend.SetMsgID(protomsg.Old_MSGID_PDK_PROFIT.UInt16())
	acc := account.AccountMgr.GetAccountByID(nAccountID)
	if acc == nil {
		tSend.WriteUInt32(uint32(nAccountID))
		tSend.WriteInt64(0)
		tSend.WriteUInt32(uint32(0))
		tSend.WriteUInt32(uint32(0))
		send_tools.Send2Account(tSend.GetData(), session)
		return
	}

	penalty := self.calc_penalty_value(acc)
	tSend.WriteUInt32(uint32(acc.AccountId))
	tSend.WriteInt64(acc.Profit)
	tSend.WriteUInt32(uint32(acc.Games))
	tSend.WriteUInt32(penalty)
	send_tools.Send2Account(tSend.GetData(), session)
	//log.Debugf("请求盈利 退出 accID:%v index:%v profit:%v games:%v penalty:%v", player.acc.AccountId, index, player.profit, player.acc.Games, penalty)
}

// 历史中奖记录
func (self *Room) Old_MSGID_PDK_AWARD_HISTORY(actor int32, msg []byte, session int64) {
	log.Debugf("客户端请求历史中奖记录")

	tNode := RoomMgr.Bonus_h[uint32(self.bet)]
	if tNode == nil {
		return
	}

	var str string
	if tNode.History_max_info == nil {
		str = string("{}")
	} else {
		data, _ := json.Marshal(tNode.History_max_info)
		str = string(data)
	}

	tSend := packet.NewPacket(nil)
	tSend.SetMsgID(protomsg.Old_MSGID_PDK_AWARD_HISTORY.UInt16())
	tSend.WriteUInt8(0)
	tSend.WriteString(str)

	nCount := uint16(0)
	nWPos := tSend.GetWritePos()
	tSend.WriteUInt16(nCount)
	for i := len(tNode.Award_history) - 1; i >= 0; i-- {
		nCount++
		sData, _ := json.Marshal(tNode.Award_history[i])
		str := string(sData)
		tSend.WriteString(str)
	}
	tSend.Rrevise(nWPos, nCount)
	send_tools.Send2Account(tSend.GetData(), session)
}

// 请求所有战绩
func (self *Room) Old_MSGID_PDK_ALL_RECORD_INFO(actor int32, msg []byte, session int64) {

	nCount := uint16(0)
	tSend := packet.NewPacket(nil)
	nWPos := tSend.GetWritePos()
	tSend.WriteUInt16(nCount)
	for index, player := range self.seats {
		if player != nil {
			nCount++
			tSend.WriteUInt32(player.acc.AccountId)
			tSend.WriteUInt8(uint8(index + 1))
			tSend.WriteString(player.acc.Name)
			tSend.WriteString(player.acc.HeadURL)
			tSend.WriteInt64(int64(player.acc.GetMoney()))
			tSend.WriteUInt32(uint32(player.acc.Games))
			tSend.WriteInt64(player.acc.Profit)
		}
	}
	tSend.Rrevise(nWPos, nCount)
	tSend.SetMsgID(protomsg.Old_MSGID_PDK_ALL_RECORD_INFO.UInt16())
	send_tools.Send2Account(tSend.GetData(), session)
}
