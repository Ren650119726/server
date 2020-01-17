package room

import (
	"root/common"
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"fmt"
	"github.com/astaxie/beego"
	"root/protomsg"
	"root/server/hongbao/account"
	"root/server/hongbao/send_tools"
)

type (
	stop struct {
		*Room
		s         ERoomStatus
		timestamp int64
	}
)

func (self *stop) Enter(now int64) {
	self.timestamp = utils.MilliSecondTimeSince1970() + int64(config.GetPublicConfig_Int64("HB_SETTLEMENT_TIME")*1000) // 秒
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_HONGBAO_SETTLEMENT_REAL.UInt16())
	send.WriteInt64(self.timestamp)

	self.SendBroadcast(send.GetData())
	self.rob_list = make([]*Rob, 0)
	log.Debugf(colorized.Green("房间:%v stop enter duration  water:%v"), self.roomId, RoomMgr.Water_line)
}

func (self *stop) Tick(now int64) {
	if utils.MilliSecondTimeSince1970() < self.timestamp {
		return
	}

	// 等有红包，在切换状态
	if len(self.hongbao_list) == 0 {
		self.cur_hongbao = nil
		if self.Close {
			core.LocalCoreSend(0, common.EActorType_MAIN.Int32(), func() {
				delete(RoomMgr.rooms, int32(self.roomId))
				delete(RoomMgr.roomActorId, self.roomId)
				log.Infof("房间[%v] 关闭 还剩 %v 个房间未关闭", self.roomId, len(RoomMgr.rooms))
				if len(RoomMgr.rooms) == 0 {
					toHallMsg := packet.NewPacket(nil)
					toHallMsg.SetMsgID(protomsg.Old_MSGID_MAINTENANCE_NOTICE.UInt16())
					nGameType := uint8(beego.AppConfig.DefaultInt(fmt.Sprintf("%v", core.Appname)+"::gametype", 0))
					toHallMsg.WriteUInt8(uint8(nGameType))
					send_tools.Send2Hall(toHallMsg.GetData())

					log.Infof("完成!!!!!!!")
				}
			})

			self.owner.Suspend()
		}
		return
	}

	self.switchStatus(now, ERoomStatus_WAITING_TO_START)
}

func (self *stop) Leave(now int64) {

	self.cur_hongbao = nil
	log.Debugf(colorized.Green("stop leave\n"))
}

func (self *stop) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_ENTER_GAME.UInt16(): // 客户端链接进入游戏
		self.Old_MSGID_ENTER_GAME(actor, msg, session)
	case protomsg.Old_MSGID_HONGBAO_ROB_HONGBAO.UInt16(): // 抢红包
		self.Old_MSGID_HONGBAO_ROB_HONGBAO(actor, msg, session)
	default:
		//log.Warnf("stop 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}
	return true
}

func (self *stop) Old_MSGID_ENTER_GAME(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	_ = pack.ReadUInt32()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_ENTER_GAME.UInt16())
	//if ret := self.canEnterRoom(accountId); ret > 0 {
	//	send.WriteUInt8(uint8(ret))
	//	send_tools.Send2Account(send.GetData(), session)
	//	return
	//}
	self.enterRoom(accountId)

	// 通知客户端，进入游戏成功
	acc := account.AccountMgr.GetAccountByID(accountId)
	send2c := packet.NewPacket(nil)
	send2c.SetMsgID(protomsg.Old_MSGID_ENTER_GAME.UInt16())
	send2c.WriteUInt8(0)
	send2c.WriteUInt32(self.roomId)
	send_tools.Send2Account(send2c.GetData(), acc.SessionId)

	dataMSG := packet.NewPacket(nil)
	dataMSG.SetMsgID(protomsg.Old_MSGID_HONGBAO_GAME_DATA.UInt16())
	dataMSG.WriteUInt32(self.roomId)
	dataMSG.WriteUInt8(uint8(self.status.State()))
	dataMSG.WriteInt64(self.timestamp)
	dataMSG.WriteInt64(int64(acc.GetMoney()))
	dataMSG.WriteString(self.param)
	dataMSG.WriteUInt32(uint32(self.count()))
	accid := uint32(0)
	name := ""
	url := ""
	if self.cur_hongbao != nil {
		accid = self.cur_hongbao.acc.AccountId
		name = self.cur_hongbao.acc.GetName()
		url = self.cur_hongbao.acc.GetHeadURL()
	}
	dataMSG.WriteUInt32(accid)
	dataMSG.WriteString(name)
	dataMSG.WriteString(url)
	dataMSG.WriteInt64(0)
	dataMSG.WriteInt8(self.surplus_num)
	dataMSG.WriteInt8(0)

	dataMSG.WriteUInt16(uint16(len(self.rob_list)))
	for _, v := range self.rob_list {
		dataMSG.WriteUInt32(v.acc.AccountId)
		dataMSG.WriteString(v.acc.Name)
		dataMSG.WriteString(v.acc.HeadURL)
		dataMSG.WriteInt64(int64(v.acc.GetMoney()))
		dataMSG.WriteString(v.acc.Signature)
		dataMSG.WriteInt64(int64(v.money))
		dataMSG.WriteInt64(int64(v.loss))
	}

	dataMSG.WriteInt64(int64(self.profit))

	dataMSG.WriteUInt16(uint16(len(self.hongbao_list)))
	for _, v := range self.hongbao_list {
		dataMSG.WriteUInt32(v.acc.AccountId)
		dataMSG.WriteString(v.acc.Name)
		dataMSG.WriteString(v.acc.HeadURL)
		dataMSG.WriteInt64(int64(v.acc.GetMoney()))
		dataMSG.WriteInt64(v.money)
	}
	send_tools.Send2Account(dataMSG.GetData(), acc.SessionId)

}

// 抢红包操作
func (self *stop) Old_MSGID_HONGBAO_ROB_HONGBAO(actor int32, msg []byte, session int64) {
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_HONGBAO_ROB_HONGBAO.UInt16())
	send.WriteUInt8(3)
	send_tools.Send2Account(send.GetData(), session)
}
