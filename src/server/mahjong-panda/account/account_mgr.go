package account

import (
	"root/core/log"
	"root/core/packet"
	"root/protomsg"
	"root/server/mahjong-panda/send_tools"
	"sync"
)

var AccountMgr = newAccountMgr()

type (
	accountMgr struct {
		Lock               sync.Mutex
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
	self.Lock.Lock()
	defer self.Lock.Unlock()
	return self.accountbyID[id]
}

func (self *accountMgr) GetAccountBySessionID(session int64) *Account {
	self.Lock.Lock()
	defer self.Lock.Unlock()
	return self.accountbySessionID[session]
}

// 广播消息, 给所有在线玩家
func (self *accountMgr) SendBroadcast(pack packet.IPacket) {
	self.Lock.Lock()
	defer self.Lock.Unlock()
	for _, acc := range self.accountbyID {
		if acc.IsOnline() == 1 && acc.Robot == 0 {
			send_tools.Send2Account(pack.GetData(), acc.SessionId)
		}
	}
}

// 创建账号
func (self *accountMgr) RecvAccount(storage *protomsg.AccountStorageData, game *protomsg.AccountGameData, st int8) {
	self.Lock.Lock()
	defer self.Lock.Unlock()
	oldAcc := self.accountbyID[storage.AccountId]
	if oldAcc == nil {
		newAcc := NewAccount(storage)
		newAcc.Robot = game.Robot
		newAcc.AutoSitDown = st
		self.accountbyID[storage.AccountId] = newAcc
	} else {
		oldRMB := oldAcc.RMB
		oldSafeRMB := oldAcc.SafeRMB
		oldAcc.AccountStorageData = storage
		oldAcc.AccountGameData = game
		if oldAcc.RoomID > 0 {
			// 玩家已经在游戏中, 不更新元宝和保险箱
			oldAcc.RMB = oldRMB
			oldAcc.SafeRMB = oldSafeRMB
		}
		if st == 1 {
			oldAcc.AutoSitDown = st
		}
	}

	// 通知大厅，可以让客户端连上游戏了
	if game.Robot == 0 {
		send := packet.NewPacket(nil)
		send.SetMsgID(protomsg.Old_MSGID_RECV_ACCOUNT_INFO.UInt16())
		send.WriteUInt32(storage.AccountId)
		send.WriteUInt32(game.RoomID)
		send_tools.Send2Hall(send.GetData())
	}

}

// 客户端连接
func (self *accountMgr) EnterAccount(accountId uint32, roomId uint32, session int64) bool {
	self.Lock.Lock()
	defer self.Lock.Unlock()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_ENTER_GAME.UInt16())

	acc := self.accountbyID[accountId]
	if acc == nil {
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return false
	}
	acc.State = STATUS_NORMAL

	// 关联账号的session
	if acc.SessionId > 0 {
		delete(self.accountbySessionID, acc.SessionId)
	}
	acc.SessionId = session
	self.accountbySessionID[session] = acc
	return true
}

// 连接断开处理
func (self *accountMgr) DisconnectAccount(accountId uint32) bool {
	self.Lock.Lock()
	defer self.Lock.Unlock()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_LEAVE_GAME.UInt16())

	acc := self.accountbyID[accountId]
	if acc == nil {
		return false
	}

	if acc.SessionId != 0 {
		delete(self.accountbySessionID, acc.SessionId)
	}

	delete(self.accountbyID, acc.AccountId)

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
	} else if sacc.SessionId != session {
		log.Errorf("作弊, session:%v accountID:%v 验证的session:%v accountId:%v Robot:%v", session, accountId, sacc.SessionId, sacc.AccountId, sacc.Robot)
		panic(nil)
	} else {
		return sacc
	}
}
