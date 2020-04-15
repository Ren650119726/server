package room

import (
	"fmt"
	"github.com/astaxie/beego"
	"math/rand"
	"root/common/config"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/utils"
	"sort"
	"time"
)

type (
	//轮轴
	wheelNode struct {
		cfPosition int   //图案位置
		ids        []int //图案id列表
	}

	// 图案中奖
	BingoPicture struct {
		ID   int
		Odds int
	}

	// 大奖翻倍信息
	bigWinConf struct {
		weight int
		odds   int
	}
)

func (self *Room) LoadConfig() {
	bets_conf := config.Get_configString("777_room", int(self.roomId), "Bet")
	self.bets = utils.SplitConf2ArrUInt64(bets_conf)
	self.addr_url = beego.AppConfig.DefaultString("DEF::setuserinfo", "")
	self.Conf_JackpotBet = int64(config.Get_configInt("777_room", int(self.roomId), "JackpotBet"))
	self.mainWheel, _ = initWheel(int64(config.Get_configInt("777_room", int(self.roomId), "Real")))

	self.bonus_jackpot_conf = make(map[int32]int)
	conf := config.Get_config("777_jackpot")
	group := config.Get_configInt("777_room", int(self.roomId), "Real")
	for id, _ := range conf {
		if config.Get_configInt("777_jackpot", id, "Group_Id") != group {
			continue
		}
		lv := config.Get_configInt("777_jackpot", id, "Level")
		initGold := int64(config.Get_configInt("777_jackpot", id, "Gold"))
		rewardID := int32(config.Get_configInt("777_jackpot", id, "Reward_id"))
		self.bonus[int32(lv)] = 0
		self.bounsInitGold[int32(lv)] = initGold
		log.Infof("奖池初始: ", self.bounsInitGold)
		self.bounsRoller[int32(lv)] = int64(config.Get_configInt("777_jackpot", id, "Roller_Rate"))
		self.bonus_jackpot_conf[rewardID] = lv
	}

	self.bingoPictureID = make(map[int]*BingoPicture)
	reward := config.Get_config("777_reward")
	for id, _ := range reward {
		n := &BingoPicture{
			ID:   id,
			Odds: config.Get_configInt("777_reward", id, "Odds"),
		}
		real := config.Get_configInt("777_reward", id, "Real")
		self.bingoPictureID[real] = n
	}

	self.multiple_conf = make(map[int]*bigWinConf)
	bigWin := config.Get_config("777_multiple")
	for id, _ := range bigWin {
		n := &bigWinConf{
			weight: config.Get_configInt("777_multiple", id, "Point"),
			odds:   config.Get_configInt("777_multiple", id, "Multiple"),
		}
		self.multiple_conf[id] = n
	}
	self.multipleWeightArr = make([][]int32, 0)
	for k, v := range self.multiple_conf {
		self.multipleWeightArr = append(self.multipleWeightArr, []int32{int32(k), int32(v.weight)})
	}
	log.Infof("房间:%v 配置加载完成", self.roomId)
}

func initWheel(group int64) (main, free []*wheelNode) {
	main = make([]*wheelNode, 0)
	free = make([]*wheelNode, 0)
	conf := config.Get_config("777_real")
	for id, _ := range conf {
		if config.Get_configInt("777_real", id, "Group_id") != int(group) {
			continue
		}
		node := new(wheelNode)
		node.cfPosition = config.Get_configInt("777_real", id, "Site")
		if node.cfPosition > 0 {
			for i := 1; i <= 3; i++ {
				value := config.Get_configInt("777_real", id, fmt.Sprintf("Real%v", i))
				node.ids = append(node.ids, value)
			}
			if t := config.Get_configInt("777_real", id, "Type"); t == 1 {
				main = append(main, node)
			} else if t == 2 {
				free = append(free, node)
			}
		}
	}
	sort.SliceStable(main, func(i, j int) bool {
		return main[i].cfPosition < main[j].cfPosition
	})
	sort.SliceStable(free, func(i, j int) bool {
		return free[i].cfPosition < free[j].cfPosition
	})
	return main, free
}

/**
该函数用于在轮轴列表中选出15个点，并且判断每条线的倍率已经总的免费次数
input: @nodes 选择的轮轴列表
return:
	@ args[0] []int32 图案一维数组
	@ args[3] 中奖总倍数
*/
//picA, int64(sumOdds), reward,pos
func (self *Room) selectWheel(nodes []*wheelNode, betNum int64) (picA []int32, sumOdds int64, bigwinID int32, lv int, re int) {
	rand.Seed(time.Now().UnixNano() + int64(rand.Int31n(int32(10000))))
	// 随机一个索引x 组成一个集合 [x-1,x,x+1]
	f := func() [3]int {
		var a [3]int
		randIndex := rand.Int31n(int32(len(nodes)))
		if int32(len(nodes)-1) == randIndex {
			a[0] = int(randIndex - 1) //70
			a[1] = int(randIndex)     //71
			a[2] = 0
		} else if 0 == randIndex {
			a[0] = int((len(nodes) - 1)) //71
			a[1] = 0
			a[2] = 1
		} else {
			a[0] = int(randIndex - 1)
			a[1] = int(randIndex)
			a[2] = int(randIndex + 1)
		}
		return a
	}

	//选出所有的图案id 组成3*3的图
	var coordinate [3][3]int
	for i := 0; i < 3; i++ {
		c := f()
		coordinate[0][i] = nodes[c[0]].ids[i]
		coordinate[1][i] = nodes[c[1]].ids[i]
		coordinate[2][i] = nodes[c[2]].ids[i]
	}

	picA = make([]int32, 0)
	for i := 0; i < 3; i++ {
		picA = append(picA, int32(coordinate[0][i]))
		picA = append(picA, int32(coordinate[1][i]))
		picA = append(picA, int32(coordinate[2][i]))
	}

	realID := 100*coordinate[1][0] + 10*coordinate[1][1] + coordinate[1][2]
	realID0 := 100*coordinate[1][0] + 10*coordinate[1][1]
	rewardID := 0
	if reword, e := self.bingoPictureID[realID]; e {
		sumOdds = int64(reword.Odds)
		rewardID = reword.ID
	} else if reword, e := self.bingoPictureID[realID0]; e {
		sumOdds = int64(reword.Odds)
		rewardID = reword.ID
	}

	if realID%10 == 0 {
		re = 2
	} else {
		re = 3
	}

	// 随机第四个图案
	index := utils.RandomWeight32(self.multipleWeightArr, 1)
	bigwinID = self.multipleWeightArr[index][0]
	sumOdds *= int64(self.multiple_conf[int(bigwinID)].odds)

	jackpotLv := 0
	if bigwinID == 1 {
		lv, e := self.bonus_jackpot_conf[int32(rewardID)]
		if e {
			jackpotLv = lv
			log.Infof("中jackpot %v", rewardID)
		}
	}

	if sumOdds != 0 {
		log.Infof(colorized.Blue("中奖:%v %v realID:%v odds:%v"), coordinate[1], bigwinID, realID, sumOdds)
	}
	return picA, sumOdds, bigwinID, jackpotLv, re
}
