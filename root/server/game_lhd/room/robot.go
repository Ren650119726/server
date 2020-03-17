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
		BetWeight       [][]int32
		AreaWeight      [][]int32
		DragonRandCount []int32
		TigerRandCount  []int32
		PeaceRatio      int
		PeaceCount      []int32
		BetFrequencies  []int32
	}

	behavior struct {
		bet         uint64
		area        protomsg.LHDAREA
		areacount   int32
		peace       bool
		peacecount  int32
		nextBetTime int64
	}
)

func (self *betting) robotbet(now int64) {
	time := utils.MilliSecondTimeSince1970()
	for id, robot := range self.robots {
		if (robot.area == protomsg.LHDAREA_LHD_AREA_Unknow && !robot.peace) || time < robot.nextBetTime || robot.nextBetTime == 0 {
			continue
		}
		acc := self.accounts[id]
		if acc == nil {
			continue
		}

		if acc.GetMoney() < robot.bet {
			if utils.Probability(10) {
				if self.leave(acc.AccountId) {
					self.leaveRoom(acc.GetAccountId())
				}
			}
			continue
		}

		if robot.areacount > 0 {
			betmsg := &protomsg.BET_LHD_REQ{
				AccountID: acc.GetAccountId(),
				Area:      robot.area,
				Bet:       robot.bet,
			}
			data, _ := proto.Marshal(betmsg)
			pack := packet.NewPacket(nil)
			pack.WriteBytes(data)
			pack.SetMsgID(protomsg.LHDMSG_CS_BET_LHD_REQ.UInt16())
			core.CoreSend(0, int32(self.roomId), pack.GetData(), 0)
			robot.areacount--
		}

		// 区域押注押完次数后判断是是否押和
		if robot.areacount == 0 && robot.peace {
			betmsg := &protomsg.BET_LHD_REQ{
				AccountID: acc.GetAccountId(),
				Area:      protomsg.LHDAREA_LHD_AREA_PEACE,
				Bet:       robot.bet,
			}
			dataluck, _ := proto.Marshal(betmsg)
			packluck := packet.NewPacket(nil)
			packluck.WriteBytes(dataluck)
			packluck.SetMsgID(protomsg.LHDMSG_CS_BET_LHD_REQ.UInt16())
			core.CoreSend(0, int32(self.roomId), packluck.GetData(), 0)
			robot.peacecount--
		}

		if robot.areacount == 0 && robot.peacecount == 0 {
			robot.nextBetTime = 0
		} else {
			robot.nextBetTime = time + int64(utils.Randx_y(int(self.robot_conf.BetFrequencies[0]), int(self.robot_conf.BetFrequencies[1])))
		}
	}
}

func (self *Room) robotQuit() {
	for _, robot := range self.accounts {
		if robot.Robot != 0 && robot.GetMoney() < uint64(self.betlimit_conf) {
			self.owner.AddTimer(int64(utils.Randx_y(100, 1000)*10), 1, func(dt int64) {
				self.leaveRoom(robot.GetAccountId())
			})
		}
	}
}
