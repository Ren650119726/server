package send_tools

import (
	"github.com/golang/protobuf/proto"
	"root/common"
	"root/core"
	"root/core/log"
	"root/core/packet"
)

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
	core.CoreSend(0, common.EActorType_CONNECT_HALL.Int32(), req.GetData(), 0)
}

func Send2Account(msgId uint16, pb proto.Message, session int64) {
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
	core.CoreSend(0, common.EActorType_SERVER.Int32(), req.GetData(), session)
}

func Send2AccountBytes(msgId uint16, data []byte, session int64) {

	req := packet.NewPacket(nil)
	req.SetMsgID(msgId)
	req.WriteBytes(data)
	core.CoreSend(0, common.EActorType_SERVER.Int32(), req.GetData(), session)
}
