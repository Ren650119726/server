package web

import (
	"root/core/db"
	"encoding/json"
	"errors"
)

type PayChannelModel struct {
	ID          uint32 `gorm:"column:gd_ID;primary_key"`                               //序号
	PayChannel  uint8  `gorm:"column:gd_PayChannel;type:tinyint; not null; default:0"` //支付渠道（1支付宝，2微信，3银联）
	PayType     string `gorm:"column:gd_PayType; type:varchar(50)"`                    //支付类型（1小水滴，2秀付）
	StartTime   string `gorm:"column:gd_StartTime;type:varchar(100)"`                  //启用时间
	EndTime     string `gorm:"column:gd_EndTime;type:varchar(100)"`                    //禁用时间
	State       uint8  `gorm:"column:gd_State;type:tinyint"`                           //状态（0禁用，1启用）
	RMB         string `gorm:"column:gd_RMB;type:varchar(255)"`                        //支持金额
	PayURL      string `gorm:"column:gd_PayURL;type:varchar(255)"`                     //支付URL
	Time        string `gorm:"column:gd_Time;type:varchar(100)"`                       //支付URL
	WeightValue uint32 `gorm:"column:WeightValue"`                                     //权重
}

func (self *PayChannelModel) Reset()         { self = &PayChannelModel{} }
func (self *PayChannelModel) String() string { return "" }
func (self *PayChannelModel) ProtoMessage()  {}

//自定义表名
func (self *PayChannelModel) TableName() string {
	return "gd_paychannel"
}

func GetPayChannel() []*PayChannelModel {
	all := []*PayChannelModel{}
	conn := db.GetWeb()
	if conn == nil {
		return all
	}

	conn.Find(&all)
	return all
}
func GetPayChannel_json() map[uint32]string {
	datas := GetPayChannel()
	strs := make(map[uint32]string, 0)
	for _, data := range datas {
		js, _ := json.Marshal(data)
		strs[data.ID] = string(js)
	}
	return strs
}

// 回存数据
func (self *PayChannelModel) Save() error {
	conn := db.GetWeb()
	if conn == nil {
		return errors.New("no db connect")
	}

	err := conn.Save(self).Error
	return err
}
