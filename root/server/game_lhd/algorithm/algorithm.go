package algorithm

import (
	"root/core/utils"
	"root/protomsg"
)

// 随机获得不重复的n张牌
func GetRandom_Card(cards []*protomsg.Card, count int) []*protomsg.Card {
	ret := make([]*protomsg.Card, 0, count)
	for i := 0; i < count; i++ {
		last := len(cards) - 1 - i
		if last == 0 {
			ret = append(ret, cards[0])
			continue
		}
		rand_val := utils.Randx_y(0, last)

		ret = append(ret, cards[rand_val])
		cards[rand_val], cards[last] = cards[last], cards[rand_val]
	}

	return ret
}
