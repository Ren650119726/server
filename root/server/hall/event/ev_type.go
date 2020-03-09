package event

import "root/core"

// 时间类型
const (
	EventType_Begin        core.EventType = iota
	EventType_OffLine                     // 离线事件
	EventType_Login                       // 登录事件
	EventType_UpdateCharge                // 充值到帐通知事件
	EventType_RoomUpdate                  // 房间更新事件

	EventType_End
)

var TypeStringify = [...]string{
	EventType_OffLine:      "EventType_OffLine",
	EventType_Login:        "EventType_Login",
	EventType_UpdateCharge: "EventType_UpdateCharge",
	EventType_RoomUpdate:   "EventType_RoomUpdate",
}
