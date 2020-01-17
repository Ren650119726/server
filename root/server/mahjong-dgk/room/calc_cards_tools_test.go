package room

import (
	"root/common"
	"root/core/log"
	"testing"
)

func Test(t *testing.T) {
	g := NewCardGroup()
	g.hand = []common.EMaJiangType{10, 12, 12, 13, 13, 13, 14, 15, 15, 16, 18}
	log.Infof("%v", g.Classification())
	log.Infof("")
}
