package types

type EGameStatus byte

// 服务器类别定义
const (
	EGameStatus_SITDOWN EGameStatus = 0 // 坐下
	EGameStatus_READY   EGameStatus = 1 // 准备好
	EGameStatus_PLAYING EGameStatus = 2 // 游戏中
)

var typeStringgame = [...]string{
	EGameStatus_SITDOWN: "sitdown",
	EGameStatus_READY:   "ready",
	EGameStatus_PLAYING: "playing",
}

func (e EGameStatus) String() string {
	return typeStringgame[e]
}

func (e EGameStatus) Int32() int32 {
	return int32(e)
}

func (e EGameStatus) UInt8() uint8 {
	return uint8(e)
}
