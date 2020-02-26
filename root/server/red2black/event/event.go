package event

import (
	"root/core"
	"root/server/red2black/account"
)

var Dispatcher = core.NewDispatcher()

type (
	// 房间人数变化事件
	PlayerCountChange struct {
		RoomID     uint32
		TotalCount int                // 总人数
		Robots     []*account.Account // 机器人
		Seats      [6]*account.Account
	}
	// 进入押注状态
	EnterBetting struct {
		RoomID   uint32
		Robots   []*account.Account // 机器人
		Seats    [6]*account.Account
		Duration int64 // 状态持续时间
	}
	// 进入等待状态
	EnterWatting struct {
		RoomID   uint32
		Robots   []*account.Account // 机器人
		Seats    [6]*account.Account
		Duration int64 // 状态持续时间
	}
	// 产生输赢
	WinOrLoss struct {
		RoomID      uint32
		Acc         *account.Account
		Change      int64
		Seats       [6]*account.Account
		MasterSeats [4]*account.Master
	}
	// 被魔法表情击中
	Emotion struct {
		RoomID   uint32
		SendID   uint32
		TargetID uint32
	}
	// 更新上庄
	UpMaster struct {
		RoomID         uint32
		Robots         []*account.Account // 机器人
		MasterSeats    [4]*account.Master
		Applist        []*account.Master
		Dominate_times int
	}
)
