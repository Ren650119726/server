package room

import (
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/red2black/account"
	"root/server/red2black/send_tools"
)

type (
	master struct {
		*Room
		s ERoomStatus
	}
)

func (self *master) Enter(now int64) {
	duration := self.status_duration[int(self.s)] // 持续时间 秒
	log.Debugf(colorized.Yellow("master enter duration:%v"), duration)

	// 广播房间玩家，切换状态
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_R2B_NEXT_STATE.UInt16())
	send.WriteUInt8(uint8(ERoomStatus_GRAB_MASTER))
	send.WriteUInt32(uint32(int64(utils.SecondTimeSince1970()+int64(duration)) * 1000))
	self.SendBroadcast(send.GetData())

	self.owner.AddTimer(int64(duration*1000), 1, func(dt int64) {
		self.switchStatus(now, ERoomStatus_START_BETTING)
	})

}

func (self *master) Tick(now int64) {

}

func (self *master) Leave(now int64) {
	log.Debugf(colorized.Yellow("master leave\n"))
}

func (self *master) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_ENTER_GAME.UInt16(): // 客户端链接进入游戏
		self.Old_MSGID_R2B_ENTER_GAME(actor, msg, session)
	case protomsg.Old_MSGID_R2B_GAME_ENTER_GAME.UInt16(): // 客户端链接进入游戏
		self.Old_MSGID_R2B_GAME_ENTER_GAME(actor, msg, session)
	default:
		log.Warnf("master 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}

	return true
}

func (self *master) Old_MSGID_R2B_ENTER_GAME(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	_ = pack.ReadUInt32()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_ENTER_GAME.UInt16())
	if ret := self.canEnterRoom(accountId); ret > 0 {
		send.WriteUInt8(uint8(ret))
		send_tools.Send2Account(send.GetData(), session)
		return
	}
	self.enterRoom(accountId)

	// 通知客户端，进入游戏成功
	acc := account.AccountMgr.GetAccountByID(accountId)
	send2c := packet.NewPacket(nil)
	send2c.SetMsgID(protomsg.Old_MSGID_ENTER_GAME.UInt16())
	send2c.WriteUInt8(0)
	send2c.WriteUInt32(self.roomId)
	send_tools.Send2Account(send2c.GetData(), acc.SessionId)
	send2acc := self.sendGameData(acc, uint32(2))

	send_tools.Send2Account(send2acc.GetData(), acc.SessionId)
}

func (self *master) Old_MSGID_R2B_GAME_ENTER_GAME(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	_ = pack.ReadUInt32()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_R2B_GAME_ENTER_GAME.UInt16())
	if ret := self.canEnterRoom(accountId); ret > 0 {
		send.WriteUInt8(uint8(ret))
		send_tools.Send2Account(send.GetData(), session)
		return
	}
	self.enterRoom(accountId)

	// 通知客户端，进入游戏成功
	acc := account.AccountMgr.GetAccountByID(accountId)
	send2c := packet.NewPacket(nil)
	send2c.SetMsgID(protomsg.Old_MSGID_R2B_GAME_ENTER_GAME.UInt16())
	send2c.WriteUInt8(0)
	send2c.WriteUInt32(self.roomId)
	send_tools.Send2Account(send2c.GetData(), acc.SessionId)
	send2acc := self.sendGameData(acc, uint32(2))

	send_tools.Send2Account(send2acc.GetData(), acc.SessionId)
}
