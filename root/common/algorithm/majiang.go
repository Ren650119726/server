package algorithm

import (
	"root/common"
	"root/core/log"
	"root/core/utils"
	"fmt"
	"sort"
)

type (
	Majiang_Sort struct {
		Cards []common.EMaJiangType
	}

	Majiang_Hu struct {
		HuType common.EMaJiangHu              // 麻将胡牌类型
		Extra  map[common.EMaJiangExtra]uint8 // 胡牌额外加翻类型和对应加几翻, 无额外加翻是nil
		Total  uint8                          // 总额外番数
	}
)

func (self *Majiang_Sort) Len() int {
	return len(self.Cards)
}
func (self *Majiang_Sort) Less(i, j int) bool {
	return self.Cards[i] < self.Cards[j]
}
func (self *Majiang_Sort) Swap(i, j int) {
	self.Cards[i], self.Cards[j] = self.Cards[j], self.Cards[i]
}

func (self *Majiang_Hu) String() string {
	return fmt.Sprintf("胡牌:%v 额外番-番数:%v 总额外番:%v", self.HuType, self.Extra, self.Total)
}

// 函数作用: 统计麻将切片中牌的个数
func MajiangStatCount(sSlice []common.EMaJiangType) map[uint8]int {
	mCount := make(map[uint8]int, 0)
	for _, value := range sSlice {
		nPoint := value.Value()
		if _, isExist := mCount[nPoint]; isExist == false {
			mCount[nPoint] = 1
		} else {
			mCount[nPoint]++
		}
	}
	return mCount
}

func dgk_CalcHuTypeByExclusion(sHand []common.EMaJiangType, sPeng, sGang [][]common.EMaJiangType, eDianPao common.EMaJiangType, mExclusion map[common.EMaJiangHu]bool) common.EMaJiangHu {
	nHandLen := len(sHand)
	if nHandLen != 2 && nHandLen != 5 && nHandLen != 8 && nHandLen != 11 && nHandLen != 14 {
		log.Errorf("手牌长度异常:%v, 正确长度应该是2,5,8,11,14之一", nHandLen)
		return common.HU_NIL
	}

	sTotal := majiang_BuildSlice(sHand, sPeng, sGang)
	mHandCount := MajiangStatCount(sHand)
	isQingYiSe := majiang_IsOnlyOneColor(sTotal)
	sCode := majiang_Marshal(sHand)
	mCodeCount := utils.StatElementCount(sCode)

	// 自摸才判断清五对和五对
	eWuDuiHuType := common.HU_NIL
	if eDianPao == 0 {
		eWuDuiHuType = dgk_IsWuDui(sHand, mHandCount, mCodeCount)
	}

	eJiangDuiDuiHuType := majiang_IsJiangDuiDui(sTotal, mCodeCount)
	isHu := majiang_IsHu(sHand, sCode)
	if eWuDuiHuType > common.HU_NIL || eJiangDuiDuiHuType > common.HU_NIL || isHu == true {

		if eJiangDuiDuiHuType > common.HU_NIL {
			if _, isExist := mExclusion[eJiangDuiDuiHuType]; isExist == false {
				return eJiangDuiDuiHuType // 将对对
			}
		}

		eDaiYaoJiuHuType := majiang_IsDaiYaoJiu(sHand, sPeng, sGang, mHandCount, isQingYiSe)
		if eDaiYaoJiuHuType > common.HU_NIL {
			if _, isExist := mExclusion[eDaiYaoJiuHuType]; isExist == false {
				return eDaiYaoJiuHuType // 清幺九和带幺九
			}
		}

		if eWuDuiHuType > common.HU_NIL {
			if _, isExist := mExclusion[eWuDuiHuType]; isExist == false {
				// 五对的清一色规则, 只判断五对是否为清一色
				return eWuDuiHuType
			}
		}

		if isHu == true {
			eDuiDuiHuType := majiang_IsDuiDuiHu(mCodeCount, isQingYiSe)
			if eDuiDuiHuType > common.HU_NIL {
				if _, isExist := mExclusion[eDuiDuiHuType]; isExist == false {
					return eDuiDuiHuType // 清对对和对对胡
				}
			}

			if isQingYiSe == true {
				if _, isExist := mExclusion[common.HU_QING_YI_SE]; isExist == false {
					return common.HU_QING_YI_SE // 清一色
				}
			} else {
				if _, isExist := mExclusion[common.HU_PING_HU]; isExist == false {
					return common.HU_PING_HU // 平胡
				}
			}
		}
	}
	return common.HU_NIL
}

func dgk_456IsHu(mHandCount map[uint8]int, isQingYiSe bool) common.EMaJiangHu {
	// 组成新手牌, 并排序
	pNew := &Majiang_Sort{}
	pNew.Cards = make([]common.EMaJiangType, 0, 11)
	for card, count := range mHandCount {
		if count > 0 {
			for i := 0; i < count; i++ {
				pNew.Cards = append(pNew.Cards, common.EMaJiangType(card))
			}
		}
	}
	sort.Sort(pNew)

	sCode := majiang_Marshal(pNew.Cards)
	isHu := majiang_IsHu(pNew.Cards, sCode)
	if isHu == true {
		if isQingYiSe == true {
			return common.HU_QING_YI_SE // 清一色
		} else {
			return common.HU_PING_HU // 平胡
		}
	}
	return common.HU_NIL
}

// 函数作用: 点炮情况下, 判断手牌中优先按照456顺子拆牌的情况下是否胡牌,以及456出现的次数
// 第一参数: 手牌切片, 长度只应该是5,8,11,14之一; 需从小到大排好序; 包含点炮的那张牌在内;
// 第二参数: 碰的牌
// 第三参数: 杠的牌
// 第四参数: 点炮的那张牌
func dgk_Calc456CountByDianPao(sHand []common.EMaJiangType, sPeng, sGang [][]common.EMaJiangType, eDianPao common.EMaJiangType) (common.EMaJiangHu, uint8) {

	nColor := eDianPao.Value() / 10
	if nColor < 1 || nColor > 3 {
		return common.HU_NIL, 0
	}
	nPoint := eDianPao.Value() % 10
	if nPoint != 5 {
		return common.HU_NIL, 0
	}

	sTotal := majiang_BuildSlice(sHand, sPeng, sGang)
	isQingYiSe := majiang_IsOnlyOneColor(sTotal)
	mHandCount := MajiangStatCount(sHand)
	n4 := (eDianPao - 1).Value()
	n5 := (eDianPao).Value()
	n6 := (eDianPao + 1).Value()
	if mHandCount[n4] > 0 && mHandCount[n5] > 0 && mHandCount[n6] > 0 {
		mHandCount[n4]--
		mHandCount[n5]--
		mHandCount[n6]--
		nHuType := dgk_456IsHu(mHandCount, isQingYiSe)
		if nHuType != common.HU_NIL {
			return nHuType, 1
		}
	}
	return common.HU_NIL, 0
}

// 函数作用: 自摸情况下, 判断手牌中优先按照456顺子拆牌的情况下是否胡牌,以及456出现的次数
// 第一参数: 手牌切片, 长度只应该是5,8,11,14之一; 需从小到大排好序;
// 第二参数: 碰的牌
// 第三参数: 杠的牌
func dgk_Calc456CountByZiMo(sHand []common.EMaJiangType, sPeng, sGang [][]common.EMaJiangType) (common.EMaJiangHu, uint8) {

	sTotal := majiang_BuildSlice(sHand, sPeng, sGang)
	isQingYiSe := majiang_IsOnlyOneColor(sTotal)
	mHandCount := MajiangStatCount(sHand)
	nMaxCount := uint8(0)

	_calc456Count := func(mHandCount map[uint8]int, nOne, nTwo, nThree uint8) (common.EMaJiangHu, uint8) {
		nCount := uint8(0)
		nMaxHuType := common.HU_NIL

		for {
			if mHandCount[nOne] > 0 && mHandCount[nTwo] > 0 && mHandCount[nThree] > 0 {
				mHandCount[nOne]--
				mHandCount[nTwo]--
				mHandCount[nThree]--
				nHuType := dgk_456IsHu(mHandCount, isQingYiSe)
				if nHuType != common.HU_NIL && nHuType >= nMaxHuType {
					nMaxHuType = nHuType
					nCount++
				}
			} else {
				break
			}
		}
		return nMaxHuType, nCount
	}

	mTestCount := make(map[uint8]int)
	for key, value := range mHandCount {
		mTestCount[key] = value
	}
	nHuTongType, nTongCount := _calc456Count(mTestCount, common.TONG_4.Value(), common.TONG_5.Value(), common.TONG_6.Value())

	mTestCount = make(map[uint8]int)
	for key, value := range mHandCount {
		mTestCount[key] = value
	}
	nHuTiaoType, nTiaoCount := _calc456Count(mTestCount, common.TIAO_4.Value(), common.TIAO_5.Value(), common.TIAO_6.Value())

	nMaxCount += nTongCount
	nMaxCount += nTiaoCount

	nMaxHuType := common.EMaJiangHu(0)
	if nHuTongType > nHuTiaoType {
		nMaxHuType = nHuTongType
	} else {
		nMaxHuType = nHuTiaoType
	}
	return nMaxHuType, nMaxCount
}

// 函数作用:传入的牌组里是否断幺九
// 返回true表示没有19牌
func majiang_IsDuanYaoJiu(sTotal []uint8) bool {
	for _, value := range sTotal {
		nPoint := value % 10
		if nPoint == 1 || nPoint == 9 {
			return false
		}
	}
	return true
}

func dgk_IsWuDui(sHand []common.EMaJiangType, mHandCount map[uint8]int, mCodeCount map[uint8]int) common.EMaJiangHu {

	if len(sHand) != 11 {
		return common.HU_NIL
	}

	// 第一种牌型: 五对 + 一单张	  	 共计11张牌
	// 第二种牌型: 四对 + 一坎 			 共计11张牌
	// 第三种牌型: 三对 + 一勾 + 一单张  共计11张牌
	// 第四种牌型: 二对 + 一勾 + 一坎    共计11张牌
	// 第五种牌型: 一对 + 二勾 + 一单张  共计11张牌
	// 第六种牌型: 二勾 + 一坎		     共计11张牌
	if (mCodeCount[2] == 5) ||
		(mCodeCount[2] == 4 && mCodeCount[3] == 1) ||
		(mCodeCount[2] == 3 && mCodeCount[4] == 1 && mCodeCount[1] == 1) ||
		(mCodeCount[2] == 2 && mCodeCount[4] == 1 && mCodeCount[3] == 1) ||
		(mCodeCount[2] == 1 && mCodeCount[4] == 2 && mCodeCount[1] == 1) ||
		(mCodeCount[4] == 2 && mCodeCount[3] == 1) {

		sTotal := make([]uint8, 0, 10)
		for nKey, nCount := range mHandCount {
			if nCount == 1 {
				continue
			} else if nCount == 3 {
				sTotal = append(sTotal, nKey)
			} else {
				sTotal = append(sTotal, nKey)
			}
		}
		isQingWuDui := majiang_IsOnlyOneColor(sTotal)
		if isQingWuDui == true {
			return common.HU_QING_WU_DUI
		}
		return common.HU_WU_DUI
	}
	return common.HU_NIL
}

// 函数作用: 判断手牌+碰的所有牌+杠的所有牌是否是带幺九牌型
func majiang_IsDaiYaoJiu(sHand []common.EMaJiangType, sPeng, sGang [][]common.EMaJiangType, mHandCount map[uint8]int, isQingYiSe bool) common.EMaJiangHu {
	for _, value := range sHand {
		if value.Value() > common.WAN_9.Value() {
			return common.HU_NIL
		}
		nPoint := value % 10
		if nPoint == 4 || nPoint == 5 || nPoint == 6 {
			return common.HU_NIL
		}
	}

	// 可能会出现111, 222, 333, 99这种牌型
	// 可能会出现111, 123, 789, 99这种牌型
	// 可能会出现111, 77, 88, 9999这种牌型
	// 可能会出现123, 77, 88, 9999这种牌型
	for value, count := range mHandCount {
		nPoint := value % 10
		if count >= 4 {
			if nPoint != 1 && nPoint != 9 {
				return common.HU_NIL
			}
		}

		if nPoint != 1 && nPoint != 9 {
			if mHandCount[value+1] >= count && mHandCount[value+2] >= count {
				continue
			}
			if mHandCount[value-2] >= count && mHandCount[value-1] >= count {
				continue
			}
			if mHandCount[value-1] >= count && mHandCount[value+1] >= count {
				continue
			}
			return common.HU_NIL
		}
	}

	for _, node := range sPeng {
		if len(node) > 0 {
			nPoint := node[0] % 10
			if nPoint != 1 && nPoint != 9 {
				return common.HU_NIL
			}
		}
	}
	for _, node := range sGang {
		if len(node) > 0 {
			nPoint := node[0] % 10
			if nPoint != 1 && nPoint != 9 {
				return common.HU_NIL
			}
		}
	}

	if isQingYiSe == true {
		return common.HU_QING_YAO_JIU // 清幺九
	} else {
		return common.HU_DAI_YAO_JIU // 带幺九
	}
}

// 函数作用: 将麻将切片转码成更简洁的uint8切片
// 第一参数: 手牌切片, 长度只应该是2,5,8,11,14之一; 需从小到大排好序;
func majiang_Marshal(sHand []common.EMaJiangType) []uint8 {
	sRet := make([]uint8, 0, len(sHand))
	sRet = append(sRet, 1)
	for i, pos := 1, 0; i < len(sHand); i++ {
		if sHand[i-1] == sHand[i] {
			sRet[pos]++
		} else if sHand[i-1]+1 == sHand[i] {
			sRet = append(sRet, 1)
			pos++
		} else {
			sRet = append(sRet, 0)
			sRet = append(sRet, 1)
			pos += 2
		}
	}
	return sRet
}

func majiang_BuildSlice(sHand []common.EMaJiangType, sPeng, sGang [][]common.EMaJiangType) []uint8 {
	sTotal := make([]uint8, 0, len(sHand))
	for _, value := range sHand {
		sTotal = append(sTotal, value.Value())
	}
	for _, node := range sPeng {
		if len(node) > 0 {
			sTotal = append(sTotal, node[0].Value())
		}
	}
	for _, node := range sGang {
		if len(node) > 0 {
			sTotal = append(sTotal, node[0].Value())
		}
	}
	return sTotal
}

// 函数作用: 判断手牌+碰的所有牌+杠的所有牌是否是清一色
// 返回参数: true表示清一色  false表示非清一色
func majiang_IsOnlyOneColor(sTotal []uint8) bool {
	var nLastGroup uint8
	for _, value := range sTotal {
		if value > common.WAN_9.Value() {
			return false
		}
		if nLastGroup == 0 {
			nLastGroup = value / 10
		} else {
			nGroup := value / 10
			if nLastGroup != nGroup {
				return false
			}
		}
	}
	return true
}

// 函数作用: 判断(手牌+碰+杠)所有牌中勾的数量
func majiang_GouCount(sHand []common.EMaJiangType, sPeng, sGang [][]common.EMaJiangType) uint8 {
	nCount := len(sGang)
	mHandCount := MajiangStatCount(sHand)
	for _, count := range mHandCount {
		if count >= 4 {
			nCount++
		}
	}
	for _, node := range sPeng {
		if len(node) > 0 {
			if _, isExist := mHandCount[node[0].Value()]; isExist == true {
				nCount++
			}
		}
	}
	return uint8(nCount)
}

// 函数作用: 判断手牌+碰的所有牌+杠的所有牌是否是将对对牌型
func majiang_IsJiangDuiDui(sTotal []uint8, mCodeCount map[uint8]int) common.EMaJiangHu {
	for _, value := range sTotal {
		if value > common.WAN_9.Value() {
			return common.HU_NIL
		}
		nPoint := value % 10
		if nPoint != 2 && nPoint != 5 && nPoint != 8 {
			return common.HU_NIL
		}
	}

	if mCodeCount != nil {
		if mCodeCount[1] > 0 || mCodeCount[2] != 1 || mCodeCount[4] > 0 {
			return common.HU_NIL
		}
	}
	return common.HU_JIANG_DUI_DUI
}

func majiang_IsDuiDuiHu(mCodeCount map[uint8]int, isQingYiSe bool) common.EMaJiangHu {
	if mCodeCount[2] != 1 || mCodeCount[1] > 0 || mCodeCount[4] > 0 {
		return common.HU_NIL
	}

	if isQingYiSe == true {
		return common.HU_QING_DUI_DUI // 清对对
	} else {
		return common.HU_DUI_DUI_HU // 对对胡
	}
}

// 至少5张牌才能用此函数判断
func majiang_IsHu(sHand []common.EMaJiangType, sCode []uint8) bool {
	nHandLen := len(sHand)
	if nHandLen == 2 {
		if sHand[0] == sHand[1] {
			return true
		}
		return false

	} else if nHandLen == 5 || nHandLen == 8 || nHandLen == 11 || nHandLen == 14 {
		nCodeLen := len(sCode)
		for i := 0; i < nCodeLen; i++ {
			nCodeCount := sCode[i]
			if nCodeCount >= 2 {
				sTestCode := make([]uint8, nCodeLen, nCodeLen)
				copy(sTestCode, sCode)
				sTestCode[i] -= 2
				isOK := calcCodeIsHu(sTestCode)
				if isOK == true {
					return true
				}
			}
		}
	}
	return false
}

// 是否是十八罗汉牌型
func majiang_IsShiBaLuoHan(sHand []common.EMaJiangType, sGang [][]common.EMaJiangType, isQingYiSe bool) common.EMaJiangHu {
	nHandLen := len(sHand)
	nGangLen := len(sGang)
	if nHandLen == 2 && nGangLen == 4 {
		if isQingYiSe == true {
			return common.HU_QING_SHI_BA_LUO_HAN
		}
		return common.HU_SHI_BA_LUO_HAN
	}
	return common.HU_NIL
}

func calcCodeIsHu(sTest []uint8) bool {
	nCodeLen := len(sTest)
	for j := 0; j < nCodeLen; {
		nTestCode := sTest[j]
		if nTestCode == 0 {
			j++
		} else if nTestCode == 3 {
			sTest[j] -= 3
			j++
		} else {
			if j+2 >= nCodeLen {
				j++
				continue
			}
			if sTest[j] > 0 && sTest[j+1] > 0 && sTest[j+2] > 0 {
				sTest[j]--
				sTest[j+1]--
				sTest[j+2]--
			} else {
				if nTestCode == 1 || nTestCode == 2 || nTestCode == 4 {
					return false
				} else {
					j++
				}
			}
		}
	}
	for _, value := range sTest {
		if value > 0 {
			return false
		}
	}
	return true
}

// 函数作用: 判断手牌是否胡牌, 若胡牌返回胡牌牌型
// 第一参数: 手牌切片, 长度只应该是2,5,8,11,14之一; 需从小到大排好序;
// 第二参数: 碰牌切片; 没有碰传空切片
// 第三参数: 杠牌切片; 没有杠传空切片
var exclusion = make(map[common.EMaJiangHu]bool)

func DGK_CalcHuType(sHand []common.EMaJiangType, sPeng, sGang [][]common.EMaJiangType, eDianPao common.EMaJiangType) common.EMaJiangHu {
	eHuType := dgk_CalcHuTypeByExclusion(sHand, sPeng, sGang, eDianPao, exclusion)
	return eHuType
}

// 函数作用: 计算所有胡牌牌型和额外加番番数
// 第一参数: 手牌切片, 长度只应该是2,5,8,11,14之一; 需从小到大排好序;
// 第二参数: 碰牌切片; 没有碰传空切片
// 第三参数: 杠牌切片; 没有杠传空切片
// 第四参数: 点炮传点炮那张牌, 自摸传0
func DGK_CalcHuAndExtra(sHand []common.EMaJiangType, sPeng, sGang [][]common.EMaJiangType, eDianPao common.EMaJiangType) []*Majiang_Hu {

	eBaseHuType := DGK_CalcHuType(sHand, sPeng, sGang, eDianPao)
	if eBaseHuType == common.HU_NIL {
		return nil
	}

	pRet := make([]*Majiang_Hu, 0, 3)
	pBase := &Majiang_Hu{HuType: eBaseHuType, Extra: nil, Total: 0}
	pRet = append(pRet, pBase)

	var sTotal []uint8
	// 五对判断是否断幺九, 需要去掉单独的一张;
	if eBaseHuType == common.HU_QING_WU_DUI || eBaseHuType == common.HU_WU_DUI {
		sTotal = make([]uint8, 0, 10)
		mHandCount := MajiangStatCount(sHand)
		for nKey, nCount := range mHandCount {
			if nCount == 1 {
				continue
			} else if nCount == 3 {
				sTotal = append(sTotal, nKey)
			} else {
				sTotal = append(sTotal, nKey)
			}
		}
	} else {
		sTotal = majiang_BuildSlice(sHand, sPeng, sGang)
	}
	isDuanYaoJiu := majiang_IsDuanYaoJiu(sTotal)
	nGouCount := majiang_GouCount(sHand, sPeng, sGang)
	if isDuanYaoJiu == true || nGouCount > 0 {
		pBase.Extra = make(map[common.EMaJiangExtra]uint8, 0)
	}
	if isDuanYaoJiu == true {
		pBase.Extra[common.EXTRA_DUAN_YAO_JIU] = 1
		pBase.Total += 1
	}
	if nGouCount > 0 {
		pBase.Extra[common.EXTRA_GOU] = nGouCount
		pBase.Total += nGouCount
	}

	// 计算所有可胡牌类型
	mExclusion := make(map[common.EMaJiangHu]bool)
	mExclusion[eBaseHuType] = true
	for {
		eHuType := dgk_CalcHuTypeByExclusion(sHand, sPeng, sGang, eDianPao, mExclusion)
		if eHuType > common.HU_NIL {
			mExclusion[eHuType] = true

			pNode := &Majiang_Hu{HuType: eHuType, Extra: nil, Total: 0}
			pRet = append(pRet, pNode)

			if isDuanYaoJiu == true || nGouCount > 0 {
				pNode.Extra = make(map[common.EMaJiangExtra]uint8, 0)
			}
			if isDuanYaoJiu == true {
				pNode.Extra[common.EXTRA_DUAN_YAO_JIU] = 1
				pNode.Total += 1
			}
			if nGouCount > 0 {
				pNode.Extra[common.EXTRA_GOU] = nGouCount
				pNode.Total += nGouCount
			}
		} else {
			break
		}
	}

	var pKa *Majiang_Hu
	eKaHuType := common.HU_NIL
	nKaCount := uint8(0)
	if eDianPao == 0 {
		eKaHuType, nKaCount = dgk_Calc456CountByZiMo(sHand, sPeng, sGang)
	} else {
		eKaHuType, nKaCount = dgk_Calc456CountByDianPao(sHand, sPeng, sGang, eDianPao)
	}
	if eKaHuType > common.HU_NIL {
		pKa = &Majiang_Hu{HuType: eKaHuType, Extra: nil, Total: 0}
		pRet = append(pRet, pKa)

		pKa.Extra = make(map[common.EMaJiangExtra]uint8, 0)
		pKa.Extra[common.EXTRA_456_KA] = nKaCount
		pKa.Total += nKaCount

		if isDuanYaoJiu == true {
			pKa.Extra[common.EXTRA_DUAN_YAO_JIU] = 1
			pKa.Total += 1
		}
		if nGouCount > 0 {
			pKa.Extra[common.EXTRA_GOU] = nGouCount
			pKa.Total += nGouCount
		}
	}
	return pRet
}

func xmmj_IsQiDui(sHand []common.EMaJiangType, mCodeCount map[uint8]int, isQingYiSe bool) common.EMaJiangHu {

	if len(sHand) != 14 {
		return common.HU_NIL
	}

	// 第一种牌型: 七对				  	 共计14张牌
	// 第二种牌型: 五对 + 一勾 			 共计14张牌
	// 第三种牌型: 三对 + 二勾			 共计14张牌
	// 第四种牌型: 一对 + 三勾			 共计14张牌
	if (mCodeCount[2] == 7) ||
		(mCodeCount[2] == 5 && mCodeCount[4] == 1) ||
		(mCodeCount[2] == 3 && mCodeCount[4] == 2) ||
		(mCodeCount[2] == 1 && mCodeCount[4] == 3) {

		if mCodeCount[4] > 0 {
			if isQingYiSe == true {
				return common.HU_QING_LONG_QI_DUI
			} else {
				return common.HU_LONG_QI_DUI
			}
		}
		if isQingYiSe == true {
			return common.HU_QING_QI_DUI
		}
		return common.HU_QI_DUI
	}
	return common.HU_NIL
}

// 函数作用: 判断手牌+碰的所有牌+杠的所有牌是否是将对对牌型
func xmmj_IsJiangDuiDui(sHand []common.EMaJiangType, sTotal []uint8, mCodeCount map[uint8]int) common.EMaJiangHu {

	eHuType := majiang_IsJiangDuiDui(sTotal, mCodeCount)
	if eHuType == common.HU_NIL {
		return common.HU_NIL
	}

	nHandLen := len(sHand)
	if nHandLen == 2 {
		return common.HU_JIANG_JIN_GOU_DIAO
	}
	return eHuType
}

func xmmj_IsDuiDuiHu(sHand []common.EMaJiangType, mCodeCount map[uint8]int, isQingYiSe bool) common.EMaJiangHu {

	eDuiDuiHuType := majiang_IsDuiDuiHu(mCodeCount, isQingYiSe)
	if eDuiDuiHuType == common.HU_NIL {
		return common.HU_NIL
	}

	nHandLen := len(sHand)
	if nHandLen == 2 {
		if isQingYiSe == true {
			return common.HU_QING_JIN_GOU_DIAO
		} else {
			return common.HU_JIN_GOU_DIAO
		}
	}
	return eDuiDuiHuType
}

func xmmj_CalcHuTypeByExclusion(sHand []common.EMaJiangType, sPeng, sGang [][]common.EMaJiangType, eDianPao common.EMaJiangType, mExclusion map[common.EMaJiangHu]bool) common.EMaJiangHu {
	nHandLen := len(sHand)
	if nHandLen != 2 && nHandLen != 5 && nHandLen != 8 && nHandLen != 11 && nHandLen != 14 {
		log.Errorf("手牌长度异常:%v, 正确长度应该是2,5,8,11,14之一", nHandLen)
		return common.HU_NIL
	}

	sTotal := majiang_BuildSlice(sHand, sPeng, sGang)
	mHandCount := MajiangStatCount(sHand)
	isQingYiSe := majiang_IsOnlyOneColor(sTotal)
	sCode := majiang_Marshal(sHand)
	mCodeCount := utils.StatElementCount(sCode)

	eQiDuiHuType := xmmj_IsQiDui(sHand, mCodeCount, isQingYiSe)
	eJiangDuiDuiHuType := xmmj_IsJiangDuiDui(sHand, sTotal, mCodeCount)
	isHu := majiang_IsHu(sHand, sCode)
	if eQiDuiHuType > common.HU_NIL || eJiangDuiDuiHuType > common.HU_NIL || isHu == true {

		eShiBaLuoHanHuType := majiang_IsShiBaLuoHan(sHand, sGang, isQingYiSe)
		if eShiBaLuoHanHuType > common.HU_NIL {
			if _, isExist := mExclusion[eShiBaLuoHanHuType]; isExist == false {
				return eShiBaLuoHanHuType // 十八罗汉和清十八罗汉
			}
		}

		if eJiangDuiDuiHuType > common.HU_NIL {
			if _, isExist := mExclusion[eJiangDuiDuiHuType]; isExist == false {
				return eJiangDuiDuiHuType // 将对对和将金钩钓
			}
		}

		eDaiYaoJiuHuTyep := majiang_IsDaiYaoJiu(sHand, sPeng, sGang, mHandCount, isQingYiSe)
		if eDaiYaoJiuHuTyep > common.HU_NIL {
			if _, isExist := mExclusion[eDaiYaoJiuHuTyep]; isExist == false {
				return eDaiYaoJiuHuTyep // 清幺九和带幺九
			}
		}

		if eQiDuiHuType > common.HU_NIL {
			if _, isExist := mExclusion[eQiDuiHuType]; isExist == false {
				return eQiDuiHuType
			}
		}

		if isHu == true {
			eDuiDuiHuType := xmmj_IsDuiDuiHu(sHand, mCodeCount, isQingYiSe)
			if eDuiDuiHuType > common.HU_NIL {
				if _, isExist := mExclusion[eDuiDuiHuType]; isExist == false {
					return eDuiDuiHuType // 金钩钓和清金钩钓和清对对和对对胡
				}
			}

			if isQingYiSe == true {
				if _, isExist := mExclusion[common.HU_QING_YI_SE]; isExist == false {
					return common.HU_QING_YI_SE // 清一色
				}
			} else {
				if _, isExist := mExclusion[common.HU_PING_HU]; isExist == false {
					return common.HU_PING_HU // 平胡
				}
			}
		}
	}
	return common.HU_NIL
}

// 函数作用: 判断手牌是否胡牌, 若胡牌返回胡牌牌型
// 第一参数: 手牌切片, 长度只应该是2,5,8,11,14之一; 需从小到大排好序;
// 第二参数: 碰牌切片; 没有碰传空切片
// 第三参数: 杠牌切片; 没有杠传空切片
func XMMJ_CalcHuType(sHand []common.EMaJiangType, sPeng, sGang [][]common.EMaJiangType, eDianPao common.EMaJiangType) common.EMaJiangHu {
	eHuType := xmmj_CalcHuTypeByExclusion(sHand, sPeng, sGang, eDianPao, exclusion)
	return eHuType
}

// 函数作用: 计算所有胡牌牌型和额外加番番数
// 第一参数: 手牌切片, 长度只应该是2,5,8,11,14之一; 需从小到大排好序;
// 第二参数: 碰牌切片; 没有碰传空切片
// 第三参数: 杠牌切片; 没有杠传空切片
// 第四参数: 点炮传点炮那张牌, 自摸传0
func XMMJ_CalcHuAndExtra(sHand []common.EMaJiangType, sPeng, sGang [][]common.EMaJiangType, eDianPao common.EMaJiangType) []*Majiang_Hu {

	eBaseHuType := XMMJ_CalcHuType(sHand, sPeng, sGang, eDianPao)
	if eBaseHuType == common.HU_NIL {
		return nil
	}

	pRet := make([]*Majiang_Hu, 0, 3)
	pBase := &Majiang_Hu{HuType: eBaseHuType, Extra: nil, Total: 0}
	pRet = append(pRet, pBase)

	sTotal := majiang_BuildSlice(sHand, sPeng, sGang)
	isDuanYaoJiu := majiang_IsDuanYaoJiu(sTotal)
	nGouCount := majiang_GouCount(sHand, sPeng, sGang)
	if eBaseHuType == common.HU_LONG_QI_DUI || eBaseHuType == common.HU_QING_LONG_QI_DUI {
		// 龙七对和清龙七对需要扣除自身成为龙七对条件那个勾
		nGouCount--
	} else if eBaseHuType == common.HU_SHI_BA_LUO_HAN {
		// 十八罗汉计算勾
		nGouCount = 0
	}

	if isDuanYaoJiu == true || nGouCount > 0 {
		pBase.Extra = make(map[common.EMaJiangExtra]uint8, 0)
	}
	if isDuanYaoJiu == true {
		pBase.Extra[common.EXTRA_DUAN_YAO_JIU] = 1
		pBase.Total += 1
	}
	if nGouCount > 0 {
		pBase.Extra[common.EXTRA_GOU] = nGouCount
		pBase.Total += nGouCount
	}

	// 计算所有可胡牌类型
	mExclusion := make(map[common.EMaJiangHu]bool)
	mExclusion[eBaseHuType] = true
	for {
		eHuType := xmmj_CalcHuTypeByExclusion(sHand, sPeng, sGang, eDianPao, mExclusion)
		if eHuType > common.HU_NIL {
			mExclusion[eHuType] = true

			pNode := &Majiang_Hu{HuType: eHuType, Extra: nil, Total: 0}
			pRet = append(pRet, pNode)

			if isDuanYaoJiu == true || nGouCount > 0 {
				pNode.Extra = make(map[common.EMaJiangExtra]uint8, 0)
			}
			if isDuanYaoJiu == true {
				pNode.Extra[common.EXTRA_DUAN_YAO_JIU] = 1
				pNode.Total += 1
			}
			if nGouCount > 0 {
				pNode.Extra[common.EXTRA_GOU] = nGouCount
				pNode.Total += nGouCount
			}
		} else {
			break
		}
	}
	return pRet
}
