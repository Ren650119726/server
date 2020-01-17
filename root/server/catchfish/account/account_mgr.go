package account

import (
	"root/core/log"
	"root/protomsg"
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


// 创建账号
func (self *accountMgr) RecvAccount(storage *protomsg.AccountStorageData, game *protomsg.AccountGameData) {

	newAcc := NewAccount(storage)
	newAcc.AccountGameData = game
	//oldAcc := self.GetAccountByID(storage.AccountId)

	self.Lock.Lock()
	defer self.Lock.Unlock()
	//if oldAcc != nil {
	//	oldRMB := oldAcc.RMB
	//	oldSafeRMB := oldAcc.SafeRMB
	//	oldAcc.AccountStorageData = storage
	//	oldAcc.AccountGameData = game
	//	if oldAcc.RoomID > 0 {
	//		// 玩家已经在游戏中, 不更新元宝和保险箱
	//		oldAcc.RMB = oldRMB
	//		oldAcc.SafeRMB = oldSafeRMB
	//	}
	//} else {
	//	self.accountbyID[storage.AccountId] = newAcc
	//}
	//self.accountbyID[storage.AccountId].State = common.STATUS_ONLINE.UInt32()
	//if newAcc.Robot == 0 {
	//	// 通知大厅，可以让客户端连上游戏了
	//	send := packet.NewPacket(nil)
	//	send.SetMsgID(protomsg.Old_MSGID_RECV_ACCOUNT_INFO.UInt16())
	//	send.WriteUInt32(storage.AccountId)
	//	send.WriteUInt32(game.RoomID)
	//	send_tools.Send2Hall(send.GetData())
	//}

}

// 客户端连接
func (self *accountMgr) EnterAccount(accountId uint32, roomId uint32, session int64) bool {
	//send := packet.NewPacket(nil)
	//send.SetMsgID(protomsg.Old_MSGID_ENTER_GAME.UInt16())
	//
	//if session == 0 {
	//	// 机器人不关联session
	//	return true
	//}
	//acc := AccountMgr.GetAccountByID(accountId)
	//if acc == nil {
	//	send.WriteUInt8(1)
	//	send_tools.Send2Account(send.GetData(), session)
	//	return false
	//}
	//
	////if acc.IsOnline() == common.STATUS_OFFLINE.UInt8() {
	////	send.WriteUInt8(7)
	////	send_tools.Send2Account(send.GetData(), session)
	////	return false
	////}
	//
	//// 关联账号的session
	//if acc.SessionId > 0 {
	//	delete(self.accountbySessionID, acc.SessionId)
	//}
	//self.Lock.Lock()
	//acc.SessionId = session
	//self.accountbySessionID[session] = acc
	//defer self.Lock.Unlock()
	return true
}

// 连接断开处理
func (self *accountMgr) DisconnectAccount(acc *Account) bool {
	self.Lock.Lock()
	delete(self.accountbyID, acc.AccountId)
	delete(self.accountbySessionID, acc.SessionId)
	self.Lock.Unlock()

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
