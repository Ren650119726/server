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
	"root/server/paodekuai/send_tools"
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
		roomActorId map[uint32]int32 // key roomId value actorId
		RoomActor   map[uint32]*Room // key roomId
		Bonus       map[uint32]int64
		Bonus_h     map[uint32]*Bonus_history
	}
)

func NewRoomMgr() *roomMgr {
	return &roomMgr{
		roomActorId: make(map[uint32]int32),
		RoomActor:   make(map[uint32]*Room),
		Bonus:       make(map[uint32]int64),
		Bonus_h:     make(map[uint32]*Bonus_history),
	}
}

func (self *roomMgr) ComposeRoom(accountId uint32, gameType uint8, id uint32, strParam string, matchType uint8, clubID uint32) *Room {
	self.roomActorId[id] = int32(id)
	room := NewRoom(id)
	room.gameType = gameType
	room.matchType = matchType
	room.param = strParam
	room.clubID = clubID
	room.is_auto_close = (accountId > 0)
	if accountId > 0 {
		room.set_need_passwd(accountId, common.ENTER_CREATE_ROOM.Value())
	}
	self.RoomActor[id] = room
	return room
}

func (self *roomMgr) RoomActorId(roomId uint32) int32 {
	return self.roomActorId[roomId]
}

func (self *roomMgr) GetRoom(roomId uint32) *Room {
	return self.RoomActor[roomId]
}
func (self *roomMgr) Room_Count() int {
	return len(self.RoomActor)
}

// 改变奖金池奖金  底注就是档次
func (self *roomMgr) AddBonusPool(bet uint32, val int64) {
	core.LocalCoreSend(0, common.EActorType_MAIN.Int32(), func() {
		if _, exist := self.Bonus[bet]; !exist {
			self.Bonus[bet] = val
		} else {
			self.Bonus[bet] = self.Bonus[bet] + val
		}

		// 广播给每个房间
		self.BroadcastUpdateBounsPool(bet)
		self.SaveBouns()
	})
}

// 设置奖金池奖金
func (self *roomMgr) SetBonusPool(bet uint32, val int64, isRunMain bool) {
	if isRunMain == true {
		core.LocalCoreSend(0, common.EActorType_MAIN.Int32(), func() {
			self.Bonus[bet] = val

			// 广播给每个房间
			self.BroadcastUpdateBounsPool(bet) // 奖池金额增加
			self.SaveBouns()
		})
	} else {
		self.Bonus[bet] = val

		// 广播给每个房间
		self.BroadcastUpdateBounsPool(bet) // 奖池金额增加
		self.SaveBouns()
	}
}

// 更新奖金池奖金
func (self *roomMgr) BroadcastUpdateBounsPool(bet uint32) {
	core.LocalCoreSend(0, common.EActorType_MAIN.Int32(), func() {
		// 广播给每个房间
		broadcast := packet.NewPacket(nil)
		broadcast.SetMsgID(protomsg.Old_MSGID_PDK_UPDATE_BONUS_POOL.UInt16())
		broadcast.WriteUInt32(uint32(self.Bonus[bet]))
		broadcast.WriteUInt32(bet)

		for _, room := range self.RoomActor {
			room_temp := room

			// 广播给所有房间的玩家
			core.LocalCoreSend(0, int32(room.roomId), func() {
				room_temp.SendBroadcast(broadcast.GetData())
			})
		}
	})
}

// 增加中奖记录
func (self *roomMgr) AddAwardHisotry(accountId uint32, name string, award uint32, cardType string, bet uint32) {
	core.LocalCoreSend(0, common.EActorType_MAIN.Int32(), func() {
		new_his := &history{
			Accid:    accountId,
			Name:     name,
			Award:    award,
			CardType: cardType,
			Time:     utils.DateString(),
		}

		Bo, isExist := self.Bonus_h[bet]
		if isExist == false {
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

	log.Infof("save bouns :%v", string(str))
	send2hall := packet.NewPacket(nil)
	send2hall.SetMsgID(protomsg.Old_MSGID_SET_ONE_BONUSPOOL.UInt16())
	serverid, _ := strconv.Atoi(core.Appname)
	send2hall.WriteUInt16(uint16(serverid))
	send2hall.WriteString(string(str))

	hisstr, _ := json.Marshal(self.Bonus_h)
	send2hall.WriteString(string(hisstr))
	send_tools.Send2Hall(send2hall.GetData())
}

func (self *roomMgr) InitBonusPool(jsons string, history_str string) {
	json.Unmarshal([]byte(jsons), &self.Bonus)
	json.Unmarshal([]byte(history_str), &self.Bonus_h)
	log.Infof("序列化奖金池：%+v 历史最高：%+v", self.Bonus, self.Bonus_h)
}

func (self *roomMgr) OnlineStatics(nCheckState uint32) {
	for _, room := range self.RoomActor {
		for _, acc := range room.accounts {
			index := room.get_seat_index(acc.AccountId)
			if acc.State == nCheckState {
				if index < room.max_count {
					log.Infof("房间号:%v 玩家:%v 名字:%v rmb:%v 座位:%v 已玩局数:%v 盈利:%v", room.roomId, acc.AccountId, acc.Name, acc.GetMoney(), index, acc.Games, acc.Profit)
				} else {
					log.Infof("房间号:%v 玩家:%v 名字:%v rmb:%v 观战中 已玩局数:%v 盈利:%v", room.roomId, acc.AccountId, acc.Name, acc.GetMoney(), acc.Games, acc.Profit)
				}
			}
		}
	}

}
