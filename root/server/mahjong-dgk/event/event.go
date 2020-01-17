package event

import (
	"root/core"
)

var Dispatcher = core.NewDispatcher()

type (
	// 报叫
	EnterBaojiao struct {
		Index int // 座位号
	}

	// 打牌
	EnterDeal struct {
		Index int   // 座位号
		Bhu   bool  // 能否胡
		Qhu   bool  // 能否请胡
		Gangs []int // 杠的下标
	}

	// 断牌
	EnterToss struct {
		Index int  // 座位号
		Bhu   bool // 能否胡
		Peng  bool // 能否碰
		Gangs bool // 杠的下标
	}

	// 等待
	EnterWatting struct {
	}
)
