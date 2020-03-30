package network

import (
	"root/common"
	"root/core"
	"root/core/log"
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
		log.Info("链接成功 23.95.130.101")
	})
	child := core.NewActor(common.EActorType_CONNECT_HALL.Int32(), connect_actor, make(chan core.IMessage, 1000))
	core.CoreRegisteActor(child)
	return true
}
func (self*Connector_secret)Stop(){

}
func (self*Connector_secret)HandleMessage(actor int32, msg []byte, session int64) bool{
	core.CoreSend(0,common.EActorType_MAIN.Int32(),msg,session)
	return true
}