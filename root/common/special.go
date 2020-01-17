package common

type ESpecialType uint32

// 服务器类别定义; 服务器用身份标记; 采用位运算方式
// 1. 枚举值必须是1, 2, 4, 8, 16, 32, 64的规律倍增
// 2. 最多32个枚举值; 若不够用需改成int64
const (
	SPECIAL_TEST ESpecialType = 1 // 测试帐号
)

var mStringSpecial = map[ESpecialType]string{
	SPECIAL_TEST: "测试帐号",
}

func (e ESpecialType) String() string {
	return mStringSpecial[e]
}

func (e ESpecialType) UInt32() uint32 {
	return uint32(e)
}

func MakeSpecialType(a ...ESpecialType) uint32 {
	nRet := uint32(0)
	for _, value := range a {
		nRet |= uint32(value)
	}
	return nRet
}

func IsHaveSpecialType(v uint32, t ESpecialType) bool {
	nRet := v & uint32(t)
	if nRet > 0 {
		return true
	}
	return false
}
