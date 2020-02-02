package logic

import (
	"regexp"
	"root/common"
	"root/core/log"
	"root/core/packet"
	"root/core/utils"
	"github.com/golang/protobuf/proto"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/hall/account"
	"root/server/hall/send_tools"
)

// 有客户端断开连接, 可能是游戏, 也可能是玩家
func (self *Hall) MSGID_CLOSE_CONNECT(actor int32, msg []byte, session int64) {
	tAccount := account.AccountMgr.GetAccountBySessionID(session)
	// 游戏进程断开连接相关处理
	if tAccount == nil {
		GameMgr.GameDisconnect(session)
		return
	}
	// 玩家断开连接相关的处理
	account.AccountMgr.RemoveAccountBySessionID(session)
	if tAccount.SessionId != session{
		log.Warnf("Account SessionID:%v, Curr SessionID:%v", tAccount.SessionId, session)
		return
	}
	tAccount.SessionId = 0
	tAccount.LogoutTime = utils.SecondTimeSince1970()
	log.Infof("<- Out Player:ID:%v, Name:%v, RoomID:%v, Money:%v SafeMoney:%v,Session:%v\r\n", tAccount.AccountId, tAccount.Name, tAccount.RoomID, tAccount.Money, tAccount.SafeMoney, session)
}

// db返回的所有玩家数据
func (self *Hall) SERVERMSG_DH_ALL_ACCOUNT_RESP(actor int32, msg []byte, session int64) {
	if session != 0 {
		log.Infof("Error: 不是来自于DB服务器的消息,SessionID:%v", session)
		return
	}
	accounts := &inner.ALL_ACCOUNT_RESP{}
	error := proto.Unmarshal(msg, accounts)
	if error != nil {
		log.Errorf("加载所有玩家数据出错 :%v", error.Error())
		return
	}
	account.AccountMgr.LoadAllAccount(accounts.AllAccount)
}
// db返回的所有邮件数据
func (self *Hall) SERVERMSG_DH_ALL_EMAIL_RESP(actor int32, msg []byte, session int64) {
	if session != 0 {
		log.Infof("Error: 不是来自于DB服务器的消息, MSGID_GH_ALL_EMAIL, SessionID:%v", session)
		return
	}
	all_email := &inner.ALL_EMAIL_RESP{}
	err := proto.Unmarshal(msg, all_email)
	if err != nil {
		log.Errorf("邮件数据读取错误:%v", err)
		return
	}
	account.EmailMgr.LoadAll(all_email)
}

func (self *Hall) MSG_CS_BIND_PHONE_REQ(actor int32, msg []byte, session int64) {
	pbMsg := packet.PBUnmarshal(msg,&protomsg.BIND_PHONE_REQ{}).(*protomsg.BIND_PHONE_REQ)
	strPhone := pbMsg.GetPhone()

	acc := account.AccountMgr.GetAccountBySessionIDAssert(session)
	if acc.Phone != ""{
		send_tools.Send2Account(protomsg.MSG_SC_BIND_PHONE_RES.UInt16(), &protomsg.BIND_PHONE_RES{Ret:1},session)
		log.Warnf("玩家已绑定手机号，不能重复绑定 acc:%v phone:%v strPhone:%v",acc.GetAccountId(),acc.GetPhone(),strPhone)
		return
	}

	m,_ := regexp.MatchString(utils.PHONE_REG,strPhone)
	if !m{
		send_tools.Send2Account(protomsg.MSG_SC_BIND_PHONE_RES.UInt16(), &protomsg.BIND_PHONE_RES{Ret:2},session)
		log.Warnf(" 手机号格式不正确 acc:%v phone:%v",acc.GetAccountId(),strPhone)
		return
	}
	acc.Phone = strPhone
	send_tools.Send2Account(protomsg.MSG_SC_BIND_PHONE_RES.UInt16(), &protomsg.BIND_PHONE_RES{Ret:0},session)
	log.Infof("玩家:[%v] 绑定手机号:[%v] 成功",acc.GetAccountId(),acc.GetPhone())
}

func (self *Hall) MSG_CS_SAFEMONEY_OPERATE_REQ(actor int32, msg []byte, session int64) {
	pbMsg := packet.PBUnmarshal(msg,&protomsg.SAFEMONEY_OPERATE_REQ{}).(*protomsg.SAFEMONEY_OPERATE_REQ)
	acc := account.AccountMgr.GetAccountBySessionIDAssert(session)

	if pbMsg.GetOperate() == 1{	// 取钱
		if pbMsg.GetOperateMoney() > acc.GetSafeMoney(){
			log.Warnf("玩家:%v 请求操作保险箱，取钱超过保险箱的钱:%+v 保险箱的钱:%v ",acc.GetAccountId(),*pbMsg,acc.GetSafeMoney())
			return
		}
		acc.SafeMoney -= pbMsg.GetOperateMoney()
		acc.AddMoney(int64(pbMsg.GetOperateMoney()),common.EOperateType_SAFE_MONEY_GET)
	}else if  pbMsg.GetOperate() == 2{ // 存钱
		if pbMsg.GetOperateMoney() < acc.GetMoney(){
			log.Warnf("玩家:%v 请求操作保险箱，存钱超过身上的钱:%+v 身上的钱:%v  ",acc.GetAccountId(),*pbMsg,acc.GetMoney())
			return
		}
		acc.SafeMoney += pbMsg.GetOperateMoney()
		acc.AddMoney(-int64(pbMsg.GetOperateMoney()),common.EOperateType_SAFE_MONEY_SAVE)
	}else{
		log.Warnf("玩家:%v 请求操作保险箱，数据错误:%+v ",acc.GetAccountId(),*pbMsg)
		return
	}

	send_tools.Send2Account(protomsg.MSG_SC_BIND_PHONE_RES.UInt16(), &protomsg.BIND_PHONE_RES{Ret:0},session)
	log.Infof("玩家:[%v] 操作保险箱:[%v] 金额:%v 成功，身上的金币:%v 保险箱:%v  ",acc.GetAccountId(),pbMsg.GetOperate(),pbMsg.GetOperateMoney(),acc.GetMoney(),acc.GetSafeMoney())
}