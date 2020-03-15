package account

import (
	"encoding/base64"
	"fmt"
	"github.com/golang/protobuf/proto"
	"root/common"
	"root/common/config"
	"root/common/model/rediskey"
	"root/core/db"
	"root/core/log"
	"root/core/utils"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/hall/event"
	"root/server/hall/send_tools"
	"root/server/hall/types"
)

var EmailMgr = newEmailMgr()

type (
	EmailMap  map[uint32]*protomsg.Email
	PersonMap map[uint32]EmailMap

	emailMgr struct {
		Emails     PersonMap
		EmailMAXID uint32
	}
)

func newEmailMgr() *emailMgr {
	ret := &emailMgr{
		EmailMAXID: 0,
		Emails:     make(PersonMap),
	}
	return ret
}

func (self *emailMgr) increase_EmailID() uint32 {
	self.EmailMAXID++
	if self.EmailMAXID >= 10 {
		self.EmailMAXID = 1
	}
	t := uint32(utils.SecondTimeSince1970())
	t = t % 10000000
	t = (t * 10) + self.EmailMAXID
	return t
}

func (self *emailMgr) SaveAll(mPlayer map[uint32]bool) {
	for accid := range mPlayer {
		self.SaveOnePlayerAllEmail(accid) // 关服回存
	}
}

func (self *emailMgr) SaveOnePlayerAllEmail(nAccountID uint32) {
	tPersonal := self.Emails[nAccountID]
	if tPersonal == nil {
		return
	}
	all_email := &inner.SAVE_EMAIL_PERSON{AccountId: nAccountID}
	for _, email := range tPersonal {
		all_email.Emails = append(all_email.Emails, email)
	}
	// 发给redis 缓存
	data, e := proto.Marshal(all_email)
	if e != nil {
		log.Errorf("序列化失败")
	} else {
		db.HSet(rediskey.PlayerId(nAccountID), "email", string(data))
	}
	send_tools.Send2DB(inner.SERVERMSG_HD_SAVE_EMAIL_PERSON.UInt16(), all_email) // 大厅
}

// 加载所有玩家邮件数据
func (self *emailMgr) LoadAll(all_emails *inner.ALL_EMAIL_RESP) {
	for _, v := range all_emails.AcccountMail {
		accid := v.AccountId
		emails := v.Emails
		if _, exist := self.Emails[accid]; !exist {
			self.Emails[accid] = make(EmailMap)
		}

		person := self.Emails[accid]
		for _, email := range emails {
			person[email.EmailID] = email
		}
	}
}

func (self *emailMgr) AddMail(recvId, emailType uint32, content string, rmb uint64) int32 {
	acc := AccountMgr.GetAccountByIDAssert(recvId)
	if acc.Robot > 0 {
		// 机器人不能接收邮件
		return -2
	}

	if content != "" {
		content = base64.StdEncoding.EncodeToString([]byte(content))
	}
	emailID := self.increase_EmailID()
	email := &protomsg.Email{
		EmailID:   emailID,
		EmailType: emailType,
		Content:   content,
		Money:     rmb,
		SendTime:  utils.SecondTimeSince1970(),
		IsRead:    0,
	}

	log.Infof("Add EmailID:%v, RecvID:%v, %v, Content:%v, nRMB:%v", emailID, recvId, types.EmailType(email.EmailType), content, rmb)
	if accMap := self.Emails[recvId]; accMap == nil {
		self.Emails[recvId] = make(EmailMap)
	}

	accMap := self.Emails[recvId]
	accMap[emailID] = email
	self.SaveOnePlayerAllEmail(recvId) // 添加一封邮件

	if acc.IsOnline() {
		self.SendEmailToClient(recvId, acc.SessionId, emailID)
	}
	return int32(emailID)
}

func (self *emailMgr) RemoveMail(accountId, emailId uint32) uint8 {
	acc := AccountMgr.GetAccountByIDAssert(accountId)
	person := self.Emails[accountId]
	if person == nil {
		log.Fatalf("RemoveMail email_node is nil :%v", accountId)
		return 2
	}
	email := person[emailId]
	if email == nil {
		return 0
	}

	if email.Money > 0 {
		eOperate := email.GetEmailType()
		acc.AddMoney(int64(email.Money), common.EOperateType(eOperate), 0)
		email.Money = 0
		event.Dispatcher.Dispatch(event.UpdateCharge{AccountID: accountId, RMB: int64(email.Money)}, event.EventType_UpdateCharge)
		log.Infof("Remove EmailID:%v, RecvID:%v, EmailType:%v, Content:%v, nRMB:%v", emailId, accountId, email.EmailType, email.Content, email.Money)
	}

	delete(person, emailId)
	self.SaveOnePlayerAllEmail(accountId)
	return 0
}

// 领取邮件
func (self *emailMgr) GetMailReward(accountID, emailId uint32) uint8 {
	acc := AccountMgr.GetAccountByIDAssert(accountID)
	personal := self.Emails[accountID]
	if personal == nil {
		log.Panicf("GetMailReward personal == nil accId:%v", accountID)
	}

	email := personal[emailId]
	if email == nil {
		log.Panicf("GetMailReward email == nil :%v emailid:%v", accountID, emailId)
	}
	email.IsRead = 1
	if email.Money <= 0 {
		log.Warnf("GetMailReward email.Money <= 0 email.Money:%v", email.Money)
		return 1
	}
	acc.AddMoney(int64(email.Money), common.EOperateType(email.GetEmailType()), 0)
	email.Money = 0
	event.Dispatcher.Dispatch(event.UpdateCharge{AccountID: accountID, RMB: int64(email.Money)}, event.EventType_UpdateCharge)
	log.Infof("GetMailReward EmailID:%v, RecvID:%v, EmailType:%v, Content:%v, nRMB:%v", emailId, accountID, email.EmailType, email.Content, email.Money)

	self.SaveOnePlayerAllEmail(accountID)
	return 0
}

// 阅读邮件
func (self *emailMgr) ReadMail(accountID, emailId uint32) {
	personal := self.Emails[accountID]
	if personal == nil {
		return
	}

	email := personal[emailId]
	if email == nil {
		return
	}
	email.IsRead = 1
}

// 获取玩家未读邮件数量
func (self *emailMgr) GetPlayerUnReadEmailNum(accountID uint32) int {
	personal := self.Emails[accountID]
	if personal == nil {
		return 0
	}

	var count = 0
	for _, email := range personal {
		if email.IsRead == 0 {
			count = count + 1
		}
	}

	return count
}

func (self *emailMgr) GetAllEmailofPerson(accountID uint32) []*protomsg.Email {
	emails := make([]*protomsg.Email, 0)
	person := self.Emails[accountID]
	if person == nil {
		return emails
	}

	for _, e := range person {
		emails = append(emails, e)
	}
	return emails
}

// 通知在线玩家有新邮件
func (self *emailMgr) SendEmailToClient(accountID uint32, session int64, emailId uint32) {
	personal := self.Emails[accountID]
	if personal == nil {
		return
	}
	email := personal[emailId]
	if email == nil {
		return
	}
	send_tools.Send2Account(protomsg.MSG_SC_EMAIL_NEW.UInt16(), &protomsg.EMAIL_NEW{New: email}, session)
}

func (self *emailMgr) PrintEmail(nAccountID uint32) {
	personal := self.Emails[nAccountID]
	if personal == nil {
		return
	}

	for _, tEmail := range personal {
		content := tEmail.Content
		if content != "" {
			bytes, _ := base64.StdEncoding.DecodeString(content)
			content = string(bytes)
		}
		fmt.Printf("邮件ID:%v, 邮件类型:%v, 元宝:%v, 发送时间:%v, 是否已读:%v, 邮件内容:%v\r\n", tEmail.EmailID, tEmail.EmailType, tEmail.Money/config.RMB_BILI, utils.GetTimeFormatString(int64(tEmail.SendTime)), tEmail.IsRead, content)
	}
}
