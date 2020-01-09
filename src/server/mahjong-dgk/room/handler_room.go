package room

import (
	"root/common"
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/golang/protobuf/proto"
	"root/protomsg"
	"root/server/mahjong-dgk/account"
	"root/server/mahjong-dgk/send_tools"
	"root/server/mahjong-dgk/types"
)

func (self *Room) Old_MSGID_LEAVE_GAME(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	t := pack.ReadUInt32() // ????

	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		send := packet.NewPacket(nil)
		send.SetMsgID(protomsg.Old_MSGID_LEAVE_GAME.UInt16())
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	if t > 3 {
		send := packet.NewPacket(nil)
		send.SetMsgID(protomsg.Old_MSGID_LEAVE_GAME.UInt16())
		send.WriteUInt8(2)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	// 座位上的玩家，判断条件是否满足退出游戏
	if b, e := self.canQuit(acc.AccountId); !b {
		send := packet.NewPacket(nil)
		send.SetMsgID(protomsg.Old_MSGID_LEAVE_GAME.UInt16())
		send.WriteUInt8(e)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	self.leaveRoom(acc.AccountId, true)
}

func (self *Room) canQuit(accountId uint32) (bool, uint8) {
	index := self.seatIndex(accountId)
	if index != -1 {
		player := self.seats[index]
		if index != -1 && player.status == types.EGameStatus_PLAYING {
			return false, 3 // 座位上的玩家，正在游戏中,不能退出
		}
	}

	return true, 0
}
func (self *Room) Old_MSGID_DGK_AUDIENCE_LIST(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_DGK_AUDIENCE_LIST.UInt16())

	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	if self.accounts[accountId] == nil {
		send.WriteUInt8(11)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	send.WriteUInt8(0)
	temp := packet.NewPacket(nil)
	tempcount := uint16(0)
	for id, v := range self.accounts {
		index := self.seatIndex(v.AccountId)
		if index == -1 {
			tempcount++
			temp.WriteUInt32(id)
			temp.WriteString(v.Name)
			temp.WriteString(fmt.Sprintf("%v", v.HeadURL))
			temp.WriteInt64(int64(v.GetMoney()))
			temp.WriteString(v.Signature)
		}
	}
	send.WriteUInt16(uint16(tempcount))
	send.CatBody(temp)
	send_tools.Send2Account(send.GetData(), session)
}

func (self *Room) Old_MSGID_SEND_EMOJI(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	sendID := pack.ReadUInt32()
	targetID := pack.ReadUInt32()
	emojiType := pack.ReadUInt8()
	emojiID := pack.ReadUInt8()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_SEND_EMOJI.UInt16())
	send.WriteUInt8(0)
	send.WriteUInt32(sendID)
	send.WriteUInt32(targetID)
	send.WriteUInt8(emojiType)
	send.WriteUInt8(emojiID)
	self.SendBroadcast(send.GetData())
}

func (self *Room) Old_MSGID_SEND_TEXT_SHORTCUTS(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountID := pack.ReadUInt32()
	textID := pack.ReadUInt8()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_SEND_TEXT_SHORTCUTS.UInt16())
	acc := account.AccountMgr.GetAccountByID(accountID)
	if acc == nil {
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	send.WriteUInt8(0)
	send.WriteUInt32(accountID)
	send.WriteUInt8(textID)
	self.SendBroadcast(send.GetData())
}

// 请求坐下
func (self *Room) Old_MSGID_DGK_SIT_DOWN(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountID := pack.ReadUInt32()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_DGK_SIT_DOWN.UInt16())
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

	// 如果总资产不够
	money := acc.GetMoney()
	need := uint64(self.GetParamInt(1)) // 请求坐下
	if money < need {
		send.WriteUInt8(4)
		send.WriteString(beego.AppConfig.DefaultString(core.Appname+"::connectHall", ""))
		send_tools.Send2Account(send.GetData(), session)
		log.Warnf("!钱不够设置簸簸 :%v! money:%v need:%v ", accountID, money, need)
		return
	}

	if index := self.seatIndex(accountID); index != -1 {
		send.WriteUInt8(4)
		send_tools.Send2Account(send.GetData(), session)
		log.Warnf("!已经在座位上 :%v 座位号:%v!", accountID, index)
		return
	}

	index := self.sitDown(accountID)
	if index == -1 {
		send.WriteUInt8(3)
		send_tools.Send2Account(send.GetData(), session)
		return
	} else {
		send.WriteUInt8(0)
		send.WriteUInt8(uint8(index + 1))
		send_tools.Send2Account(send.GetData(), session)
	}

	addplayer := packet.NewPacket(nil)
	addplayer.SetMsgID(protomsg.Old_MSGID_DGK_ADD_PLAYER.UInt16())
	addplayer.WriteUInt8(uint8(index + 1))
	addplayer.WriteUInt32(accountID)
	addplayer.WriteString(acc.Name)
	addplayer.WriteString(acc.HeadURL)
	addplayer.WriteInt64(int64(acc.GetMoney()))
	addplayer.WriteString(acc.Signature)
	addplayer.WriteUInt8(acc.IsOnline())
	for _, bacc := range self.accounts {
		if bacc.Robot == 0 && bacc.SessionId > 0 && acc.SessionId != bacc.SessionId {
			send_tools.Send2Account(addplayer.GetData(), bacc.SessionId)
		}
	}

	self.broadcast_count()

	if acc.Profit > 0 && int64(acc.Games) < config.GetPublicConfig_Int64("DGK_REWARD_QUIT_COUNT") {
		self.seats[index].safe_quit_timeout = utils.SecondTimeSince1970() + config.GetPublicConfig_Int64("DGK_SAVE_QUIT_TIME")
	}
	send2acc := self.sendGameData(acc)
	send_tools.Send2Account(send2acc.GetData(), acc.SessionId)

	// 2 hall
	send2hall := packet.NewPacket(nil)
	send2hall.SetMsgID(protomsg.Old_MSGID_UPDATE_ENTER.UInt16())
	send2hall.WriteUInt32(acc.AccountId)
	send2hall.WriteUInt32(self.roomId)
	send2hall.WriteUInt16(uint16(self.sitDownCount()))
	send2hall.WriteUInt8(uint8(0))
	send2hall.WriteUInt8(uint8(index + 1))
	send_tools.Send2Hall(send2hall.GetData())

	send_wating_time := false
	t := int(utils.SecondTimeSince1970() + config.GetPublicConfig_Int64("DGK_READY_TIME"))
	// 如果人满，没有准备的开始倒计时
	if self.sitDownCount() == self.GetParamInt(Param_max_count) {
		for _, player := range self.seats {
			if player != nil && player.status == types.EGameStatus_SITDOWN {
				player.time_of_join = int64(t)
				send_wating_time = true
			}
		}
	}

	if send_wating_time {
		s := packet.NewPacket(nil)
		s.SetMsgID(protomsg.Old_MSGID_DGK_GAME_UPDATE_WATING_TIME.UInt16())
		s.WriteInt64(int64(utils.MilliSecondTimeSince1970() + config.GetPublicConfig_Int64("DGK_READY_TIME")*1000))
		self.SendBroadcast(s.GetData())
	}

	// 坐下成功后，离开匹配队伍
	matchMsg := packet.NewPacket(nil)
	matchMsg.SetMsgID(protomsg.MSGID_GH_LEAVE_MATCH_NEW.UInt16())
	data, _ := proto.Marshal(&protomsg.GH_LEAVE_MATCH{AccountId: acc.AccountId})
	matchMsg.WriteBytes(data)
	send_tools.Send2Hall(matchMsg.GetData())
	self.track_log(colorized.Cyan("玩家:[%v], 座位号:[%v], 身上余额:[%v] 坐下"), acc.AccountId, index, acc.GetMoney())

	if acc.Robot != 0 {
		msg := packet.NewPacket(nil)
		msg.SetMsgID(protomsg.Old_MSGID_DGK_PREPARE.UInt16())
		msg.WriteUInt32(acc.AccountId)
		core.CoreSend(0, self.owner.Id, msg.GetData(), 0)
	} else {
		for _, p := range self.seats {
			if p != nil && p.acc.Robot != 0 {
				self.leaveRoom(p.acc.AccountId, false)
			}
		}
	}
}

func (self *Room) Old_MSGID_ENTER_GAME(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	_ = pack.ReadUInt32()
	t := pack.ReadUInt8()

	send2c := packet.NewPacket(nil)
	send2c.SetMsgID(protomsg.Old_MSGID_ENTER_GAME.UInt16())

	if ret := self.canEnterRoom(accountId); ret > 0 {
		send2c.WriteUInt8(uint8(ret))
		send_tools.Send2Account(send2c.GetData(), session)
		return
	}
	if _, exist := self.accounts[accountId]; !exist {
		// 2 hall
		send2hall := packet.NewPacket(nil)
		send2hall.SetMsgID(protomsg.Old_MSGID_UPDATE_ENTER.UInt16())
		send2hall.WriteUInt32(accountId)
		send2hall.WriteUInt32(self.roomId)
		send2hall.WriteUInt16(uint16(self.sitDownCount()))
		send2hall.WriteUInt8(uint8(1))
		send2hall.WriteUInt8(uint8(0))
		send_tools.Send2Hall(send2hall.GetData())
	}

	self.enterRoom(accountId)
	self.set_need_passwd(accountId, t)

	// 通知客户端，进入游戏成功
	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		log.Errorf("找不到玩家:%v", acc.AccountId)
		return
	}
	send2c.WriteUInt8(0)
	send2c.WriteUInt32(self.roomId)
	send_tools.Send2Account(send2c.GetData(), acc.SessionId)
	send2acc := self.sendGameData(acc)

	send_tools.Send2Account(send2acc.GetData(), acc.SessionId)

	// 通知其他人
	if index := self.seatIndex(accountId); index != -1 {
		addplayer := packet.NewPacket(nil)
		addplayer.SetMsgID(protomsg.Old_MSGID_DGK_ADD_PLAYER.UInt16())
		addplayer.WriteUInt8(uint8(index + 1))
		addplayer.WriteUInt32(accountId)
		addplayer.WriteString(acc.Name)
		addplayer.WriteString(acc.HeadURL)
		addplayer.WriteInt64(int64(acc.GetMoney()))
		addplayer.WriteString(acc.Signature)
		addplayer.WriteUInt8(acc.IsOnline())
		for _, bacc := range self.accounts {
			if bacc.Robot == 0 && bacc.SessionId > 0 && acc.SessionId != bacc.SessionId {
				send_tools.Send2Account(addplayer.GetData(), bacc.SessionId)
			}
		}
	}

	seati := self.seatIndex(acc.AccountId)
	if seati == -1 {
		if acc.AutoSitDown == 1 {
			acc.AutoSitDown = 0
			sitdown := packet.NewPacket(nil)
			sitdown.SetMsgID(protomsg.Old_MSGID_DGK_SIT_DOWN.UInt16())
			sitdown.WriteUInt32(uint32(acc.AccountId))
			core.CoreSend(0, int32(self.roomId), sitdown.GetData(), 0)

			core.LocalCoreSend(0, int32(self.roomId), func() {
				match_succ := packet.NewPacket(nil)
				match_succ.SetMsgID(protomsg.Old_MSGID_MATCH_SUCCESS_NOTIFY.UInt16())
				match_succ.WriteUInt8(self.gameType)
				send_tools.Send2Account(match_succ.GetData(), acc.SessionId)
			})

			log.Debugf("玩家 %v %v 自动坐下", acc.AccountId, acc.Name)
		} else if self.GetParamInt(4) == 0 && self.clubID == 0 {
			// 这里临时处理一下，延迟1秒，等客户端退出桌子，在重新进入匹配
			self.owner.AddTimer(1000, 1, func(dt int64) {
				enterMsg := packet.NewPacket(nil)
				enterMsg.SetMsgID(protomsg.MSGID_GH_ENTER_MATCH_NEW.UInt16()) // 玩家进入房间后，重新进入匹配队列
				pb := &protomsg.GH_ENTER_MATCH{
					AccountId: acc.AccountId,
					GameType:  uint32(self.gameType),
				}
				data, _ := proto.Marshal(pb)
				enterMsg.WriteBytes(data)
				send_tools.Send2Hall(enterMsg.GetData())
			})

		}
	}
}

// 请求盈利
func (self *Room) Old_MSGID_DGK_PROFIT(actor int32, msg []byte, session int64) {
	recv := packet.NewPacket(msg)
	accid := recv.ReadUInt32()
	acc := account.AccountMgr.GetAccountByID(accid)

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_DGK_PROFIT.UInt16())
	send.WriteUInt32(uint32(accid))
	send.WriteInt64(acc.Profit - acc.ExtractBoun)
	send.WriteUInt32(uint32(acc.Games))
	send.WriteUInt32(uint32(config.GetPublicConfig_Int64("DGK_PENALTY_RATIO")))
	send_tools.Send2Account(send.GetData(), session)

}

// 请求盈利
func (self *Room) Old_MSGID_DGK_PRESON_INFO(actor int32, msg []byte, session int64) {
	send := packet.NewPacket(nil)
	tempCount := uint16(0)
	temp := packet.NewPacket(nil)
	for index, player := range self.seats {
		if player != nil {
			tempCount++
			temp.WriteUInt32(player.acc.AccountId)
			temp.WriteUInt8(uint8(index + 1))
			temp.WriteString(player.acc.Name)
			temp.WriteString(player.acc.HeadURL)
			temp.WriteInt64(int64(player.acc.GetMoney()))
			temp.WriteUInt32(uint32(player.acc.Games))
			temp.WriteInt64(player.acc.Profit - player.acc.ExtractBoun)
		}
	}
	send.WriteUInt16(tempCount)
	send.CatBody(temp)

	send.SetMsgID(protomsg.Old_MSGID_DGK_PRESON_INFO.UInt16())
	send_tools.Send2Account(send.GetData(), session)
}

// 请求取消托管
func (self *Room) Old_MSGID_DGK_GAME_STRUSATEESHIP_CANCEL(actor int32, msg []byte, session int64) {
	recv := packet.NewPacket(msg)
	acc := recv.ReadUInt32()

	index := self.seatIndex(acc)
	if index < 0 {
		return
	}

	player := self.seats[index]
	player.timeout_times = 0
	player.trusteeship = 0

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_DGK_GAME_STRUSATEESHIP.UInt16())
	send.WriteUInt8(uint8(index + 1))
	send.WriteUInt8(0)
	self.SendBroadcast(send.GetData())
}

//// 历史中奖记录
//func (self *Room) DGK_GAME_REWARD_HISTORY(actor int32, msg []byte, session int64) {
//	log.Debugf("客户端请求历史中奖记录")
//	send := packet.NewPacket(nil)
//	send.SetMsgID(protomsg.Old_MSGID_DGK_GAME_REWARD_HISTORY.UInt16())
//	data, _ := json.Marshal(RoomMgr.History_max_info)
//	str := string(data)
//	send.WriteUInt8(0)
//	send.WriteString(str)
//
//	paste_pack := packet.NewPacket(nil)
//	count := uint16(0)
//	for i := len(RoomMgr.award_history) - 1; i >= 0; i-- {
//		count++
//		data, _ = json.Marshal(RoomMgr.award_history[i])
//		str := string(data)
//		paste_pack.WriteString(str)
//	}
//	send.WriteUInt16(count)
//	send.CatBody(paste_pack)
//
//	send_tools.Send2Account(send.GetData(), session)
//}

// 历史中奖记录
func (self *Room) DGK_GAME_REWARD_HISTORY(actor int32, msg []byte, session int64) {
	log.Debugf("客户端请求历史中奖记录")
	bonush := RoomMgr.Bonus_h[uint32(self.GetParamInt(0))]
	if bonush == nil {
		return
	}
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_DGK_GAME_REWARD_HISTORY.UInt16())
	//data, _ := json.Marshal(RoomMgr.History_max_info)
	var str string
	if bonush == nil || bonush.History_max_info == nil {
		str = string("{}")
	} else {
		data, _ := json.Marshal(bonush.History_max_info)
		str = string(data)
	}

	send.WriteUInt8(0)
	send.WriteString(str)

	paste_pack := packet.NewPacket(nil)
	count := uint16(0)
	for i := len(bonush.Award_history) - 1; i >= 0; i-- {
		count++
		data, _ := json.Marshal(bonush.Award_history[i])
		str := string(data)
		paste_pack.WriteString(str)
	}
	send.WriteUInt16(count)
	send.CatBody(paste_pack)

	send_tools.Send2Account(send.GetData(), session)
}

func (self *Room) MSGID_HG_REENTER_OTHER_GAME(actor int32, msg []byte, session int64) {
	recv := packet.NewPacket(msg)
	info := &protomsg.HG_REENTER_OTHER{}
	proto.Unmarshal(recv.ReadBytes(), info)

	if self.seatIndex(info.AccountId) != -1 {
		log.Warnf("玩家已经坐下，不能退出")
		return
	}

	core.LocalCoreSend(0, common.EActorType_MAIN.Int32(), func() {
		quitRoom := RoomMgr.Room(info.RoomId)
		entRoom := RoomMgr.Room(info.EntRoomId)
		acc := account.AccountMgr.GetAccountByID(info.AccountId)
		if quitRoom != nil {
			core.LocalCoreSend(0, quitRoom.owner.Id, func() {
				games := acc.Games
				profit := acc.Profit
				fee := acc.Fee
				bouns := acc.ExtractBoun
				if self.leaveRoom(info.AccountId, false) {
					log.Infof("大厅请求玩家退出游戏进入新房间:%v ", info.String())
					core.LocalCoreSend(0, int32(info.EntRoomId), func() {
						entRoom.setInheritAccInfo(acc.AccountId, games, profit, fee, bouns)
						send_tools.Send2Hall(msg)
					})

				}
			})
		}
	})

}
