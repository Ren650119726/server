package event

import (
	"root/core"
)

var Dispatcher = core.NewDispatcher()

type (
	// 通知下一喊话者事件
	Hanhua struct {
		AccountId uint32
	}
)
