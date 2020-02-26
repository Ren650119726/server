package room

import (
	"github.com/golang/protobuf/proto"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/game_red2black/algorithm"
)

type (
	waitting struct {
		*Room
		s             ERoomStatus
		start_timestamp int64
		end_timestamp int64
		enterMgr *protomsg.StatusMsg
	}
)

func (self *waitting) Enter(now int64) {
	duration := self.status_duration[self.s]
	self.start_timestamp = utils.MilliSecondTimeSince1970()
	self.end_timestamp = self.start_timestamp + duration
	log.Debugf(colorized.Blue("waitting enter duration:%v"), duration)

	self.betPlayers = make(map[uint32]map[protomsg.RED2BLACKAREA]int64)	// 清理押注过的玩家
	// 踢出下线的玩家
	for _, acc := range self.accounts {
		if !acc.IsOnline(){
			self.leaveRoom(acc.AccountId)
			continue
		}
	}

	// 随机获得6张牌
	self.GameCards = algorithm.GetRandom_Card(self.RoomCards,6)
	log.Infof("房间等待开始显示:%v 张 本局牌:%+v ",self.showNum,self.GameCards)

	// 组装消息
	bet,err := proto.Marshal(&protomsg.Status_Wait{
		//todo .....................................................
	})
	if err != nil {
		log.Panicf("错误:%v ",err.Error())
	}

	self.enterMgr = &protomsg.StatusMsg{
		Status:           protomsg.RED2BLACKGAMESTATUS(self.s),
		Status_StartTime: uint64(self.start_timestamp),
		Status_EndTime:   uint64(self.end_timestamp),
		RedCards:self.GameCards[0:self.showNum],
		BlackCards:self.GameCards[3:3+self.showNum],
		AreaBetVal:self.areaBetVal(true),
		Status_Data:      bet,
	}
	self.SendBroadcast(protomsg.RED2BLACKMSG_SC_SWITCH_GAME_STATUS_BROADCAST.UInt16(),&protomsg.SWITCH_GAME_STATUS_BROADCAST{
		NextStatus:self.enterMgr,
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
		self.owner.Suspend()
		log.Infof("房间关闭完成")
	}
}

func (self *waitting) leave(accid uint32) bool {
	return true
}

func (self *waitting) enterData(accountId uint32) *protomsg.StatusMsg {
	return self.enterMgr
}

func (self *waitting) Leave(now int64) {
	log.Debugf(colorized.Blue("waitting leave\n"))
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