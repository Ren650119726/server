package algorithm

import (
	"root/common"
	"root/core/log"
	"root/core/utils"
	"math/rand"
	"sort"
)

var PDK_HN_ZHANIAO_CARD = common.Card_info{common.ECardType_HONGTAO.UInt8(), 10}

var pdk_hn_cards_one = []common.Card_info{
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
	{common.ECardType_HEITAO.UInt8(), 14},
	{common.ECardType_HEITAO.UInt8(), 102},
}

var pdk_hn_cards_two = []common.Card_info{
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
	{common.ECardType_FANGKUAI.UInt8(), 14},
	{common.ECardType_FANGKUAI.UInt8(), 102},

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
	{common.ECardType_MEIHUA.UInt8(), 14},

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
	{common.ECardType_HONGTAO.UInt8(), 14},

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
}

// 预先配置所有牌型的最小点数; 用于智能出牌策略
// 调整需配合pdk_hn_intelligent_out_card_type一起改
var pdk_hn_intelligent_duizi_card = []common.Card_info{{1, 2}, {1, 2}}
var pdk_hn_intelligent_danzhang_card = []common.Card_info{{1, 2}}
var pdk_hn_intelligent_out_card = [][]common.Card_info{
	{{1, 5}, {1, 5}, {1, 5}, {1, 4}, {1, 4}, {1, 4}, {1, 3}, {1, 3}, {1, 3}, {1, 2}, {1, 2}, {1, 2}, {1, 13}, {1, 14}, {1, 11}, {1, 12}}, // 飞机带翅膀 (4飞机) 副牌四张
	{{1, 6}, {1, 6}, {1, 6}, {1, 5}, {1, 5}, {1, 5}, {1, 4}, {1, 4}, {1, 4}, {1, 3}, {1, 3}, {1, 3}, {1, 2}, {1, 2}, {1, 2}, {1, 12}},    // 飞机带翅膀 (5飞机) 副牌一张
	{{1, 9}, {1, 9}, {1, 8}, {1, 8}, {1, 7}, {1, 7}, {1, 6}, {1, 6}, {1, 5}, {1, 5}, {1, 4}, {1, 4}, {1, 3}, {1, 3}, {1, 2}, {1, 2}},     // 双顺
	{{1, 4}, {1, 4}, {1, 4}, {1, 3}, {1, 3}, {1, 3}, {1, 2}, {1, 2}, {1, 2}, {1, 10}, {1, 11}, {1, 12}, {1, 14}, {1, 7}, {1, 8}},         // 飞机带翅膀 (3飞机) 副牌六张
	{{1, 5}, {1, 5}, {1, 5}, {1, 4}, {1, 4}, {1, 4}, {1, 3}, {1, 3}, {1, 3}, {1, 2}, {1, 2}, {1, 2}, {1, 7}, {1, 8}, {1, 9}},             // 飞机带翅膀 (4飞机) 副牌三张
	{{1, 6}, {1, 6}, {1, 6}, {1, 5}, {1, 5}, {1, 5}, {1, 4}, {1, 4}, {1, 4}, {1, 3}, {1, 3}, {1, 3}, {1, 2}, {1, 2}, {1, 2}},             // 三顺
	{{1, 8}, {1, 8}, {1, 7}, {1, 7}, {1, 6}, {1, 6}, {1, 5}, {1, 5}, {1, 4}, {1, 4}, {1, 3}, {1, 3}, {1, 2}, {1, 2}},                     // 双顺
	{{1, 5}, {1, 5}, {1, 5}, {1, 4}, {1, 4}, {1, 4}, {1, 3}, {1, 3}, {1, 3}, {1, 2}, {1, 2}, {1, 2}, {1, 7}, {1, 8}},                     // 飞机带翅膀 (4飞机) 副牌二张
	{{1, 4}, {1, 4}, {1, 4}, {1, 3}, {1, 3}, {1, 3}, {1, 2}, {1, 2}, {1, 2}, {1, 7}, {1, 8}, {1, 9}, {1, 11}, {1, 12}},                   // 飞机带翅膀 (3飞机) 副牌五张
	{{1, 5}, {1, 5}, {1, 5}, {1, 4}, {1, 4}, {1, 4}, {1, 3}, {1, 3}, {1, 3}, {1, 2}, {1, 2}, {1, 2}, {1, 7}},                             // 飞机带翅膀 (4飞机) 副牌一张
	{{1, 4}, {1, 4}, {1, 4}, {1, 3}, {1, 3}, {1, 3}, {1, 2}, {1, 2}, {1, 2}, {1, 7}, {1, 8}, {1, 9}, {1, 11}},                            // 飞机带翅膀 (3飞机) 副牌四张
	{{1, 13}, {1, 12}, {1, 11}, {1, 10}, {1, 9}, {1, 8}, {1, 7}, {1, 6}, {1, 5}, {1, 4}, {1, 3}, {1, 2}},                                 // 单顺
	{{1, 5}, {1, 5}, {1, 5}, {1, 4}, {1, 4}, {1, 4}, {1, 3}, {1, 3}, {1, 3}, {1, 2}, {1, 2}, {1, 2}},                                     // 三顺
	{{1, 7}, {1, 7}, {1, 6}, {1, 6}, {1, 5}, {1, 5}, {1, 4}, {1, 4}, {1, 3}, {1, 3}, {1, 2}, {1, 2}},                                     // 双顺
	{{1, 4}, {1, 4}, {1, 4}, {1, 3}, {1, 3}, {1, 3}, {1, 2}, {1, 2}, {1, 2}, {1, 7}, {1, 8}, {1, 9}},                                     // 飞机带翅膀 (3飞机) 副牌三张
	{{1, 12}, {1, 11}, {1, 10}, {1, 9}, {1, 8}, {1, 7}, {1, 6}, {1, 5}, {1, 4}, {1, 3}, {1, 2}},                                          // 单顺
	{{1, 4}, {1, 4}, {1, 4}, {1, 3}, {1, 3}, {1, 3}, {1, 2}, {1, 2}, {1, 2}, {1, 7}, {1, 8}},                                             // 飞机带翅膀 (3飞机) 副牌二张
	{{1, 11}, {1, 10}, {1, 9}, {1, 8}, {1, 7}, {1, 6}, {1, 5}, {1, 4}, {1, 3}, {1, 2}},                                                   // 单顺
	{{1, 6}, {1, 6}, {1, 5}, {1, 5}, {1, 4}, {1, 4}, {1, 3}, {1, 3}, {1, 2}, {1, 2}},                                                     // 双顺
	{{1, 4}, {1, 4}, {1, 4}, {1, 3}, {1, 3}, {1, 3}, {1, 2}, {1, 2}, {1, 2}, {1, 7}},                                                     // 飞机带翅膀 (3飞机) 副牌一张
	{{1, 3}, {1, 3}, {1, 3}, {1, 2}, {1, 2}, {1, 2}, {1, 7}, {1, 8}, {1, 9}, {1, 6}},                                                     // 飞机带翅膀 (2飞机) 副牌四张
	{{1, 10}, {1, 9}, {1, 8}, {1, 7}, {1, 6}, {1, 5}, {1, 4}, {1, 3}, {1, 2}},                                                            // 单顺
	{{1, 4}, {1, 4}, {1, 4}, {1, 3}, {1, 3}, {1, 3}, {1, 2}, {1, 2}, {1, 2}},                                                             // 三顺
	{{1, 3}, {1, 3}, {1, 3}, {1, 2}, {1, 2}, {1, 2}, {1, 7}, {1, 8}, {1, 9}},                                                             // 飞机带翅膀 (2飞机) 副牌三张
	{{1, 9}, {1, 8}, {1, 7}, {1, 6}, {1, 5}, {1, 4}, {1, 3}, {1, 2}},                                                                     // 单顺
	{{1, 5}, {1, 5}, {1, 4}, {1, 4}, {1, 3}, {1, 3}, {1, 2}, {1, 2}},                                                                     // 双顺
	{{1, 3}, {1, 3}, {1, 3}, {1, 2}, {1, 2}, {1, 2}, {1, 7}, {1, 8}},                                                                     // 飞机带翅膀 (2飞机) 副牌两张
	{{1, 3}, {1, 3}, {1, 3}, {1, 2}, {1, 2}, {1, 2}, {1, 7}},                                                                             // 飞机带翅膀 (2飞机) 副牌一张
	{{1, 8}, {1, 7}, {1, 6}, {1, 5}, {1, 4}, {1, 3}, {1, 2}},                                                                             // 单顺
	{{1, 2}, {1, 2}, {1, 2}, {1, 2}, {1, 6}, {1, 9}, {1, 7}},                                                                             // 四张带副牌三张
	{{1, 7}, {1, 6}, {1, 5}, {1, 4}, {1, 3}, {1, 2}},                                                                                     // 单顺
	{{1, 3}, {1, 3}, {1, 3}, {1, 2}, {1, 2}, {1, 2}},                                                                                     // 三顺
	{{1, 4}, {1, 4}, {1, 3}, {1, 3}, {1, 2}, {1, 2}},                                                                                     // 双顺
	{{1, 2}, {1, 2}, {1, 2}, {1, 2}, {1, 8}, {1, 7}},                                                                                     // 四张带副牌二张
	{{1, 2}, {1, 2}, {1, 2}, {1, 8}, {1, 7}},                                                                                             // 三张带副牌二张
	{{1, 6}, {1, 5}, {1, 4}, {1, 3}, {1, 2}},                                                                                             // 单顺
	{{1, 2}, {1, 2}, {1, 2}, {1, 2}, {1, 7}},                                                                                             // 三张带副牌二张
	{{1, 3}, {1, 3}, {1, 2}, {1, 2}},                                                                                                     // 双顺
	{{1, 2}, {1, 2}, {1, 2}, {1, 7}},                                                                                                     // 三张带副牌一张
	{{1, 2}, {1, 2}, {1, 2}},                                                                                                             // 三张
}

// 配合上面牌组的牌型
var pdk_hn_intelligent_out_card_type = []uint8{
	common.PDK_FEI_JI_4.Value(),
	common.PDK_FEI_JI_5.Value(),
	common.PDK_SHUANG_SHUN.Value(),
	common.PDK_FEI_JI_3.Value(),
	common.PDK_FEI_JI_4.Value(),
	common.PDK_SAN_SHUN.Value(),
	common.PDK_SHUANG_SHUN.Value(),
	common.PDK_FEI_JI_4.Value(),
	common.PDK_FEI_JI_3.Value(),
	common.PDK_FEI_JI_4.Value(),
	common.PDK_FEI_JI_3.Value(),
	common.PDK_SHUN_ZI.Value(),
	common.PDK_SAN_SHUN.Value(),
	common.PDK_SHUANG_SHUN.Value(),
	common.PDK_FEI_JI_3.Value(),
	common.PDK_SHUN_ZI.Value(),
	common.PDK_FEI_JI_3.Value(),
	common.PDK_SHUN_ZI.Value(),
	common.PDK_SHUANG_SHUN.Value(),
	common.PDK_FEI_JI_3.Value(),
	common.PDK_FEI_JI_2.Value(),
	common.PDK_SHUN_ZI.Value(),
	common.PDK_SAN_SHUN.Value(),
	common.PDK_FEI_JI_2.Value(),
	common.PDK_SHUN_ZI.Value(),
	common.PDK_SHUANG_SHUN.Value(),
	common.PDK_FEI_JI_2.Value(),
	common.PDK_FEI_JI_2.Value(),
	common.PDK_SHUN_ZI.Value(),
	common.PDK_SI_DAI.Value(),
	common.PDK_SHUN_ZI.Value(),
	common.PDK_SAN_SHUN.Value(),
	common.PDK_SHUANG_SHUN.Value(),
	common.PDK_SI_DAI.Value(),
	common.PDK_SAN_DAI.Value(),
	common.PDK_SHUN_ZI.Value(),
	common.PDK_SAN_DAI.Value(),
	common.PDK_SHUANG_SHUN.Value(),
	common.PDK_SAN_DAI.Value(),
	common.PDK_SAN_ZHANG.Value(),
}

// 洗牌, 按照洗牌规则
// 第一参数, 每人15张牌: 总牌45张; 去掉大小王, 去掉三个2, 去掉三个A, 去掉一个K
// 第一参数, 每人16张牌: 总牌48张; 去掉大小王, 去掉三个2, 去掉一个A
func PaoDeKuai_HN_ShuffleCard(nShuffleMode uint8) []common.Card_info {

	rand.Seed(utils.SecondTimeSince1970())
	var sCard []common.Card_info
	if nShuffleMode == 15 {
		utils.RandomSlice(pdk_hn_cards_one)
		sCard = append(sCard, pdk_hn_cards_one...)
	} else if nShuffleMode == 16 {
		utils.RandomSlice(pdk_hn_cards_two)
		sCard = append(sCard, pdk_hn_cards_two...)
	}
	return sCard
}

// 函数作用: 判断传入的牌切片对应的牌型
func PaoDeKuai_HN_CalcCardType(sCard []common.Card_info) (common.EPaoDeKuai, uint8) {

	nLen := len(sCard)
	if nLen <= 0 || nLen > 16 {
		log.Errorf("传入切片长度异常: %v", nLen)
		return common.PDK_NIL, 0 // 无牌型
	}
	if nLen == 1 {
		return common.PDK_DAN_ZHANG, 0 // 单张
	}
	c1 := sCard[0]
	c2 := sCard[1]
	if nLen == 2 {
		if c1[1] == c2[1] {
			return common.PDK_DUI_ZI, 0 // 对子
		}
		return common.PDK_NIL, 0 // 无牌型
	}
	c3 := sCard[2]
	if nLen == 3 {
		if c1[1] == c2[1] && c2[1] == c3[1] {
			return common.PDK_SAN_ZHANG, 0 // 三张
		}
		return common.PDK_NIL, 0 // 无牌型
	}
	c4 := sCard[3]
	if nLen == 4 {
		if c1[1] == c2[1] && c2[1] == c3[1] && c3[1] == c4[1] {
			return common.PDK_ZHA_DAN, 0 // 炸弹
		} else if c1[1] == c2[1] && c3[1] == c4[1] && c2[1]-1 == c3[1] {
			return common.PDK_SHUANG_SHUN, 0 // 双顺
		} else if c2[1] == c3[1] && c3[1] == c4[1] {
			return common.PDK_SAN_DAI, 1 // 三张带一张
		} else if c1[1] == c2[1] && c2[1] == c3[1] {
			return common.PDK_SAN_DAI, 1 // 三张带一张
		}
		return common.PDK_NIL, 0 // 无牌型
	}
	c5 := sCard[4]
	if nLen == 5 {
		if c1[1] == c2[1] && c2[1] == c3[1] && c3[1] == c4[1] {
			return common.PDK_SAN_DAI, 2 // 三张带二张 (炸弹带一张)
		} else if c2[1] == c3[1] && c3[1] == c4[1] && c4[1] == c5[1] {
			return common.PDK_SAN_DAI, 2 // 三张带二张 (炸弹带一张)
		} else if c1[1]-1 == c2[1] && c2[1]-1 == c3[1] && c3[1]-1 == c4[1] && c4[1]-1 == c5[1] {
			return common.PDK_SHUN_ZI, 0 // 顺子
		} else if c1[1] == c2[1] && c2[1] == c3[1] {
			return common.PDK_SAN_DAI, 2 // 三张带二张
		} else if c2[1] == c3[1] && c3[1] == c4[1] {
			return common.PDK_SAN_DAI, 2 // 三张带二张
		} else if c3[1] == c4[1] && c4[1] == c5[1] {
			return common.PDK_SAN_DAI, 2 // 三张带二张
		}
		return common.PDK_NIL, 0 // 无牌型
	}
	if nLen == 6 {
		c6 := sCard[5]
		if c1[1] == c2[1] && c2[1] == c3[1] && c3[1] == c4[1] {
			return common.PDK_SI_DAI, 2 // 四张带二张 (炸弹带二张)
		} else if c2[1] == c3[1] && c3[1] == c4[1] && c4[1] == c5[1] {
			return common.PDK_SI_DAI, 2 // 四张带二张 (炸弹带二张)
		} else if c3[1] == c4[1] && c4[1] == c5[1] && c5[1] == c6[1] {
			return common.PDK_SI_DAI, 2 // 四张带二张 (炸弹带二张)
		}
	}
	if nLen == 7 {
		c6 := sCard[5]
		c7 := sCard[6]
		if c1[1] == c2[1] && c2[1] == c3[1] && c3[1] == c4[1] {
			return common.PDK_SI_DAI, 3 // 四张带三张 (炸弹带三张)
		} else if c2[1] == c3[1] && c3[1] == c4[1] && c4[1] == c5[1] {
			return common.PDK_SI_DAI, 3 // 四张带三张 (炸弹带三张)
		} else if c3[1] == c4[1] && c4[1] == c5[1] && c5[1] == c6[1] {
			return common.PDK_SI_DAI, 3 // 四张带三张 (炸弹带三张)
		} else if c4[1] == c5[1] && c5[1] == c6[1] && c6[1] == c7[1] {
			return common.PDK_SI_DAI, 3 // 四张带三张 (炸弹带三张)
		}
	}

	eShunZi := pdk_HN_IsShunZi(sCard)
	if eShunZi != common.PDK_NIL {
		return eShunZi, 0 // 顺子
	}
	eShuangShun := pdk_HN_IsShuangShun(sCard)
	if eShuangShun != common.PDK_NIL {
		return eShuangShun, 0 // 双顺
	}
	eSanShun := pdk_HN_IsSanShun(sCard)
	if eSanShun != common.PDK_NIL {
		return eSanShun, 0 // 三顺
	}
	eFeiJiDaiCB, nSubLen := pdk_HN_IsFeiJiDaiCB(sCard)
	if eFeiJiDaiCB != common.PDK_NIL {
		return eFeiJiDaiCB, nSubLen // 飞机带翅膀
	}
	return common.PDK_NIL, 0 // 无牌型
}

func PaoDeKuai_HN_IsHaveBiggerCard(sCard, sInCard []common.Card_info) bool {

	nCardLen := uint8(len(sCard))
	nInLen := uint8(len(sInCard))
	if nInLen < 1 {
		return false
	}
	nInPoint := sInCard[0][1]
	mHandCount := Poker_StatPointCount(sCard)
	eInType, _ := PaoDeKuai_HN_CalcCardType(sInCard)
	switch eInType {
	case common.PDK_DAN_ZHANG: // 单张
		eEndCard := sCard[0]
		if eEndCard[1] > nInPoint {
			return true
		}
		// 从大到小的方式, 找不到, 再找炸弹
		nOutPoint := pdk_HN_GetOutPointByCount(sCard, mHandCount, 0, 4)
		if nOutPoint > 0 {
			return true
		}
	case common.PDK_DUI_ZI: // 对子
		nOutPoint := pdk_HN_GetOutPointByCount(sCard, mHandCount, nInPoint, 2)
		if nOutPoint > 0 {
			return true
		}
		// 从大到小的方式, 找不到, 再找炸弹
		nOutPoint = pdk_HN_GetOutPointByCount(sCard, mHandCount, 0, 4)
		if nOutPoint > 0 {
			return true
		}
	case common.PDK_SAN_ZHANG: // 三张
		nOutPoint := pdk_HN_GetOutPointByCount(sCard, mHandCount, nInPoint, 3)
		if nOutPoint > 0 {
			return true
		}
		// 从大到小的方式, 找不到, 再找炸弹
		nOutPoint = pdk_HN_GetOutPointByCount(sCard, mHandCount, 0, 4)
		if nOutPoint > 0 {
			return true
		}
	case common.PDK_SAN_DAI: // 三张带二张
		if nInLen == 5 {
			nInPoint = pdk_HN_GetInPointByCount(sInCard, 3)
			nOutPoint := pdk_HN_GetOutPointByCount(sCard, mHandCount, nInPoint, 3)
			if nOutPoint > 0 && nCardLen >= 5 {
				return true
			}
			// 从大到小的方式, 找不到, 再找炸弹
			nOutPoint = pdk_HN_GetOutPointByCount(sCard, mHandCount, 0, 4)
			if nOutPoint > 0 {
				return true
			}
		}
	case common.PDK_SHUN_ZI: // 顺子
		nStartPoint, nEndPoint := pdk_HN_GetCardByShunZi(sCard, mHandCount, nInPoint, nInLen, 1)
		if nStartPoint > 0 && nEndPoint > 0 {
			return true
		}
		// 从大到小的方式, 找不到, 再找炸弹
		nOutPoint := pdk_HN_GetOutPointByCount(sCard, mHandCount, 0, 4)
		if nOutPoint > 0 {
			return true
		}
	case common.PDK_SHUANG_SHUN: // 双顺
		nStartPoint, nEndPoint := pdk_HN_GetCardByShunZi(sCard, mHandCount, nInPoint, nInLen, 2)
		if nStartPoint > 0 && nEndPoint > 0 {
			return true
		}
		// 从大到小的方式, 找不到, 再找炸弹
		nOutPoint := pdk_HN_GetOutPointByCount(sCard, mHandCount, 0, 4)
		if nOutPoint > 0 {
			return true
		}
	case common.PDK_SAN_SHUN: // 三顺
		nStartPoint, nEndPoint := pdk_HN_GetCardByShunZi(sCard, mHandCount, nInPoint, nInLen, 3)
		if nStartPoint > 0 && nEndPoint > 0 {
			return true
		}
		// 从大到小的方式, 找不到, 再找炸弹
		nOutPoint := pdk_HN_GetOutPointByCount(sCard, mHandCount, 0, 4)
		if nOutPoint > 0 {
			return true
		}
	case common.PDK_FEI_JI_2: // 2飞机带4张
		if nInLen == 10 {
			nInPoint = pdk_HN_GetInPointByCount(sInCard, 3)
			nStartPoint, nEndPoint := pdk_HN_GetCardByFeiJi(sCard, mHandCount, nInPoint, 6)
			if nStartPoint > 0 && nEndPoint > 0 && nCardLen >= 10 {
				return true
			}
			// 从大到小的方式, 找不到, 再找炸弹
			nOutPoint := pdk_HN_GetOutPointByCount(sCard, mHandCount, 0, 4)
			if nOutPoint > 0 {
				return true
			}
		}
	case common.PDK_FEI_JI_3: // 3飞机带6张
		if nInLen == 15 {
			nInPoint = pdk_HN_GetInPointByCount(sInCard, 3)
			nStartPoint, nEndPoint := pdk_HN_GetCardByFeiJi(sCard, mHandCount, nInPoint, 9)
			if nStartPoint > 0 && nEndPoint > 0 && nCardLen >= 15 {
				return true
			}
			// 从大到小的方式, 找不到, 再找炸弹
			nOutPoint := pdk_HN_GetOutPointByCount(sCard, mHandCount, 0, 4)
			if nOutPoint > 0 {
				return true
			}
		}
	case common.PDK_SI_DAI: // 四张带三张
		if nInLen == 7 {
			nInPoint = pdk_HN_GetInPointByCount(sInCard, 4)
			nOutPoint := pdk_HN_GetOutPointByCount(sCard, mHandCount, nInPoint, 4)
			if nOutPoint > 0 && nCardLen >= 7 {
				return true
			}
			// 从大到小的方式, 找不到, 再找炸弹
			nOutPoint = pdk_HN_GetOutPointByCount(sCard, mHandCount, 0, 4)
			if nOutPoint > 0 {
				return true
			}
		}
	case common.PDK_ZHA_DAN: // 炸弹
		nOutPoint := pdk_HN_GetOutPointByCount(sCard, mHandCount, nInPoint, 4)
		if nOutPoint > 0 {
			return true
		}
	}
	return false
}

func PaoDeKuai_HN_IsAllowOutCard(sCard, sOut []common.Card_info, is_sizhang_dai bool) (bool, uint8) {
	if sOut == nil {
		return false, 0
	}

	nLen := len(sCard)
	nOutLen := len(sOut)
	nOutCardType, nSubLen := PaoDeKuai_HN_CalcCardType(sOut)
	isYiBaShuai := nLen == nOutLen
	switch nOutCardType {
	case common.PDK_NIL:
		return false, nOutCardType.Value()
	case common.PDK_SAN_ZHANG, common.PDK_SAN_SHUN:
		if isYiBaShuai == false {
			return false, nOutCardType.Value()
		}
	case common.PDK_SAN_DAI:
		if isYiBaShuai == false && nSubLen != 2 {
			return false, nOutCardType.Value()
		}
	case common.PDK_SHUN_ZI:
		if nOutLen < 5 {
			return false, nOutCardType.Value()
		}
	case common.PDK_FEI_JI_2:
		if isYiBaShuai == false && nSubLen != 4 {
			return false, nOutCardType.Value()
		}
	case common.PDK_FEI_JI_3:
		if isYiBaShuai == false && nSubLen != 6 {
			return false, nOutCardType.Value()
		}
	case common.PDK_FEI_JI_4:
		if isYiBaShuai == false && nSubLen != 8 {
			return false, nOutCardType.Value()
		}
	case common.PDK_FEI_JI_5:
		if isYiBaShuai == false && nSubLen != 10 {
			return false, nOutCardType.Value()
		}
	case common.PDK_SI_DAI:
		if is_sizhang_dai == false {
			return false, nOutCardType.Value()
		}
		if isYiBaShuai == false && nSubLen != 3 {
			return false, nOutCardType.Value()
		}
	}
	return true, nOutCardType.Value()
}

func PaoDeKuai_HN_CalcCardValue(sCard []common.Card_info) uint32 {

	nValue := uint32(0)
	// 找组合
	for i := 0; i < len(pdk_hn_intelligent_out_card); i++ {
		sIn := pdk_hn_intelligent_out_card[i]
		nCardType := pdk_hn_intelligent_out_card_type[i]
		isHave := PaoDeKuai_HN_IsHaveBiggerCard(sCard, sIn)
		if isHave == true {
			nValue += uint32(nCardType)*100 + uint32(len(sIn))
		}
	}

	// 找单对子
	isHave := PaoDeKuai_HN_IsHaveBiggerCard(sCard, pdk_hn_intelligent_duizi_card)
	if isHave == true {
		nValue += uint32(common.PDK_DUI_ZI)*100 + uint32(len(pdk_hn_intelligent_duizi_card))
	}

	// 找单张
	for _, card := range sCard {
		if card[1] == 102 {
			nValue++
		}
	}
	return nValue
}

func PaoDeKuai_HN_IntelligentGetCard(sCard []common.Card_info, isBigToSmall, isSiZhangDai, isIntelligent bool) []common.Card_info {

	// 找组合
	for i := 0; i < len(pdk_hn_intelligent_out_card); i++ {
		sIn := pdk_hn_intelligent_out_card[i]
		sOut := PaoDeKuai_HN_GetBigCard(sCard, sIn, isBigToSmall, isIntelligent)
		isOut, _ := PaoDeKuai_HN_IsAllowOutCard(sCard, sOut, isSiZhangDai)
		if sOut != nil && isOut == true {
			return sOut
		}
	}

	// 找单对子
	sOut := PaoDeKuai_HN_GetBigCard(sCard, pdk_hn_intelligent_duizi_card, isBigToSmall, isIntelligent)
	if sOut != nil {
		return sOut
	}

	// 找单张
	sOut = PaoDeKuai_HN_GetBigCard(sCard, pdk_hn_intelligent_danzhang_card, isBigToSmall, isIntelligent)
	return sOut
}

func PaoDeKuai_HN_GetBigCard(sCard, sInCard []common.Card_info, isBigToSmall, isIntelligent bool) []common.Card_info {

	nCardLen := uint8(len(sCard))
	nInLen := uint8(len(sInCard))
	if nInLen < 1 {
		return nil
	}

	nInPoint := sInCard[0][1]
	mHandCount := Poker_StatPointCount(sCard)
	eInType, _ := PaoDeKuai_HN_CalcCardType(sInCard)
	switch eInType {
	case common.PDK_DAN_ZHANG: // 单张
		if isBigToSmall == true {
			if isIntelligent == true {
				// 智能出牌, 先找非对子的最大单张
				for i := 0; i < int(nCardLen); i++ {
					card := sCard[i]
					point := card[1]
					if point > nInPoint && mHandCount[point] == 1 {
						sRet := []common.Card_info{card}
						return sRet
					}
				}
				// 智能出牌, 都是非对子, 拆最大单张出
				for i := 0; i < int(nCardLen); i++ {
					card := sCard[i]
					point := card[1]
					if point > nInPoint {
						sRet := []common.Card_info{card}
						return sRet
					}
				}
			} else {
				// 下家报单, 最大单张出
				eEndCard := sCard[0]
				if eEndCard[1] > nInPoint {
					sRet := []common.Card_info{eEndCard}
					return sRet
				}
			}
		} else {
			// 智能出牌, 先找非对子的最小单张
			for i := int(nCardLen) - 1; i >= 0; i-- {
				card := sCard[i]
				point := card[1]
				if point > nInPoint && mHandCount[point] == 1 {
					sRet := []common.Card_info{card}
					return sRet
				}
			}
			// 智能出牌, 都是非对子, 拆最小单张出
			for i := int(nCardLen) - 1; i >= 0; i-- {
				card := sCard[i]
				point := card[1]
				if point > nInPoint {
					sRet := []common.Card_info{card}
					return sRet
				}
			}
		}
		// 找不到的情况下, 找炸弹
		if isIntelligent == false {
			sBombRet := pdk_HN_BuildCardByPoint(sCard, mHandCount, 0, 4)
			return sBombRet
		} else {
			return nil
		}
	case common.PDK_DUI_ZI: // 对子
		sRet := pdk_HN_BuildCardByPoint(sCard, mHandCount, nInPoint, 2)
		if isIntelligent == false && sRet == nil {
			// 找不到的情况下, 找炸弹
			sBombRet := pdk_HN_BuildCardByPoint(sCard, mHandCount, 0, 4)
			return sBombRet
		} else {
			return sRet
		}
	case common.PDK_SAN_ZHANG: // 三张
		sRet := pdk_HN_BuildCardByPoint(sCard, mHandCount, nInPoint, 3)
		if isIntelligent == false && sRet == nil {
			// 找不到的情况下, 找炸弹
			sBombRet := pdk_HN_BuildCardByPoint(sCard, mHandCount, 0, 4)
			return sBombRet
		} else {
			return sRet
		}
	case common.PDK_SAN_DAI: // 三张带二张
		if isIntelligent == false && nInLen < 5 {
			return nil
		}
		nInPoint = pdk_HN_GetInPointByCount(sInCard, 3)
		sRet := pdk_HN_BuildCardByPoint(sCard, mHandCount, nInPoint, 3)
		sSubRet := pdk_HN_BuildCardBySubCard(sCard, sRet, mHandCount, 2, isIntelligent)
		if sSubRet != nil {
			if isIntelligent == true || len(sSubRet) == 2 {
				sRet = append(sRet, sSubRet...)
				return sRet
			} else {
				sRet = nil
			}
		} else {
			sRet = nil
		}
		if isIntelligent == false && sRet == nil {
			// 找不到的情况下, 找炸弹
			sBombRet := pdk_HN_BuildCardByPoint(sCard, mHandCount, 0, 4)
			return sBombRet
		} else {
			return sRet
		}
	case common.PDK_SHUN_ZI: // 顺子
		sRet := pdk_HN_BuildCardByShunZi(sCard, mHandCount, nInPoint, nInLen, 1)
		if isIntelligent == false && sRet == nil {
			// 找不到的情况下, 找炸弹
			sBombRet := pdk_HN_BuildCardByPoint(sCard, mHandCount, 0, 4)
			return sBombRet
		} else {
			return sRet
		}
	case common.PDK_SHUANG_SHUN: // 双顺
		sRet := pdk_HN_BuildCardByShunZi(sCard, mHandCount, nInPoint, nInLen, 2)
		if isIntelligent == false && sRet == nil {
			// 找不到的情况下, 找炸弹
			sBombRet := pdk_HN_BuildCardByPoint(sCard, mHandCount, 0, 4)
			return sBombRet
		} else {
			return sRet
		}
	case common.PDK_SAN_SHUN: // 三顺
		sRet := pdk_HN_BuildCardByShunZi(sCard, mHandCount, nInPoint, nInLen, 3)
		if isIntelligent == false && sRet == nil {
			// 找不到的情况下, 找炸弹
			sBombRet := pdk_HN_BuildCardByPoint(sCard, mHandCount, 0, 4)
			return sBombRet
		} else {
			return sRet
		}
	case common.PDK_FEI_JI_2: // 2飞机带4张
		if isIntelligent == false && nInLen < 10 {
			return nil
		}
		nInPoint = pdk_HN_GetInPointByCount(sInCard, 3)
		sRet := pdk_HN_BuildCardByFeiJi(sCard, mHandCount, nInPoint, 6)
		sSubRet := pdk_HN_BuildCardBySubCard(sCard, sRet, mHandCount, 4, isIntelligent)
		if sSubRet != nil {
			if isIntelligent == true || len(sSubRet) == 4 {
				sRet = append(sRet, sSubRet...)
				return sRet
			} else {
				sRet = nil
			}
		} else {
			sRet = nil
		}
		if isIntelligent == false && sRet == nil {
			// 找不到的情况下, 找炸弹
			sBombRet := pdk_HN_BuildCardByPoint(sCard, mHandCount, 0, 4)
			return sBombRet
		} else {
			return sRet
		}
	case common.PDK_FEI_JI_3: // 3飞机带6张
		if isIntelligent == false && nInLen < 15 {
			return nil
		}
		nInPoint = pdk_HN_GetInPointByCount(sInCard, 3)
		sRet := pdk_HN_BuildCardByFeiJi(sCard, mHandCount, nInPoint, 9)
		sSubRet := pdk_HN_BuildCardBySubCard(sCard, sRet, mHandCount, 6, isIntelligent)
		if sSubRet != nil {
			if isIntelligent == true || len(sSubRet) == 6 {
				sRet = append(sRet, sSubRet...)
				return sRet
			} else {
				sRet = nil
			}
		} else {
			sRet = nil
		}
		if isIntelligent == false && sRet == nil {
			// 找不到的情况下, 找炸弹
			sBombRet := pdk_HN_BuildCardByPoint(sCard, mHandCount, 0, 4)
			return sBombRet
		} else {
			return sRet
		}
	case common.PDK_SI_DAI: // 四张带三张
		if isIntelligent == false && nInLen < 7 {
			return nil
		}
		nInPoint = pdk_HN_GetInPointByCount(sInCard, 4)
		sRet := pdk_HN_BuildCardBySiZhangDai(sCard, mHandCount, nInPoint, 3, isIntelligent)
		if isIntelligent == false && sRet == nil {
			// 找不到的情况下, 找炸弹
			sBombRet := pdk_HN_BuildCardByPoint(sCard, mHandCount, 0, 4)
			return sBombRet
		} else {
			return sRet
		}
	case common.PDK_ZHA_DAN: // 炸弹
		sRet := pdk_HN_BuildCardByPoint(sCard, mHandCount, nInPoint, 4)
		return sRet
	}
	return nil
}

// 双顺
func pdk_HN_IsShuangShun(sCard []common.Card_info) common.EPaoDeKuai {
	nLen := len(sCard)
	if nLen < 4 {
		return common.PDK_NIL // 无牌型
	}
	isRet := Poker_Is_AAKK(sCard)
	if isRet == true {
		return common.PDK_SHUANG_SHUN // 双顺
	}
	return common.PDK_NIL // 无牌型
}

// 顺子
func pdk_HN_IsShunZi(sCard []common.Card_info) common.EPaoDeKuai {
	nLen := len(sCard)
	if nLen < 5 {
		return common.PDK_NIL // 无牌型
	}
	isRet := Poker_Is_AKQJ10(sCard)
	if isRet == true {
		return common.PDK_SHUN_ZI // 顺子
	}
	return common.PDK_NIL // 无牌型
}

// 三顺
func pdk_HN_IsSanShun(sCard []common.Card_info) common.EPaoDeKuai {
	nLen := len(sCard)
	if nLen < 6 {
		return common.PDK_NIL // 无牌型
	}
	isSanShun := Poker_Is_AAAKKK(sCard)
	if isSanShun == true {
		return common.PDK_SAN_SHUN // 三顺
	}
	return common.PDK_NIL // 无牌型
}

// 迭代判断子飞机
func pdk_HN_IsSubFeiJi(sCard []common.Card_info, sPoint []uint8) (common.EPaoDeKuai, uint8) {
	nPointLen := uint8(len(sPoint))
	if nPointLen < 2 {
		return common.PDK_NIL, 0 // 无牌型
	}

	if sPoint[0]-1 == sPoint[1] {
		isShunZi := true
		for i := 1; i < int(nPointLen)-1; i++ {
			sNodeCurr := sPoint[i]
			sNodeNext := sPoint[i+1]
			// 不能算顺子
			if sNodeCurr-1 != sNodeNext {
				isShunZi = false
				break
			}
		}
		if isShunZi == true {
			nShunZiLen := nPointLen * 3
			nSubLen := uint8(len(sCard)) - nShunZiLen
			nMaxSubLen := nPointLen * 2
			if nSubLen > nMaxSubLen {
				// 超出可允许带副牌张数
				return common.PDK_NIL, 0 // 无牌型
			}
			switch nPointLen {
			case 2:
				return common.PDK_FEI_JI_2, nSubLen
			case 3:
				return common.PDK_FEI_JI_3, nSubLen
			case 4:
				return common.PDK_FEI_JI_4, nSubLen
			case 5:
				return common.PDK_FEI_JI_5, nSubLen
			default:
				return common.PDK_NIL, 0 // 无牌型
			}
		}
	}
	return common.PDK_NIL, 0 // 无牌型
}

// 飞机带
func pdk_HN_IsFeiJiDaiCB(sCard []common.Card_info) (common.EPaoDeKuai, uint8) {
	nLen := len(sCard)
	if nLen < 7 {
		return common.PDK_NIL, 0 // 无牌型
	}
	nShunZiLen := uint8(0)
	mCardPoint := Poker_StatPointCount(sCard)
	var sPoint []uint8
	for nPoint, nCount := range mCardPoint {
		if nCount >= 3 {
			sPoint = append(sPoint, nPoint)
			nShunZiLen += 3
		}
	}
	if sPoint != nil {
		nPointLen := uint8(len(sPoint))
		if nPointLen >= 2 {
			sort.Slice(sPoint, func(i, j int) bool {
				if sPoint[i] > sPoint[j] {
					return true
				}
				return false
			})

			for n := 0; n < int(nPointLen)-1; n++ {
				for i := nPointLen; i >= 2; i-- {
					if int(i)-n >= 2 {
						sSubPoint := sPoint[n:i]
						eType, nSubLen := pdk_HN_IsSubFeiJi(sCard, sSubPoint)
						if eType != common.PDK_NIL {
							return eType, nSubLen
						}
					}
				}
			}
		}
	}
	return common.PDK_NIL, 0 // 无牌型
}

// 根据输入牌张数, 获得牌点数; 作为输入点数使用;
func pdk_HN_GetInPointByCount(sInCard []common.Card_info, nPerCount int) uint8 {
	nOutPoint := uint8(0)
	mInCardCount := Poker_StatPointCount(sInCard)
	for point, count := range mInCardCount {
		if count >= nPerCount {
			if nOutPoint < point {
				nOutPoint = point
			}
		}
	}
	return nOutPoint
}

// 根据输入牌点数和要求张数, 获得输出点数
func pdk_HN_GetOutPointByCount(sCard []common.Card_info, mHandCount map[uint8]int, nInPoint uint8, nPerCount int) uint8 {

	// 先找完全匹配的最小
	nOutPoint := uint8(255)
	for point, count := range mHandCount {
		if count == nPerCount && point > nInPoint {
			if nOutPoint > point {
				nOutPoint = point
			}
		}
	}
	if nOutPoint == 255 {
		// 如果找不到完全匹配的, 找兼容的
		for point, count := range mHandCount {
			if count >= nPerCount && point > nInPoint {
				if nOutPoint > point {
					nOutPoint = point
				}
			}
		}
		if nOutPoint == 255 {
			nOutPoint = 0
		}
	}
	return nOutPoint
}

// 找出拥有大于指定点数指定张数的牌
func pdk_HN_BuildCardByPoint(sCard []common.Card_info, mHandCount map[uint8]int, nInPoint uint8, nPerCount int) []common.Card_info {
	nOutPoint := pdk_HN_GetOutPointByCount(sCard, mHandCount, nInPoint, nPerCount)
	if nOutPoint <= 0 {
		return nil
	}
	nAddCount := nPerCount
	var sRet []common.Card_info
	nCardLen := len(sCard)
	for i := 0; i < nCardLen; i++ {
		card := sCard[i]
		if card[1] == nOutPoint {
			sRet = append(sRet, card)
			nAddCount--
			if nAddCount <= 0 {
				return sRet
			}
		}
	}
	return nil
}

// 连续添加多张副带的牌(可是单张, 可是对子等)
func pdk_HN_BuildCardBySubCard(sCard, sExclude []common.Card_info, mHandCount map[uint8]int, nSubCount uint8, isIntelligent bool) []common.Card_info {
	if sExclude == nil {
		return nil
	}
	mExclude := make(map[uint32]bool)
	for _, card := range sExclude {
		nKey := uint32(card[0])*100 + uint32(card[1])
		mExclude[nKey] = true
	}
	nCardLen := len(sCard)
	nTotalLen := len(sExclude)
	nAddCount := nSubCount
	var sRet []common.Card_info
	for i := nCardLen - 1; i >= 0; i-- {
		card := sCard[i]
		point := card[1]
		nKey := uint32(card[0])*100 + uint32(point)
		if _, isExist := mExclude[nKey]; isExist == false {
			if mHandCount[point] == 1 && point < 102 {
				// 排除掉添加的牌
				mExclude[nKey] = true

				// 先添加单张
				sRet = append(sRet, card)
				nAddCount--
				nTotalLen++
				if isIntelligent == true {
					if nAddCount <= 0 || nTotalLen >= nCardLen {
						return sRet
					}
				} else {
					if nAddCount <= 0 {
						return sRet
					}
				}
			}
		}
	}

	for i := nCardLen - 1; i >= 0; i-- {
		card := sCard[i]
		point := card[1]
		nKey := uint32(card[0])*100 + uint32(point)
		if _, isExist := mExclude[nKey]; isExist == false {
			// 排除掉添加的牌
			mExclude[nKey] = true

			// 如果单张不够, 拆对子等
			sRet = append(sRet, card)
			nAddCount--
			nTotalLen++
			if isIntelligent == true {
				if nAddCount <= 0 || nTotalLen >= nCardLen {
					return sRet
				}
			} else {
				if nAddCount <= 0 {
					return sRet
				}
			}
		}
	}
	return nil
}

// 查找顺子牌型, 每张牌型添加指定张数
// 可查找类型: 11, 22, 33, 44
// 可查找类型: 1, 2, 3, 4, 5
func pdk_HN_GetCardByShunZi(sCard []common.Card_info, mHandCount map[uint8]int, nStart, nInLen uint8, nPerCount int) (uint8, uint8) {
	nStartPoint := nStart + 1
	nEndPoint := nStartPoint + 1 - nInLen/uint8(nPerCount)
	if nStartPoint > 14 {
		return 0, 0
	}
	nCheck := uint8(0)
	for nStartPoint <= 14 {
		nCheck = 0
		for point := nStartPoint; point >= nEndPoint; point-- {
			if mHandCount[point] < nPerCount {
				break
			} else {
				nCheck++
			}
		}
		if nCheck*uint8(nPerCount) >= nInLen {
			break
		}
		nStartPoint++
		nEndPoint++
	}
	if nCheck*uint8(nPerCount) < nInLen {
		return 0, 0
	}
	return nStartPoint, nEndPoint
}

// 查找飞机牌型, 每张牌型添加指定张数
// 可查找类型: 111, 222, 333, 444
func pdk_HN_GetCardByFeiJi(sCard []common.Card_info, mHandCount map[uint8]int, nStart, nInLen uint8) (uint8, uint8) {
	nStartPoint, nEndPoint := pdk_HN_GetCardByShunZi(sCard, mHandCount, nStart, nInLen, 3)
	return nStartPoint, nEndPoint
}

// 查找顺子牌型, 每张牌型添加指定张数
// 可查找类型: 11, 22, 33, 44
// 可查找类型: 1, 2, 3, 4, 5
func pdk_HN_BuildCardByShunZi(sCard []common.Card_info, mHandCount map[uint8]int, nStart, nInLen uint8, nPerCount int) []common.Card_info {
	nStartPoint, nEndPoint := pdk_HN_GetCardByShunZi(sCard, mHandCount, nStart, nInLen, nPerCount)
	if nStartPoint <= 0 || nEndPoint <= 0 {
		return nil
	}

	nCardLen := uint8(len(sCard))
	var sRet []common.Card_info
	mOutCount := make(map[uint8]int)
	for i := uint8(0); i < nCardLen; i++ {
		if sRet != nil && len(sRet) >= int(nInLen) {
			break
		}
		card := sCard[i]
		point := card[1]
		if point >= nEndPoint && point <= nStartPoint {
			if count, isExist := mOutCount[point]; isExist == false {
				mOutCount[point] = 1
				sRet = append(sRet, card)
			} else {
				if count < nPerCount {
					sRet = append(sRet, card)
				}
				mOutCount[point]++
			}
		}
	}
	return sRet
}

// 查找顺子牌型, 每张牌型添加指定张数
// 可查找类型: 3333, 444;
// 可查找类型: 5555, 467;
func pdk_HN_BuildCardBySiZhangDai(sCard []common.Card_info, mHandCount map[uint8]int, nInPoint, nSubLen uint8, isIntelligent bool) []common.Card_info {
	sRet := pdk_HN_BuildCardByPoint(sCard, mHandCount, nInPoint, 4)
	sSubRet := pdk_HN_BuildCardBySubCard(sCard, sRet, mHandCount, nSubLen, isIntelligent)
	if sSubRet != nil && len(sSubRet) >= int(nSubLen) {
		sRet = append(sRet, sSubRet...)
		return sRet
	}
	return nil
}

// 查找飞机牌型, 每张牌型添加指定张数
// 可查找类型: 111, 222, 333, 444
func pdk_HN_BuildCardByFeiJi(sCard []common.Card_info, mHandCount map[uint8]int, nStart, nInLen uint8) []common.Card_info {
	sRet := pdk_HN_BuildCardByShunZi(sCard, mHandCount, nStart, nInLen, 3)
	return sRet
}
