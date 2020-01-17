package algorithm

import (
	"root/core/log"
	"testing"
)

func Test(t *testing.T) {
	te_val := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	te_val = append(te_val[:0], te_val[1:]...)
	//j := algorithm.MajiangCalcHuType([]common.EMaJiangType{common.TIAO_3, common.TIAO_3, common.TIAO_4, common.TIAO_4, common.TIAO_4},
	//	[][]common.EMaJiangType{{common.TONG_1, common.TONG_1, common.TONG_1}},
	//	[][]common.EMaJiangType{{common.TIAO_1, common.TIAO_1, common.TIAO_1, common.TIAO_1}})
	log.Infof("%v", te_val)
}
