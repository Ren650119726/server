package room

import (
	"root/common"
	ag "root/common/algorithm"
	"fmt"
	"root/server/mahjong-dgk/algorithm"
)

var FAN_RATIO = [...]int{
	1: 1,
	2: 2,
	3: 4,
	4: 8,
	5: 16,
	6: 32,
}

////////////////////////////////////////////////////////////////////////////////////////////////////////
func NewCardGroup() *CardGroup {
	return &CardGroup{
		hand: make([]common.EMaJiangType, 0, 0),
		peng: make([][]common.EMaJiangType, 0, 0),
		gang: make([][]common.EMaJiangType, 0, 0),
	}
}

func (self *CardGroup) String() string {
	return fmt.Sprintf("手牌:%v 碰:%v 杠:%v", self.hand, self.peng, self.gang)
}

// 手牌是否胡
func (self *GamePlayer) IsHu_() common.EMaJiangHu {
	return ag.DGK_CalcHuType(self.cards.hand, self.cards.peng, self.cards.gang, 0)
}

// 打出牌后还有叫，返回可以打的牌下标
func (self *GamePlayer) Master_All_Jiao() []int {
	ret := []int{}
	for index := 0; index < len(self.cards.hand); index++ {
		new := []common.EMaJiangType{}
		new = append(new, self.cards.hand[:index]...)
		new = append(new, self.cards.hand[index+1:]...)
		if len(algorithm.Jiao_(new, [][]common.EMaJiangType{}, [][]common.EMaJiangType{})) != 0 {
			ret = append(ret, index)
		}
	}

	return ret
}

type statisics struct {
	len        int
	coincide   [5]int
	continuous [10]int

	single []common.EMaJiangType
}

// 杠、刻、对子，4顺，3顺 归类
func (self *CardGroup) Classification() map[int]*statisics {
	ret := make(map[int]*statisics)
	all_t := [][]common.EMaJiangType{}
	offset := 0
	if len(self.hand) == 0 {
		return ret
	}
	t := int(self.hand[0]) / 10
	l := len(self.hand)
	for i := 0; i < l; i++ {
		card := self.hand[i]
		card_t := int(card) / 10
		// 最后一张牌，或者牌型变化时，归类
		if card_t != t {
			t = card_t
			all_t = append(all_t, self.hand[offset:i])
			offset = i
		} else if i == l-1 {
			all_t = append(all_t, self.hand[offset:])
		}
	}

	// 统计每种类型 杠 、刻、 顺的数量
	for _, cards := range all_t {
		tcard := int(cards[0]) / 10
		ret[tcard] = &statisics{single: make([]common.EMaJiangType, 0), len: len(cards)}

		same_count := 1       // 相同的数量
		continuous_count := 1 // 连续的数量
		previous_card := cards[0]

		lcards := len(cards)
		for i := 1; i <= lcards; i++ {
			reset_same := true
			reset_continuous := true
			var card common.EMaJiangType
			if i == lcards {
				reset_same = true
				reset_continuous = true
			} else {
				card = cards[i]
				if card == previous_card {
					same_count++
					reset_same = false
					reset_continuous = false
				} else if card == previous_card+1 {
					continuous_count++
					reset_continuous = false
				}
			}

			if same_count == 1 && continuous_count == 1 {
				ret[tcard].single = append(ret[tcard].single, previous_card)
			}
			if same_count > 1 && reset_same {
				ttt := ret[tcard]
				ttt.coincide[same_count]++
				same_count = 1
			}

			if continuous_count > 1 && reset_continuous {
				ret[tcard].continuous[continuous_count]++
				continuous_count = 1
			}

			previous_card = card
		}
	}
	return ret
}
