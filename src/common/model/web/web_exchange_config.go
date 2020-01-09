package web

import (
	"root/core/db"
	"encoding/json"
	"errors"
)

type Exchange_configModel struct {
	ID              uint32 `gorm:"column:gd_id;primary_key"` //序号
	Exchange_config string `gorm:"column:gd_exchange_config;type:text"`
}

func (self *Exchange_configModel) Reset()         { self = &Exchange_configModel{} }
func (self *Exchange_configModel) String() string { return "" }
func (self *Exchange_configModel) ProtoMessage()  {}

//自定义表名
func (self *Exchange_configModel) TableName() string {
	return "gd_exchange_config"
}

func GetExchange_config() *Exchange_configModel {
	all := []*Exchange_configModel{}
	conn := db.GetWeb()
	if conn == nil {
		return nil
	}

	conn.Find(&all)
	if len(all) == 1 {
		return all[0]
	}
	return nil
}
func GetExchange_config_json() string {
	data := GetExchange_config()
	if data == nil {
		return ""
	}
	js, _ := json.Marshal(data)

	return string(js)
}

// 回存数据
func (self *Exchange_configModel) Save() error {
	conn := db.GetWeb()
	if conn == nil {
		return errors.New("no db connect")
	}

	err := conn.Save(self).Error
	return err
}
