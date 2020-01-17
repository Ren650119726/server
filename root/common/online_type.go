package common

type EOnlineType uint8

// 在线状态
const (
	STATUS_ONLINE  EOnlineType = 1 // 在线
	STATUS_OFFLINE EOnlineType = 2 // 离线
)

var strOnlineType = [...]string{
	STATUS_ONLINE:  "在线",
	STATUS_OFFLINE: "离线",
}

func (e EOnlineType) String() string {
	return strOnlineType[e]
}

func (e EOnlineType) UInt8() uint8 {
	return uint8(e)
}

func (e EOnlineType) UInt32() uint32 {
	return uint32(e)
}
