package logic

import (
	"root/core/log"
	"root/core/packet"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/hall/account"
	"root/server/hall/event"
	"root/server/hall/send_tools"
)

// 玩家向大厅请求进入房间
func (self *Hall) MSG_CS_ENTER_ROOM_REQ(actor int32, msg []byte, session int64) {
	pbMsg := packet.PBUnmarshal(msg, &protomsg.ENTER_ROOM_REQ{}).(*protomsg.ENTER_ROOM_REQ)
	acc := account.AccountMgr.GetAccountBySessionIDAssert(session)
	if acc.GetRoomID() != 0 && acc.GetRoomID() != pbMsg.RoomID {
		log.Warnf("玩家:%v已经在房间:[%v]内，不能进入新房间:[%v]", acc.GetAccountId(), acc.GetRoomID(), pbMsg.GetRoomID())
		return
	}

	// 给游戏服发送玩家数据
	room := GameMgr.rooms[pbMsg.GetRoomID()]
	if room == nil {
		log.Warnf("找不到房间:%v", pbMsg.GetRoomID())
		send_tools.Send2Account(protomsg.MSG_SC_ENTER_ROOM_RES.UInt16(), &protomsg.ENTER_ROOM_RES{Ret: 2, RoomID: pbMsg.GetRoomID()}, session)
		return
	}
	node := GameMgr.nodes[room.serverID]
	if node == nil {
		log.Warnf("找不到服务器节点 accID:%v roomID:%v, serverID:%v ", acc.GetAccountId(), pbMsg.GetRoomID(), room.serverID)
		return
	}
	sendPB := &inner.PLAYER_DATA_REQ{
		Account:     acc.AccountStorageData,
		AccountData: acc.AccountGameData,
		RoomID:      pbMsg.GetRoomID(),
		Reback:      true,
	}
	send_tools.Send2Game(inner.SERVERMSG_HG_PLAYER_DATA_REQ.UInt16(), sendPB, node.session)
	log.Infof("玩家:[%v] 请求进入房间:%v 给游戏:%v 发送数据 ", acc.GetAccountId(), pbMsg.GetRoomID(), room.serverID)
}

// 游戏通知大厅，可以让玩家进入房间
func (self *Hall) SERVERMSG_GH_PLAYER_DATA_RES(actor int32, msg []byte, session int64) {
	pbMsg := packet.PBUnmarshal(msg, &inner.PLAYER_DATA_RES{}).(*inner.PLAYER_DATA_RES)
	acc := account.AccountMgr.GetAccountByIDAssert(pbMsg.GetAccountID())
	if acc.GetRoomID() != 0 && acc.GetRoomID() != pbMsg.RoomID {
		log.Warnf("玩家:%v已经在房间:[%v]内，不能进入新房间:[%v]", acc.GetAccountId(), acc.GetRoomID(), pbMsg.GetRoomID())
		return
	}
	roominfo := GameMgr.rooms[pbMsg.RoomID]
	if roominfo == nil {
		log.Warnf("找不到房间:%v", roominfo)
		return
	}

	node := GameMgr.nodes[roominfo.serverID]
	if node == nil {
		log.Warnf("找不到房间所在节点:%v %v ", roominfo, roominfo.serverID)
		return
	}
	send_tools.Send2Account(protomsg.MSG_SC_ENTER_ROOM_RES.UInt16(), &protomsg.ENTER_ROOM_RES{Ret: 0, RoomID: pbMsg.GetRoomID(), GameType: node.gameType}, acc.SessionId)
	log.Infof("通知玩家:%v 进入房间:%v 游戏:%v ", acc.GetAccountId(), pbMsg.GetRoomID(), node.gameType)
}

// 游戏通知大厅，玩家进入房间
func (self *Hall) SERVERMSG_GH_PLAYER_ENTER_ROOM(actor int32, msg []byte, session int64) {
	pbMsg := packet.PBUnmarshal(msg, &inner.PLAYER_ENTER_ROOM{}).(*inner.PLAYER_ENTER_ROOM)
	acc := account.AccountMgr.GetAccountByIDAssert(pbMsg.GetAccountID())
	acc.RoomID = pbMsg.GetRoomID()
	room := GameMgr.rooms[acc.RoomID]
	if room == nil {
		log.Warnf("找不到大厅房间信息:%v ", acc.RoomID)
		return
	}
	room.PlayerCount = pbMsg.GetPlayerCount()
	room.RobotCount = pbMsg.GetRobotCount()
	event.Dispatcher.Dispatch(event.RoomUpdate{
		RoomID:      room.roomID,
		PlayerCount: room.PlayerCount,
		RobotCount:  room.RobotCount,
	}, event.EventType_RoomUpdate)

	log.Infof("玩家%v robot:%v 进入房间:%v 当前玩家数量:%v 当前机器人数量:%v", acc.GetAccountId(), acc.GetRobot(), pbMsg.GetRoomID(), pbMsg.GetPlayerCount(), pbMsg.GetRobotCount())
}

// 游戏通知大厅，玩家退出房间
func (self *Hall) SERVERMSG_GH_PLAYER_LEAVE_ROOM(actor int32, msg []byte, session int64) {
	pbMsg := packet.PBUnmarshal(msg, &inner.PLAYER_LEAVE_ROOM{}).(*inner.PLAYER_LEAVE_ROOM)
	acc := account.AccountMgr.GetAccountByIDAssert(pbMsg.GetAccountID())
	if acc.GetRoomID() == 0 {
		log.Warnf("房间:%v 通知大厅玩家:%v 退出房间，但是玩家不在任何房间内！！！需要排查异常情况", pbMsg.GetRoomID(), acc.GetAccountId())
		return
	}
	log.Infof("玩家%v robot:%v 离开房间:%v 当前玩家数量:%v 当前机器人数量:%v", acc.GetAccountId(), acc.GetRobot(), pbMsg.GetRoomID(), pbMsg.GetPlayerCount(), pbMsg.GetRobotCount())
	room := GameMgr.rooms[acc.RoomID]
	if room == nil {
		log.Warnf("找不到大厅房间信息:%v ", acc.RoomID)
		return
	}
	event.Dispatcher.Dispatch(event.RoomUpdate{
		RoomID:      room.roomID,
		PlayerCount: room.PlayerCount,
		RobotCount:  room.RobotCount,
	}, event.EventType_RoomUpdate)
	room.PlayerCount = pbMsg.GetPlayerCount()
	room.RobotCount = pbMsg.GetRobotCount()
	acc.RoomID = 0
}
