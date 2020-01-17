package account

import (
	"root/common"
	"root/common/model/rediskey"
	"root/core/db"
	"root/core/log"
	"root/core/utils"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/fruitMary/send_tools"
)
type (
	Account struct {
		*protomsg.AccountStorageData
		*protomsg.AccountGameData
		SessionId int64
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

func (self *Account) AddMoney(iValue int64, index uint8, operate common.EOperateType) {
	if iValue == 0 {
		return
	}

	money := int64(self.Money) + iValue
	if money < 0 || money > 999999999 {
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
func (self *Account) UpdateEnter(roomId uint32, count uint16, watch uint8) {
	//send := packet.NewPacket(nil)
	//send.SetMsgID(protomsg.Old_MSGID_UPDATE_ENTER.UInt16())
	//
	//send.WriteUInt32(self.AccountId)
	//send.WriteUInt32(roomId)
	//send.WriteUInt16(count)
	//send.WriteUInt8(watch)
	//send_tools.Send2Hall(send.GetData())
}

func (self *Account) UpdateLeave(roomId uint32, count uint16, watch uint8) {
	//send := packet.NewPacket(nil)
	//send.SetMsgID(protomsg.Old_MSGID_UPDATE_LEAVE.UInt16())
	//
	//send.WriteUInt32(self.AccountId)
	//send.WriteUInt32(roomId)
	//send.WriteUInt16(count)
	//send.WriteUInt8(watch)
	//send_tools.Send2Hall(send.GetData())
}
