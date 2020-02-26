package room

import (
	"root/common/config"
	"root/core"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/red2black/event"
)

const PREVIOUS_AVAILABLE_SEATCOUNT = 2 // 给玩家预留的座位数量

type (
	Robot_SeatDown struct {
		Room *Room
	}
)

func New_Behavior_SeatDown(room *Room) {
	obj := &Robot_SeatDown{Room: room}
	event.Dispatcher.AddEventListener(event.EventType_PlayerCountChange, obj)
}

func (self *Robot_SeatDown) OnEvent(e core.Event, t core.EventType) {
	switch t {
	case event.EventType_PlayerCountChange:
		wrap := e.(core.WrapEvent)
		data := wrap.Event.(*event.PlayerCountChange)
		self.Room.owner.AddTimer(int64(utils.Randx_y(10, 1000)), 1, func(dt int64) {
			self.seat_logic(data)
		})

	}
}

func (self *Robot_SeatDown) seat_logic(ev *event.PlayerCountChange) {
	seats := ev.Seats // 座位上的所有人
	seatCount := 0
	available_indexs := make([]int, 0)
	for index, acc := range seats {
		if acc != nil {
			seatCount++
		} else {
			available_indexs = append(available_indexs, index)
		}
	}

	inSeat := func(accid uint32) bool {
		for _, acc := range seats {
			if acc != nil && acc.AccountId == accid {
				return true
			}
		}
		return false
	}

	// 给玩家预留2个位置
	if len(available_indexs) <= PREVIOUS_AVAILABLE_SEATCOUNT {
		return
	}

	for _, robot := range ev.Robots {
		if inSeat(robot.AccountId) {
			continue
		}
		if robot.GetMoney() < uint64(config.GetPublicConfig_Int64("R2B_UP_SEAT_MONEY")) {
			continue
		}

		hit := utils.Probability(60)
		if !hit {
			continue
		}

		rand_seat := utils.Randx_y(0, len(available_indexs))
		available_index := available_indexs[rand_seat]
		available_indexs = append(available_indexs[:rand_seat], available_indexs[rand_seat+1:]...)
		pack := packet.NewPacket(nil)
		pack.SetMsgID(protomsg.Old_MSGID_R2B_UP_SEAT.UInt16())
		pack.WriteUInt32(robot.AccountId)
		pack.WriteUInt8(uint8(available_index + 1))
		core.CoreSend(0, int32(ev.RoomID), pack.GetData(), 0)

		// 给玩家预留2个位置
		if len(available_indexs) <= PREVIOUS_AVAILABLE_SEATCOUNT {
			return
		}
	}
}
