package room

import (
	"root/core"
	"root/core/utils"
	"root/server/red2black/account"
	"root/server/red2black/event"
)

type (
	Robot_Quit struct {
		Room *Room
	}
)

// {min,max,weight} ，退出范围的权重
var quit_conf = [][]int32{{10, 15, 7}, {16, 20, 15}, {21, 40, 40}, {41, 999999, 50}} // 押注区域权重

func New_Behavior_Quit(room *Room) {
	obj := &Robot_Quit{Room: room}
	event.Dispatcher.AddEventListener(event.EventType_EnterWatting, obj)
}

func (self *Robot_Quit) OnEvent(e core.Event, t core.EventType) {
	switch t {
	case event.EventType_EnterWatting:
		wrap := e.(core.WrapEvent)
		data := wrap.Event.(*event.EnterWatting)
		self.quit_logic(data)
	}
}

// go to kidnap
func (self *Robot_Quit) quit_logic(ev *event.EnterWatting) {
	for _, robot := range ev.Robots {
		times := robot.Games // 游戏局数
		for _, quitConf := range quit_conf {
			if quitConf[0] <= times && times <= quitConf[1] {
				if utils.Probability(int(quitConf[2])) {
					quitaccid := robot.AccountId
					core.LocalCoreSend(0, int32(ev.RoomID), func() {
						quit_timer := utils.Randx_y(0, int(ev.Duration))
						self.Room.owner.AddTimer(int64(quit_timer*1000), 1, func(dt int64) {
							acc := account.AccountMgr.GetAccountByID(quitaccid)
							if acc != nil {
								self.Room.leaveRoom(quitaccid, false)
							}

						})
					})
				}
				break
			}
		}
	}
}
