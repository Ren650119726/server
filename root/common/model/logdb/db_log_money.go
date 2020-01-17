package logdb

import (
	"strconv"
)

type MoneyModel struct {
	ID          uint32 `gorm:"column:log_ID; primary_key; auto_increment"`                                 //序号
	AccountID   uint32 `gorm:"column:log_AccountID; primary_key; TYPE:int unsigned; not null"`             //玩家AccountID
	ChangeValue int64  `gorm:"column:log_ChangeValue; not null; default 0"`                                //改变值
	Value       int64  `gorm:"column:log_Value; not null; default 0"`                                      //改变后剩余值
	Operate     uint8  `gorm:"column:log_Operate; not null; default 0"`                                    //改变原因
	Time        string `gorm:"column:log_Time; type:varchar(20); not null; default:'2016-01-01 23:59:59'"` //改变时间
	RoomID      uint32 `gorm:"column:log_RoomID; not null; default 0"`                                     //房间ID
	logPortion  int
}

func (self *MoneyModel) Reset()                    { self = &MoneyModel{} }
func (self *MoneyModel) String() string            { return "" }
func (self *MoneyModel) ProtoMessage()             {}
func (self *MoneyModel) Portion(i int) *MoneyModel { self.logPortion = i; return self }

//自定义表名
func (self *MoneyModel) TableName() string {
	//strSuffix := time.Now().Format("2006-01-02")
	//return "log_rmb_" + strSuffix
	return "log_money_" + strconv.Itoa(int(self.logPortion))
}