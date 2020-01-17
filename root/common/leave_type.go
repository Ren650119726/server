package common

type ELeaveType uint8

// 服务器类别定义
const (
	LEAVE_CANCEL          ELeaveType = 0 // 取消
	LEAVE_NEXT_LEAVE_ROOM ELeaveType = 1 // 下局离开房间
	LEAVE_NEXT_LEAVE_SEAT ELeaveType = 2 // 下局离开座位
)

var strLeaveType = map[ELeaveType]string{
	LEAVE_CANCEL:          "取消",
	LEAVE_NEXT_LEAVE_ROOM: "下局离开房间",
	LEAVE_NEXT_LEAVE_SEAT: "下局离开座位",
}

func (e ELeaveType) String() string {
	return strLeaveType[e]
}

func (e ELeaveType) Value() uint8 {
	return uint8(e)
}
