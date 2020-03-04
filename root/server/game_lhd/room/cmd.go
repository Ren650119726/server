package room

import (
	"root/common"
	"root/common/config"
	"root/core"
	"root/core/packet"
	"root/protomsg/inner"
)

var ServerActor *core.Actor

func init() {
	core.Cmd.Regist("tohall", tohall, true)
	core.Cmd.Regist("todb", todb, true)
	core.Cmd.Regist("reload", reload, true)
	core.Cmd.Regist("info", info, true)
	core.Cmd.Regist("stop", Close, true)
}

func tohall(s []string) {
	send := packet.NewPacket(nil)
	send.SetMsgID(inner.SERVERMSG_SS_TEST_NETWORK.UInt16())
	core.CoreSend(0, common.EActorType_CONNECT_HALL.Int32(), send.GetData(), 0)
}

func todb(s []string) {
	send := packet.NewPacket(nil)
	send.SetMsgID(inner.SERVERMSG_SS_TEST_NETWORK.UInt16())

	core.CoreSend(0, common.EActorType_CONNECT_DB.Int32(), send.GetData(), 0)
}

func reload(s []string) {
	config.Load_Conf()

	msg := packet.NewPacket(nil)
	msg.SetMsgID(inner.SERVERMSG_SS_RELOAD_CONFIG.UInt16())
	for roomID, _ := range RoomMgr.rooms {
		core.CoreSend(0, int32(roomID), msg.GetData(), 0)
	}
}

func info(s []string) {

}

func Close(s []string) {
	ServerActor.Suspend()

	send := packet.NewPacket(nil)
	send.SetMsgID(inner.SERVERMSG_SS_CLOSE_SERVER.UInt16())
	core.CoreSend(0, common.EActorType_MAIN.Int32(), send.GetData(), 0)
	for _, room := range RoomMgr.rooms {
		core.CoreSend(0, int32(room), send.GetData(), 0)
	}
}
