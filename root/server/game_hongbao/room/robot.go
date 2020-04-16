package room

import (
	"root/core/log"
	"root/core/utils"
	"root/protomsg"
	"root/server/game_hongbao/account"
	"root/server/game_hongbao/send_tools"
)

// 自动发红包
func (self *Room) autoAssignHB(dt int64) {
	conf := utils.SplitConf2ArrInt64(self.Robot_Send_Interval)
	time := utils.Randx_y(int(conf[0]), int(conf[1]))
	self.owner.AddTimer(int64(time), 1, self.autoAssignHB)
	log.Infof("nextAutoAssignTime :%v ", time)

	conf = utils.SplitConf2ArrInt64(self.Robot_Send_Value)
	odds := utils.Randx_y(int(conf[0]), int(conf[1]))
	val := uint64(self.Min_Red * odds)

	arr := []uint32{}
	for k, _ := range self.Red_Odds {
		arr = append(arr, k)
	}
	num := utils.Randx_y(1, self.Robot_Send_Count+1)
	asshb := &protomsg.ASSIGN_HB_REQ{
		AccountID:  0,
		Value:      val,
		Count:      arr[utils.Randx_y(0, len(arr))],
		BombNumber: uint32(utils.Randx_y(0, 10)),
		Num:        uint32(num),
	}

	robots := self.robots()
	if len(robots) == 0 {
		return
	}
	var robot *account.Account
	randi := utils.Randx_y(0, len(robots)/2)
	for i := randi; i < len(robots); i++ {
		if robots[i].GetMoney() >= val {
			robot = robots[i]
			break
		}
	}

	if robot != nil {
		asshb.AccountID = robot.AccountId
		send_tools.Send2Room(protomsg.HBMSG_CS_ASSIGN_HB_REQ.UInt16(), asshb, int32(self.roomId))
		log.Infof("机器人 %v 发红包:%+v ", robot.AccountId, asshb)
	}

}

// 自动抢红包
func (self *Room) autoGrabHB(dt int64) {
	robots := self.robots()
	if len(robots) == 0 {
		return
	}

	for _, hb := range self.hbList {
		for i := 0; i < len(hb.arr); i++ {
			randi := utils.Randx_y(0, len(robots))
			for i := randi; i < len(robots); i++ {
				if _, e := hb.grabs[robots[i].AccountId]; !e {
					bombValue := uint64(self.Red_Odds[uint32(hb.count)] * hb.value / 100)
					if robots[i].GetMoney() < bombValue {
						continue
					}

					robotID := robots[i].AccountId
					hbID := hb.hbID
					self.owner.AddTimer(int64(utils.Randx_y(1, 50)*100), 1, func(dt int64) {
						send_tools.Send2Room(protomsg.HBMSG_CS_GRAB_HB_REQ.UInt16(), &protomsg.GRAB_HB_REQ{
							AccountID: robotID,
							ID:        uint32(hbID),
						}, int32(self.roomId))
					})
					break
				}
			}
		}
	}

}

// 自动退
func (self *Room) robotAutoQuit(dt int64) {
	robots := self.robots()
	for _, robot := range robots {
		if robot.GetMoney() < uint64(self.Min_Red*2) || utils.Probability(10) {
			if self.canleave(robot.AccountId) {
				self.leaveRoom(robot.AccountId)
			}
		}
	}
}
