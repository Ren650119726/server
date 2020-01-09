package types

type EGameStatus byte

// 服务器类别定义
const (
	EGameStatus_SITDOWN EGameStatus = 1 // 坐下
	EGameStatus_JOIN    EGameStatus = 2 // 加入游戏
	EGameStatus_PREPARE EGameStatus = 3 // 设置簸簸完成
	EGameStatus_PLAYING EGameStatus = 4 // 游戏中

	EGameStatus_GIVE_UP EGameStatus = 5 // 放弃耍游戏
)

var typeStringgame = [...]string{
	EGameStatus_SITDOWN: "sitdown",
	EGameStatus_JOIN:    "join",
	EGameStatus_PREPARE: "prepare",
	EGameStatus_PLAYING: "playing",
	EGameStatus_GIVE_UP: "give up",
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
