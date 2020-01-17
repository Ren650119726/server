package room

import (
	"root/core/packet"
	"root/protomsg"
)

// 玩家进入游戏
func (self *Room) FRUITMARYMSG_CS_ENTER_GAME_FRUITMARY_REQ(actor int32, msg []byte, session int64) {
	enterPB := packet.PBUnmarshal(msg,&protomsg.ENTER_GAME_FRUITMARY_REQ{}).(*protomsg.ENTER_GAME_FRUITMARY_REQ)
	self.enterRoom(enterPB.GetAccountID())
}

// 玩家进入游戏
func (self *Room) FRUITMARYMSG_CS_LEAVE_GAME_FRUITMARY_REQ(actor int32, msg []byte, session int64) {
	enterPB := packet.PBUnmarshal(msg,&protomsg.LEAVE_GAME_FRUITMARY_REQ{}).(*protomsg.LEAVE_GAME_FRUITMARY_REQ)
	self.leaveRoom(enterPB.GetAccountID())
}
