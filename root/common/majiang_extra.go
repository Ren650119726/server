package common

// 麻将额外加翻类型
type EMaJiangExtra uint8

const (
	EXTRA_NIL          EMaJiangExtra = 0  // 无
	EXTRA_DUAN_YAO_JIU EMaJiangExtra = 1  // 断幺九
	EXTRA_456_KA       EMaJiangExtra = 2  // 456卡
	EXTRA_GOU          EMaJiangExtra = 3  // 勾
	EXTRA_GANGSHANGHUA EMaJiangExtra = 4  // 杠上花
	EXTRA_GANGSHANGPAO EMaJiangExtra = 5  // 杠上炮
	EXTRA_QIANGGANGHU  EMaJiangExtra = 6  // 抢杠胡
	EXTRA_HAIDIPAO     EMaJiangExtra = 7  // 海底炮
	EXTRA_HAIDIHUA     EMaJiangExtra = 8  // 海底花
	EXTRA_BAOHU        EMaJiangExtra = 9  // 报胡
	EXTRA_ZHUABAOHU    EMaJiangExtra = 10 // 抓报胡
	EXTRA_ZHUAQINGHU   EMaJiangExtra = 11 // 抓请胡
	EXTRA_ZIMO         EMaJiangExtra = 12 // 自摸
	EXTRA_MENQ         EMaJiangExtra = 13 // 门清
	EXTRA_JIAXINWU     EMaJiangExtra = 14 // 夹心五
)

var strMaJiangExtraType = map[EMaJiangExtra]string{
	EXTRA_NIL:          "无",
	EXTRA_DUAN_YAO_JIU: "断幺九",
	EXTRA_456_KA:       "456卡",
	EXTRA_GOU:          "勾",
	EXTRA_GANGSHANGHUA: "杠上花",
	EXTRA_GANGSHANGPAO: "杠上炮",
	EXTRA_QIANGGANGHU:  "抢杠胡",
	EXTRA_HAIDIPAO:     "海底炮",
	EXTRA_HAIDIHUA:     "海底花",
	EXTRA_BAOHU:        "报胡",
	EXTRA_ZHUABAOHU:    "抓报胡",
	EXTRA_ZHUAQINGHU:   "抓请胡",
	EXTRA_ZIMO:         "自摸",
	EXTRA_MENQ:         "门清",
	EXTRA_JIAXINWU:     "夹心五",
}

func (e EMaJiangExtra) String() string {
	return strMaJiangExtraType[e]
}

func (e EMaJiangExtra) Value() uint8 {
	return uint8(e)
}
