package room

import (
	"root/common"
	"root/core"
	"root/core/log"
	"root/core/packet"
	"root/protomsg"
	"root/server/hongbao/send_tools"
	"strconv"
)

var RoomMgr = NewRoomMgr()

type (
	roomMgr struct {
		roomActorId   map[uint32]int32 // key roomId value actorId
		rooms         map[int32]*Room  // key roomId value Room
		Water_line    int64
		IsMaintenance bool
	}
)

func NewRoomMgr() *roomMgr {
	return &roomMgr{
		roomActorId:   make(map[uint32]int32),
		rooms:         make(map[int32]*Room),
		IsMaintenance: false,
	}
}

func (self *roomMgr) InitRoomMgr() {
	nServerID, _ := strconv.Atoi(core.Appname)
	tSend := packet.NewPacket(nil)
	tSend.SetMsgID(protomsg.MSGID_GET_ONE_WATERLINE.UInt16())
	tSend.WriteUInt16(uint16(nServerID))
	send_tools.Send2Hall(tSend.GetData())

}

func (self *roomMgr) Room(roomId uint32) *Room {
	return self.rooms[int32(roomId)]
}
func (self *roomMgr) SaveWaterLine() {
	nServerID, _ := strconv.Atoi(core.Appname)
	tSend := packet.NewPacket(nil)
	tSend.SetMsgID(protomsg.MSGID_SET_ONE_WATERLINE.UInt16())
	tSend.WriteUInt16(uint16(nServerID))
	tSend.WriteUInt8(common.EGameTypeHONG_BAO.Value())
	tSend.WriteString(strconv.FormatInt(self.Water_line, 10))
	send_tools.Send2Hall(tSend.GetData())
}

func (self *roomMgr) CreateRoom(accountId uint32, gameType uint8, id uint32, strParam string, matchType uint8, clubID uint32) *Room {
	self.roomActorId[id] = int32(id)
	room := NewRoom(id)
	room.gameType = gameType
	room.matchType = matchType
	room.param = strParam
	room.clubID = clubID
	log.Debugf("创建房间id:%v  param:%v", id, strParam)
	self.rooms[int32(id)] = room
	return room
}
func (self *roomMgr) RoomCount() int {
	return len(self.roomActorId)
}

func (self *roomMgr) RoomActorId(roomId uint32) int32 {
	return self.roomActorId[roomId]
}
