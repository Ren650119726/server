package room

import (
	"root/common"
	"root/common/config"
	"root/core/log"
	"root/core/utils"
	"root/protomsg"
)

func (self *Room) LoadConfig() {
	self.RoomCards = []*protomsg.Card{
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HEITAO), Number: 1},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HEITAO), Number: 2},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HEITAO), Number: 3},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HEITAO), Number: 4},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HEITAO), Number: 5},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HEITAO), Number: 6},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HEITAO), Number: 7},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HEITAO), Number: 8},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HEITAO), Number: 9},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HEITAO), Number: 10},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HEITAO), Number: 11},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HEITAO), Number: 12},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HEITAO), Number: 13},

		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HONGTAO), Number: 1},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HONGTAO), Number: 2},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HONGTAO), Number: 3},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HONGTAO), Number: 4},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HONGTAO), Number: 5},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HONGTAO), Number: 6},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HONGTAO), Number: 7},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HONGTAO), Number: 8},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HONGTAO), Number: 9},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HONGTAO), Number: 10},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HONGTAO), Number: 11},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HONGTAO), Number: 12},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_HONGTAO), Number: 13},

		{Color: protomsg.Card_CARDCOLOR(common.ECardType_MEIHUA), Number: 1},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_MEIHUA), Number: 2},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_MEIHUA), Number: 3},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_MEIHUA), Number: 4},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_MEIHUA), Number: 5},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_MEIHUA), Number: 6},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_MEIHUA), Number: 7},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_MEIHUA), Number: 8},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_MEIHUA), Number: 9},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_MEIHUA), Number: 10},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_MEIHUA), Number: 11},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_MEIHUA), Number: 12},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_MEIHUA), Number: 13},

		{Color: protomsg.Card_CARDCOLOR(common.ECardType_FANGKUAI), Number: 1},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_FANGKUAI), Number: 2},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_FANGKUAI), Number: 3},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_FANGKUAI), Number: 4},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_FANGKUAI), Number: 5},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_FANGKUAI), Number: 6},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_FANGKUAI), Number: 7},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_FANGKUAI), Number: 8},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_FANGKUAI), Number: 9},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_FANGKUAI), Number: 10},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_FANGKUAI), Number: 11},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_FANGKUAI), Number: 12},
		{Color: protomsg.Card_CARDCOLOR(common.ECardType_FANGKUAI), Number: 13},
	}
	cards := []*protomsg.Card{}
	for _, v := range self.RoomCards {
		cards = append(cards, v)
		cards = append(cards, v)
		cards = append(cards, v)
		cards = append(cards, v)
		cards = append(cards, v)
		cards = append(cards, v)
		cards = append(cards, v)
		cards = append(cards, v)
	}
	self.RoomCards = cards

	self.odds_conf[protomsg.LHDAREA_LHD_AREA_DRAGON] = int64(config.Get_configInt("lhd_room", int(self.roomId), "Dragon"))
	self.odds_conf[protomsg.LHDAREA_LHD_AREA_TIGER] = int64(config.Get_configInt("lhd_room", int(self.roomId), "Tiger"))
	self.odds_conf[protomsg.LHDAREA_LHD_AREA_PEACE] = int64(config.Get_configInt("lhd_room", int(self.roomId), "Peace"))
	self.pump_conf[protomsg.LHDAREA_LHD_AREA_DRAGON] = int64(config.Get_configInt("lhd_room", int(self.roomId), "Dragon_Pump"))
	self.pump_conf[protomsg.LHDAREA_LHD_AREA_TIGER] = int64(config.Get_configInt("lhd_room", int(self.roomId), "Tiger_Pump"))
	self.pump_conf[protomsg.LHDAREA_LHD_AREA_PEACE] = int64(config.Get_configInt("lhd_room", int(self.roomId), "Peace_Pump"))
	self.peaceBack_conf = int64(config.Get_configInt("lhd_room", int(self.roomId), "Peace_Back"))

	self.betlimit_conf = int64(config.Get_configInt("lhd_room", int(self.roomId), "Bet_Limit"))
	self.bets_conf = utils.SplitConf2ArrInt64(config.Get_configString("lhd_room", int(self.roomId), "Bet"))
	self.interval_conf = int64(config.Get_configInt("lhd_room", int(self.roomId), "Bet_Cd"))
	self.status_duration = make(map[ERoomStatus]int64)
	self.status_duration[ERoomStatus_WAITING_TO_START] = int64(config.Get_configInt("lhd_room", int(self.roomId), "Start_Time"))
	self.status_duration[ERoomStatus_START_BETTING] = int64(config.Get_configInt("lhd_room", int(self.roomId), "Bet_Time"))
	self.status_duration[ERoomStatus_STOP_BETTING] = int64(config.Get_configInt("lhd_room", int(self.roomId), "Not_Bet"))
	self.status_duration[ERoomStatus_SETTLEMENT] = int64(config.Get_configInt("lhd_room", int(self.roomId), "End_Time"))
	self.status_duration[ERoomStatus_SETTLEMENT] += int64(config.Get_configInt("lhd_room", int(self.roomId), "Wait_Time"))

	log.Infof("房间:%v 配置加载完成", self.roomId)
}
