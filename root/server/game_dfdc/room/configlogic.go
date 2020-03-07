package room

import (
	"fmt"
	"math/rand"
	"root/common/config"
	"root/core/log"
	"root/core/utils"
	"root/protomsg"
	"sort"
	"time"
)

type (
	//图案节点
	pictureNode struct {
		cfId    int //图案id
		cfOdd_3 int //图案3连赔率
		cfOdd_4 int //图案4连赔率
		cfOdd_5 int //图案5连赔率
	}
	//轮轴
	wheelNode struct {
		cfPosition int   //图案位置
		ids        []int //图案id列表
	}
)

func (self *Room) LoadConfig() {
	bets_conf := config.Get_configString("dfdc_room", int(self.roomId), "Bet")
	self.bets = utils.SplitConf2ArrUInt64(bets_conf)
	self.addr_url = config.GetPublicConfig_String(5)
	self.Conf_JackpotBet = int64(config.Get_configInt("dfdc_room", int(self.roomId), "JackpotBet"))
	self.Conf_Bet_Probability = int64(config.Get_configInt("dfdc_room", int(self.roomId), "Bet_Probability"))

	self.mapPictureNodes = make(map[int32]*protomsg.ENTER_GAME_DFDC_RES_DfdcRatio)
	for _, id := range protomsg.DFDCID_value {
		if id == 0 {
			continue
		}
		self.mapPictureNodes[id] = &protomsg.ENTER_GAME_DFDC_RES_DfdcRatio{
			ID:     protomsg.DFDCID(id),
			Same_3: int32(config.Get_configInt("dfdc_pattern", int(id), "Odds3")),
			Same_4: int32(config.Get_configInt("dfdc_pattern", int(id), "Odds4")),
			Same_5: int32(config.Get_configInt("dfdc_pattern", int(id), "Odds5")),
		}
	}

	self.mainWheel, self.freeWheel = initWheel(int64(config.Get_configInt("dfdc_room", int(self.roomId), "Real")))

	self.bonusRatio = make([][]int32, 0)
	conf := config.Get_config("dfdc_jackpot")
	group := config.Get_configInt("dfdc_room", int(self.roomId), "Real")
	for id, _ := range conf {
		if config.Get_configInt("dfdc_jackpot", id, "Group_Id") != group {
			continue
		}
		lv := config.Get_configInt("dfdc_jackpot", id, "Level")
		ratio := config.Get_configInt("dfdc_jackpot", id, "Reward_Probability")
		initGold := int64(config.Get_configInt("dfdc_jackpot", id, "Gold"))
		self.bonus[int32(lv)] = initGold
		self.bounsInitGold[int32(lv)] = initGold
		self.bonusRatio = append(self.bonusRatio, []int32{int32(lv), int32(ratio)})
		self.bounsRoller[int32(lv)] = int64(config.Get_configInt("dfdc_jackpot", id, "Roller_Rate"))
	}
	log.Infof("房间:%v 配置加载完成", self.roomId)
}

func initWheel(group int64) (main, free []*wheelNode) {
	main = make([]*wheelNode, 0)
	free = make([]*wheelNode, 0)
	conf := config.Get_config("dfdc_real")
	for id, _ := range conf {
		if config.Get_configInt("dfdc_real", id, "Group_id") != int(group) {
			continue
		}
		node := new(wheelNode)
		node.cfPosition = config.Get_configInt("dfdc_real", id, "Site")
		if node.cfPosition > 0 {
			for i := 1; i <= 5; i++ {
				value := config.Get_configInt("dfdc_real", id, fmt.Sprintf("Real%v", i))
				node.ids = append(node.ids, value)
			}
			if t := config.Get_configInt("dfdc_real", id, "Type"); t == 1 {
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

// 图案id 连续个数
// 返回 赔率,免费次数
func (self *Room) getOddsByPictureId(cfId int32, count int) (int32, int32) {
	odds := int32(0)
	fee := int32(0)

	pPic := self.mapPictureNodes[cfId]
	if nil == pPic {
		log.Errorf("配置解析错误 函数:getOddsByPictureId cfId:%d", cfId)
		return 0, 0
	}
	switch count {
	case 3:
		{
			odds = pPic.Same_3
			fee = int32(config.Get_configInt("dfdc_pattern", int(cfId), "Free3"))
			break
		}
	case 4:
		{
			odds = pPic.Same_4
			fee = int32(config.Get_configInt("dfdc_pattern", int(cfId), "Free4"))
			break
		}
	case 5:
		{
			odds = pPic.Same_5
			fee = int32(config.Get_configInt("dfdc_pattern", int(cfId), "Free5"))
			break
		}
	default:
		{
			break
		}
	}
	return odds, fee
}

/**
该函数用于在轮轴列表中选出15个点，并且判断每条线的倍率已经总的免费次数
input: @nodes 选择的轮轴列表
return:
	@ args[0] awardResluts 中奖列表
	@ args[0] []int32 图案一维数组
	@ args[1] int 增加的免费次数
	@ args[2] float32 总倍数
	@ args[2] int 小玛利连续次数
	@ args[3] 中奖总倍数
	@ args[4] 获得大奖的数量
*/
//picA, freeCount, int64(sumOdds), reward,pos
func (self *Room) selectWheel(nodes []*wheelNode, betNum int64) (picA []int32, freeCount int, sumOdds int64, showPos []*protomsg.DFDCPosition) {
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

	//选出所有的图案id 组成3*5的图
	var coordinate [3][5]int
	for i := 0; i < 5; i++ {
		c := f()
		coordinate[0][i] = int(self.mapPictureNodes[int32(nodes[c[0]].ids[i])].ID)
		coordinate[1][i] = int(self.mapPictureNodes[int32(nodes[c[1]].ids[i])].ID)
		coordinate[2][i] = int(self.mapPictureNodes[int32(nodes[c[2]].ids[i])].ID)
	}

	resultMap := make(map[int][]int)
	tempPos := make(map[int][]*protomsg.DFDCPosition, 0)
	for i := 0; i < 3; i++ {
		val := coordinate[i][0] // 获得第一列的图案
		if val == 1 {
			log.Panicf("第一列出现了福字，请策划检查:%v", coordinate)
		}
		arr, e := resultMap[val]
		if !e {
			resultMap[val] = []int{1, 0, 0, 0, 0}
		} else {
			arr[0]++
		}
		tempPos[val] = append(tempPos[val], &protomsg.DFDCPosition{Px: int32(i), Py: 0})
	}

	for val, arr := range resultMap {
		totalline := arr[0]
		continous := 1

		// 遍历后4列，统计图案出现数量
		for y := 1; y < 5; y++ {
			// 遍历每一列的值
			for x := 0; x < 3; x++ {
				if cv := coordinate[x][y]; cv == val || cv == 1 {
					arr[y]++
					tempPos[val] = append(tempPos[val], &protomsg.DFDCPosition{Px: int32(x), Py: int32(y)})
				}
			}
			if arr[y] != 0 {
				totalline *= arr[y]
				continous++
			} else {
				break
			}
		}

		if continous <= 2 {
			continue
		}

		showPos = append(showPos, tempPos[val]...)
		odds, free := self.getOddsByPictureId(int32(val), continous)
		sumOdds += int64(odds) * int64(totalline)
		freeCount += int(free)
		//if odds > 0 {
		//	log.Infof("图案:%v 最大连数:%v 赔率:%v totalline:%v fee:%v arr:%v pos:%+v",val,continous,odds,totalline,free,arr,tempPos[val])
		//}
	}
	//if sumOdds > 0 {
	//	for i:=0;i < 3;i++{
	//		log.Infof("%v", coordinate[i])
	//	}
	//}

	picA = make([]int32, 0)
	for i := 0; i < 5; i++ {
		picA = append(picA, int32(coordinate[0][i]))
		picA = append(picA, int32(coordinate[1][i]))
		picA = append(picA, int32(coordinate[2][i]))
	}

	return picA, freeCount, sumOdds, showPos
}
