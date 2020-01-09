package logic

import (
	"root/core/packet"
	"root/protomsg"
	"root/server/hall/account"
	"root/server/hall/send_tools"
)

// Client请求改变邮件状态为已读
func (self *Hall) Old_MSGID_CHANGE_EMAIL_TO_READED(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	emailId := pack.ReadUInt32()
	account.CheckSession(accountId, session)
	account.EmailMgr.ReadMail(accountId, emailId)
}

// Client请求领取邮件内奖励
func (self *Hall) Old_MSGID_GET_EMAIL_REWARD(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	emailId := pack.ReadUInt32()
	account.CheckSession(accountId, session)
	nRet := account.EmailMgr.GetMailReward(accountId, emailId)

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_GET_EMAIL_REWARD.UInt16())
	send.WriteUInt8(nRet)
	send.WriteUInt32(emailId)
	send_tools.Send2Account(send.GetData(), session)
}

// Client请求邮件列表
func (self *Hall) Old_MSGID_SEND_EMALL_LIST(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	account.CheckSession(accountId, session)
	account.EmailMgr.SendEmailToClient(accountId, session, 0)
}

// Client请求删除邮件
func (self *Hall) Old_MSGID_DELETE_EMAIL(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	emailId := pack.ReadUInt32()
	account.CheckSession(accountId, session)
	nRet := account.EmailMgr.RemoveMail(accountId, emailId)

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_DELETE_EMAIL.UInt16())
	send.WriteUInt8(nRet)
	send.WriteUInt32(emailId)
	send_tools.Send2Account(send.GetData(), session)
}
