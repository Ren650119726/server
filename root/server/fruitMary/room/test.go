package room

import (
	"root/common"
	"root/common/config"
	"root/core"
	"root/core/packet"
	"root/protomsg/inner"
)

func init() {
	core.Cmd.Regist("tohall", tohall, true)
	core.Cmd.Regist("todb", todb, true)
	core.Cmd.Regist("reload", reload, true)
	core.Cmd.Regist("info", info, true)
	core.Cmd.Regist("closegame", Close, true)

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
	config.LoadPublic_Conf()
}

func info(s []string) {

}
func Close(s []string) {
	a := core.GetActor(int32(common.EActorType_SERVER))
	if a != nil {
		a.Suspend()
	}
}
