package room

import (
	"root/common"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"fmt"
	"github.com/astaxie/beego"
	"root/protomsg"
	"root/server/mahjong-dgk/account"
	"root/server/mahjong-dgk/send_tools"
	"root/server/mahjong-dgk/types"
)

type (
	close struct {
		*Room
		s types.ERoomStatus
	}
)

func (self *close) Enter(now int64) {
	self.track_log(colorized.Yellow("Close enter %v\n"), self.roomId)
	send2hall := packet.NewPacket(nil)
	send2hall.SetMsgID(protomsg.Old_MSGID_UPDATE_DESTROY_ROOM.UInt16())
	send2hall.WriteUInt16(1)
	send2hall.WriteUInt32(uint32(self.roomId))
	send_tools.Send2Hall(send2hall.GetData())

	a := core.GetActor(int32(self.roomId))
	a.Suspend()

	core.LocalCoreSend(int32(self.roomId), common.EActorType_MAIN.Int32(), func() {
		delete(RoomMgr.roomActor, self.roomId)
		delete(RoomMgr.roomActorId, self.roomId)

		if len(RoomMgr.roomActor) == 0 {
			toHallMsg := packet.NewPacket(nil)
			toHallMsg.SetMsgID(protomsg.Old_MSGID_MAINTENANCE_NOTICE.UInt16())
			nGameType := uint8(beego.AppConfig.DefaultInt(fmt.Sprintf("%v", core.Appname)+"::gametype", 0))
			toHallMsg.WriteUInt8(uint8(nGameType))
			send_tools.Send2Hall(toHallMsg.GetData())

			log.Infof(colorized.White("所有房间退出完成！"))
		}
	})
}

func (self *close) Tick(now int64) {

}

func (self *close) CombineMSG(packet packet.IPacket, acc *account.Account) {

}
func (self *close) Leave(now int64) {

	self.track_log(colorized.Yellow("Close leave\n"))
}

///////////////////////////////// handler ///////////////////////////////////////////////////
func (self *close) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	default:
		log.Warnf("Close 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}

	return true
}
