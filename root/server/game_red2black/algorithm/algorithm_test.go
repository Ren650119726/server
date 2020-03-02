package algorithm

import (
	"fmt"
	"root/protomsg"
	"testing"
)

func Test(t *testing.T) {
	red := []*protomsg.Card{
		{2, protomsg.Card_CARDCOLOR_4},
		{6, protomsg.Card_CARDCOLOR_2},
		{1, protomsg.Card_CARDCOLOR_3},
	}

	black := []*protomsg.Card{
		{8, protomsg.Card_CARDCOLOR_4},
		{2, protomsg.Card_CARDCOLOR_2},
		{7, protomsg.Card_CARDCOLOR_2},
	}

	reddWin, tred, tblack := Compare(red, black)

	fmt.Println(reddWin, tred, tblack)
}
