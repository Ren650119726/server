package common

// 麻将胡牌类型
type EMaJiangHu uint8

const (
	HU_NIL                 EMaJiangHu = 0  // 未胡牌
	HU_PING_HU             EMaJiangHu = 1  // 平胡
	HU_DUI_DUI_HU          EMaJiangHu = 2  // 对对胡
	HU_QING_YI_SE          EMaJiangHu = 3  // 清一色
	HU_WU_DUI              EMaJiangHu = 4  // 五对
	HU_QING_DUI_DUI        EMaJiangHu = 5  // 清对对
	HU_QING_WU_DUI         EMaJiangHu = 6  // 清五对
	HU_DAI_YAO_JIU         EMaJiangHu = 7  // 带幺九
	HU_QING_YAO_JIU        EMaJiangHu = 8  // 清幺九
	HU_JIANG_DUI_DUI       EMaJiangHu = 9  // 将对对
	HU_TIAN                EMaJiangHu = 10 // 天胡
	HU_DI                  EMaJiangHu = 11 // 地胡
	HU_QI_DUI              EMaJiangHu = 12 // 七对
	HU_QING_QI_DUI         EMaJiangHu = 13 // 清七对
	HU_JIN_GOU_DIAO        EMaJiangHu = 14 // 金钩钓
	HU_LONG_QI_DUI         EMaJiangHu = 15 // 龙七对
	HU_JIANG_JIN_GOU_DIAO  EMaJiangHu = 16 // 将金钩钓
	HU_QING_JIN_GOU_DIAO   EMaJiangHu = 17 // 清金钩钓
	HU_QING_LONG_QI_DUI    EMaJiangHu = 18 // 清龙七对
	HU_SHI_BA_LUO_HAN      EMaJiangHu = 19 // 十八罗汉
	HU_QING_SHI_BA_LUO_HAN EMaJiangHu = 20 // 清十八罗汉
)

var strMaJiangHuType = map[EMaJiangHu]string{
	HU_NIL:                 "未胡牌",
	HU_PING_HU:             "平胡",
	HU_DUI_DUI_HU:          "对对胡",
	HU_QING_YI_SE:          "清一色",
	HU_WU_DUI:              "五对",
	HU_QING_DUI_DUI:        "清对对",
	HU_QING_WU_DUI:         "清五对",
	HU_DAI_YAO_JIU:         "带幺九",
	HU_QING_YAO_JIU:        "清幺九",
	HU_JIANG_DUI_DUI:       "将对对",
	HU_TIAN:                "天胡",
	HU_DI:                  "地胡",
	HU_QI_DUI:              "七对",
	HU_QING_QI_DUI:         "清七对",
	HU_JIN_GOU_DIAO:        "金钩钓",
	HU_LONG_QI_DUI:         "龙七对",
	HU_JIANG_JIN_GOU_DIAO:  "将金钩钓",
	HU_QING_JIN_GOU_DIAO:   "清金钩钓",
	HU_QING_LONG_QI_DUI:    "清龙七对",
	HU_SHI_BA_LUO_HAN:      "十八罗汉",
	HU_QING_SHI_BA_LUO_HAN: "清十八罗汉",
}

func (e EMaJiangHu) String() string {
	return strMaJiangHuType[e]
}

func (e EMaJiangHu) Value() uint8 {
	return uint8(e)
}
