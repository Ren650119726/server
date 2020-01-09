package room

import (
	"root/common/config"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/mahjong-dgk/send_tools"
	"root/server/mahjong-panda/account"
	"root/server/mahjong-panda/types"
)

type (
	settlement struct {
		*Room
		s         types.ERoomStatus
		timestamp int64
	}
)

func (self *settlement) Enter(now int64) {
	duration := config.GetPublicConfig_Int64("PANDA_SETTLEMENT_TIME") // 持续时间 秒
	if self.liuju {
		duration += 2
	}
	self.timestamp = utils.SecondTimeSince1970() + int64(duration)
	self.settle_total_profit.WriteUInt16(uint16(len(self.seats)))
	servicepack2 := packet.NewPacket(nil)
	for index, player := range self.seats {
		player.timeout_times = 0
		player.trusteeship = 0
		player.decide_t = 0
		player.money_after = player.acc.GetMoney()
		cur_profit := int64(player.money_after) - int64(player.money_before)
		player.acc.Profit += cur_profit // 总盈利
		self.settle_total_profit.WriteInt8(int8(index + 1))
		self.settle_total_profit.WriteInt64(cur_profit)
		self.settle_total_profit.WriteUInt16(uint16(len(player.cards.hand)))
		for _, v := range player.cards.hand {
			self.settle_total_profit.WriteInt8(int8(v))
		}

		self.track_log(colorized.Green("accid:%v profit:%v, after:%v before:%v cur_profit:%v"), player.acc.AccountId, player.acc.Profit, int64(player.money_after), int64(player.money_before), cur_profit)

		servicepack2.WriteUInt32(player.acc.AccountId)
		servicepack2.WriteInt64(int64(player.acc.GetMoney()))
		servicepack2.WriteInt64(int64(cur_profit))
		servicepack2.WriteString("")
	}

	msg := self.SettleMsg()
	if self.liuju {
		msg.WriteInt8(1)
	} else {
		msg.WriteInt8(0)
	}

	msg.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_SETTLE.UInt16())
	self.SendBroadcast(msg.GetData())

	updateAccount := packet.NewPacket(nil)
	updateAccount.SetMsgID(protomsg.Old_MSGID_UPDATE_ACCOUNT.UInt16())
	updateAccount.WriteUInt32(self.roomId)
	updateAccount.WriteUInt8(0)
	updateAccount.WriteUInt16(uint16(len(self.seats)))
	updateAccount.CatBody(servicepack2)
	send_tools.Send2Hall(updateAccount.GetData())
	self.track_log(colorized.Green("settlement enter"))

}

func (self *settlement) SaveQuit(accid uint32) bool {
	return false
}
func (self *settlement) Tick(now int64) {
	if now >= self.timestamp {
		self.switchStatus(now, types.ERoomStatus_WAITING)
		return
	}
}

func (self *settlement) CombineMSG(packet packet.IPacket, acc *account.Account) {
	packet.WriteInt64(self.timestamp * 1000)
	msg := self.SettleMsg()
	packet.CatBody(msg)
}

func (self *settlement) Leave(now int64) {
	self.track_log(colorized.Green("settlement leave\n"))
}

///////////////////////////////// handler ///////////////////////////////////////////////////
func (self *settlement) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	default:
		log.Warnf("settlement 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}
	return true
}
