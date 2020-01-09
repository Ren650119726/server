package logic

import (
	"root/common"
	"root/core"
	"root/core/packet"
	"root/protomsg"
)

func CMD_ToHall(s []string) {
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_SS_TEST_NETWORK.UInt16())
	core.CoreSend(0, common.EActorType_CONNECT_HALL.Int32(), send.GetData(), 0)
}
