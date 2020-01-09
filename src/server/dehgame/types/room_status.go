package types

type ERoomStatus byte

// 服务器类别定义
const (
	ERoomStatus_WAITING     ERoomStatus = 1 // 等待开始
	ERoomStatus_SETBOBO     ERoomStatus = 2 // 设置簸簸
	ERoomStatus_PLAYING     ERoomStatus = 3 // 开始游戏
	ERoomStatus_ARRANGEMENT ERoomStatus = 4 // 分牌
	ERoomStatus_SETTLEMENT  ERoomStatus = 5 // 结算

	ERoomStatus_CLOSE ERoomStatus = 6 // 关闭

)

var typeStringify = [...]string{
	ERoomStatus_WAITING:     "wating",
	ERoomStatus_SETBOBO:     "setbobo",
	ERoomStatus_PLAYING:     "playing",
	ERoomStatus_ARRANGEMENT: "arrangement",
	ERoomStatus_SETTLEMENT:  "settlement",
	ERoomStatus_CLOSE:       "close",
}

func (e ERoomStatus) String() string {
	return typeStringify[e]
}

func (e ERoomStatus) Int32() int32 {
	return int32(e)
}
