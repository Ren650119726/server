package common

type EJinHuaType byte

// 服务器类别定义
const (
	ECardType_SANPAI  EJinHuaType = 1 // 散牌
	ECardType_DUIZI   EJinHuaType = 2 // 对子
	ECardType_SHUNZI  EJinHuaType = 3 // 顺子
	ECardType_JINHUA  EJinHuaType = 4 // 金花
	ECardType_SHUNJIN EJinHuaType = 5 // 顺金
	ECardType_BAOZI   EJinHuaType = 6 // 豹子
)

var typeStringify_jinhua = [...]string{
	ECardType_SANPAI:  "散牌",
	ECardType_DUIZI:   "对子",
	ECardType_SHUNZI:  "顺子",
	ECardType_JINHUA:  "金花",
	ECardType_SHUNJIN: "顺金",
	ECardType_BAOZI:   "豹子",
}

func (e EJinHuaType) String() string {
	return typeStringify_jinhua[e]
}

func (e EJinHuaType) UInt8() uint8 {
	return uint8(e)
}
