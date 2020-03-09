package account

import (
	"root/common"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/utils"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/hall/logcache"
	"root/server/hall/send_tools"
)

type (
	Account struct {
		*protomsg.AccountStorageData
		*protomsg.AccountGameData
		SessionId int64
		Store     bool
	}
)

func NewAccount(storageData *protomsg.AccountStorageData) *Account {
	return &Account{
		AccountStorageData: storageData,
		AccountGameData:    &protomsg.AccountGameData{},
		SessionId:          0,
		Store:              false,
	}
}

// 参数: 是否忽略IsChangeData条件;
// 传true表示无条件回存  传false表示要满足有数据改变才回存
func (self *Account) Save() {
	if self.Robot == 0 {
		self.Store = false
		log.Infof(colorized.White("回存玩家:%v money:%v safemoney:%v"), self.AccountId, self.Money, self.SafeMoney)
		send_tools.Send2DB(inner.SERVERMSG_HD_SAVE_ACCOUNT.UInt16(), &inner.SAVE_ACCOUNT{AccData: self.AccountStorageData})
	}
}

func (self *Account) IsOnline() bool {
	return self.LoginTime-self.LogoutTime > 0
}

// 操作保险箱
func (self *Account) AddSafeMoney(iValue int64, operate common.EOperateType) {
	if iValue == 0 {
		return
	}

	iRMB := int64(self.Money) - iValue
	iSafeRMB := int64(self.SafeMoney) + iValue
	if iRMB < 0 || iRMB > 999999999 {
		log.Errorf("钱越界了 Money 玩家ID:%v Money:%v 保险箱:%v, 保险箱想要改变:%v", self.AccountId, self.Money, self.SafeMoney, iValue)
		return
	} else if iSafeRMB < 0 || iSafeRMB > 999999999 {
		log.Errorf("钱越界了 SafeMoney 玩家ID:%v Money:%v 保险箱:%v, 保险箱想要改变:%v", self.AccountId, self.Money, self.SafeMoney, iValue)
		return
	}

	if iValue < 0 {
		// 从保险箱取钱; 日志从AddMoney函数记录
		self.SafeMoney = uint64(iSafeRMB)
		self.AddMoney(-iValue, common.EOperateType_SAFE_MONEY_GET)
		log.Infof("玩家ID:%v 从保险箱取出金额:%v  操作后身上:%v  保险箱剩余:%v", self.AccountId, -iValue, self.Money, self.SafeMoney)
	} else {
		// 存钱到保险箱; 日志从AddMoney函数记录
		self.SafeMoney = uint64(iSafeRMB)
		self.AddMoney(-iValue, common.EOperateType_SAFE_MONEY_SAVE)
		log.Infof("玩家ID:%v 存入保险箱金额:%v  操作后身上:%v  保险箱剩余:%v", self.AccountId, iValue, self.Money, self.SafeMoney)
	}

	//db.HSet(rediskey.PlayerId(uint32(self.AccountId)), "SafeMoney", self.SafeMoney)
}

func (self *Account) AddMoney(iValue int64, operate common.EOperateType) {
	if iValue == 0 {
		return
	}

	money := int64(self.Money) + iValue
	if money < 0 || money > 9999999999999 {
		log.Errorf("钱越界了 :[%]", money)
		return
	}

	if self.Robot == 0 {
		strTime := utils.DateString()
		logpb := &inner.MONEYCHANGE{
			AccountID:   self.AccountId,
			ChangeValue: iValue,
			Value:       money,
			Operate:     uint32(operate),
			Time:        strTime,
			RoomID:      0,
		}
		logcache.LogCache.AddMoneyChangeLog(logpb) // 金币改变
		//db.HSet(rediskey.PlayerId(uint32(self.AccountId)), "Money", self.Money)
	}
	self.Money = uint64(money)
	self.Store = true

	send_tools.Send2Account(protomsg.MSG_SC_UPDATE_MONEY.UInt16(), &protomsg.UPDATE_MONEY{
		Operate: int32(operate),
		Money:   self.Money,
	}, self.SessionId)
}
