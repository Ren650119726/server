package algorithm

import (
	"root/common"
	"root/core/log"
	"sort"
)

/////////////////////////////////////////////////////////////////////////////
// 通用扑克牌算法
/////////////////////////////////////////////////////////////////////////////

// 扑克牌通用排序算法, 若牌组中有王, 王排到末尾
// 排序规则, 点数大的排到前面; 点数相同的, 花色大的排到前面;
func Poker_SortCard(sCard []common.Card_info) {
	sort.Slice(sCard, func(i, j int) bool {
		nColor1 := sCard[i][0]
		nPoint1 := sCard[i][1]
		nColor2 := sCard[j][0]
		nPoint2 := sCard[j][1]
		if nColor1 == common.ECardType_JKEOR.UInt8() && nColor2 != common.ECardType_JKEOR.UInt8() {
			return false
		} else if nColor1 != common.ECardType_JKEOR.UInt8() && nColor2 == common.ECardType_JKEOR.UInt8() {
			return true
		} else if nPoint1 > nPoint2 {
			return true
		} else if nPoint1 == nPoint2 {
			if nColor1 > nColor2 {
				return true
			}
		}
		return false
	})
}

// 统计传入牌组中点数对应的张数, (若牌组中有王, 王不参与统计计算)
// 返回说明: 		map[key]value
// key[uint8类型]:  表示牌点数
// value[int类型]:  表示有几张对应点数的牌
func Poker_StatPointCount(sCard []common.Card_info) map[uint8]int {
	mCount := make(map[uint8]int, 0)
	for _, node := range sCard {
		nPoint := node[1]
		nColor := node[0]
		if nColor == common.ECardType_JKEOR.UInt8() {
			continue
		}
		if _, isExist := mCount[nPoint]; isExist == false {
			mCount[nPoint] = 1
		} else {
			mCount[nPoint]++
		}
	}
	return mCount
}

// 删除手牌中指定牌组 (可传一张牌, 也可传一组牌)
// 返回删除牌后的剩余牌组
func Poker_RemoveCard(sCard []common.Card_info, sRemoveCard interface{}) []common.Card_info {
	nLen := len(sCard)
	switch sRemoveCard.(type) {
	case common.Card_info:
		sRemove := sRemoveCard.(common.Card_info)
		for i := 0; i < nLen; i++ {
			card := sCard[i]
			if card[0] == sRemove[0] && card[1] == sRemove[1] {
				sCard = append(sCard[:i], sCard[i+1:]...)
				return sCard
			}
		}
	case []common.Card_info:
		sRemoveList := sRemoveCard.([]common.Card_info)
		nRemoveLen := len(sRemoveList)
		var sNewCard []common.Card_info
		for i := 0; i < nLen; i++ {
			card := sCard[i]
			isInRemove := false
			for _, remove := range sRemoveList {
				if card[0] == remove[0] && card[1] == remove[1] {
					isInRemove = true
					break
				}
			}
			if isInRemove == false {
				sNewCard = append(sNewCard, card)
			}
		}
		if sNewCard == nil {
			sNewCard = []common.Card_info{}
		}
		nEndLen := nLen - nRemoveLen
		if len(sNewCard) != nEndLen {
			log.Errorf("删除指定牌后,应剩余:%v张 %v; 现剩余牌数不正确:%v张 %v", nEndLen, sRemoveList, len(sNewCard), sNewCard)
		}
		return sNewCard
	default:
		log.Errorf("错误的参数类型:%v", sRemoveCard)
		return sCard
	}
	return sCard
}

// 函数作用: 牌中是否包含指定的牌
// 第一参数: 待检测的牌切片
// 第二参数: 指定的牌
// 返回说明: 返回true表示包含
func Poker_IsHaveCard(sCard []common.Card_info, sInCard interface{}) bool {
	nLen := len(sCard)
	switch sInCard.(type) {
	case common.Card_info:
		eInCard := sInCard.(common.Card_info)
		for _, card := range sCard {
			if card[0] == eInCard[0] && card[1] == eInCard[1] {
				return true
			}
		}
		return false
	case []common.Card_info:
		sIn := sInCard.([]common.Card_info)
		nHave := 0
		for _, one := range sIn {
			for i := 0; i < nLen; i++ {
				card := sCard[i]
				if card[0] == one[0] && card[1] == one[1] {
					nHave++
					break
				}
			}
		}
		if nHave == len(sIn) {
			return true
		}
		return false
	default:
		log.Errorf("错误的参数类型:%v", sInCard)
		return false
	}
}

// 获取传入牌组中有几张王
func Poker_GetJokerCount(sCard []common.Card_info) uint8 {
	nCount := uint8(0)
	for _, node := range sCard {
		nColor := node[0]
		if nColor == common.ECardType_JKEOR.UInt8() {
			nCount++
		}
	}
	return nCount
}

// 判断传入牌组是否是同花色, (若牌组中有王, 王不参与判断)
// 返回true, 表示同花色
func Poker_IsOneColor(sCard []common.Card_info) bool {
	nColor := sCard[0][0]
	for i := 1; i < len(sCard); i++ {
		if nColor == common.ECardType_JKEOR.UInt8() {
			// 排除掉王
			continue
		}
		if nColor != sCard[i][0] {
			return false
		}
	}
	return true
}

// 判断传入牌组是否是顺子, (牌组中不包含王)
// 返回true, 表示顺子
func Poker_Is_AKQJ10(sCard []common.Card_info) bool {
	nLen := len(sCard)
	for i := 0; i < nLen-1; i++ {
		tOne := sCard[i]
		tNext := sCard[i+1]
		if tOne[0] == common.ECardType_JKEOR.UInt8() || tNext[0] == common.ECardType_JKEOR.UInt8() {
			log.Errorf("传入牌组中包含了王, 牌组长度:%v; 牌组:%v", nLen, sCard)
			return false
		}
		if tOne[1]-1 != tNext[1] {
			return false
		}
	}
	return true
}

// 判断传入牌组是否是双顺, (牌组中不包含王)
// 返回true, 表示是双顺
func Poker_Is_AAKK(sCard []common.Card_info) bool {
	nLen := len(sCard)
	c1 := sCard[0]
	c2 := sCard[1]
	if nLen%2 == 0 && c1[1] == c2[1] {
		for j := 2; j < nLen-1; j += 2 {
			sNodeLast := sCard[j-1]
			sNodeCurr := sCard[j]
			sNodeNext := sCard[j+1]
			if sNodeLast[0] == common.ECardType_JKEOR.UInt8() || sNodeCurr[0] == common.ECardType_JKEOR.UInt8() || sNodeNext[0] == common.ECardType_JKEOR.UInt8() {
				log.Errorf("传入牌组中包含了王, 牌组长度:%v; 牌组:%v", nLen, sCard)
				return false
			}
			if sNodeLast[1]-1 != sNodeCurr[1] || sNodeCurr[1] != sNodeNext[1] {
				return false
			}
		}
		return true
	}
	return false
}

// 判断传入牌组是否是三顺, (牌组中不包含王)
// 返回true, 表示是三顺
func Poker_Is_AAAKKK(sCard []common.Card_info) bool {
	nLen := len(sCard)
	c1 := sCard[0]
	c2 := sCard[1]
	c3 := sCard[2]
	if nLen%3 == 0 && c1[1] == c2[1] && c2[1] == c3[1] {
		for j := 3; j < nLen-2; j += 3 {
			sNode1 := sCard[j-1]
			sNode2 := sCard[j]
			sNode3 := sCard[j+1]
			sNode4 := sCard[j+2]
			if sNode1[0] == common.ECardType_JKEOR.UInt8() || sNode2[0] == common.ECardType_JKEOR.UInt8() || sNode3[0] == common.ECardType_JKEOR.UInt8() || sNode4[0] == common.ECardType_JKEOR.UInt8() {
				log.Errorf("传入牌组中包含了王, 牌组长度:%v; 牌组:%v", nLen, sCard)
				return false
			}
			if sNode1[1]-1 != sNode2[1] || sNode2[1] != sNode3[1] || sNode3[1] != sNode4[1] {
				return false
			}
		}
		return true
	}
	return false
}

// 判断传入int切片组是否是顺子, (牌组中不包含王)
// 返回true, 表示顺子
func Ints_Is_ShunZi(sCard []int) bool {
	if sCard == nil {
		return false
	}
	nLen := len(sCard)
	if nLen < 2 {
		return false
	}

	for i := 0; i < nLen-1; i++ {
		nOne := sCard[i]
		nNext := sCard[i+1]
		if nOne-1 != nNext {
			return false
		}
	}
	return true
}
