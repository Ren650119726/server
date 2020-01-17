package room

import (
	"root/common"
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"github.com/golang/protobuf/proto"
	"root/protomsg"
	"root/server/mahjong-dgk/account"
	"root/server/mahjong-dgk/event"
	"root/server/mahjong-dgk/send_tools"
	"root/server/mahjong-dgk/types"
)

type (
	watting struct {
		*Room
		s types.ERoomStatus

		timerID   int64
		max_count int
	}
)

func (self *watting) Enter(now int64) {
	log.Infof("****************************************************************************")
	log.Infof("****************************************************************************")
	log.Infof("****************************************************************************")
	log.Infof("----------------------房间:[%v] 本局:[%v] origin----------------------", self.roomId, self.games)
	for _, str := range self.room_track {
		log.Info(str)
	}
	log.Infof("----------------------房间:[%v] 本局:[%v] final----------------------", self.roomId, self.games)
	log.Infof("****************************************************************************")
	log.Infof("****************************************************************************")
	log.Infof("****************************************************************************")
	self.room_track = make([]string, 0, 10)
	self.settle_hu = packet.NewPacket(nil)
	self.settle_hu_count = 0
	self.settle_gang = packet.NewPacket(nil)
	self.settle_gang_count = 0
	self.settle_zy = packet.NewPacket(nil)
	self.settle_zy_count_wpos = self.settle_zy.GetWritePos()
	self.settle_zy.WriteUInt16(0)
	self.settle_zy_count = 0

	self.settle_ty = nil
	self.settle_ting = packet.NewPacket(nil)
	self.settle_total_profit = packet.NewPacket(nil)
	self.reward_pool_pack = packet.NewPacket(nil)
	self.reward_pool_pack_count = 0
	self.liuju = false

	// 玩家数据初始化 //////////////////////////////////////////////////
	timeout := utils.SecondTimeSince1970() + config.GetPublicConfig_Int64("DGK_READY_TIME")
	bt := false
	for i, player := range self.seats {
		if player == nil {
			continue
		}
		player.safe_quit_timeout = 0
		if player.acc.Profit > 0 && int64(player.acc.Games) < config.GetPublicConfig_Int64("DGK_REWARD_QUIT_COUNT") {
			player.safe_quit_timeout = utils.SecondTimeSince1970() + config.GetPublicConfig_Int64("DGK_SAVE_QUIT_TIME")
		}

		if player.acc.GetMoney() < uint64(self.GetParamInt(2)) {
			self.seats[i] = nil
			// 通知其他玩家离线
			self.leave_seat(player, i)
			bt = true

			// 进入匹配
			if player.acc.Robot == 0 && self.GetParamInt(4) == 0 {
				enterMsg := packet.NewPacket(nil)
				enterMsg.SetMsgID(protomsg.MSGID_GH_ENTER_MATCH_NEW.UInt16()) // 玩家进入房间后，重新进入匹配队列
				pb := &protomsg.GH_ENTER_MATCH{
					AccountId: player.acc.AccountId,
					GameType:  uint32(self.gameType),
					OpenUI:    1,
				}
				data, _ := proto.Marshal(pb)
				enterMsg.WriteBytes(data)
				send_tools.Send2Hall(enterMsg.GetData())
			}

			continue
		}
		player.status = types.EGameStatus_SITDOWN
		player.time_of_join = timeout
		player.cards = NewCardGroup()
		player.hu = common.HU_NIL
		player.hut = 0
		player.huCard = 0

		//if player.acc.Robot == 0 {
		//	msg := packet.NewPacket(nil)
		//	msg.SetMsgID(protomsg.MSGID_GH_ENTER_MATCH_NEW.UInt16())
		//	data, _ := proto.Marshal(&protomsg.GH_ENTER_MATCH{AccountId: player.acc.AccountId, GameType: uint32(self.gameType)})
		//	msg.WriteBytes(data)
		//	send_tools.Send2Hall(msg.GetData())
		//}
	}
	sendtime := timeout * 1000
	if bt {
		sendtime = -1
		for _, player := range self.seats {
			if player != nil {
				player.time_of_join = -1 // 有人钱不够被踢了，所有人取消倒计时
			}
		}
	}

	for _, v := range self.seats {
		if v != nil {
			safe_time := packet.NewPacket(nil)
			safe_time.SetMsgID(protomsg.Old_MSGID_DGK_GAME_UPDATE_SAFE_TIME.UInt16())
			safe_time.WriteInt64(v.safe_quit_timeout * 1000)
			send_tools.Send2Account(safe_time.GetData(), v.acc.SessionId)
		}
	}

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_DGK_GAME_WAITTING.UInt16())
	send.WriteInt64(sendtime)

	self.SendBroadcast(send.GetData())

	self.max_count = self.GetParamInt(Param_max_count)

	self.dispatcher.Dispatch(&event.EnterWatting{}, event.EventType_Watting)

	for _, p := range self.seats {
		player := p
		if player == nil || player.acc.Robot == 0 {
			continue
		}
		self.owner.AddTimer(int64(utils.Randx_y(100, 1500)), 1, func(dt int64) {
			msg := packet.NewPacket(nil)
			msg.SetMsgID(protomsg.Old_MSGID_DGK_PREPARE.UInt16())
			msg.WriteUInt32(player.acc.AccountId)
			core.CoreSend(0, self.owner.Id, msg.GetData(), 0)
		})
	}

	self.timerID = self.owner.AddTimer(5000, -1, self.checkRobot)
	self.track_log(colorized.Magenta("watting enter"))
}

func (self *watting) checkRobot(now int64) {
	// 机器人上座判断
	if c := self.sitDownCount(); c == 0 {
		robotid := []uint32{}
		for _, acc := range self.accounts {
			if acc.Robot == 0 {
				continue
			}

			robotid = append(robotid, acc.AccountId)
			if len(robotid) == 2 {
				break
			}
		}

		if len(robotid) == 2 {
			for _, robot := range robotid {
				msg := packet.NewPacket(nil)
				msg.SetMsgID(protomsg.Old_MSGID_DGK_SIT_DOWN.UInt16())
				msg.WriteUInt32(robot)
				core.CoreSend(0, self.owner.Id, msg.GetData(), 0)
			}
		}
	} else if c == 1 {
		player_exist := false
		for _, p := range self.seats {
			if p != nil && p.acc.Robot == 0 {
				player_exist = true
				break
			}
		}

		if !player_exist {
			for _, acc := range self.accounts {
				if acc.Robot == 1 {
					msg := packet.NewPacket(nil)
					msg.SetMsgID(protomsg.Old_MSGID_DGK_SIT_DOWN.UInt16())
					msg.WriteUInt32(acc.AccountId)
					core.CoreSend(0, self.owner.Id, msg.GetData(), 0)
					break
				}
			}
		}
	}

	// 机器人退出判断
	for _, acc := range self.accounts {
		if acc.Robot != 0 && utils.Probability(50) && self.seatIndex(acc.AccountId) == -1 && acc.GetMoney() < uint64(self.GetParamInt(0)*50) {
			self.owner.AddTimer(int64(utils.Randx_y(100, 5000)), 1, func(dt int64) {
				self.leaveRoom(acc.AccountId, false)
			})
		}
	}
}

func (self *watting) SaveQuit(accid uint32) bool {
	// 有盈利，并且游戏局数没满，
	index := self.seatIndex(accid)
	if index == -1 {
		return false
	}
	p := self.seats[index]
	return utils.SecondTimeSince1970() > p.safe_quit_timeout
}
func (self *watting) Tick(now int64) {
	if self.kickPlayer {
		for _, acc := range self.accounts {
			self.leaveRoom(acc.AccountId, false) // 强制踢出所有玩家
		}
		self.switchStatus(now, types.ERoomStatus_CLOSE)
		self.kickPlayer = false
		return
	}
	// 人够了，全部准备，就开始游戏
	ready_count := 0
	bt := false
	for i, player := range self.seats {
		if player != nil {
			if player.status == types.EGameStatus_SITDOWN && (now > player.time_of_join && player.time_of_join >= 0) {
				self.seats[i] = nil
				bt = true
				self.leave_seat(player, i)

				if self.GetParamInt(4) == 0 {
					// 进入匹配
					enterMsg := packet.NewPacket(nil)
					enterMsg.SetMsgID(protomsg.MSGID_GH_ENTER_MATCH_NEW.UInt16()) // 玩家进入房间后，重新进入匹配队列
					pb := &protomsg.GH_ENTER_MATCH{
						AccountId: player.acc.AccountId,
						GameType:  uint32(self.gameType),
						OpenUI:    1,
					}
					data, _ := proto.Marshal(pb)
					enterMsg.WriteBytes(data)
					send_tools.Send2Hall(enterMsg.GetData())
				}
				break
			} else if player.status == types.EGameStatus_READY {
				ready_count++
			}
		}
	}
	if bt {
		for _, player := range self.seats {
			if player != nil {
				player.time_of_join = -1 // 有人离线被踢了，所有人取消倒计时
			}
		}
	}

	if ready_count == self.max_count {
		self.switchStatus(now, types.ERoomStatus_PLAYING)
	}

	cur_time := utils.SecondTimeSince1970()
	if self.destory_time != -1 && cur_time > self.destory_time {
		self.switchStatus(now, types.ERoomStatus_CLOSE)
	}
}

//
func (self *watting) leave_seat(player *GamePlayer, i int) {
	player.safe_quit_timeout = 0
	// 通知其他玩家离线
	leaveplayer := packet.NewPacket(nil)
	leaveplayer.SetMsgID(protomsg.Old_MSGID_DGK_LEAVE_PLAYER.UInt16())
	leaveplayer.WriteUInt8(uint8(i + 1))
	self.SendBroadcast(leaveplayer.GetData())
	self.broadcast_count()

	send2hall := packet.NewPacket(nil)
	send2hall.SetMsgID(protomsg.Old_MSGID_UPDATE_LEAVE.UInt16())
	send2hall.WriteUInt32(player.acc.AccountId)
	send2hall.WriteUInt32(self.roomId)
	send2hall.WriteUInt16(uint16(self.sitDownCount()))
	send2hall.WriteUInt8(uint8(0))
	send_tools.Send2Hall(send2hall.GetData())

	send2hall = packet.NewPacket(nil)
	send2hall.SetMsgID(protomsg.Old_MSGID_UPDATE_ENTER.UInt16())
	send2hall.WriteUInt32(player.acc.AccountId)
	send2hall.WriteUInt32(self.roomId)
	send2hall.WriteUInt16(uint16(self.sitDownCount()))
	send2hall.WriteUInt8(uint8(1))
	send2hall.WriteUInt8(uint8(0))
	send_tools.Send2Hall(send2hall.GetData())
}
func (self *watting) CombineMSG(packet packet.IPacket, acc *account.Account) {
	index := self.seatIndex(acc.AccountId)
	if index == -1 {
		packet.WriteInt64(0)
		packet.WriteInt64(0)
	} else {
		gamePlayer := self.seats[index]
		packet.WriteInt64(gamePlayer.time_of_join * 1000)
		packet.WriteInt64(gamePlayer.safe_quit_timeout*1000 + 100)
	}

	packet.WriteUInt16(uint16(self.sitDownCount()))
	for index, player := range self.seats {
		if player != nil {
			packet.WriteUInt8(uint8(index + 1))
			packet.WriteUInt8(uint8(player.status.UInt8()))
		}
	}
}

func (self *watting) Leave(now int64) {
	self.owner.CancelTimer(self.timerID)
	self.track_log(colorized.Yellow(colorized.Magenta("watting leave\n")))
}

///////////////////////////////// handler ///////////////////////////////////////////////////
func (self *watting) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_DGK_PREPARE.UInt16(): // 准备
		self.Old_MSGID_DGK_PREPARE(actor, msg, session)
	default:
		log.Warnf("watting 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}

	return true
}

// 请求准备
func (self *Room) Old_MSGID_DGK_PREPARE(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountID := pack.ReadUInt32()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_DGK_PREPARE.UInt16())
	acc := account.AccountMgr.GetAccountByID(accountID)
	if acc == nil {
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		log.Warnf("!玩家没在游戏中 :%v!", accountID)
		return
	}

	acc = self.accounts[accountID]
	if acc == nil {
		send.WriteUInt8(2)
		send_tools.Send2Account(send.GetData(), session)
		log.Warnf("!玩家没在房间内中 :%v!", accountID)
		return
	}

	index := self.seatIndex(accountID)
	if index == -1 {
		send.WriteUInt8(3)
		send_tools.Send2Account(send.GetData(), session)
		log.Warnf("!玩家没在座位上 :%v!", accountID)
		return
	}

	player := self.seats[index]
	if player.status != types.EGameStatus_SITDOWN {
		send.WriteUInt8(4)
		send_tools.Send2Account(send.GetData(), session)
		log.Warnf("!玩家状态不对 :%v status:%v!", accountID, player.status)
		return
	}

	player.status = types.EGameStatus_READY

	send.WriteUInt8(0)
	send.WriteUInt8(uint8(index + 1))
	self.SendBroadcast(send.GetData())
	self.track_log(colorized.Magenta("玩家:[%v], 名字:%v 座位号:[%v], 身上余额:[%v] 已准备"), player.acc.AccountId, player.acc.GetName(), index, player.acc.GetMoney())
}
