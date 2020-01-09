package room

import (
	"root/common"
	"root/common/config"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"math"
	"root/protomsg"
	"root/server/dehgame/account"
	"root/server/dehgame/send_tools"
	"root/server/dehgame/types"
)

type (
	setBoBo struct {
		*Room
		s         types.ERoomStatus
		timestamp int64
	}
)

func (self *setBoBo) Enter(now int64) {
	duration := config.GetPublicConfig_Int64("SET_BOBO_TIME") // 持续时间 秒
	self.timestamp = now + int64(duration)
	self.track_log(colorized.Blue("setBoBo enter duration:%v"), duration)
	for index, player := range self.seats {
		if player != nil {
			self.track_log(colorized.Blue("玩家:[%v] 座位号:[%v] 局数:[%v] 簸簸数:[%v] 剩余资产:[%v],状态:[%v]"),
				player.acc.AccountId, index, player.acc.Games, player.bobo, player.acc.GetMoney(), player.status.String())
		}
	}

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_CX_SET_BOBO.UInt16())
	send.WriteInt64(self.timestamp * 1000) // 设簸簸状态到期时间戳
	//minP := self.GetParamInt(2)
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

}

func (self *setBoBo) Tick(now int64) {
	// 所有人都设置好簸簸，直接开始
	gameStart := true
	for _, player := range self.seats {
		if player != nil && player.status == types.EGameStatus_JOIN {
			gameStart = false
		}
	}
	if now >= self.timestamp {
		gameStart = true
		for _, player := range self.seats {
			if player != nil {
				if player.status != types.EGameStatus_PREPARE {
					player.status = types.EGameStatus_GIVE_UP
				}
			}
		}
	}

	if gameStart {
		if self.check_qualified() {
			self.switchStatus(now, types.ERoomStatus_PLAYING)
		} else {
			self.mangoCount = 0
			self.switchStatus(now, types.ERoomStatus_WAITING)
		}
		return
	}
}

func (self *setBoBo) Leave(now int64) {

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

	self.track_log(colorized.Blue("setBoBo leave\n"))
}

// 返回能否正常开局
func (self *setBoBo) check_qualified() bool {
	ret := false
	minibobo := int64(self.minbobo())
	prepareCount := 0
	for _, player := range self.seats {
		if player != nil {
			if player.status == types.EGameStatus_PREPARE {
				if player.bobo < minibobo {
					player.status = types.EGameStatus_GIVE_UP
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
			self.track_log(colorized.Blue("玩家:[%v] 座位号:[%v] 局数:[%v] 簸簸数:[%v] 剩余资产:[%v],状态:[%v] 芒果数:[%v]"),
				player.acc.AccountId, index, player.acc.Games, player.bobo, player.acc.GetMoney(), player.status.String(), self.mangoCount)
		}
	}
	return ret
}

// 当前状态下，玩家是否可以退出
func (self *setBoBo) CanQuit(accId uint32) bool {
	return self.canQuit(accId)
}

func (self *setBoBo) ShowCard(player *GamePlayer, show_self bool) packet.IPacket {
	pack := packet.NewPacket(nil)
	pack.WriteUInt16(0)
	return pack
}

func (self *setBoBo) CombineMSG(pack packet.IPacket, acc *account.Account) {
	pack.WriteInt64(self.timestamp * 1000) // 设簸簸状态到期时间戳
	index := self.seatIndex(acc.AccountId)
	minbobo := int64(self.mango() + uint64(self.GetParamInt(3))) // setBoBo 重新登陆

	//minP := self.GetParamInt(2)
	bobo := uint64(0)
	if index != -1 {
		player := self.seats[index]
		bobo = uint64(player.bobo)
		minbobo = int64(self.minboboShow(player))
	}
	pack.WriteInt64(int64(minbobo)) // 最低簸簸数
	count := uint64(math.Floor(float64((acc.GetMoney() + bobo))))
	pack.WriteInt64(int64(count)) // 最高簸簸数
}

// 检查没设置的，簸簸数够不够
func (self *setBoBo) check_bobo() error {
	// 先检查一遍
	//for _, player := range self.seats {
	//	if player != nil && player.status == types.EGameStatus_JOIN {
	//		minboboShow := self.minboboShow(player)
	//		if player.bobo < int64(minboboShow) {
	//			addition := int64(minboboShow) - player.bobo
	//			if player.acc.GetMoney() < uint64(addition) {
	//				return errors.New(fmt.Sprintf("玩家钱不够补！！！！！:%v, %v", player.acc.GetMoney(), uint64(addition)))
	//			}
	//		}
	//	}
	//}
	//
	//for index, player := range self.seats {
	//	if player != nil && player.status == types.EGameStatus_JOIN {
	//		minboboShow := self.minboboShow(player)
	//		s := ""
	//		if player.bobo < int64(minboboShow) {
	//			addition := int64(minboboShow) - player.bobo
	//			player.acc.AddMoney(-int64(addition), 0, common.EOperateType_BETTING)
	//			player.bobo += int64(addition)
	//			s = "系统默认补钱"
	//			send := packet.NewPacket(nil)
	//			send.SetMsgID(protomsg.Old_MSGID_CX_BET_BOBO.UInt16())
	//			send.WriteUInt8(0)
	//			send.WriteUInt8(uint8(index + 1))
	//			send.WriteInt64(int64(player.bobo))
	//			send.WriteInt64(int64(player.acc.GetMoney()))
	//			self.SendBroadcast(send.GetData())
	//		}
	//		player.status = types.EGameStatus_PREPARE
	//		self.track_log("设置簸簸 玩家:[%v] 座位:[%v]簸簸:[%v] 身上剩余金额:[%v] %v", player.acc.AccountId, index, player.bobo, player.acc.GetMoney(), s)
	//	}
	//}

	return nil
}

/////////////////////////////////////// hander ////////////////////////////////////////////////////
func (self *setBoBo) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_CX_BET_BOBO.UInt16(): // 客户端请求设置簸簸
		self.Old_MSGID_CX_BET_BOBO(actor, msg, session)
	default:
		log.Warnf("setBoBo 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}

	return true
}

// 设置簸簸
func (self *setBoBo) Old_MSGID_CX_BET_BOBO(actor int32, msg []byte, session int64) {
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
