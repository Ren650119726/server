package logic

import (
	"fmt"
	"root/common"
	"root/common/config"
	"root/core/log"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/hall/account"
	"root/server/hall/send_tools"
	"root/server/hall/server"
)

// db返回的水位线
func (self *Hall) SERVERMSG_DH_ALL_WATER_LINE(actor int32, msg []byte, session int64) {
	all_water_line := packet.PBUnmarshal(msg,&inner.ALL_WATER_LINE{}).(*inner.ALL_WATER_LINE)
	for _, v := range all_water_line.Line {
		serverID := uint16(v.ServerID)
		if _, exist := HallMgr.mWaterLine[serverID]; !exist {
			HallMgr.mWaterLine[serverID] = &waterLine{
				GameType:  uint8(v.GameType),
				WaterLine: v.WaterLine,
			}
		}
	}
}

func (self *Hall) MSGID_GET_ONE_WATER_LINE(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	nServerID := pack.ReadUInt16()

	tNode := HallMgr.mWaterLine[nServerID]
	if tNode != nil {
		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(protomsg.MSGID_GET_ONE_WATERLINE.UInt16())
		tSend.WriteString(tNode.WaterLine)
		send_tools.Send2Game(tSend.GetData(), session)
	}
}

func (self *Hall) MSGID_SET_ONE_WATER_LINE(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	nServerID := pack.ReadUInt16()
	nGameType := pack.ReadUInt8()
	strWaterLine := pack.ReadString()

	isGameServerSession := server.ServerMgr.IsGameServerSession(session)
	if isGameServerSession == false {
		log.Infof("Error: Session:%v 游戏通知大厅设置水位线错误, ServerID:%v 游戏:%v 水位线:%v", session, nServerID, common.EGameType(nGameType), strWaterLine)
		return
	}

	tNode := HallMgr.mWaterLine[nServerID]
	if tNode != nil {
		tNode.GameType = nGameType
		tNode.WaterLine = strWaterLine
	} else {
		HallMgr.mWaterLine[nServerID] = &waterLine{
			GameType:  nGameType,
			WaterLine: strWaterLine,
		}
	}

	nNowTime := utils.SecondTimeSince1970()
	if nNextTime, isExist := HallMgr.mSaveWaterLineTime[nServerID]; isExist == true {
		if nNowTime >= nNextTime {
			tLine := &protomsg.WaterLine{ServerID: uint32(nServerID), GameType: uint32(tNode.GameType), WaterLine: tNode.WaterLine}
			send_tools.Send2DB(account.AccountMgr.StampNum(), protomsg.MSGID_HG_SAVE_WATER_LINE.UInt16(), &protomsg.SAVE_WATER_LINE{Line: tLine}, false)
			log.Infof("回存水位线数据: %+v ", tLine)

			SAVE_WATER_LINE_TIME := config.GetPublicConfig_Int64("SAVE_WATER_LINE_TIME")
			HallMgr.mSaveWaterLineTime[nServerID] = nNowTime + SAVE_WATER_LINE_TIME
		}
	} else {
		SAVE_WATER_LINE_TIME := config.GetPublicConfig_Int64("SAVE_WATER_LINE_TIME")
		HallMgr.mSaveWaterLineTime[nServerID] = nNowTime + SAVE_WATER_LINE_TIME
	}
}

func (self *Hall) Old_MSGID_GET_ROOM_LIST(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	nGameType := pack.ReadUInt8()
	nClubID := pack.ReadUInt32()

	HallMgr.sendRoomList(nGameType, nClubID, session)
}

func (self *Hall) Old_MSGID_OPEN_DESK_UPDATE(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	nAccountID := pack.ReadUInt32()
	nGameType := pack.ReadUInt8()
	nMatchType := pack.ReadUInt8()
	nClubID := pack.ReadUInt32()

	tAccount := account.CheckSession(nAccountID, session)
	nRet := HallMgr.openDeskUpdate(nGameType, nMatchType, tAccount, nClubID)
	if nRet == 0 {
		HallMgr.checkPreCreateRoomCount(nGameType, nMatchType, tAccount, nClubID)
	}

	tSend := packet.NewPacket(nil)
	tSend.SetMsgID(protomsg.Old_MSGID_OPEN_DESK_UPDATE.UInt16())
	tSend.WriteUInt8(nRet)
	send_tools.Send2Account(tSend.GetData(), session)

}

func (self *Hall) Old_MSGID_CLOSE_DESK_UPDATE(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	nAccountID := pack.ReadUInt32()
	nGameType := pack.ReadUInt8()
	nMatchType := pack.ReadUInt8()
	nClubID := pack.ReadUInt32()

	tAccount := account.CheckSession(nAccountID, session)
	nRet := HallMgr.closeDeskUpdate(nGameType, nMatchType, tAccount, nClubID)

	tSend := packet.NewPacket(nil)
	tSend.SetMsgID(protomsg.Old_MSGID_CLOSE_DESK_UPDATE.UInt16())
	tSend.WriteUInt8(nRet)
	send_tools.Send2Account(tSend.GetData(), session)

}

func (self *Hall) Old_MSGID_CREATE_ROOM(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	nAccountID := pack.ReadUInt32()
	nGameType := pack.ReadUInt8()
	strParam := pack.ReadString()
	nClubID := pack.ReadUInt32()

	if HallMgr.OpenDesk == 0 {
		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(protomsg.Old_MSGID_ENTER_ROOM.UInt16())
		tSend.WriteUInt8(30)
		send_tools.Send2Account(tSend.GetData(), session)
		return
	}

	tAccount := account.CheckSession(nAccountID, session)
	if tAccount == nil {
		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(protomsg.Old_MSGID_ENTER_ROOM.UInt16())
		tSend.WriteUInt8(1)
		send_tools.Send2Account(tSend.GetData(), session)
		return
	}

	nMatchType := uint8(0)
	if nGameType == common.EGameTypePAO_DE_KUAI.Value() || nGameType == common.EGameTypeXMMJ.Value() || nGameType == common.EGameTypeDGK.Value() {
		sParam := utils.SplitConf2ArrInt32(strParam, "|")
		if len(sParam) >= 4 {
			// 房间档次按照房间参数中人数来确定; 3人档次3  4人档次4
			nMatchType = uint8(sParam[3])
		} else {
			log.Warnf("客户端创建 %v 房间时, 读取参数错误; 参数:%v", common.EGameType(nGameType), strParam)
		}
	} else if nGameType == common.EGameTypeNIU_NIU.Value() || nGameType == common.EGameTypeWUHUA_NIUNIU.Value() {
		nMatchType = uint8(config.NN_MAX_PLAYER)
	} else if nGameType == common.EGameTypeTEN_NIU_NIU.Value() {
		nMatchType = uint8(config.TNN_MAX_PLAYER)
	} else if nGameType == common.EGameTypeSAN_GONG.Value() {
		nMatchType = uint8(config.SG_MAX_PLAYER)
	} else if nGameType == common.EGameTypePDK_HN.Value() {
		nMatchType = uint8(config.PDK_HN_MAX_PLAYER)
	}

	if nClubID != 0 && !ClubMgr.IsOpenGame(nClubID, nGameType) {
		log.Warnf("玩家:%v 请求创建俱乐部:%v %v 游戏 失败! 俱乐部未开放此游戏", tAccount.AccountId, nClubID, common.EGameType(nGameType).String())
		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(protomsg.Old_MSGID_CREATE_ROOM.UInt16())
		tSend.WriteUInt8(7)
		send_tools.Send2Account(tSend.GetData(), session)
		return
	}

	nRet := HallMgr.createRoom(tAccount, nGameType, nMatchType, strParam, protomsg.Old_MSGID_CREATE_ROOM.UInt16(), nClubID)
	if nRet > 0 {
		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(protomsg.Old_MSGID_CREATE_ROOM.UInt16())
		tSend.WriteUInt8(nRet)
		send_tools.Send2Account(tSend.GetData(), session)
		return
	}
}

func (self *Hall) Old_MSGID_CREATE_ROOM_RESULT(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	nAccountID := pack.ReadUInt32()
	nNewRoomID := pack.ReadUInt32()
	nResult := pack.ReadUInt8()

	isGameServerSession := server.ServerMgr.IsGameServerSession(session)
	if isGameServerSession == false {
		log.Infof("Error: Session:%v 游戏通知大厅创建房间结果错误, 玩家ID:%v 新房间:%v", session, nAccountID, nNewRoomID)
		return
	}

	tCreate := HallMgr.mCreateTable[nNewRoomID]
	delete(HallMgr.mCreateTable, nNewRoomID)
	if tCreate == nil {
		log.Warnf("ResponseCreateRoom Can't find CreateNode:%v, NewRoomID:%v, Result:%v", nAccountID, nNewRoomID, nResult)
		return
	}

	isSystemCreate := tCreate.nAnswerProtocol == protomsg.Old_MSGID_SYSTEM_CREATE_ROOM.UInt16()
	tAccount := account.AccountMgr.GetAccountByID(nAccountID)
	if tAccount == nil && isSystemCreate == false {
		log.Warnf("ResponseCreateRoom Can't find AccountID:%v, NewRoomID:%v, Result:%v", nAccountID, nNewRoomID, nResult)
		return
	}

	if tAccount != nil && tAccount.RoomID > 0 && isSystemCreate == false {
		if tAccount.Robot == 0 {
			tSend := packet.NewPacket(nil)
			tSend.SetMsgID(tCreate.nAnswerProtocol)
			tSend.WriteUInt8(2)
			send_tools.Send2Account(tSend.GetData(), tAccount.SessionId)
		}
		return
	}

	if HallMgr.nMaintenanceTime > 0 {
		if tAccount != nil && tAccount.Robot == 0 && isSystemCreate == false {
			tSend := packet.NewPacket(nil)
			tSend.SetMsgID(tCreate.nAnswerProtocol)
			tSend.WriteUInt8(4)
			send_tools.Send2Account(tSend.GetData(), tAccount.SessionId)
		}
		return
	}

	tServerNode := server.ServerMgr.GetServerNode(tCreate.nServerID)
	if tServerNode == nil {
		if tAccount != nil && tAccount.Robot == 0 && isSystemCreate == false {
			tSend := packet.NewPacket(nil)
			tSend.SetMsgID(tCreate.nAnswerProtocol)
			tSend.WriteUInt8(4)
			send_tools.Send2Account(tSend.GetData(), tAccount.SessionId)
		}
		return
	}

	if tAccount != nil && tAccount.Robot == 0 {
		nCheckBindCode := config.GetPublicConfig_Int64("CHECK_BIND_CODE")
		if nCheckBindCode == 1 && tAccount.BindCode <= 0 {
			if tAccount != nil && isSystemCreate == false {
				tSend := packet.NewPacket(nil)
				tSend.SetMsgID(tCreate.nAnswerProtocol)
				tSend.WriteUInt8(5)
				send_tools.Send2Account(tSend.GetData(), tAccount.SessionId)
			}
			return
		}
	}

	if nResult > 0 {
		if tAccount != nil && tAccount.Robot == 0 && isSystemCreate == false {
			tSend := packet.NewPacket(nil)
			tSend.SetMsgID(tCreate.nAnswerProtocol)
			tSend.WriteUInt8(nResult)
			send_tools.Send2Account(tSend.GetData(), tAccount.SessionId)
		}
		return
	}

	tNewRoom := HallMgr.newRoom(tAccount, tCreate.nGameType, tCreate.nNewRoomID, tCreate.nServerID, tCreate.nMatchType, tCreate.strParam, tCreate.nAnswerProtocol, tCreate.nClubID, tCreate.nClubmgr)
	if tAccount != nil && tNewRoom != nil {
		if isSystemCreate == false {
			tAccount.SendAccountInfo2Game(tCreate.nNewRoomID, tServerNode.SessionID, 0)
			if tAccount.Robot > 0 {
				account.AccountMgr.SendRobotEnterGame(tCreate.nGameType, tAccount.AccountId, tCreate.nNewRoomID, tServerNode.SessionID)
			}
		}
	}

	if tNewRoom != nil {
		HallMgr.assignmentGameNodeIP(tNewRoom)
		HallMgr.updateDeskList(tNewRoom, 4)
	}
}

func (self *Hall) Old_MSGID_ENTER_ROOM(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	nAccountID := pack.ReadUInt32()
	nRoomID := pack.ReadUInt32()
	auto_sitdown := pack.ReadInt8()

	if HallMgr.OpenDesk == 0 {
		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(protomsg.Old_MSGID_ENTER_ROOM.UInt16())
		tSend.WriteUInt8(30)
		send_tools.Send2Account(tSend.GetData(), session)
		return
	}

	if HallMgr.nMaintenanceTime > 0 {
		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(protomsg.Old_MSGID_ENTER_ROOM.UInt16())
		tSend.WriteUInt8(4)
		send_tools.Send2Account(tSend.GetData(), session)
		return
	}

	tAccount := account.AccountMgr.GetAccountByID(nAccountID)
	if tAccount == nil {
		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(protomsg.Old_MSGID_ENTER_ROOM.UInt16())
		tSend.WriteUInt8(1)
		send_tools.Send2Account(tSend.GetData(), session)
		return
	}

	nCheckBindCode := config.GetPublicConfig_Int64("CHECK_BIND_CODE")
	if nCheckBindCode == 1 && tAccount.BindCode <= 0 && tAccount.Robot == 0 {
		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(protomsg.Old_MSGID_ENTER_ROOM.UInt16())
		tSend.WriteUInt8(5)
		send_tools.Send2Account(tSend.GetData(), session)
		return
	}

	if tAccount.RoomID > 0 && tAccount.RoomID != nRoomID {
		isInRoom := HallMgr.IsInRoom(tAccount.RoomID, tAccount.AccountId)
		if isInRoom == false {
			tAccount.RoomID = 0
			tAccount.GameType = 0
			tAccount.Index = 0
			tAccount.MatchType = 0
		} else {
			tSend := packet.NewPacket(nil)
			tSend.SetMsgID(protomsg.Old_MSGID_ENTER_ROOM.UInt16())
			tSend.WriteUInt8(2)
			send_tools.Send2Account(tSend.GetData(), session)
			return
		}
	}

	tRoom := HallMgr.GetRoom(nRoomID)
	if tRoom == nil {
		// 容错处理
		if tAccount.RoomID == nRoomID {
			tAccount.RoomID = 0
			tAccount.GameType = 0
			tAccount.Index = 0
			tAccount.MatchType = 0
		}
		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(protomsg.Old_MSGID_ENTER_ROOM.UInt16())
		tSend.WriteUInt8(3)
		send_tools.Send2Account(tSend.GetData(), session)
		return
	}

	tServerNode := server.ServerMgr.GetServerNode(tRoom.nServerID)
	if tServerNode == nil || tServerNode.IsMaintenance == true {
		// 容错处理
		if tAccount.RoomID == nRoomID {
			tAccount.RoomID = 0
			tAccount.GameType = 0
			tAccount.Index = 0
			tAccount.MatchType = 0
		}
		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(protomsg.Old_MSGID_ENTER_ROOM.UInt16())
		tSend.WriteUInt8(4)
		send_tools.Send2Account(tSend.GetData(), session)
		return
	}

	nRet := HallMgr.canEnterRoom(tAccount, tRoom)
	if nRet > 0 {
		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(protomsg.Old_MSGID_ENTER_ROOM.UInt16())
		tSend.WriteUInt8(nRet)
		send_tools.Send2Account(tSend.GetData(), session)
		return
	}
	log.Debugf("session;%v ", tServerNode.SessionID)
	tAccount.SendAccountInfo2Game(nRoomID, tServerNode.SessionID, auto_sitdown)
}

func (self *Hall) Old_MSGID_UPDATE_SERVICE_FEE(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	nGameType := pack.ReadUInt8()
	nRoomID := pack.ReadUInt32()
	nPlayerCount := pack.ReadUInt16()
	strNowTime := utils.DateString()

	isGameServerSession := server.ServerMgr.IsGameServerSession(session)
	if isGameServerSession == false {
		log.Infof("Error: Session:%v 游戏同步服务费到大厅错误, 所在游戏:%v 所在房间:%v", session, common.EGameType(nGameType), nRoomID)
		return
	}

	nClubID := uint32(0)
	tRoom := HallMgr.GetRoom(nRoomID)
	if tRoom != nil {
		nClubID = tRoom.clubID
	}

	for i := 0; i < int(nPlayerCount); i++ {
		nAccountID := pack.ReadUInt32()
		nServiceFee := pack.ReadUInt32()
		tAccount := account.AccountMgr.GetAccountByID(nAccountID)
		if tAccount != nil && tAccount.Robot == 0 && nServiceFee > 0 {
			strSQL := fmt.Sprintf("(%v,%v,%v,'%v',%v,%v),", nAccountID, nServiceFee, nGameType, strNowTime, nRoomID, nClubID)
			HallMgr.AddServiceFeeLog(nAccountID, strSQL)
		}
	}
}

func (self *Hall) Old_MSGID_UPDATE_ACCOUNT(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	nRoomID := pack.ReadUInt32() // 未使用字段 nRoomID
	_ = pack.ReadUInt8()         // 未使用字段 nModeType
	nPlayerCount := pack.ReadUInt16()

	isGameServerSession := server.ServerMgr.IsGameServerSession(session)
	if isGameServerSession == false {
		log.Infof("Error: Session:%v 游戏同步玩家数据到大厅错误, 所在房间:%v", session, nRoomID)
		return
	}

	//tRoom := HallMgr.GetRoom(nRoomID)
	//if tRoom == nil {
	//	log.Warnf("UPDATE_ACCOUNT Can't find RoomID:%v", nRoomID)
	//	return
	//}
	//var maxChange int64
	//var maxCard string
	//var maxName string
	nNowTime := utils.MilliSecondTimeSince1970()
	GAME_RECORD_COUNT := config.GetPublicConfig_Int64("GAME_RECORD_COUNT")
	for i := 0; i < int(nPlayerCount); i++ {
		nAccountID := pack.ReadUInt32()
		nRMB := pack.ReadInt64()
		iChange := pack.ReadInt64()
		_ = pack.ReadString() // strCard

		tAccount := account.AccountMgr.GetAccountByID(nAccountID)
		if tAccount != nil {
			// 记录元宝值
			tAccount.RMB = uint64(nRMB)
			tAccount.IsChangeData = true

			// 添加战绩记录
			if tAccount.Robot == 0 {
				tRecord := &account.GameRecord{GameType: uint8(tAccount.GameType), Change: iChange, Time: nNowTime}
				tAccount.GameRecord = append(tAccount.GameRecord, tRecord)
				if int64(len(tAccount.GameRecord)) > GAME_RECORD_COUNT {
					tAccount.GameRecord = tAccount.GameRecord[1:]
				}
			}

			//// 找出当局最大赢家
			//if iChange > maxChange {
			//	maxChange = iChange
			//	maxCard = strCard
			//	maxName = tAccount.Name
			//}
		}
	}

	//// 判断是否需要播放小喇叭
	//GAME_BROADCAST_NOTIFY := config.GetPublicConfig_Mapi("GAME_BROADCAST_NOTIFY")
	//NEED_NOTIFY_VALUE, isExist := GAME_BROADCAST_NOTIFY[int(tRoom.nGameType)]
	//if isExist == true && maxChange >= int64(NEED_NOTIFY_VALUE) {
	//	var combinationOfString string
	//	if maxCard == "" {
	//		strText := config.GetPublicConfig_String("PROFIT_BROADCAST_NOTIFY")
	//		fRMB := float32(maxChange) / float32(config.RMB_BILI)
	//		combinationOfString = fmt.Sprintf(strText, maxName, common.EGameType(tRoom.nGameType).String(), fRMB)
	//	} else {
	//		strText := config.GetPublicConfig_String("PROFIT_BROADCAST_NOTIFY_CARD")
	//		fRMB := float32(maxChange) / float32(config.RMB_BILI)
	//		combinationOfString = fmt.Sprintf(strText, maxName, common.EGameType(tRoom.nGameType).String(), maxCard, fRMB)
	//	}
	//	speaker.SpeakerMgr.sendBroadcast(2, combinationOfString)
	//}
}

func (self *Hall) Old_MSGID_UPDATE_ENTER(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	nAccountID := pack.ReadUInt32()
	nRoomID := pack.ReadUInt32()
	nCount := pack.ReadUInt16()
	nWatch := pack.ReadUInt8()
	nIndex := pack.ReadUInt8()

	isGameServerSession := server.ServerMgr.IsGameServerSession(session)
	if isGameServerSession == false {
		log.Infof("Error: Session:%v 游戏通知大厅玩家进入错误, 玩家ID:%v 房间:%v", session, nAccountID, nRoomID)
		return
	}

	tRoom := HallMgr.GetRoom(nRoomID)
	if tRoom == nil {
		log.Warnf("UPDATE_ENTER Can't find RoomID:%v", nRoomID)
		return
	}

	tAccount := account.AccountMgr.GetAccountByID(nAccountID)
	if tAccount == nil {
		log.Warnf("UPDATE_ENTER Can't find AccountID:%v", nAccountID)
		nAccountID = 0
	} else {
		tAccount.RoomID = nRoomID
		tAccount.GameType = uint32(tRoom.nGameType)
		tAccount.MatchType = tRoom.nMatchType
		tAccount.Index = uint32(nIndex)

		if tAccount.Robot > 0 {
			tAccount.LoginTime = utils.SecondTimeSince1970() // 机器人进入房间，设置为在线
		}

		//if tAccount.Robot == 0 {
		//  log.Infof("UPDATE_ENTER, AccountID:%v, RoomID:%v, GameType:%v, Count:%v, Watch:%v", tAccount.AccountId, tAccount.RoomID, tAccount.GameType, nCount, nWatch)
		//}
	}

	if nWatch == 0 {
		if _, isExist := tRoom.mSeatID[nAccountID]; isExist == false {
			tRoom.mSeatID[nAccountID] = true
			HallMgr.updateDeskList(tRoom, 2)
			HallMgr.checkPreCreateRoomCount(tRoom.nGameType, tRoom.nMatchType, tAccount, tRoom.clubID)
		} else {
			log.Warnf("Error UPDATE_ENTER SitDown, Game:%v AccountID:%v RoomID:%v Count:%v Watch:%v Index:%v", common.EGameType(tRoom.nGameType), nAccountID, nRoomID, nCount, nWatch, nIndex)
		}
	} else {
		if _, isExist := tRoom.mWatchID[nAccountID]; isExist == false {
			tRoom.mWatchID[nAccountID] = true
		} else {
			log.Warnf("Error UPDATE_ENTER Watch, Game:%v AccountID:%v RoomID:%v Count:%v Watch:%v Index:%v", common.EGameType(tRoom.nGameType), nAccountID, nRoomID, nCount, nWatch, nIndex)
		}
	}

}
func (self *Hall) Old_MSGID_UPDATE_INDEX(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	nAccountID := pack.ReadUInt32()
	nIndex := pack.ReadUInt8()

	isGameServerSession := server.ServerMgr.IsGameServerSession(session)
	if isGameServerSession == false {
		log.Infof("Error: Session:%v 游戏通知大厅玩家更新座位索引错误, 玩家ID:%v 座位:%v", session, nAccountID, nIndex)
		return
	}

	tAccount := account.AccountMgr.GetAccountByID(nAccountID)
	if tAccount != nil {
		tAccount.Index = uint32(nIndex)

		tRoom := HallMgr.GetRoom(tAccount.RoomID)
		if tRoom != nil {
			HallMgr.updateDeskList(tRoom, 2)
		}

		//if tAccount.Robot == 0 {
		//	log.Infof("UPDATE_INDEX, AccountID:%v, Set Index:%v", tAccount.AccountId, nIndex)
		//}
	}
}

func (self *Hall) Old_MSGID_UPDATE_LEAVE(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	nAccountID := pack.ReadUInt32()
	nRoomID := pack.ReadUInt32()
	nCount := pack.ReadUInt16()
	nWatch := pack.ReadUInt8()

	isGameServerSession := server.ServerMgr.IsGameServerSession(session)
	if isGameServerSession == false {
		log.Infof("Error: Session:%v 游戏通知大厅玩家离开错误, 玩家ID:%v 房间:%v", session, nAccountID, nRoomID)
		return
	}

	tAccount := account.AccountMgr.GetAccountByID(nAccountID)
	if tAccount == nil {
		log.Warnf("Old_MSGID_UPDATE_LEAVE Can't find AccountID:%v", nAccountID)
	} else {
		//if tAccount.Robot == 0 {
		//  log.Infof("UPDATE_LEAVE, AccountID:%v, RoomID:%v, GameType:%v, Count:%v, Watch:%v", tAccount.AccountId, tAccount.RoomID, tAccount.GameType, nCount, nWatch)
		//}
		tAccount.RoomID = 0
		tAccount.GameType = 0
		tAccount.MatchType = 0
		tAccount.Index = 0
		if tAccount.Robot > 0 {
			tAccount.IsUse = false
			tAccount.RMB = 0
			tAccount.LogoutTime = utils.SecondTimeSince1970() // 机器人退出房间，设置为在线
		}
	}

	tRoom := HallMgr.GetRoom(nRoomID)
	if tRoom == nil {
		log.Warnf("UPDATE_LEAVE Can't find RoomID:%v", nRoomID)
	} else {
		if nWatch == 0 {
			if _, isExist := tRoom.mSeatID[nAccountID]; isExist == true {
				delete(tRoom.mSeatID, nAccountID)
				HallMgr.updateDeskList(tRoom, 2)
			} else {
				log.Warnf("Error Old_MSGID_UPDATE_LEAVE SitDown, Game:%v AccountID:%v RoomID:%v Count:%v Watch:%v", common.EGameType(tRoom.nGameType), nAccountID, nRoomID, nCount, nWatch)
			}
		} else {

			if _, isExist := tRoom.mWatchID[nAccountID]; isExist == true {
				delete(tRoom.mWatchID, nAccountID)
			} else {
				log.Warnf("Error Old_MSGID_UPDATE_LEAVE Watch, Game:%v AccountID:%v RoomID:%v Count:%v Watch:%v", common.EGameType(tRoom.nGameType), nAccountID, nRoomID, nCount, nWatch)
			}
		}
	}
}

func (self *Hall) Old_MSGID_UPDATE_DESTROY_ROOM(actor int32, msg []byte, session int64) {
	isGameServerSession := server.ServerMgr.IsGameServerSession(session)
	if isGameServerSession == false {
		log.Infof("Error: Session:%v 游戏通知大厅销毁房间错误", session)
		return
	}

	pack := packet.NewPacket(msg)
	nRoomCount := pack.ReadUInt16()
	for i := 0; i < int(nRoomCount); i++ {
		nRoomID := pack.ReadUInt32()
		tRoom := HallMgr.GetRoom(nRoomID)
		if tRoom != nil {
			HallMgr.destroyRoom(tRoom)
			log.Infof("UPDATE_DESTROY_ROOM DestroyRoom, ServerID:%v Game:%v RoomID:%v", tRoom.nServerID, common.EGameType(tRoom.nGameType), tRoom.nRoomID)
		}
	}
}

func (self *Hall) MSGID_SAVE_LOG(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	strLog := pack.ReadString()

	isGameServerSession := server.ServerMgr.IsGameServerSession(session)
	if isGameServerSession == false {
		log.Infof("Error: Session:%v 游戏通知大厅回存日志错误, 日志:%v", session, strLog)
		return
	}
	// 日志类型; 可用于处理同一类型缓存后一条消息回存到DB
	//nLogType := pack.ReadUInt16()

	send_tools.SQLLog(account.AccountMgr.StampNum(), strLog)
}