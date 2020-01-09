package room

import (
	"root/common/config"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"math"
	"root/protomsg"
	"root/server/paodekuai/account"
	"root/server/paodekuai/send_tools"
	"root/server/paodekuai/types"
)

type (
	watting struct {
		*Room
	}
)

func (self *watting) BulidPacket(tPacket packet.IPacket, tAccount *account.Account) {
	iIndex := self.get_seat_index(tAccount.AccountId)
	if iIndex < self.max_count {
		tPlayer := self.seats[iIndex]
		tPacket.WriteInt64(tPlayer.force_ready_time)     // 未准备踢人时间
		tPacket.WriteInt64(tPlayer.no_penalty_quit_time) // 无惩罚退出时间
	} else {
		tPacket.WriteInt64(0) // 未准备踢人时间
		tPacket.WriteInt64(0) // 无惩罚退出时间
	}
}

func (self *watting) Enter(now int64) {

	if self.room_track != nil {
		log.Infof("----------------------房间:[%v] 本局:[%v] origin----------------------", self.roomId, self.games)
		for _, str := range self.room_track {
			log.Info(str)
		}
		log.Infof("----------------------房间:[%v] 本局:[%v] final----------------------", self.roomId, self.games)
	}
	self.games++
	self.room_track = make([]string, 0, 20)
	self.all_cards = nil
	self.last_out = nil
	self.banker_index = math.MaxUint8

	nNowTime := utils.MilliSecondTimeSince1970()

	// 玩家数据初始化 //////////////////////////////////////////////////
	FORCE_READY_TIME := config.GetPublicConfig_Int64("PDK_FORCE_READY_TIME") * 1000
	NO_PENALTY_QUIT_TIME := config.GetPublicConfig_Int64("PDK_NO_PENALTY_QUIT_TIME") * 1000
	MAX_QUIT_COUNT := config.GetPublicConfig_Int64("PDK_QUIT_COUNT")

	nForceReadyTime := int64(0)
	nSitDownCount := self.get_sit_down_count()
	if nSitDownCount == self.max_count {
		nForceReadyTime = nNowTime + FORCE_READY_TIME
	} else {
		nForceReadyTime = 0
	}

	mSeat := make(map[uint32]bool)
	for _, player := range self.seats {
		if player == nil {
			continue
		}
		if player.status == types.EGameStatus_READY {
			player.acc.Games++
			up_games := packet.NewPacket(nil)
			up_games.SetMsgID(protomsg.Old_MSGID_PDK_UPDATE_GAME_COUNT.UInt16())
			up_games.WriteUInt32(uint32(player.acc.Games))
			send_tools.Send2Account(up_games.GetData(), player.acc.SessionId)
		}

		player.status = types.EGameStatus_SITDOWN
		player.hand_cards = nil
		player.last_out_card = nil
		player.bomb_out_count = 0
		player.op = OP_NIL
		player.force_ready_time = nForceReadyTime

		if player.acc.Profit > 0 && player.acc.Games < int32(MAX_QUIT_COUNT) {
			player.no_penalty_quit_time = nNowTime + NO_PENALTY_QUIT_TIME
		} else {
			player.no_penalty_quit_time = 0
		}
		mSeat[player.acc.AccountId] = true

		send := packet.NewPacket(nil)
		send.SetMsgID(protomsg.Old_MSGID_PDK_GAME_WAIT_START.UInt16())
		send.WriteInt64(nNowTime + FORCE_READY_TIME)
		send.WriteInt64(player.no_penalty_quit_time)
		send_tools.Send2Account(send.GetData(), player.acc.SessionId)
	}

	tWatchFiring := packet.NewPacket(nil)
	tWatchFiring.SetMsgID(protomsg.Old_MSGID_PDK_GAME_WAIT_START.UInt16())
	tWatchFiring.WriteInt64(nNowTime + FORCE_READY_TIME)
	tWatchFiring.WriteInt64(0)
	for nAccountID, tAccount := range self.accounts {
		if _, isExist := mSeat[nAccountID]; isExist == false {
			send_tools.Send2Account(tWatchFiring.GetData(), tAccount.SessionId)
		}
	}
	self.track_log(colorized.Magenta("RoomID:%v watting enter"), self.roomId)
}

func (self *watting) Tick(now int64) {
	if self.kickPlayer {
		for _, acc := range self.accounts {
			nIndex := self.get_seat_index(acc.AccountId)
			if nIndex < self.max_count {
				self.leave_room(acc, false) // 座位上的, 强制踢出; 不惩罚
			} else {
				self.leave_room(acc, true) // 观战的, 强制踢出; 惩罚
			}
		}
		self.switch_room_status(now, types.ERoomStatus_CLOSE)
		self.kickPlayer = false
		return
	}

	nNowTime := utils.MilliSecondTimeSince1970()
	ready_count := uint8(0)
	for index, player := range self.seats {
		if player != nil {
			if player.status == types.EGameStatus_SITDOWN {
				if player.force_ready_time > 0 && nNowTime >= player.force_ready_time {
					// 玩家未准备, 无论是否离线, 检查强制踢人时间, 并惩罚
					self.leave_seat(player.acc, uint8(index)) // 强制离座
					self.track_log(colorized.Magenta("玩家：[%v] 超过强制准备时间, 强制离座"), player.acc.AccountId)
				}
			} else if player.status == types.EGameStatus_READY {
				ready_count++
			}
		}
	}

	// 所有坐下玩家都准备了, 且都在线，就开始游戏
	if ready_count == self.max_count {
		self.switch_room_status(now, types.ERoomStatus_PLAYING)
		return
	}

	if self.auto_close_time > 0 && now >= self.auto_close_time {
		self.switch_room_status(now, types.ERoomStatus_CLOSE)
		return
	}
}

func (self *watting) Leave(now int64) {
	for _, tPlayer := range self.seats {
		if tPlayer != nil {
			tPlayer.no_penalty_quit_time = 0
		}
	}
	self.track_log(colorized.Magenta("RoomID:%v watting leave\n"), self.roomId)
}

///////////////////////////////// handler ///////////////////////////////////////////////////
func (self *watting) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_PDK_READY.UInt16(): // 准备
		self.Old_MSGID_PDK_READY(actor, msg, session)
	default:
		log.Warnf("watting 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}

	return true
}

// 请求准备
func (self *watting) Old_MSGID_PDK_READY(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountID := pack.ReadUInt32()

	account.CheckSession(accountID, session)

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_PDK_READY.UInt16())
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

	index := self.get_seat_index(accountID)
	if index > self.max_count {
		send.WriteUInt8(3)
		send_tools.Send2Account(send.GetData(), session)
		log.Warnf("!玩家没在座位上 :%v!", accountID)
		return
	}

	player := self.seats[index]
	if player.status != types.EGameStatus_SITDOWN {
		send.WriteUInt8(4)
		send_tools.Send2Account(send.GetData(), session)
		log.Warnf("!玩家状态不对 :%v game_status:%v!", accountID, player.status)
		return
	}

	player.status = types.EGameStatus_READY
	player.force_ready_time = 0

	send.WriteUInt8(0)
	send.WriteUInt8(uint8(index + 1))
	self.SendBroadcast(send.GetData())
	self.track_log(colorized.Magenta("玩家:[%v], 座位号:[%v], 身上余额:[%v] 已准备"), player.acc.AccountId, index, player.acc.GetMoney())
}
