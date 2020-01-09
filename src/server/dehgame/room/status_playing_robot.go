package room

import (
	"root/common"
	"root/core"
	"root/core/log"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/dehgame/event"
	"root/server/dehgame/types"
)

func (self *playing) OnEvent(e core.Event, t core.EventType) {
	switch t {
	case event.EventType_hanhua:
		wrap := e.(core.WrapEvent)
		hanhua := wrap.Event.(*event.Hanhua)
		index := self.seatIndex(hanhua.AccountId)
		if index == -1 {
			log.Errorf("玩家%v 不再座位上", hanhua.AccountId)
			return
		}
		player := self.seats[index]
		speech := []int{1, 2, 3, 4, 5}
		for i := 4; i >= 0; i-- {
			randVal := utils.Randx_y(0, i)
			if i := self.CanSpeech(player.acc.AccountId, types.ESpeechStatus(speech[randVal])); i == 0 {
				speech = append(speech[:randVal], speech[randVal+1:]...)
			} else {
				j := packet.NewPacket(nil)
				j.WriteUInt32(uint32(player.acc.AccountId))
				switch speech[randVal] {
				case 1:
					j.SetMsgID(protomsg.Old_MSGID_CX_HANHUA_XIU.UInt16())
				case 2:
					j.SetMsgID(protomsg.Old_MSGID_CX_HANHUA_DIU.UInt16())
				case 3:
					j.SetMsgID(protomsg.Old_MSGID_CX_HANHUA_DA.UInt16())
					j.WriteInt64(int64(self.Da_Val(player.acc.AccountId)))
				case 4:
					j.SetMsgID(protomsg.Old_MSGID_CX_HANHUA_QIAO.UInt16())
				case 5:
					j.SetMsgID(protomsg.Old_MSGID_CX_HANHUA_GEN.UInt16())
					j.WriteInt64(int64(self.Da_Val(player.acc.AccountId)))
				}
				core.CoreSend(0, common.EActorType_MAIN.Int32(), j.GetData(), 0)
				break
			}
		}
	}
}
