package logic

import (
	"root/core"
	"root/core/log"
	"root/core/packet"
	"root/server/catchfish/account"
	"root/server/catchfish/room"
)

func (self *CatchFish) Old_MSGID_CREATE_ROOM(actor int32, msg []byte, session int64) {

}

func (self *CatchFish) Old_MSGID_RECV_ACCOUNT_INFO(actor int32, data []byte, session int64) {
	//account.AccountMgr.RecvAccount(storage, game)
}

func (self *CatchFish) Old_MSGID_ENTER_GAME(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	roomId := pack.ReadUInt32()
	if room.RoomMgr.IsMaintenance == true {
		return
	}

	b := account.AccountMgr.EnterAccount(accountId, roomId, session)

	if b {
		actorId := int32(roomId)
		if actorId == 0 {
			log.Warnf("玩家连上R2b 但是找不到房间所在actor roomId:%v", roomId)
			return
		}

		core.CoreSend(0, actorId, msg, session)
	}
}