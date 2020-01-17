package inst

import (
	"root/core/db"
	"errors"
)

/*
TinyBlob 最大 255
Blob 最大 65K
MediumBlob 最大16M
LongBlob 最大 4G
*/
// 角色表
type EmailModel struct {
	AccountId uint32 `gorm:"column:gd_AccountID; primary_key"`
	Data      []byte `gorm:"column:gd_Email; type:mediumblob; not null"`
}

//自定义表名
func (self *EmailModel) TableName() string {
	return "gd_email"
}

func GetAllEmail() []*EmailModel {
	all := []*EmailModel{}
	conn := db.GetInst()
	if conn == nil {
		return all
	}

	conn.Find(&all)
	return all
}
func (self *EmailModel) GetEmail(accid uint32) error {
	conn := db.GetInst()
	if conn == nil {
		return nil
	}

	self.AccountId = accid
	err := conn.FirstOrInit(self, self).Error
	return err
}

// 回存数据
func (self *EmailModel) Save() error {
	conn := db.GetInst()
	if conn == nil {
		return errors.New("no db connect")
	}

	err := conn.Save(self).Error
	return err
}

//// 将sql角色数据转换成proto数据
//func (self *EmailModel) Convert2PB() (emails []*msgserver.GlobalEmail, IncEmailStamp int64) {
//	globalemails := &msgserver.GlobalEmails{}
//	err := proto.Unmarshal(self.Data, globalemails)
//	if err != nil {
//		log.Errorf("解析邮件数据出错:%v", err.Error())
//		return nil, 0
//	}
//
//	return globalemails.Emails, self.IncEmailStamp
//}

//// 将proto数据换换成model
//func (self *EmailModel) ConvertFromPB(models *msgserver.GlobalEmails, incEmailStamp int64) {
//	self.Data, _ = proto.Marshal(models)
//	self.IncEmailStamp = incEmailStamp
//}
