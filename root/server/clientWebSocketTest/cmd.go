package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"root/core"
	"root/core/log"
	"root/core/packet"
	"root/protomsg"
)

func init() {
	core.Cmd.Regist("login", login, true)
}

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
	Clinet_Global.SendMessage(req.GetData())
}

func login(s []string) {
	if len(s) < 1 {
		fmt.Printf("× 参数错误 \r\n")
		return
	}

	acc := s[0]
	Send2Hall(protomsg.MSG_CS_LOGIN_HALL_REQ.UInt16(),&protomsg.LOGIN_HALL_REQ{
		LoginType: uint32(1),	// 1 游客 2 手机 3 微信
		OSType:    1,
		Unique:    acc,
		Sign:      "",
	})
}
