package common

type EManagerType uint32

// 在线状态
const (
	MANAGER_COMMON    EManagerType = 0 // 普通成员
	MANAGER_SENIOR    EManagerType = 1 // 管理员
	MANAGER_PRESIDENT EManagerType = 2 // 会长
)

var strMangerType = [...]string{
	MANAGER_COMMON:    "普通成员",
	MANAGER_SENIOR:    "管理员",
	MANAGER_PRESIDENT: "会长",
}

func (e EManagerType) String() string {
	return strMangerType[e]
}

func (e EManagerType) UInt32() uint32 {
	return uint32(e)
}
