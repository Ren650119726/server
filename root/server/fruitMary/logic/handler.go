package logic

import (
	"root/core/log"
	"root/core/packet"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/fruitMary/account"
	"root/server/fruitMary/send_tools"
)

// 大厅发送的玩家数据
func (self *FruitMary) SERVERMSG_HG_PLAYER_DATA_REQ(actor int32, msg []byte, session int64) {
	accPB := packet.PBUnmarshal(msg,&inner.PLAYER_DATA_REQ{}).(*inner.PLAYER_DATA_REQ)
	if acc := account.AccountMgr.GetAccountByID(accPB.GetAccount().GetAccountId());acc != nil{
		acc.AccountStorageData = accPB.GetAccount()
		acc.AccountGameData = accPB.GetAccountData()
	}else {
		account.AccountMgr.SetAccountByID(account.NewAccount(accPB.GetAccount()))
	}
	
	send_tools.Send2Hall(inner.SERVERMSG_GH_PLAYER_DATA_RES.UInt16(),&inner.PLAYER_DATA_RES{
		Ret:       0,
		AccountID: accPB.GetAccount().GetAccountId(),
		RoomID:    accPB.GetRoomID(),
	})
	log.Infof("大厅发送玩家数据:%v 想进入的房间:%v",accPB.GetAccount().GetAccountId(),accPB.GetRoomID())
}

// 玩家请求进入小玛利房间
func (self *FruitMary) FRUITMARYMSG_CS_ENTER_GAME_FRUITMARY_REQ(actor int32, data []byte, session int64) int32{
	enterPB := packet.PBUnmarshal(data,&protomsg.ENTER_GAME_FRUITMARY_REQ{}).(*protomsg.ENTER_GAME_FRUITMARY_REQ)
	acc := account.AccountMgr.GetAccountByIDAssert(enterPB.GetAccountID())
	acc.SessionId = session
	account.AccountMgr.SetAccountBySession(acc,session)

	actorId := int32(enterPB.GetRoomID())
	if actorId == 0 {
		log.Warnf("玩家连上小玛利 但是找不到房间所在actor roomId:%v", enterPB.GetRoomID())
		return 0
	}

	acc.RoomID = enterPB.GetRoomID()
	return int32(enterPB.GetRoomID())
}