package room

import (
	"root/common"
	"testing"
)

func Test(t *testing.T) {
	g := NewCardGroup()
	g.hand = []common.EMaJiangType{23, 25, 26, 27, 27, 28, 29, 33, 34, 35, 36, 36, 38, 39}

	auto_decide(g)
}

func auto_decide(g *CardGroup) int {
	hand := g.hand
	tt := []int{0, 0, 0, 0}
	for _, v := range hand {
		t := int(v) / 10
		tt[t]++
	}

	min := tt[1]
	equalt := 0
	t := 1
	for i := 2; i < len(tt); i++ {
		if min > tt[i] {
			equalt = 0
			t = i
			min = tt[i]
		} else if min == tt[i] {
			equalt = i
		}
	}

	if equalt != 0 {
		set := g.Classification()

		coincide_choose := 0
		coincide_choose_num := 0
		for i := 4; i >= 0; i-- {
			if set[t].coincide[i] <= set[equalt].coincide[i] {
				coincide_choose = t
				coincide_choose_num = i
				break
			} else if set[t].coincide[i] > set[equalt].coincide[i] {
				coincide_choose = equalt
				coincide_choose_num = i
				break
			}
		}

		continuous_choose := 0
		continuous_choose_num := 0
		for i := 4; i >= 0; i-- {
			if set[t].continuous[i] <= set[equalt].continuous[i] {
				continuous_choose = t
				continuous_choose_num = i
				break
			} else if set[t].continuous[i] > set[equalt].continuous[i] {
				continuous_choose = equalt
				continuous_choose_num = i
				break
			}
		}

		if coincide_choose_num >= continuous_choose_num {
			return coincide_choose
		} else {
			return continuous_choose
		}
	} else {
		return t
	}
}
