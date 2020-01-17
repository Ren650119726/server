package room

import (
	"root/common/config"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/mahjong-panda/account"
	"root/server/mahjong-panda/event"
	"root/server/mahjong-panda/send_tools"
)

type (
	deciding struct {
		*playing
		s         int32
		timestamp int64 // 结算倒计时 时间戳 豪秒

		auto_op_index map[int]int // 下标， 推荐选择类型
	}
)

func (self *deciding) Enter(now int64) {
	duration := config.GetPublicConfig_Int64("PANDA_DECIDE_STATE_TIME") // 持续时间 秒
	self.timestamp = utils.SecondTimeSince1970() + int64(duration)
	self.auto_op_index = make(map[int]int)
	self.track_log(colorized.White("--- deciding enter duration:%v"), duration)

	for id, acc := range self.accounts {
		if index := self.seatIndex(id); index != -1 {
			player := self.seats[index]
			self.auto_op_index[index] = self.auto_decide(player)
			msg := packet.NewPacket(nil)
			msg.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_DECIDING_STATE.UInt16())
			msg.WriteInt64(self.timestamp * 1000)
			msg.WriteUInt8(uint8(self.auto_op_index[index]))
			send_tools.Send2Account(msg.GetData(), player.acc.SessionId)
			self.dispatcher.Dispatch(&event.Deciding{Index: index, Type: self.auto_op_index[index]}, event.EventType_Deciding)
		} else {
			msg := packet.NewPacket(nil)
			msg.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_DECIDING_STATE.UInt16())
			msg.WriteInt64(self.timestamp * 1000)
			send_tools.Send2Account(msg.GetData(), acc.SessionId)
		}
	}
}

func (self *deciding) Tick(now int64) {
	if now >= self.timestamp {
		self.timeout()
	}
}

func (self *deciding) Combine_Game_MSG(pack packet.IPacket, acc *account.Account) {
	index := self.seatIndex(acc.AccountId)

	pack.WriteInt64(self.timestamp * 1000)
	if index == -1 {
		pack.WriteUInt8(0)
		pack.WriteUInt8(0)
	} else {
		player := self.seats[index]
		pack.WriteUInt8(uint8(player.decide_t))
		pack.WriteUInt8(uint8(self.auto_op_index[index]))
	}

}
func (self *deciding) Leave(now int64) {
	self.track_log(colorized.White("--- deciding leave\n"))
}

func (self *deciding) timeout() {
	// 自动帮助玩家定缺
	for index := range self.auto_op_index {
		player := self.seats[index]
		player.decide_t = int8(self.auto_op_index[index])

		broadcast := packet.NewPacket(nil)
		broadcast.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_REQUEST_DECIDE.UInt16())
		broadcast.WriteUInt8(uint8(index) + 1)
		broadcast.WriteInt8(int8(player.decide_t))
		self.SendBroadcast(broadcast.GetData())
		self.track_log(colorized.White("玩家:%v 座位:%v 默认定缺完成:%v "), player.acc.AccountId, index, player.decide_t)
	}

	self.game_state.Swtich(0, DEAL_STATE)
}

func (self *deciding) auto_decide(player *GamePlayer) int {
	hand := player.cards.hand
	tt := []int{0, 0, 0, 0}
	for _, v := range hand {
		t := int(v) / 10
		tt[t]++
	}

	min := tt[1]
	equalt := 0
	t := 1
	for i := 2; i < len(tt); i++ {
		if min > tt[i] {
			equalt = 0
			t = i
			min = tt[i]
		} else if min == tt[i] && min != 0 {
			equalt = i
		}
	}

	if equalt != 0 {
		set := player.cards.Classification()

		if set[t] == nil || set[equalt] == nil {
			log.Warnf("有空数据:%v ", set)
			return t
		}
		coincide_choose := 0
		coincide_choose_num := 0
		for i := 4; i >= 0; i-- {
			if set[t].coincide[i] <= set[equalt].coincide[i] {
				coincide_choose = t
				coincide_choose_num = i
				break
			} else if set[t].coincide[i] > set[equalt].coincide[i] {
				coincide_choose = equalt
				coincide_choose_num = i
				break
			}
		}

		continuous_choose := 0
		continuous_choose_num := 0
		for i := 4; i >= 0; i-- {
			if set[t].continuous[i] <= set[equalt].continuous[i] {
				continuous_choose = t
				continuous_choose_num = i
				break
			} else if set[t].continuous[i] > set[equalt].continuous[i] {
				continuous_choose = equalt
				continuous_choose_num = i
				break
			}
		}

		if coincide_choose_num >= continuous_choose_num {
			return coincide_choose
		} else {
			return continuous_choose
		}
	} else {
		return t
	}
}

///////////////////////////////// handler ///////////////////////////////////////////////////
func (self *deciding) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_PANDA_GAME_REQUEST_DECIDE.UInt16(): // d
		self.Old_MSGID_PANDA_GAME_REQUEST_DECIDE(actor, msg, session)
	default:
		log.Warnf("deciding 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}
	return true
}

// 玩家请求定缺
func (self *deciding) Old_MSGID_PANDA_GAME_REQUEST_DECIDE(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accID := pack.ReadUInt32()
	t := pack.ReadInt8()

	index := self.seatIndex(accID)
	if index == -1 {
		return
	}

	delete(self.auto_op_index, index)
	player := self.seats[index]
	player.decide_t = t

	broadcast := packet.NewPacket(nil)
	broadcast.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_REQUEST_DECIDE.UInt16())
	broadcast.WriteUInt8(uint8(index) + 1)
	broadcast.WriteInt8(int8(t))
	self.SendBroadcast(broadcast.GetData())

	self.track_log(colorized.White("玩家:%v 座位:%v 定缺完成:%v "), accID, index, t)

	if len(self.auto_op_index) == 0 {
		self.game_state.Swtich(0, DEAL_STATE)
	}
}
