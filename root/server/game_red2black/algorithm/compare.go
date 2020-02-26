package algorithm

import (
	"root/core/log"
	"root/protomsg"
	"sort"
)

type (
	Card_sorte struct {
		S []*protomsg.Card
		A bool // true A做为1    false A做为14
	}
)

var compare_map = make(map[protomsg.RED2BLACKCARDTYPE]func([]*protomsg.Card, []*protomsg.Card) bool)

func init() {
	compare_map[protomsg.RED2BLACKCARDTYPE_RED2BLACK_CARDTYPE_1] = compare_sanpai
	compare_map[protomsg.RED2BLACKCARDTYPE_RED2BLACK_CARDTYPE_2] = compare_duizi
	compare_map[protomsg.RED2BLACKCARDTYPE_RED2BLACK_CARDTYPE_3] = compare_shunzi
	compare_map[protomsg.RED2BLACKCARDTYPE_RED2BLACK_CARDTYPE_4] = compare_jinhua
	compare_map[protomsg.RED2BLACKCARDTYPE_RED2BLACK_CARDTYPE_5] = compare_shunjin
	compare_map[protomsg.RED2BLACKCARDTYPE_RED2BLACK_CARDTYPE_6] = compare_baozi
}
func Compare(a []*protomsg.Card, b []*protomsg.Card) (awin bool, ta protomsg.RED2BLACKCARDTYPE, tb protomsg.RED2BLACKCARDTYPE) {
	ta = JudgeCardType(a)
	tb = JudgeCardType(b)
	if ta == tb {
		return compare_map[ta](a, b), ta, tb
	} else {
		return ta > tb, ta, tb
	}

}

func compare_sanpai(a []*protomsg.Card, b []*protomsg.Card) bool {
	sortTool := &Card_sorte{}
	sortTool.S = a
	sortTool.A = false
	sort.Sort(sortTool)
	sortTool.S = b
	sortTool.A = false
	sort.Sort(sortTool)

	// 先比较点数
	for i := 0; i < 3; i++ {
		if a[i].Number == b[i].Number {
			continue
		}
		ta := a[i].Number
		tb := b[i].Number
		if ta == 1 {
			ta = 14
		}
		if tb == 1 {
			tb = 14
		}
		return ta > tb
	}

	// 再比较花色
	for i := 0; i < 3; i++ {
		if a[i].Color == b[i].Color {
			continue
		}
		return a[i].Color > b[i].Color
	}

	log.Errorf("散牌没有比较出大小 a:%v b:%v", a, b)
	return false
}

func compare_duizi(a []*protomsg.Card, b []*protomsg.Card) bool {
	sortTool := &Card_sorte{}
	sortTool.S = a
	sort.Sort(sortTool)
	sortTool.S = b
	sort.Sort(sortTool)

	if a[1].Number != b[1].Number {

		ta := a[1].Number
		tb := b[1].Number
		if ta == 1 {
			ta = 14
		}
		if tb == 1 {
			tb = 14
		}
		return ta > tb
	}

	single_a := a[0]
	if a[1].Number == a[0].Number {
		single_a = a[2]
	}
	single_b := b[0]
	if b[1].Number == b[0].Number {
		single_b = b[2]
	}

	if single_a.Number != single_b.Number {
		ta := single_a.Number
		tb := single_b.Number
		if ta == 1 {
			ta = 14
		}
		if tb == 1 {
			tb = 14
		}

		return ta > tb
	}

	// 比花色
	if a[1].Color != b[1].Color {
		return a[1].Color > b[1].Color
	}

	// 比单牌花色
	if single_a.Color != single_b.Color {
		return single_a.Color > single_b.Color
	}
	return false
}

func compare_shunzi(a []*protomsg.Card, b []*protomsg.Card) bool {
	maxa := shunzi(a)
	maxb := shunzi(b)

	if maxa != maxb {
		return maxa > maxb
	}

	sortTool := &Card_sorte{}
	if maxa == 14 {
		sortTool.A = false
	} else {
		sortTool.A = true
	}
	sortTool.S = a
	sort.Sort(sortTool)
	sortTool.S = b
	sort.Sort(sortTool)

	// 比较花色
	for i := 0; i < 3; i++ {
		if a[i].Color == b[i].Color {
			continue
		}
		return a[i].Color > b[i].Color
	}

	log.Errorf("顺子没有比较出大小 a:%v b:%v", a, b)
	return false
}

func compare_jinhua(a []*protomsg.Card, b []*protomsg.Card) bool {
	sortTool := &Card_sorte{}
	sortTool.S = a
	sort.Sort(sortTool)
	sortTool.S = b
	sort.Sort(sortTool)

	// 先比较点数
	for i := 0; i < 3; i++ {
		if a[i].Number == b[i].Number {
			continue
		}
		ta := a[i].Number
		tb := b[i].Number
		if ta == 1 {
			ta = 14
		}
		if tb == 1 {
			tb = 14
		}

		return ta > tb
	}

	if a[0].Color != b[0].Color {
		return a[0].Color > b[0].Color
	}

	log.Errorf("散牌没有比较出大小 a:%v b:%v", a, b)
	return false
}

func compare_shunjin(a []*protomsg.Card, b []*protomsg.Card) bool {
	maxa := shunzi(a)
	maxb := shunzi(b)

	if maxa != maxb {
		return maxa > maxb
	}

	if a[0].Color != b[0].Color {
		return a[0].Color > b[0].Color
	}

	log.Errorf("顺子没有比较出大小 a:%v b:%v", a, b)
	return false
}

func compare_baozi(a []*protomsg.Card, b []*protomsg.Card) bool {
	ta := a[2].Number
	tb := b[2].Number
	if ta == 1 {
		ta = 14
	}
	if tb == 1 {
		tb = 14
	}
	if ta != tb {
		return ta > tb
	}

	sortTool := &Card_sorte{}
	sortTool.S = a
	sort.Sort(sortTool)
	sortTool.S = b
	sort.Sort(sortTool)

	// 比较花色
	for i := 0; i < 3; i++ {
		if a[i].Color == b[i].Color {
			continue
		}
		ta := a[i].Number
		tb := b[i].Number
		if ta == 1 {
			ta = 14
		}
		if tb == 1 {
			tb = 14
		}

		return ta > tb
	}

	log.Errorf("豹子没有比较出大小 a:%v b:%v", a, b)
	return false
}

////////////////////////////////////////////////////////////////////////////////////////////////////////
func (self *Card_sorte) Len() int {
	return len(self.S)
}
func (self *Card_sorte) Less(i, j int) bool {
	valinum := self.S[i].Number
	valjnum := self.S[j].Number
	valihua := self.S[j].Color
	valjhua := self.S[j].Color

	if !self.A {
		if valinum == 1 {
			valinum = 14
		}
		if valjnum == 1 {
			valjnum = 14
		}
	}

	if valinum == valjnum {
		return valihua > valjhua
	} else {
		return valinum > valjnum
	}
}
func (self *Card_sorte) Swap(i, j int) {
	self.S[i], self.S[j] = self.S[j], self.S[i]
}
