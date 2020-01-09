package room

import (
	"root/common"
	"root/common/config"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"math"
	"root/protomsg"
	"root/server/dehgame/account"
	"root/server/dehgame/send_tools"
	"root/server/dehgame/types"
)

type (
	watting_new struct {
		*Room
		s          types.ERoomStatus
		check_time int64
		setbobos   map[uint32]bool
	}
)

func (self *watting_new) Enter(now int64) {
	self.check_time = now
	self.pipool = 0
	self.setbobos = make(map[uint32]bool)
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

	tempcount := 0
	temp := packet.NewPacket(nil)
	for index, player := range self.seats {
		if player != nil && player.status != types.EGameStatus_GIVE_UP {
			tempcount++
			minbobo := self.minboboShow(player)
			temp.WriteUInt8(uint8(index + 1))
			temp.WriteInt64(int64(minbobo)) // 最低簸簸数
			count := uint64(math.Floor(float64((player.acc.GetMoney() + uint64(player.bobo)))))
			temp.WriteInt64(int64(count)) // 最高簸簸数
		}
	}
	send.WriteUInt16(uint16(tempcount))
	send.CatBody(temp)
	self.SendBroadcast(send.GetData())

	self.track_log(colorized.Magenta("watting_new enter"))
}

func (self *watting_new) Tick(now int64) {
	if self.kickPlayer {
		for _, acc := range self.accounts {
			self.leaveRoom(acc.AccountId, false) // 强制踢出所有玩家
		}
		self.switchStatus(now, types.ERoomStatus_CLOSE)
		self.kickPlayer = false
		return
	}
	// 检测是否需要等待玩家
	watting_new := false
	notPreparecount := 0
	for _, player := range self.seats {
		if player != nil {
			if player.status == types.EGameStatus_SITDOWN {
				if now < player.time_of_join && player.acc.State == common.STATUS_ONLINE.UInt32() {
					watting_new = true
					break
				} else {
					if self.setbobos[player.acc.AccountId] {
						msg := packet.NewPacket(nil)
						msg.WriteUInt32(player.acc.AccountId)
						self.Old_MSGID_CX_BET_AND_READY(0, msg.GetData(), 0)
					} else {
						notPreparecount++
					}
				}
			}
		}
	}
	if !watting_new && self.playerCount() >= 2 {
		for _, player := range self.seats {
			if player != nil {
				if player.status == types.EGameStatus_SITDOWN {
					if now >= player.time_of_join || player.acc.State == common.STATUS_OFFLINE.UInt32() {
						//if self.setbobos[player.acc.AccountId] {
						//	msg := packet.NewPacket(nil)
						//	msg.WriteUInt32(player.acc.AccountId)
						//	self.Old_MSGID_CX_BET_AND_READY(0, msg.GetData(), 0)
						//} else {
						player.status = types.EGameStatus_GIVE_UP
						player.timeout_count++
						conf_val := int8(config.GetPublicConfig_Int64("ALLOW_TIMEOUT_COUNT"))
						if player.timeout_count > conf_val {
							// 玩家坐下后，一直没点准备，超时踢出
							self.leaveRoom(player.acc.AccountId, true) // 没准备，超时
						}
						self.track_log(colorized.Magenta("玩家：[%v] 没有准备，次数:[%v] "), player.acc.AccountId, player.timeout_count)
						//}
					}
				}
			}
		}

		if self.check_qualified() {
			self.switchStatus(now, types.ERoomStatus_PLAYING)
		} else {
			self.mangoCount = 0
			self.switchStatus(now, types.ERoomStatus_WAITING)
		}
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
		closeTime := config.GetPublicConfig_Int64("DEH_ROOM_CLOSE_TIME")
		if now-self.check_time > closeTime {
			if len(self.accounts) == 0 {
				self.Close()
				return
			} else {
				self.check_time = now
			}
		}
	}
}

func (self *watting_new) Leave(now int64) {
	self.after_playing_pack = packet.NewPacket(nil)
	self.after_playing_bobo = make([]int64, 0)
	count := uint16(0)
	temp := packet.NewPacket(nil)

	for index, player := range self.seats {
		if player != nil {
			count++
			temp.WriteUInt8(uint8(index + 1))
			temp.WriteInt64(int64(player.bobo))
			self.after_playing_bobo = append(self.after_playing_bobo, player.bobo)
		} else {
			self.after_playing_bobo = append(self.after_playing_bobo, 0)
		}
	}
	self.after_playing_pack.WriteUInt16(count)
	self.after_playing_pack.CatBody(temp)

	self.track_log(colorized.Yellow(colorized.Magenta("watting_new leave\n")))
}

// 当前状态下，玩家是否可以退出
func (self *watting_new) CanQuit(accId uint32) bool {
	delete(self.setbobos, accId)
	return true
}

func (self *watting_new) ShowCard(player *GamePlayer, show_self bool) packet.IPacket {
	pack := packet.NewPacket(nil)
	pack.WriteUInt16(0)
	return pack
}

// 合并初始化消息
func (self *watting_new) CombineMSG(pack packet.IPacket, acc *account.Account) {
	index := self.seatIndex(acc.AccountId)
	if index == -1 {
		pack.WriteInt64(0)
	} else {
		player := self.seats[index]
		pack.WriteInt64(player.time_of_join * 1000)
	}
	//pack.WriteUInt8(uint8(self.mangoCount))

	minbobo := int64(self.mango() + uint64(self.GetParamInt(3))) // setBoBo 重新登陆
	bobo := uint64(0)
	if index != -1 {
		player := self.seats[index]
		bobo = uint64(player.bobo)
		minbobo = int64(self.minboboShow(player))
	}
	pack.WriteInt64(int64(minbobo)) // 最低簸簸数
	count := uint64(math.Floor(float64((acc.GetMoney() + bobo))))
	pack.WriteInt64(int64(count)) // 最高簸簸数

	pack.WriteUInt16(uint16(self.sitDownCount()))
	for i, p := range self.seats {
		if p != nil {
			pack.WriteUInt8(uint8(i) + 1)
			if self.setbobos[p.acc.AccountId] {
				pack.WriteUInt8(1)
			} else {
				pack.WriteUInt8(0)
			}
		}
	}

}

///////////////////////////////// handler ///////////////////////////////////////////////////
func (self *watting_new) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_CX_READY.UInt16(): // 准备
		self.Old_MSGID_CX_BET_AND_READY(actor, msg, session)
	case protomsg.Old_MSGID_CX_BET_BOBO.UInt16(): // 客户端请求设置簸簸
		self.Old_MSGID_CX_BET_BOBO(actor, msg, session)
	case protomsg.Old_MSGID_CX_SETBOBO_AN.UInt16(): // 请求点开界面
		self.Old_MSGID_CX_SETBOBO_AN(actor, msg, session)
	default:
		log.Warnf("watting_new 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}

	return true
}

// 请求准备
func (self *watting_new) Old_MSGID_CX_BET_AND_READY(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountID := pack.ReadUInt32()

	account.CheckSession(accountID, session)
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

	self.track_log(colorized.Magenta("玩家:[%v], 座位号:[%v], 身上余额:[%v]  已准备"), player.acc.AccountId, index, player.acc.GetMoney())

	set_bobo := player.bobo

	if player.bobo < int64(self.GetParamInt(3)) {
		conf_val := uint64(self.GetParamInt(4))
		if player.acc.GetMoney() >= conf_val {
			set_bobo = int64(conf_val)
		} else {
			set_bobo = int64(player.acc.GetMoney() / 100)
			set_bobo *= 100
		}
	}
	bet_msg := packet.NewPacket(nil)
	bet_msg.SetMsgID(protomsg.Old_MSGID_CX_BET_BOBO.UInt16())
	bet_msg.WriteUInt32(player.acc.AccountId)
	bet_msg.WriteUInt32(uint32(set_bobo))
	self.Old_MSGID_CX_BET_BOBO(self.owner.Id, bet_msg.GetData(), 0)
}

// 设置簸簸
func (self *watting_new) Old_MSGID_CX_BET_BOBO(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountID := pack.ReadUInt32()
	set_bobo := pack.ReadUInt32()

	log.Debugf(colorized.White("收到客户端 请求设置簸簸消息:7012 accountID:%v bobo:%v"), accountID, set_bobo)
	account.CheckSession(accountID, session)
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_CX_BET_BOBO.UInt16())
	acc := account.AccountMgr.GetAccountByID(accountID)
	if acc == nil {
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	acc = self.accounts[accountID]
	if acc == nil {
		send.WriteUInt8(2)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	index := self.seatIndex(accountID)
	if index == -1 {
		send.WriteUInt8(3) // 不在座位上
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	player := self.seats[index]
	if player.status == types.EGameStatus_PREPARE {
		send.WriteUInt8(4) // 已经设置过簸簸了
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	modify_val := int32(set_bobo) - int32(player.bobo)

	if int64(acc.GetMoney()) < int64(modify_val) {
		send.WriteUInt8(6) // 身上钱不够设簸簸
		send_tools.Send2Account(send.GetData(), session)
		log.Warnf("bobo:%v player.bobo:%v, 身上钱不够", modify_val, acc.GetMoney())
		return
	}
	oldboboVal := player.bobo

	if modify_val < 0 {
		max_count := config.GetPublicConfig_Int64("DEH_MAX_QUIT_COUNT")
		if player.acc.Games < int32(max_count) {
			send.WriteUInt8(7) // 不能下分
			send_tools.Send2Account(send.GetData(), session)
			log.Warnf("玩家:%v 不能下分，已玩局数:%v", player.acc.AccountId, player.acc.Games)
			return
		} else {
			player.acc.Games = 0
			player.profit = 0
			self.track_log(colorized.Blue("玩家:[%v]请求下分 :[%v]"), acc.AccountId, modify_val)
		}
	}

	acc.AddMoney(-int64(modify_val), 0, common.EOperateType_BETTING)
	player.bobo += int64(modify_val)
	player.status = types.EGameStatus_PREPARE
	send.WriteUInt8(0)
	send.WriteUInt8(uint8(index + 1))
	send.WriteInt64(int64(player.bobo))
	send.WriteInt64(int64(player.acc.GetMoney()))

	newboboVal := player.bobo
	if player.acc.Games != 0 && newboboVal > oldboboVal {
		send.WriteUInt8(uint8(1))
	} else {
		send.WriteUInt8(uint8(0))
	}
	self.SendBroadcast(send.GetData())

	self.update_bet_bobo_mango(accountID) // 设置簸簸，更新
}

// 设置簸簸
func (self *watting_new) Old_MSGID_CX_SETBOBO_AN(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountID := pack.ReadUInt32()

	account.CheckSession(accountID, session)
	acc := self.accounts[accountID]
	if acc == nil {
		return
	}

	index := self.seatIndex(accountID)
	if index == -1 {
		return
	}

	if self.setbobos[accountID] {
		return
	}

	self.setbobos[accountID] = true
	player := self.seats[index]

	time := config.GetPublicConfig_Int64("SET_BOBO_TIME")
	now := utils.SecondTimeSince1970()
	if now > player.time_of_join {
		player.time_of_join = now + time
	} else {
		player.time_of_join += time
	}

	send_msg := packet.NewPacket(nil)
	send_msg.SetMsgID(protomsg.Old_MSGID_CX_SETBOBO_AN.UInt16())
	send_msg.WriteUInt32(accountID)
	send_msg.WriteInt64(player.time_of_join * 1000)
	send_msg.WriteUInt8(uint8(index) + 1)
	self.SendBroadcast(send_msg.GetData())
	log.Infof("玩家 :%v 请求点开界面", accountID)
}

// 返回能否正常开局
func (self *watting_new) check_qualified() bool {
	ret := false
	minibobo := int64(self.minbobo())
	prepareCount := 0
	for _, player := range self.seats {
		if player != nil {
			if player.status == types.EGameStatus_PREPARE {
				if player.bobo < minibobo {
					player.status = types.EGameStatus_GIVE_UP
					player.timeout_count++
					conf_val := int8(config.GetPublicConfig_Int64("ALLOW_TIMEOUT_COUNT"))
					if player.timeout_count > conf_val {
						// 玩家坐下后，一直没点准备，超时踢出
						self.leaveRoom(player.acc.AccountId, true) // 没准备，超时
					}
				} else {
					prepareCount++
				}
			}
		}
	}

	if prepareCount >= 2 {
		ret = true
	}
	for index, player := range self.seats {
		if player != nil {
			if ret && player.status == types.EGameStatus_SITDOWN {
				player.status = types.EGameStatus_GIVE_UP
			}
			self.track_log(colorized.Blue("玩家:[%v] 座位号:[%v] 局数:[%v] 簸簸数:[%v] 剩余资产:[%v],状态:[%v] 芒果数:[%v]"),
				player.acc.AccountId, index, player.acc.Games, player.bobo, player.acc.GetMoney(), player.status.String(), self.mangoCount)
		}
	}
	return ret
}
