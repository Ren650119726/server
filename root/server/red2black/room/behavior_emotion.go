package room

import (
	"root/common/config"
	"root/core"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/red2black/event"
)

type (
	Robot_Emotion struct {
		Room *Room
	}
)

func New_Behavior_Emotion(room *Room) {
	obj := &Robot_Emotion{Room: room}
	event.Dispatcher.AddEventListener(event.EventType_EnterWatting, obj)
	event.Dispatcher.AddEventListener(event.EventType_WinOrLoss, obj)
	event.Dispatcher.AddEventListener(event.EventType_Emotion, obj)
}

func (self *Robot_Emotion) OnEvent(e core.Event, t core.EventType) {
	switch t {
	case event.EventType_EnterWatting:
		wrap := e.(core.WrapEvent)
		data := wrap.Event.(*event.EnterWatting)
		self.Emotion_Watting_logic(data)
	case event.EventType_WinOrLoss:
		wrap := e.(core.WrapEvent)
		data := wrap.Event.(*event.WinOrLoss)
		self.Emotion_WinOrLoss_logic(data)
	case event.EventType_Emotion:
		wrap := e.(core.WrapEvent)
		data := wrap.Event.(*event.Emotion)
		self.Emotion_Emotion_logic(data)
	}
}

func (self *Robot_Emotion) Emotion_Watting_logic(ev *event.EnterWatting) {
	available_index := make([]uint32, 0)
	for _, acc := range ev.Seats {
		if acc != nil && acc.Robot > 0 {
			available_index = append(available_index, acc.AccountId)
		}
	}

	if len(available_index) <= 1 {
		return
	}

	if !utils.Probability(25) {
		return
	}

	randindex := utils.Randx_y(0, len(available_index))
	sendindex := available_index[randindex]
	available_index = append(available_index[:randindex], available_index[randindex+1:]...)
	randindex = utils.Randx_y(0, len(available_index))
	receiveID := available_index[randindex]

	pack := packet.NewPacket(nil)
	pack.SetMsgID(protomsg.Old_MSGID_SEND_EMOJI.UInt16())
	pack.WriteUInt32(uint32(sendindex))
	pack.WriteUInt32(uint32(receiveID))
	pack.WriteUInt8(2)
	emotions := config.GetPublicConfig_Slice("ROBOT_COMMON_MAGIC_EMOJI")
	randindex = utils.Randx_y(0, len(emotions))
	pack.WriteUInt8(uint8(emotions[randindex]))
	self.Room.owner.AddTimer(int64(utils.Randx_y(1, 50)*100), 1, func(dt int64) {
		core.CoreSend(0, int32(self.Room.roomId), pack.GetData(), 0)
	})

}

func (self *Robot_Emotion) Emotion_WinOrLoss_logic(ev *event.WinOrLoss) {
	for _, acc := range ev.Seats {
		if acc != nil && acc.Robot > 0 && acc.AccountId == ev.Acc.AccountId {
			if ev.Change > 50000 && utils.Probability(40) {
				pack := packet.NewPacket(nil)
				pack.SetMsgID(protomsg.Old_MSGID_SEND_EMOJI.UInt16())
				pack.WriteUInt32(uint32(acc.AccountId))
				pack.WriteUInt32(uint32(acc.AccountId))
				pack.WriteUInt8(1)
				emotions := config.GetPublicConfig_Slice("ROBOT_COMMON_EMOJI")
				randindex := utils.Randx_y(0, len(emotions))
				pack.WriteUInt8(uint8(emotions[randindex]))
				self.Room.owner.AddTimer(int64(utils.Randx_y(1, 50)*100), 1, func(dt int64) {
					core.CoreSend(0, int32(self.Room.roomId), pack.GetData(), 0)
				})
				return
			}

			if ev.Change > 0 {
				send := utils.Probability(20)
				if send {
					if acc != nil {
						pack := packet.NewPacket(nil)
						pack.SetMsgID(protomsg.Old_MSGID_SEND_EMOJI.UInt16())
						pack.WriteUInt32(uint32(acc.AccountId))
						pack.WriteUInt32(uint32(acc.AccountId))
						pack.WriteUInt8(1)
						emotions := config.GetPublicConfig_Slice("ROBOT_COMMON_EMOJI")
						randindex := utils.Randx_y(0, len(emotions))
						pack.WriteUInt8(uint8(emotions[randindex]))
						self.Room.owner.AddTimer(int64(utils.Randx_y(1, 50)*100), 1, func(dt int64) {
							core.CoreSend(0, int32(self.Room.roomId), pack.GetData(), 0)
						})
					}
				}
			}

			break
		}
	}
	for _, acc := range ev.MasterSeats {
		if acc != nil && acc.Robot > 0 && ev.Acc.AccountId == acc.AccountId {
			if ev.Change > 50000 && utils.Probability(40) {
				pack := packet.NewPacket(nil)
				pack.SetMsgID(protomsg.Old_MSGID_SEND_EMOJI.UInt16())
				pack.WriteUInt32(uint32(acc.AccountId))
				pack.WriteUInt32(uint32(acc.AccountId))
				pack.WriteUInt8(1)
				emotions := config.GetPublicConfig_Slice("ROBOT_COMMON_EMOJI")
				randindex := utils.Randx_y(0, len(emotions))
				pack.WriteUInt8(uint8(emotions[randindex]))
				self.Room.owner.AddTimer(int64(utils.Randx_y(1, 50)*100), 1, func(dt int64) {
					core.CoreSend(0, int32(self.Room.roomId), pack.GetData(), 0)
				})
				return
			}
		}
	}
}

func (self *Robot_Emotion) Emotion_Emotion_logic(ev *event.Emotion) {
	if utils.Probability(30) {
		pack := packet.NewPacket(nil)
		pack.SetMsgID(protomsg.Old_MSGID_SEND_EMOJI.UInt16())
		pack.WriteUInt32(uint32(ev.TargetID))
		pack.WriteUInt32(uint32(ev.SendID))
		pack.WriteUInt8(2)
		emotions := config.GetPublicConfig_Slice("ROBOT_COMMON_MAGIC_EMOJI")
		randindex := utils.Randx_y(0, len(emotions))
		pack.WriteUInt8(uint8(emotions[randindex]))
		self.Room.owner.AddTimer(int64(utils.Randx_y(20, 40)*100), 1, func(dt int64) {
			core.CoreSend(0, int32(self.Room.roomId), pack.GetData(), 0)
		})
	}
}
