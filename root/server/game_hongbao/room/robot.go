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
	var robot *account.Account
	robots := self.robots()
	randi := utils.Randx_y(0, len(robots)/2)
	for i := randi; i < len(robots); i++ {
		if robots[i].GetMoney() >= val {
			robot = robots[i]
			break
		}
	}

	if robot != nil {
		asshb.AccountID = robot.AccountId
		send_tools.Send2Main(protomsg.HBMSG_CS_ASSIGN_HB_REQ.UInt16(), asshb)
		log.Infof("机器人 %v 发红包:%+v ", robot.AccountId, asshb)
	}

}
