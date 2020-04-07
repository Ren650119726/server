package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"root/core"
	"root/core/log"
	"root/core/packet"
	"root/protomsg"
	"strconv"
)

func init() {
	core.Cmd.Regist("login", login, true)
	core.Cmd.Regist("time", time, true)
	core.Cmd.Regist("engame", engame, true)
	core.Cmd.Regist("start1", start1, true)
	core.Cmd.Regist("show", show, true)

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
	client_Global.SendMessage(req.GetData())
}

func login(s []string) {
	if len(s) < 1 {
		fmt.Printf("× 参数错误 \r\n")
		return
	}

	acc := s[0]
	Send2Hall(protomsg.MSG_CS_LOGIN_HALL_REQ.UInt16(), &protomsg.LOGIN_HALL_REQ{
		LoginType: uint32(1), // 1 游客 2 手机 3 微信
		OSType:    1,
		Unique:    acc,
		Sign:      "",
	})
}

func time(s []string) {
	Send2Hall(protomsg.MSG_CS_SYNC_SERVER_TIME.UInt16(), nil)
}

func engame(s []string) {
	if len(s) < 1 {
		fmt.Printf("× 参数错误 \r\n")
		return
	}
	room, _ := strconv.Atoi(s[0])
	Send2Hall(protomsg.MSG_CS_ENTER_ROOM_REQ.UInt16(), &protomsg.ENTER_ROOM_REQ{RoomID: uint32(room)})
}

func start1(s []string) {
	if len(s) < 2 {
		fmt.Printf("× 参数错误 \r\n")
		return
	}
	bet, _ := strconv.Atoi(s[0])
	c, _ := strconv.Atoi(s[1])
	if c == 0 {
		c = 1
	}
	count = c
	log.Infof("请求开始:%v", c)
	Send2Game(protomsg.S777MSG_CS_START_S777_REQ.UInt16(), &protomsg.START_S777_REQ{Bet: uint64(bet)})

}
func show(s []string) {
	log.Infof("count:%v fee:%v", count, fee)
}
