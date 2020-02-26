package account

import (
	"root/common"
	"root/common/model/rediskey"
	"root/core/db"
	"root/core/log"
	"root/core/utils"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/game_dfdc/send_tools"
)
type (
	Account struct {
		*protomsg.AccountStorageData
		*protomsg.AccountGameData
		SessionId int64
		FeeCount  int32
		LastBet   uint64
		Forbid bool
		StaticFee int64
	}
)

func NewAccount(storageData *protomsg.AccountStorageData) *Account {
	gameData := &protomsg.AccountGameData{}
	return &Account{
		AccountStorageData: storageData,
		AccountGameData:    gameData,
		SessionId:          0,
	}
}

func (self *Account) IsOnline() bool {
	return self.LoginTime - self.LogoutTime > 0
}

func (self *Account) AddMoney(iValue int64, operate common.EOperateType) {
	if iValue == 0 {
		return
	}

	money := int64(self.Money) + iValue
	if money < 0 || money > 9999999999999 {
		log.Errorf("钱越界了 :[%] accid:%v ", money, self.AccountId)
		return
	}
	if self.Robot == 0 {
		strTime := utils.DateString()
		moneyChange := &inner.MONEYCHANGE{
			AccountID:   self.GetAccountId(),
			ChangeValue: iValue,
			Value:       money,
			Operate:     uint32(operate),
			Time:        strTime,
			RoomID:      self.GetRoomID(),
		}
		send_tools.Send2Hall(inner.SERVERMSG_GH_MONEYCHANGE.UInt16(),moneyChange)
		db.HSet(rediskey.PlayerId(uint32(self.AccountId)), "Money", self.Money)
	}
	self.Money = uint64(money)
}

func (self *Account) GetMoney() uint64 {
	return self.Money
}