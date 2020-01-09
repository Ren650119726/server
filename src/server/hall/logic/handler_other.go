package logic

import (
	"root/common"
	"root/common/config"
	"root/common/tools"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"fmt"
	"github.com/golang/protobuf/proto"
	"root/protomsg"
	"root/server/hall/account"
	"root/server/hall/send_tools"
	"root/server/hall/server"
)


// game请求hall建立连接
func (self *Hall) Old_MSGID_SS_MAPING(actor int32, msg []byte, session int64) {

	pack := packet.NewPacket(msg)
	nServerID := pack.ReadUInt16()
	nHaveRoomCount := pack.ReadUInt32()
	strSign := pack.ReadString()

	GAME_TO_HALL_MAP_KEY := config.GetPublicConfig_String("GAME_TO_HALL_MAP_KEY")
	strCheckSign := fmt.Sprintf("%v%v%v", nServerID, nHaveRoomCount, GAME_TO_HALL_MAP_KEY)
	strCheckSign = tools.MD5(strCheckSign)
	if strCheckSign != strSign {
		log.Warnf("Error: 游戏与大厅建立连接时验签失败 not match; CheckSign:%v != Sign:%v; ServerID:%v RoomCount:%v", strCheckSign, strSign, nServerID, nHaveRoomCount)
		return
	}

	tCheck := account.AccountMgr.GetAccountBySessionID(session)
	if tCheck != nil {
		log.Infof("Error: Session:%v 游戏通知大厅已启动时错误, ServerID:%v, HaveRoom:%v", session, nServerID, nHaveRoomCount)
		return
	}

	HallMgr.mapServerNode(nHaveRoomCount, nServerID, session)
}

// game通知大厅设置维护标记
func (self *Hall) MSGID_GH_SET_MAINTENANCE(actor int32, msg []byte, session int64) {

	pack := packet.NewPacket(msg)
	nServerID := pack.ReadUInt16()
	isOpen := pack.ReadUInt8() == 1

	isGameServerSession := server.ServerMgr.IsGameServerSession(session)
	if isGameServerSession == false {
		log.Infof("Error: Session:%v 游戏通知大厅设置维护标记错误, ServerID:%v", session, nServerID)
		return
	}

	tServerNode := server.ServerMgr.GetServerNode(nServerID)
	if tServerNode != nil {
		tServerNode.IsMaintenance = isOpen
		log.Infof(" ====> 游戏:%v 设置维护标记:%v", common.EGameType(tServerNode.GameType), isOpen)
	}
}

// 处理游戏服同步货币
func (self *Hall) Old_MSGID_SYNC_TO_HALL_MONEY(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	nLastMoney := pack.ReadInt64()
	nLastSafeMoney := pack.ReadInt64()

	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		log.Errorf("can find acc :%v", accountId)
		return
	}

	isGameServerSession := server.ServerMgr.IsGameServerSession(session)
	if isGameServerSession == false {
		log.Infof("Error: Session:%v 游戏同步元宝到大厅错误, 玩家ID:%v 所在游戏:%v 最后RMB:%v 最后SafeRMB:%v", session, accountId, common.EGameType(acc.GameType), nLastMoney, nLastSafeMoney)
		return
	}

	acc.SyncMoney(uint64(nLastMoney), uint64(nLastSafeMoney))
}

// 处理游戏服同步货币
func (self *Hall) MSGID_SAVE_RMB_CHANGE_LOG(actor int32, msg []byte, session int64) {

	pack := packet.NewPacket(msg)
	nAccountID := pack.ReadUInt32()
	iChangeValue := pack.ReadInt64()
	nRMBMoney := pack.ReadInt64()
	nIndex := pack.ReadUInt8()
	nOperate := pack.ReadUInt8()
	strTime := pack.ReadString()
	nRoomID := pack.ReadUInt32()
	nGameType := pack.ReadUInt8()

	isGameServerSession := server.ServerMgr.IsGameServerSession(session)
	if isGameServerSession == false {
		log.Infof("Error: Session:%v 游戏通知大厅RMB变更日志错误, 玩家ID:%v 游戏:%v 房间:%v 改变值:%v 改变后:%v 操作原因:%v", session, nAccountID, common.EGameType(nGameType), nRoomID, iChangeValue, nRMBMoney, common.EOperateType(nOperate))
		return
	}

	if iChangeValue != 0 {
		strSQL := fmt.Sprintf("(%v, %v, %v, %v, %v, '%v', %v, %v),", nAccountID, iChangeValue, nRMBMoney, nIndex, nOperate, strTime, nRoomID, nGameType)
		account.AccountMgr.AddRMBChangeLog(nAccountID, strSQL)
	}
}

func (self *Hall) Old_MSGID_SEND_RANK_LIST(actor int32, msg []byte, session int64) {
	tPack := packet.NewPacket(msg)
	nAccountID := tPack.ReadUInt32()

	rank.RankMgr.SendRankToClient(nAccountID, session)
}

func (self *Hall) Old_MSGID_MAINTENANCE_NOTICE(actor int32, msg []byte, session int64) {
	tPack := packet.NewPacket(msg)
	nGameType := tPack.ReadUInt8()

	isGameServerSession := server.ServerMgr.IsGameServerSession(session)
	if isGameServerSession == false {
		log.Infof("Error: Session:%v 游戏通知大厅可关闭异常, 游戏:%v", session, common.EGameType(nGameType))
		return
	}

	nCanCloseCount := 0
	mServer := server.ServerMgr.GetAllServerList()
	for _, tNode := range mServer {
		if tNode.GameType == nGameType {
			tNode.CanClose = true
		}

		if tNode.CanClose == true {
			nCanCloseCount++
		}
	}

	nTotalCount := len(mServer)
	if nTotalCount == nCanCloseCount {
		log.Infof(colorized.Yellow("!!!!!!!!!!!!!!所有游戏都可关闭!!!!!!!!!!!!!!"))
		log.Infof(colorized.Yellow("!!!!!!!!!!!!!!所有游戏都可关闭!!!!!!!!!!!!!!"))
		log.Infof(colorized.Yellow("!!!!!!!!!!!!!!所有游戏都可关闭!!!!!!!!!!!!!!"))
	} else {
		for _, tNode := range mServer {
			if tNode.CanClose == false {
				log.Infof(colorized.Yellow("  ==> %v 未关闭, 待定游戏进程通知!"), common.EGameType(tNode.GameType))
			}
		}
		log.Infof(colorized.Yellow("还剩:%v个游戏未关闭"), nTotalCount-nCanCloseCount)
	}
}

func (self *Hall) MSGID_OPERATE_SAFE_BOX(actor int32, msg []byte, session int64) {
	tPack := packet.NewPacket(msg)
	nMsgID := tPack.GetMsgID()

	nAccountID := tPack.ReadUInt32()
	nOpType := tPack.ReadUInt8()
	iOpVal := tPack.ReadInt64()

	tSend := packet.NewPacket(nil)
	tSend.SetMsgID(protomsg.MSGID_OPERATE_SAFE_BOX.UInt16())

	var acc *account.Account
	if nMsgID == protomsg.MSGID_OPERATE_SAFE_BOX.UInt16() {
		acc = account.CheckSession(nAccountID, session)
	} else {
		isGameServerSession := server.ServerMgr.IsGameServerSession(session)
		if isGameServerSession == false {
			log.Infof("Error: Session:%v 游戏通知保险箱操作返回异常, AccID:%v 操作:%v 改变值:%v", session, nAccountID, nOpType, iOpVal)
			return
		}
		acc = account.AccountMgr.GetAccountByID(nAccountID)
	}
	if acc == nil {
		tSend.WriteUInt8(1)
		send_tools.Send2Account(tSend.GetData(), session)
		return
	}

	if acc.Robot > 0 {
		// 模拟机器人发消息 不处理
		strOp := ""
		if nOpType == 1 {
			strOp = "取出"
		} else {
			strOp = "存入"
		}
		log.Infof("Error: 玩家ID:%v 违规操作保险箱, 自己是机器人, 消息:%v %v 金额:%v", acc.AccountId, nMsgID, strOp, iOpVal)
		return
	}

	if iOpVal == 0 {
		tSend.WriteUInt8(2)
		send_tools.Send2Account(tSend.GetData(), session)
		return
	}
	nRet := uint8(0)
	if nMsgID == protomsg.MSGID_OPERATE_SAFE_BOX_RESPOND.UInt16() {
		nRet = acc.OperateSafeBox(nOpType, iOpVal, false)
	} else {
		nRet = acc.OperateSafeBox(nOpType, iOpVal, true)
	}

	if nRet == 1 {
		// 转发到游戏去处理操作保险箱
		nGameSessionID := HallMgr.getGameSessionID(acc.RoomID)
		if nGameSessionID > 0 {
			tSend.SetMsgID(protomsg.MSGID_CHANGE_SAFE_RMB.UInt16())
			tSend.WriteUInt32(nAccountID)
			tSend.WriteUInt8(nOpType)
			tSend.WriteInt64(iOpVal)
			send_tools.Send2Game(tSend.GetData(), nGameSessionID)
		} else {
			// 游戏进程异常, 稍后重试
			tSend.WriteUInt8(4)
			send_tools.Send2Account(tSend.GetData(), session)
		}
	} else if nRet == 0 {
		//操作成功
		tSend.WriteUInt8(0)
		tSend.WriteUInt8(nOpType)
		tSend.WriteInt64(int64(acc.RMB))
		tSend.WriteInt64(int64(acc.SafeRMB))
		send_tools.Send2Account(tSend.GetData(), session)
	} else {
		// 操作失败
		tSend.WriteUInt8(nRet)
		send_tools.Send2Account(tSend.GetData(), session)
	}
}
func (self *Hall) MSGID_CH_SELECT_MATCH_RESULT(actor int32, msg []byte, session int64) {
	tPack := packet.NewPacket(msg)
	matchid := tPack.ReadUInt32()
	nAccountID := tPack.ReadUInt32()
	result := tPack.ReadInt8()

	MatchList.DealWithResult(nAccountID, matchid, result)
}
func (self *Hall) MSGID_HG_REENTER_OTHER_GAME(actor int32, msg []byte, session int64) {
	tPack := packet.NewPacket(msg)
	info := &protomsg.HG_REENTER_OTHER{}
	proto.Unmarshal(tPack.ReadBytes(), info)

	backmsg := packet.NewPacket(nil)
	backmsg.SetMsgID(protomsg.Old_MSGID_ENTER_ROOM.UInt16())
	backmsg.WriteUInt32(info.AccountId)
	backmsg.WriteUInt32(info.EntRoomId)
	backmsg.WriteInt8(1)
	core.CoreSend(0, common.EActorType_MAIN.Int32(), backmsg.GetData(), 0)

}
