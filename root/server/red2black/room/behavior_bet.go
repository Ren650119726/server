package room

import (
	"root/core"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/red2black/event"
)

const TOTAL_TIME = 20

type (
	Robot_Bet struct {
		Room *Room
	}
)

var bet_conf = [][]int32{{500, 40}, {1000, 35}, {5000, 15}, {10000, 5}, {50000, 5}} // 下注金额权重
var bet_area_conf = [][]int32{{1, 48}, {2, 48}, {3, 4}}                             // 押注区域权重
var bet_timer_conf = [][]int32{
	{0, 0}, {0, 0}, {0, 2}, {0, 3}, {0, 4},
	{0, 5}, {0, 7}, {0, 10}, {0, 25}, {0, 30},
	{0, 40}, {0, 50}, {0, 50}, {0, 40}, {0, 30},
	{0, 25}, {0, 20}, {0, 15}, {0, 10}, {0, 10},
	{0, 10}, {0, 10}, {0, 5}, {0, 0}, {0, 0}} // 押注时间权重

func New_Behavior_Bet(room *Room) {
	obj := &Robot_Bet{Room: room}
	event.Dispatcher.AddEventListener(event.EventType_EnterBetting, obj)
}

func (self *Robot_Bet) OnEvent(e core.Event, t core.EventType) {
	switch t {
	case event.EventType_EnterBetting:
		wrap := e.(core.WrapEvent)
		data := wrap.Event.(*event.EnterBetting)
		self.bet_logic(data)
	}
}

func (self *Robot_Bet) bet_logic(ev *event.EnterBetting) {
	if self.Room.total_master_val() == 0 {
		return
	}
	for _, acc := range ev.Robots {
		join := utils.Probability(50)
		if !join {
			quit := utils.Probability(20)
			if quit {
				// 不押注的机器人有一定几率退出游戏
				accid := acc.AccountId
				core.LocalCoreSend(0, int32(ev.RoomID), func() {

					quit_timer := utils.Randx_y(0, int(ev.Duration))
					self.Room.owner.AddTimer(int64(quit_timer*1000), 1, func(dt int64) {
						self.Room.leaveRoom(accid, false)
					})

				})
			}
			continue
		}

		total_val := acc.GetMoney()
		total_time := utils.Randx_y(0, TOTAL_TIME)
		next_millisecond := 1
		betcount := 0
		for i := 0; i < total_time; i++ {
			if total_val <= 0 {
				break
			}
			index := utils.RandomWeight32(bet_conf, 1)
			bet_val := bet_conf[index][0]
			if uint64(bet_val) > total_val {
				continue
			}

			total_val -= uint64(bet_val)

			area_index := utils.RandomWeight32(bet_area_conf, 1)
			bet_area := bet_area_conf[area_index][0]

			total_area := len(bet_timer_conf)
			rand_area := utils.RandomWeight32(bet_timer_conf, 1)
			average := float64(ev.Duration) / float64(total_area)
			millisecond_rand := utils.Randx_y(0, 1000)
			result_area := average * float64(rand_area)
			if int(result_area) == 0 {
				pack := packet.NewPacket(nil)
				pack.SetMsgID(protomsg.Old_MSGID_R2B_BETTING.UInt16())
				pack.WriteUInt32(acc.AccountId)
				pack.WriteUInt8(uint8(bet_area))
				pack.WriteUInt32(uint32(bet_val))
				core.CoreSend(0, int32(ev.RoomID), pack.GetData(), 0)
			} else {
				accid := acc.AccountId
				roomid := ev.RoomID
				self.Room.owner.AddTimer(int64(result_area*1000)+int64(millisecond_rand)+int64(next_millisecond), 1, func(dt int64) {
					pack := packet.NewPacket(nil)
					pack.SetMsgID(protomsg.Old_MSGID_R2B_BETTING.UInt16())
					pack.WriteUInt32(accid)
					pack.WriteUInt8(uint8(bet_area))
					pack.WriteUInt32(uint32(bet_val))
					core.CoreSend(0, int32(roomid), pack.GetData(), 0)
				})
			}
			next_millisecond += 10
			betcount++
		}
	}

}
