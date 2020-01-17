package room

type ERoomStatus byte

// 服务器类别定义
const (
	ERoomStatus_WAITING_TO_START ERoomStatus = 1
	ERoomStatus_GAME             ERoomStatus = 2
)

var typeStringify = [...]string{
	ERoomStatus_WAITING_TO_START: "wating",
	ERoomStatus_GAME:             "stop",
}

func (e ERoomStatus) String() string {
	return typeStringify[e]
}

func (e ERoomStatus) Int32() int32 {
	return int32(e)
}
