package send_tools

import (
	"root/common"
	"root/core"
	"root/core/packet"
	"root/protomsg"
)

func Send2Hall(msg []byte) {
	core.CoreSend(0, common.EActorType_CONNECT_HALL.Int32(), msg, 0)
}

func Send2Account(data []byte, session int64) {

	core.CoreSend(0, common.EActorType_SERVER.Int32(), data, session)
}

func SQLLog(sytnax string) {
	pack := packet.NewPacket(nil)
	pack.SetMsgID(protomsg.Old_MSGID_SS_REQUEST_LUA.UInt16())
	pack.WriteString(sytnax)
	pack.WriteUInt8(1) // db类型(0 实例数据，1 日志数据)
	pack.WriteUInt16(0)
	core.CoreSend(0, common.EActorType_CONNECT_DB.Int32(), pack.GetData(), 0)
}
