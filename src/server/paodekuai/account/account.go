package account

import (
	"root/common"
	"root/common/model/rediskey"
	"root/core/db"
	"root/core/log"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/paodekuai/send_tools"
)

type (
	Account struct {
		*protomsg.AccountStorageData
		*protomsg.AccountGameData

		Games     int32 // 游戏连续耍了多少局
		Profit    int64 // 获利的钱，负数表示输
		SessionId int64
	}
)

func NewAccount(storageData *protomsg.AccountStorageData) *Account {
	gameData := &protomsg.AccountGameData{
		State: common.STATUS_ONLINE.UInt32(),
	}

	acc := &Account{
		AccountStorageData: storageData,
		AccountGameData:    gameData,
		Games:              0,
		Profit:             0,
		SessionId:          0,
	}

	return acc
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
		log.Errorf("钱越界了, ID:%v 身上:%v 改变:%v 下标:%v 操作:%v", self.AccountId, self.RMB, iValue, index, operate)
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
		tSave.WriteUInt8(common.EGameTypePAO_DE_KUAI.Value())
		send_tools.Send2Hall(tSave.GetData())
		db.HSet(rediskey.PlayerId(uint32(self.AccountId)), "Money", self.RMB)
	}

	self.SyncToHall_Money(self.RMB)
}

func (self *Account) GetMoney() uint64 {
	return self.RMB
}

func (self *Account) SyncToHall_Money(val uint64) {

	if self.Robot == 0 {
		send := packet.NewPacket(nil)
		send.SetMsgID(protomsg.Old_MSGID_SYNC_TO_HALL_MONEY.UInt16())

		send.WriteUInt32(self.AccountId)
		send.WriteInt64(int64(self.RMB))
		send.WriteInt64(int64(self.SafeRMB))

		send_tools.Send2Hall(send.GetData())
	}
}
