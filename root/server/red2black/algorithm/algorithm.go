package algorithm

import (
	"root/common"
	"root/core/log"
	"root/core/utils"
	"math/rand"
)

type Card_info [2]uint8

var cards = []Card_info{
	{common.ECardType_HEITAO.UInt8(), 1},
	{common.ECardType_HEITAO.UInt8(), 2},
	{common.ECardType_HEITAO.UInt8(), 3},
	{common.ECardType_HEITAO.UInt8(), 4},
	{common.ECardType_HEITAO.UInt8(), 5},
	{common.ECardType_HEITAO.UInt8(), 6},
	{common.ECardType_HEITAO.UInt8(), 7},
	{common.ECardType_HEITAO.UInt8(), 8},
	{common.ECardType_HEITAO.UInt8(), 9},
	{common.ECardType_HEITAO.UInt8(), 10},
	{common.ECardType_HEITAO.UInt8(), 11},
	{common.ECardType_HEITAO.UInt8(), 12},
	{common.ECardType_HEITAO.UInt8(), 13},

	{common.ECardType_HONGTAO.UInt8(), 1},
	{common.ECardType_HONGTAO.UInt8(), 2},
	{common.ECardType_HONGTAO.UInt8(), 3},
	{common.ECardType_HONGTAO.UInt8(), 4},
	{common.ECardType_HONGTAO.UInt8(), 5},
	{common.ECardType_HONGTAO.UInt8(), 6},
	{common.ECardType_HONGTAO.UInt8(), 7},
	{common.ECardType_HONGTAO.UInt8(), 8},
	{common.ECardType_HONGTAO.UInt8(), 9},
	{common.ECardType_HONGTAO.UInt8(), 10},
	{common.ECardType_HONGTAO.UInt8(), 11},
	{common.ECardType_HONGTAO.UInt8(), 12},
	{common.ECardType_HONGTAO.UInt8(), 13},

	{common.ECardType_MEIHUA.UInt8(), 1},
	{common.ECardType_MEIHUA.UInt8(), 2},
	{common.ECardType_MEIHUA.UInt8(), 3},
	{common.ECardType_MEIHUA.UInt8(), 4},
	{common.ECardType_MEIHUA.UInt8(), 5},
	{common.ECardType_MEIHUA.UInt8(), 6},
	{common.ECardType_MEIHUA.UInt8(), 7},
	{common.ECardType_MEIHUA.UInt8(), 8},
	{common.ECardType_MEIHUA.UInt8(), 9},
	{common.ECardType_MEIHUA.UInt8(), 10},
	{common.ECardType_MEIHUA.UInt8(), 11},
	{common.ECardType_MEIHUA.UInt8(), 12},
	{common.ECardType_MEIHUA.UInt8(), 13},

	{common.ECardType_FANGKUAI.UInt8(), 1},
	{common.ECardType_FANGKUAI.UInt8(), 2},
	{common.ECardType_FANGKUAI.UInt8(), 3},
	{common.ECardType_FANGKUAI.UInt8(), 4},
	{common.ECardType_FANGKUAI.UInt8(), 5},
	{common.ECardType_FANGKUAI.UInt8(), 6},
	{common.ECardType_FANGKUAI.UInt8(), 7},
	{common.ECardType_FANGKUAI.UInt8(), 8},
	{common.ECardType_FANGKUAI.UInt8(), 9},
	{common.ECardType_FANGKUAI.UInt8(), 10},
	{common.ECardType_FANGKUAI.UInt8(), 11},
	{common.ECardType_FANGKUAI.UInt8(), 12},
	{common.ECardType_FANGKUAI.UInt8(), 13},
}

// 随机获得不重复的n张牌
func GetRandom_Card(count int) []Card_info {
	rand.Seed(utils.SecondTimeSince1970())
	ret := make([]Card_info, 0, count)
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
func JudgeCardType(cards []Card_info) common.EJinHuaType {
	if len(cards) != 3 {
		log.Errorf("传入的牌不是3张 len:%v", len(cards))
		return 0
	}
	// 豹子
	if cards[0][1] == cards[1][1] && cards[1][1] == cards[2][1] {
		return common.ECardType_BAOZI
	}

	// 有没有顺子
	maxShunzi := shunzi(cards)
	// 有没有金花
	jinhua := false
	if cards[0][0] == cards[1][0] && cards[1][0] == cards[2][0] {
		jinhua = true
	}

	// 顺金
	if maxShunzi > 0 && jinhua {
		return common.ECardType_SHUNJIN
	}

	// 金花
	if jinhua {
		return common.ECardType_JINHUA
	}
	// 顺子
	if maxShunzi > 0 {
		return common.ECardType_SHUNZI
	}
	// 对子
	if cards[0][1] == cards[1][1] || cards[0][1] == cards[2][1] || cards[1][1] == cards[2][1] {
		return common.ECardType_DUIZI
	}

	// 散牌
	return common.ECardType_SANPAI
}

// 判断是否有顺子，并且获得最大的一张牌, 没有顺子max == 0
func shunzi(cards []Card_info) (max uint8) {
	temp := [15]uint8{}

	for _, v := range cards {
		if v[1] >= uint8(len(temp)) {
			log.Errorf("牌错误：%v ", v, cards)
			return 0
		}

		temp[v[1]] += 1
		if v[1] == 1 {
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
