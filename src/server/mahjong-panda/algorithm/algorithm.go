package algorithm

import (
	"root/common"
	"root/common/algorithm"
	"root/core/utils"
	"math/rand"
)

type (
	Jiao_Card struct {
		Card common.EMaJiangType
		Hu   common.EMaJiangHu
		T    int8 // 1 自摸 2 点炮
	}
)

var ALL_TYPE_CARD_GROUP = []common.EMaJiangType{
	common.TONG_1, common.TONG_2, common.TONG_3, common.TONG_4, common.TONG_5, common.TONG_6, common.TONG_7, common.TONG_8, common.TONG_9,
	common.TIAO_1, common.TIAO_2, common.TIAO_3, common.TIAO_4, common.TIAO_5, common.TIAO_6, common.TIAO_7, common.TIAO_8, common.TIAO_9,
	common.WAN_1, common.WAN_2, common.WAN_3, common.WAN_4, common.WAN_5, common.WAN_6, common.WAN_7, common.WAN_8, common.WAN_9,
}

// 随机获得不重复的n张牌
func GetRandom_Card(count int) []common.EMaJiangType {
	var cards = []common.EMaJiangType{
		common.TONG_1, common.TONG_1, common.TONG_1, common.TONG_1,
		common.TONG_2, common.TONG_2, common.TONG_2, common.TONG_2,
		common.TONG_3, common.TONG_3, common.TONG_3, common.TONG_3,
		common.TONG_4, common.TONG_4, common.TONG_4, common.TONG_4,
		common.TONG_5, common.TONG_5, common.TONG_5, common.TONG_5,
		common.TONG_6, common.TONG_6, common.TONG_6, common.TONG_6,
		common.TONG_7, common.TONG_7, common.TONG_7, common.TONG_7,
		common.TONG_8, common.TONG_8, common.TONG_8, common.TONG_8,
		common.TONG_9, common.TONG_9, common.TONG_9, common.TONG_9,

		common.TIAO_1, common.TIAO_1, common.TIAO_1, common.TIAO_1,
		common.TIAO_2, common.TIAO_2, common.TIAO_2, common.TIAO_2,
		common.TIAO_3, common.TIAO_3, common.TIAO_3, common.TIAO_3,
		common.TIAO_4, common.TIAO_4, common.TIAO_4, common.TIAO_4,
		common.TIAO_5, common.TIAO_5, common.TIAO_5, common.TIAO_5,
		common.TIAO_6, common.TIAO_6, common.TIAO_6, common.TIAO_6,
		common.TIAO_7, common.TIAO_7, common.TIAO_7, common.TIAO_7,
		common.TIAO_8, common.TIAO_8, common.TIAO_8, common.TIAO_8,
		common.TIAO_9, common.TIAO_9, common.TIAO_9, common.TIAO_9,

		//common.WAN_1, common.WAN_1, common.WAN_1, common.WAN_1,
		//common.WAN_2, common.WAN_2, common.WAN_2, common.WAN_2,
		//common.WAN_3, common.WAN_3, common.WAN_3, common.WAN_3,
		//common.WAN_4, common.WAN_4, common.WAN_4, common.WAN_4,
		//common.WAN_5, common.WAN_5, common.WAN_5, common.WAN_5,
		//common.WAN_6, common.WAN_6, common.WAN_6, common.WAN_6,
		//common.WAN_7, common.WAN_7, common.WAN_7, common.WAN_7,
		//common.WAN_8, common.WAN_8, common.WAN_8, common.WAN_8,
		//common.WAN_9, common.WAN_9, common.WAN_9, common.WAN_9,
	}

	if count > len(cards) {
		return nil
	}

	rand.Seed(utils.SecondTimeSince1970())
	ret := make([]common.EMaJiangType, 0, count)
	for i := 0; i < count; i++ {
		last := len(cards) - 1 - i
		rand_val := utils.Randx_y(0, last)
		ret = append(ret, cards[rand_val])
		cards[rand_val], cards[last] = cards[last], cards[rand_val]
	}
	return ret
}

// 传入当前牌，附加 一张单牌，能否胡牌
func jiao(sHand []common.EMaJiangType, sPeng, sGang [][]common.EMaJiangType, card common.EMaJiangType) common.EMaJiangHu {
	var cards []common.EMaJiangType
	if card != 0 {
		cards, _ = InsertCard(sHand, card)
	} else {
		cards = sHand
	}

	h := algorithm.DGK_CalcHuType(cards, sPeng, sGang, card)
	return h
}

//
func Jiao_(sHand []common.EMaJiangType, sPeng, sGang [][]common.EMaJiangType) []Jiao_Card {
	retV := make([]Jiao_Card, 0, 0)
	for _, v := range ALL_TYPE_CARD_GROUP {
		// 点炮
		h := jiao(sHand, sPeng, sGang, v)
		if h != common.HU_NIL {
			retV = append(retV, Jiao_Card{Card: v, Hu: h, T: 2})
		}
		// 自摸
		//new, _ := InsertCard(sHand, v)
		//h = jiao(new, sPeng, sGang, 0)
		//if h != common.HU_NIL {
		//	insert := true
		//	for _, vv := range retV {
		//		if vv.Hu == h && vv.Card == v {
		//			insert = false
		//			break
		//		}
		//	}
		//	if insert {
		//		retV = append(retV, Jiao_Card{Card: v, Hu: h, T: 1})
		//	}
		//
		//}

	}
	return retV
}

// 将一张单牌插入一组有序的牌组中 小->大
func InsertCard(cards []common.EMaJiangType, card common.EMaJiangType) (ret []common.EMaJiangType, index int) {
	new := cards
	p1 := card
	index = len(new)
	for i := len(new) - 1; i >= 0; i-- {
		if p1 >= new[i] {
			t := []common.EMaJiangType{}
			t = append(t, new[:i+1]...)
			t = append(t, p1)
			if i+1 < len(new) {
				t = append(t, new[i+1:]...)
			}
			new = t
			index = i + 1
			break
		} else if i == 0 {
			t := []common.EMaJiangType{}
			t = append(t, p1)
			t = append(t, new...)
			new = t
			index = 0
		}
	}
	return new, index
}

// 检查有没有暗杠 和 弯杠
func AllGang(hand []common.EMaJiangType, peng [][]common.EMaJiangType) []int {
	ret := []int{}
	new := hand
	for _, v := range peng {
		p1 := v[0]
		for i := len(new) - 1; i >= 0; i-- {
			if p1 >= new[i] {
				t := []common.EMaJiangType{}
				t = append(t, new[:i+1]...)
				t = append(t, v...)
				if i+1 < len(new) {
					t = append(t, new[i+1:]...)
				}
				new = t
				break
			} else if i == 0 {
				t := []common.EMaJiangType{}
				t = append(t, v...)
				t = append(t, new...)
				new = t
			}
		}
	}

	count := 0
	val := new[0]
	for _, v := range new {
		if val == v {
			count++
		} else {
			val = v
			count = 1
		}

		if count == 4 {
			for i, hv := range hand {
				if val == hv {
					ret = append(ret, i)
					break
				}
			}

		}
	}
	return ret
}

// 检查能不能碰牌
func CheckPeng(hand []common.EMaJiangType, card common.EMaJiangType) int {
	index := -1
	for i, v := range hand {
		if i+1 >= len(hand) {
			return index
		} else if v == card && hand[i+1] == card {
			return i
		}
	}
	return index
}

// 检查能不能直杠
func CheckGang(hand []common.EMaJiangType, card common.EMaJiangType) int {
	index := -1

	for i, v := range hand {
		if i+2 >= len(hand) {
			return index
		} else if v == card && hand[i+1] == card && hand[i+2] == card {
			return i
		}
	}
	return index
}

// 检查能不能暗杠
func CheckGang_AN(hand []common.EMaJiangType) []int {
	index := []int{}

	for i := 0; i < len(hand); {
		card := hand[i]
		if i+3 <= len(hand) && hand[i+1] == card && hand[i+2] == card && hand[i+3] == card {
			index = append(index, i)
			i += 4
		} else {
			i++
		}
	}
	return index
}
