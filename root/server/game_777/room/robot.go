package room

import (
	"root/common"
	"root/core/utils"
)

// 3秒 更新一次
func (self *Room) updateRobot(dt int64) {
	for _, acc := range self.accounts {
		if acc.Robot != 0 {
			if utils.Probability(50) {
				val := uint64(0)
				var randbetindex int
				randbetindex = utils.Randx_y(0, len(self.bets)/2)
				if acc.GetMoney() < self.bets[randbetindex] {
					continue
				}
				val += self.bets[randbetindex]
				if utils.Probability(50) {
					acc.AddMoney(-int64(val), common.EOperateType_S777_BET)
				} else if utils.Probability(80) {
					acc.AddMoney(int64(val), common.EOperateType_S777_WIN)
				} else {
					acc.AddMoney(int64(val*uint64(utils.Randx_y(2, 4))), common.EOperateType_S777_WIN)
				}
			}
			if acc.GetMoney() < self.bets[0] {
				if utils.Probability(20) {
					self.leaveRoom(acc.GetAccountId())
				}
			}
		}
	}

}
