package algorithm

import (
	"root/common"
	"root/core/log"
	"root/core/utils"
	"math/rand"
)

var PDK_DY_FIRST_CARD = common.Card_info{common.ECardType_HEITAO.UInt8(), 5}
var PDK_DY_FIXED_RESERVATION_CARD = common.Card_info{common.ECardType_FANGKUAI.UInt8(), 5}

const PDK_DY_MAX_CARDS = 10

var pdk_dy_cards_one = []common.Card_info{
	{common.ECardType_FANGKUAI.UInt8(), 5},
	{common.ECardType_FANGKUAI.UInt8(), 6},
	{common.ECardType_FANGKUAI.UInt8(), 7},
	{common.ECardType_FANGKUAI.UInt8(), 8},
	{common.ECardType_FANGKUAI.UInt8(), 9},
	{common.ECardType_FANGKUAI.UInt8(), 10},
	{common.ECardType_FANGKUAI.UInt8(), 11},
	{common.ECardType_FANGKUAI.UInt8(), 12},
	{common.ECardType_FANGKUAI.UInt8(), 13},
	{common.ECardType_FANGKUAI.UInt8(), 14},

	{common.ECardType_MEIHUA.UInt8(), 5},
	{common.ECardType_MEIHUA.UInt8(), 6},
	{common.ECardType_MEIHUA.UInt8(), 7},
	{common.ECardType_MEIHUA.UInt8(), 8},
	{common.ECardType_MEIHUA.UInt8(), 9},
	{common.ECardType_MEIHUA.UInt8(), 10},
	{common.ECardType_MEIHUA.UInt8(), 11},
	{common.ECardType_MEIHUA.UInt8(), 12},
	{common.ECardType_MEIHUA.UInt8(), 13},
	{common.ECardType_MEIHUA.UInt8(), 14},

	{common.ECardType_HONGTAO.UInt8(), 5},
	{common.ECardType_HONGTAO.UInt8(), 6},
	{common.ECardType_HONGTAO.UInt8(), 7},
	{common.ECardType_HONGTAO.UInt8(), 8},
	{common.ECardType_HONGTAO.UInt8(), 9},
	{common.ECardType_HONGTAO.UInt8(), 10},
	{common.ECardType_HONGTAO.UInt8(), 11},
	{common.ECardType_HONGTAO.UInt8(), 12},
	{common.ECardType_HONGTAO.UInt8(), 13},
	{common.ECardType_HONGTAO.UInt8(), 14},

	{common.ECardType_HEITAO.UInt8(), 6},
	{common.ECardType_HEITAO.UInt8(), 7},
	{common.ECardType_HEITAO.UInt8(), 8},
	{common.ECardType_HEITAO.UInt8(), 9},
	{common.ECardType_HEITAO.UInt8(), 10},
	{common.ECardType_HEITAO.UInt8(), 11},
	{common.ECardType_HEITAO.UInt8(), 12},
	{common.ECardType_HEITAO.UInt8(), 13},
	{common.ECardType_HEITAO.UInt8(), 14},
}

var pdk_dy_cards_two = []common.Card_info{
	{common.ECardType_MEIHUA.UInt8(), 5},
	{common.ECardType_MEIHUA.UInt8(), 6},
	{common.ECardType_MEIHUA.UInt8(), 7},
	{common.ECardType_MEIHUA.UInt8(), 8},
	{common.ECardType_MEIHUA.UInt8(), 9},
	{common.ECardType_MEIHUA.UInt8(), 10},
	{common.ECardType_MEIHUA.UInt8(), 11},
	{common.ECardType_MEIHUA.UInt8(), 12},
	{common.ECardType_MEIHUA.UInt8(), 13},

	{common.ECardType_HONGTAO.UInt8(), 5},
	{common.ECardType_HONGTAO.UInt8(), 6},
	{common.ECardType_HONGTAO.UInt8(), 7},
	{common.ECardType_HONGTAO.UInt8(), 8},
	{common.ECardType_HONGTAO.UInt8(), 9},
	{common.ECardType_HONGTAO.UInt8(), 10},
	{common.ECardType_HONGTAO.UInt8(), 11},
	{common.ECardType_HONGTAO.UInt8(), 12},
	{common.ECardType_HONGTAO.UInt8(), 13},
	{common.ECardType_HONGTAO.UInt8(), 14},

	{common.ECardType_HEITAO.UInt8(), 6},
	{common.ECardType_HEITAO.UInt8(), 7},
	{common.ECardType_HEITAO.UInt8(), 8},
	{common.ECardType_HEITAO.UInt8(), 9},
	{common.ECardType_HEITAO.UInt8(), 10},
	{common.ECardType_HEITAO.UInt8(), 11},
	{common.ECardType_HEITAO.UInt8(), 12},
	{common.ECardType_HEITAO.UInt8(), 13},
	{common.ECardType_HEITAO.UInt8(), 14},
}

// 洗牌, 按照洗牌规则
// 规则一: 总牌40张; 固定保留黑桃5, 其余39张牌中随机29张
// 规则二: 总牌40张; 去掉梅花A和所有方块; 固定保留黑桃5和方块5
func PaoDeKuai_DY_ShuffleCard(nShuffleMode uint8) []common.Card_info {

	rand.Seed(utils.SecondTimeSince1970())
	var sCard []common.Card_info
	if nShuffleMode == 1 {
		sCard = append(sCard, pdk_dy_cards_one[:29]...)
		sCard = append(sCard, PDK_DY_FIRST_CARD)
	} else {
		sCard = append(sCard, pdk_dy_cards_two...)
		sCard = append(sCard, PDK_DY_FIXED_RESERVATION_CARD)
		sCard = append(sCard, PDK_DY_FIRST_CARD)
	}
	utils.RandomSlice(sCard)
	return sCard
}

// 函数作用: 判断传入的牌切片对应的牌型
func PaoDeKuai_DY_CalcCardType(sCard []common.Card_info) common.EPaoDeKuai {

	nLen := len(sCard)
	if nLen <= 0 || nLen > 10 {
		log.Errorf("传入切片长度异常: %v", nLen)
		return common.PDK_NIL // 无牌型
	}

	if nLen == 1 {
		return common.PDK_DAN_ZHANG // 单张
	}

	c1 := sCard[0]
	c2 := sCard[1]
	if nLen == 2 {
		if c1[1] == c2[1] {
			return common.PDK_DUI_ZI // 对子
		}
		return common.PDK_NIL // 无牌型
	}

	c3 := sCard[2]
	if nLen == 3 {
		if c1[1] == c2[1] && c2[1] == c3[1] {
			return common.PDK_SAN_ZHANG // 三筒
		} else if c1[1]-1 == c2[1] && c2[1]-1 == c3[1] {
			return common.PDK_SHUN_ZI // 顺子
		}
		return common.PDK_NIL // 无牌型
	}

	c4 := sCard[3]
	if nLen == 4 {
		if c1[1] == c2[1] && c2[1] == c3[1] && c3[1] == c4[1] {
			return common.PDK_ZHA_DAN // 炸弹
		} else if c1[1]-1 == c2[1] && c2[1]-1 == c3[1] && c3[1]-1 == c4[1] {
			return common.PDK_SHUN_ZI // 顺子
		} else if c1[1] == c2[1] && c3[1] == c4[1] && c2[1]-1 == c3[1] {
			return common.PDK_SHUANG_SHUN // 连对
		}
		return common.PDK_NIL // 无牌型
	}

	if nLen%2 == 0 && c1[1] == c2[1] {
		isLianDui := true
		for j := 2; j < nLen-1; j += 2 {
			sNodeLast := sCard[j-1]
			sNodeCurr := sCard[j]
			sNodeNext := sCard[j+1]
			if sNodeLast[1]-1 != sNodeCurr[1] || sNodeCurr[1] != sNodeNext[1] {
				isLianDui = false
				break
			}
		}
		if isLianDui == true {
			return common.PDK_SHUANG_SHUN // 连对
		}
	}

	if c1[1]-1 == c2[1] {
		isShunZi := true
		for i := 1; i < nLen-1; i++ {
			sNodeCurr := sCard[i]
			sNodeNext := sCard[i+1]
			if sNodeCurr[1]-1 != sNodeNext[1] {
				isShunZi = false
				break
			}
		}
		if isShunZi == true {
			return common.PDK_SHUN_ZI // 顺子
		}
	}
	return common.PDK_NIL // 无牌型
}

// 判断能否关牌操作
func PaoDeKuai_DY_CanGuanPai(sCard []common.Card_info) bool {
	mCount := Poker_StatPointCount(sCard)
	if mCount[5] >= 4 || mCount[14] >= 4 {
		return true
	}
	return false
}

func PaoDeKuai_DY_OneIsBiggerTwo(sCardOne, sCardTwo []common.Card_info) bool {

	nOneType := PaoDeKuai_DY_CalcCardType(sCardOne)
	nTwoType := PaoDeKuai_DY_CalcCardType(sCardTwo)
	if nOneType == common.PDK_NIL || nTwoType == common.PDK_NIL {
		log.Error("无牌型无法比较大小")
		return false
	}

	nOneLen := len(sCardOne)
	nTwoLen := len(sCardTwo)
	if nOneType == nTwoType {
		if nOneLen == nTwoLen {
			if sCardOne[0][1] > sCardTwo[0][1] {
				return true
			}
		} else {
			log.Warnf("牌型相同的情况下, 牌的数量不相同, One:%v  Two:%v", sCardOne, sCardTwo)
			return false
		}
	} else if nOneType == common.PDK_ZHA_DAN && nTwoType < common.PDK_ZHA_DAN {
		return true
	}
	return false
}

func PaoDeKuai_DY_IsHaveBiggerCard(sCard, sInCard []common.Card_info) bool {

	//nCardLen := uint8(len(sCard))
	nInLen := uint8(len(sInCard))
	if nInLen < 1 {
		return false
	}
	nInPoint := sInCard[0][1]
	mHandCount := Poker_StatPointCount(sCard)
	eInType := PaoDeKuai_DY_CalcCardType(sInCard)
	switch eInType {
	case common.PDK_DAN_ZHANG: // 单张
		eEndCard := sCard[0]
		if eEndCard[1] > nInPoint {
			return true
		}
		// 从大到小的方式, 单张找不到, 再找炸弹
		nOutPoint := pdk_DY_GetOutPointByCount(sCard, mHandCount, 0, 4)
		if nOutPoint > 0 {
			return true
		}
	case common.PDK_DUI_ZI: // 对子
		nOutPoint := pdk_DY_GetOutPointByCount(sCard, mHandCount, nInPoint, 2)
		if nOutPoint > 0 {
			return true
		}
	case common.PDK_SAN_ZHANG: // 三筒
		nOutPoint := pdk_DY_GetOutPointByCount(sCard, mHandCount, nInPoint, 3)
		if nOutPoint > 0 {
			return true
		}
	case common.PDK_SHUN_ZI: // 顺子
		nStartPoint, nEndPoint := pdk_DY_GetCardByShunZi(sCard, mHandCount, nInPoint, nInLen, 1)
		if nStartPoint > 0 && nEndPoint > 0 {
			return true
		}
	case common.PDK_SHUANG_SHUN: // 连对
		nStartPoint, nEndPoint := pdk_DY_GetCardByShunZi(sCard, mHandCount, nInPoint, nInLen, 2)
		if nStartPoint > 0 && nEndPoint > 0 {
			return true
		}
	case common.PDK_ZHA_DAN: // 炸弹
		nOutPoint := pdk_DY_GetOutPointByCount(sCard, mHandCount, nInPoint, 4)
		if nOutPoint > 0 {
			return true
		}
	}
	return false
}

func PaoDeKuai_DY_GetBigCard(sCard, sInCard []common.Card_info, isBigToSmall bool) []common.Card_info {

	nCardLen := uint8(len(sCard))
	nInLen := uint8(len(sInCard))
	if nInLen < 1 {
		return nil
	}

	nInPoint := sInCard[0][1]
	mHandCount := Poker_StatPointCount(sCard)
	eInType := PaoDeKuai_DY_CalcCardType(sInCard)
	switch eInType {
	case common.PDK_DAN_ZHANG: // 单张
		if isBigToSmall == true {
			eEndCard := sCard[0]
			if eEndCard[1] > nInPoint {
				sRet := []common.Card_info{eEndCard}
				return sRet
			}
			// 找不到单张的情况下, 找炸弹
			sRet := pdk_DY_BuildCardByPoint(sCard, mHandCount, 0, 4)
			return sRet
		} else {
			for i := int(nCardLen) - 1; i >= 0; i-- {
				card := sCard[i]
				if card[1] > nInPoint {
					sRet := []common.Card_info{card}
					return sRet
				}
			}
			// 找不到单张的情况下, 找炸弹
			sRet := pdk_DY_BuildCardByPoint(sCard, mHandCount, 0, 4)
			return sRet
		}
	case common.PDK_DUI_ZI: // 对子
		sRet := pdk_DY_BuildCardByPoint(sCard, mHandCount, nInPoint, 2)
		return sRet
	case common.PDK_SAN_ZHANG: // 三筒
		sRet := pdk_DY_BuildCardByPoint(sCard, mHandCount, nInPoint, 3)
		return sRet
	case common.PDK_SHUN_ZI: // 顺子
		sRet := pdk_DY_BuildCardByShunZi(sCard, mHandCount, nInPoint, nInLen, 1)
		return sRet
	case common.PDK_SHUANG_SHUN: // 连对
		sRet := pdk_DY_BuildCardByShunZi(sCard, mHandCount, nInPoint, nInLen, 2)
		return sRet
	case common.PDK_ZHA_DAN: // 炸弹
		sRet := pdk_DY_BuildCardByPoint(sCard, mHandCount, nInPoint, 4)
		return sRet
	}
	return nil
}

// 连续添加多张同点数的牌
func pdk_DY_GetOutPointByCount(sCard []common.Card_info, mHandCount map[uint8]int, nInPoint uint8, nPerCount int) uint8 {
	nOutPoint := uint8(0)
	for point, count := range mHandCount {
		if count >= nPerCount && point > nInPoint {
			if nOutPoint < point {
				nOutPoint = point
			}
		}
	}
	return nOutPoint
}

// 连续添加多张同点数的牌
func pdk_DY_BuildCardByPoint(sCard []common.Card_info, mHandCount map[uint8]int, nInPoint uint8, nPerCount int) []common.Card_info {

	nOutPoint := pdk_DY_GetOutPointByCount(sCard, mHandCount, nInPoint, nPerCount)
	if nOutPoint <= 0 {
		return nil
	}

	nAddCount := nPerCount
	var sRet []common.Card_info
	nCardLen := len(sCard)
	for i := 0; i < nCardLen; i++ {
		card := sCard[i]
		if card[1] == nOutPoint {
			sRet = append(sRet, card)
			nAddCount--
			if nAddCount <= 0 {
				return sRet
			}
		}
	}
	return nil
}

// 查找顺子牌型, 每张牌型添加指定张数
// 可查找类型: 11, 22, 33, 44
// 可查找类型: 1, 2, 3, 4, 5
func pdk_DY_GetCardByShunZi(sCard []common.Card_info, mHandCount map[uint8]int, nStart, nInLen uint8, nPerCount int) (uint8, uint8) {
	nStartPoint := nStart + 1
	nEndPoint := nStartPoint + 1 - nInLen/uint8(nPerCount)
	if nStartPoint > 14 {
		return 0, 0
	}
	nCheck := uint8(0)
	for nStartPoint <= 14 {
		nCheck = 0
		for point := nStartPoint; point >= nEndPoint; point-- {
			if mHandCount[point] < nPerCount {
				break
			} else {
				nCheck++
			}
		}
		if nCheck*uint8(nPerCount) >= nInLen {
			break
		}
		nStartPoint++
		nEndPoint++
	}
	if nCheck*uint8(nPerCount) < nInLen {
		return 0, 0
	}
	return nStartPoint, nEndPoint
}

// 查找顺子牌型, 每张牌型添加指定张数
// 可查找类型: 11, 22, 33, 44
// 可查找类型: 1, 2, 3, 4, 5
func pdk_DY_BuildCardByShunZi(sCard []common.Card_info, mHandCount map[uint8]int, nStart, nInLen uint8, nPerCount int) []common.Card_info {

	nStartPoint, nEndPoint := pdk_DY_GetCardByShunZi(sCard, mHandCount, nStart, nInLen, nPerCount)
	if nStartPoint <= 0 || nEndPoint <= 0 {
		return nil
	}

	nCardLen := uint8(len(sCard))
	var sRet []common.Card_info
	mOutCount := make(map[uint8]int)
	for i := uint8(0); i < nCardLen; i++ {
		if sRet != nil && len(sRet) >= int(nInLen) {
			break
		}
		card := sCard[i]
		point := card[1]
		if point >= nEndPoint && point <= nStartPoint {
			if count, isExist := mOutCount[point]; isExist == false {
				mOutCount[point] = 1
				sRet = append(sRet, card)
			} else {
				if count < nPerCount {
					sRet = append(sRet, card)
				}
				mOutCount[point]++
			}
		}
	}
	return sRet
}
