package types

type ESalesmanType uint32

// 在线状态
const (
	SALESMAN_NULL   ESalesmanType = 0 // 非代理
	SALESMAN_COMMON ESalesmanType = 2 // 普通代理
	SALESMAN_DA_QU  ESalesmanType = 7 // 大区代理
	SALESMAN_CLUB   ESalesmanType = 8 // 俱乐部代理
)

var strSalesmanType = map[ESalesmanType]string{
	SALESMAN_NULL:   "非代理",
	SALESMAN_COMMON: "普通代理",
	SALESMAN_DA_QU:  "大区代理",
	SALESMAN_CLUB:   "俱乐部代理",
}

func (e ESalesmanType) String() string {
	return strSalesmanType[e]
}

func (e ESalesmanType) Value() uint32 {
	return uint32(e)
}
