package types

type EGameStatus byte

// 服务器类别定义
const (
	EGameStatus_SITDOWN EGameStatus = 1 // 坐下
	EGameStatus_READY   EGameStatus = 2 // 准备好
)

var typeStringgame = [...]string{
	EGameStatus_SITDOWN: "sitdown",
	EGameStatus_READY:   "ready",
}

func (e EGameStatus) String() string {
	return typeStringgame[e]
}

func (e EGameStatus) UInt8() uint8 {
	return uint8(e)
}
