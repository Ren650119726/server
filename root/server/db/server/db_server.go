package server

import (
	"github.com/astaxie/beego"
	"root/common"
	"root/common/model/inst"
	"root/common/model/logdb"
	"root/common/model/web"
	"root/core"
	"root/core/db"
	"root/core/log"
	"root/core/network"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg/inner"
	"root/server/db/send_tools"
	"root/server/db/types"
)

/*
 * log
 */
type (
	DBServer struct {
		owner    *core.Actor
		serverid int32
	}
)

// 创建一个DCServer
func NewDBServer() *DBServer {
	dc := &DBServer{}
	return dc
}

//func ZeroUpdate(dt int64) {
//	dailyRegistTable()
//	log.Info("每日注册日志表")
//}

// actor初始化(actor接口定义)
func (self *DBServer) Init(actor *core.Actor) bool {
	// 获取serverid
	self.owner = actor
	self.serverid = int32(beego.AppConfig.DefaultInt(core.Appname+"::sid", 0))
	//self.owner.AddEverydayTimer("0:00:00", ZeroUpdate)

	// 注册mysql表
	nStartTime := utils.MilliSecondTimeSince1970()
	registTable()
	dailyRegistTable()
	nEndTime := utils.MilliSecondTimeSince1970()
	log.Infof("注册数据库表花费:%v毫秒", nEndTime-nStartTime)

	//// 获取redis数据的actor
	//redisActor := NewRedisHandler()
	//child := core.NewActor(types.EActorType_REDIS.Int32(), redisActor, make(chan core.IMessage, 100))
	//core.CoreRegisteActor(child)

	// 获取mysql数据的actor
	mysqlActor := NewMysqlHandler()
	child := core.NewActor(types.EActorType_MYSQL.Int32(), mysqlActor, make(chan core.IMessage, 5000))
	core.CoreRegisteActor(child)

	//// 获取log数据的actor
	//logActor := NewLogHandler()
	//child = core.NewActor(types.EActorType_LOG.Int32(), logActor, make(chan core.IMessage, 1000))
	//core.CoreRegisteActor(child)

	// 监听端口
	var customer []*core.Actor
	customer = append(customer, self.owner)
	listen_actor := network.NewNetworkServer(customer, beego.AppConfig.DefaultString(core.Appname+"::listen", ""),
		beego.AppConfig.DefaultString(core.Appname+"::listenHttp", ""))
	child = core.NewActor(common.EActorType_SERVER.Int32(), listen_actor, make(chan core.IMessage, 100000))
	core.CoreRegisteActor(child)

	go GRPC_SERVER()
	// 读取所有名字
	return true
}


func registTable() {
	// 角色表
	db.RegisteModel(&inst.AccountModel{})
	db.RegisteModel(&inst.EmailModel{})
	db.RegisteModel(&inst.WaterLineModel{})
	db.RegisteModel(&inst.RoomBonusModel{})

	// 日志表
	db.RegisteModel_log(&logdb.MoneyDailyModel{})
	db.RegisteModel_log(&logdb.BonusPoolModel{})
	db.RegisteModel_log(&logdb.RechargeModel{})

	// web表
	db.RegisteModel_web(&web.WebConfigModel{})
	db.RegisteModel_web(&web.Exchange_configModel{})
	db.RegisteModel_web(&web.PayChannelModel{})
}

func dailyRegistTable() {
	rmb := &logdb.MoneyModel{}
	db.RegisteModel_log(rmb.Portion(0))
	db.RegisteModel_log(rmb.Portion(1))
	db.RegisteModel_log(rmb.Portion(2))
	db.RegisteModel_log(rmb.Portion(3))
	db.RegisteModel_log(rmb.Portion(4))
	db.RegisteModel_log(rmb.Portion(5))
	db.RegisteModel_log(rmb.Portion(6))
	db.RegisteModel_log(rmb.Portion(7))
	db.RegisteModel_log(rmb.Portion(8))
	db.RegisteModel_log(rmb.Portion(9))

	service := &logdb.ServiceModel{}
	db.RegisteModel_log(service.Portion(0))
	db.RegisteModel_log(service.Portion(1))
	db.RegisteModel_log(service.Portion(2))
	db.RegisteModel_log(service.Portion(3))
	db.RegisteModel_log(service.Portion(4))
	db.RegisteModel_log(service.Portion(5))
	db.RegisteModel_log(service.Portion(6))
	db.RegisteModel_log(service.Portion(7))
	db.RegisteModel_log(service.Portion(8))
	db.RegisteModel_log(service.Portion(9))
}

// 停止回收相关资源
func (self *DBServer) Stop() {}

// actor消息处理
func (self *DBServer) HandleMessage(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case inner.SERVERMSG_HD_HELLO_DB.UInt16(): // 大厅请求建立连接
		send_tools.Hall_session = session
		log.Infof("大厅连接成功")
	case utils.ID_DISCONNECT:
		if session == send_tools.Hall_session {
			log.Infof("大厅断开链接")
			send_tools.Hall_session = 0
		}
	case inner.SERVERMSG_SS_TEST_NETWORK.UInt16():
		log.Infof("收到测试网络消息 SessionID:%v", session)
		req := packet.NewPacket(nil)
		req.SetMsgID(inner.SERVERMSG_SS_TEST_NETWORK.UInt16())
		core.CoreSend(0, common.EActorType_SERVER.Int32(), req.GetData(), session)
		core.CoreSend(self.owner.Id, types.EActorType_MYSQL.Int32(), msg, session)
	default:
		core.CoreSend(self.owner.Id, types.EActorType_MYSQL.Int32(), msg, session)
	}
	return true
}
