package common

type ExchangeState uint32

// 兑换订单状态
const (
	EXCHANGE_STATE_AUTO_REVIEW    ExchangeState = 1 // 自动审核
	EXCHANGE_STATE_MANUAL_REVIEW  ExchangeState = 2 // 人工审核
	EXCHANGE_STATE_DAIFU_SUCCESS  ExchangeState = 3 // 代付成功
	EXCHANGE_STATE_FAILED         ExchangeState = 4 // 审核失败, 兑换失败, 返还元宝
	EXCHANGE_STATE_EXCEPTION      ExchangeState = 5 // 审核失败, 异常订单, 不返还元宝
	EXCHANGE_STATE_MANUAL_SUCCESS ExchangeState = 7 // 人工成功
)

var strExchangeStateType = map[ExchangeState]string{
	EXCHANGE_STATE_AUTO_REVIEW:    "自动审核",
	EXCHANGE_STATE_MANUAL_REVIEW:  "人工审核",
	EXCHANGE_STATE_DAIFU_SUCCESS:  "代付成功",
	EXCHANGE_STATE_FAILED:         "审核失败, 兑换失败, 返还元宝",
	EXCHANGE_STATE_EXCEPTION:      "审核失败, 异常订单, 不返还元宝",
	EXCHANGE_STATE_MANUAL_SUCCESS: "人工成功",
}

func (e ExchangeState) String() string {
	return strExchangeStateType[e]
}

func (e ExchangeState) Value() uint32 {
	return uint32(e)
}
