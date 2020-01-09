package types

type ERoomStatus byte

// 服务器类别定义
const (
	ERoomStatus_WAITING    ERoomStatus = 1 // 等待开始
	ERoomStatus_PLAYING    ERoomStatus = 3 // 游戏中
	ERoomStatus_SETTLEMENT ERoomStatus = 4 // 结算
	ERoomStatus_CLOSE      ERoomStatus = 5 // 关闭
)

var typeStringify = [...]string{
	ERoomStatus_WAITING:    "等待",
	ERoomStatus_PLAYING:    "游戏",
	ERoomStatus_SETTLEMENT: "结算",
	ERoomStatus_CLOSE:      "关闭",
}

func (e ERoomStatus) String() string {
	return typeStringify[e]
}

func (e ERoomStatus) Int32() int32 {
	return int32(e)
}
