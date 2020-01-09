package account

import (
	"root/common"
	"root/common/model/rediskey"
	"root/core/db"
	"root/core/log"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/hongbao/send_tools"
)

const (
	BET_KIND = 3
)

type (
	Account struct {
		*protomsg.AccountStorageData
		*protomsg.AccountGameData
		SessionId int64
		Quit_flag bool
	}
)

func NewAccount(storageData *protomsg.AccountStorageData) *Account {
	gameData := &protomsg.AccountGameData{
		State: common.STATUS_OFFLINE.UInt32(),
	}
	return &Account{
		AccountStorageData: storageData,
		AccountGameData:    gameData,
		SessionId:          0,
		Quit_flag:          false,
	}
}

func (self *Account) IsOnline() uint8 {
	return uint8(self.State)
}

func (self *Account) AddMoney(iValue int64, index uint8, operate common.EOperateType) {
	if iValue == 0 {
		return
	}

	money := int64(self.RMB) + iValue
	if money < 0 || money > 999999999 {
		log.Errorf("钱越界了 :[%] accid:%v ", money, self.AccountId)
		return
	}
	self.RMB = uint64(money)

	if self.Robot == 0 {
		strTime := utils.DateString()
		tSave := packet.NewPacket(nil)
		tSave.SetMsgID(protomsg.MSGID_SAVE_RMB_CHANGE_LOG.UInt16())
		tSave.WriteUInt32(self.AccountId)
		tSave.WriteInt64(iValue)
		tSave.WriteInt64(money)
		tSave.WriteUInt8(index)
		tSave.WriteUInt8(uint8(operate))
		tSave.WriteString(strTime)
		tSave.WriteUInt32(self.RoomID)
		tSave.WriteUInt8(common.EGameTypeHONG_BAO.Value())
		send_tools.Send2Hall(tSave.GetData())
		db.HSet(rediskey.PlayerId(uint32(self.AccountId)), "Money", self.RMB)
	}

	self.SyncToHall_Money(self.RMB)

}

func (self *Account) GetMoney() uint64 {
	return self.RMB
}

func (self *Account) SyncToHall_Money(val uint64) {
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_SYNC_TO_HALL_MONEY.UInt16())

	send.WriteUInt32(self.AccountId)
	send.WriteInt64(int64(self.RMB))
	send.WriteInt64(int64(self.SafeRMB))

	send_tools.Send2Hall(send.GetData())
}

func (self *Account) UpdateEnter(roomId uint32, count uint16, watch uint8) {
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_UPDATE_ENTER.UInt16())

	send.WriteUInt32(self.AccountId)
	send.WriteUInt32(roomId)
	send.WriteUInt16(count)
	send.WriteUInt8(watch)
	send_tools.Send2Hall(send.GetData())
}

func (self *Account) UpdateLeave(roomId uint32, count uint16, watch uint8) {
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_UPDATE_LEAVE.UInt16())

	send.WriteUInt32(self.AccountId)
	send.WriteUInt32(roomId)
	send.WriteUInt16(count)
	send.WriteUInt8(watch)
	send_tools.Send2Hall(send.GetData())
}
