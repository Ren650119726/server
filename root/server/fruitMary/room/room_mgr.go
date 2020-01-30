package room

import (
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/protomsg/inner"
	"root/server/fruitMary/send_tools"
	"strconv"
)

var RoomMgr = NewRoomMgr()

/*
	因为没个房间都是单独的一个线程，原则上，所有房间的逻辑、房间内相关的数据改动、数据获取，都需要抛事件或消息给房间，让房间自己处理
    roomMgr 只做为创建房间，分流房间消息用，不能处理房间逻辑
*/
type (
	roomMgr struct {
		rooms         map[uint32]uint32  // key roomId value actorID
		Water_line    int64
		IsMaintenance bool
	}
)

func NewRoomMgr() *roomMgr {
	return &roomMgr{
		rooms:         make(map[uint32]uint32),
		IsMaintenance: false,
	}
}

func (self *roomMgr) InitRoomMgr() {
		// 初始化房间
		for id,_ := range config.Global_mary_room_config {
			self.CreateRoom(uint32(id))
		}
}

func (self *roomMgr) SendRoomInfo2Hall() {
	sid,_ := strconv.Atoi(core.Appname)
	rooms := []uint32{}
	for id,_ := range self.rooms{
		rooms = append(rooms,id)
	}
	send_tools.Send2Hall(inner.SERVERMSG_GH_ROOM_INFO.UInt16(),&inner.ROOM_INFO{
		ServerID: uint32(sid),
		RoomsID:    rooms,
	})
}
func (self *roomMgr) SaveWaterLine() {
}

func (self *roomMgr) CreateRoom(id uint32)  {
	room := NewRoom(id)
	self.rooms[id] = id
	if id < 1000{
		log.Panicf("房间ID 不能小于1000 id:%v jsonParam:%v",id)
	}
	core.CoreRegisteActor(core.NewActor(int32(id), room, make(chan core.IMessage, 5000)))
	jsonInfo := config.Global_mary_room_config[int(id)]
	log.Infof("创建房间:%v jsoninfo:%v",id,jsonInfo)
}

func (self *roomMgr) RoomCount() int {
	return len(self.rooms)
}
