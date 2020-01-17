package logdb

type RechargeModel struct {
	ID        uint32 `gorm:"column:log_ID; primary_key; auto_increment"`                                 //序号
	Order     string `gorm:"column:log_Order; TYPE:varchar(64) NOT NULL; DEFAULT ''; unique_index"`      //唯一索引, 订单号
	AccountID uint32 `gorm:"column:log_AccountID; not null; default 0"`                                  //帐号ID
	RMB       int64  `gorm:"column:log_RMB; not null; default 0"`                                        //邮件元宝
	Operator  string `gorm:"column:log_Operator; TYPE:varchar(100) not null; default:''"`                //操作者
	State     uint8  `gorm:"column:log_State; not null; default 0"`                                      //状态
	Time      string `gorm:"column:log_Time; type:varchar(20); not null; default:'2016-01-01 23:59:59'"` //日志时间
	Type      uint8  `gorm:"column:log_Type; not null; default 0"`                                       //操作类型
}

func (self *RechargeModel) Reset()         { self = &RechargeModel{} }
func (self *RechargeModel) String() string { return "" }
func (self *RechargeModel) ProtoMessage()  {}

//自定义表名
func (self *RechargeModel) TableName() string {
	//strSuffix := time.Now().Format("2006-01-02")
	//return "log_recharge_" + strSuffix
	return "log_recharge"
}
