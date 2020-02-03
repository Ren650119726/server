package inst

import (
	"root/core/db"
	"errors"
)

// 房间奖金池
type RoomBonusModel struct {
	RoomID    uint32 `gorm:"column:gd_RoomID; primary_key"`
	Value     uint64 `gorm:"column:gd_Value; not null; default:0"`
}

func (self *RoomBonusModel) Reset()         { self = &RoomBonusModel{} }
func (self *RoomBonusModel) String() string { return "" }
func (self *RoomBonusModel) ProtoMessage()  {}

//自定义表名
func (self *RoomBonusModel) TableName() string {
	return "gd_room_bonus"
}

func GetAllRoomBonus() []*RoomBonusModel {
	all := []*RoomBonusModel{}
	conn := db.GetInst()
	if conn == nil {
		return all
	}

	conn.Find(&all)
	return all
}

// 回存数据
func (self *RoomBonusModel) Save() error {
	conn := db.GetInst()
	if conn == nil {
		return errors.New("no db connect")
	}

	err := conn.Save(self).Error
	return err
}

