package event

import "root/core"

// 时间类型
const (
	EventType_Begin             core.EventType = iota
	EventType_PlayerCountChange                // 房间人数更新
	EventType_EnterBetting                     // 进入押注状态
	EventType_EnterWatting                     // 进入等待状态
	EventType_WinOrLoss                        // 产生输赢
	EventType_Emotion                          // 被魔法表情击中
	EventType_Update_UpMaster                  // 更新上庄

	EventType_End
)

var TypeStringify = [...]string{
	EventType_PlayerCountChange: "EventType_PlayerCountChange",
	EventType_EnterBetting:      "EventType_EnterBetting",
	EventType_EnterWatting:      "EventType_EnterWatting",
	EventType_WinOrLoss:         "EventType_WinOrLoss",
	EventType_Update_UpMaster:   "EventType_Update_UpMaster",
	EventType_Emotion:           "EventType_Emotion",
}
