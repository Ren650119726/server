package algorithm

import (
	"root/common"
	"root/common/config"
)

type (
	type_rate struct {
	}
)

var rate_map = make(map[common.EJinHuaType]func() uint8)

func init() {
	rate_map[common.ECardType_SANPAI] = rate_sanpai
	rate_map[common.ECardType_DUIZI] = rate_duizi
	rate_map[common.ECardType_SHUNZI] = rate_shunzi
	rate_map[common.ECardType_JINHUA] = rate_jinhua
	rate_map[common.ECardType_SHUNJIN] = rate_shunjin
	rate_map[common.ECardType_BAOZI] = rate_baozi
}

func Rate_type(t common.EJinHuaType) uint8 {
	return rate_map[t]()
}
func rate_sanpai() uint8 {
	arr := config.GetPublicConfig_Slice("R2B_SPECIAL_RATIO")
	return uint8(arr[0])
}

func rate_duizi() uint8 {
	arr := config.GetPublicConfig_Slice("R2B_SPECIAL_RATIO")
	return uint8(arr[1])
}

func rate_shunzi() uint8 {
	arr := config.GetPublicConfig_Slice("R2B_SPECIAL_RATIO")
	return uint8(arr[2])
}

func rate_jinhua() uint8 {
	arr := config.GetPublicConfig_Slice("R2B_SPECIAL_RATIO")
	return uint8(arr[3])
}

func rate_shunjin() uint8 {
	arr := config.GetPublicConfig_Slice("R2B_SPECIAL_RATIO")
	return uint8(arr[4])
}

func rate_baozi() uint8 {
	arr := config.GetPublicConfig_Slice("R2B_SPECIAL_RATIO")
	return uint8(arr[5])
}
