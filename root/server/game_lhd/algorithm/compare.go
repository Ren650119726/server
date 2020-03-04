package algorithm

import (
	"root/protomsg"
)

type (
	Card_sorte struct {
		S []*protomsg.Card
		A bool // true A做为1    false A做为14
	}
)

func Compare(a *protomsg.Card, b *protomsg.Card) (ret protomsg.LHDAREA) {
	if a.Number > b.Number {
		return protomsg.LHDAREA_LHD_AREA_DRAGON
	} else if a.Number == b.Number {
		return protomsg.LHDAREA_LHD_AREA_PEACE
	} else {
		return protomsg.LHDAREA_LHD_AREA_TIGER
	}
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
