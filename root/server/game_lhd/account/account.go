package account

import (
	"root/common"
	"root/common/model/rediskey"
	"root/core/db"
	"root/core/log"
	"root/core/utils"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/game_lhd/send_tools"
)

const (
	BET_KIND = 3
)

type (
	Account struct {
		*protomsg.AccountStorageData
		*protomsg.AccountGameData
		Games          int32 // 游戏连续耍了多少局
		SessionId      int64
		BetVal         [BET_KIND + 1]uint32 // 下注金额 1龙、2虎、3和
		IsAllowBetting bool                 // 当前局能否下注
		Betcount       int32                // 下注缓存数量
		CLeanTime      int64                // 清除下注时间
	}

	Master struct {
		*Account
		Share int64 // 认购份额
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
	return self.LoginTime-self.LogoutTime > 0
}

func (self *Account) AddMoney(iValue int64, operate common.EOperateType) {
	if iValue == 0 {
		return
	}

	money := int64(self.Money) + iValue
	if money < 0 || money > 99999999999 {
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
		send_tools.Send2Hall(inner.SERVERMSG_GH_MONEYCHANGE.UInt16(), moneyChange) // game_lhd
		db.HSet(rediskey.PlayerId(uint32(self.AccountId)), "Money", self.Money)
	}
	self.Money = uint64(money)
}

func (self *Account) GetMoney() uint64 {
	return self.Money
}
