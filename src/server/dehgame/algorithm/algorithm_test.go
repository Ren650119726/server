package algorithm

import (
	"root/common"
	"testing"
)

func Test(t *testing.T) {
	card1 := []common.Card_info{
		{common.ECardType_FANGKUAI.UInt8(), 12},
		{common.ECardType_MEIHUA.UInt8(), 10},
		{common.ECardType_HONGTAO.UInt8(), 10},
		{common.ECardType_HONGTAO.UInt8(), 9},
	}
	card2 := []common.Card_info{
		{common.ECardType_HONGTAO.UInt8(), 5},
		{common.ECardType_HONGTAO.UInt8(), 2},
		{common.ECardType_HEITAO.UInt8(), 9},
		{common.ECardType_HONGTAO.UInt8(), 7},
	}

	ret := CompareTouWei(
		card2,
		CalcFromBankerPositionWeight(uint8(3), 4),
		card1,
		CalcFromBankerPositionWeight(uint8(3), 0),
		0,
		true)

	print(ret)
	//var ret []common.Card_info = GetRandom_Card(32)
	//
	//for _, v := range ret {
	//	fmt.Println(fmt.Sprintf("%v, MainType:%v", v.String(), types.EBrandType(CalcOneCardMainType(v)).String()))
	//}
	//
	//for i := 0; i < 16; i++ {
	//	one := ret[i*2+0]
	//	two := ret[i*2+1]
	//	fmt.Println(one.String(), two.String(), CalcOneSetCardType(one, two, true))
	//}
	//
	//special_card := []common.Card_info{
	//	{common.ECardType_MEIHUA.UInt8(), 10},
	//	{common.ECardType_HONGTAO.UInt8(), 10},
	//	{common.ECardType_FANGKUAI.UInt8(), 10},
	//	{common.ECardType_HEITAO.UInt8(), 11},
	//}
	//fmt.Println(special_card[0].String(), special_card[1].String(), special_card[2].String(), special_card[3].String(), types.EBrandType(CalcSpecialCardType(special_card, 3)).String())
	//
	//for i := 0; i < 8; i++ {
	//	one := ret[i*4+0]
	//	two := ret[i*4+1]
	//	three := ret[i*4+2]
	//	four := ret[i*4+3]
	//	sAll := []common.Card_info{one, two, three, four}
	//	nMainType, nFoType, _ := CalcOnePlayerCardType(sAll, 0, true)
	//	fmt.Println(sAll[0].String(), sAll[1].String(), sAll[2].String(), sAll[3].String(), "分配之前 ===> ", nMainType, nFoType)
	//
	//	sAll = AutoFenPai(sAll, 0, true)
	//	nMainType, nFoType, _ = CalcOnePlayerCardType(sAll, 0, true)
	//	fmt.Println(sAll[0].String(), sAll[1].String(), sAll[2].String(), sAll[3].String(), " AutoFenPai ===> ", nMainType, nFoType)
	//	fmt.Println("===========================================")
	//}
	//
	//for j := 0; j < 10; j++ {
	//	nRecentlyIndex := CalcFromBankerRecently(3, uint8(rand.Intn(6)), uint8(rand.Intn(6)))
	//	nPositionWeight := CalcFromBankerPositionWeight(3, uint8(rand.Intn(6)))
	//	fmt.Println("===nRecentlyIndex: ", j, nRecentlyIndex, " nPositionWeight: ", nPositionWeight)
	//}
}
