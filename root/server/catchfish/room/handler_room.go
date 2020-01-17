package room

import (
	"root/core/packet"
	"root/server/catchfish/account"
)

func (self *Room) Old_MSGID_LEAVE_GAME(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	account.CheckSession(accountId, session)
	self.leaveRoom(accountId)

}
