package algorithm

import (
	"root/common"
	"root/core/utils"
	"fmt"
	"math/rand"
)

// 普通牛一到牛九的映射
var comm_niu_type = map[uint8]uint8{
	1: 1,
	2: 3,
	3: 5,
	4: 7,
	5: 9,
	6: 11,
	7: 13,
	8: 15,
	9: 17,
}

// 金牌牛一到牛九的映射
var gold_niu_type = map[uint8]uint8{
	1: 2,
	2: 4,
	3: 6,
	4: 8,
	5: 10,
	6: 12,
	7: 14,
	8: 16,
	9: 18,
}

var nn_point_string = [...]string{
	0:  "0",
	1:  "A",
	2:  "2",
	3:  "3",
	4:  "4",
	5:  "5",
	6:  "6",
	7:  "7",
	8:  "8",
	9:  "9",
	10: "10",
	11: "J",
	12: "Q",
	13: "K",
	14: "A",
}

// 不带大小王
var nn_cards_one = []common.Card_info{
	{common.ECardType_FANGKUAI.UInt8(), 1},
	{common.ECardType_FANGKUAI.UInt8(), 2},
	{common.ECardType_FANGKUAI.UInt8(), 3},
	{common.ECardType_FANGKUAI.UInt8(), 4},
	{common.ECardType_FANGKUAI.UInt8(), 5},
	{common.ECardType_FANGKUAI.UInt8(), 6},
	{common.ECardType_FANGKUAI.UInt8(), 7},
	{common.ECardType_FANGKUAI.UInt8(), 8},
	{common.ECardType_FANGKUAI.UInt8(), 9},
	{common.ECardType_FANGKUAI.UInt8(), 10},
	{common.ECardType_FANGKUAI.UInt8(), 11},
	{common.ECardType_FANGKUAI.UInt8(), 12},
	{common.ECardType_FANGKUAI.UInt8(), 13},

	{common.ECardType_MEIHUA.UInt8(), 1},
	{common.ECardType_MEIHUA.UInt8(), 2},
	{common.ECardType_MEIHUA.UInt8(), 3},
	{common.ECardType_MEIHUA.UInt8(), 4},
	{common.ECardType_MEIHUA.UInt8(), 5},
	{common.ECardType_MEIHUA.UInt8(), 6},
	{common.ECardType_MEIHUA.UInt8(), 7},
	{common.ECardType_MEIHUA.UInt8(), 8},
	{common.ECardType_MEIHUA.UInt8(), 9},
	{common.ECardType_MEIHUA.UInt8(), 10},
	{common.ECardType_MEIHUA.UInt8(), 11},
	{common.ECardType_MEIHUA.UInt8(), 12},
	{common.ECardType_MEIHUA.UInt8(), 13},

	{common.ECardType_HEITAO.UInt8(), 1},
	{common.ECardType_HEITAO.UInt8(), 2},
	{common.ECardType_HEITAO.UInt8(), 3},
	{common.ECardType_HEITAO.UInt8(), 4},
	{common.ECardType_HEITAO.UInt8(), 5},
	{common.ECardType_HEITAO.UInt8(), 6},
	{common.ECardType_HEITAO.UInt8(), 7},
	{common.ECardType_HEITAO.UInt8(), 8},
	{common.ECardType_HEITAO.UInt8(), 9},
	{common.ECardType_HEITAO.UInt8(), 10},
	{common.ECardType_HEITAO.UInt8(), 11},
	{common.ECardType_HEITAO.UInt8(), 12},
	{common.ECardType_HEITAO.UInt8(), 13},

	{common.ECardType_HONGTAO.UInt8(), 1},
	{common.ECardType_HONGTAO.UInt8(), 2},
	{common.ECardType_HONGTAO.UInt8(), 3},
	{common.ECardType_HONGTAO.UInt8(), 4},
	{common.ECardType_HONGTAO.UInt8(), 5},
	{common.ECardType_HONGTAO.UInt8(), 6},
	{common.ECardType_HONGTAO.UInt8(), 7},
	{common.ECardType_HONGTAO.UInt8(), 8},
	{common.ECardType_HONGTAO.UInt8(), 9},
	{common.ECardType_HONGTAO.UInt8(), 10},
	{common.ECardType_HONGTAO.UInt8(), 11},
	{common.ECardType_HONGTAO.UInt8(), 12},
	{common.ECardType_HONGTAO.UInt8(), 13},
}

// 带大小王
var nn_cards_two = []common.Card_info{
	{common.ECardType_JKEOR.UInt8(), 2},
	{common.ECardType_JKEOR.UInt8(), 3},

	{common.ECardType_FANGKUAI.UInt8(), 1},
	{common.ECardType_FANGKUAI.UInt8(), 2},
	{common.ECardType_FANGKUAI.UInt8(), 3},
	{common.ECardType_FANGKUAI.UInt8(), 4},
	{common.ECardType_FANGKUAI.UInt8(), 5},
	{common.ECardType_FANGKUAI.UInt8(), 6},
	{common.ECardType_FANGKUAI.UInt8(), 7},
	{common.ECardType_FANGKUAI.UInt8(), 8},
	{common.ECardType_FANGKUAI.UInt8(), 9},
	{common.ECardType_FANGKUAI.UInt8(), 10},
	{common.ECardType_FANGKUAI.UInt8(), 11},
	{common.ECardType_FANGKUAI.UInt8(), 12},
	{common.ECardType_FANGKUAI.UInt8(), 13},

	{common.ECardType_MEIHUA.UInt8(), 1},
	{common.ECardType_MEIHUA.UInt8(), 2},
	{common.ECardType_MEIHUA.UInt8(), 3},
	{common.ECardType_MEIHUA.UInt8(), 4},
	{common.ECardType_MEIHUA.UInt8(), 5},
	{common.ECardType_MEIHUA.UInt8(), 6},
	{common.ECardType_MEIHUA.UInt8(), 7},
	{common.ECardType_MEIHUA.UInt8(), 8},
	{common.ECardType_MEIHUA.UInt8(), 9},
	{common.ECardType_MEIHUA.UInt8(), 10},
	{common.ECardType_MEIHUA.UInt8(), 11},
	{common.ECardType_MEIHUA.UInt8(), 12},
	{common.ECardType_MEIHUA.UInt8(), 13},

	{common.ECardType_HEITAO.UInt8(), 1},
	{common.ECardType_HEITAO.UInt8(), 2},
	{common.ECardType_HEITAO.UInt8(), 3},
	{common.ECardType_HEITAO.UInt8(), 4},
	{common.ECardType_HEITAO.UInt8(), 5},
	{common.ECardType_HEITAO.UInt8(), 6},
	{common.ECardType_HEITAO.UInt8(), 7},
	{common.ECardType_HEITAO.UInt8(), 8},
	{common.ECardType_HEITAO.UInt8(), 9},
	{common.ECardType_HEITAO.UInt8(), 10},
	{common.ECardType_HEITAO.UInt8(), 11},
	{common.ECardType_HEITAO.UInt8(), 12},
	{common.ECardType_HEITAO.UInt8(), 13},

	{common.ECardType_HONGTAO.UInt8(), 1},
	{common.ECardType_HONGTAO.UInt8(), 2},
	{common.ECardType_HONGTAO.UInt8(), 3},
	{common.ECardType_HONGTAO.UInt8(), 4},
	{common.ECardType_HONGTAO.UInt8(), 5},
	{common.ECardType_HONGTAO.UInt8(), 6},
	{common.ECardType_HONGTAO.UInt8(), 7},
	{common.ECardType_HONGTAO.UInt8(), 8},
	{common.ECardType_HONGTAO.UInt8(), 9},
	{common.ECardType_HONGTAO.UInt8(), 10},
	{common.ECardType_HONGTAO.UInt8(), 11},
	{common.ECardType_HONGTAO.UInt8(), 12},
	{common.ECardType_HONGTAO.UInt8(), 13},
}

// 洗牌, 按照洗牌规则
// 参数0: 不带大小王
// 参数1: 带大小王
func NiuNiu_ShuffleCard(nHaveJoker uint8) []common.Card_info {
	rand.Seed(utils.SecondTimeSince1970())
	var sCard []common.Card_info
	if nHaveJoker == 0 {
		sCard = append(sCard, nn_cards_one...)
	} else {
		sCard = append(sCard, nn_cards_two...)
	}
	utils.RandomSlice(sCard)
	return sCard
}

// 从所有牌中发出5张牌
// 第一参数: 游戏类型
// 第二参数: 所有牌
// 第三参数: 是否允许中奖
// 第四参数: 中奖配置
// 第一返回: 剩余所有牌
// 第二返回: 一个玩家的牌 (该组牌已排序, 从大到小)
// 第三返回: 该玩家牌型
func NiuNiu_FiringOnePlayerCard(nGameType uint8, sAllCard []common.Card_info, isAllowAward bool, mAwardScale map[int]int) (sAll, sOne []common.Card_info, nType uint8) {
	if isAllowAward == true {
		sOne := sAllCard[:5]
		sAllCard = sAllCard[5:]
		nCardType := NiuNiu_CalcCardTypeAndSort(sOne, nGameType)
		return sAllCard, sOne, nCardType
	} else {
		nChangeCount := 0
		var sOne []common.Card_info
		var nCardType uint8
		for {
			nChangeCount++
			sOne = sAllCard[:5]
			nCardType = NiuNiu_CalcCardTypeAndSort(sOne, nGameType)
			nAwardScale := mAwardScale[int(nCardType)]
			if nAwardScale == 0 || nChangeCount > 10 {
				break
			} else {
				utils.RandomSlice(sAllCard)
			}
		}
		sAllCard = sAllCard[5:]
		return sAllCard, sOne, nCardType
	}
}

func NiuNiu_CardToString(sCard []common.Card_info, nIndex uint8) string {
	tOne := sCard[0]
	tTwo := sCard[1]
	tThree := sCard[2]
	tFour := sCard[3]
	tFives := sCard[4]
	strText := fmt.Sprintf("%v :-> %v%2v,%v%2v,%v%2v,%v%2v,%v%2v", nIndex, common.ECardType(tOne[0]), nn_point_string[tOne[1]], common.ECardType(tTwo[0]), nn_point_string[tTwo[1]], common.ECardType(tThree[0]), nn_point_string[tThree[1]], common.ECardType(tFour[0]), nn_point_string[tFour[1]], common.ECardType(tFives[0]), nn_point_string[tFives[1]])
	return strText
}

// 判断传入牌组是否是顺子 (传入牌组需先排序)(传入参数中不能包含王)
// 返回true, 表示是顺子
func NiuNiu_IsShunZi(sCard []common.Card_info) bool {
	nCardLen := len(sCard)
	if nCardLen < 5 {
		return false
	}

	tOne := sCard[0]
	tTwo := sCard[1]
	tThree := sCard[2]
	tFour := sCard[3]
	tFives := sCard[4]
	if tOne[1] == 13 && tTwo[1] == 12 && tThree[1] == 11 && tFour[1] == 10 && tFives[1] == 1 {
		return true
	}
	isRet := Poker_Is_AKQJ10(sCard)
	return isRet
}

// 判断传入牌组是否是炸弹 (传入牌组需先排序)(传入参数中不能包含王)
// 返回true, 表示是炸弹
func NiuNiu_IsZhaDan(sCard []common.Card_info) bool {
	nCardLen := len(sCard)
	if nCardLen < 5 {
		return false
	}
	mCount := Poker_StatPointCount(sCard)
	for _, nCount := range mCount {
		if nCount >= 4 {
			return true
		}
	}
	return false
}

// 判断传入牌组是否是葫芦 (传入牌组需先排序)(传入参数中不能包含王)
// 返回true, 表示是葫芦
func NiuNiu_IsHuLu(sCard []common.Card_info) bool {
	nCardLen := len(sCard)
	if nCardLen < 5 {
		return false
	}

	tOne := sCard[0]
	tTwo := sCard[1]
	tThree := sCard[2]
	tFour := sCard[3]
	tFives := sCard[4]
	if tOne[1] == tTwo[1] && tTwo[1] == tThree[1] && tThree[1] != tFour[1] && tFour[1] == tFives[1] {
		return true
	} else if tOne[1] == tTwo[1] && tTwo[1] != tThree[1] && tThree[1] == tFour[1] && tFour[1] == tFives[1] {
		return true
	}
	return false
}

func NiuNiu_HaveNiu(sCard []common.Card_info, nGameType uint8) uint8 {
	nCardLen := len(sCard)
	if nCardLen < 5 {
		return 0
	}

	nCard1 := uint8(utils.Min(10, int(sCard[0][1])))
	nCard2 := uint8(utils.Min(10, int(sCard[1][1])))
	nCard3 := uint8(utils.Min(10, int(sCard[2][1])))
	nCard4 := uint8(utils.Min(10, int(sCard[3][1])))
	nCard5 := uint8(utils.Min(10, int(sCard[4][1])))
	nCommTail := uint8(0)
	nGoldTail := uint8(0)
	if nGameType == common.EGameTypeWUHUA_NIUNIU.Value() {
		if nCard1 == nCard2 && nCard2 == nCard3 {
			nGoldTail = nCard4 + nCard5
		} else if nCard1 == nCard2 && nCard2 == nCard4 {
			nGoldTail = nCard3 + nCard5
		} else if nCard1 == nCard2 && nCard2 == nCard5 {
			nGoldTail = nCard3 + nCard4
		} else if nCard1 == nCard3 && nCard3 == nCard4 {
			nGoldTail = nCard2 + nCard5
		} else if nCard1 == nCard3 && nCard3 == nCard5 {
			nGoldTail = nCard2 + nCard4
		} else if nCard1 == nCard4 && nCard4 == nCard5 {
			nGoldTail = nCard2 + nCard3
		} else if nCard2 == nCard3 && nCard3 == nCard4 {
			nGoldTail = nCard1 + nCard5
		} else if nCard2 == nCard3 && nCard3 == nCard5 {
			nGoldTail = nCard1 + nCard4
		} else if nCard2 == nCard4 && nCard4 == nCard5 {
			nGoldTail = nCard1 + nCard3
		} else if nCard3 == nCard4 && nCard4 == nCard5 {
			nGoldTail = nCard1 + nCard2
		}
		if nGoldTail > 0 {
			nGoldTail = nGoldTail % 10
			if nGoldTail == 0 {
				nGoldTail = 20
			} else {
				nGoldTail = gold_niu_type[nGoldTail]
			}
		}
	}

	if (nCard1+nCard2+nCard3)%10 == 0 {
		nCommTail = nCard4 + nCard5
	} else if (nCard1+nCard2+nCard4)%10 == 0 {
		nCommTail = nCard3 + nCard5
	} else if (nCard1+nCard2+nCard5)%10 == 0 {
		nCommTail = nCard3 + nCard4
	} else if (nCard1+nCard3+nCard4)%10 == 0 {
		nCommTail = nCard2 + nCard5
	} else if (nCard1+nCard3+nCard5)%10 == 0 {
		nCommTail = nCard2 + nCard4
	} else if (nCard1+nCard4+nCard5)%10 == 0 {
		nCommTail = nCard2 + nCard3
	} else if (nCard2+nCard3+nCard4)%10 == 0 {
		nCommTail = nCard1 + nCard5
	} else if (nCard2+nCard3+nCard5)%10 == 0 {
		nCommTail = nCard1 + nCard4
	} else if (nCard2+nCard4+nCard5)%10 == 0 {
		nCommTail = nCard1 + nCard3
	} else if (nCard3+nCard4+nCard5)%10 == 0 {
		nCommTail = nCard1 + nCard2
	}
	if nCommTail > 0 {
		nCommTail = nCommTail % 10
		if nCommTail == 0 {
			nCommTail = 19
		} else {
			nCommTail = comm_niu_type[nCommTail]
		}
	}
	nTail := utils.MaxInt(int(nGoldTail), int(nCommTail))
	return uint8(nTail)
}

func niuniu_BaseCalcCardType(sCard []common.Card_info, nGameType uint8, isOneColor bool) uint8 {
	isShunZi := NiuNiu_IsShunZi(sCard)
	if isShunZi == true && isOneColor == true {
		return common.NN_TONGHUASHUN_18.UInt8() // 同花顺
	}

	nTotal := uint32(0)
	nCount := uint8(0)
	for _, value := range sCard {
		nPoint := value[1]
		nTotal += uint32(nPoint)
		if nPoint > 10 {
			nCount++
		}
	}

	// 开启特殊牌型
	if nGameType == common.EGameTypeWUHUA_NIUNIU.Value() {
		if nTotal >= 40 {
			return common.NN_SISHI_17.UInt8() // 四十
		}
	}

	if sCard[0][1] < 5 && nTotal <= 10 {
		return common.NN_WUXIAONIU_16.UInt8() // 五小牛
	}

	isZhaDan := NiuNiu_IsZhaDan(sCard)
	if isZhaDan == true {
		return common.NN_ZHADAN_15.UInt8() // 炸弹
	}

	isHuLu := NiuNiu_IsHuLu(sCard)
	if isHuLu == true {
		return common.NN_HULU_14.UInt8() // 葫芦
	}

	if isOneColor == true {
		return common.NN_TONGHUA_13.UInt8() // 同花
	}

	if nCount >= 5 {
		return common.NN_WUHUANIU_12.UInt8() // 五花牛
	}

	if isShunZi == true {
		return common.NN_SHUNZI_11.UInt8() // 顺子
	}

	nTail := NiuNiu_HaveNiu(sCard, nGameType)
	return uint8(nTail)
}

// 函数作用: 判断传入的牌切片对应的牌型
// 第一参数: 牌
// 第二参数: 游戏类型
// 第三参数: 是否开启同花顺等特殊牌型 (1开启, 0关闭)
func NiuNiu_CalcCardTypeAndSort(sCard []common.Card_info, nGameType uint8) uint8 {

	Poker_SortCard(sCard)

	nJokerCount := uint8(0)
	if nGameType == common.EGameTypeTEN_NIU_NIU.Value() {
		nJokerCount = Poker_GetJokerCount(sCard)
		if nJokerCount == 1 {
			sNewCard := []common.Card_info{sCard[0], sCard[1], sCard[2], sCard[3]}
			nRet := niuniu_CalcHaveJokerCardType(sNewCard, nGameType, nJokerCount)
			return nRet
		} else if nJokerCount == 2 {
			sNewCard := []common.Card_info{sCard[0], sCard[1], sCard[2]}
			nRet := niuniu_CalcHaveJokerCardType(sNewCard, nGameType, nJokerCount)
			return nRet
		}
	}

	isOneColor := Poker_IsOneColor(sCard)
	nRet := niuniu_BaseCalcCardType(sCard, nGameType, isOneColor)
	return nRet
}

func NiuNiu_CompareResults(nOneType uint8, tOneCard []common.Card_info, nTwoType uint8, tTwoCard []common.Card_info) uint8 {

	if nOneType == nTwoType {
		// AKQJ10排序后为KQJ10A 故增加下面特殊判断
		// A和B牌型同为同花顺或同为顺子才判断A牌的花色
		if nTwoType == common.NN_TONGHUASHUN_18.UInt8() || nTwoType == common.NN_SHUNZI_11.UInt8() {
			if tOneCard[0][1] == 13 && tOneCard[4][1] == 1 && tTwoCard[0][1] == 13 && tTwoCard[4][1] == 1 {
				if tOneCard[4][0] > tTwoCard[4][0] {
					return 1
				} else {
					return 2
				}
			}
		}
	}

	if nOneType > nTwoType {
		return 1
	} else if nOneType < nTwoType {
		return 2
	} else {
		if tOneCard[0][1] > tTwoCard[0][1] {
			return 1
		} else if tOneCard[0][1] < tTwoCard[0][1] {
			return 2
		} else {
			if tOneCard[0][0] > tTwoCard[0][0] {
				return 1
			} else {
				return 2
			}
		}
	}
}

func niuniu_CalcHaveJokerCardType(sNotHaveJokerCard []common.Card_info, nGameType, nJokerCount uint8) uint8 {

	isOneColor := Poker_IsOneColor(sNotHaveJokerCard)
	nMaxCardType := uint8(0)
	nCardLen := len(sNotHaveJokerCard)
	nRet := uint8(0)
	if nJokerCount == 1 {
		for i := 0; i < nCardLen; i++ {
			curr := sNotHaveJokerCard[i]
			check_color := curr[0]
			for j := int8(-1); j <= 1; j++ {
				check_point := uint8(int8(curr[1]) + j)
				if check_point >= 1 && check_point <= 13 {
					check_one := common.Card_info{check_color, check_point}
					check_card := sNotHaveJokerCard[:]
					check_card = append(check_card, check_one)
					Poker_SortCard(check_card)
					nRet = niuniu_BaseCalcCardType(check_card, nGameType, isOneColor)
					if nMaxCardType < nRet {
						nMaxCardType = nRet
					}
				}
			}
		}
		if nMaxCardType < 19 {
			nCard1 := uint8(utils.Min(10, int(sNotHaveJokerCard[0][1])))
			nCard2 := uint8(utils.Min(10, int(sNotHaveJokerCard[1][1])))
			nCard3 := uint8(utils.Min(10, int(sNotHaveJokerCard[2][1])))
			nCard4 := uint8(utils.Min(10, int(sNotHaveJokerCard[3][1])))
			// 一张王的情况下, 剩余四张牌中有3张能组成牛的, 则为牛牛牌型
			if (nCard1+nCard2+nCard3)%10 == 0 || (nCard1+nCard2+nCard4)%10 == 0 || (nCard1+nCard3+nCard4)%10 == 0 || (nCard2+nCard3+nCard4)%10 == 0 {
				return 19
			}
			// 一张王的情况下, 剩余四张牌中有2张能组成牛的, 则为牛牛牌型
			if (nCard1+nCard2)%10 == 0 || (nCard1+nCard3)%10 == 0 || (nCard1+nCard4)%10 == 0 || (nCard2+nCard3)%10 == 0 || (nCard2+nCard4)%10 == 0 || (nCard3+nCard4)%10 == 0 {
				return 19
			}

			nMax := 0
			nMax = utils.MaxInt(nMax, int(nCard1%10+nCard2%10)%10)
			nMax = utils.MaxInt(nMax, int(nCard1%10+nCard3%10)%10)
			nMax = utils.MaxInt(nMax, int(nCard1%10+nCard4%10)%10)
			nMax = utils.MaxInt(nMax, int(nCard2%10+nCard3%10)%10)
			nMax = utils.MaxInt(nMax, int(nCard2%10+nCard4%10)%10)
			nMax = utils.MaxInt(nMax, int(nCard3%10+nCard4%10)%10)
			nRet = uint8(nMax)
			if nRet == 0 {
				nRet = 19
			} else {
				nRet = comm_niu_type[nRet]
			}
			if nMaxCardType < nRet {
				nMaxCardType = nRet
			}
		}
		return nMaxCardType
	} else if nJokerCount == 2 {
		for i := 0; i < nCardLen; i++ {
			curr := sNotHaveJokerCard[i]
			check_color := curr[0]
			for j := int8(-2); j <= 1; j++ {
				check_point_j := uint8(int8(curr[1]) + j)
				for k := int8(-1); k <= 2; k++ {
					check_point_k := uint8(int8(curr[1]) + k)
					if check_point_j >= 1 && check_point_j <= 13 && check_point_k >= 1 && check_point_k <= 13 {
						check_one := common.Card_info{check_color, check_point_j}
						check_two := common.Card_info{check_color, check_point_k}
						check_card := sNotHaveJokerCard[:]
						check_card = append(check_card, check_one, check_two)
						Poker_SortCard(check_card)
						nRet = niuniu_BaseCalcCardType(check_card, nGameType, isOneColor)
						if nMaxCardType < nRet {
							nMaxCardType = nRet
						}
					}
				}
			}
		}
		if nMaxCardType < 19 {
			return 19
		}
		return nMaxCardType
	} else {
		return 0
	}
}
