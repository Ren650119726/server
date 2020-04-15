package room

import (
	"root/common/config"
	"root/core/log"
	"root/core/utils"
)

func (self *Room) LoadConfig() {
	self.conf = &conf{Red_Odds: make(map[uint32]int64)}
	self.Min_Red = config.Get_configInt("red_room", int(self.roomId), "Min_Red")
	self.Max_Red = config.Get_configInt("red_room", int(self.roomId), "Max_Red")
	self.Red_Count = config.Get_configInt("red_room", int(self.roomId), "Red_Count")
	self.Pump = config.Get_configInt("red_room", int(self.roomId), "Pump")
	self.Robot_Send_Interval = config.Get_configString("red_room", int(self.roomId), "Robot_Send_Interval")
	self.Robot_Send_Count = config.Get_configInt("red_room", int(self.roomId), "Robot_Send_Count")
	self.Robot_Send_Value = config.Get_configString("red_room", int(self.roomId), "Robot_Send_Value")
	self.Rand_Point = config.Get_configInt("red_room", int(self.roomId), "Rand_Point")
	self.Red_Max = uint64(config.Get_configInt("red_room", int(self.roomId), "Red_Max"))

	red_odds := config.Get_config("red_odds")
	for k, _ := range red_odds {
		num := config.Get_configInt("red_odds", k, "Pack_Num")
		odds := config.Get_configInt("red_odds", k, "Odds")
		self.Red_Odds[uint32(num)] = int64(odds)
	}
	log.Infof("房间:%v 配置加载完成", self.roomId)
}

// money 红包金额
// num   切分的红包数量
// ratio 每个红包随机比例 / 10000
// bombc 至少需要几个雷
// bombNum 雷号
func hongbao_slice(money, num, ratio int64, bombc, bombNum int) []int64 {
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
	for i := 0; i < int(num-1); i++ {
		r := utils.Randx_y(0, int(ratio/100)+1) * 100
		add_val := money * int64(r) / 10000
		ret[i] += add_val
		money -= add_val
	}
	ret[int(num-1)] += money

	if bombc != 0 {
		// 先把雷全部放到前面
		bombi := 0
		for i := 0; i < int(num); i++ {
			if ret[i]%10 == int64(bombNum) {
				ret[i], ret[bombi] = ret[bombi], ret[i]
				bombi++
				bombc--
			}
		}

		// 剩余需要的雷数，大于剩下不是雷的红包个数，做一个安全处理，并且给出警告
		if bombc >= len(ret[bombi:]) {
			log.Warnf("Error 需要的雷是否太多？？！！金额:%v 包数:%v 比率:%v 雷数:%v 雷号:%v arr:%v", money, num, ratio, bombc, bombNum, ret)
			bombc = int(num - 1)
		}
		for i := bombi; i < int(num); i++ {
			if bombc <= 0 {
				break
			}
			unit := ret[i] % 10
			val := bombNum - int(unit)
			ret[i] += int64(val)
			ret[i+1] -= int64(val)
			bombc--
		}
	}

	// 做一个随机处理
	for i := 0; i < int(num); i++ {
		ri := utils.Randx_y(i, int(num))
		ret[i], ret[ri] = ret[ri], ret[i]
	}
	return ret
}
