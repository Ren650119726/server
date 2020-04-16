package send_tools

import (
	"github.com/golang/protobuf/proto"
	"root/common"
	"root/core"
	"root/core/log"
	"root/core/packet"
)

func Proto2PacketBytes(msgId uint16, pb proto.Message) []byte {
	var bytes []byte
	if pb == nil {
		bytes = []byte{}
	} else {
		data, error := proto.Marshal(pb)
		if error != nil {
			log.Errorf("发送数据出错 :%v", error.Error())
			return nil
		}
		bytes = data
	}
	req := packet.NewPacket(nil)
	req.SetMsgID(msgId)
	req.WriteBytes(bytes)
	return req.GetData()
}

func Send2Hall(msgId uint16, pb proto.Message) {
	core.CoreSend(0, common.EActorType_CONNECT_HALL.Int32(), Proto2PacketBytes(msgId, pb), 0)
}

func Send2Main(msgId uint16, pb proto.Message) {
	core.CoreSend(0, common.EActorType_MAIN.Int32(), Proto2PacketBytes(msgId, pb), 0)
}

func Send2Room(msgId uint16, pb proto.Message, roomID int32) {
	core.CoreSend(0, roomID, Proto2PacketBytes(msgId, pb), 0)
}

func Send2Account(msgId uint16, pb proto.Message, session int64) {
	core.CoreSend(0, common.EActorType_SERVER.Int32(), Proto2PacketBytes(msgId, pb), session)
}
