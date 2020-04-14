package room

import (
	"root/common/config"
	"root/core/log"
	"root/core/utils"
)

func (self *Room) LoadConfig() {
	self.conf = &conf{}
	self.Min_Red = config.Get_configInt("red_room", int(self.roomId), "Min_Red")
	self.Max_Red = config.Get_configInt("red_room", int(self.roomId), "Max_Red")
	self.Red_Count = config.Get_configInt("red_room", int(self.roomId), "Red_Count")
	self.Pump = config.Get_configInt("red_room", int(self.roomId), "Pump")
	self.Robot_Num = config.Get_configInt("red_room", int(self.roomId), "Robot_Num")
	self.Robot_Send_Interval = config.Get_configString("red_room", int(self.roomId), "Robot_Send_Interval")
	self.Robot_Send_Count = config.Get_configInt("red_room", int(self.roomId), "Robot_Send_Count")
	self.Robot_Send_Value = config.Get_configString("red_room", int(self.roomId), "Robot_Send_Value")
	self.Rand_Point = config.Get_configInt("red_room", int(self.roomId), "Rand_Point")

	log.Infof("房间:%v 配置加载完成", self.roomId)
}

// money 红包金额
// num   切分的红包数量
// ratio 每个红包随机比例 / 10000
func hongbao_slice(money, num, ratio int64) []int64 {
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

	// 做一个随机处理
	for i := 0; i < int(num); i++ {
		ri := utils.Randx_y(i, int(num))
		ret[i], ret[ri] = ret[ri], ret[i]
	}
	return ret
}
