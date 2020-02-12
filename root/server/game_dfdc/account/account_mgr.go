package account

import (
	"root/core/log"
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
func (self *accountMgr) SetAccountByID(acc *Account) {
	self.Lock.Lock()
	defer self.Lock.Unlock()
	self.accountbyID[acc.GetAccountId()] = acc
}
func (self *accountMgr) SetAccountBySession(acc *Account,session int64) {
	self.Lock.Lock()
	defer self.Lock.Unlock()
	if _,e := self.accountbySessionID[acc.SessionId];e{
		delete(self.accountbySessionID,acc.SessionId)
	}
	self.accountbySessionID[session] = acc
}

func (self *accountMgr) GetAccountByIDAssert(id uint32) *Account {
	self.Lock.Lock()
	defer self.Lock.Unlock()

	acc :=  self.accountbyID[id]
	if acc == nil {
		log.Panicf("找不到玩家:%v ",id)
	}
	return acc
}

func (self *accountMgr) GetAccountBySessionID(session int64) *Account {
	self.Lock.Lock()
	defer self.Lock.Unlock()
	return self.accountbySessionID[session]
}

func (self *accountMgr) GetAccountBySessionIDAssert(session int64) *Account {
	self.Lock.Lock()
	defer self.Lock.Unlock()
	acc :=  self.accountbySessionID[session]
	if acc == nil {
		log.Panicf("找不到玩家:%v ",session)
	}
	return acc
}

// 客户端连接
func (self *accountMgr) EnterAccount(accountId uint32, roomId uint32, session int64) bool {
	if session == 0 {
		// 机器人不关联session
		return true
	}
	acc := AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		return false
	}

	// 关联账号的session
	if acc.SessionId > 0 {
		delete(self.accountbySessionID, acc.SessionId)
	}
	self.Lock.Lock()
	acc.SessionId = session
	self.accountbySessionID[session] = acc
	defer self.Lock.Unlock()
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