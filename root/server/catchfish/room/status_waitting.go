package room

import (
	"root/common/config"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
)

type (
	waitting struct {
		*Room
		s         ERoomStatus
		timestamp int64

		normal_deal []int64 // 普通红包
		bomb_deal   []int64 // 踩雷红包
		senddata    bool

		conf_floor_line   int64
		conf_ceiling_line int64
		conf_enforce      int64
	}
)

func (self *waitting) Enter(now int64) {
	self.conf_floor_line = config.GetPublicConfig_Int64("HB_FLOOR_LINE")
}

func (self *waitting) Tick(now int64) {
	curTime := utils.MilliSecondTimeSince1970()
	if curTime >= self.timestamp {
		if !self.senddata {
			self.settlement()
		}

		// 切换到等待状态
		self.switchStatus(0, ERoomStatus_GAME)
		return
	}
}

// 结算
func (self *waitting) settlement() {

}
func (self *waitting) Leave(now int64) {

	log.Debugf(colorized.Blue("waitting leave\n"))
}

func (self *waitting) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	//case protomsg.Old_MSGID_ENTER_GAME.UInt16(): // 客户端链接进入游戏
	//	self.Old_MSGID_ENTER_GAME(actor, msg, session)
	//case protomsg.Old_MSGID_HONGBAO_ROB_HONGBAO.UInt16(): // 抢红包
	//	self.Old_MSGID_HONGBAO_ROB_HONGBAO(actor, msg, session)
	default:
		log.Warnf("waitting 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}

	return true
}

// 进入游戏
func (self *waitting) Old_MSGID_ENTER_GAME(actor int32, msg []byte, session int64) {

}

// 抢红包操作
func (self *waitting) Old_MSGID_HONGBAO_ROB_HONGBAO(actor int32, msg []byte, session int64) {
}
