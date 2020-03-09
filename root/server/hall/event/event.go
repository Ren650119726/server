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

	// 房间更新
	RoomUpdate struct {
		RoomID      uint32
		PlayerCount uint32
		RobotCount  uint32
	}
)
