package account

import (
	"root/common"
	"root/core/log"
	"root/core/packet"
	"root/protomsg"
	"root/server/dehgame/send_tools"
)

var AccountMgr = newAccountMgr()

type (
	accountMgr struct {
		accountbyID        map[uint32]*Account
		accountbySessionID map[int64]*Account
	}
)

func newAccountMgr() *accountMgr {
	return &accountMgr{
		accountbyID:        make(map[uint32]*Account),
		accountbySessionID: make(map[int64]*Account),
	}
}

func (self *accountMgr) GetAccountByID(id uint32) *Account {
	return self.accountbyID[id]
}

func (self *accountMgr) GetAccountBySessionID(session int64) *Account {
	return self.accountbySessionID[session]
}

// 广播消息, 给所有在线玩家
func (self *accountMgr) SendBroadcast(pack packet.IPacket) {
	for _, acc := range self.accountbyID {
		if acc.IsOnline() == common.STATUS_ONLINE.UInt8() && acc.Robot == 0 {
			send_tools.Send2Account(pack.GetData(), acc.SessionId)
		}
	}
}

//
//func (self *accountMgr) UpdateStart(nRoomID uint32, nStart uint8) {
//
//	send := packet.NewPacket(nil)
//	send.SetMsgID(protomsg.Old_MSGID_UPDATE_START.UInt16())
//	send.WriteUInt32(nRoomID)
//	send.WriteUInt8(nStart)
//	send_tools.Send2Hall(send.GetData())
//}

// 创建账号
func (self *accountMgr) RecvAccount(storage *protomsg.AccountStorageData, game *protomsg.AccountGameData) {
	oldAcc := self.GetAccountByID(storage.AccountId)
	if oldAcc == nil {
		newAcc := NewAccount(storage)
		//newAcc.AccountGameData = game
		self.accountbyID[storage.AccountId] = newAcc
	} else {
		oldAcc.AccountStorageData = storage
		oldAcc.AccountGameData = game
	}

	// 通知大厅，可以让客户端连上游戏了
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_RECV_ACCOUNT_INFO.UInt16())
	send.WriteUInt32(storage.AccountId)
	send.WriteUInt32(game.RoomID)
	send_tools.Send2Hall(send.GetData())
}

// 客户端连接
func (self *accountMgr) EnterAccount(accountId uint32, roomId uint32, session int64) bool {
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_ENTER_GAME.UInt16())
	acc := AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return false
	}
	acc.State = common.STATUS_ONLINE.UInt32()

	// 关联账号的session
	if acc.SessionId > 0 {
		delete(self.accountbySessionID, acc.SessionId)
	}
	acc.SessionId = session
	self.accountbySessionID[session] = acc
	return true
}

// 连接断开处理
func (self *accountMgr) DisconnectAccount(session int64) bool {
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_LEAVE_GAME.UInt16())

	acc := self.GetAccountBySessionID(session)
	if acc == nil {
		return false
	}

	delete(self.accountbyID, acc.AccountId)
	delete(self.accountbySessionID, session)

	return true
}
func CheckSession(accountId uint32, session int64) *Account {
	sacc := AccountMgr.GetAccountByID(accountId)
	var seAcc *Account
	seAccId := uint32(0)
	seAccName := ""
	if session != 0 {
		seAcc = AccountMgr.GetAccountBySessionID(session)
		if seAcc != nil {
			seAccId = seAcc.AccountId
			seAccName = seAcc.Name
		}
	}

	if sacc == nil {
		log.Errorf("作弊, session:%v 验证的accountId:%v session对应的玩家 accid:%v,name:%v", session, accountId, seAccId, seAccName)
		panic(nil)
	} else if sacc.SessionId != session && session != 0 {
		log.Errorf("作弊, session:%v accountID:%v 验证的session:%v accountId:%v Robot:%v", session, accountId, sacc.SessionId, sacc.AccountId, sacc.Robot)
		panic(nil)
	} else {
		return sacc
	}
}
