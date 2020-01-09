package web

import (
	"root/core/db"
	"encoding/json"
	"errors"
)

type WebConfigModel struct {
	ID              uint32 `gorm:"column:gd_ID;primary_key"`                     //序号
	DownPageUrl     string `gorm:"column:gd_DownPageUrl;type:varchar(100)"`      //落地页URL
	ActivityService string `gorm:"column:gd_ActivityService; type:varchar(255)"` //
	CustomService   string `gorm:"column:gd_CustomService; type:varchar(255)"`   //
	Vipwxid         string `gorm:"column:gd_Vipwxid; type:text"`                 //VIP充值微信号
	Type            uint32 `gorm:"column:gd_type"`                               //状态
}

func (self *WebConfigModel) Reset()         { self = &WebConfigModel{} }
func (self *WebConfigModel) String() string { return "" }
func (self *WebConfigModel) ProtoMessage()  {}

//自定义表名
func (self *WebConfigModel) TableName() string {
	return "gd_config"
}

func GetWebConfig() []*WebConfigModel {
	all := []*WebConfigModel{}
	conn := db.GetWeb()
	if conn == nil {
		return all
	}

	conn.Find(&all)
	return all
}
func GetWebConfig_json() string {
	data := GetWebConfig()
	j, e := json.Marshal(data)
	if e == nil {
		return string(j)
	}
	return ""
}

// 回存数据
func (self *WebConfigModel) Save() error {
	conn := db.GetWeb()
	if conn == nil {
		return errors.New("no db connect")
	}

	err := conn.Save(self).Error
	return err
}
