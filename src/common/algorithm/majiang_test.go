package algorithm

import (
	"root/common"
	"fmt"
	"testing"
)

func TestMajiang(t *testing.T) {

	sHand := []common.EMaJiangType{
		common.TONG_2,
		common.TONG_2,
		common.TONG_3,
		common.TONG_3,
		common.TONG_4,
		common.TONG_4,
		common.TONG_5,
		common.TONG_5,
		common.TONG_7,
		common.TONG_7,
		common.TIAO_2,
	}

	sPeng := [][]common.EMaJiangType{}
	nType := DGK_CalcHuAndExtra(sHand, sPeng, sPeng, 0)
	fmt.Println(nType)
}
