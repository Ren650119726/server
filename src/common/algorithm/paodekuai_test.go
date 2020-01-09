package algorithm

import (
	"root/common"
	"fmt"
	"testing"
)

func TestPaoDeKuai(t *testing.T) {
	//sNiuNiu := [][]common.Card_info{
	//	{{5, 2}, {5, 3}, {3, 6}, {3, 7}, {2, 8}},
	//	{{5, 2}, {4, 5}, {5, 3}, {3, 7}, {2, 8}},
	//	{{5, 2}, {4, 5}, {3, 6}, {5, 3}, {2, 8}},
	//	{{5, 2}, {4, 5}, {3, 6}, {3, 7}, {5, 3}},
	//	{{4, 4}, {5, 2}, {5, 3}, {3, 7}, {2, 8}},
	//	{{4, 4}, {5, 2}, {3, 6}, {5, 3}, {2, 8}},
	//	{{4, 4}, {5, 2}, {3, 6}, {3, 7}, {5, 3}},
	//	{{4, 4}, {4, 5}, {5, 2}, {5, 3}, {2, 8}},
	//	{{4, 4}, {4, 5}, {5, 2}, {3, 7}, {5, 3}},
	//	{{4, 4}, {4, 5}, {3, 6}, {5, 2}, {5, 3}},
	//}
	//for index, card := range sNiuNiu {
	//	nType := NiuNiu_CalcCardTypeAndSort(card, 15)
	//	fmt.Println(index, nType, common.ENiuNiuType(nType))
	//}

	sNiuNiu := []common.Card_info{{4, 10}, {4, 5}, {3, 4}, {2, 3}, {5, 3}}
	nType := NiuNiu_CalcCardTypeAndSort(sNiuNiu, 15)
	fmt.Println(nType, common.ENiuNiuType(nType))

	sHand := []common.Card_info{
		{1, 12},
		{2, 12},
		{3, 12},
		{4, 11},
		{1, 11},
		{2, 11},
		{2, 7},
		{2, 6},
		{2, 3},
		{2, 3},
	}

	nValue := PaoDeKuai_HN_CalcCardValue(sHand)
	sRet1 := PaoDeKuai_HN_GetBigCard(sHand, pdk_hn_intelligent_danzhang_card, false, false)
	sRet2 := PaoDeKuai_HN_GetBigCard(sHand, pdk_hn_intelligent_danzhang_card, true, false)
	fmt.Println("===============sRet", sRet1, sRet2, nValue)

	sIn := []common.Card_info{
		{1, 12},
		{2, 12},
		{2, 12},
		{3, 12},
	}
	//Poker_SortCard(sHand)
	//eType, nSubLen := PaoDeKuai_HN_CalcCardType(sIn)
	isRet := PaoDeKuai_HN_IsHaveBiggerCard(sIn, sHand)
	//sRet := PaoDeKuai_HN_GetBigCard(sHand, sIn, false)
	fmt.Println("手牌", sHand)
	//fmt.Println("输入", sIn, eType, nSubLen)
	fmt.Println("结果", isRet)
	//sHand = Poker_RemoveCard(sHand, sRet)
	//fmt.Println("删后", sHand)

	sTemp := []common.Card_info{
		{1, 13},
		{2, 7},
		{2, 1},
	}
	nRet := niuniu_CalcHaveJokerCardType(sTemp, 1, 2)
	fmt.Println(nRet, "=============")
}
