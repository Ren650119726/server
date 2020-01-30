package main

import (
	"bytes"
	"fmt"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/protomsg"
)

type (
	Logic struct {
		owner          *core.Actor
		init           bool // 重新建立连接是否需要拉取所有数据
		ListenActor     *core.Actor
	}
)

func NewLogic() *Logic {
	return &Logic{}
}
var Clinet_Global *Client
func (self *Logic) Init(actor *core.Actor) bool {
	self.owner = actor

	Clinet_Global = NewWebsocketClient("47.108.87.29:41000","/connect")
	Clinet_Global.connect()
	fmt.Println("connected success :",Clinet_Global.ws.RemoteAddr())
	go func() {
		for {
			recv := make([]byte,65535)
			n,err := Clinet_Global.ws.Read(recv)
			if err != nil{
				log.Warnf("err:%v",err.Error())
				continue
			}
			//log.Infof("读出%v个字节",n)
			recv = recv[0:n]
			buffer := new(bytes.Buffer)
			buffer.Write(recv)
			_, content, errcode := self.decode(buffer)
			if errcode != 0{
				log.Warnf("错误:%v",errcode)
			}
			self.HandleMessage(0, content,0)
		}
	}()

	self.owner.AddTimer(30000,-1, func(dt int64) {
		Send2Hall(protomsg.MSG_CLIENT_KEEPALIVE.UInt16(),nil)
	})
	return true
}

// 解密&处理粘包
func (self *Logic) decode(buffer *bytes.Buffer) (uint16, []byte, int) {
	buff_len := uint16(buffer.Len())
	if buff_len < packet.PACKET_HEAD_LEN { // 长度不够
		return 0, nil, 0
	}

	origin := buffer.Bytes()
	// 获取长度
	var msg_len uint16 = 0
	msg_len = uint16(origin[4])
	msg_len |= uint16(origin[5]) << 8
	if msg_len < packet.PACKET_HEAD_LEN || msg_len > packet.PACKET_BUFFER_LEN || msg_len > buff_len {
		return 0, nil, 0
	}

	content := buffer.Next(int(msg_len))

	newBuf := make([]byte, 0, msg_len)
	newBuf = append(newBuf, content...)

	// 解析msgid
	var msgid uint16 = 0
	msgid = uint16(newBuf[2])
	msgid |= uint16(newBuf[3]) << 8
	return msgid, newBuf, 0
}

func (self *Logic) registerHall() {

}

func (self *Logic) Stop() {

}

func (self *Logic) HandleMessage(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.MSG_SC_LOGIN_HALL_RES.UInt16():
		pb := packet.PBUnmarshal(pack.ReadBytes(),&protomsg.LOGIN_HALL_RES{}).(*protomsg.LOGIN_HALL_RES)
		log.Infof(colorized.Blue("登陆成功：%+v"),pb)
	}
	return false
}