package common

// 麻将胡牌类型
type EPaoDeKuai uint8

const (
	PDK_NIL         EPaoDeKuai = 0  // 无牌型
	PDK_DAN_ZHANG   EPaoDeKuai = 1  // 单张
	PDK_DUI_ZI      EPaoDeKuai = 2  // 对子
	PDK_SAN_ZHANG   EPaoDeKuai = 3  // 三张
	PDK_SAN_DAI     EPaoDeKuai = 4  // 三张带副牌 标准(555, 66) 一把甩(555, 6)
	PDK_SHUN_ZI     EPaoDeKuai = 5  // 顺子 (5, 6, 7, 8, 9)
	PDK_SHUANG_SHUN EPaoDeKuai = 6  // 双顺 (55, 66) (JJ, QQ)
	PDK_SAN_SHUN    EPaoDeKuai = 7  // 三顺 (555,666) (JJJ, QQQ)
	PDK_FEI_JI_2    EPaoDeKuai = 8  // 飞机带翅膀 (2飞机) 标准(555,666, 77, 88) 一把甩(777, 888, K)
	PDK_FEI_JI_3    EPaoDeKuai = 9  // 飞机带翅膀 (3飞机)
	PDK_FEI_JI_4    EPaoDeKuai = 10 // 飞机带翅膀 (4飞机)
	PDK_FEI_JI_5    EPaoDeKuai = 11 // 飞机带翅膀 (5飞机)
	PDK_SI_DAI      EPaoDeKuai = 12 // 四张带副牌  标准(5555, 666) 一把甩(5555, 6) or (5555, 66)
	PDK_ZHA_DAN     EPaoDeKuai = 13 // 炸弹 (5555) (KKKK)
)

var strPaoDeKuaiType = map[EPaoDeKuai]string{
	PDK_NIL:         "无牌型",
	PDK_DAN_ZHANG:   "单张",
	PDK_DUI_ZI:      "对子",
	PDK_SAN_ZHANG:   "三张",
	PDK_SAN_DAI:     "三张带",
	PDK_SHUN_ZI:     "顺子",
	PDK_SHUANG_SHUN: "双顺",
	PDK_SAN_SHUN:    "三顺",
	PDK_FEI_JI_2:    "二飞机带",
	PDK_FEI_JI_3:    "三飞机带",
	PDK_FEI_JI_4:    "四飞机带",
	PDK_FEI_JI_5:    "五飞机带",
	PDK_SI_DAI:      "四张带",
	PDK_ZHA_DAN:     "炸弹",
}

func (e EPaoDeKuai) String() string {
	return strPaoDeKuaiType[e]
}

func (e EPaoDeKuai) Value() uint8 {
	return uint8(e)
}
