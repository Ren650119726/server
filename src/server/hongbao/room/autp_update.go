package room

import (
	"root/common/config"
	"root/core"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/hongbao/account"
)

func (self *Room) auto_push_hongbao(now int64) {
	cur_count := len(self.hongbao_list)
	if cur_count < 5 || (cur_count < 10 && utils.Probability(70)) {
		self.rand_push_X(10)
		return
	}

	if utils.Probability(20) {
		self.rand_push_X(1)
	}
}

// 随机发N个红包
func (self *Room) rand_push_X(count int) {
	num := utils.Randx_y(1, count+1)
	for i := 0; i < num; i++ {
		next_section := i * 1000
		self.owner.AddTimer(int64(utils.Randx_y(500+next_section, 2000+next_section)), 1, func(dt int64) {
			robots := self.Robots()
			if count := len(robots); count != 0 {
				robot := robots[utils.Randx_y(0, count)]
				self.robot_push_hongbao(robot)
			}
		})
	}
}
func (self *Room) robot_push_hongbao(robot *account.Account) {
	str_conf := config.GetPublicConfig_String("HB_ROBOT_HONGBAO")
	ratio := utils.SplitConf2Arr_ArrInt64(str_conf)
	rate := ratio[utils.RandomWeight64(ratio, 1)][0]

	bet := self.GetParamInt(0)
	max_rate := self.GetParamInt(3)
	if rate > int64(max_rate) {
		rate = int64(max_rate)
	}

	need := int64(bet) * rate

	surplus := robot.GetMoney()

	if surplus < uint64(bet) {
		return
	}

	if surplus < uint64(need) {
		rate = 1
	}

	push_hongbao := packet.NewPacket(nil)
	push_hongbao.SetMsgID(protomsg.Old_MSGID_HONGBAO_POST_HONGBAO.UInt16())
	push_hongbao.WriteUInt32(robot.AccountId)
	push_hongbao.WriteUInt16(uint16(rate))
	push_hongbao.WriteInt8(int8(utils.Randx_y(0, 10)))
	core.CoreSend(0, int32(self.roomId), push_hongbao.GetData(), 0)
}

// 抢红包
func (self *Room) robot_rob_hongbao() {
	robots := self.Robots() // 所有机器人
	for _, robot := range robots {
		if utils.Probability(65) {
			accid := robot.AccountId
			self.owner.AddTimer(int64(utils.Randx_y(100, 3500)), 1, func(dt int64) {
				robmsg := packet.NewPacket(nil)
				robmsg.SetMsgID(protomsg.Old_MSGID_HONGBAO_ROB_HONGBAO.UInt16())
				robmsg.WriteUInt32(accid)
				core.CoreSend(0, int32(self.roomId), robmsg.GetData(), 0)
			})
		}
	}
}

// 机器人自动退出
func (self *Room) robot_quit(now int64) {
	robots := self.Robots() // 所有机器人

	i := 0
	for _, robot := range robots {
		probability := false

		if robot.GetMoney() < uint64(self.GetParamInt(0)*3) {
			if utils.Probability(80) {
				probability = true
			}
		} else {
			if utils.Probability(1) {
				probability = true
			}
		}

		if (probability) && !robot.Quit_flag {
			t := false
			for _, hongbao := range self.hongbao_list {
				if robot.AccountId == hongbao.acc.AccountId {
					t = true
					break
				}
			}

			if t {
				continue
			}
			i++
			robot.Quit_flag = true
			accid := robot.AccountId
			self.owner.AddTimer(int64(utils.Randx_y(i*1500, 5000+i*1500)), 1, func(dt int64) {
				self.leaveRoom(accid)
			})
		}
	}
}
