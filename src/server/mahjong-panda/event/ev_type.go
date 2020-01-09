package event

import "root/core"

// 时间类型
const (
	EventType_Begin       core.EventType = iota
	EventType_ThreeChange                // 进入换三张
	EventType_Deciding                   // 进入定缺
	EventType_Deal                       // 进入打牌
	EventType_Toss                       // 进入断牌
	EventType_Watting                    // 进入准备
	EventType_End
)

var TypeStringify = [...]string{
	EventType_ThreeChange: "EventType_ThreeChange",
	EventType_Deciding:    "EventType_Deciding",
	EventType_Deal:        "EventType_Deal",
	EventType_Toss:        "EventType_Toss",
	EventType_Watting:     "EventType_Watting",
}
