package logic

import (
	"root/common"
	"root/common/config"
	"root/common/tools"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/network"
	"root/core/packet"
	"fmt"
	"github.com/astaxie/beego"
	"root/protomsg"
	"root/server/paodekuai/account"
	"root/server/paodekuai/room"
	"root/server/paodekuai/send_tools"
	"strconv"
)

type (
	PDK struct {
		owner     *core.Actor
		con_count int
	}
)

func NewPDK() *PDK {
	return &PDK{}
}

func (self *PDK) Init(actor *core.Actor) bool {
	// 先处理脚本
	if core.ScriptDir != "" {
		core.InitScript(core.ScriptDir)
	}

	self.owner = actor

	//连接hall
	hallIP := core.CoreAppConfString("connectHall")
	connectHall_actor := network.NewTCPClient(self.owner, func() string {
		return core.CoreAppConfString("connectHall")
	}, self.registerHall)
	child := core.NewActor(common.EActorType_CONNECT_HALL.Int32(), connectHall_actor, make(chan core.IMessage, 1000))
	core.CoreRegisteActor(child)
	log.Infof(colorized.Yellow("[%v]连接hall..."), hallIP)
	return true
}

// 向hall注册
func (self *PDK) registerHall() {
	// 发送注册消息登记自身信息
	serverId, _ := strconv.Atoi(core.Appname)
	count := uint32(room.RoomMgr.Room_Count())

	GAME_TO_HALL_MAP_KEY := config.GetPublicConfig_String("GAME_TO_HALL_MAP_KEY")
	strSign := fmt.Sprintf("%v%v%v", serverId, count, GAME_TO_HALL_MAP_KEY)
	strSign = tools.MD5(strSign)

	// 组装消息
	tMaping := packet.NewPacket(nil)
	tMaping.SetMsgID(protomsg.Old_MSGID_SS_MAPING.UInt16())
	tMaping.WriteUInt16(uint16(serverId))
	tMaping.WriteUInt32(count)
	tMaping.WriteString(strSign)
	core.CoreSend(self.owner.Id, common.EActorType_CONNECT_HALL.Int32(), tMaping.GetData(), 0)
	log.Infof("连接大厅成功  ServerID:%v", serverId)

	tBonus := packet.NewPacket(nil)
	tBonus.SetMsgID(protomsg.Old_MSGID_GET_ONE_BONUSPOOL.UInt16())
	tBonus.WriteUInt16(uint16(serverId))
	send_tools.Send2Hall(tBonus.GetData())
	self.StartService()
}

func (self *PDK) StartService() {
	// 监听端口，客户端连接用
	var customer []*core.Actor
	customer = append(customer, self.owner)
	listen_actor := network.NewTCPServer(customer, beego.AppConfig.DefaultString(core.Appname+"::listen", ""))
	child := core.NewActor(common.EActorType_SERVER.Int32(), listen_actor, make(chan core.IMessage, 10000))
	core.CoreRegisteActor(child)

}

func (self *PDK) Stop() {

}

func (self *PDK) HandleMessage(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_CREATE_ROOM.UInt16(): // 大厅请求创建房间
		self.Old_MSGID_CREATE_ROOM(actor, msg, session)
	case protomsg.Old_MSGID_RECV_ACCOUNT_INFO.UInt16(): // 同步账号数据
		self.Old_MSGID_RECV_ACCOUNT_INFO(actor, msg, session)
	case protomsg.Old_MSGID_ENTER_GAME.UInt16(): // 客户端链接进入游戏
		self.Old_MSGID_ENTER_GAME(actor, msg, session)
	case protomsg.Old_MSGID_MAINTENANCE_NOTICE.UInt16(): // 关服通知
		self.Old_MSGID_MAINTENANCE_NOTICE(actor, msg, session)
	case protomsg.Old_MSGID_BACKSTAGE_CLOSE_ROOM.UInt16(): // 后台关房间操作
		self.Old_MSGID_BACKSTAGE_CLOSE_ROOM(actor, msg, session)
	case protomsg.Old_MSGID_GET_ONE_BONUSPOOL.UInt16(): // 请求奖金池
		self.Old_MSGID_GET_ONE_BONUSPOOL(actor, msg, session)
	case protomsg.Old_MSGID_BACKSTAGE_SET_BONUSPOOL.UInt16(): // 后台设置奖金池
		self.Old_MSGID_BACKSTAGE_SET_BONUSPOOL(actor, msg, session)
	case protomsg.Old_MSGID_CHANGE_PLAYER_INFO.UInt16(): // 修改玩家信息
		self.Old_MSGID_CHANGE_PLAYER_INFO(actor, msg, session)
	case protomsg.Old_MSGID_CLIENT_KEEPALIVE.UInt16():
		send_tools.Send2Account(msg, session)
	case protomsg.Old_MSGID_SS_TEST_NETWORK.UInt16():
		log.Infof("收到测试网络消息 SessionID:%v", session)
	case protomsg.Old_MSGID_HG_TEST_NETWORK.UInt16():
		log.Infof("收到来自大厅的测试网络消息 SessionID:%v", session)
		req := packet.NewPacket(nil)
		req.SetMsgID(protomsg.Old_MSGID_HG_TEST_NETWORK.UInt16())
		send_tools.Send2Hall(req.GetData())
	default: // 客户端游戏消息，统一发送给房间处理
		acc := account.AccountMgr.GetAccountBySessionID(session)
		if acc == nil {
			//log.Warnf("找不到session 关联的玩家 session:%v msgId：%v", session, pack.GetMsgID())
			return false
		}
		if acc.RoomID == 0 {
			log.Warnf("玩家roomid ==0 accid %v session:%v msgId：%v", acc.AccountId, session, pack.GetMsgID())
			return false
		}
		core.CoreSend(self.owner.Id, int32(acc.RoomID), msg, session)
		break
	}
	return true
}
