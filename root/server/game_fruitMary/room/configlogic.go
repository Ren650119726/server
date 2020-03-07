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
		cfOdd_2 int //图案2连赔率
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
	bets_conf := config.Get_configString("mary_room", int(self.roomId), "Bet")
	self.bets = utils.SplitConf2ArrUInt64(bets_conf)
	self.addr_url = config.GetPublicConfig_String(5)

	self.basics = int64(config.Get_configInt("mary_room", int(self.roomId), "JackpotBase"))
	self.jackpotRate = uint64(config.Get_configInt("mary_room", int(self.roomId), "JackpotRole"))
	self.jackLimit = int64(config.Get_configInt("mary_room", int(self.roomId), "JackpotBet"))

	self.FruitRatio = make(map[int32]*protomsg.ENTER_GAME_FRUITMARY_RES_FruitRatio)
	self.weight_ratio = make([][]int32, 0)
	for _, ID := range protomsg.Fruit2ID_value {
		if ID == 0 {
			continue
		}
		r := &protomsg.ENTER_GAME_FRUITMARY_RES_FruitRatio{
			ID:     protomsg.Fruit2ID(ID),
			Single: int32(config.Get_configInt("mary_bonuspattern", int(ID), "Odds1")),
			Same_2: int32(config.Get_configInt("mary_bonuspattern", int(ID), "Odds2")),
			Same_3: int32(config.Get_configInt("mary_bonuspattern", int(ID), "Odds3")),
			Same_4: int32(config.Get_configInt("mary_bonuspattern", int(ID), "Odds4")),
		}
		self.FruitRatio[ID] = r
		self.weight_ratio = append(self.weight_ratio, []int32{ID, int32(config.Get_configInt("mary_bonuspattern", int(ID), "Weight"))})
	}

	self.mapPictureNodes = make(map[int]*pictureNode)
	for _, id := range protomsg.Fruit1ID_value {
		if id == 0 {
			continue
		}
		self.mapPictureNodes[int(id)] = &pictureNode{
			cfId:    int(id),
			cfOdd_2: config.Get_configInt("mary_pattern", int(id), "Odds2"),
			cfOdd_3: config.Get_configInt("mary_pattern", int(id), "Odds3"),
			cfOdd_4: config.Get_configInt("mary_pattern", int(id), "Odds4"),
			cfOdd_5: config.Get_configInt("mary_pattern", int(id), "Odds5"),
		}
	}

	self.lineConf = make([][5]int, 10, 10)
	conf := config.Get_config("mary_lines")
	for id, _ := range conf {
		for i := 1; i <= 5; i++ {
			val := config.Get_configInt("mary_lines", id, fmt.Sprintf("site%v", i))
			self.lineConf[id][i-1] = val - 1
		}
	}
	self.mainWheel, self.freeWheel, self.maryWheel = initWheel(int64(config.Get_configInt("mary_room", int(self.roomId), "Real")))

	log.Infof("房间:%v 配置加载完成", self.roomId)
}

func initWheel(group int64) (main, free, mary []*wheelNode) {
	main = make([]*wheelNode, 0)
	free = make([]*wheelNode, 0)
	mary = make([]*wheelNode, 0)
	conf := config.Get_config("mary_real")
	for id, _ := range conf {
		if config.Get_configInt("mary_real", id, "Group_id") != int(group) {
			continue
		}
		node := new(wheelNode)
		node.cfPosition = config.Get_configInt("mary_real", id, "Site")
		if node.cfPosition > 0 {
			for i := 1; i <= 5; i++ {
				value := config.Get_configInt("mary_real", id, fmt.Sprintf("Real%v", i))
				node.ids = append(node.ids, value)
			}
			if t := config.Get_configInt("mary_real", id, "Type"); t == 1 {
				main = append(main, node)
			} else if t == 2 {
				free = append(free, node)
			} else if t == 3 {
				mary = append(mary, node)
			}
		}
	}
	sort.SliceStable(main, func(i, j int) bool {
		return main[i].cfPosition < main[j].cfPosition
	})
	sort.SliceStable(free, func(i, j int) bool {
		return free[i].cfPosition < free[j].cfPosition
	})
	sort.SliceStable(mary, func(i, j int) bool {
		return mary[i].cfPosition < mary[j].cfPosition
	})
	return main, free, mary
}

// 图案id 连续个数
// 返回赔率
func (self *Room) getOddsByPictureId(cfId int, count int) int {
	odds := int(0)

	pPic := self.mapPictureNodes[cfId]
	if nil == pPic {
		log.Errorf("配置解析错误 函数:getOddsByPictureId cfId:%d", cfId)
		return 0
	}
	switch count {
	case 2:
		{
			odds = pPic.cfOdd_2
			break
		}
	case 3:
		{
			odds = pPic.cfOdd_3
			break
		}
	case 4:
		{
			odds = pPic.cfOdd_4
			break
		}
	case 5:
		{
			odds = pPic.cfOdd_5
			break
		}
	default:
		{
			break
		}
	}
	return odds
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
func (self *Room) selectWheel(nodes []*wheelNode, betNum int64, isKill, test bool) ([]*protomsg.FRUITMARY_Result, []int32, int, int, int64, int64, []*protomsg.FRUITMARYPosition) {
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

	//选出所有的图案id
	var b [3][5]int

	ccc := 0
	spcifity_2_count := 0
	pos := []*protomsg.FRUITMARYPosition{}
	f5 := []*protomsg.FRUITMARYPosition{nil, nil, nil, nil, nil}
	for i := 0; i < 5; i++ {
		c := f()
		b[0][i] = self.mapPictureNodes[nodes[c[0]].ids[i]].cfId
		b[1][i] = self.mapPictureNodes[nodes[c[1]].ids[i]].cfId
		b[2][i] = self.mapPictureNodes[nodes[c[2]].ids[i]].cfId

		e := false

		for j := 0; j < 3; j++ {
			if 2 == b[j][i] {
				f5[i] = &protomsg.FRUITMARYPosition{Px: int32(j), Py: int32(i)}
				ccc++
				if spcifity_2_count < ccc {
					spcifity_2_count = ccc
				}
				e = true
				break
			}
		}
		if !e {
			ccc = 0
		}
	}
	freeCount := 0
	if spcifity_2_count >= 3 {
		ct := 0
		for i, v := range f5 {
			if v != nil {
				ct++
				if ct == spcifity_2_count {
					pos = f5[i+1-spcifity_2_count : i+1]
					break
				}
			} else {
				ct = 0
			}
		}
		freeCount = config.Get_configInt("mary_pattern", 2, fmt.Sprintf("Free%v", spcifity_2_count))
		log.Infof("中免费 坐标:%v", pos)
	}

	tmp := make([]*protomsg.FRUITMARY_Result, 0)
	sumOdds := 0
	reward := int64(0)
	bingocount := 0
	maryCount := 0
	// 判断所有中奖线路
	for lid, line := range self.lineConf {
		if lid == 0 {
			continue
		}
		positions := make([]*protomsg.FRUITMARYPosition, 0)
		tempArr := []int{} // 中奖线图片组
		tempposs := []*protomsg.FRUITMARYPosition{}
		for _, pos := range line {
			x := pos % 3
			y := pos / 3
			id := b[x][y]
			tempArr = append(tempArr, id)
			tempposs = append(tempposs, &protomsg.FRUITMARYPosition{Px: int32(x), Py: int32(y)})
		}

		count, spicify_1, bingo, marypos := self.win(tempArr, tempposs)

		ii := 0
		for _, pos := range line {
			if ii >= count {
				break
			}
			x := pos % 3
			y := pos / 3
			positions = append(positions, &protomsg.FRUITMARYPosition{int32(x), int32(y)})
			ii++
		}

		if spicify_1 >= 3 {
			ct := 0
			for i, v := range marypos {
				if v != nil {
					ct++
					if ct == spicify_1 {
						pos = append(pos, marypos[i+1-spicify_1:i+1]...)
					}
				} else {
					ct = 0
				}
			}
			maryCount += config.Get_configInt("mary_pattern", 1, fmt.Sprintf("Bouns%v", spicify_1))
			log.Infof("中小玛利 坐标:%v", pos)
		}
		// 中奖金了
		if bingo == 3 && count >= 3 {
			reward = (self.basics * betNum) + (self.bonus)
			val := config.Get_configInt("mary_pattern", 3, fmt.Sprintf("Jackpot%v", count))
			reward = reward * int64(val) / 10000
			if reward != 0 {
				log.Infof("中大奖了！！！！！中獎綫:%v bingo == 3 count:%v reward:%v val:%v self.basics:%v betNum:%v self.bonus:%v", lid, count, reward, val, self.basics, betNum, self.bonus)
			}
		}

		m := self.getOddsByPictureId(bingo, count)
		sumOdds += m
		if m > 0 {
			if !test {
				for i := 0; i < 3; i++ {
					log.Infof("%v", b[i])
				}
				log.Infof("检测图片组:%v 中獎綫:%v bingo == %v count:%v 单线赔率:%v 总赔率:%v ", tempArr, lid, bingo, count, m, sumOdds)
			}
			bingocount++
			tmp = append(tmp, &protomsg.FRUITMARY_Result{LineId: int32(lid), Count: int32(count), Odds: int32(m), Positions: positions})
		}
	}

	picA := make([]int32, 0)
	for i := 0; i < 5; i++ {
		picA = append(picA, int32(b[0][i]))
		picA = append(picA, int32(b[1][i]))
		picA = append(picA, int32(b[2][i]))
	}

	return tmp, picA, freeCount, maryCount, int64(sumOdds), reward, pos
}

// 返回中奖的连数，以及触发小玛利次数, 中奖的图片ID,中小玛利次数的图片坐标
func (self *Room) win(arr []int, inpos []*protomsg.FRUITMARYPosition) (count, maxMary, bingo int, pos []*protomsg.FRUITMARYPosition) {
	//判断是否中奖
	number := arr[0]
	count = 1
	maryCount := 0
	cont := true
	next := true
	pos = []*protomsg.FRUITMARYPosition{nil, nil, nil, nil, nil}
	for i := 0; i < 5; i++ {
		// 特殊判断连续1的个数
		if arr[i] == 1 {
			pos[i] = inpos[i]
			maryCount++
			if maxMary < maryCount {
				maxMary = maryCount
			}
		} else {
			maryCount = 0
		}

		if i == 0 {
			continue
		}

		if number == 1 && next {
			if arr[i] != 2 && arr[i] != 3 {
				number = arr[i]
			} else {
				next = false
			}
		}

		if !cont || (arr[i] != number && (arr[i] != 1 || number == 2 || number == 3)) {
			cont = false
			continue
		}
		count++
	}
	bingo = number
	return count, maxMary, bingo, pos
}
