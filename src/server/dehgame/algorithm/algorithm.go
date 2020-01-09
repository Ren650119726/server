package algorithm

import (
	"root/common"
	"root/core/log"
	"root/core/utils"
	"math/rand"
	"root/server/dehgame/types"
)

var cards = []common.Card_info{
	{common.ECardType_HEITAO.UInt8(), 4},
	{common.ECardType_HEITAO.UInt8(), 5},
	{common.ECardType_HEITAO.UInt8(), 6},
	{common.ECardType_HEITAO.UInt8(), 7},
	{common.ECardType_HEITAO.UInt8(), 8},
	{common.ECardType_HEITAO.UInt8(), 9},
	{common.ECardType_HEITAO.UInt8(), 10},
	{common.ECardType_HEITAO.UInt8(), 11},

	{common.ECardType_HONGTAO.UInt8(), 2},
	{common.ECardType_HONGTAO.UInt8(), 3},
	{common.ECardType_HONGTAO.UInt8(), 4},
	{common.ECardType_HONGTAO.UInt8(), 5},
	{common.ECardType_HONGTAO.UInt8(), 6},
	{common.ECardType_HONGTAO.UInt8(), 7},
	{common.ECardType_HONGTAO.UInt8(), 8},
	{common.ECardType_HONGTAO.UInt8(), 9},
	{common.ECardType_HONGTAO.UInt8(), 10},
	{common.ECardType_HONGTAO.UInt8(), 12},

	{common.ECardType_MEIHUA.UInt8(), 4},
	{common.ECardType_MEIHUA.UInt8(), 6},
	{common.ECardType_MEIHUA.UInt8(), 7},
	{common.ECardType_MEIHUA.UInt8(), 8},
	{common.ECardType_MEIHUA.UInt8(), 10},
	{common.ECardType_MEIHUA.UInt8(), 11},

	{common.ECardType_FANGKUAI.UInt8(), 2},
	{common.ECardType_FANGKUAI.UInt8(), 4},
	{common.ECardType_FANGKUAI.UInt8(), 6},
	{common.ECardType_FANGKUAI.UInt8(), 7},
	{common.ECardType_FANGKUAI.UInt8(), 8},
	{common.ECardType_FANGKUAI.UInt8(), 10},
	{common.ECardType_FANGKUAI.UInt8(), 12},

	{common.ECardType_JKEOR.UInt8(), 6},
}

// 随机获得不重复的n张牌
func GetRandom_Card(count int) []common.Card_info {

	if count > len(cards) {
		return nil
	}

	rand.Seed(utils.SecondTimeSince1970())
	ret := make([]common.Card_info, 0, count)
	for i := 0; i < count; i++ {
		last := len(cards) - 1 - i
		rand_val := utils.Randx_y(0, last)
		ret = append(ret, cards[rand_val])
		cards[rand_val], cards[last] = cards[last], cards[rand_val]
	}
	return ret
}

// 计算一张牌的主牌型
func CalcOneCardMainType(sCard common.Card_info) uint8 {
	switch sCard[0] {
	case common.ECardType_HONGTAO.UInt8(), common.ECardType_FANGKUAI.UInt8():
		switch sCard[1] {
		case 12:
			return types.BRAND_TIAN_PAI.Value()
		case 2:
			return types.BRAND_DI_PAI.Value()
		case 8:
			return types.BRAND_REN_PAI.Value()
		case 4:
			return types.BRAND_HE_PAI.Value()
		case 10, 6, 7:
			return types.BRAND_HU_SHI_MAO_GAO.Value()
		default:
			return types.BRAND_LAN_PAI.Value()
		}
	case common.ECardType_MEIHUA.UInt8(), common.ECardType_HEITAO.UInt8():
		switch sCard[1] {
		case 11:
			return types.BRAND_HU_SHI_MAO_GAO.Value()
		case 10, 6, 4:
			return types.BRAND_MEI_BAN_SAN.Value()
		default:
			return types.BRAND_LAN_PAI.Value()
		}
	default:
		return types.BRAND_LAN_PAI.Value()
	}
}

// 计算一个玩家的两张牌组的牌型 (头为一组, 尾为一组)
// 第一参数: 一组中的第一张牌 (头的第一张  或者   尾的第一张)
// 第二参数: 一组中的第二张牌 (头的第二张  或者   尾的第二张)
// 第三参数: 地九王是否算大牌, 算大牌传true; 反之传false
// 第一返回: 一组牌的牌型   (主牌型 + 牌点数)
func CalcOneSetCardType(sOneCard, sTwoCard common.Card_info, isOpenDiJiuWang bool) uint8 {

	nOneType := CalcOneCardMainType(sOneCard)
	nTwoType := CalcOneCardMainType(sTwoCard)

	if nOneType == nTwoType {
		switch nOneType {
		case types.BRAND_TIAN_PAI.Value():
			return types.BRAND_TIAN_DUI.Value()
		case types.BRAND_DI_PAI.Value():
			return types.BRAND_DI_DUI.Value()
		case types.BRAND_REN_PAI.Value():
			return types.BRAND_REN_DUI.Value()
		case types.BRAND_HE_PAI.Value():
			return types.BRAND_HE_DUI.Value()
		case types.BRAND_MEI_BAN_SAN.Value():
			if sOneCard[1] == sTwoCard[1] {
				return types.BRAND_MEI_BAN_SAN_DUI.Value()
			}
		case types.BRAND_HU_SHI_MAO_GAO.Value():
			if sOneCard[1] == sTwoCard[1] {
				return types.BRAND_HU_SHI_MAO_GAO_DUI.Value()
			}
		case types.BRAND_LAN_PAI.Value():
			if sOneCard[1] == sTwoCard[1] {
				return types.BRAND_LAN_DUI.Value()
			}
		}

		if (sOneCard[0] == common.ECardType_JKEOR.UInt8() && sTwoCard[1] == 3) || (sOneCard[1] == 3 && sTwoCard[0] == common.ECardType_JKEOR.UInt8()) {
			return types.BRAND_DING_ER_HUANG.Value()
		}

		nPoint := (sOneCard[1] + sTwoCard[1]) % 10
		if nPoint == 0 {
			// 尾点为0, 不区分主牌大小
			return 0
		}
		return nOneType + nPoint
	}

	if nOneType < nTwoType {
		nOneType, nTwoType = nTwoType, nOneType
		sOneCard, sTwoCard = sTwoCard, sOneCard
	}

	if nOneType == types.BRAND_TIAN_PAI.Value() {
		if sTwoCard[1] == 9 {
			return types.BRAND_TIAN_WANG.Value()
		} else if sTwoCard[1] == 8 {
			return types.BRAND_TIAN_GANG.Value()
		}
	} else if nOneType == types.BRAND_DI_PAI.Value() {

		if isOpenDiJiuWang == true && sTwoCard[1] == 9 {
			return types.BRAND_DI_WANG.Value()
		} else if sTwoCard[1] == 8 {
			return types.BRAND_DI_GANG.Value()
		}
	}

	nPoint2 := (sOneCard[1] + sTwoCard[1]) % 10
	if nPoint2 == 0 {
		// 尾点为0, 不区分主牌大小
		return 0
	}
	return nOneType + nPoint2
}

// 计算一个玩家的牌是否是三花十或者三花六  否则返回0
// 参数一: 一个玩家的所有牌
// 参数二: 需要检测特殊牌型的张数, 例如玩家发完第三张牌丢牌, 就传3; 否则传4;
// 返回值: 0表示无特殊排序;
// 返回值: types.BRAND_SAN_HUA_LIU表示三花六
// 返回值: types.BRAND_SAN_HUA_SHI表示三花十
func CalcSpecialCardType(sAll []common.Card_info, nCheckLen uint8) uint8 {
	if nCheckLen < 3 || nCheckLen > 4 {
		return 0
	}

	sCheckCount := []uint8{0, 0, 0, 0, 0, 0}
	for i := 0; i < int(nCheckLen); i++ {
		color := sAll[i][0]
		point := sAll[i][1]

		if color == common.ECardType_JKEOR.UInt8() {
			sCheckCount[0] = 1
		} else if point == 6 {
			if color == common.ECardType_HONGTAO.UInt8() || color == common.ECardType_FANGKUAI.UInt8() {
				sCheckCount[1] = 1
			} else if color == common.ECardType_MEIHUA.UInt8() || color == common.ECardType_HEITAO.UInt8() {
				sCheckCount[2] = 1
			}
		} else if point == 11 {
			sCheckCount[3] = 1
		} else if point == 10 {
			if color == common.ECardType_MEIHUA.UInt8() || color == common.ECardType_HEITAO.UInt8() {
				sCheckCount[4] = 1
			} else if color == common.ECardType_HONGTAO.UInt8() || color == common.ECardType_FANGKUAI.UInt8() {
				sCheckCount[5] = 1
			}
		}
	}

	if sCheckCount[0] > 0 && sCheckCount[1] > 0 && sCheckCount[2] > 0 {
		return types.BRAND_SAN_HUA_LIU.Value()
	} else if sCheckCount[3] > 0 && sCheckCount[4] > 0 && sCheckCount[5] > 0 {
		return types.BRAND_SAN_HUA_SHI.Value()
	}
	return 0
}

// 计算一个玩家4张牌的牌型; 会返回重新排序后的牌下标(修正为头必须大于尾);
// 第一参数: 玩家的4张牌, 可以是未分牌的4张牌
// 第二参数: 如果允许出现三花十或者三花六,长度应该传3,或者4; 否则传0
// 第三参数: 如果允许出现地九王, 传true; 否则传false
// 返回参数1: 头的牌型
// 返回参数2: 尾的牌型
// 返回参数3: 返回修正后的牌切片 (前2个是头牌, 后2个是尾牌)
func CalcOnePlayerCardType(sAll []common.Card_info, nCheckLen uint8, isOpenDiJiuWang bool) (uint8, uint8, []common.Card_info) {

	nSpecial := CalcSpecialCardType(sAll, nCheckLen)
	if nSpecial > 0 {
		return nSpecial, 0, sAll
	}

	one := sAll[0]
	two := sAll[1]
	three := sAll[2]
	four := sAll[3]

	nTouSet := CalcOneSetCardType(one, two, isOpenDiJiuWang)
	nWeiSet := CalcOneSetCardType(three, four, isOpenDiJiuWang)

	if nTouSet < types.BRAND_DI_GANG.Value() && nWeiSet < types.BRAND_DI_GANG.Value() {
		nTouPoint := nTouSet % 10
		nWeiPoint := nWeiSet % 10
		nTouMain := nTouSet - nTouPoint
		nWeiMain := nWeiSet - nWeiPoint
		if nTouPoint < nWeiPoint {
			sCard := []common.Card_info{three, four, one, two}
			return nWeiSet, nTouSet, sCard
		} else if nTouPoint > nWeiPoint {
			return nTouSet, nWeiSet, sAll
		} else {
			if nTouMain < nWeiMain {
				sCard := []common.Card_info{three, four, one, two}
				return nWeiSet, nTouSet, sCard
			} else {
				return nTouSet, nWeiSet, sAll
			}
		}
	} else {
		if nTouSet < nWeiSet {
			sCard := []common.Card_info{three, four, one, two}
			return nWeiSet, nTouSet, sCard
		} else {
			return nTouSet, nWeiSet, sAll
		}
	}
}

func CalcReceiveAward(nTouSet, nWeiSet uint8) types.ESpecialCard {
	if nTouSet == types.BRAND_DING_ER_HUANG.Value() {
		if nWeiSet == types.BRAND_TIAN_DUI.Value() {
			return types.SPECIAL_CARD_TIAN_DING
		} else if nWeiSet == types.BRAND_DI_DUI.Value() {
			return types.SPECIAL_CARD_DI_DING
		} else if nWeiSet >= types.BRAND_LAN_DUI.Value() {
			return types.SPECIAL_CARD_DUO_DING
		}
	} else if nTouSet >= types.BRAND_LAN_DUI.Value() && nWeiSet >= types.BRAND_LAN_DUI.Value() {
		return types.SPECIAL_CARD_DUO_DUO
	}
	return types.SPECIAL_CARD_NIL
}

var sCheckTurnIndex = [...][]uint8{
	0: {0, 1, 2, 3, 4, 5},
	1: {1, 2, 3, 4, 5, 0},
	2: {2, 3, 4, 5, 0, 1},
	3: {3, 4, 5, 0, 1, 2},
	4: {4, 5, 0, 1, 2, 3},
	5: {5, 0, 1, 2, 3, 4},
}

// 计算离庄家最近的玩家座位下标
// 第一参数: 庄家下标
// 第二参数: One玩家下标
// 第三参数: Two玩家下标
// 返回说明: 返回离庄家最近的座位下标 (One or Two)
func CalcFromBankerRecently(nBankerIndex, nOneIndex, nTwoIndex uint8) uint8 {
	sCheck := sCheckTurnIndex[nBankerIndex]
	for _, nValue := range sCheck {
		if nValue == nOneIndex {
			return nOneIndex
		}

		if nValue == nTwoIndex {
			return nTwoIndex
		}
	}
	return nOneIndex
}

// 计算玩家座位权重
// 第一参数: 庄家下标
// 第二参数: 玩家下标
// 返回说明: 返回玩家的座位权重 (返回值越小, 位置权重越大)
func CalcFromBankerPositionWeight(nBankerIndex, nPlayerIndex uint8) uint8 {
	sCheck := sCheckTurnIndex[nBankerIndex]
	for nIndex, nValue := range sCheck {
		if nValue == nPlayerIndex {
			return uint8(nIndex)
		}
	}
	return 99
}

// 自动分配, 将4张牌安装一定规律筛选最较为最优的头尾分牌
// 较为最优说明:
// 4张牌排除重复有3种组合  [1,2  3,4] [1,3  2,4] [1,4  2,3]
// 将3种组合分别计算各自的头尾牌型. 满足以下规则将返回该分牌序列
// 1. 头牌型中有地九王, 地杠, 天九王, 天杠, 对子, 丁二红等牌型的, 将头最大的序列返回
// 2. 若头牌中没有上面类型牌型的, 将尾最大的序列返回
func AutoFenPai(sAll []common.Card_info, nCheckLen uint8, isOpenDiJiuWang bool) []common.Card_info {

	sGroup1 := []common.Card_info{sAll[0], sAll[1], sAll[2], sAll[3]}
	nTouSet1, nWeiSet1, sNewCard1 := CalcOnePlayerCardType(sGroup1, nCheckLen, isOpenDiJiuWang)

	sGroup2 := []common.Card_info{sAll[0], sAll[2], sAll[1], sAll[3]}
	nTouSet2, nWeiSet2, sNewCard2 := CalcOnePlayerCardType(sGroup2, nCheckLen, isOpenDiJiuWang)

	sGroup3 := []common.Card_info{sAll[0], sAll[3], sAll[1], sAll[2]}
	nTouSet3, nWeiSet3, sNewCard3 := CalcOnePlayerCardType(sGroup3, nCheckLen, isOpenDiJiuWang)

	// 优先返回头大的牌组
	var nMaxTouSet uint8
	var sRetMaxTouCard []common.Card_info
	if (nTouSet1%10 >= 9 || nTouSet1 >= types.BRAND_DI_GANG.Value()) && nTouSet1 > nMaxTouSet {
		nMaxTouSet = nTouSet1
		sRetMaxTouCard = sNewCard1
	}
	if (nTouSet2%10 >= 9 || nTouSet2 >= types.BRAND_DI_GANG.Value()) && nTouSet2 > nMaxTouSet {
		nMaxTouSet = nTouSet2
		sRetMaxTouCard = sNewCard2
	}
	if (nTouSet3%10 >= 9 || nTouSet3 >= types.BRAND_DI_GANG.Value()) && nTouSet3 > nMaxTouSet {
		nMaxTouSet = nTouSet3
		sRetMaxTouCard = sNewCard3
	}
	if nMaxTouSet > 0 {
		return sRetMaxTouCard
	}

	// 其次选择尾大的牌组
	var nMaxWeiSet uint8
	var sRetMaxWeiCard []common.Card_info
	nRet := CompareCardSet(nWeiSet1, nWeiSet2)
	if nRet != 2 {
		sRetMaxWeiCard = sNewCard1
		nMaxWeiSet = nWeiSet1
	} else {
		sRetMaxWeiCard = sNewCard2
		nMaxWeiSet = nWeiSet2
	}

	nRet = CompareCardSet(nMaxWeiSet, nWeiSet3)
	if nRet == 2 {
		sRetMaxWeiCard = sNewCard3
	}
	return sRetMaxWeiCard
}

// 比较两个玩家的牌组 (头和头比, 尾和尾比)
// 返回1表示 One >  Two
// 返回2表示 One <  Two
// 返回0表示 One == Two
func CompareCardSet(nOneSet uint8, nTwoSet uint8) uint8 {

	var nRet uint8 = 0
	if nOneSet < types.BRAND_DI_GANG.Value() && nTwoSet < types.BRAND_DI_GANG.Value() {
		nPointA := nOneSet % 10
		nPointB := nTwoSet % 10
		nMainA := nOneSet - nPointA
		nMainB := nTwoSet - nPointB
		if nPointA > nPointB {
			nRet = 1
		} else if nPointA < nPointB {
			nRet = 2
		} else {
			if nMainA > nMainB {
				nRet = 1
			} else if nMainA < nMainB {
				nRet = 2
			} else {
				nRet = 0
			}
		}
	} else {
		if nOneSet > nTwoSet {
			nRet = 1
		} else if nOneSet < nTwoSet {
			nRet = 2
		} else {
			nRet = 0
		}
	}
	return nRet
}

// 比较两个玩家的头尾, 判断One是否大于Two
// 第一参数: 玩家1的4张牌
// 第二参数: 玩家1的位置权重
// 第三参数: 玩家2的4张牌
// 第四参数: 玩家2的位置权重
// 第五参数: 检测是否开启三花十或三花六, 若开启传3或者4; 否则传0
// 第六参数: 是否开启地九王为大牌, 开启传true; 否则传false
// 第七参数: 是否判断One的尾牌组的牌型是否大于Two
// 返回1表示 One >  Two
// 返回2表示 One <  Two
// 返回0表示 One == Two
func CompareTouWei(sOneAll []common.Card_info, nOneWeight uint8, sTwoAll []common.Card_info, nTwoWeight uint8, nCheckLen uint8, isOpenDiJiuWang bool) uint8 {

	nOneTouSet, nOneWeiSet, _ := CalcOnePlayerCardType(sOneAll, nCheckLen, isOpenDiJiuWang)
	nTwoTouSet, nTwoWeiSet, _ := CalcOnePlayerCardType(sTwoAll, nCheckLen, isOpenDiJiuWang)

	nTouRet := CompareCardSet(nOneTouSet, nTwoTouSet)
	nWeiRet := CompareCardSet(nOneWeiSet, nTwoWeiSet)

	if nTouRet == 0 {
		if nOneWeight < nTwoWeight {
			nTouRet = 1
		} else {
			nTouRet = 2
		}
	}

	if nWeiRet == 0 {
		if nOneWeight < nTwoWeight {
			nWeiRet = 1
		} else {
			nWeiRet = 2
		}
	}

	if nTouRet == 1 && nWeiRet == 1 {
		return 1
	} else if nTouRet == 2 && nWeiRet == 2 {
		return 2
	} else if nTouRet == 1 && nWeiRet == 2 {
		return 0
	} else if nTouRet == 2 && nWeiRet == 1 {
		return 0
	} else {
		log.Errorf("丁二红出现异常的比牌大小, 头:%v, 尾:%v, OneWeight:%v, TwoWeight:%v", nTouRet, nWeiRet, nOneWeight, nTwoWeight)
	}
	return 0
}

// one和two比较单张牌的主类型大小
// 返回1表示 one大于two
// 返回0表示 one等于two
// 返回2表示 one小于two
func CompareOneCardMainType(sOneCard common.Card_info, sTwoCard common.Card_info) uint8 {
	nOneType := CalcOneCardMainType(sOneCard)
	nTwoType := CalcOneCardMainType(sTwoCard)

	if nOneType < nTwoType {
		return 2
	} else if nOneType == nTwoType {
		return 0
	}
	return 1
}
