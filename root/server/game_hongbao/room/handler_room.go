package room

import (
	"root/common"
	"root/core/log"
	"root/core/packet"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/game_hongbao/account"
	"root/server/game_hongbao/send_tools"
)

// 玩家进入游戏
func (self *Room) HBMSG_CS_ENTER_GAME_HB_REQ(actor int32, msg []byte, session int64) {
	enterPB := packet.PBUnmarshal(msg, &protomsg.ENTER_GAME_HB_REQ{}).(*protomsg.ENTER_GAME_HB_REQ)
	self.enterRoom(enterPB.GetAccountID())
}

// 玩家离开
func (self *Room) HBMSG_CS_LEAVE_GAME_HB_REQ(actor int32, msg []byte, session int64) {
	enterPB := packet.PBUnmarshal(msg, &protomsg.LEAVE_GAME_HB_REQ{}).(*protomsg.LEAVE_GAME_HB_REQ)
	ret := uint32(1)
	if self.canleave(enterPB.GetAccountID()) {
		ret = 0
	}
	send_tools.Send2Account(protomsg.HBMSG_SC_LEAVE_GAME_HB_RES.UInt16(), &protomsg.LEAVE_GAME_HB_RES{
		Ret:    ret,
		RoomID: self.roomId,
	}, session)
}

// 玩家请求发红包
func (self *Room) HBMSG_CS_ASSIGN_HB_REQ(actor int32, msg []byte, session int64) {
	assignHB := packet.PBUnmarshal(msg, &protomsg.ASSIGN_HB_REQ{}).(*protomsg.ASSIGN_HB_REQ)
	account.AccountMgr.GetAccountByIDAssert(assignHB.GetAccountID())

}

// 请求抢红包
func (self *Room) HBMSG_CS_GRAB_HB_REQ(actor int32, msg []byte, session int64) {
	//start := packet.PBUnmarshal(msg, &protomsg.START_HB_REQ{}).(*protomsg.START_HB_REQ)
	//msgBetNum := start.GetBet()
	//acc := account.AccountMgr.GetAccountBySessionIDAssert(session)

}

func (self *Room) HBMSG_CS_PLAYERS_HB_LIST_REQ(actor int32, msg []byte, session int64) {
	account.AccountMgr.GetAccountBySessionIDAssert(session)

	ret := &protomsg.PLAYERS_HB_LIST_RES{}
	ret.Players = make([]*protomsg.AccountStorageData, 0)
	for _, p := range self.accounts {
		ret.Players = append(ret.Players, p.AccountStorageData)
	}
	send_tools.Send2Account(protomsg.HBMSG_SC_PLAYERS_HB_LIST_RES.UInt16(), ret, session)
}

// 大厅请求修改玩家数据
func (self *Room) SERVERMSG_HG_NOTIFY_ALTER_DATE(actor int32, msg []byte, session int64) {
	if session != 0 {
		log.Warnf("此消息只能大厅发送 %v", session)
		return
	}
	data := packet.PBUnmarshal(msg, &inner.NOTIFY_ALTER_DATE{}).(*inner.NOTIFY_ALTER_DATE)
	acc := account.AccountMgr.GetAccountByIDAssert(data.GetAccountID())
	if data.GetType() == 1 { // 修改金币
		changeValue := int(data.GetAlterValue())
		if changeValue < 0 && -changeValue > int(acc.GetMoney()) {
			changeValue = int(-acc.GetMoney())
		}
		acc.AddMoney(int64(changeValue), common.EOperateType(data.GetOperateType()))
	} else if data.GetType() == 2 { // 修改杀数
		acc.Kill = int32(data.GetAlterValue())
	}
}
