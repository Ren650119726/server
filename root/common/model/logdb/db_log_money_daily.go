package logdb

type MoneyDailyModel struct {
	ID      uint   `gorm:"column:log_ID; primary_key; auto_increment"` //序号
	RMB     int64  `gorm:"column:log_RMB; not null"`
	SafeRMB int64  `gorm:"column:log_SafeRMB; not null"`
	Time    string `gorm:"column:log_Time; type:varchar(20); not null; default:'2016-01-01 23:59:59'"`
}

func (self *MoneyDailyModel) Reset()         { self = &MoneyDailyModel{} }
func (self *MoneyDailyModel) String() string { return "" }
func (self *MoneyDailyModel) ProtoMessage()  {}

//自定义表名
func (self *MoneyDailyModel) TableName() string {
	return "log_money_daily"
}
