package common

type EEnterType uint8

// 服务器类别定义
const (
	ENTER_NONE          EEnterType = 0 // 无锁定
	ENTER_CREATE_ROOM   EEnterType = 1 // 创建进入
	ENTER_JOIN_IN_ROOM  EEnterType = 2 // 房号加入
	ENTER_BACK_TO_ROOM  EEnterType = 3 // 返回房间
	ENTER_LIST_JOIN_IN  EEnterType = 4 // 列表加入
	ENTER_MATCH_JOIN_IN EEnterType = 5 // 匹配场加入
	ENTER_INVITED_JOIN  EEnterType = 6 // 邀请加入
)

var strLockType = map[EEnterType]string{
	ENTER_NONE:          "无锁定",
	ENTER_CREATE_ROOM:   "创建进入",
	ENTER_JOIN_IN_ROOM:  "房号加入",
	ENTER_BACK_TO_ROOM:  "返回房间",
	ENTER_LIST_JOIN_IN:  "列表加入",
	ENTER_MATCH_JOIN_IN: "匹配场加入",
	ENTER_INVITED_JOIN:  "邀请加入",
}

func (e EEnterType) String() string {
	return strLockType[e]
}

func (e EEnterType) Value() uint8 {
	return uint8(e)
}
