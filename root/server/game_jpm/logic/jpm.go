package logic

import (
	"fmt"
	"github.com/astaxie/beego"
	"root/common"
	"root/common/config"
	"root/common/tools"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/network"
	"root/core/packet"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/game_jpm/account"
	"root/server/game_jpm/room"
	"root/server/game_jpm/send_tools"
	"strconv"
)

type (
	jpm struct {
		owner *core.Actor
		init  bool // 是否是第一次启动程序
		close bool // 关服
	}
)

func Newjpm() *jpm {
	return &jpm{}
}

func (self *jpm) Init(actor *core.Actor) bool {
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
func (self *jpm) registerHall() {
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
	if !self.init {
		room.RoomMgr.InitRoomMgr()
		self.StartService()
	}
	room.RoomMgr.SendRoomInfo2Hall()

	self.init = true
}

func (self *jpm) StartService() {
	// 监听端口，客户端连接用
	var customer []*core.Actor
	customer = append(customer, self.owner)
	listen_actor := network.NewTCPServer(customer, beego.AppConfig.DefaultString(core.Appname+"::listen", ""),
		beego.AppConfig.DefaultString(core.Appname+"::listenHttp", ""))
	room.ServerActor = core.NewActor(common.EActorType_SERVER.Int32(), listen_actor, make(chan core.IMessage, 10000))
	core.CoreRegisteActor(room.ServerActor)
}

func (self *jpm) Stop() {

}

func (self *jpm) HandleMessage(actor int32, msg []byte, session int64) bool {
	if self.close {
		return true
	}
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case inner.SERVERMSG_HG_NOTIFY_ALTER_DATE.UInt16(): // 大厅通知修改玩家数据
		data := packet.PBUnmarshal(pack.ReadBytes(), &inner.NOTIFY_ALTER_DATE{}).(*inner.NOTIFY_ALTER_DATE)
		core.CoreSend(self.owner.Id, int32(data.GetRoomID()), msg, session)
	case inner.SERVERMSG_SS_RELOAD_CONFIG.UInt16():
		config.Load_Conf()
		room.RoomMgr.BraodcastReload()
	case inner.SERVERMSG_SS_CLOSE_SERVER.UInt16():
		self.close = true
		self.owner.AddTimer(1000, -1, func(dt int64) {
			if room.RoomMgr.RoomCount() == 0 {
				log.Infof("所有房间关闭完成，可以关闭服务器!")
				self.owner.Suspend()
			}
		})
		for _, actor := range room.RoomMgr.Rooms {
			core.CoreSend(self.owner.Id, int32(actor), msg, session)
		}
	case protomsg.MSG_CLIENT_KEEPALIVE.UInt16(): // 心跳
		send_tools.Send2Account(protomsg.MSG_CLIENT_KEEPALIVE.UInt16(), nil, session)
	case inner.SERVERMSG_HG_PLAYER_DATA_REQ.UInt16(): // 大厅发送玩家数据
		self.SERVERMSG_HG_PLAYER_DATA_REQ(actor, pack.ReadBytes(), session)
	case protomsg.JPMMSG_CS_ENTER_GAME_JPM_REQ.UInt16(): // 请求进入小玛利房间
		actor := self.JPMMSG_CS_ENTER_GAME_JPM_REQ(actor, pack.ReadBytes(), session)
		core.CoreSend(self.owner.Id, actor, msg, session)
	case inner.SERVERMSG_HG_ROOM_BONUS_RES.UInt16(): // 大厅返回水池金额
		data := packet.PBUnmarshal(pack.ReadBytes(), &inner.ROOM_BONUS_RES{}).(*inner.ROOM_BONUS_RES)
		core.CoreSend(self.owner.Id, int32(data.GetRoomID()), msg, session)
	case protomsg.MSG_CLIENT_KEEPALIVE.UInt16():
		send_tools.Send2Account(protomsg.MSG_CLIENT_KEEPALIVE.UInt16(), nil, session)

	case inner.SERVERMSG_SS_TEST_NETWORK.UInt16():
		log.Infof("收到来自大厅的测试网络消息 SessionID:%v", session)
	default: // 客户端游戏消息，统一发送给房间处理
		acc := account.AccountMgr.GetAccountBySessionID(session)
		if acc == nil {
			log.Warnf("找不到session 关联的玩家 session:%v msgId：%v", session, pack.GetMsgID())
			return false
		}
		if room.RoomMgr.Exist(acc.RoomID) {
			core.CoreSend(self.owner.Id, int32(acc.RoomID), msg, session)
		}
		break
	}
	return true
}

func (self *jpm) Old_MSGID_MAINTENANCE_NOTICE(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	if session != 0 {
		log.Warnf("Error, 异常session:%v 处理消息编号:%v", session, pack.GetMsgID())
		return
	}
	room.Close(nil)
}
