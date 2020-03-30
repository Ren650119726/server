package main

import (
	"root/common"
	"root/core"
	"root/core/network"
)

type (
	logic struct {
		owner     *core.Actor
	}
)

func NewLogic() *logic {
	return &logic{}
}

func (self *logic) Init(actor *core.Actor) bool {
	self.owner = actor

	// 监听端口，客户端连接用
	var customer []*core.Actor
	customer = append(customer, self.owner)
	listen_actor := network.NewNetworkServer(customer, "0.0.0.0:8760", "")
	child := core.NewActor(common.EActorType_SERVER.Int32(), listen_actor, make(chan core.IMessage, 10000))
	core.CoreRegisteActor(child)
	return true
}

func (self *logic) Stop() {

}

func (self *logic) HandleMessage(actor int32, msg []byte, session int64) bool {

	return true
}
