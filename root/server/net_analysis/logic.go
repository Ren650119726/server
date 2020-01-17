package main

import (
	"root/core"
	"root/core/log"
	"root/core/packet"
)

type (
	logic struct {
		owner *core.Actor
	}
)

func NewLogic() *logic {
	return &logic{}
}

func (self *logic) Init(actor *core.Actor) bool {
	self.owner = actor

	go captrue_packet()
	core.Cmd.Regist("s", self.seat, false)
	return true
}

func (self *logic) Stop() {

}

func (self *logic) registerR2B() {
}

func (self *logic) seat([]string) {
}

func (self *logic) HandleMessage(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	default:
		log.Warnf("waitting 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}

	return true
}
