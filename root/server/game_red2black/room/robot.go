package room

import (
	"github.com/golang/protobuf/proto"
	"root/core"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
)

type (
	robot_config struct {
		BetWeight      [][]int32
		RedBlackWeight [][]int32
		RedRandCount   []int32
		BlackRandCount []int32
		LuckRatio      int
		LuckCount      []int32
		BetFrequencies []int32
	}

	behavior struct {
		bet         uint64
		area        protomsg.RED2BLACKAREA
		areacount   int32
		luck        bool
		luckcount   int32
		nextBetTime int64
	}
)

func (self *betting) robotbet(now int64) {
	time := utils.MilliSecondTimeSince1970()
	for id, robot := range self.robots {
		if (robot.area == protomsg.RED2BLACKAREA_RED2BLACK_AREA_Unknow && !robot.luck) || time < robot.nextBetTime || robot.nextBetTime == 0 {
			continue
		}
		acc := self.accounts[id]
		if acc == nil {
			continue
		}

		if acc.GetMoney() < robot.bet {
			if utils.Probability(10) {
				self.leaveRoom(acc.GetAccountId())
			}
			continue
		}

		betmsg := &protomsg.BET_RED2BLACK_REQ{
			AccountID: acc.GetAccountId(),
			Area:      robot.area,
			Bet:       robot.bet,
		}
		data, _ := proto.Marshal(betmsg)
		pack := packet.NewPacket(nil)
		pack.WriteBytes(data)
		pack.SetMsgID(protomsg.RED2BLACKMSG_CS_BET_RED2BLACK_REQ.UInt16())
		core.CoreSend(0, int32(self.roomId), pack.GetData(), 0)
		robot.areacount--

		// 区域押注押完次数后判断是否幸运一击
		if robot.areacount == 0 && robot.luck {
			betmsg = &protomsg.BET_RED2BLACK_REQ{
				AccountID: acc.GetAccountId(),
				Area:      protomsg.RED2BLACKAREA_RED2BLACK_AREA_LUCK,
				Bet:       robot.bet,
			}
			dataluck, _ := proto.Marshal(betmsg)
			packluck := packet.NewPacket(nil)
			packluck.WriteBytes(dataluck)
			packluck.SetMsgID(protomsg.RED2BLACKMSG_CS_BET_RED2BLACK_REQ.UInt16())
			core.CoreSend(0, int32(self.roomId), packluck.GetData(), 0)
			robot.luckcount--
		}

		if robot.areacount == 0 && robot.luckcount == 0 {
			robot.nextBetTime = 0
		} else {
			robot.nextBetTime = time + int64(utils.Randx_y(int(self.robot_conf.BetFrequencies[0]), int(self.robot_conf.BetFrequencies[1])))
		}
	}
}

func (self *Room) robotQuit() {
	for _, robot := range self.accounts {
		if robot.Robot != 0 && robot.GetMoney() < uint64(self.betlimit) {
			self.owner.AddTimer(int64(utils.Randx_y(100, 500)*10), 1, func(dt int64) {
				self.leaveRoom(robot.GetAccountId())
			})
		}
	}
}
