package inst

import (
	"errors"
	"root/core/db"
)

// 水位线表
type WaterLineModel struct {
	RoomID    uint32 `gorm:"column:gd_RoomID; primary_key"`
	WaterLine string `gorm:"column:gd_WaterLine; type:varchar(255); not null; default:''"`
}

func (self *WaterLineModel) Reset()         { self = &WaterLineModel{} }
func (self *WaterLineModel) String() string { return "" }
func (self *WaterLineModel) ProtoMessage()  {}

//自定义表名
func (self *WaterLineModel) TableName() string {
	return "gd_water_line"
}

func GetAllWaterLine() []*WaterLineModel {
	all := []*WaterLineModel{}
	conn := db.GetInst()
	if conn == nil {
		return all
	}

	conn.Find(&all)
	return all
}

// 回存数据
func (self *WaterLineModel) Save() error {
	conn := db.GetInst()
	if conn == nil {
		return errors.New("no db connect")
	}

	err := conn.Save(self).Error
	return err
}
