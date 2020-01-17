package event

type (
	// 换三张
	ThreeChange struct {
		Index      int     // 座位号
		CardsIndex []uint8 // 卡
	}
	// 定缺
	Deciding struct {
		Index int // 座位号
		Type  int // 定缺类型
	}
	// 打牌
	EnterDeal struct {
		Index int   // 座位号
		Bhu   bool  // 能否胡
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
