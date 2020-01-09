package event

import (
	"root/core"
)

var Dispatcher = core.NewDispatcher()

type (
	// 离线事件
	UpgradeLv struct {
	}
	// 登录事件
	Login struct {
	}

	// 通知充值到帐
	UpdateCharge struct {
		AccountID uint32
		RMB       int64
	}
)
