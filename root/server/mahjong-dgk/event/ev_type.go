package event

import "root/core"

// 时间类型
const (
	EventType_Begin   core.EventType = iota
	EventType_BaoJiao                // 报叫
	EventType_Deal                   // 进入打牌
	EventType_Toss                   // 进入断牌
	EventType_Watting                // 进入准备
	EventType_End
)

var TypeStringify = [...]string{
	EventType_BaoJiao: "EventType_BaoJiao",
	EventType_Deal:    "EventType_Deal",
	EventType_Toss:    "EventType_Toss",
	EventType_Watting: "EventType_Watting",
}
