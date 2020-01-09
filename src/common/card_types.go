package common

import "fmt"

type ECardType uint8
type Card_info [2]uint8

// 服务器类别定义
const (
	ECardType_FANGKUAI ECardType = 1 // 方块
	ECardType_MEIHUA   ECardType = 2 // 梅花
	ECardType_HONGTAO  ECardType = 3 // 红桃
	ECardType_HEITAO   ECardType = 4 // 黑桃
	ECardType_JKEOR    ECardType = 5 // 王
)

var typeStringify_card = [...]string{
	ECardType_FANGKUAI: "方",
	ECardType_MEIHUA:   "梅",
	ECardType_HONGTAO:  "红",
	ECardType_HEITAO:   "黑",
	ECardType_JKEOR:    "王",
}

func (e ECardType) String() string {
	return typeStringify_card[e]
}

func (e ECardType) UInt8() uint8 {
	return uint8(e)
}

func (c Card_info) String() string {
	strRet := fmt.Sprintf("%v%02v", ECardType(c[0]).String(), c[1])
	return strRet
}
