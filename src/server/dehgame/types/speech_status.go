package types

type ESpeechStatus byte

// 服务器类别定义
const (
	NIL  ESpeechStatus = 0
	XIU  ESpeechStatus = 1
	DIU  ESpeechStatus = 2
	DA   ESpeechStatus = 3
	QIAO ESpeechStatus = 4
	GEN  ESpeechStatus = 5
)

var typeString_speech = [...]string{
	XIU:  "休",
	DIU:  "丢",
	DA:   "大",
	QIAO: "敲",
	GEN:  "跟",
}

func (e ESpeechStatus) String() string {
	return typeString_speech[e]
}

func (e ESpeechStatus) Int32() int32 {
	return int32(e)
}

func (e ESpeechStatus) UInt8() uint8 {
	return uint8(e)
}
