package server

import (
	"github.com/golang/protobuf/proto"
	"github.com/jinzhu/gorm"
	"root/common"
	"root/common/model/inst"
	"root/common/tools"
	"root/core"
	"root/core/db"
	"root/core/log"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/db/send_tools"
)

type mysql_server struct {
	owner *core.Actor
}

func NewMysqlHandler() *mysql_server {
	dc := &mysql_server{}
	return dc
}

// actor初始化(actor接口定义)
func (self *mysql_server) Init(actor *core.Actor) bool {
	self.owner = actor

	return true
}

// 停止回收相关资源
func (self *mysql_server) Stop() {

}

// actor消息处理
func (self *mysql_server) HandleMessage(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)

	switch pack.GetMsgID() {
	case inner.SERVERMSG_HD_SQL_SYNTAX.UInt16(): // 执行sql syntax
		self.SERVERMSG_HD_SQL_SYNTAX(pack.ReadBytes(), session)
	case inner.SERVERMSG_HD_ALL_DATA.UInt16(): // 大厅请求所有数据
		self.SERVERMSG_HD_ALL_DATA(pack.ReadBytes(), session)
	case inner.SERVERMSG_HD_SAVE_ACCOUNT.UInt16(): // 回存玩家数据
		if !self.SERVERMSG_HD_SAVE_ACCOUNT(pack.ReadBytes(), session) {
			core.CoreSend(self.owner.Id, self.owner.Id, pack.GetData(), session)
		}
	case inner.SERVERMSG_HD_SAVE_EMAIL_PERSON.UInt16(): // 大厅回存所有邮件数据
		self.SERVERMSG_HD_SAVE_EMAIL_PERSON(pack, session)
	case inner.SERVERMSG_HD_SAVE_WATER_LINE.UInt16(): // 回存水位线
		self.SERVERMSG_HD_SAVE_WATER_LINE(pack, session)
	case inner.SERVERMSG_HD_SAVE_ROOM_BONUS.UInt16(): // 回存水池
		self.SERVERMSG_HD_SAVE_ROOM_BONUS(pack, session)
	case inner.SERVERMSG_HD_SAVE_ALL.UInt16(): // 通知大厅完成所有数据回存
		if dataInst := db.GetInst().Where("select 1"); dataInst.Error == nil {
			if dataLog := db.GetLog().Where("select 1"); dataLog.Error == nil {
				nRetMsgID := pack.GetMsgID()
				send_tools.Send2Hall(nRetMsgID, nil)
			} else {
				log.Warnf("infomation log:%v ", dataLog.Error)
				core.CoreSend(self.owner.Id, self.owner.Id, msg, session)
			}
		} else {
			log.Warnf("infomation inst:%v ", dataInst.Error)
			core.CoreSend(self.owner.Id, self.owner.Id, msg, session)
		}
	case inner.SERVERMSG_SS_TEST_NETWORK.UInt16():
		log.Infof("mysql actor收到测试网络消息 SessionID:%v", session)
		req := packet.NewPacket(nil)
		req.SetMsgID(inner.SERVERMSG_SS_TEST_NETWORK.UInt16())
		core.CoreSend(0, common.EActorType_SERVER.Int32(), req.GetData(), session)
	default:
		log.Errorf("no handler msgid:%v", pack.GetMsgID())
	}
	return true
}

// 执行sql syntax
func (self *mysql_server) SERVERMSG_HD_SQL_SYNTAX(pbmsg []byte, session int64) {
	data := packet.PBUnmarshal(pbmsg, &inner.SQL_SYNTAX{}).(*inner.SQL_SYNTAX)
	sql_sytnax := data.GetSQLSyntax()
	db_type := data.GetDataBaseType()

	var mysql *gorm.DB
	if db_type == 0 {
		mysql = db.GetInst()
	} else if db_type == 1 {
		mysql = db.GetLog()
	} else {
		log.Errorf("位置的类型:%v", db_type)
		return
	}
	if mysql == nil {
		log.Warnf("获取数据库失败")
		return
	}
	log.Infof("SERVERMSG_HD_SQL_SYNTAX:%v", sql_sytnax)
	time_before := utils.MilliSecondTimeSince1970()
	err := mysql.Exec(sql_sytnax).Error
	time_after := utils.MilliSecondTimeSince1970()

	if time_diff := time_after - time_before; time_diff > 100 {
		log.Warnf("回存执行时间:%v sytnax:%v", time_diff, sql_sytnax)
	}
	if err != nil {
		log.Errorf("执行LUA中的SQL错误:%v 错误:%v", sql_sytnax, err)
	}
}

const maxsend = 200

// 大厅请求所有数据
func (self *mysql_server) SERVERMSG_HD_ALL_DATA(pbmsg []byte, session int64) {
	// 所有账号-----------------------------------------------------------------
	models := inst.GetAllAccount()
	sendAccounts := []*protomsg.AccountStorageData{}
	count := maxsend
	for _, accModel := range models {
		pbacc := &protomsg.AccountStorageData{}
		tools.CopyProtoData(accModel, pbacc) // 将grom model数据转换成proto数据
		sendAccounts = append(sendAccounts, pbacc)
		count--
		if count <= 0 {
			pack := &inner.ALL_ACCOUNT_RESP{
				AllAccount: sendAccounts,
			}
			send_tools.Send2Hall(inner.SERVERMSG_DH_ALL_ACCOUNT_RESP.UInt16(), pack)
			sendAccounts = []*protomsg.AccountStorageData{}
			count = maxsend
		}
	}
	sendAccount := &inner.ALL_ACCOUNT_RESP{
		AllAccount: sendAccounts,
	}
	send_tools.Send2Hall(inner.SERVERMSG_DH_ALL_ACCOUNT_RESP.UInt16(), sendAccount)

	// 所有邮件-----------------------------------------------------------------
	all_email_model := inst.GetAllEmail()
	count = maxsend
	sendEmail := &inner.ALL_EMAIL_RESP{}
	for _, v := range all_email_model {
		pbData := &inner.SAVE_EMAIL_PERSON{}
		proto.Unmarshal(v.Data, pbData)
		sendEmail.AcccountMail = append(sendEmail.AcccountMail, pbData)
		count--
		if count <= 0 {
			send_tools.Send2Hall(inner.SERVERMSG_DH_ALL_EMAIL_RESP.UInt16(), sendEmail)
			sendEmail = &inner.ALL_EMAIL_RESP{}
			count = maxsend
		}
	}
	send_tools.Send2Hall(inner.SERVERMSG_DH_ALL_EMAIL_RESP.UInt16(), sendEmail)

	// 所有水位线-----------------------------------------------------------------
	all_water_line := inst.GetAllWaterLine()
	count = maxsend
	sendWaterLine := &inner.ALL_WATER_LINE{}
	for _, waterline := range all_water_line {
		pbline := &inner.SAVE_WATER_LINE{}
		tools.CopyProtoData(waterline, pbline)
		sendWaterLine.Line = append(sendWaterLine.Line, pbline)
		count--
		if count <= 0 {
			send_tools.Send2Hall(inner.SERVERMSG_DH_ALL_WATER_LINE.UInt16(), sendWaterLine)
			sendWaterLine = &inner.ALL_WATER_LINE{}
			count = maxsend
		}
	}
	send_tools.Send2Hall(inner.SERVERMSG_DH_ALL_WATER_LINE.UInt16(), sendWaterLine)

	// 所有房间水池-----------------------------------------------------------------
	all_room_bonus := inst.GetAllRoomBonus()
	count = maxsend
	sendRoomBonus := &inner.ALL_ROOM_BONUS{}
	for _, room_bouns := range all_room_bonus {
		pb := &inner.SAVE_ROOM_BONUS{}
		tools.CopyProtoData(room_bouns, pb)
		sendRoomBonus.Bonus = append(sendRoomBonus.Bonus, pb)
		count--
		if count <= 0 {
			send_tools.Send2Hall(inner.SERVERMSG_DH_ALL_ROOM_BONUS.UInt16(), sendRoomBonus)
			sendWaterLine = &inner.ALL_WATER_LINE{}
			count = maxsend
		}
	}
	send_tools.Send2Hall(inner.SERVERMSG_DH_ALL_ROOM_BONUS.UInt16(), sendRoomBonus)

	// todo 所有数据发送完毕
	send_tools.Send2Hall(inner.SERVERMSG_DH_FINISH_DATA.UInt16(), nil)
}

// 回存玩家数据
func (self *mysql_server) SERVERMSG_HD_SAVE_ACCOUNT(pbmsg []byte, session int64) bool {
	pbData := packet.PBUnmarshal(pbmsg, &inner.SAVE_ACCOUNT{}).(*inner.SAVE_ACCOUNT)
	accModel := &inst.AccountModel{}
	tools.CopyProtoData(pbData.GetAccData(), accModel)

	time_before := utils.MilliSecondTimeSince1970()
	save_err := accModel.Save()
	time_after := utils.MilliSecondTimeSince1970()

	log.Infof("MSGID_HG_SAVE_ACCOUNT:%v", pbData.String())
	if time_diff := time_after - time_before; time_diff > 200 {
		log.Warnf("回存玩家数据时间:%v sytnax:%v", time_diff, pbData.String())
	}
	if save_err != nil {
		log.Warnf("回存玩家数据出错,数据:%+v 错误:%v", *accModel, save_err.Error())
		return false
	}
	return true
}

// 大厅回存邮件数据
func (self *mysql_server) SERVERMSG_HD_SAVE_EMAIL_PERSON(pack packet.IPacket, session int64) {
	pbData := pack.ReadBytes()
	emails := &inner.SAVE_EMAIL_PERSON{}
	err := proto.Unmarshal(pbData, emails)
	if err != nil {
		log.Errorf("解析出错:%v", err)
		return
	}
	saveEmail := inst.EmailModel{
		AccountId: emails.AccountId,
		Data:      pbData,
	}
	log.Infof("MSGID_HG_SAVE_EMAIL_PERSON:%v", emails.String())
	time_before := utils.MilliSecondTimeSince1970()
	save_err := saveEmail.Save()
	time_after := utils.MilliSecondTimeSince1970()

	if time_diff := time_after - time_before; time_diff > 200 {
		log.Warnf("回存邮件数据时间:%v sytnax:%v", time_diff, emails.String())
	}
	if save_err != nil {
		log.Warnf("回存邮件数据出错,数据:%+v 错误:%v", saveEmail, save_err.Error())
		core.CoreSend(self.owner.Id, self.owner.Id, pack.GetData(), session)
	}
}

// 回存水位线
func (self *mysql_server) SERVERMSG_HD_SAVE_WATER_LINE(pack packet.IPacket, session int64) {
	pbData := pack.ReadBytes()
	water_line := &inner.SAVE_WATER_LINE{}
	err := proto.Unmarshal(pbData, water_line)
	if err != nil {
		log.Errorf("解析出错:%v", err)
		return
	}
	saveWaterLine := &inst.WaterLineModel{}
	tools.CopyProtoData(water_line, saveWaterLine)

	log.Infof("MSGID_HG_SAVE_WATER_LINE:%v", water_line.String())
	time_before := utils.MilliSecondTimeSince1970()
	save_err := saveWaterLine.Save()
	time_after := utils.MilliSecondTimeSince1970()

	if time_diff := time_after - time_before; time_diff > 200 {
		log.Warnf("回存水位线时间:%v sytnax:%+v", time_diff, *saveWaterLine)
	}
	if save_err != nil {
		log.Warnf("回存水位线数据出错,数据:%+v 错误:%v", *saveWaterLine, save_err.Error())
	}
}

// 回存房间水池
func (self *mysql_server) SERVERMSG_HD_SAVE_ROOM_BONUS(pack packet.IPacket, session int64) {
	pbData := pack.ReadBytes()
	room_bonus := &inner.ROOM_BONUS_SAVE{}
	err := proto.Unmarshal(pbData, room_bonus)
	if err != nil {
		log.Errorf("解析出错:%v", err)
		return
	}
	saveRoomBouns := &inst.RoomBonusModel{}
	tools.CopyProtoData(room_bonus, saveRoomBouns)

	log.Infof("MSGID_HG_SAVE_WATER_LINE:%v", room_bonus.String())
	time_before := utils.MilliSecondTimeSince1970()
	save_err := saveRoomBouns.Save()
	time_after := utils.MilliSecondTimeSince1970()

	if time_diff := time_after - time_before; time_diff > 200 {
		log.Warnf("回存水位线时间:%v sytnax:%+v", time_diff, *saveRoomBouns)
	}
	if save_err != nil {
		log.Warnf("回存水位线数据出错,数据:%+v 错误:%v", *saveRoomBouns, save_err.Error())
	}
}
