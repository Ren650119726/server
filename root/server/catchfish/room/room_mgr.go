package room

import (
	"root/core"
	"root/core/log"
	"root/protomsg/inner"
	"root/server/catchfish/send_tools"
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
	sid,_ := strconv.Atoi(core.Appname)
	rooms := []*inner.Room{}
	for id,_ := range self.rooms{
		rooms = append(rooms,&inner.Room{
			RoomID:      id,
			ServerID:    uint32(sid),
			PlayerCount: 0,	// 定期更新通知大厅或者初始化完成后，抛事件，让房间主动发送各自信息给大厅
		})
	}

	send_tools.Send2Hall(inner.SERVERMSG_GH_ROOM_INFO.UInt16(),&inner.ROOM_INFO{
		ServerID: uint32(sid),
		Rooms:    rooms,
	})
}

func (self *roomMgr) SaveWaterLine() {
}

func (self *roomMgr) CreateRoom(id,gameType,serverID uint32,jsonParam string)  {
	room := NewRoom(id,gameType,serverID,jsonParam)
	self.rooms[id] = id
	if id < 1000{
		log.Panicf("房间ID 不能小于1000 id:%v jsonParam:%v",id,jsonParam )
	}
	core.CoreRegisteActor(core.NewActor(int32(id), room, make(chan core.IMessage, 5000)))
}

func (self *roomMgr) RoomCount() int {
	return len(self.rooms)
}
