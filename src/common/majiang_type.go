package common

// 麻将牌的枚举
type EMaJiangType uint8

const (
	TONG_1 EMaJiangType = 11 // 1筒
	TONG_2 EMaJiangType = 12 // 2筒
	TONG_3 EMaJiangType = 13 // 3筒
	TONG_4 EMaJiangType = 14 // 4筒
	TONG_5 EMaJiangType = 15 // 5筒
	TONG_6 EMaJiangType = 16 // 6筒
	TONG_7 EMaJiangType = 17 // 7筒
	TONG_8 EMaJiangType = 18 // 8筒
	TONG_9 EMaJiangType = 19 // 9筒

	TIAO_1 EMaJiangType = 21 // 1条
	TIAO_2 EMaJiangType = 22 // 2条
	TIAO_3 EMaJiangType = 23 // 3条
	TIAO_4 EMaJiangType = 24 // 4条
	TIAO_5 EMaJiangType = 25 // 5条
	TIAO_6 EMaJiangType = 26 // 6条
	TIAO_7 EMaJiangType = 27 // 7条
	TIAO_8 EMaJiangType = 28 // 8条
	TIAO_9 EMaJiangType = 29 // 9条

	WAN_1 EMaJiangType = 31 // 1萬
	WAN_2 EMaJiangType = 32 // 2萬
	WAN_3 EMaJiangType = 33 // 3萬
	WAN_4 EMaJiangType = 34 // 4萬
	WAN_5 EMaJiangType = 35 // 5萬
	WAN_6 EMaJiangType = 36 // 6萬
	WAN_7 EMaJiangType = 37 // 7萬
	WAN_8 EMaJiangType = 38 // 8萬
	WAN_9 EMaJiangType = 39 // 9萬

	BAI_BAN    EMaJiangType = 50 // 白板
	HONG_ZHONG EMaJiangType = 52 // 红中
	FA_CAI     EMaJiangType = 54 // 發财

	DONG EMaJiangType = 61 // 东
	NAN  EMaJiangType = 63 // 南
	XI   EMaJiangType = 65 // 西
	BAI  EMaJiangType = 67 // 北

	CUN_TIAN  EMaJiangType = 71 // 春
	XIA_TIAN  EMaJiangType = 73 // 夏
	QIU_TIAN  EMaJiangType = 75 // 秋
	DONG_TIAN EMaJiangType = 77 // 东

	MEI EMaJiangType = 81 // 梅
	LAN EMaJiangType = 83 // 兰
	ZHU EMaJiangType = 85 // 竹
	JU  EMaJiangType = 87 // 菊
)

var strMaJiangType = map[EMaJiangType]string{
	TONG_1: "一筒",
	TONG_2: "二筒",
	TONG_3: "三筒",
	TONG_4: "四筒",
	TONG_5: "五筒",
	TONG_6: "六筒",
	TONG_7: "七筒",
	TONG_8: "八筒",
	TONG_9: "九筒",

	TIAO_1: "一条",
	TIAO_2: "二条",
	TIAO_3: "三条",
	TIAO_4: "四条",
	TIAO_5: "五条",
	TIAO_6: "六条",
	TIAO_7: "七条",
	TIAO_8: "八条",
	TIAO_9: "九条",

	WAN_1: "一萬",
	WAN_2: "二萬",
	WAN_3: "三萬",
	WAN_4: "四萬",
	WAN_5: "五萬",
	WAN_6: "六萬",
	WAN_7: "七萬",
	WAN_8: "八萬",
	WAN_9: "九萬",

	BAI_BAN:    "白板",
	HONG_ZHONG: "红中",
	FA_CAI:     "發财",

	DONG: "东",
	NAN:  "南",
	XI:   "西",
	BAI:  "北",

	CUN_TIAN:  "春",
	XIA_TIAN:  "夏",
	QIU_TIAN:  "秋",
	DONG_TIAN: "东",

	MEI: "梅",
	LAN: "兰",
	ZHU: "竹",
	JU:  "菊",
}

func (e EMaJiangType) String() string {
	return strMaJiangType[e]
}

func (e EMaJiangType) Value() uint8 {
	return uint8(e)
}
