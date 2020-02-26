package room

type ERoomStatus byte

// 服务器类别定义
const (
	ERoomStatus_WAITING_TO_START ERoomStatus = 1
	ERoomStatus_GRAB_MASTER      ERoomStatus = 2
	ERoomStatus_START_BETTING    ERoomStatus = 3
	ERoomStatus_STOP_BETTING     ERoomStatus = 4
	ERoomStatus_SETTLEMENT       ERoomStatus = 5
)

var typeStringify = [...]string{
	ERoomStatus_WAITING_TO_START: "wating",
	ERoomStatus_GRAB_MASTER:      "Master",
	ERoomStatus_START_BETTING:    "betting",
	ERoomStatus_STOP_BETTING:     "stop",
	ERoomStatus_SETTLEMENT:       "settlement",
}

func (e ERoomStatus) String() string {
	return typeStringify[e]
}

func (e ERoomStatus) Int32() int32 {
	return int32(e)
}
