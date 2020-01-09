package logic

import (
	"github.com/astaxie/beego"
	"root/common"
	"root/common/config"
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
		owner          *core.Actor
		init           bool // 重新建立连接是否需要拉取所有数据
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

	self.SetNode()

	// 初始化定时器
	self.owner.AddTimer(utils.MILLISECONDS_OF_SECOND*10, -1, OnSaveAccount)
	self.owner.AddTimer(utils.MILLISECONDS_OF_SECOND*20, -1, OnSpeakerUpdate)
	self.owner.AddTimer(utils.MILLISECONDS_OF_SECOND*30, -1, OnThirtySecondsUpdate)
	self.owner.AddTimer(utils.MILLISECONDS_OF_MINUTE, -1, OneMinuteUpdate)
	self.owner.AddTimer(utils.MILLISECONDS_OF_MINUTE*5, -1, FiveMinuteUpdate)
	self.owner.AddTimer(utils.MILLISECONDS_OF_HOUR, -1, OneHourUpdate)
	self.owner.AddTimer(utils.MILLISECONDS_OF_SECOND, -1, SecondUpdate)
	self.owner.AddEverydayTimer("23:59:59", ZeroUpdate)
	core.Cmd.Regist("help", CMD_Help, true)
	core.Cmd.Regist("reload", self.CMD_LoadConfig, true)
	core.Cmd.Regist("stopall", CMD_StopAll, true)
	core.Cmd.Regist("kickall", CMD_KickAll, true)
	core.Cmd.Regist("saveall", CMD_SaveAll, true)
	core.Cmd.Regist("force-saveall", CMD_ForceSaveAll, true)
	core.Cmd.Regist("count", CMD_Count, true)
	core.Cmd.Regist("player", CMD_Player, true)
	core.Cmd.Regist("to", CMD_To, true)
	core.Cmd.Regist("on", CMD_On, true)
	core.Cmd.Regist("off", CMD_Off, true)
	core.Cmd.Regist("room", CMD_Room, true)
	core.Cmd.Regist("print-sp", CMD_Print_Speaker, true)
	core.Cmd.Regist("add-sp", CMD_Add_Speaker, true)
	core.Cmd.Regist("del-sp", CMD_Del_Speaker, true)
	core.Cmd.Regist("print-email", CMD_Print_Email, true)
	core.Cmd.Regist("del-email", CMD_Del_Email, true)
	core.Cmd.Regist("up-sf", CMD_SetUp_SalesmenType, true)
	core.Cmd.Regist("down-sf", CMD_SetDown_SalesmenType, true)
	core.Cmd.Regist("set-ch", CMD_Set_ChannelID, true)
	core.Cmd.Regist("check", CMD_Check, true)
	core.Cmd.Regist("add-rmb", CMD_Add_RMB, true)
	core.Cmd.Regist("get-robot-id", CMD_Get_Robot_ID, true)
	core.Cmd.Regist("todb", CMD_ToDB, true)
	core.Cmd.Regist("test", CMD_Test, true)
	core.Cmd.Regist("node", self.CMD_Node, true)
	core.Cmd.Regist("open", self.Open, true)
	core.Cmd.Regist("weihu", self.WeiHuGame, true)
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
	listen_actor := network.NewTCPServer(customer, beego.AppConfig.DefaultString(core.Appname+"::listen", ""))
	HallMgr.ListenActor = core.NewActor(common.EActorType_SERVER.Int32(), listen_actor, make(chan core.IMessage, 10000))
	core.CoreRegisteActor(HallMgr.ListenActor)

	go HallMgr.runExchangeRequest()

	strServerIP := utils.GetLocalIP()
	HallMgr.PrintSign(strServerIP)
}

func (self *Hall) SERVERMSG_HD_SAVE_ALL() {
	log.Info("数据回存完毕!!! 按下先关闭所有游戏服务器")
	log.Info("数据回存完毕!!! 按下再关闭HallServer服务器")
}
func (self *Hall) Stop() {

}

func (self *Hall) HandleMessage(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case utils.ID_DISCONNECT: // 客户端或游戏进程断开连接
		self.MSGID_CLOSE_CONNECT(actor, msg, session)
	case protomsg.MSG_CLIENT_KEEPALIVE.UInt16():
		send_tools.Send2Account(protomsg.MSG_CLIENT_KEEPALIVE.UInt16(),nil, session)
	//case protomsg.Old_MSGID_SS_MAPING.UInt16(): // game向hall请求连接
		//self.Old_MSGID_SS_MAPING(actor, msg, session)

	//---------------------------- 数据库加载和回存 ---------------------------------------------
	case inner.SERVERMSG_DH_ALL_ACCOUNT_RESP.UInt16(): // 加载所有账号
		self.SERVERMSG_DH_ALL_ACCOUNT_RESP(actor, pack.ReadBytes(), session)
	case inner.SERVERMSG_DH_ALL_EMAIL_RESP.UInt16(): // 加载账号邮件
		self.SERVERMSG_DH_ALL_EMAIL_RESP(actor, pack.ReadBytes(), session)
	case inner.SERVERMSG_DH_ALL_WATER_LINE.UInt16(): // 水位线
		self.SERVERMSG_DH_ALL_WATER_LINE(actor, pack.ReadBytes(), session)
	case inner.SERVERMSG_DH_FINISH_DATA.UInt16(): // 所有数据初始化完成
		self.SERVERMSG_DH_FINISH_DATA(session)
	case inner.SERVERMSG_HD_SAVE_ALL.UInt16(): // 所有数据完成回存
		self.SERVERMSG_HD_SAVE_ALL()

	//---------------------------- 帐号相关 ---------------------------------------------
	case protomsg.MSG_CS_LOGIN_HALL_REQ.UInt16(): // 登录请求
		self.MSG_LOGIN_HALL(actor, msg, session)

	//---------------------------- 大厅相关 ---------------------------------------------
	case protomsg.MSG_CS_SYNC_SERVER_TIME.UInt16(): // 客户端同步服务器时间
		self.MSG_CS_SYNC_SERVER_TIME(actor, msg, session)
	case protomsg.MSG_CS_BIND_PHONE.UInt16(): // 客户端绑定帐号
		self.MSG_CS_BIND_PHONE(actor, pack.ReadBytes(), session)
	case protomsg.Old_MSGID_CREATE_ROOM.UInt16(): // 客户端申请创建房间
		self.Old_MSGID_CREATE_ROOM(actor, msg, session)
	case protomsg.Old_MSGID_CREATE_ROOM_RET.UInt16(): // 客户端申请创建房间结果
		self.Old_MSGID_CREATE_ROOM_RESULT(actor, msg, session)
	case protomsg.Old_MSGID_ENTER_ROOM.UInt16(): // 客户端进入房间
		self.Old_MSGID_ENTER_ROOM(actor, msg, session)
	case protomsg.Old_MSGID_RECV_ACCOUNT_INFO.UInt16(): // 游戏通知接收帐号数据完成
		self.Old_MSGID_RECV_ACCOUNT_INFO(actor, msg, session)
	case protomsg.Old_MSGID_SYNC_TO_HALL_MONEY.UInt16(): // 游戏同步元宝到大厅
		self.Old_MSGID_SYNC_TO_HALL_MONEY(actor, msg, session)
	case protomsg.MSGID_SAVE_RMB_CHANGE_LOG.UInt16(): // 元宝变动日志缓存到大厅, 统一处理
		self.MSGID_SAVE_RMB_CHANGE_LOG(actor, msg, session)
	case protomsg.Old_MSGID_GET_ROOM_LIST.UInt16(): // 客户端请求房间列表
		self.Old_MSGID_GET_ROOM_LIST(actor, msg, session)
	case protomsg.Old_MSGID_OPEN_DESK_UPDATE.UInt16(): // 客户端开启桌子更新
		self.Old_MSGID_OPEN_DESK_UPDATE(actor, msg, session)
	case protomsg.Old_MSGID_CLOSE_DESK_UPDATE.UInt16(): // 客户端关闭桌子更新
		self.Old_MSGID_CLOSE_DESK_UPDATE(actor, msg, session)
	case protomsg.Old_MSGID_UPDATE_SERVICE_FEE.UInt16(): // 游戏通知大厅更新玩家服务费
		self.Old_MSGID_UPDATE_SERVICE_FEE(actor, msg, session)
	case protomsg.Old_MSGID_UPDATE_ACCOUNT.UInt16(): // 游戏通知大厅更新玩家数据
		self.Old_MSGID_UPDATE_ACCOUNT(actor, msg, session)
	case protomsg.Old_MSGID_UPDATE_INDEX.UInt16(): // 游戏通知玩家更新座位索引
		self.Old_MSGID_UPDATE_INDEX(actor, msg, session)
	case protomsg.Old_MSGID_UPDATE_ENTER.UInt16(): // 游戏通知玩家进入房间
		self.Old_MSGID_UPDATE_ENTER(actor, msg, session)
	case protomsg.Old_MSGID_UPDATE_LEAVE.UInt16(): // 游戏通知玩家离开房间
		self.Old_MSGID_UPDATE_LEAVE(actor, msg, session)
	case protomsg.Old_MSGID_UPDATE_DESTROY_ROOM.UInt16(): // 游戏通知大厅销毁房间
		self.Old_MSGID_UPDATE_DESTROY_ROOM(actor, msg, session)
	case protomsg.Old_MSGID_SEND_RANK_LIST.UInt16(): // 客户端请求排行榜数据
		self.Old_MSGID_SEND_RANK_LIST(actor, msg, session)
	case protomsg.MSGID_OPERATE_SAFE_BOX.UInt16(): // 客户端操作保险箱
		self.MSGID_OPERATE_SAFE_BOX(actor, msg, session)
	case protomsg.MSGID_OPERATE_SAFE_BOX_RESPOND.UInt16(): // 游戏返回操作保险箱
		self.MSGID_OPERATE_SAFE_BOX(actor, msg, session)
	case protomsg.Old_MSGID_MAINTENANCE_NOTICE.UInt16(): // 游戏可关闭通知大厅
		self.Old_MSGID_MAINTENANCE_NOTICE(actor, msg, session)
	case protomsg.MSGID_CH_SELEChT_MATCH_RESULT.UInt16(): // 选择匹配结果
		self.MSGID_CH_SELECT_MATCH_RESULT(actor, msg, session)
	case protomsg.MSGID_HG_REENTER_OTHER_GAME.UInt16():
		self.MSGID_HG_REENTER_OTHER_GAME(actor, msg, session)

	//---------------------------- 房间相关 ---------------------------------------------
	//case protomsg.MSGID_GET_ONE_WATERLINE.UInt16(): // 请求水位线数据
		self.MSGID_GET_ONE_WATER_LINE(actor, msg, session)
	//case protomsg.MSGID_SET_ONE_WATERLINE.UInt16(): // 设置水位线数据
		self.MSGID_SET_ONE_WATER_LINE(actor, msg, session)
	//case protomsg.MSGID_SAVE_LOG.UInt16(): // 回存游戏产生的日志
		self.MSGID_SAVE_LOG(actor, msg, session)


		//---------------------------- 邮件 ---------------------------------------------
	case protomsg.Old_MSGID_CHANGE_EMAIL_TO_READED.UInt16(): // Client请求改变邮件状态为已读
		self.Old_MSGID_CHANGE_EMAIL_TO_READED(actor, msg, session)
	case protomsg.Old_MSGID_GET_EMAIL_REWARD.UInt16(): // Client请求领取邮件内奖励
		self.Old_MSGID_GET_EMAIL_REWARD(actor, msg, session)
	case protomsg.Old_MSGID_SEND_EMALL_LIST.UInt16(): // Client请求邮件列表
		self.Old_MSGID_SEND_EMALL_LIST(actor, msg, session)
	case protomsg.Old_MSGID_DELETE_EMAIL.UInt16(): // Client请求删除邮件
		self.Old_MSGID_DELETE_EMAIL(actor, msg, session)

	case inner.SERVERMSG_SS_TEST_NETWORK.UInt16():
		log.Infof("收到测试网络消息 SessionID:%v", session)
		req := packet.NewPacket(nil)
		req.SetMsgID(inner.SERVERMSG_SS_TEST_NETWORK.UInt16())
		send_tools.Send2Game(req.GetData(), session)
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

func (self *Hall) cachePacket(pack packet.IPacket) {
}

func (self *Hall) SetNode() {
	_, strlocalIP, _ := config.IsTestServer()
	DD_HALL_IP := config.GetPublicConfig_String("DD_HALL_IP")
	if strlocalIP == DD_HALL_IP {
		conf := config.GetPublicConfig_String("DD_ROOM_NODE_LIST")
		HallMgr.ipNodes = utils.SplitConf2Mapis(conf)
	} else {
		conf := config.GetPublicConfig_String("HH_ROOM_NODE_LIST")
		HallMgr.ipNodes = utils.SplitConf2Mapis(conf)
	}

	log.Infof("setnode :%v", HallMgr.ipNodes)
}

func (self *Hall) AlterWebData() {
	if self.webTimer == 0 {
		self.webTimer = self.owner.AddTimer(int64(1000*60*5), 1, func(dt int64) {
			self.webTimer = 0
			HallMgr.SaveWebData()
		})
	}
}
