package common

type ENiuNiuType byte

// 服务器类别定义
const (
	NN_NIU_0          ENiuNiuType = 0  // 无牛
	NN_NIU_1          ENiuNiuType = 1  // 牛一
	NN_JP_NIU_1       ENiuNiuType = 2  // 金牌牛一
	NN_NIU_2          ENiuNiuType = 3  // 牛二
	NN_JP_NIU_2       ENiuNiuType = 4  // 金牌牛二
	NN_NIU_3          ENiuNiuType = 5  // 牛三
	NN_JP_NIU_3       ENiuNiuType = 6  // 金牌牛三
	NN_NIU_4          ENiuNiuType = 7  // 牛四
	NN_JP_NIU_4       ENiuNiuType = 8  // 金牌牛四
	NN_NIU_5          ENiuNiuType = 9  // 牛五
	NN_JP_NIU_5       ENiuNiuType = 10 // 金牌牛五
	NN_NIU_6          ENiuNiuType = 11 // 牛六
	NN_JP_NIU_6       ENiuNiuType = 12 // 金牌牛六
	NN_NIU_7          ENiuNiuType = 13 // 牛七
	NN_JP_NIU_7       ENiuNiuType = 14 // 金牌牛七
	NN_NIU_8          ENiuNiuType = 15 // 牛八
	NN_JP_NIU_8       ENiuNiuType = 16 // 金牌牛八
	NN_NIU_9          ENiuNiuType = 17 // 牛九
	NN_JP_NIU_9       ENiuNiuType = 18 // 金牌牛九
	NN_NIU_10         ENiuNiuType = 19 // 牛牛
	NN_JP_NIU_10      ENiuNiuType = 20 // 金牌牛牛
	NN_SHUNZI_11      ENiuNiuType = 21 // 顺子
	NN_WUHUANIU_12    ENiuNiuType = 22 // 五花牛
	NN_TONGHUA_13     ENiuNiuType = 23 // 同花
	NN_HULU_14        ENiuNiuType = 24 // 葫芦
	NN_ZHADAN_15      ENiuNiuType = 25 // 炸弹
	NN_WUXIAONIU_16   ENiuNiuType = 26 // 五小牛
	NN_SISHI_17       ENiuNiuType = 27 // 四十
	NN_TONGHUASHUN_18 ENiuNiuType = 28 // 同花顺
)

var typeStringify_niuniu = [...]string{
	NN_NIU_0:          "无牛",
	NN_NIU_1:          "牛一",
	NN_JP_NIU_1:       "金牌牛一",
	NN_NIU_2:          "牛二",
	NN_JP_NIU_2:       "金牌牛二",
	NN_NIU_3:          "牛三",
	NN_JP_NIU_3:       "金牌牛三",
	NN_NIU_4:          "牛四",
	NN_JP_NIU_4:       "金牌牛四",
	NN_NIU_5:          "牛五",
	NN_JP_NIU_5:       "金牌牛五",
	NN_NIU_6:          "牛六",
	NN_JP_NIU_6:       "金牌牛六",
	NN_NIU_7:          "牛七",
	NN_JP_NIU_7:       "金牌牛七",
	NN_NIU_8:          "牛八",
	NN_JP_NIU_8:       "金牌牛八",
	NN_NIU_9:          "牛九",
	NN_JP_NIU_9:       "金牌牛九",
	NN_NIU_10:         "牛牛",
	NN_JP_NIU_10:      "金牌牛牛",
	NN_SHUNZI_11:      "顺子",
	NN_WUHUANIU_12:    "五花牛",
	NN_TONGHUA_13:     "同花",
	NN_HULU_14:        "葫芦",
	NN_ZHADAN_15:      "炸弹",
	NN_WUXIAONIU_16:   "五小牛",
	NN_SISHI_17:       "四十",
	NN_TONGHUASHUN_18: "同花顺",
}

func (e ENiuNiuType) String() string {
	return typeStringify_niuniu[e]
}

func (e ENiuNiuType) UInt8() uint8 {
	return uint8(e)
}
