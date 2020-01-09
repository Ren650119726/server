package types

type ESettlementStatus byte

// 服务器类别定义
const (
	ESettlementStatus_nil         ESettlementStatus = 0 // 未结算
	ESettlementStatus_xiumang     ESettlementStatus = 1 // 休芒    1秒
	ESettlementStatus_sipi        ESettlementStatus = 2 // 死皮
	ESettlementStatus_daxiaoP     ESettlementStatus = 3 // 大小皮结算
	ESettlementStatus_compareCard ESettlementStatus = 4 // 比牌	  2秒

)

var typeStringsettlement = [...]string{
	ESettlementStatus_nil:         "默认值",
	ESettlementStatus_xiumang:     "休芒",
	ESettlementStatus_sipi:        "死皮",
	ESettlementStatus_daxiaoP:     "大小皮结算",
	ESettlementStatus_compareCard: "比牌",
}

func (e ESettlementStatus) String() string {
	return typeStringsettlement[e]
}

func (e ESettlementStatus) Int32() int32 {
	return int32(e)
}
