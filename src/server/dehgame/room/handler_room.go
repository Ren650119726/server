package room

import (
	"root/common"
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"encoding/json"
	"fmt"
	"root/protomsg"
	"root/server/dehgame/account"
	"root/server/dehgame/send_tools"
	"root/server/dehgame/types"
)

func (self *Room) Old_MSGID_CX_LEAVE_GAME(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	t := pack.ReadUInt32() // ????

	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		send := packet.NewPacket(nil)
		send.SetMsgID(protomsg.Old_MSGID_LEAVE_GAME.UInt16())
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	if t > 3 {
		send := packet.NewPacket(nil)
		send.SetMsgID(protomsg.Old_MSGID_LEAVE_GAME.UInt16())
		send.WriteUInt8(2)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	if self.seatIndex(accountId) != -1 && !self.status_obj().CanQuit(accountId) {
		send := packet.NewPacket(nil)
		send.SetMsgID(protomsg.Old_MSGID_LEAVE_GAME.UInt16())
		send.WriteUInt8(3)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	//send.WriteUInt8(0)

	// 统一走断线流程
	self.Disconnect(session)

	index := self.seatIndex(acc.AccountId)
	if index != -1 {
		if self.leaveRoom(acc.AccountId, true) { // 主动退出
			core.LocalCoreSend(self.owner.Id, common.EActorType_MAIN.Int32(), func() {
				account.AccountMgr.DisconnectAccount(session)
			})
		}
	}

}

func (self *Room) canQuit(accountId uint32) bool {
	index := self.seatIndex(accountId)
	if index == -1 {
		return true
	}
	player := self.seats[index]
	if player.status == types.EGameStatus_GIVE_UP {
		return true
	}
	return false
}
func (self *Room) Old_MSGID_CX_PLAYER_LIST(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_CX_PLAYER_LIST.UInt16())

	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	if self.accounts[accountId] == nil {
		send.WriteUInt8(11)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	send.WriteUInt8(0)
	temp := packet.NewPacket(nil)
	tempcount := uint16(0)
	for id, v := range self.accounts {
		index := self.seatIndex(v.AccountId)
		if index == -1 {
			tempcount++
			temp.WriteUInt32(id)
			temp.WriteString(v.Name)
			temp.WriteString(fmt.Sprintf("%v", v.HeadURL))
			temp.WriteInt64(int64(v.GetMoney()))
			temp.WriteString(v.Signature)
		}
	}
	send.WriteUInt16(uint16(tempcount))
	send.CatBody(temp)
	send_tools.Send2Account(send.GetData(), session)
}

func (self *Room) Old_MSGID_SEND_EMOJI(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	sendID := pack.ReadUInt32()
	targetID := pack.ReadUInt32()
	emojiType := pack.ReadUInt8()
	emojiID := pack.ReadUInt8()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_SEND_EMOJI.UInt16())
	send.WriteUInt8(0)
	send.WriteUInt32(sendID)
	send.WriteUInt32(targetID)
	send.WriteUInt8(emojiType)
	send.WriteUInt8(emojiID)
	self.SendBroadcast(send.GetData())
}

func (self *Room) Old_MSGID_SEND_TEXT_SHORTCUTS(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountID := pack.ReadUInt32()
	textID := pack.ReadUInt8()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_SEND_TEXT_SHORTCUTS.UInt16())
	acc := account.AccountMgr.GetAccountByID(accountID)
	if acc == nil {
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	send.WriteUInt8(0)
	send.WriteUInt32(accountID)
	send.WriteUInt8(textID)
	self.SendBroadcast(send.GetData())
}

// 请求坐下
func (self *Room) Old_MSGID_CX_SIT_DOWN(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountID := pack.ReadUInt32()

	account.CheckSession(accountID, session)
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_CX_SIT_DOWN.UInt16())
	acc := account.AccountMgr.GetAccountByID(accountID)
	if acc == nil {
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	acc = self.accounts[accountID]
	if acc == nil {
		send.WriteUInt8(2)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	// 如果总资产不够
	money := acc.GetMoney()
	need := uint64(self.GetParamInt(4)) // 请求坐下
	if money < need {
		send.WriteUInt8(4)
		send_tools.Send2Account(send.GetData(), session)
		log.Warnf("!钱不够设置簸簸 :%v!", accountID)
		return
	}

	if index := self.seatIndex(accountID); index != -1 {
		send.WriteUInt8(4)
		send_tools.Send2Account(send.GetData(), session)
		log.Warnf("!已经在座位上 :%v 座位号:%v!", accountID, index)
		return
	}

	index := self.sitDown(accountID)
	if index == -1 {
		send.WriteUInt8(3)
		send_tools.Send2Account(send.GetData(), session)
		return
	} else {
		send.WriteUInt8(0)
		send.WriteUInt8(uint8(index + 1))
		send_tools.Send2Account(send.GetData(), session)
	}

	addplayer := packet.NewPacket(nil)
	addplayer.SetMsgID(protomsg.Old_MSGID_CX_PLAYER_JOIN.UInt16())
	addplayer.WriteUInt8(uint8(index + 1))
	addplayer.WriteUInt32(accountID)
	addplayer.WriteString(acc.Name)
	addplayer.WriteString(acc.HeadURL)
	addplayer.WriteInt64(int64(acc.GetMoney()))
	addplayer.WriteString(acc.Signature)
	addplayer.WriteUInt8(acc.IsOnline())
	for _, bacc := range self.accounts {
		if bacc.Robot == 0 && bacc.SessionId > 0 && acc.SessionId != bacc.SessionId {
			send_tools.Send2Account(addplayer.GetData(), bacc.SessionId)
		}
	}

	self.broadcast_count()

	send2acc := self.sendGameData(acc)
	self.status_obj().CombineMSG(send2acc, acc)
	send_tools.Send2Account(send2acc.GetData(), acc.SessionId)

	send2leave := packet.NewPacket(nil)
	send2leave.SetMsgID(protomsg.Old_MSGID_UPDATE_LEAVE.UInt16())
	send2leave.WriteUInt32(acc.AccountId)
	send2leave.WriteUInt32(self.roomId)
	send2leave.WriteUInt16(uint16(self.sitDownCount()))
	send2leave.WriteUInt8(1)
	send_tools.Send2Hall(send2leave.GetData())

	// 2 hall
	send2enter := packet.NewPacket(nil)
	send2enter.SetMsgID(protomsg.Old_MSGID_UPDATE_ENTER.UInt16())
	send2enter.WriteUInt32(acc.AccountId)
	send2enter.WriteUInt32(self.roomId)
	send2enter.WriteUInt16(uint16(self.sitDownCount()))
	send2enter.WriteUInt8(uint8(0))
	send2enter.WriteUInt8(uint8(index + 1))
	send_tools.Send2Hall(send2enter.GetData())
	log.Infof(colorized.Yellow("玩家:[%v], 座位号:[%v], 身上余额:[%v] 坐下"), acc.AccountId, index, acc.GetMoney())
}

func (self *Room) Old_MSGID_CX_ENTER_GAME(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	_ = pack.ReadUInt32()

	send2c := packet.NewPacket(nil)
	send2c.SetMsgID(protomsg.Old_MSGID_ENTER_GAME.UInt16())

	if ret := self.canEnterRoom(accountId); ret > 0 {
		send2c.WriteUInt8(uint8(ret))
		send_tools.Send2Account(send2c.GetData(), session)
		return
	}

	if _, exist := self.accounts[accountId]; !exist {
		// 2 hall
		send2hall := packet.NewPacket(nil)
		send2hall.SetMsgID(protomsg.Old_MSGID_UPDATE_ENTER.UInt16())
		send2hall.WriteUInt32(accountId)
		send2hall.WriteUInt32(self.roomId)
		send2hall.WriteUInt16(uint16(self.sitDownCount()))
		send2hall.WriteUInt8(uint8(1))
		send2hall.WriteUInt8(uint8(0))
		send_tools.Send2Hall(send2hall.GetData())
	}

	self.enterRoom(accountId)

	// 通知客户端，进入游戏成功
	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		log.Error("找不到玩家:%v", acc.AccountId)
		return
	}
	send2c.WriteUInt8(0)
	send2c.WriteUInt32(self.roomId)
	send_tools.Send2Account(send2c.GetData(), acc.SessionId)
	send2acc := self.sendGameData(acc)

	self.status_obj().CombineMSG(send2acc, acc)
	send_tools.Send2Account(send2acc.GetData(), acc.SessionId)

	// 通知其他人
	if index := self.seatIndex(accountId); index != -1 {
		addplayer := packet.NewPacket(nil)
		addplayer.SetMsgID(protomsg.Old_MSGID_CX_PLAYER_JOIN.UInt16())
		addplayer.WriteUInt8(uint8(index + 1))
		addplayer.WriteUInt32(accountId)
		addplayer.WriteString(acc.Name)
		addplayer.WriteString(acc.HeadURL)
		addplayer.WriteInt64(int64(acc.GetMoney()))
		addplayer.WriteString(acc.Signature)
		addplayer.WriteUInt8(acc.IsOnline())
		for _, bacc := range self.accounts {
			if bacc.Robot == 0 && bacc.SessionId > 0 && acc.SessionId != bacc.SessionId {
				send_tools.Send2Account(addplayer.GetData(), bacc.SessionId)
			}
		}
	}

}

// 历史中奖记录
func (self *Room) Old_MSGID_CX_AWARD_HISTORY(actor int32, msg []byte, session int64) {
	log.Debugf("客户端请求历史中奖记录")
	RoomMgr.Bonus_h.RLock()
	bonush := RoomMgr.Bonus_h.m[uint32(self.GetParamInt(0))]
	RoomMgr.Bonus_h.RUnlock()
	if bonush == nil {
		return
	}
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_CX_AWARD_HISTORY.UInt16())
	//data, _ := json.Marshal(RoomMgr.History_max_info)
	var str string
	if bonush == nil || bonush.History_max_info == nil {
		str = string("{}")
	} else {
		data, _ := json.Marshal(bonush.History_max_info)
		str = string(data)
	}

	send.WriteUInt8(0)
	send.WriteString(str)

	paste_pack := packet.NewPacket(nil)
	count := uint16(0)
	for i := len(bonush.Award_history) - 1; i >= 0; i-- {
		count++
		data, _ := json.Marshal(bonush.Award_history[i])
		str := string(data)
		paste_pack.WriteString(str)
	}
	send.WriteUInt16(count)
	send.CatBody(paste_pack)

	send_tools.Send2Account(send.GetData(), session)
}

// 请求盈利
func (self *Room) Old_MSGID_CX_PROFIT_VAL(actor int32, msg []byte, session int64) {
	recv := packet.NewPacket(msg)
	acc := recv.ReadUInt32()

	index := self.seatIndex(acc)
	if index == -1 {
		send := packet.NewPacket(nil)
		send.SetMsgID(protomsg.Old_MSGID_CX_PROFIT_VAL.UInt16())
		send.WriteUInt32(uint32(acc))
		send.WriteInt64(0)
		send.WriteUInt32(uint32(0))
		send.WriteUInt32(uint32(0))
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	player := self.seats[index]

	penalty := player.profit * config.GetPublicConfig_Int64("QUIT_PENALTY")
	max_count := config.GetPublicConfig_Int64("DEH_MAX_QUIT_COUNT")
	if penalty < 0 || self.sitDownCount() == 1 || player.acc.Games >= int32(max_count) {
		penalty = 0
	}

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_CX_PROFIT_VAL.UInt16())
	send.WriteUInt32(uint32(player.acc.AccountId))
	send.WriteInt64(player.profit + player.extractDec - player.extractBoun)
	send.WriteUInt32(uint32(player.acc.Games))
	send.WriteUInt32(uint32(penalty))
	send_tools.Send2Account(send.GetData(), session)
	log.Debugf("请求盈利 退出 accID:%v index:%v profit:%v games:%v penalty:%v max_count:%v", player.acc.AccountId, index, player.profit, player.acc.Games, penalty, max_count)
}

// 请求战绩
func (self *Room) Old_MSGID_CX_SHOW_PERSON_INFO(actor int32, msg []byte, session int64) {

	send := packet.NewPacket(nil)
	tempCount := uint16(0)
	temp := packet.NewPacket(nil)
	for index, player := range self.seats {
		if player != nil {
			tempCount++
			temp.WriteUInt32(player.acc.AccountId)
			temp.WriteUInt8(uint8(index + 1))
			temp.WriteString(player.acc.Name)
			temp.WriteString(player.acc.HeadURL)
			temp.WriteInt64(int64(player.acc.GetMoney()))
			temp.WriteInt64(int64(player.bobo))
			temp.WriteUInt32(uint32(player.acc.Games))
			temp.WriteInt64(player.profit + player.extractDec - player.extractBoun)
		}
	}
	send.WriteUInt16(tempCount)
	send.CatBody(temp)

	send.SetMsgID(protomsg.Old_MSGID_CX_SHOW_PERSON_INFO.UInt16())
	send_tools.Send2Account(send.GetData(), session)
}
