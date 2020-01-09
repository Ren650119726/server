package send_tools

import (
	"root/common"
	"root/core"
	"root/core/log"
	"root/core/packet"
	"github.com/golang/protobuf/proto"
	"root/protomsg"
)

func Send2DB(msgId uint16, pb proto.Message) {
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
	core.CoreSend(0, common.EActorType_CONNECT_DB.Int32(), req.GetData(), 0)
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

// Web方将一直阻塞, 直到收到指定长度的内容或者\r\n结束阻塞
// 返回字符串error表示失败
// 返回json字符串表示成功内容
func Send2Web(str string, session int64) {
	data := str + "\r\n"
	core.CoreSend(0, common.EActorType_SERVER.Int32(), []byte(data), session)
}

func Send2Game(data []byte, session int64) {
	core.CoreSend(0, common.EActorType_SERVER.Int32(), data, session)
}

func SQLLog(sytnax string) {
	log.Infof("SQL_LOG: " + sytnax)
	pack := packet.NewPacket(nil)
	pack.SetMsgID(protomsg.Old_MSGID_SS_REQUEST_LUA.UInt16())
	pack.WriteString(sytnax)
	pack.WriteUInt8(1) // db类型(0 实例数据，1 日志数据)
	pack.WriteUInt16(0)
	core.CoreSend(0, common.EActorType_CONNECT_DB.Int32(), pack.GetData(), 0)

	// 发给大厅缓存
	cache := packet.NewPacket(nil)
	cache.SetMsgID(protomsg.MSGID_PACKET_CACHE_LIST.UInt16())
	cache.CatBody(pack)
	core.CoreSend(0, common.EActorType_MAIN.Int32(), cache.GetData(), 0)
}
