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
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/hall/event"
	"root/server/hall/send_tools"
	"root/server/hall/types"
	"strings"
)

var EmailMgr = newEmailMgr()

type (
	EmailMap  map[uint32]*protomsg.Email
	PersonMap map[uint32]EmailMap

	emailMgr struct {
		Emails       PersonMap
		EmailMAXID   uint32
		rechargeLog  []string
		isBatSaveLog bool
	}
)

func newEmailMgr() *emailMgr {
	ret := &emailMgr{
		EmailMAXID:   0,
		Emails:       make(PersonMap),
		rechargeLog:  make([]string, 0, 100),
		isBatSaveLog: false,
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

func (self *emailMgr) SetBatSaveLog(isOpen bool) {
	self.isBatSaveLog = isOpen
}

func (self *emailMgr) AddRechargeLog(strOrder string, nAccountID uint32, nRMB uint64, strOperator string, nState uint8, strTime string, nEmailType uint32) {

	if self.isBatSaveLog == true {
		strLog := fmt.Sprintf("('%v',%v, %v, '%v', %v, '%v', %v),", strOrder, nAccountID, nRMB, strOperator, nState, strTime, nEmailType)
		self.rechargeLog = append(self.rechargeLog, strLog)

		RECHARGE_LOG_COUNT := config.GetPublicConfig_Int64("RECHARGE_LOG_COUNT")
		if int64(len(self.rechargeLog)) > RECHARGE_LOG_COUNT {
			self.SendRechargeLog()
		}
	} else {
		strLog := fmt.Sprintf("INSERT INTO log_recharge (log_Order, log_AccountID, log_RMB, log_Operator, log_State, log_Time, log_Type) VALUES ('%v',%v, %v, '%v', %v, '%v', %v)", strOrder, nAccountID, nRMB, strOperator, nState, strTime, nEmailType)
		send_tools.SQLLog(AccountMgr.StampNum(), strLog)
	}
}

func (self *emailMgr) SendRechargeLog() {

	nLen := len(self.rechargeLog)
	if nLen <= 0 {
		return
	}

	strSQL := "INSERT INTO log_recharge (log_Order, log_AccountID, log_RMB, log_Operator, log_State, log_Time, log_Type) VALUES "
	for i := 0; i < nLen; i++ {
		strNode := self.rechargeLog[i]
		strSQL += strNode
	}

	strSQL = strings.TrimRight(strSQL, ",")
	send_tools.SQLLog(strSQL)

	self.rechargeLog = make([]string, 0, 100)
}

func (self *emailMgr) Update() {
	for _, person := range self.Emails {
		for id, email := range person {
			if email.EmailType != types.EMAIL_SALESMAN.Value() && email.IsRead == 1 && email.Money == 0 {
				delete(person, id)
			}
		}
	}
}

func (self *emailMgr) SaveAll(mPlayer map[uint32]bool) {
	for accid := range mPlayer {
		self.SaveOnePlayerAllEmail(accid)
	}
}

func (self *emailMgr) SaveOnePlayerAllEmail(nAccountID uint32) {
	tPersonal := self.Emails[nAccountID]
	if tPersonal == nil {
		return
	}
	all_email := &protomsg.SAVE_EMAIL_PERSON{AccountId: nAccountID}
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
func (self *emailMgr) LoadAll(all_emails *protomsg.ALL_EMAIL_RESP) {
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
	acc := AccountMgr.GetAccountByID(recvId)
	if acc == nil {
		log.Fatalf("AddMail acc == nil :%v", recvId)
		return -1
	}
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

	if email.Money > 0 {
		strOrder := fmt.Sprintf("%v_%v", emailID, utils.MilliSecondTimeSince1970())
		self.AddRechargeLog(strOrder, recvId, email.Money, "Hall", 2, utils.DateString(), types.EmailType(email.EmailType).Value())
	}

	accMap := self.Emails[recvId]
	accMap[emailID] = email
	self.SaveOnePlayerAllEmail(recvId)

	if acc.IsOnline() {
		self.SendEmailToClient(recvId, acc.SessionId, emailID)
	}
	return int32(emailID)
}

// 仅用于后台删除邮件使用, 避免后台删除邮件自动领取元宝
func (self *emailMgr) ResetEmailRMB(accountId, emailId uint32) uint8 {
	acc := AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		return 1
	}

	person := self.Emails[accountId]
	if person == nil {
		return 2
	}

	email := person[emailId]
	if email == nil {
		return 2
	}

	if email.Money > 0 {
		strOrder := fmt.Sprintf("%v_%v", email.EmailID, utils.MilliSecondTimeSince1970())
		self.AddRechargeLog(strOrder, accountId, email.Money, "Hall", 4, utils.DateString(), types.EmailType(email.EmailType).Value())
		email.Money = 0
	}
	return 0
}

// 仅用于后台降级代理身份使用, 降级后成为玩家才删除代理邮件
func (self *emailMgr) RemoveSalesmanTypeEmail(accountId uint32) uint8 {
	acc := AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		return 1
	}

	person := self.Emails[accountId]
	if person == nil {
		return 2
	}

	for nEmailID, tEmail := range person {
		if tEmail.EmailType == types.EMAIL_SALESMAN.Value() {
			delete(person, nEmailID)
			self.SaveOnePlayerAllEmail(accountId)
			return 0
		}
	}
	return 3
}

func (self *emailMgr) RemoveMail(accountId, emailId uint32) uint8 {
	acc := AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		log.Fatalf("RemoveMail acc is nil :%v", accountId)
		return 1
	}

	person := self.Emails[accountId]
	if person == nil {
		log.Fatalf("RemoveMail email_node is nil :%v", accountId)
		return 2
	}

	email := person[emailId]
	if email == nil {
		// 非异常, 多次点击删除
		return 2
	}

	if email.EmailType == types.EMAIL_SALESMAN.Value() {
		return 2
	}

	if acc.CanAddCharge() == false {
		return 2
	}

	if email.Money > 0 {
		eOperate := common.EOperateType_EMAILL
		if email.EmailType == types.EMAIL_REBATE.Value() {
			eOperate = common.EOperateType_REBATE
		} else if email.EmailType == types.EMAIL_OFFLINE_CHARGE.Value() {
			eOperate = common.EOperateType_OFFLINE_CHARGE
		} else if email.EmailType == types.EMAIL_ONLINE_CHARGE.Value() {
			eOperate = common.EOperateType_ONLINE_CHARGE
		}
		acc.AddMoney(int64(email.Money), common.EOperateType(eOperate))
		event.Dispatcher.Dispatch(event.UpdateCharge{AccountID: accountId, RMB: int64(email.Money)}, event.EventType_UpdateCharge)
		log.Infof("Remove EmailID:%v, RecvID:%v, EmailType:%v, Content:%v, nRMB:%v", emailId, accountId, email.EmailType, email.Content, email.Money)

		strOrder := fmt.Sprintf("%v_%v", email.EmailID, utils.MilliSecondTimeSince1970())
		self.AddRechargeLog(strOrder, accountId, email.Money, "Hall", 3, utils.DateString(), types.EmailType(email.EmailType).Value())
	}

	delete(person, emailId)
	if email.Money > 0 {
		email.Money = 0
	}
	self.SaveOnePlayerAllEmail(accountId)
	return 0
}

// 领取邮件
func (self *emailMgr) GetMailReward(accountID, emailId uint32) uint8 {
	acc := AccountMgr.GetAccountByID(accountID)
	if acc == nil {
		log.Fatalf("RemoveMail acc == nil :%v", accountID)
		return 1
	}

	personal := self.Emails[accountID]
	if personal == nil {
		log.Warnf("GetMailReward personal == nil accId:%v", accountID)
		return 2
	}

	email := personal[emailId]
	if email == nil {
		return 2
	}

	if acc.CanAddCharge() == false {
		log.Warnf("GetMailReward AccountID:%v, RoomID:%v, GameType:%v", accountID, acc.RoomID, acc.GameType)
		return 3
	}

	email.IsRead = 1
	if email.Money <= 0 {
		log.Warnf("GetMailReward email.Money <= 0 email.Money:%v", email.Money)
		return 3
	}

	eOperate := common.EOperateType_EMAILL
	if email.EmailType == types.EMAIL_REBATE.Value() {
		eOperate = common.EOperateType_REBATE
	} else if email.EmailType == types.EMAIL_OFFLINE_CHARGE.Value() {
		eOperate = common.EOperateType_OFFLINE_CHARGE
	} else if email.EmailType == types.EMAIL_ONLINE_CHARGE.Value() {
		eOperate = common.EOperateType_ONLINE_CHARGE
	}
	acc.AddMoney(int64(email.Money), common.EOperateType(eOperate))
	event.Dispatcher.Dispatch(event.UpdateCharge{AccountID: accountID, RMB: int64(email.Money)}, event.EventType_UpdateCharge)
	log.Infof("Remove EmailID:%v, RecvID:%v, EmailType:%v, Content:%v, nRMB:%v", emailId, accountID, email.EmailType, email.Content, email.Money)

	strOrder := fmt.Sprintf("%v_%v", email.EmailID, utils.MilliSecondTimeSince1970())
	self.AddRechargeLog(strOrder, accountID, email.Money, "Hall", 3, utils.DateString(), types.EmailType(email.EmailType).Value())

	delete(personal, emailId)
	email.Money = 0
	self.SaveOnePlayerAllEmail(accountID)
	return 0
}

// 阅读邮件
func (self *emailMgr) ReadMail(accountID, emailId uint32) {
	acc := AccountMgr.GetAccountByID(accountID)
	if acc == nil {
		log.Fatalf("RemoveMail acc == nil :%v", accountID)
		return
	}

	personal := self.Emails[accountID]
	if personal == nil {
		return
	}

	email := personal[emailId]
	if email == nil {
		return
	}

	email.IsRead = 1
	self.SaveOnePlayerAllEmail(accountID)
}

// 获取玩家未读邮件数量
func (self *emailMgr) GetPlayerUnReadEmailNum(accountID uint32) int {
	acc := AccountMgr.GetAccountByID(accountID)
	if acc == nil {
		log.Fatalf("GetPlayerUnReadEmailNum acc == nil :%v", accountID)
		return 0
	}

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

// 是否有未领取的充值邮件
// 无需要提醒邮件, 返回0
// 返回1, 充值邮件提醒
// 返回2, 充值返回提醒
func (self *emailMgr) HasChargeEmail(accountID uint32) uint8 {
	acc := AccountMgr.GetAccountByID(accountID)
	if acc == nil {
		log.Fatalf("GetPlayerUnReadEmailNum acc == nil :%v", accountID)
		return 0
	}

	personal := self.Emails[accountID]
	if personal == nil {
		return 0
	}

	for _, email := range personal {
		if email.EmailType == types.EMAIL_OFFLINE_CHARGE.Value() || email.EmailType == types.EMAIL_ONLINE_CHARGE.Value() {
			return 1
		} else if email.EmailType == types.EMAIL_EXCHANGE_RETURN.Value() {
			return 2
		}
	}
	return 0
}

// 通知在线玩家有新邮件
func (self *emailMgr) SendEmailToClient(accountID uint32, session int64, emailId uint32) {
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_SEND_EMALL_LIST.UInt16())

	personal := self.Emails[accountID]
	if personal == nil {
		send.WriteUInt16(0)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	if emailId == 0 {
		count := uint16(0)
		for range personal {
			count++
		}
		send.WriteUInt16(count)
		if count != 0 {
			for id, email := range personal {
				content := email.Content
				if content != "" {
					bytes, _ := base64.StdEncoding.DecodeString(content)
					content = string(bytes)
				}

				send.WriteUInt32(id)
				send.WriteUInt8(uint8(email.EmailType))
				send.WriteString(content)
				send.WriteInt64(email.SendTime)
				send.WriteInt64(int64(email.Money))
				send.WriteUInt8(uint8(email.IsRead))
			}
		}
		send_tools.Send2Account(send.GetData(), session)
	} else {
		email := personal[emailId]
		if email == nil {
			return
		}
		content := email.Content
		if content != "" {
			bytes, _ := base64.StdEncoding.DecodeString(content)
			content = string(bytes)
		}

		send.WriteUInt16(1)
		send.WriteUInt32(emailId)
		send.WriteUInt8(uint8(email.EmailType))
		send.WriteString(content)
		send.WriteInt64(email.SendTime)
		send.WriteInt64(int64(email.Money))
		send.WriteUInt8(uint8(email.IsRead))

		send_tools.Send2Account(send.GetData(), session)

		if email.EmailType == types.EMAIL_OFFLINE_CHARGE.Value() || email.EmailType == types.EMAIL_ONLINE_CHARGE.Value() {
			send2 := packet.NewPacket(nil)
			send2.SetMsgID(protomsg.Old_MSGID_EXCHANGE_ORDER_LIST_TIPS.UInt16())
			send2.WriteUInt8(1) // 充值邮件提醒类型
			send_tools.Send2Account(send2.GetData(), session)
		} else if email.EmailType == types.EMAIL_EXCHANGE_RETURN.Value() {
			send2 := packet.NewPacket(nil)
			send2.SetMsgID(protomsg.Old_MSGID_EXCHANGE_ORDER_LIST_TIPS.UInt16())
			send2.WriteUInt8(2) // 充值返回提醒类型
			send_tools.Send2Account(send2.GetData(), session)
		}
	}
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

func (self *emailMgr) LogAllEmail() {

	for accid, personMap := range self.Emails {
		all_email := &protomsg.SAVE_EMAIL_PERSON{AccountId: accid}
		for _, email := range personMap {
			all_email.Emails = append(all_email.Emails, email)
		}
		sBinaryContent, _ := proto.Marshal(all_email)
		log.Infof("REPLACE INTO gd_email (gd_AccountID, gd_Email) VALUES(%v, '%v')", accid, sBinaryContent)
	}
}
