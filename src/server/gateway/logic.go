package main

import (
	"root/common"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/network"
	"root/core/packet"
	"root/core/utils"
	"fmt"
	"github.com/astaxie/beego"
)

const GATEWAY = 100

type (
	group struct {
		c_session int64
		s_actorId int64
	}

	logic struct {
		owner     *core.Actor
		con_count int

		cmap map[int64]*group
		smap map[int64]*group

		inc_actor int64
	}
)

func NewLogic() *logic {
	return &logic{}
}

func (self *logic) Init(actor *core.Actor) bool {
	self.owner = actor
	self.cmap = make(map[int64]*group)
	self.smap = make(map[int64]*group)
	self.inc_actor = 1000

	// 监听端口，客户端连接用
	var customer []*core.Actor
	customer = append(customer, self.owner)
	listen_actor := network.NewTCPServer(customer, beego.AppConfig.DefaultString(core.Appname+"::gateway", ""))
	child := core.NewActor(common.EActorType_SERVER.Int32(), listen_actor, make(chan core.IMessage, 10000))
	core.CoreRegisteActor(child)

	return true
}

func (self *logic) Stop() {

}

func (self *logic) HandleMessage(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)

	// c端来的消息，直接流给s处理
	if actor == common.EActorType_SERVER.Int32() {
		if pack.GetMsgID() == utils.ID_DISCONNECT {
			g, _ := self.cmap[session]
			delete(self.cmap, session)
			log.Infof("client 断开链接 sess:%v", session)
			if g != nil {
				log.Infof("清除完成 sess:%v actor:%v", session, g.s_actorId)
				delete(self.smap, g.s_actorId)
				if g.s_actorId != -1 {
					a := core.GetActor(int32(g.s_actorId))
					if a != nil {
						a.Suspend()
					}

				}
			}
			return true
		}

		conIP := core.CoreAppConfString("listen")
		// 所有客户端直接转发给服务器
		g, e := self.cmap[session]
		if !e {

			self.inc_actor++
			actorId := self.inc_actor

			newg := &group{
				c_session: session,
				s_actorId: actorId,
			}

			self.cmap[session] = newg
			self.smap[actorId] = newg
			connect_actor := network.NewTCPClient(self.owner,
				func() string {
					return core.CoreAppConfString("listen")
				}, func() {
					log.Infof(colorized.Yellow("c:[%v]sess:%v ----> s:[%v]映射建立成功actorId:[%v] 附带发送msgId:%v"), core.GetRemoteIP(newg.c_session), newg.c_session, conIP, actorId, pack.GetMsgID())
					core.CoreSend(actor, int32(actorId), pack.GetData(), session)
				})

			child := core.NewActor(int32(actorId), connect_actor, make(chan core.IMessage, 1000))
			log.Infof(colorized.White("初始化c:[%v]到s:[%v]的映射... inc_actor:%v"), core.GetRemoteIP(session), conIP, self.inc_actor)
			core.CoreRegisteActor(child)
		} else {
			s := g.s_actorId
			if s == -1 {
				log.Warnf(" c:[%v] ----> s:[%v]映射未建立成功 客户发来的消息msgId:%v", core.GetRemoteIP(g.c_session), conIP, pack.GetMsgID())
				return false
			} else {
				core.CoreSend(actor, int32(s), msg, session)
				if pack.GetMsgID() != 2110 {
					fmt.Printf("c->s:%v actor:%v session:%v\n", pack.GetMsgID(), actor, g.c_session)
				}
			}
		}
	} else { // s端来的消息，直接流给c
		g, e := self.smap[int64(actor)]
		if !e {
			log.Warnf("不存在的映射：%v", actor)
		} else {
			core.CoreSend(actor, common.EActorType_SERVER.Int32(), msg, g.c_session)
			if pack.GetMsgID() != 2110 {
				fmt.Printf("s->c:%v actor:%v session:%v\n", pack.GetMsgID(), actor, g.c_session)
			}
		}

	}

	return true
}
