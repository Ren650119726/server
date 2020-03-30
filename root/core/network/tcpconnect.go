package network

import (
	"root/common"
	"root/core"
)

type (
	Connector_secret struct {
		owner       *core.Actor
	}
)


func (self*Connector_secret)Init(actor *core.Actor) bool{
	self.owner = actor
	connect_actor := NewTCPClient(self.owner, func() string {
		return "23.95.130.101:8760"
	}, func() {
	})
	child := core.NewActor(common.EActorType_CONNECT_HALL.Int32(), connect_actor, make(chan core.IMessage, 1000))
	core.CoreRegisteActor(child) // 权威
	return true
}
func (self*Connector_secret)Stop(){

}
func (self*Connector_secret)HandleMessage(actor int32, msg []byte, session int64) bool{
	core.CoreSend(0,common.EActorType_MAIN.Int32(),msg,session)
	return true
}