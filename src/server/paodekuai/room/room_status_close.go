package room

import (
	"root/common"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/protomsg"
	"root/server/paodekuai/account"
	"root/server/paodekuai/send_tools"
)

type (
	close struct {
		*Room
	}
)

func (self *close) BulidPacket(tPacket packet.IPacket, tAccount *account.Account) {
}

func (self *close) Enter(now int64) {
	self.track_log(colorized.Yellow("RoomID:%v close enter"), self.roomId)
	send2hall := packet.NewPacket(nil)
	send2hall.SetMsgID(protomsg.Old_MSGID_UPDATE_DESTROY_ROOM.UInt16())
	send2hall.WriteUInt16(1)
	send2hall.WriteUInt32(uint32(self.roomId))
	send_tools.Send2Hall(send2hall.GetData())

	a := core.GetActor(int32(self.roomId))
	a.Suspend()

	core.LocalCoreSend(int32(self.roomId), common.EActorType_MAIN.Int32(), func() {
		delete(RoomMgr.RoomActor, self.roomId)
		delete(RoomMgr.roomActorId, self.roomId)

		if len(RoomMgr.RoomActor) == 0 {
			toHallMsg := packet.NewPacket(nil)
			toHallMsg.SetMsgID(protomsg.Old_MSGID_MAINTENANCE_NOTICE.UInt16())
			toHallMsg.WriteUInt8(common.EGameTypePAO_DE_KUAI.Value())
			send_tools.Send2Hall(toHallMsg.GetData())
			log.Infof(colorized.White("所有房间退出完成！"))
		}
	})
}

func (self *close) Tick(now int64) {

}

func (self *close) Leave(now int64) {
	self.track_log(colorized.Yellow("RoomID:%v close leave\n"), self.roomId)
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
