package logic

import (
	"regexp"
	"root/common"
	"root/common/config"
	"root/core/log"
	"root/core/packet"
	"root/core/utils"
	"fmt"
	"github.com/golang/protobuf/proto"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/hall/account"
	"root/server/hall/send_tools"
	"root/server/hall/server"
	"root/server/hall/types"
	"unicode/utf8"
)

// 有客户端断开连接, 可能是游戏, 也可能是玩家
func (self *Hall) MSGID_CLOSE_CONNECT(actor int32, msg []byte, session int64) {
	tAccount := account.AccountMgr.GetAccountBySessionID(session)

	// 游戏进程断开连接相关处理
	if tAccount == nil {
		HallMgr.UnMapServerNode(session)
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

	account.AccountMgr.SendExitLog(tAccount.AccountId)
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

func (self *Hall) Old_MSGID_RECV_ACCOUNT_INFO(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	nAccountID := pack.ReadUInt32()
	nRoomID := pack.ReadUInt32()

	tAccount := account.AccountMgr.GetAccountByID(nAccountID)
	if tAccount == nil {
		log.Warnf("Can't find AccountID:%v", nAccountID)
		return
	}

	isGameServerSession := server.ServerMgr.IsGameServerSession(session)
	if isGameServerSession == false {
		log.Infof("Error: Session:%v 游戏通知大厅可进房间错误, 玩家ID:%v 所在房间:%v 想进入房间:%v", session, nAccountID,  tAccount.RoomID, nRoomID)
		return
	}

	tRoom := HallMgr.GetRoom(nRoomID)
	if tRoom == nil {
		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(protomsg.Old_MSGID_ENTER_ROOM.UInt16())
		tSend.WriteUInt8(3)
		send_tools.Send2Account(tSend.GetData(), tAccount.SessionId)
		return
	}

	if tAccount.Robot > 0 {
		log.Warnf("In GameType:%v, The Robot does't need to respond to this protocol:%v", common.EGameType(tRoom.nGameType), protomsg.Old_MSGID_RECV_ACCOUNT_INFO.UInt16())
		return
	}

	tServerNode := server.ServerMgr.GetServerNode(tRoom.nServerID)
	if tServerNode == nil || HallMgr.nMaintenanceTime > 0 {
		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(protomsg.Old_MSGID_ENTER_ROOM.UInt16())
		tSend.WriteUInt8(4)
		send_tools.Send2Account(tSend.GetData(), tAccount.SessionId)
		return
	}

	// 记录玩家已经进入该房间; 若客户端未连接上游戏房间; 则一直保存该状态和房间ID
	tAccount.RoomID = nRoomID
	tAccount.GameType = uint32(tRoom.nGameType)

	HallMgr.clearDeskList(tRoom.nGameType, tAccount.AccountId, tRoom.clubID)

	// 通知客户端进入房间
	tSend := packet.NewPacket(nil)
	tSend.SetMsgID(protomsg.Old_MSGID_ENTER_ROOM.UInt16())
	tSend.WriteUInt8(0)
	tSend.WriteUInt32(nRoomID)
	tSend.WriteUInt8(tRoom.nGameType)

	// 返回房间对应的游戏节点IP; 以桌子为单位加强防御; 避免所有桌子都掉线
	strGameNodeIP := "Error IP"
	isTestServer, _, strRealIP := config.IsTestServer()
	if isTestServer == true {
		strGameNodeIP = strRealIP
	} else {
		if tRoom.nGameNodeID > 0 {
			strGameNodeIP = HallMgr.ipNodes[int(tRoom.nGameNodeID)]
		} else {
			sIPList := make([]string, 0, 10)
			for _, strGameNodeIP := range HallMgr.ipNodes {
				sIPList = append(sIPList, strGameNodeIP)
			}
			nIPLen := len(sIPList)
			if len(sIPList) > 0 {
				strGameNodeIP = sIPList[utils.Randx_y(0, nIPLen)]
			}
		}
	}
	tSend.WriteString(strGameNodeIP)
	tSend.WriteUInt32(tRoom.clubID)
	send_tools.Send2Account(tSend.GetData(), tAccount.SessionId)
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


func (self *Hall) MSG_CS_BIND_PHONE(actor int32, msg []byte, session int64) {
	pbMsg := packet.PBUnmarshal(msg,&protomsg.BIND_PHONE_REQ{}).(*protomsg.BIND_PHONE_REQ)
	strPhone := pbMsg.GetPhone()

	acc := account.AccountMgr.GetAccountBySessionID(session)
	if acc.Phone != ""{
		send_tools.Send2Account(protomsg.MSG_SC_BIND_PHONE.UInt16(), &protomsg.BIND_PHONE_RESP{Ret:1},session)
		log.Warnf("玩家已绑定手机号，不能重复绑定 acc:%v phone:%v strPhone:%v",acc.GetAccountId(),acc.GetPhone(),strPhone)
		return
	}

	m,_ := regexp.MatchString(utils.PHONE_REG,strPhone)
	if !m{
		send_tools.Send2Account(protomsg.MSG_SC_BIND_PHONE.UInt16(), &protomsg.BIND_PHONE_RESP{Ret:2},session)
		log.Warnf(" 手机号格式不正确 acc:%v phone:%v",acc.GetAccountId(),strPhone)
		return
	}
	acc.Phone = strPhone
	send_tools.Send2Account(protomsg.MSG_SC_BIND_PHONE.UInt16(), &protomsg.BIND_PHONE_RESP{Ret:0},session)
	log.Infof("玩家:[%v] 绑定手机号:[%v] 成功",acc.GetAccountId(),acc.GetPhone())
}