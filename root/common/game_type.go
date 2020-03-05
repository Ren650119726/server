package common

type EGameType uint8

// 服务器类别定义
const (
	EGameTypeNIL       EGameType = 0 // 默认值
	EGameTypeCATCHFISH EGameType = 1 // 捕鱼
	EGameTypeFRUITMARY EGameType = 2 // 水果小玛丽
	EGameTypeDFDC      EGameType = 3 // 多福多财
	EGameTypeJPM       EGameType = 4 // 金瓶梅
	EGameTypeLUCKFRUIT EGameType = 5 // 幸运水果机
	EGameTypeRED2BLACK EGameType = 6 // 红黑大战
	EGameTypeLHD       EGameType = 7 // 龙虎斗

)

var GameTypeByID = map[EGameType]string{
	EGameTypeNIL:       "大厅",
	EGameTypeCATCHFISH: "捕鱼",
	EGameTypeFRUITMARY: "水果小玛丽",
	EGameTypeDFDC:      "多福多财",
	EGameTypeJPM:       "金瓶梅",
	EGameTypeLUCKFRUIT: "幸运水果机",
	EGameTypeRED2BLACK: "红黑大战",
	EGameTypeLHD:       "龙虎斗",
}

var GameTypeByString = map[string]EGameType{
	"大厅":    EGameTypeNIL,
	"捕鱼":    EGameTypeCATCHFISH,
	"水果小玛丽": EGameTypeFRUITMARY,
	"多福多财":  EGameTypeDFDC,
	"金瓶梅":   EGameTypeJPM,
	"幸运水果机": EGameTypeLUCKFRUIT,
	"红黑大战":  EGameTypeRED2BLACK,
	"龙虎斗":   EGameTypeLHD,
}

func (e EGameType) String() string {
	return GameTypeByID[e]
}

func (e EGameType) Value() uint8 {
	return uint8(e)
}
