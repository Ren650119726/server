package logic

import (
	"root/common"
	"root/core"
	"root/core/log"
	"root/core/packet"
	"root/core/utils"
	"fmt"
	"github.com/golang/protobuf/proto"
	"root/protomsg"
	"root/server/mahjong-dgk/account"
	"root/server/mahjong-dgk/room"
	"root/server/mahjong-dgk/send_tools"
)

func (self *DGK) Old_MSGID_CREATE_ROOM(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountID := pack.ReadUInt32()
	newRoomID := pack.ReadUInt32()
	gameType := pack.ReadUInt8()
	trParam := pack.ReadString()
	matchType := pack.ReadUInt8()
	clubID := pack.ReadUInt32()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_CREATE_ROOM_RET.UInt16())
	if room.RoomMgr.IsMaintenance == true {
		send.WriteUInt32(accountID)
		send.WriteUInt32(newRoomID)
		send.WriteUInt8(4)
		send_tools.Send2Hall(send.GetData())
		return
	}
	//////////////////////// test /////////////////////////////
	//if room.RoomMgr.Room_Count() == 1 {
	//	return
	//}

	room_actor := room.RoomMgr.ComposeRoom(accountID, gameType, newRoomID, trParam, matchType, clubID)
	child := core.NewActor(int32(newRoomID), room_actor, make(chan core.IMessage, 10000))
	core.CoreRegisteActor(child)

	actorId := room.RoomMgr.RoomActorId(newRoomID)
	if actorId != 0 {
		send.WriteUInt32(accountID)
		send.WriteUInt32(newRoomID)
		send.WriteUInt8(0)
		core.CoreSend(0, common.EActorType_CONNECT_HALL.Int32(), send.GetData(), 0)
		return
	}
}

func (self *DGK) Old_MSGID_RECV_ACCOUNT_INFO(actor int32, data []byte, session int64) {
	pack := packet.NewPacket(data)
	accId := pack.ReadUInt32()
	name := pack.ReadString()
	headURL := pack.ReadString()
	ip := pack.ReadString()
	rmb := pack.ReadUInt64()
	special := pack.ReadUInt32() // nSpecial
	robot := pack.ReadUInt8()    // robot
	signature := pack.ReadString()
	roomId := pack.ReadUInt32()
	safeRMB := pack.ReadUInt64()
	sd := pack.ReadInt8()

	if session != 0 {
		log.Warnf("Error, 异常session:%v 处理消息编号:%v", session, pack.GetMsgID())
		return
	}
	if room.RoomMgr.IsMaintenance == true {
		//    维护中, 不允许玩家进入
		return
	}

	storage := &protomsg.AccountStorageData{}
	game := &protomsg.AccountGameData{}
	game.RoomID = roomId
	game.Robot = uint32(robot)
	storage.AccountId = accId
	storage.HeadURL = headURL
	storage.Name = name
	storage.RMB = rmb
	storage.SafeRMB = safeRMB
	storage.Signature = signature
	storage.ActiveIP = ip
	storage.Special = uint32(special)

	account.AccountMgr.RecvAccount(storage, game, sd)

}

func (self *DGK) Old_MSGID_ENTER_GAME(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	roomId := pack.ReadUInt32()
	if room.RoomMgr.IsMaintenance == true {
		//    维护中, 不允许玩家进入
		send := packet.NewPacket(nil)
		send.SetMsgID(protomsg.Old_MSGID_ENTER_GAME.UInt16())
		send.WriteUInt8(4)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	b := account.AccountMgr.EnterAccount(accountId, roomId, session)

	if b {
		actorId := room.RoomMgr.RoomActorId(roomId)
		if actorId == 0 {
			log.Errorf("玩家连上断勾卡 但是找不到房间所在actor roomId:%v", roomId)
			return
		}

		core.CoreSend(0, actorId, msg, session)
	}
}

func (self *DGK) Old_MSGID_MAINTENANCE_NOTICE(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	if session != 0 {
		log.Warnf("Error, 异常session:%v 处理消息编号:%v", session, pack.GetMsgID())
		return
	}

	room.CloseServer(nil)
}
func (self *DGK) Old_MSGID_BACKSTAGE_CLOSE_ROOM(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	nRoomID := pack.ReadUInt32()

	if session != 0 {
		log.Warnf("Error, 异常session:%v 处理消息编号:%v", session, pack.GetMsgID())
		return
	}

	room := room.RoomMgr.Room(nRoomID)
	if room != nil {
		core.LocalCoreSend(0, int32(nRoomID), func() {
			room.Close()
			log.Infof("Backage CloseRoom, RoomID:%v", nRoomID)
		})
	}
}

func (self *DGK) Old_MSGID_BACKSTAGE_SET_BONUSPOOL(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	nBet := pack.ReadUInt32()
	nSetValue := pack.ReadUInt32()

	if session != 0 {
		log.Warnf("Error, 异常session:%v 处理消息编号:%v", session, pack.GetMsgID())
		return
	}

	nOldValue := room.RoomMgr.Bonus[nBet]
	room.RoomMgr.Set_bonus(nBet, uint64(nSetValue), false)
	nNewValue := room.RoomMgr.Bonus[nBet]

	log.Infof("后台设置奖金池, ServerID:%v, Bet:%v OldValue:%v, NewValue:%v", core.Appname, nBet, nOldValue, nNewValue)

	strLog := fmt.Sprintf("INSERT INTO log_bonus_pool (log_ServerID, log_Bet, log_OldValue, log_NewValue, log_Time) VALUES (%v, %v, %v, %v, '%v');", core.Appname, nBet, nOldValue, nNewValue, utils.DateString())
	tSendToHall := packet.NewPacket(nil)
	tSendToHall.SetMsgID(protomsg.MSGID_SAVE_LOG.UInt16())
	tSendToHall.WriteString(strLog)
	tSendToHall.WriteUInt16(2) // 日志类型, 大厅可将多条日志组装成一个消息回存
	send_tools.Send2Hall(tSendToHall.GetData())
}

func (self *DGK) Old_MSGID_GET_ONE_BONUSPOOL(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	bonus_pool_str := pack.ReadString()
	his_str := pack.ReadString()

	room.RoomMgr.InitBouns(bonus_pool_str, his_str)
}

func (self *DGK) Old_MSGID_CHANGE_PLAYER_INFO(actor int32, msg []byte, session int64) {
	tPack := packet.NewPacket(msg)
	nAccountID := tPack.ReadUInt32()
	nChangeType := tPack.ReadUInt8()
	strString := tPack.ReadString()

	if session != 0 {
		log.Warnf("Error, 异常session:%v 处理消息编号:%v", session, tPack.GetMsgID())
		return
	}

	acc := account.AccountMgr.GetAccountByID(nAccountID)
	if acc == nil {
		return
	}

	tRoom := room.RoomMgr.Room(acc.RoomID)
	if tRoom == nil {
		return
	}

	if nChangeType == 1 {
		acc.Name = strString
	} else if nChangeType == 2 {
		acc.HeadURL = strString
	}

	core.LocalCoreSend(self.owner.Id, int32(acc.RoomID), func() {
		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(protomsg.Old_MSGID_CHANGE_PLAYER_INFO.UInt16())
		tSend.WriteUInt32(nAccountID)
		tSend.WriteUInt8(nChangeType)
		tSend.WriteString(strString)
		tRoom.SendBroadcast(tSend.GetData())
	})
}

func (self *DGK) MSGID_HG_REENTER_OTHER_GAME(actor int32, msg []byte, session int64) {
	recv := packet.NewPacket(msg)
	info := &protomsg.HG_REENTER_OTHER{}
	proto.Unmarshal(recv.ReadBytes(), info)

	if session != 0 {
		log.Warnf("Error, 异常session:%v 处理消息编号:%v", session, recv.GetMsgID())
		return
	}

	core.CoreSend(self.owner.Id, int32(info.RoomId), msg, session)
}
