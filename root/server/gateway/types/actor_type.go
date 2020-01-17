package types

type EActorType byte

// 服务器类别定义
const (
	EActorType_Unknown EActorType = 0
)

var typeStringify = [...]string{
	EActorType_Unknown: "Unknown",
}

func (e EActorType) String() string {
	return typeStringify[e]
}

func (e EActorType) Int32() int32 {
	return int32(e)
}
