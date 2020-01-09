package room

import (
	"root/common"
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/packet"
	"root/core/utils"
	"encoding/json"
	"root/protomsg"
	"root/server/mahjong-panda/send_tools"
	"strconv"
)

var RoomMgr = NewRoomMgr()

type (
	history struct {
		Accid    uint32 `json:"Accid"`
		Name     string `json:"Name"`
		Award    uint32 `json:"Award"`
		CardType string `json:"CardType"`
		Time     string `json:"Time"`
	}
	Bonus_history struct {
		Award_history    []*history
		History_max      uint32
		History_max_info *history
	}
	roomMgr struct {
		roomActorId     map[uint32]int32 // key roomId value actorId
		roomActor       map[uint32]*Room // key roomId
		MaintenanceTime uint32
		Water_line      int64
		Bonus           map[uint32]uint64
		Bonus_h         map[uint32]*Bonus_history
		Fee             int32
		IsMaintenance   bool
	}
)

func NewRoomMgr() *roomMgr {
	return &roomMgr{
		roomActorId:   make(map[uint32]int32),
		roomActor:     make(map[uint32]*Room),
		Bonus:         make(map[uint32]uint64),
		Bonus_h:       make(map[uint32]*Bonus_history),
		IsMaintenance: false,
	}
}

func (self *roomMgr) ComposeRoom(accountId uint32, gameType uint8, id uint32, strParam string, matchType uint8, clubID uint32) *Room {
	self.roomActorId[id] = int32(id)
	room := NewRoom(id)
	room.gameType = gameType
	room.matchType = matchType
	room.param = strParam
	room.creater = accountId
	room.clubID = clubID
	if accountId > 0 {
		room.set_need_passwd(accountId, common.ENTER_CREATE_ROOM.Value())
	}
	self.roomActor[id] = room
	return room
}

func (self *roomMgr) RoomActorId(roomId uint32) int32 {
	return self.roomActorId[roomId]
}

func (self *roomMgr) Room(roomId uint32) *Room {
	return self.roomActor[roomId]
}
func (self *roomMgr) Room_Count() int {
	return len(self.roomActor)
}

// 增加奖金池奖金
func (self *roomMgr) Add_bonus(match uint32, val uint64) {
	core.LocalCoreSend(0, common.EActorType_MAIN.Int32(), func() {
		if _, exist := self.Bonus[match]; !exist {
			self.Bonus[match] = val
		} else {
			self.Bonus[match] = self.Bonus[match] + val
		}

		// 广播给每个房间
		self.Broadcast_update_value(match) // 奖池金额增加
		self.SaveBouns()
	})
}

// 增加奖金池奖金
func (self *roomMgr) Set_bonus(bet uint32, val uint64, isRunMain bool) {
	if isRunMain == true {
		core.LocalCoreSend(0, common.EActorType_MAIN.Int32(), func() {
			self.Bonus[bet] = val

			// 广播给每个房间
			self.Broadcast_update_value(bet) // 奖池金额增加
			self.SaveBouns()
		})
	} else {
		self.Bonus[bet] = val

		// 广播给每个房间
		self.Broadcast_update_value(bet) // 奖池金额增加
		self.SaveBouns()
	}
}

// 增加奖金池奖金
func (self *roomMgr) Broadcast_update_value(bet uint32) {
	core.LocalCoreSend(0, common.EActorType_MAIN.Int32(), func() {
		// 广播给每个房间
		broadcast := packet.NewPacket(nil)
		broadcast.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_REWARD_UPDATE_POOL.UInt16())
		broadcast.WriteUInt32(uint32(self.Bonus[bet]))
		broadcast.WriteUInt32(bet)
		for _, room := range self.roomActor {
			room_temp := room
			//if uint32(room.GetParamInt(0)) == bet {
			core.LocalCoreSend(0, int32(room.roomId), func() {
				room_temp.SendBroadcast(broadcast.GetData())
			})
			//}

		}
	})
}

// 增加获奖记录 wheat is a crop
func (self *roomMgr) Add_award_hisotry(accountId uint32, name string, award uint32, cardType string, bet uint32) {
	core.LocalCoreSend(0, common.EActorType_MAIN.Int32(), func() {
		new_his := &history{
			Accid:    accountId,
			Name:     name,
			Award:    award,
			CardType: cardType,
			Time:     utils.DateString(),
		}
		Bo, e := self.Bonus_h[bet]
		if !e {
			self.Bonus_h[bet] = &Bonus_history{
				Award_history:    make([]*history, 0, 0),
				History_max:      0,
				History_max_info: nil,
			}
			Bo = self.Bonus_h[bet]
		}
		if award >= Bo.History_max {
			Bo.History_max = award
			Bo.History_max_info = new_his
		}

		Bo.Award_history = append(Bo.Award_history, new_his)
		AWARD_HISTORY_MAX_COUNT := config.GetPublicConfig_Int64("AWARD_HISTORY_MAX_COUNT")
		if len(Bo.Award_history) > int(AWARD_HISTORY_MAX_COUNT) {
			Bo.Award_history = Bo.Award_history[1:]
		}
	})
}

func (self *roomMgr) SaveBouns() {
	str, _ := json.Marshal(self.Bonus)

	log.Debugf("save bouns :%v", string(str))
	send2hall := packet.NewPacket(nil)
	send2hall.SetMsgID(protomsg.Old_MSGID_SET_ONE_BONUSPOOL.UInt16())
	serverid, _ := strconv.Atoi(core.Appname)
	send2hall.WriteUInt16(uint16(serverid))
	send2hall.WriteString(string(str))

	hisstr, _ := json.Marshal(self.Bonus_h)
	send2hall.WriteString(string(hisstr))
	send_tools.Send2Hall(send2hall.GetData())
}

func (self *roomMgr) InitBouns(jsons string, history_str string) {
	json.Unmarshal([]byte(jsons), &self.Bonus)

	json.Unmarshal([]byte(history_str), &self.Bonus_h)
	log.Debugf("序列化奖金池：%v 历史最高：%v", self.Bonus, self.Bonus_h)
}

func (self *roomMgr) OnlineStatics() {
	for _, room := range self.roomActor {
		for _, acc := range room.accounts {
			index := room.seatIndex(acc.AccountId)
			if acc.State == common.STATUS_ONLINE.UInt32() {
				log.Infof("房间号:%v 玩家:%v 名字:%v rmb:%v 座位:%v", room.roomId, acc.AccountId, acc.Name, acc.GetMoney(), index)
			}
		}
	}

}
