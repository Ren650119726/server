package logic

import (
	"root/core/packet"
	"root/protomsg"
	"root/server/hall/account"
	"root/server/hall/send_tools"
)

// Client请求邮件列表
func (self *Hall) MSG_CS_EMAILS_REQ(actor int32, msg []byte, session int64) {
	acc := account.AccountMgr.GetAccountBySessionIDAssert(session)
	emails := account.EmailMgr.GetAllEmailofPerson(acc.AccountId)
	send_tools.Send2Account(protomsg.MSG_SC_EMAILS_RES.UInt16(), &protomsg.EMAILS_RES{Emails:emails},session)
}

// Client请求改变邮件状态为已读
func (self *Hall) MSG_CS_EMAIL_READ_REQ(actor int32, msg []byte, session int64) {
	pbData := packet.PBUnmarshal(msg, &protomsg.EMAIL_READ_REQ{}).(*protomsg.EMAIL_READ_REQ)
	acc := account.AccountMgr.GetAccountBySessionIDAssert(session)
	account.EmailMgr.ReadMail(acc.AccountId,pbData.GetEmailID())
	send_tools.Send2Account(protomsg.MSG_SC_EMAIL_READ_RES.UInt16(), &protomsg.EMAIL_READ_RES{EmailID:pbData.GetEmailID()},session)
}

// Client请求领取邮件内奖励
func (self *Hall) MSG_CS_EMAIL_REWARD_REQ(actor int32, msg []byte, session int64) {
	pbData := packet.PBUnmarshal(msg, &protomsg.EMAIL_REWARD_REQ{}).(*protomsg.EMAIL_REWARD_REQ)
	acc := account.AccountMgr.GetAccountBySessionIDAssert(session)
	ret := account.EmailMgr.GetMailReward(acc.AccountId,pbData.GetEmailID())
	response := &protomsg.EMAIL_REWARD_RES{Ret:int32(ret),EmailID:pbData.GetEmailID()}
	send_tools.Send2Account(protomsg.MSG_SC_EMAIL_REWARD_RES.UInt16(),response,session)
}

// Client请求删除邮件
func (self *Hall) MSG_CS_EMAIL_DEL_REQ(actor int32, msg []byte, session int64) {
	pbData := packet.PBUnmarshal(msg, &protomsg.EMAIL_DEL_REQ{}).(*protomsg.EMAIL_DEL_REQ)
	acc := account.AccountMgr.GetAccountBySessionIDAssert(session)
	account.EmailMgr.RemoveMail(acc.AccountId,pbData.GetEmailID())
	send_tools.Send2Account(protomsg.MSG_SC_EMAIL_DEL_RES.UInt16(),&protomsg.EMAIL_DEL_RES{EmailID:pbData.GetEmailID()},session)
}
