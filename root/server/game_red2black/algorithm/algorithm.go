package algorithm

import (
	"math/rand"
	"root/core/log"
	"root/core/utils"
	"root/protomsg"
)


// 随机获得不重复的n张牌
func GetRandom_Card(cards []*protomsg.Card, count int)[]*protomsg.Card {
	rand.Seed(utils.SecondTimeSince1970())
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

// 传入3张牌，判断出牌型
func JudgeCardType(cards []*protomsg.Card) protomsg.RED2BLACKCARDTYPE {
	if len(cards) != 3 {
		log.Errorf("传入的牌不是3张 len:%+v", len(cards))
		return 0
	}
	// 豹子
	if cards[0].Number == cards[1].Number && cards[1].Number == cards[2].Number {
		return protomsg.RED2BLACKCARDTYPE_RED2BLACK_CARDTYPE_6
	}

	// 有没有顺子
	maxShunzi := shunzi(cards)
	// 有没有金花
	jinhua := false
	if cards[0].Color == cards[1].Color && cards[1].Color == cards[2].Color {
		jinhua = true
	}

	// 顺金
	if maxShunzi > 0 && jinhua {
		return protomsg.RED2BLACKCARDTYPE_RED2BLACK_CARDTYPE_5
	}

	// 金花
	if jinhua {
		return protomsg.RED2BLACKCARDTYPE_RED2BLACK_CARDTYPE_4
	}
	// 顺子
	if maxShunzi > 0 {
		return protomsg.RED2BLACKCARDTYPE_RED2BLACK_CARDTYPE_3
	}
	// 对子
	if cards[0].Number == cards[1].Number || cards[0].Number == cards[2].Number || cards[1].Number == cards[2].Number {
		return protomsg.RED2BLACKCARDTYPE_RED2BLACK_CARDTYPE_2
	}

	// 散牌
	return protomsg.RED2BLACKCARDTYPE_RED2BLACK_CARDTYPE_1
}

// 判断是否有顺子，并且获得最大的一张牌, 没有顺子max == 0
func shunzi(cards []*protomsg.Card) (max uint8) {
	temp := [15]uint8{}

	for _, card := range cards {
		if card.Number >= int32(len(temp)) {
			log.Errorf("牌错误：%card ", card, cards)
			return 0
		}

		temp[card.Number] += 1
		if card.Number == 1 {
			temp[14] += 1
		}
	}

	// 如果有连续三张 就找到了顺子
	continu := 0
	for i, v := range temp {
		if v > 0 {
			continu++
			if continu == 3 {
				max = uint8(i)
				return
			}
		} else {
			continu = 0
		}
	}
	return
}
