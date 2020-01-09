package types

import "root/common"

type EActorType byte

// 服务器类别定义
const (
	EActorType_Unknown common.EActorType = 0
	EActorType_PROCESS common.EActorType = 1 // 主逻辑actor
)

var typeStringify = [...]string{
	EActorType_Unknown: "Unknown",
	EActorType_PROCESS: "Process",
}

func (e EActorType) String() string {
	return typeStringify[e]
}

func (e EActorType) Int32() int32 {
	return int32(e)
}
