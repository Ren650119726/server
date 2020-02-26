package logic

import (
	"root/common"
	"root/core"
	"root/core/log"
	"root/core/packet"
	"root/protomsg"
	"root/server/red2black/account"
	"root/server/red2black/room"
)

func (self *Red2Black) Old_MSGID_CREATE_ROOM(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountID := pack.ReadUInt32()
	newRoomID := pack.ReadUInt32()
	gameType := pack.ReadUInt8()
	trParam := pack.ReadString()
	matchType := pack.ReadUInt8()
	clubID := pack.ReadUInt32()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_CREATE_ROOM_RET.UInt16())

	if room.RoomMgr.RoomCount() > 0 {
		send.WriteUInt32(accountID)
		send.WriteUInt32(newRoomID)
		send.WriteUInt8(11)
		core.CoreSend(0, common.EActorType_CONNECT_HALL.Int32(), send.GetData(), 0)
		return
	}
	room_actor := room.RoomMgr.CreateRoom(accountID, gameType, newRoomID, trParam, matchType, clubID)
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

func (self *Red2Black) Old_MSGID_RECV_ACCOUNT_INFO(actor int32, data []byte, session int64) {
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

	if session != 0 {
		log.Warnf("Error, 异常session:%v 处理消息编号:%v", session, pack.GetMsgID())
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

	account.AccountMgr.RecvAccount(storage, game)
}

func (self *Red2Black) Old_MSGID_R2B_ENTER_GAME(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	roomId := pack.ReadUInt32()
	b := account.AccountMgr.EnterAccount(accountId, roomId, session)

	if b {
		actorId := room.RoomMgr.RoomActorId(roomId)
		if actorId == 0 {
			log.Errorf("玩家连上R2b 但是找不到房间所在actor roomId:%v", roomId)
			return
		}

		core.CoreSend(0, actorId, msg, session)
	}
}

func (self *Red2Black) Old_MSGID_CHANGE_PLAYER_INFO(actor int32, msg []byte, session int64) {
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

	if room.RoomMgr.Global_room == nil || acc.RoomID == 0 {
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
		room.RoomMgr.Global_room.SendBroadcast(tSend.GetData())
	})
}
