package common

type EGameType uint8

// 服务器类别定义
const (
	EGameTypeNIL                EGameType = 0   // 默认值
	EGameTypeNIU_NIU            EGameType = 1   // 牛牛
	EGameTypeJIN_HUA            EGameType = 2   // 金花
	EGameTypeCHE_XUAN           EGameType = 3   // 扯旋
	EGameTypeWU_ZI_QI           EGameType = 4   // 五子棋
	EGameTypeSHEN_SHOU_ZHI_ZHAN EGameType = 5   // 神兽之战
	EGameTypeTUI_TONG_ZI        EGameType = 6   // 推筒子
	EGameTypeSHI_SAN_SHUI       EGameType = 7   // 十三水
	EGameTypeLONG_HU_DOU        EGameType = 8   // 龙虎斗
	EGameTypeHONG_HEI_DA_ZHAN   EGameType = 9   // 红黑大战
	EGameTypeHONG_BAO           EGameType = 11  // 红包接龙
	EGameTypeDGK                EGameType = 12  // 断勾卡
	EGameTypePAO_DE_KUAI        EGameType = 13  // 跑得快
	EGameTypeXMMJ               EGameType = 14  // 熊猫麻将
	EGameTypeTEN_NIU_NIU        EGameType = 15  // 十人牛牛
	EGameTypeFQZS               EGameType = 16  // 飞禽走兽
	EGameTypePDK_HN             EGameType = 17  // 跑得快_湖南版本
	EGameTypeSAN_GONG           EGameType = 18  // 三公
	EGameTypeWUHUA_NIUNIU       EGameType = 101 // 无花牛牛 (无花比比)
	EGameTypeDING_ER_HONG       EGameType = 103 // 丁二红
)

var GameTypeByID = map[EGameType]string{
	EGameTypeNIL:                "大厅",
	EGameTypeNIU_NIU:            "牛牛",
	EGameTypeJIN_HUA:            "金花",
	EGameTypeCHE_XUAN:           "扯旋",
	EGameTypeWU_ZI_QI:           "五子棋",
	EGameTypeSHEN_SHOU_ZHI_ZHAN: "神兽之战",
	EGameTypeTUI_TONG_ZI:        "推筒子",
	EGameTypeSHI_SAN_SHUI:       "十三水",
	EGameTypeLONG_HU_DOU:        "龙虎斗",
	EGameTypeHONG_HEI_DA_ZHAN:   "红黑大战",
	EGameTypeHONG_BAO:           "全民抢红包",
	EGameTypeDGK:                "断勾卡",
	EGameTypePAO_DE_KUAI:        "跑得快",
	EGameTypeXMMJ:               "熊猫麻将",
	EGameTypeTEN_NIU_NIU:        "十人牛牛",
	EGameTypeFQZS:               "飞禽走兽",
	EGameTypePDK_HN:             "跑得快_湖南",
	EGameTypeSAN_GONG:           "三公",
	EGameTypeWUHUA_NIUNIU:       "无花牛牛",
	EGameTypeDING_ER_HONG:       "丁二红",
}

var GameTypeByString = map[string]EGameType{
	"大厅":    EGameTypeNIL,
	"牛牛":    EGameTypeNIU_NIU,
	"金花":    EGameTypeJIN_HUA,
	"扯旋":    EGameTypeCHE_XUAN,
	"五子棋":   EGameTypeWU_ZI_QI,
	"神兽之战":  EGameTypeSHEN_SHOU_ZHI_ZHAN,
	"推筒子":   EGameTypeTUI_TONG_ZI,
	"十三水":   EGameTypeSHI_SAN_SHUI,
	"龙虎斗":   EGameTypeLONG_HU_DOU,
	"红黑大战":  EGameTypeHONG_HEI_DA_ZHAN,
	"红包":    EGameTypeHONG_BAO,
	"断勾卡":   EGameTypeDGK,
	"跑得快":   EGameTypePAO_DE_KUAI,
	"熊猫麻将":  EGameTypeXMMJ,
	"十人牛牛":  EGameTypeTEN_NIU_NIU,
	"飞禽走兽":  EGameTypeFQZS,
	"跑得快湖南": EGameTypePDK_HN,
	"三公":    EGameTypeSAN_GONG,
	"无花牛牛":  EGameTypeWUHUA_NIUNIU,
	"丁二红":   EGameTypeDING_ER_HONG,
}

func (e EGameType) String() string {
	return GameTypeByID[e]
}

func (e EGameType) Value() uint8 {
	return uint8(e)
}
