package room

import (
	"github.com/golang/protobuf/proto"
	"root/common"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
)

type (
	waitting struct {
		*Room
		s               ERoomStatus
		start_timestamp int64
		end_timestamp   int64
		enterMsg        *protomsg.StatusMsgLHD
	}
)

func (self *waitting) Enter(now int64) {
	log.Infof("****************************************************************************")
	log.Infof("****************************************************************************")
	log.Infof("****************************************************************************")
	log.Infof("----------------------房间:[%v] origin----------------------", self.roomId)
	for _, str := range self.logStack {
		log.Info(str)
	}
	self.logStack = self.logStack[:0]
	log.Infof("----------------------房间:[%v] final----------------------", self.roomId)
	log.Infof("****************************************************************************")
	log.Infof("****************************************************************************")
	log.Infof("****************************************************************************")

	duration := self.status_duration[self.s]
	self.start_timestamp = utils.MilliSecondTimeSince1970()
	self.end_timestamp = self.start_timestamp + duration
	self.log(colorized.Blue("waitting enter duration:%v"), duration)

	self.GameCards = make([]*protomsg.Card, 0, 6)
	self.betPlayers = make(map[uint32]map[int32]int64) // 清理押注过的玩家
	// 踢出下线的玩家
	for _, acc := range self.accounts {
		if !acc.IsOnline() {
			self.leaveRoom(acc.AccountId)
			continue
		}
	}
	// 组装消息
	wait, err := proto.Marshal(&protomsg.Status_Wait{
		//todo .....................................................
	})
	if err != nil {
		log.Panicf("错误:%v ", err.Error())
	}

	betval := self.areaBetVal(true)
	self.enterMsg = &protomsg.StatusMsgLHD{
		Status:           protomsg.LHDGAMESTATUS(self.s),
		Status_StartTime: uint64(self.start_timestamp),
		Status_EndTime:   uint64(self.end_timestamp),
		AreaBetVal:       betval,
		AreaBetVal_Own:   map[int32]int64{},
		Status_Data:      wait,
	}
	self.SendBroadcast(protomsg.LHDMSG_SC_SWITCH_GAME_STATUS_BROADCAST_LHD.UInt16(), &protomsg.SWITCH_GAME_STATUS_BROADCAST_LHD{
		NextStatus: self.enterMsg,
	})
}

func (self *waitting) Tick(now int64) {
	if now >= self.end_timestamp {
		self.switchStatus(now, ERoomStatus_START_BETTING)
		return
	}

	if self.Close {
		for _, acc := range self.accounts {
			self.leaveRoom(acc.AccountId)
		}
		roomid := self.roomId
		core.LocalCoreSend(self.owner.Id, common.EActorType_MAIN.Int32(), func() {
			delete(RoomMgr.Rooms, roomid)
		})
		self.owner.Suspend()
		log.Infof("房间关闭完成")
	}
}

func (self *waitting) leave(accid uint32) bool {
	return true
}

func (self *waitting) enterData(accountId uint32) *protomsg.StatusMsgLHD {
	return self.enterMsg
}

func (self *waitting) Leave(now int64) {
	self.log(colorized.Blue("waitting leave\n"))
	self.log(colorized.Blue(""))
}

func (self *waitting) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	default:
		log.Warnf("waitting 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}

	return true
}
