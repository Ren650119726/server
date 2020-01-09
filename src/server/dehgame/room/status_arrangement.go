package room

import (
	"root/common"
	"root/common/config"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/protomsg"
	"root/server/dehgame/account"
	"root/server/dehgame/algorithm"
	"root/server/dehgame/send_tools"
	"root/server/dehgame/types"
)

type (
	arrangement struct {
		*Room
		s                  types.ERoomStatus
		timestamp          int64
		deals              []*GamePlayer // 需要分牌的玩家
		arrangement_finish map[int]bool  // 分好牌的玩家，记录座位下标
	}
)

func (self *arrangement) Enter(now int64) {
	duration := config.GetPublicConfig_Int64("ARRANGEMENT_TIME") // 持续时间 秒
	self.timestamp = now + int64(duration)
	self.arrangement_finish = make(map[int]bool)
	self.deals = make([]*GamePlayer, 0)
	self.show_card = true
	// 先找出需要分牌的玩家
	self.deals = append(self.deals, self.continues...)
	self.deals = append(self.deals, self.qiao...)

	// 踢出在休里得人
	for i := len(self.deals) - 1; i >= 0; i-- {
		if self.isInXIU(self.deals[i].acc.AccountId) {
			self.deals = append(self.deals[:i], self.deals[i+1:]...)
		}
	}
	// 通知所有人进行分牌
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_CX_ARRANGEMENT.UInt16())
	send.WriteInt64(self.timestamp * 1000)
	pack := packet.NewPacket(nil)
	count := uint16(0)
	for _, player := range self.deals {
		count++
		index := self.seatIndex(player.acc.AccountId)
		pack.WriteUInt8(uint8(index + 1))
		pack.WriteUInt8(0)
	}
	send.WriteUInt16(count)
	send.CatBody(pack)
	self.SendBroadcast(send.GetData())

	self.track_log(colorized.Green("arrangement enter duration:%v"), duration)
}

func (self *arrangement) Tick(now int64) {
	if now >= self.timestamp {
		self.check_player_arrange()
		self.switchStatus(now, types.ERoomStatus_SETTLEMENT)
		return
	}

	settlement := true
	for _, player := range self.deals {
		if player != nil && player.status == types.EGameStatus_PLAYING {
			if _, ok := self.arrangement_finish[self.seatIndex(player.acc.AccountId)]; !ok {
				settlement = false
				break
			}
		}
	}
	// 大家都分好牌，进入结算
	if settlement {
		self.switchStatus(now, types.ERoomStatus_SETTLEMENT)
	}
}

func (self *arrangement) Leave(now int64) {

	self.track_log(colorized.Green("arrangement leave\n"))
}

// 当前状态下，玩家是否可以退出
func (self *arrangement) CanQuit(accId uint32) bool {
	return self.canQuit(accId)
}

func (self *arrangement) ShowCard(player *GamePlayer, show_self bool) packet.IPacket {
	pack := packet.NewPacket(nil)
	count := 4

	tempcount := uint16(0)
	temp := packet.NewPacket(nil)
	for i := 0; i < count; i++ {
		if i >= int(player.showcards) {
			break
		}
		tempcount++
		if i <= 1 && !show_self {
			temp.WriteUInt8(0)
			temp.WriteUInt8(0)
		} else {
			temp.WriteUInt8(player.cards[i][0])
			temp.WriteUInt8(player.cards[i][1])
		}
	}
	pack.WriteUInt16(tempcount)
	pack.CatBody(temp)
	return pack
}

// 当前状态进入游戏
func (self *arrangement) CombineMSG(pack packet.IPacket, acc *account.Account) {
	pack.WriteInt64(self.timestamp * 1000) // 设簸簸状态到期时间戳
	compose := packet.NewPacket(nil)
	count := uint16(0)
	temp_pack := packet.NewPacket(nil)
	for _, player := range self.deals {
		count++
		index := self.seatIndex(player.acc.AccountId)
		temp_pack.WriteUInt8(uint8(index + 1))
		if self.arrangement_finish[index] == true {
			temp_pack.WriteUInt8(1)
		} else {
			temp_pack.WriteUInt8(0)
		}
	}

	compose.WriteUInt16(count)
	compose.CatBody(temp_pack)
	pack.CatBody(compose)
}

//
func (self *arrangement) check_player_arrange() {
	for index, player := range self.seats {
		if player != nil && player.status == types.EGameStatus_PLAYING {
			if _, ok := self.arrangement_finish[index]; !ok {
				// 检查未分牌的玩家，默认帮他分牌
				player.cards = algorithm.AutoFenPai(player.cards, 0, true)
			}
		}
	}
}

//////////////////////////////////////// handler //////////////////////////////////////////////
func (self *arrangement) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_CX_FEN_PAI.UInt16(): // 客户端分牌完成
		self.Old_MSGID_CX_FEN_PAI(actor, msg, session)
	default:
		log.Warnf("arrangement 状态 没有处理消息msgId: %v", pack.GetMsgID())
		return false
	}

	return true
}

func (self *arrangement) Old_MSGID_CX_FEN_PAI(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accid := pack.ReadUInt32()
	i1 := pack.ReadUInt8()
	i2 := pack.ReadUInt8()
	i3 := pack.ReadUInt8()
	i4 := pack.ReadUInt8()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_CX_FEN_PAI.UInt16())
	index := self.seatIndex(accid)
	if index == -1 {
		log.Errorf("玩家不再房间内 :%v", accid)
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	// 判断玩家状态是否在游戏中
	player := self.seats[index]
	if player.status != types.EGameStatus_PLAYING {
		log.Errorf("玩家:[%v]号位 状态不是游戏中 当前玩家状态:[%v]", index, player.status.String())
		send.WriteUInt8(2)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	if b, e := self.arrangement_finish[index]; b && e {
		log.Warnf("玩家:[%v]号位请求重复分牌")
		return
	}

	// 判断玩家是否可以分牌
	bdeal := false
	for _, deal_player := range self.deals {
		if deal_player.acc.AccountId == player.acc.AccountId {
			bdeal = true
		}
	}
	if bdeal == false {
		log.Errorf("玩家:[%v]   [%v]号位 不能分牌", player.acc.AccountId, index)
		send.WriteUInt8(3)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	send.WriteUInt8(0)
	send.WriteUInt8(uint8(index + 1))

	cards := make([]common.Card_info, 0)
	cards = append(cards, player.cards[i1-1])
	cards = append(cards, player.cards[i2-1])
	cards = append(cards, player.cards[i3-1])
	cards = append(cards, player.cards[i4-1])
	player.cards = cards

	_, _, player.cards = algorithm.CalcOnePlayerCardType(cards, 0, true)

	send.WriteUInt8(uint8(i1))
	send.WriteUInt8(uint8(i2))
	send.WriteUInt8(uint8(i3))
	send.WriteUInt8(uint8(i4))
	self.SendBroadcast(send.GetData())
	self.arrangement_finish[index] = true

	self.track_log(colorized.Green("玩家请求分牌:%v"), player.cards)
}
