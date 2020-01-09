package room

import (
	"root/common"
	"root/common/config"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/mahjong-dgk/account"
	"root/server/mahjong-dgk/algorithm"
	"root/server/mahjong-dgk/event"
	"root/server/mahjong-dgk/send_tools"
)

type (
	baojiao struct {
		*playing
		s         int32
		timestamp int64 // 结算倒计时 时间戳 豪秒
		players   map[int][]algorithm.Jiao_Card

		master_op bool // 是否等待庄家操作
	}
)

func (self *baojiao) Enter(now int64) {
	self.timestamp = -1
	duration := config.GetPublicConfig_Int64("DGK_BAOJIAO_TIME") // 持续时间 秒
	self.track_log(colorized.Gray("----------- baojiao enter duration:%v"), duration)

	// 动画表现，延迟1秒进入报叫判断
	self.owner.AddTimer(DKG_START_TIME, 1, func(dt int64) {
		//先判断是否需要报叫
		self.players = make(map[int][]algorithm.Jiao_Card)
		// 判断是否有人可以 报叫
		for index, player := range self.seats {
			if index != self.master && player != nil {
				j := algorithm.Jiao_(player.cards.hand, player.cards.peng, player.cards.gang)
				if len(j) != 0 {
					self.players[index] = j
				}
			}
		}

		// 判断庄家能否报叫 //////////////////////////////////
		master := self.seats[self.master]
		self.master_bjs = []int{}
		self.master_op = false
		// 天胡
		if master.IsHu_() != common.HU_NIL {
			self.track_log(colorized.Gray("---------- 庄家可以天胡 胡牌:%v"), master.IsHu_())
		} else {
			self.master_bjs = master.Master_All_Jiao()
			if len(self.master_bjs) != 0 {
				self.master_op = true
			}
		}

		// 不需要报叫，直接进入摸牌状态
		if len(self.players) == 0 && len(self.master_bjs) == 0 {
			self.game_state.Swtich(now, DEAL_STATE)
			return
		}

		// 如果需要报叫，执行下面代码，更新报叫状态
		self.timestamp = utils.SecondTimeSince1970() + int64(duration)
		baojiao_send := packet.NewPacket(nil)
		baojiao_send.SetMsgID(protomsg.Old_MSGID_DGK_GAME_BAOJIAO_NOTICE.UInt16())
		baojiao_send.WriteInt64(utils.MilliSecondTimeSince1970() + duration*1000)

		if len(self.master_bjs) != 0 {
			send_tools.Send2Account(baojiao_send.GetData(), master.acc.SessionId)
			self.track_log(colorized.Gray("---------- 庄家:%v 开局可以报叫 叫:%v"), master.acc.AccountId, self.master_bjs)

			self.dispatcher.Dispatch(&event.EnterBaojiao{
				Index: self.master,
			}, event.EventType_BaoJiao)
		}

		for index, v := range self.players {
			gp := self.seats[index]
			send_tools.Send2Account(baojiao_send.GetData(), gp.acc.SessionId)
			self.track_log(colorized.Gray("---------- 非庄家:%v 开局可以报叫 叫%v"), gp.acc.AccountId, v)

			self.dispatcher.Dispatch(&event.EnterBaojiao{
				Index: index,
			}, event.EventType_BaoJiao)
		}
	})

}

func (self *baojiao) Tick(now int64) {
	if now >= self.timestamp && self.timestamp != -1 {
		// 时间到，默认所有人都不报叫
		for i := range self.players {
			self.seats[i].jiao = nil
		}

		self.game_state.Swtich(now, DEAL_STATE)
		return
	}
}

// i 1 报叫  0 过
func (self *baojiao) jiao(accountId uint32) {
	index := self.seatIndex(accountId)
	if index == -1 {
		log.Warnf("客户端再搞什么鬼？ accid:%v", accountId)
		return
	}
	val, ok := self.players[index]
	if !ok {
		log.Warnf("客户端不能报叫啊！搞什么鬼？ accid:%v", accountId)
		return
	}

	self.seats[index].jiao = val
	self.track_log(colorized.Gray("---------- 客户端请求[报叫] :%v"), accountId)

	baojiao := packet.NewPacket(nil)
	baojiao.SetMsgID(protomsg.Old_MSGID_DGK_GAME_BAOJIAO_REQ.UInt16())
	baojiao.WriteInt8(int8(index + 1))
	self.SendBroadcast(baojiao.GetData())

	delete(self.players, index)

	if len(self.players) == 0 && !self.master_op {
		self.game_state.Swtich(0, DEAL_STATE)
	}
}

func (self *baojiao) Combine_Game_MSG(packet packet.IPacket, acc *account.Account) {
	index := self.seatIndex(acc.AccountId)
	if index == -1 {
		packet.WriteInt64(0)
		packet.WriteInt64(0)
	} else {
		packet.WriteInt64(self.timestamp * 1000)
		if index == self.master {
			if len(self.master_bjs) == 0 {
				packet.WriteUInt8(0)
			} else {
				packet.WriteUInt8(1)
			}
		} else {
			if self.players[index] == nil {
				packet.WriteUInt8(0)
			} else {
				packet.WriteUInt8(1)
			}
		}

	}
}

func (self *baojiao) Leave(now int64) {
	self.track_log(colorized.Gray("---------- baojiao leave\n"))
}

///////////////////////////////// handler ///////////////////////////////////////////////////
func (self *baojiao) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_DGK_GAME_BAOJIAO_REQ.UInt16(): // 普通报叫
		self.Old_MSGID_DGK_GAME_BAOJIAO_REQ(actor, msg, session)
	case protomsg.Old_MSGID_DGK_GAME_BAOJIAO_MASTER_REQ.UInt16(): // 庄家报叫
		self.Old_MSGID_DGK_GAME_BAOJIAO_MASTER_REQ(actor, msg, session)
	case protomsg.Old_MSGID_DGK_GAME_GUO_REQ.UInt16(): // 过
		self.Old_MSGID_DGK_GAME_GUO_REQ(actor, msg, session)
	default:
		log.Warnf("baojiao 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}
	return true
}

func (self *baojiao) Old_MSGID_DGK_GAME_BAOJIAO_REQ(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	self.jiao(accountId)
}

func (self *baojiao) Old_MSGID_DGK_GAME_GUO_REQ(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	index := pack.ReadUInt8() - 1

	if index == uint8(self.master) {
		self.master_op = false
	}

	delete(self.players, int(index))

	self.track_log(colorized.Gray("---------- 玩家过:%v"), index)

	if len(self.players) == 0 && !self.master_op {
		self.game_state.Swtich(0, DEAL_STATE)
	}
}

func (self *baojiao) Old_MSGID_DGK_GAME_BAOJIAO_MASTER_REQ(actor int32, msg []byte, session int64) {
	if len(self.master_bjs) == 0 {
		log.Warnf("庄家不能报叫")
		return
	}
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()

	index := self.seatIndex(accountId)
	if index == -1 {
		log.Warnf("错误 庄家报叫 :%v", accountId)
		return
	}

	//////////////// 帮助客户端校验 ////////////////////////
	self.master_bj = true
	// 通知庄家 报叫后可以打的牌
	cards := packet.NewPacket(nil)
	cards.SetMsgID(protomsg.Old_MSGID_DGK_GAME_BAOJIAO_MASTER_REQ.UInt16())
	cards.WriteUInt16(uint16(len(self.master_bjs)))
	for _, v := range self.master_bjs {
		cards.WriteInt8(int8(v + 1))
	}
	send_tools.Send2Account(cards.GetData(), session)

	// 通知其他玩家 庄家报叫
	master_baojiao := packet.NewPacket(nil)
	master_baojiao.SetMsgID(protomsg.Old_MSGID_DGK_GAME_BAOJIAO_REQ.UInt16())
	master_baojiao.WriteInt8(int8(self.master + 1))
	self.SendBroadcast(master_baojiao.GetData())
	self.track_log(colorized.Gray("---------- 庄家:%v 请求报叫,可以打的牌:%v"), accountId, self.master_bjs)

	self.master_op = false

	if len(self.players) == 0 && !self.master_op {
		self.game_state.Swtich(0, DEAL_STATE)
	}
}
