package room

import (
	"root/common"
	"root/common/config"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/red2black/account"
	"root/server/red2black/event"
	"root/server/red2black/send_tools"
)

type (
	betting struct {
		*Room
		sendMark1   []func(packet.IPacket)
		sendMark2   []func(packet.IPacket)
		s           ERoomStatus
		timestamp   int64
		updateTimer int64
		bet_count   int

		bets map[uint32]uint32 // key accid val betval
	}
)

func (self *betting) Enter(now int64) {
	self.total_bet_player_val = 0
	self.sendMark1 = make([]func(packet.IPacket), 0)
	self.sendMark2 = make([]func(packet.IPacket), 0)
	self.bets = make(map[uint32]uint32)

	self.updateTimer = self.owner.AddTimer(utils.MILLISECONDS_OF_SECOND*0.1, -1, self.updateBetting)
	duration := self.status_duration[int(self.s)] // 持续时间 秒
	self.timestamp = now + int64(duration)
	log.Debugf(colorized.Yellow("betting enter duration:%v"), duration)
	// 广播房间玩家，切换状态
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_R2B_NEXT_STATE.UInt16())
	send.WriteUInt8(uint8(ERoomStatus_START_BETTING))
	send.WriteUInt32(uint32(duration * 1000))
	self.SendBroadcast(send.GetData())

	event.Dispatcher.Dispatch(&event.EnterBetting{
		RoomID:   self.roomId,
		Robots:   self.Robots(),
		Seats:    self.seats,
		Duration: int64(duration),
	}, event.EventType_EnterBetting)
	self.bet_count = 0

}

func (self *betting) Tick(now int64) {
	if now >= self.timestamp {
		self.switchStatus(now, ERoomStatus_STOP_BETTING)
		return
	}
}

func (self *betting) Leave(now int64) {
	self.owner.CancelTimer(self.updateTimer)
	log.Debugf(colorized.Gray("self.bet_count:%v"), self.bet_count)
	log.Debugf(colorized.Yellow("betting leave\n"))
}

func (self *betting) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_ENTER_GAME.UInt16(): // 客户端链接进入游戏
		self.Old_MSGID_R2B_ENTER_GAME(actor, msg, session)
	case protomsg.Old_MSGID_R2B_GAME_ENTER_GAME.UInt16(): // 客户端链接进入游戏
		self.Old_MSGID_R2B_GAME_ENTER_GAME(actor, msg, session)
	case protomsg.Old_MSGID_R2B_BETTING.UInt16():
		self.Old_MSGID_R2B_BETTING(actor, msg, session)
	default:
		log.Warnf("betting 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}

	return true
}

func (self *betting) Old_MSGID_R2B_ENTER_GAME(actor int32, msg []byte, session int64) {
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

	now := utils.SecondTimeSince1970()
	// 通知客户端，进入游戏成功
	acc := account.AccountMgr.GetAccountByID(accountId)
	send2c := packet.NewPacket(nil)
	send2c.SetMsgID(protomsg.Old_MSGID_ENTER_GAME.UInt16())
	send2c.WriteUInt8(0)
	send2c.WriteUInt32(self.roomId)
	send_tools.Send2Account(send2c.GetData(), acc.SessionId)
	send2acc := self.sendGameData(acc, uint32(self.timestamp-now))

	send_tools.Send2Account(send2acc.GetData(), acc.SessionId)

	if acc.GetMoney() < uint64(config.GetPublicConfig_Int64("R2B_LIMIT_VAL")) {
		acc.IsAllowBetting = false
	} else {
		acc.IsAllowBetting = true
	}
}

func (self *betting) Old_MSGID_R2B_GAME_ENTER_GAME(actor int32, msg []byte, session int64) {
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

	now := utils.SecondTimeSince1970()
	// 通知客户端，进入游戏成功
	acc := account.AccountMgr.GetAccountByID(accountId)
	send2c := packet.NewPacket(nil)
	send2c.SetMsgID(protomsg.Old_MSGID_R2B_GAME_ENTER_GAME.UInt16())
	send2c.WriteUInt8(0)
	send2c.WriteUInt32(self.roomId)
	send_tools.Send2Account(send2c.GetData(), acc.SessionId)
	send2acc := self.sendGameData(acc, uint32(self.timestamp-now))

	send_tools.Send2Account(send2acc.GetData(), acc.SessionId)

	if acc.GetMoney() < uint64(config.GetPublicConfig_Int64("R2B_LIMIT_VAL")) {
		acc.IsAllowBetting = false
	} else {
		acc.IsAllowBetting = true
	}
}

func (self *betting) Old_MSGID_R2B_BETTING(actor int32, msg []byte, session int64) {
	if self.total_master_val() == 0 {
		log.Warnf("没有庄家,不能下注")
		return
	}
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	index := pack.ReadUInt8() // 押注区域
	betValue := pack.ReadUInt32()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_R2B_BETTING.UInt16())

	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	if index <= 0 || index > 3 {
		send.WriteUInt8(11)
		send_tools.Send2Account(send.GetData(), session)
		log.Debugf("index error:%v", index)
		return
	}

	if self.accounts[acc.AccountId] == nil {
		send.WriteUInt8(12)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	// 庄家不能下注 r2bnew
	if self.SeatMasterIndex(accountId) != -1 {
		send.WriteUInt8(13)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	if !acc.IsAllowBetting {
		send.WriteUInt8(20)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	if acc.GetMoney() < uint64(betValue) {
		send.WriteUInt8(14)
		send_tools.Send2Account(send.GetData(), session)
		log.Debugf("acc.GetMoney():%v < :%v", acc.GetMoney(), betValue)
		return
	}

	bets := self.total_bet()
	bets[index] += betValue
	total_val := self.total_master_val()
	total_money := int(total_val * config.GetPublicConfig_Int64("R2B_DOMINATE_MONEY"))
	if index == 3 && bets[3] > uint32(total_money/20) {
		send.WriteUInt8(15)
		send_tools.Send2Account(send.GetData(), session)
		log.Warnf("当前押注 特牌:%v 区域:%v 押注金额已经操过可押注的总值 %v total_val:%v", betValue, index, self.total_bet(), total_val)
		return
	}

	choushui := uint32(bets[2])

	if index == 1 {
		choushui = bets[2] * uint32(100-config.GetPublicConfig_Int64("R2B_SYSTEM_FEE")) / uint32(100)
	} else if index == 2 {
		choushui = bets[1] * uint32(100-config.GetPublicConfig_Int64("R2B_SYSTEM_FEE")) / uint32(100)
	} else if index == 3 {
		if bets[1] < bets[2] {
			choushui = bets[1] * uint32(100-config.GetPublicConfig_Int64("R2B_SYSTEM_FEE")) / uint32(100)
		} else {
			choushui = bets[2] * uint32(100-config.GetPublicConfig_Int64("R2B_SYSTEM_FEE")) / uint32(100)
		}
	}
	// 检查是否操过总上限 r2bnew
	if !self.check_bet(bets, total_val, int(choushui)) {
		send.WriteUInt8(15)
		send_tools.Send2Account(send.GetData(), session)
		log.Warnf("当前押注:%v 区域:%v 押注金额已经操过可押注的总值 %v total_val:%v", betValue, index, self.total_bet(), total_val)
		return
	}

	if acc.Robot == 0 {
		self.total_bet_player_val += int64(betValue)
	}
	acc.AddMoney(-int64(betValue), index, common.EOperateType_BETTING)
	acc.BetVal[index] += betValue
	self.bets[acc.AccountId] = acc.BetVal[index]

	send.WriteUInt8(0)
	send.WriteUInt8(1) //1为单播, 客户端处理自己的筹码飞和金币改变   --2为广播
	send.WriteUInt32(acc.AccountId)
	send.WriteUInt8(index)
	send.WriteUInt32(betValue)
	send.WriteUInt64(acc.GetMoney())
	send_tools.Send2Account(send.GetData(), acc.SessionId)

	cacheFun1 := func(broadcast packet.IPacket) {
		broadcast.WriteUInt32(accountId)
		broadcast.WriteUInt8(index)
		broadcast.WriteUInt32(betValue)
	}
	self.sendMark1 = append(self.sendMark1, cacheFun1)

	cacheFun2 := func(broadcast packet.IPacket) {
		broadcast.WriteUInt32(accountId)
		broadcast.WriteInt64(int64(acc.GetMoney()))
	}
	self.sendMark2 = append(self.sendMark2, cacheFun2)
	self.bet_count++
}

func (self *betting) updateBetting(dt int64) {
	if len(self.sendMark1) <= 0 {
		return
	}

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_R2B_BETTING.UInt16())
	send.WriteUInt8(0)
	send.WriteUInt8(2)
	send.WriteUInt16(uint16(len(self.sendMark1)))
	for _, f := range self.sendMark1 {
		f(send)
	}

	send.WriteUInt16(uint16(len(self.sendMark2)))
	for _, f := range self.sendMark2 {
		f(send)
	}

	self.sendMark1 = self.sendMark1[0:0]
	self.sendMark2 = self.sendMark2[0:0]
	self.SendBroadcast(send.GetData())
}
