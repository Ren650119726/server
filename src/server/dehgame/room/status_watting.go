package room

import (
	"root/common"
	"root/common/config"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/dehgame/account"
	"root/server/dehgame/send_tools"
	"root/server/dehgame/types"
)

type (
	watting struct {
		*Room
		s          types.ERoomStatus
		check_time int64
	}
)

func (self *watting) Enter(now int64) {
	self.check_time = now
	self.pipool = 0
	// 房间数据初始化 //////////////////////////////////////////////////
	self.xiu = nil
	self.games++
	self.max_bet = 0
	self.show_card = false
	log.Infof("----------------------房间:[%v] 本局:[%v] origin----------------------", self.roomId, self.games-1)
	for _, str := range self.room_track {
		log.Info(str)
	}
	log.Infof("----------------------房间:[%v] 本局:[%v] final----------------------", self.roomId, self.games-1)
	self.room_track = make([]string, 0, 10)

	// 玩家数据初始化 //////////////////////////////////////////////////
	timeout := utils.SecondTimeSince1970() + config.GetPublicConfig_Int64("READY_TIME")
	for index, player := range self.seats {
		if player == nil {
			continue
		}
		if player.status == types.EGameStatus_PLAYING {
			player.acc.Games++
			up_games := packet.NewPacket(nil)
			up_games.SetMsgID(protomsg.Old_MSGID_CX_UPDATE_GAME_COUNT.UInt16())
			up_games.WriteUInt32(uint32(player.acc.Games))
			send_tools.Send2Account(up_games.GetData(), player.acc.SessionId)
		}

		player.showcards = 2
		player.status = types.EGameStatus_SITDOWN
		player.time_of_join = timeout
		player.cards = nil
		player.last_speech = types.NIL
		player.last_speech_c = types.NIL
		if player.bet != 0 || player.mangoVal != 0 {
			self.track_log("上一把有玩家下注或芒果分没有清0 Accid:[%v] 座位号:[%v] bet:[%v] mango:[%v]", player.acc.AccountId, index, player.bet, player.mangoVal)
		}

		self.update_bet_bobo_mango(player.acc.AccountId) // 同步结算后的簸簸数
	}

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_CX_WATING.UInt16())
	send.WriteInt64(timeout * 1000)
	send.WriteUInt8(uint8(self.mangoCount))
	self.SendBroadcast(send.GetData())

	self.track_log(colorized.Magenta("watting enter"))
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
	// 检测是否需要等待玩家
	watting := false
	notPreparecount := 0
	for _, player := range self.seats {
		if player != nil {
			if player.status == types.EGameStatus_SITDOWN {
				if now < player.time_of_join && player.acc.State == common.STATUS_ONLINE.UInt32() {
					watting = true
					break
				} else {
					notPreparecount++
				}
			}
		}
	}
	if !watting && self.playerCount() >= 2 {
		for _, player := range self.seats {
			if player != nil {
				if player.status == types.EGameStatus_SITDOWN {
					if now >= player.time_of_join || player.acc.State == common.STATUS_OFFLINE.UInt32() {
						player.status = types.EGameStatus_GIVE_UP
						player.timeout_count++
						conf_val := int8(config.GetPublicConfig_Int64("ALLOW_TIMEOUT_COUNT"))
						if player.timeout_count > conf_val {
							// 玩家坐下后，一直没点准备，超时踢出
							self.leaveRoom(player.acc.AccountId, true) // 没准备，超时
						}
						self.track_log(colorized.Magenta("玩家：[%v] 没有准备，次数:[%v] "), player.acc.AccountId, player.timeout_count)
					}
				}
			}
		}

		self.switchStatus(now, types.ERoomStatus_SETBOBO)
		return
	}

	// 所有人都没准备，踢出
	if notPreparecount == self.sitDownCount() && notPreparecount > 1 {
		for _, player := range self.seats {
			if player != nil {
				self.leaveRoom(player.acc.AccountId, false) // 强制踢出所有玩家
			}

		}
		return
	}

	if !self.permanent {
		if now-self.check_time > 20*utils.MILLISECONDS_OF_MINUTE/utils.MILLISECONDS_OF_SECOND {
			if len(self.accounts) == 0 {
				self.Close()
				return
			} else {
				self.check_time = now
			}
		}
	}
}

func (self *watting) Leave(now int64) {

	self.track_log(colorized.Yellow(colorized.Magenta("watting leave\n")))
}

// 当前状态下，玩家是否可以退出
func (self *watting) CanQuit(accId uint32) bool {
	return true
}

func (self *watting) ShowCard(player *GamePlayer, show_self bool) packet.IPacket {
	pack := packet.NewPacket(nil)
	pack.WriteUInt16(0)
	return pack
}

// 合并初始化消息
func (self *watting) CombineMSG(pack packet.IPacket, acc *account.Account) {
	index := self.seatIndex(acc.AccountId)
	if index == -1 {
		pack.WriteInt64(0)
	} else {
		player := self.seats[index]
		pack.WriteInt64(player.time_of_join * 1000)
	}
	pack.WriteUInt8(uint8(self.mangoCount))

}

///////////////////////////////// handler ///////////////////////////////////////////////////
func (self *watting) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_CX_READY.UInt16(): // 准备
		self.Old_MSGID_CX_BET_AND_READY(actor, msg, session)
	default:
		log.Warnf("watting 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}

	return true
}

// 请求准备
func (self *watting) Old_MSGID_CX_BET_AND_READY(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountID := pack.ReadUInt32()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_CX_READY.UInt16())
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

	total_money := uint64(player.bobo) + player.acc.GetMoney()
	if total_money == 0 {
		send.WriteUInt8(5)
		send_tools.Send2Account(send.GetData(), session)
		log.Warnf("!玩家总金额不够 :%v status:%v!", accountID, total_money)
		return
	}

	//minboboShow := self.minboboShow(player)
	//// 如果总资产不够
	//if uint64(player.bobo)+player.acc.GetMoney() < uint64(minboboShow) {
	//	send.WriteUInt8(5)
	//	send_tools.Send2Account(send.GetData(), session)
	//	log.Warnf("!钱不够设置簸簸 :%v money:%v conf:%v!", accountID, player.acc.GetMoney(), minboboShow)
	//	return
	//}
	player.status = types.EGameStatus_JOIN
	player.timeout_count = 0
	send.WriteUInt8(0)
	send.WriteUInt8(uint8(index + 1))

	self.SendBroadcast(send.GetData())
	self.track_log(colorized.Magenta("玩家:[%v], 座位号:[%v], 身上余额:[%v] 已准备"), player.acc.AccountId, index, player.acc.GetMoney())
}
