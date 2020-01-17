package common

type EGameType uint8

// 服务器类别定义
const (
	EGameTypeNIL                 EGameType = 0   // 默认值
	EGameTypeCATCHFISH           EGameType = 1   // 捕鱼
	EGameTypeFRUITMARY           EGameType = 2   // 水果小玛丽

)

var GameTypeByID = map[EGameType]string{
	EGameTypeNIL:                "大厅",
	EGameTypeCATCHFISH:          "捕鱼",
	EGameTypeFRUITMARY:          "水果小玛丽",
}

var GameTypeByString = map[string]EGameType{
	"大厅":    		EGameTypeNIL,
	"捕鱼":    		EGameTypeCATCHFISH,
	"水果小玛丽":    EGameTypeFRUITMARY,
}

func (e EGameType) String() string {
	return GameTypeByID[e]
}

func (e EGameType) Value() uint8 {
	return uint8(e)
}
