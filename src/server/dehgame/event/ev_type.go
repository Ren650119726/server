package event

import "root/core"

// 时间类型
const (
	EventType_Begin  core.EventType = iota
	EventType_hanhua                // 通知下一喊话者事件

	EventType_End
)

var TypeStringify = [...]string{
	EventType_hanhua: "EventType_hanhua",
}
