package room

import (
	"root/common"
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/packet"
	"root/protomsg"
	"root/server/mahjong-panda/send_tools"
	"strconv"
)

func init() {
	core.Cmd.Regist("room", createRoom, false)
	core.Cmd.Regist("info", infoRoom, false)
	core.Cmd.Regist("list", roomList, false)
	core.Cmd.Regist("closegame", CloseServer, false)

	core.Cmd.Regist("tohall", tohall, false)
	core.Cmd.Regist("todb", todb, false)
	core.Cmd.Regist("reload", reload, false)
	core.Cmd.Regist("on", online, false)
	core.Cmd.Regist("setbonus", setbouns, false)
	core.Cmd.Regist("sp", sp, false)
}

// 创建房间
func createRoom([]string) {
	p := packet.NewPacket(nil)
	p.SetMsgID(protomsg.Old_MSGID_CREATE_ROOM.UInt16())
	p.WriteUInt32(0)
	p.WriteUInt32(999)
	p.WriteUInt8(1)
	p.WriteUInt8(2)
	p.WriteString("1|0|100|200|5000|400|0")
	p.WriteUInt8(1)
	p.WriteUInt32(0)
	core.CoreSend(0, common.EActorType_MAIN.Int32(), p.GetData(), 0)
}

// 房间信息
func infoRoom(s []string) {
	roomid, _ := strconv.Atoi(s[0])
	Room := RoomMgr.Room(uint32(roomid))
	if Room == nil {
		log.Warnf("找不到房间%v", roomid)
	} else {

	}
}

// 房间信息
func roomList(s []string) {
	for _, room := range RoomMgr.roomActor {
		if room != nil {
			log.Infof("rooms:%v 坐下的人:%v 总人数:%v", room.roomId, room.sitDownCount(), len(room.accounts))
		}

	}
}
func tohall(s []string) {
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_SS_TEST_NETWORK.UInt16())
	core.CoreSend(0, common.EActorType_CONNECT_HALL.Int32(), send.GetData(), 0)
}

func todb(s []string) {
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_SS_TEST_NETWORK.UInt16())

	core.CoreSend(0, common.EActorType_CONNECT_DB.Int32(), send.GetData(), 0)
}

func CloseServer(s []string) {
	for _, room := range RoomMgr.roomActor {
		room.Close()
	}
	RoomMgr.IsMaintenance = true

	nServerID, _ := strconv.Atoi(core.Appname)
	tSend := packet.NewPacket(nil)
	tSend.SetMsgID(protomsg.MSGID_GH_SET_MAINTENANCE.UInt16())
	tSend.WriteUInt16(uint16(nServerID))
	tSend.WriteUInt8(1) // 设置进入维护状态
	send_tools.Send2Hall(tSend.GetData())

	a := core.GetActor(int32(common.EActorType_SERVER))
	if a != nil {
		a.Suspend()
	}
}

func reload(s []string) {
	config.LoadPublic_Conf()
}

func online(s []string) {
	RoomMgr.OnlineStatics()
}

func setbouns(s []string) {
	bet, _ := strconv.Atoi(s[0])
	addition, _ := strconv.Atoi(s[1])
	core.LocalCoreSend(0, common.EActorType_MAIN.Int32(), func() {
		RoomMgr.Set_bonus(uint32(bet), uint64(addition), false)
	})
}

func sp(s []string) {
	roomid, _ := strconv.Atoi(s[0])
	Room := RoomMgr.Room(uint32(roomid))
	if Room == nil {
		log.Warnf("找不到房间%v", roomid)
	} else {
		Room.t = true
	}
}
