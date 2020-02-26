package algorithm

import (
	"common"
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	red := []Card_info{
		{common.ECardType_HEITAO.UInt8(), 7},
		{common.ECardType_HEITAO.UInt8(), 5},
		{common.ECardType_HONGTAO.UInt8(), 2},
	}

	black := []Card_info{
		{common.ECardType_MEIHUA.UInt8(), 6},
		{common.ECardType_HEITAO.UInt8(), 7},
		{common.ECardType_HEITAO.UInt8(), 5},
	}

	reddWin, tred, tblack := Compare(red, black)

	fmt.Println(reddWin, tred, tblack)
}
