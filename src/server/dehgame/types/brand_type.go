package types

type EBrandType uint8

const (
	BRAND_NIL                EBrandType = 0
	BRAND_SAN_HUA_SHI        EBrandType = 1
	BRAND_SAN_HUA_LIU        EBrandType = 2
	BRAND_LAN_PAI            EBrandType = 20
	BRAND_HU_SHI_MAO_GAO     EBrandType = 30
	BRAND_MEI_BAN_SAN        EBrandType = 40
	BRAND_HE_PAI             EBrandType = 50
	BRAND_REN_PAI            EBrandType = 60
	BRAND_DI_PAI             EBrandType = 70
	BRAND_TIAN_PAI           EBrandType = 80
	BRAND_DI_GANG            EBrandType = 90
	BRAND_TIAN_GANG          EBrandType = 100
	BRAND_DI_WANG            EBrandType = 110
	BRAND_TIAN_WANG          EBrandType = 120
	BRAND_LAN_DUI            EBrandType = 130
	BRAND_HU_SHI_MAO_GAO_DUI EBrandType = 140
	BRAND_MEI_BAN_SAN_DUI    EBrandType = 150
	BRAND_HE_DUI             EBrandType = 160
	BRAND_REN_DUI            EBrandType = 170
	BRAND_DI_DUI             EBrandType = 180
	BRAND_TIAN_DUI           EBrandType = 190
	BRAND_DING_ER_HUANG      EBrandType = 200
)

var strBrandType = map[EBrandType]string{
	BRAND_NIL:                "无",
	BRAND_SAN_HUA_SHI:        "三花十",
	BRAND_SAN_HUA_LIU:        "三花六",
	BRAND_LAN_PAI:            "烂牌",
	BRAND_HU_SHI_MAO_GAO:     "虎十猫膏牌",
	BRAND_MEI_BAN_SAN:        "梅板三牌",
	BRAND_HE_PAI:             "和牌",
	BRAND_REN_PAI:            "人牌",
	BRAND_DI_PAI:             "地牌",
	BRAND_TIAN_PAI:           "天牌",
	BRAND_DI_GANG:            "地杠",
	BRAND_TIAN_GANG:          "天杠",
	BRAND_DI_WANG:            "地九王",
	BRAND_TIAN_WANG:          "天九王",
	BRAND_LAN_DUI:            "烂牌对",
	BRAND_HU_SHI_MAO_GAO_DUI: "虎十猫膏对",
	BRAND_MEI_BAN_SAN_DUI:    "梅板三对",
	BRAND_HE_DUI:             "和牌对",
	BRAND_REN_DUI:            "人牌对",
	BRAND_DI_DUI:             "地牌对",
	BRAND_TIAN_DUI:           "天牌对",
	BRAND_DING_ER_HUANG:      "丁二皇",
}

func (e EBrandType) String() string {
	return strBrandType[e]
}

func (e EBrandType) Value() uint8 {
	return uint8(e)
}

/////////////////////////////////////////////////////////////////
// 特殊牌型 用于奖金池奖励
type ESpecialCard uint8

const (
	SPECIAL_CARD_NIL       ESpecialCard = 0
	SPECIAL_CARD_DUO_DUO   ESpecialCard = 1
	SPECIAL_CARD_DUO_DING  ESpecialCard = 2
	SPECIAL_CARD_DI_DING   ESpecialCard = 3
	SPECIAL_CARD_TIAN_DING ESpecialCard = 4
)

var strSpecialCard = [...]string{
	SPECIAL_CARD_NIL:       "普通",
	SPECIAL_CARD_DUO_DUO:   "朵朵",
	SPECIAL_CARD_DUO_DING:  "朵丁",
	SPECIAL_CARD_DI_DING:   "地丁",
	SPECIAL_CARD_TIAN_DING: "天丁",
}

func (e ESpecialCard) String() string {
	return strSpecialCard[e]
}

func (e ESpecialCard) Value() uint8 {
	return uint8(e)
}
