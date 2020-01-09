package logdb

type LoginModel struct {
	ID         uint32 `gorm:"column:log_ID; primary_key; auto_increment"`                            //序号
	AccountID  uint32 `gorm:"column:log_AccountID; primary_key; type:int unsigned; not null"`        //玩家AccountID
	LoginTime  string `gorm:"column:log_LoginTime; type:varchar(20); default:'2016-01-01 23:59:59'"` //登录时间
	LogoutTime string `gorm:"column:log_LogoutTime; type:varchar(20); default:''"`                   //玩家离线时间
}

func (self *LoginModel) Reset()         { self = &LoginModel{} }
func (self *LoginModel) String() string { return "" }
func (self *LoginModel) ProtoMessage()  {}

//自定义表名
func (self *LoginModel) TableName() string {
	return "log_login"
}
