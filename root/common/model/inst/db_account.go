package inst

import (
	"root/core/db"
	"errors"
)

// 账号表
type AccountModel struct {
	AccountId          uint32 `gorm:"column:gd_AccountID; primary_key;"`                                              //帐号ID,主键,唯一
	UnDevice           string `gorm:"column:gd_UnDevice; type:varchar(64); not null; default:''"`                     //账号绑定设备唯一码
	Phone      	       string `gorm:"column:gd_Phone; type:varchar(26); not null; default:''"`                        //帐号绑定手机字符串
	WeiXin     		   string `gorm:"column:gd_WeiXin; type:varchar(64); not null; default:''"`                       //帐号绑定微信唯一码
	Name       		   string `gorm:"column:gd_Name; type:varchar(60); not null; default:''"`                         //名字
	HeadURL       	   string `gorm:"column:gd_URL; type:varchar(100); not null; default:''"`                         //头像URL
	Money      		   uint64 `gorm:"column:gd_Money; not null; default:0"`                                           //货币
	SafeMoney    	   uint64 `gorm:"column:gd_SafeMoney; not null; default:0"`                                       //保险箱货币
	ActiveTime 		   string `gorm:"column:gd_ActiveTime; type:varchar(20); not null; default:'2018-11-01 23:59:59'"` //账号激活时间
	ActiveIP   		   string `gorm:"column:gd_ActiveIP; type:varchar(20); not null; default:'0.0.0.0'"`              //激活帐号IP
	FrozenTime         uint64 `gorm:"column:gd_FrozenTime; not null; default:0"`                                      //冻结限制到期时间
	Salesman           int8  `gorm:"column:gd_Salesman; not null; default:0"`                                        //代理身份 0非代理 1级代理 2级代理
	SalesTime          string `gorm:"column:gd_SalesTime; type:varchar(20); not null; default:'2018-11-01 23:59:59'"` //成为推广员时间
	Special            uint8  `gorm:"column:gd_Special; not null; default:0"`                                         //特殊账号 0 不是 1 是
	OSType             uint8  `gorm:"column:gd_OSType; not null; default:0"`                                          //系统类型 1Windos, 2安卓, 3苹果
	LoginTime          int64  `gorm:"column:gd_LoginTime; not null; default:0"`
	LogoutTime         int64  `gorm:"column:gd_LogoutTime; not null; default:0"`
	Robot              uint32 `gorm:"column:gd_Robot; not null; default:0"`
	Kill               int32 `gorm:"column:gd_Kill; not null; default:0"`
}

func (self *AccountModel) Reset()         { self = &AccountModel{} }
func (self *AccountModel) String() string { return "" }
func (self *AccountModel) ProtoMessage()  {}

//自定义表名
func (self *AccountModel) TableName() string {
	return "gd_account"
}

func GetAllAccount() []*AccountModel {
	all := []*AccountModel{}
	conn := db.GetInst()
	if conn == nil {
		return all
	}

	conn.Find(&all)
	return all
}

func (self *AccountModel) FindbyAccountID(accountId uint32) error {
	conn := db.GetInst()
	if conn == nil {
		return errors.New("no db connect")
	}

	self.AccountId = accountId
	err := conn.FirstOrInit(self, self).Error
	return err
}

func (self *AccountModel) Save() error {
	conn := db.GetInst()
	if conn == nil {
		return errors.New("no db connect")
	}

	err := conn.Save(self).Error
	return err
}
