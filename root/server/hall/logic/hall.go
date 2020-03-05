package logic

import (
	"github.com/astaxie/beego"
	"root/common"
	"root/core"
	"root/core/log"
	"root/core/network"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/hall/account"
	"root/server/hall/send_tools"
)

type (
	Hall struct {
		owner       *core.Actor
		init        bool // 重新建立连接是否需要拉取所有数据
		ListenActor *core.Actor
	}
)

func NewHall() *Hall {
	return &Hall{}
}

func (self *Hall) Init(actor *core.Actor) bool {
	// 先处理脚本
	if core.ScriptDir != "" {
		core.InitScript(core.ScriptDir)
	}
	self.owner = actor
	// 连接DB
	connectDB_actor := network.NewTCPClient(self.owner, func() string {
		return core.CoreAppConfString("connectDB")
	}, self.registerDB)
	child := core.NewActor(common.EActorType_CONNECT_DB.Int32(), connectDB_actor, make(chan core.IMessage, 10000))
	core.CoreRegisteActor(child)

	// 初始化定时器
	self.owner.AddTimer(utils.MILLISECONDS_OF_SECOND*20, -1, OnSpeakerUpdate)
	self.owner.AddTimer(utils.MILLISECONDS_OF_SECOND*30, -1, OnThirtySecondsUpdate)
	self.owner.AddTimer(utils.MILLISECONDS_OF_MINUTE, -1, OneMinuteUpdate)
	self.owner.AddTimer(utils.MILLISECONDS_OF_MINUTE*5, -1, FiveMinuteUpdate)
	self.owner.AddTimer(utils.MILLISECONDS_OF_HOUR, -1, OneHourUpdate)
	self.owner.AddTimer(utils.MILLISECONDS_OF_SECOND, -1, SecondUpdate)
	self.owner.AddEverydayTimer("23:59:59", ZeroUpdate)
	core.Cmd.Regist("help", CMD_Help, true)
	core.Cmd.Regist("reload", self.CMD_LoadConfig, true)
	core.Cmd.Regist("player", CMD_Player, true)
	core.Cmd.Regist("on", CMD_On, true)
	core.Cmd.Regist("off", CMD_Off, true)
	core.Cmd.Regist("print-sp", CMD_Print_Speaker, true)
	core.Cmd.Regist("add-sp", CMD_Add_Speaker, true)
	core.Cmd.Regist("del-sp", CMD_Del_Speaker, true)
	core.Cmd.Regist("print-email", CMD_Print_Email, true)
	core.Cmd.Regist("del-email", CMD_Del_Email, true)
	core.Cmd.Regist("money", CMD_Add_Money, true)
	core.Cmd.Regist("kill", CMD_Kill, true)
	core.Cmd.Regist("todb", CMD_ToDB, true)
	core.Cmd.Regist("save", CMD_Save, true)
	core.Cmd.Regist("room", self.CMD_RoomInfo, true)
	core.Cmd.Regist("stop", self.CMD_Stop, true)
	core.Cmd.Regist("close", self.CMD_Close, true)

	return true
}

func (self *Hall) registerDB() {
	// 向db建立连接
	send_tools.Send2DB(inner.SERVERMSG_HD_HELLO_DB.UInt16(), nil)
	if !self.init {
		// 像DB请求所有玩家数据
		send_tools.Send2DB(inner.SERVERMSG_HD_ALL_DATA.UInt16(), nil)
		self.init = true
	}
}

func (self *Hall) SERVERMSG_DH_FINISH_DATA(session int64) {
	if session != 0 {
		log.Infof("Error: 不是来自于DB服务器的消息, FinishRecvAllData, SessionID:%v", session)
		return
	}
	// 所有玩家和机器人都初始化完成以后, 再将玩家和机器人的ID排除掉
	account.AccountMgr.CollatingIDAssign()
	// 大厅完成所有数据初始化, 开启监听，让客户端可连接
	var customer []*core.Actor
	customer = append(customer, self.owner)
	listen_actor := network.NewTCPServer(customer, beego.AppConfig.DefaultString(core.Appname+"::listen", ""), beego.AppConfig.DefaultString(core.Appname+"::listenHttp", ""))
	self.ListenActor = core.NewActor(common.EActorType_SERVER.Int32(), listen_actor, make(chan core.IMessage, 10000))
	core.CoreRegisteActor(self.ListenActor)
	strServerIP := utils.GetLocalIP()
	GameMgr.PrintSign(strServerIP)
}

func (self *Hall) SERVERMSG_HD_SAVE_ALL() {
	log.Info("数据回存完毕!!! 按下先关闭所有游戏服务器")
	log.Info("数据回存完毕!!! 按下再关闭HallServer服务器")
}
func (self *Hall) Stop() {

}

func (self *Hall) HandleMessage(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	//if name,e := protomsg.MSG_name[int32(pack.GetMsgID())];e{
	//	log.Infof("收到消息:%v %v ", pack.GetMsgID(),name)
	//}else{
	//	log.Infof("收到消息:%v %v ", pack.GetMsgID(),inner.SERVERMSG_name[int32(pack.GetMsgID())])
	//}
	switch pack.GetMsgID() {
	case utils.ID_DISCONNECT: // 客户端或游戏进程断开连接
		self.MSGID_CLOSE_CONNECT(actor, msg, session)
	case protomsg.MSG_CLIENT_KEEPALIVE.UInt16(): // 心跳
		send_tools.Send2Account(protomsg.MSG_CLIENT_KEEPALIVE.UInt16(), nil, session)
	case inner.SERVERMSG_GH_GAME_CONNECT_HALL.UInt16(): // game向hall请求连接
		self.SERVERMSG_GH_GAME_CONNECT_HALL(actor, pack.ReadBytes(), session)
	case inner.SERVERMSG_GH_CLOSE_SERVER_FIN.UInt16():
		for _, node := range GameMgr.nodes {
			if node.session == session {
				log.Infof("游戏:%v 完成关闭！", common.EGameType(node.gameType))
				break
			}
		}
	case inner.SERVERMSG_GH_ROOM_INFO.UInt16(): // game向hall 发送房间信息
		self.SERVERMSG_GH_ROOM_INFO(actor, pack.ReadBytes(), session)

	//---------------------------- 数据库加载和回存 ---------------------------------------------
	case inner.SERVERMSG_DH_ALL_ACCOUNT_RESP.UInt16(): // 加载所有账号
		self.SERVERMSG_DH_ALL_ACCOUNT_RESP(actor, pack.ReadBytes(), session)
	case inner.SERVERMSG_DH_ALL_EMAIL_RESP.UInt16(): // 加载账号邮件
		self.SERVERMSG_DH_ALL_EMAIL_RESP(actor, pack.ReadBytes(), session)
	case inner.SERVERMSG_DH_ALL_WATER_LINE.UInt16(): // 水位线
		self.SERVERMSG_DH_ALL_WATER_LINE(actor, pack.ReadBytes(), session)
	case inner.SERVERMSG_DH_ALL_ROOM_BONUS.UInt16(): // 水池
		self.SERVERMSG_DH_ALL_ROOM_BONUS(actor, pack.ReadBytes(), session)
	case inner.SERVERMSG_DH_FINISH_DATA.UInt16(): // 所有数据初始化完成
		self.SERVERMSG_DH_FINISH_DATA(session)
	case inner.SERVERMSG_HD_SAVE_ALL.UInt16(): // 所有数据完成回存
		self.SERVERMSG_HD_SAVE_ALL()

	//---------------------------- 大厅相关 -----------------------------------------------------
	case protomsg.MSG_CS_LOGIN_HALL_REQ.UInt16(): // 登录请求
		self.MSG_LOGIN_HALL(actor, pack.ReadBytes(), session)
	case protomsg.MSG_CS_SYNC_SERVER_TIME.UInt16(): // 客户端同步服务器时间
		self.MSG_CS_SYNC_SERVER_TIME(actor, pack.ReadBytes(), session)
	case protomsg.MSG_CS_BIND_PHONE_REQ.UInt16(): // 客户端绑定帐号
		self.MSG_CS_BIND_PHONE_REQ(actor, pack.ReadBytes(), session)
	case protomsg.MSG_CS_SAFEMONEY_OPERATE_REQ.UInt16(): // 客户端操作保险箱
		self.MSG_CS_SAFEMONEY_OPERATE_REQ(actor, pack.ReadBytes(), session)
	case protomsg.MSG_CS_ENTER_ROOM_REQ.UInt16(): // 玩家请求进入房间
		self.MSG_CS_ENTER_ROOM_REQ(actor, pack.ReadBytes(), session)
	case inner.SERVERMSG_GH_PLAYER_DATA_RES.UInt16(): // 游戏回复大厅，收到玩家数据，通知玩家进入游戏
		self.SERVERMSG_GH_PLAYER_DATA_RES(actor, pack.ReadBytes(), session)
	case inner.SERVERMSG_GH_PLAYER_ENTER_ROOM.UInt16(): // 游戏通知大厅，玩家进入房间
		self.SERVERMSG_GH_PLAYER_ENTER_ROOM(actor, pack.ReadBytes(), session)
	case inner.SERVERMSG_GH_PLAYER_LEAVE_ROOM.UInt16(): // 游戏通知大厅，玩家进入房间
		self.SERVERMSG_GH_PLAYER_LEAVE_ROOM(actor, pack.ReadBytes(), session)

		//---------------------------- 邮件 ---------------------------------------------
	case protomsg.MSG_CS_EMAILS_REQ.UInt16(): // Client请求邮件列表
		self.MSG_CS_EMAILS_REQ(actor, pack.ReadBytes(), session)
	case protomsg.MSG_CS_EMAIL_READ_REQ.UInt16(): // Client请求阅读一封未读邮件
		self.MSG_CS_EMAIL_READ_REQ(actor, pack.ReadBytes(), session)
	case protomsg.MSG_CS_EMAIL_REWARD_REQ.UInt16(): // Client请求领取邮件内奖励
		self.MSG_CS_EMAIL_REWARD_REQ(actor, pack.ReadBytes(), session)
	case protomsg.MSG_CS_EMAIL_DEL_REQ.UInt16(): // Client请求删除邮件
		self.MSG_CS_EMAIL_DEL_REQ(actor, pack.ReadBytes(), session)

	//---------------------------- 游戏相关 ---------------------------------------------
	case inner.SERVERMSG_GH_SERVERFEE_LOG.UInt16(): // 服务费日志
		self.SERVERMSG_GH_SERVERFEE_LOG(actor, pack.ReadBytes(), session)
	case inner.SERVERMSG_GH_MONEYCHANGE.UInt16(): // 金币改变日志
		self.SERVERMSG_GH_MONEYCHANGE(actor, pack.ReadBytes(), session)
	case inner.SERVERMSG_GH_ROOM_BONUS_REQ.UInt16(): // 游戏请求水池金额
		self.SERVERMSG_GH_ROOM_BONUS_REQ(actor, pack.ReadBytes(), session)
	case inner.SERVERMSG_GH_ROOM_BONUS_SAVE.UInt16(): // 游戏请求回存水池金额
		self.SERVERMSG_GH_ROOM_BONUS_SAVE(actor, pack.ReadBytes(), session)
	case inner.SERVERMSG_GH_ROOM_PROFIT_SAVE.UInt16(): // 游戏请求回存盈利金额
		self.SERVERMSG_GH_ROOM_PROFIT_SAVE(actor, pack.ReadBytes(), session)

	case inner.SERVERMSG_SS_TEST_NETWORK.UInt16():
		log.Infof("收到测试网络消息 SessionID:%v", session)
		send_tools.Send2Game(inner.SERVERMSG_SS_TEST_NETWORK.UInt16(), nil, session)
	default:
		tAccount := account.AccountMgr.GetAccountBySessionID(session)
		if tAccount != nil {
			log.Infof("Error: HandleMessage don`t find handler, msgid:%v AccountID:%v Name:%v actor:%v session:%v", pack.GetMsgID(), tAccount.AccountId, tAccount.Name, actor, session)
		} else {
			log.Infof("Error: HandleMessage don`t find handler, msgid:%v actor:%v session:%v", pack.GetMsgID(), actor, session)
		}
		break
	}
	return true
}
