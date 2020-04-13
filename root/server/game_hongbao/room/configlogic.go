package room

import (
	"root/common/config"
	"root/core/log"
	"root/core/utils"
)

func (self *Room) LoadConfig() {
	_ = config.Get_configString("jpm_room", int(self.roomId), "Bet")

	log.Infof("房间:%v 配置加载完成", self.roomId)
}

// money 红包金额
// num   切分的红包数量
// ratio 每个红包随机比例 / 10000
func honagbao_slice(money, num, ratio int64) []int64 {
	if money < num {
		log.Errorf("钱不够分 money:%v num:%v,ratio:%v", money, num, ratio)
		return []int64{}
	}
	ret := make([]int64, num, num)
	// 每个红包，先分配最少金额
	for i := 0; i < int(num); i++ {
		ret[i] = 1
	}

	money -= num
	if money == 0 {
		return ret
	}

	// 根据随机比例，依次添加到红包中
	for i := 0; i < int(num); i++ {
		r := utils.Randx_y(0, int(ratio/100)+1) * 100
		add_val := money * int64(r) / 10000
		if add_val == 0 {
			break
		}
		ret[i] += add_val
		money -= add_val
	}

	return ret
}
