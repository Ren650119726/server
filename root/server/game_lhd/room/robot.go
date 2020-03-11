package room

import (
	"github.com/golang/protobuf/proto"
	"root/core"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
)

func (self *betting) robotbet(now int64) {
	for _, acc := range self.accounts {
		if acc.Robot != 0 {
			if utils.Probability(20) {
				betWeight := [][]int32{{0, 60}, {1, 20}, {2, 10}, {3, 8}, {4, 2}}
				i := utils.RandomWeight32(betWeight, 1)
				bet := uint64(self.bets_conf[uint64(betWeight[i][0])])
				if acc.GetMoney() < bet {
					if utils.Probability(10) {
						self.leaveRoom(acc.GetAccountId())
					}
					continue
				}

				areaWeight := [][]int32{{1, 42}, {2, 42}, {3, 6}}
				i = utils.RandomWeight32(areaWeight, 1)
				area := areaWeight[i][0]
				betmsg := &protomsg.BET_RED2BLACK_REQ{
					AccountID: acc.GetAccountId(),
					Area:      protomsg.RED2BLACKAREA(area),
					Bet:       bet,
				}
				data, _ := proto.Marshal(betmsg)
				pack := packet.NewPacket(nil)
				pack.WriteBytes(data)
				pack.SetMsgID(protomsg.LHDMSG_CS_BET_LHD_REQ.UInt16())

				self.owner.AddTimer(int64(utils.Randx_y(5, 30)*100), 1, func(dt int64) {
					core.CoreSend(0, int32(self.roomId), pack.GetData(), 0)
				})
			}

		}
	}
}

func (self *Room) robotQuit() {
	for _, robot := range self.accounts {
		if robot.Robot != 0 && robot.GetMoney() < uint64(self.betlimit_conf) {
			self.owner.AddTimer(int64(utils.Randx_y(100, 500)*10), 1, func(dt int64) {
				self.leaveRoom(robot.GetAccountId())
			})
		}
	}
}
