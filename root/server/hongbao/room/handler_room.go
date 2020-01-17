package room

import (
	"root/common"
	"root/core/log"
	"root/core/packet"
	"root/protomsg"
	"root/server/hongbao/account"
	"root/server/hongbao/send_tools"
)

func (self *Room) Old_MSGID_LEAVE_GAME(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	t := pack.ReadUInt32() // ????

	account.CheckSession(accountId, session)

	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		send := packet.NewPacket(nil)
		send.SetMsgID(protomsg.Old_MSGID_LEAVE_GAME.UInt16())
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	if t > 3 {
		send := packet.NewPacket(nil)
		send.SetMsgID(protomsg.Old_MSGID_LEAVE_GAME.UInt16())
		send.WriteUInt8(2)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	for _, v := range self.rob_list {
		if v.acc.AccountId == accountId {
			return
		}
	}
	for _, v := range self.hongbao_list {
		if v.acc.AccountId == accountId {
			send := packet.NewPacket(nil)
			send.SetMsgID(protomsg.Old_MSGID_LEAVE_GAME.UInt16())
			send.WriteUInt8(3)
			send_tools.Send2Account(send.GetData(), session)
			return
		}
	}

	if self.cur_hongbao != nil && self.cur_hongbao.acc.AccountId == accountId {
		send := packet.NewPacket(nil)
		send.SetMsgID(protomsg.Old_MSGID_LEAVE_GAME.UInt16())
		send.WriteUInt8(3)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_LEAVE_GAME.UInt16())
	send.WriteUInt8(0)
	send_tools.Send2Account(send.GetData(), session)

	self.leaveRoom(accountId)

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

// 发红包
func (self *Room) Old_MSGID_HONGBAO_POST_HONGBAO(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountID := pack.ReadUInt32()
	ratio := pack.ReadUInt16()
	bomb := pack.ReadInt8()

	account.CheckSession(accountID, session)

	if self.Close {
		return
	}
	if ratio > uint16(self.GetParamInt(3)) {
		log.Errorf("倍数操作配置值 ratio:%v   self.GetParamInt(3):%v", ratio, self.GetParamInt(3))
		return
	}

	bt := self.GetParamInt(0) // 底注金额
	need := int64(ratio) * int64(bt)

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_HONGBAO_POST_HONGBAO.UInt16())
	acc := account.AccountMgr.GetAccountByID(accountID)
	if acc == nil {
		send.WriteUInt8(3)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	if acc.GetMoney() < uint64(need) {
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), acc.SessionId)
		return
	}

	if len(self.hongbao_list) != 0 {
		for _, v := range self.rob_list {
			if v.acc.AccountId == acc.AccountId {
				send.WriteUInt8(2)
				send_tools.Send2Account(send.GetData(), acc.SessionId)
				return
			}
		}
	}

	acc.AddMoney(-int64(need), 3, common.EOperateType_HONGBAO)

	hongbao := &HongBao{
		acc:      acc,
		money:    need,
		bomb_num: bomb,
	}

	self.hongbao_list = append(self.hongbao_list, hongbao)

	send.WriteUInt8(0)
	send.WriteInt64(int64(acc.GetMoney()))
	send_tools.Send2Account(send.GetData(), acc.SessionId)

	broadcast_send := packet.NewPacket(nil)
	broadcast_send.SetMsgID(protomsg.Old_MSGID_HONGBAO_INCREASE_POST.UInt16())
	broadcast_send.WriteUInt32(acc.AccountId)
	broadcast_send.WriteString(acc.Name)
	broadcast_send.WriteString(acc.HeadURL)
	broadcast_send.WriteInt64(int64(acc.GetMoney()))
	broadcast_send.WriteInt64(need)
	self.SendBroadcast(broadcast_send.GetData())

	// 发红包的扣服务费
	log.Debugf("房间:%v 玩家:%v 请求发红包，money:%v bomb_num:%v, 底注金额:%v", self.roomId, acc.AccountId, need, bomb, bt)
}

// 请求玩家列表
func (self *Room) Old_MSGID_HONGBAO_PLAYER_LIST(actor int32, msg []byte, session int64) {
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_HONGBAO_PLAYER_LIST.UInt16())
	send.WriteUInt16(uint16(len(self.accounts)))
	for accid, acc := range self.accounts {
		send.WriteUInt32(accid)
		send.WriteString(acc.Name)
		send.WriteString(acc.HeadURL)
		send.WriteInt64(int64(acc.GetMoney()))
		send.WriteString(acc.Signature)
	}
	send_tools.Send2Account(send.GetData(), session)
}
