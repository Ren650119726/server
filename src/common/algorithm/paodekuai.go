package algorithm

import (
	"root/common"
)

// 下家报单, 判断传入牌是否是最大的单张
func PaoDeKuai_IsMaxSingleCard(sCard []common.Card_info, eInCard common.Card_info) bool {
	eEndCard := sCard[0]
	if eInCard[1] >= eEndCard[1] {
		return true
	}
	return false
}
