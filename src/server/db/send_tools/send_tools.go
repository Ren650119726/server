package send_tools

import (
	"root/common"
	"root/core"
	"root/core/log"
	"root/core/packet"
	"github.com/golang/protobuf/proto"
)

var Hall_session int64

func Send2Hall(msgId uint16, pb proto.Message) {
	var bytes []byte
	if pb == nil {
		bytes = []byte{}
	} else {
		data, error := proto.Marshal(pb)
		if error != nil {
			log.Errorf("发送数据出错 :%v", error.Error())
			return
		}
		bytes = data
	}

	req := packet.NewPacket(nil)
	req.SetMsgID(msgId)
	req.WriteBytes(bytes)
	core.CoreSend(0, common.EActorType_SERVER.Int32(), req.GetData(), Hall_session)
}

func Send2Hall_pack(msgId uint16, pack packet.IPacket) {
	core.CoreSend(0, common.EActorType_SERVER.Int32(), pack.GetData(), Hall_session)
}
