package logic

import (
	"root/common"
	"root/common/config"
	"root/common/model/rediskey"
	"root/common/tools"
	"root/core"
	"root/core/db"
	"root/core/log"
	"root/core/network"
	"root/core/packet"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/yuin/gopher-lua"
	"strconv"

	"root/protomsg"
)

type (
	Process struct {
		owner  *core.Actor
		listen *core.Actor
	}
)

func NewProcess() *Process {
	return &Process{}
}

func (self *Process) Init(actor *core.Actor) bool {
	// 先处理脚本
	if core.ScriptDir != "" {
		core.InitScript(core.ScriptDir)
	}

	self.owner = actor

	// 连接hall
	connectHall_actor := network.NewTCPClient(self.owner, func() string {
		return core.CoreAppConfString("connectHall")
	}, self.registerHall)
	child := core.NewActor(common.EActorType_CONNECT_HALL.Int32(), connectHall_actor, make(chan core.IMessage, 1000))
	core.CoreRegisteActor(child)

	core.Cmd.Regist("tohall", CMD_ToHall, true)
	return true
}

func (self *Process) open_listen() {
	// 监听端口，客户端连接用
	var customer []*core.Actor
	customer = append(customer, self.owner)
	listen_actor := network.NewTCPServer(customer, beego.AppConfig.DefaultString(core.Appname+"::listen", ""))
	self.listen = core.NewActor(common.EActorType_SERVER.Int32(), listen_actor, make(chan core.IMessage, 10000))
	core.CoreRegisteActor(self.listen)

}

func (self *Process) stop() {
	self.listen.Suspend()

}

// 向hall注册
func (self *Process) registerHall() {
	// 发送注册消息登记自身信息
	sid, _ := strconv.Atoi(core.Appname)

	if err := core.Global_Lua.CallByParam(lua.P{
		Fn:      core.Global_Lua.GetGlobal("lua_GetRoomCount"),
		NRet:    1,
		Protect: true,
	}); err != nil {
		log.Errorf("%v", err.Error())
		return
	}
	Lval := core.Global_Lua.Get(-1)
	count := Lval.(lua.LNumber)
	core.Global_Lua.Pop(1)

	GAME_TO_HALL_MAP_KEY := config.GetPublicConfig_String("GAME_TO_HALL_MAP_KEY")
	strSign := fmt.Sprintf("%v%v%v", sid, count, GAME_TO_HALL_MAP_KEY)
	strSign = tools.MD5(strSign)

	// 组装消息
	Databytes := packet.NewPacket(nil)
	Databytes.WriteUInt16(uint16(sid))
	Databytes.WriteUInt32(uint32(count))
	Databytes.WriteString(strSign)
	Databytes.SetMsgID(protomsg.Old_MSGID_SS_MAPING.UInt16())
	core.CoreSend(self.owner.Id, common.EActorType_CONNECT_HALL.Int32(), Databytes.GetData(), 0)
	log.Infof("连接hall成功 房间数量:%v", count)

	// 通知Lua连接大厅成功, 可以开始初始化工作
	tSendToLua := packet.NewPacket(nil)
	tSendToLua.SetMsgID(protomsg.Old_MSGID_SS_HALL_CONNECT_SUCCESS.UInt16())
	core.CoreSend(self.owner.Id, common.EActorType_MAIN.Int32(), tSendToLua.GetData(), 0)
}

func (self *Process) Stop() {

}

func (self *Process) HandleMessage(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_SS_LUA_SAVE_REDIS.UInt16():
		accid := pack.ReadUInt32()
		field := pack.ReadString()
		val := pack.ReadUInt64()
		db.HSet(rediskey.PlayerId(uint32(accid)), field, val)

	case protomsg.Old_MSGID_CLIENT_KEEPALIVE.UInt16():
		core.CoreSend(0, common.EActorType_SERVER.Int32(), msg, session)
	case protomsg.Old_MSGID_SS_START_LISTEN.UInt16():
		self.open_listen()
	case protomsg.Old_MSGID_SS_STOP_SERVER.UInt16():
		self.stop()
	case protomsg.Old_MSGID_SS_TEST_NETWORK.UInt16():
		log.Infof("收到测试网络消息 SessionID:%v", session)
	case protomsg.Old_MSGID_HG_TEST_NETWORK.UInt16():
		log.Infof("收到来自大厅的测试网络消息 SessionID:%v", session)
		req := packet.NewPacket(nil)
		req.SetMsgID(protomsg.Old_MSGID_HG_TEST_NETWORK.UInt16())
		core.CoreSend(0, common.EActorType_CONNECT_HALL.Int32(), req.GetData(), 0)
	default:
		if !core.MsgProcess(pack.GetMsgID(), msg, session) {
			log.Warn("HandleMessage don`t find handler, msgid =", pack.GetMsgID(), " actor =", actor, " session =", session)
		}
		break
	}
	return true
}
