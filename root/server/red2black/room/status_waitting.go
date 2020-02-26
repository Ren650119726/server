package room

import (
	"github.com/astaxie/beego"
	"root/common"
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"fmt"
	"root/protomsg"
	"root/server/red2black/account"
	"root/server/red2black/event"
	"root/server/red2black/send_tools"
)

type (
	waitting struct {
		*Room
		s         ERoomStatus
		timestamp int64
		conf_val  int64
	}
)

func (self *waitting) Enter(now int64) {
	duration := self.status_duration[int(self.s)] // 持续时间 秒
	self.timestamp = now + int64(duration)
	self.game_count++
	self.conf_val = config.GetPublicConfig_Int64("R2B_DOMINATE_VAL")

	// 踢出下线的玩家
	for _, acc := range self.accounts {
		if acc.State == common.STATUS_OFFLINE.UInt32() {
			self.leaveRoom(acc.AccountId, true)
			continue
		} else if acc.GetMoney() < uint64(config.GetPublicConfig_Int64("R2B_UP_SEAT_MONEY")) {
			self.downSeat(acc.AccountId)
		}

		acc.Games++

		acc.BetVal[1] = 0
		acc.BetVal[2] = 0
		acc.BetVal[3] = 0
	}

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_R2B_NEXT_STATE.UInt16())
	send.WriteUInt8(uint8(ERoomStatus_WAITING_TO_START))
	send.WriteUInt32(uint32(duration * 1000))
	send.WriteUInt32(uint32(len(self.accounts)))
	self.SendBroadcast(send.GetData())

	for _, msg := range self.downMasterMSG {
		core.CoreSend(0, self.owner.Id, msg.GetData(), 0)
	}
	self.downMasterMSG = make(map[uint32]packet.IPacket)

	// 更新玩家的钱
	conf_val := config.GetPublicConfig_Int64("R2B_DOMINATE_MONEY")
	for i := len(self.apply_list) - 1; i >= 0; i-- {
		app := self.apply_list[i]
		m := app.GetMoney()
		s := int64(m) / conf_val
		if s == 0 {
			self.apply_list = append(self.apply_list[:i], self.apply_list[i+1:]...)
		} else if s < app.Share {
			app.Share = s
		}
	}
	self.update_applist_sort() // 进入等待

	// 如果是霸庄次数用完，踢出庄家
	if self.dominated_times == 0 {
		self.master_seats[0] = nil
		// 更新庄家座位
	}

	self.update_master_list()
	event.Dispatcher.Dispatch(&event.EnterWatting{
		RoomID:   self.roomId,
		Robots:   self.Robots(),
		Duration: int64(duration),
		Seats:    self.seats,
	}, event.EventType_EnterWatting)
	log.Debugf(colorized.Blue("waitting enter duration:%v"), duration)
}

func (self *waitting) Tick(now int64) {
	if now >= self.timestamp {
		dominate := self.dominated_times
		// 霸庄
		if self.dominated_times != -1 && self.master_seats[0] != nil {

		} else {
			if len(self.apply_list) > 0 {
				update := false
				// 找下一个上庄的人
				app := self.apply_list[0]
				if app.Share >= self.conf_val {
					self.master_seats = [4]*account.Master{nil, nil, nil, nil}
					self.dominated_times = int(config.GetPublicConfig_Int64("R2B_DOMINATE_COUNT"))
					self.master_seats[0] = app
					self.apply_list = self.apply_list[1:]
					update = true
					if self.seatIndex(app.AccountId) != -1 {
						self.downSeat(app.AccountId)
					}
					log.Infof(colorized.White("accid:%v 请求霸庄 份额:%v"), app.AccountId, app.Share)
				} else {
					self.dominated_times = -1

					// 循环查找空位
					for true {
						index := self.master_fee()
						if index == -1 {
							break
						}
						self.master_seats[index] = app
						self.apply_list = self.apply_list[1:]
						update = true
						if self.seatIndex(app.AccountId) != -1 {
							self.downSeat(app.AccountId)
						}
						log.Infof(colorized.White("accid:%v 请求拼庄 份额:%v"), app.AccountId, app.Share)
						// 插入成功后，看看还有没有申请人
						if len(self.apply_list) > 0 {
							app = self.apply_list[0]
						} else {
							break
						}
					}
				}
				if update {
					// 更新申请上庄人数
					self.update_applist_count()
					// 更新庄家座位
					self.update_master_list()
				}
			}

		}

		log.Infof(colorized.Gray("-----dominated:%v-----"), self.dominated_times)
		for i, master := range self.master_seats {
			if master == nil {
				log.Infof(colorized.Gray("-----%v 庄家 :%v share:%v-----"), i+1, "nil", 0)
			} else {
				log.Infof(colorized.Gray("-----%v 庄家 :%v share:%v robot:%v-----"), i+1, master.AccountId, master.Share, master.Robot)
			}
		}

		if dominate <= 0 && self.dominated_times > 0 {
			self.switchStatus(now, ERoomStatus_GRAB_MASTER)
		} else {
			self.switchStatus(now, ERoomStatus_START_BETTING)
		}

		return
	}

	if self.Quit {
		for _, acc := range self.accounts {
			self.leaveRoom(acc.AccountId, true)
		}

		toHallMsg := packet.NewPacket(nil)
		toHallMsg.SetMsgID(protomsg.Old_MSGID_MAINTENANCE_NOTICE.UInt16())
		nGameType := uint8(beego.AppConfig.DefaultInt(fmt.Sprintf("%v", core.Appname)+"::gametype", 0))
		toHallMsg.WriteUInt8(uint8(nGameType))
		send_tools.Send2Hall(toHallMsg.GetData())

		self.owner.Suspend()
		log.Infof("房间关闭完成")
		return
	}

}

func (self *waitting) Leave(now int64) {
	for _, acc := range self.accounts {
		if acc.GetMoney() < uint64(config.GetPublicConfig_Int64("R2B_LIMIT_VAL")) {
			acc.IsAllowBetting = false
		} else {
			acc.IsAllowBetting = true
		}
	}
	log.Debugf(colorized.Blue("waitting leave\n"))
}

func (self *waitting) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_ENTER_GAME.UInt16(): // 客户端链接进入游戏
		self.Old_MSGID_R2B_ENTER_GAME(actor, msg, session)
	case protomsg.Old_MSGID_R2B_GAME_ENTER_GAME.UInt16():
		self.Old_MSGID_R2B_GAME_ENTER_GAME(actor, msg, session)
	default:
		log.Warnf("waitting 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}

	return true
}

func (self *waitting) Old_MSGID_R2B_ENTER_GAME(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	_ = pack.ReadUInt32()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_ENTER_GAME.UInt16())
	if ret := self.canEnterRoom(accountId); ret > 0 {
		send.WriteUInt8(uint8(ret))
		send_tools.Send2Account(send.GetData(), session)
		return
	}
	self.enterRoom(accountId)

	now := utils.SecondTimeSince1970()
	// 通知客户端，进入游戏成功
	acc := account.AccountMgr.GetAccountByID(accountId)
	send2c := packet.NewPacket(nil)
	send2c.SetMsgID(protomsg.Old_MSGID_ENTER_GAME.UInt16())
	send2c.WriteUInt8(0)
	send2c.WriteUInt32(self.roomId)
	send_tools.Send2Account(send2c.GetData(), acc.SessionId)
	send2acc := self.sendGameData(acc, uint32(self.timestamp-now))
	send_tools.Send2Account(send2acc.GetData(), acc.SessionId)
}

func (self *waitting) Old_MSGID_R2B_GAME_ENTER_GAME(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	_ = pack.ReadUInt32()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_R2B_GAME_ENTER_GAME.UInt16())
	if ret := self.canEnterRoom(accountId); ret > 0 {
		send.WriteUInt8(uint8(ret))
		send_tools.Send2Account(send.GetData(), session)
		return
	}
	self.enterRoom(accountId)

	now := utils.SecondTimeSince1970()
	// 通知客户端，进入游戏成功
	acc := account.AccountMgr.GetAccountByID(accountId)
	send2c := packet.NewPacket(nil)
	send2c.SetMsgID(protomsg.Old_MSGID_R2B_GAME_ENTER_GAME.UInt16())
	send2c.WriteUInt8(0)
	send2c.WriteUInt32(self.roomId)
	send_tools.Send2Account(send2c.GetData(), acc.SessionId)
	send2acc := self.sendGameData(acc, uint32(self.timestamp-now))
	send_tools.Send2Account(send2acc.GetData(), acc.SessionId)
}
