package algorithm

import (
	"root/common"
	"root/core/log"
	"sort"
)

type (
	Card_sorte struct {
		S []Card_info
		A bool // true A做为1    false A做为14
	}
)

var compare_map = make(map[common.EJinHuaType]func([]Card_info, []Card_info) bool)

func init() {
	compare_map[common.ECardType_SANPAI] = compare_sanpai
	compare_map[common.ECardType_DUIZI] = compare_duizi
	compare_map[common.ECardType_SHUNZI] = compare_shunzi
	compare_map[common.ECardType_JINHUA] = compare_jinhua
	compare_map[common.ECardType_SHUNJIN] = compare_shunjin
	compare_map[common.ECardType_BAOZI] = compare_baozi
}
func Compare(a []Card_info, b []Card_info) (awin bool, ta common.EJinHuaType, tb common.EJinHuaType) {
	ta = JudgeCardType(a)
	tb = JudgeCardType(b)
	if ta == tb {
		return compare_map[ta](a, b), ta, tb
	} else {
		return ta > tb, ta, tb
	}

}

func compare_sanpai(a []Card_info, b []Card_info) bool {
	sortTool := &Card_sorte{}
	sortTool.S = a
	sortTool.A = false
	sort.Sort(sortTool)
	sortTool.S = b
	sortTool.A = false
	sort.Sort(sortTool)

	// 先比较点数
	for i := 0; i < 3; i++ {
		if a[i][1] == b[i][1] {
			continue
		}
		ta := a[i][1]
		tb := b[i][1]
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
		if a[i][0] == b[i][0] {
			continue
		}
		return a[i][0] > b[i][0]
	}

	log.Errorf("散牌没有比较出大小 a:%v b:%v", a, b)
	return false
}

func compare_duizi(a []Card_info, b []Card_info) bool {
	sortTool := &Card_sorte{}
	sortTool.S = a
	sort.Sort(sortTool)
	sortTool.S = b
	sort.Sort(sortTool)

	if a[1][1] != b[1][1] {

		ta := a[1][1]
		tb := b[1][1]
		if ta == 1 {
			ta = 14
		}
		if tb == 1 {
			tb = 14
		}
		return ta > tb
	}

	single_a := a[0]
	if a[1][1] == a[0][1] {
		single_a = a[2]
	}
	single_b := b[0]
	if b[1][1] == b[0][1] {
		single_b = b[2]
	}

	if single_a[1] != single_b[1] {
		ta := single_a[1]
		tb := single_b[1]
		if ta == 1 {
			ta = 14
		}
		if tb == 1 {
			tb = 14
		}

		return ta > tb
	}

	// 比花色
	if a[1][0] != b[1][0] {
		return a[1][0] > b[1][0]
	}

	// 比单牌花色
	if single_a[0] != single_b[0] {
		return single_a[0] > single_b[0]
	}
	return false
}

func compare_shunzi(a []Card_info, b []Card_info) bool {
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
		if a[i][0] == b[i][0] {
			continue
		}
		return a[i][0] > b[i][0]
	}

	log.Errorf("顺子没有比较出大小 a:%v b:%v", a, b)
	return false
}

func compare_jinhua(a []Card_info, b []Card_info) bool {
	sortTool := &Card_sorte{}
	sortTool.S = a
	sort.Sort(sortTool)
	sortTool.S = b
	sort.Sort(sortTool)

	// 先比较点数
	for i := 0; i < 3; i++ {
		if a[i][1] == b[i][1] {
			continue
		}
		ta := a[i][1]
		tb := b[i][1]
		if ta == 1 {
			ta = 14
		}
		if tb == 1 {
			tb = 14
		}

		return ta > tb
	}

	if a[0][0] != b[0][0] {
		return a[0][0] > b[0][0]
	}

	log.Errorf("散牌没有比较出大小 a:%v b:%v", a, b)
	return false
}

func compare_shunjin(a []Card_info, b []Card_info) bool {
	maxa := shunzi(a)
	maxb := shunzi(b)

	if maxa != maxb {
		return maxa > maxb
	}

	if a[0][0] != b[0][0] {
		return a[0][0] > b[0][0]
	}

	log.Errorf("顺子没有比较出大小 a:%v b:%v", a, b)
	return false
}

func compare_baozi(a []Card_info, b []Card_info) bool {
	ta := a[2][1]
	tb := b[2][1]
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
		if a[i][0] == b[i][0] {
			continue
		}
		ta := a[i][1]
		tb := b[i][1]
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
	valinum := self.S[i][1]
	valjnum := self.S[j][1]
	valihua := self.S[j][0]
	valjhua := self.S[j][0]

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
