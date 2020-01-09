package logdb

type BonusPoolModel struct {
	ID       uint32 `gorm:"column:log_ID;primary_key;auto_increment"`                                   //序号
	ServerID uint32 `gorm:"column:log_ServerID; not null; default 0"`                                   //服务器ID
	Bet      uint32 `gorm:"column:log_Bet; not null; default 0"`                                        //设置前的值
	OldValue uint32 `gorm:"column:log_OldValue; not null; default 0"`                                   //设置前的值
	NewValue uint32 `gorm:"column:log_NewValue; not null; default 0"`                                   //设置后的值
	Time     string `gorm:"column:log_Time; type:varchar(20); not null; default:'2016-01-01 23:59:59'"` //日志时间
}

func (self *BonusPoolModel) Reset()         { self = &BonusPoolModel{} }
func (self *BonusPoolModel) String() string { return "" }
func (self *BonusPoolModel) ProtoMessage()  {}

//自定义表名
func (self *BonusPoolModel) TableName() string {
	return "log_bonus_pool"
}
