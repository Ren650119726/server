package logic

import (
	"fmt"
	"github.com/astaxie/beego"
	"root/common"
	"root/common/tools"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/network"
	"root/core/packet"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/catchfish/account"
	"root/server/catchfish/room"
	"root/server/catchfish/send_tools"
	"strconv"
)

type (
	CatchFish struct {
		owner     *core.Actor
		con_count int
	}
)

func NewHongbao() *CatchFish {
	return &CatchFish{}
}

func (self *CatchFish) Init(actor *core.Actor) bool {
	// 先处理脚本
	if core.ScriptDir != "" {
		core.InitScript(core.ScriptDir)
	}

	self.owner = actor

	// 连接hall
	hallIP := core.CoreAppConfString("connectHall")
	connectHall_actor := network.NewTCPClient(self.owner, func() string {
		return core.CoreAppConfString("connectHall")
	}, self.registerHall)
	child := core.NewActor(common.EActorType_CONNECT_HALL.Int32(), connectHall_actor, make(chan core.IMessage, 1000))
	core.CoreRegisteActor(child)
	log.Infof(colorized.Yellow("连接hall:[%v]"), hallIP)
	return true
}

// 向hall注册
func (self *CatchFish) registerHall() {
	// 发送注册消息登记自身信息
	sid, _ := strconv.Atoi(core.Appname)
	count := uint32(room.RoomMgr.RoomCount())

	GAME_TO_HALL_MAP_KEY := "fwef32f3245435"
	strSign := fmt.Sprintf("%v%v%v", sid, count, GAME_TO_HALL_MAP_KEY)
	strSign = tools.MD5(strSign)

	// 组装消息
	send_tools.Send2Hall(inner.SERVERMSG_GH_GAME_CONNECT_HALL.UInt16(), &inner.GAME_CONNECT_HALL{
		ServerID: uint32(sid),
		GameType: uint32(beego.AppConfig.DefaultInt(fmt.Sprintf("%v::gametype", sid), 0)),
	})
	log.Infof("连接大厅成功  sid:%v", sid)
	room.RoomMgr.InitRoomMgr()
	self.StartService()
}

func (self *CatchFish) StartService() {
	// 监听端口，客户端连接用
	var customer []*core.Actor
	customer = append(customer, self.owner)
	listen_actor := network.NewNetworkServer(customer, beego.AppConfig.DefaultString(core.Appname+"::listen", ""))
	child := core.NewActor(common.EActorType_SERVER.Int32(), listen_actor, make(chan core.IMessage, 10000))
	core.CoreRegisteActor(child)
}

func (self *CatchFish) Stop() {

}

func (self *CatchFish) HandleMessage(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	//case protomsg.Old_MSGID_CREATE_ROOM.UInt16(): // 大厅请求创建房间
	//	self.Old_MSGID_CREATE_ROOM(actor, msg, session)
	//case protomsg.Old_MSGID_RECV_ACCOUNT_INFO.UInt16(): // 同步账号数据
	//	self.Old_MSGID_RECV_ACCOUNT_INFO(actor, msg, session)
	//case protomsg.Old_MSGID_ENTER_GAME.UInt16(): // 客户端链接进入游戏
	//	self.Old_MSGID_ENTER_GAME(actor, msg, session)
	case protomsg.MSG_CLIENT_KEEPALIVE.UInt16():
		send_tools.Send2Account(protomsg.MSG_CLIENT_KEEPALIVE.UInt16(), nil, session)

	case inner.SERVERMSG_SS_TEST_NETWORK.UInt16():
		log.Infof("收到来自大厅的测试网络消息 SessionID:%v", session)
		req := packet.NewPacket(nil)
		req.SetMsgID(inner.SERVERMSG_SS_TEST_NETWORK.UInt16())
		send_tools.Send2Hall(inner.SERVERMSG_SS_TEST_NETWORK.UInt16(), nil)
	default: // 客户端游戏消息，统一发送给房间处理
		acc := account.AccountMgr.GetAccountBySessionID(session)
		if acc == nil {
			log.Warnf("找不到session 关联的玩家 session:%v msgId：%v", session, pack.GetMsgID())
			return false
		}
		core.CoreSend(self.owner.Id, int32(acc.RoomID), msg, session)
		break
	}
	return true
}

func (self *CatchFish) Old_MSGID_MAINTENANCE_NOTICE(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	if session != 0 {
		log.Warnf("Error, 异常session:%v 处理消息编号:%v", session, pack.GetMsgID())
		return
	}
	room.Close(nil)
}
