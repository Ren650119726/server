package speaker

import (
	"root/core/log"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/hall/account"
)

var SpeakerMgr = newSpeakerMgr()

type (
	// 小喇叭
	speaker struct {
		NextTime     int64  // 下次执行时间
		DelTime      int64  // 删除时间
		IntervalTime uint16 // 间隔时间
		Type         uint8  // 小喇叭广播类型 (1只在大厅  2大厅和所有游戏)
		Content      string // 小喇叭内容
	}

	speakerMgr struct {
		speakerSlice []*speaker
	}
)

func newSpeakerMgr() *speakerMgr {
	return &speakerMgr{
		speakerSlice: []*speaker{},
	}
}

func (self *speakerMgr) AddSpeaker(nNextTime int64, nDelTime int64, nIntervalTime uint16, nType uint8, strContent string) {
	tSpeaker := &speaker{
		NextTime:     nNextTime,
		Content:      strContent,
		DelTime:      nDelTime,
		IntervalTime: nIntervalTime,
		Type:         nType,
	}
	self.speakerSlice = append(self.speakerSlice, tSpeaker)
	log.Infof("添加后台小喇叭, StartTime:%v, EndTime:%v, IntervalTime:%v秒, Type:%v, 内容:%v", utils.GetTimeFormatString(tSpeaker.NextTime), utils.GetTimeFormatString(tSpeaker.DelTime), tSpeaker.IntervalTime, tSpeaker.Type, tSpeaker.Content)
}

func (self *speakerMgr) PrintAll() {
	for nID, tNode := range self.speakerSlice {
		log.Infof("小喇叭ID:%v, StartTime:%v, EndTime:%v, IntervalTime:%v秒, Type:%v, 内容:%v", nID, utils.GetTimeFormatString(tNode.NextTime), utils.GetTimeFormatString(tNode.DelTime), tNode.IntervalTime, tNode.Type, tNode.Content)
	}
}

func (self *speakerMgr) RemoveSpeaker(iIndex int) {

	// 删除所有小喇叭
	if iIndex < 0 {
		self.speakerSlice = []*speaker{}
		return
	}

	// 删除指定下标的小喇叭
	nLastIndex := len(self.speakerSlice)
	if iIndex >= nLastIndex {
		return
	}

	nLastIndex--
	self.speakerSlice[iIndex], self.speakerSlice[nLastIndex] = self.speakerSlice[nLastIndex], self.speakerSlice[iIndex]
	self.speakerSlice = self.speakerSlice[:nLastIndex]
}

// nType值1表示只在大厅播放
// nType值2表示在大厅和所有游戏房间播放
func (self *speakerMgr) SendBroadcast(nType uint8, Content string) {
	tSend := packet.NewPacket(nil)
	tSend.SetMsgID(protomsg.Old_MSGID_SEND_SPEAKERS.UInt16())
	tSend.WriteUInt8(nType)
	tSend.WriteString(Content)
	account.AccountMgr.SendBroadcast(tSend, nType)
}

func (self *speakerMgr) Update() {
	nNowTime := utils.SecondTimeSince1970()
	nLastIndex := len(self.speakerSlice)
	for i := nLastIndex - 1; i >= 0; i-- {
		tNode := self.speakerSlice[i]
		if nNowTime >= int64(tNode.NextTime) {
			tNode.NextTime += int64(tNode.IntervalTime)
			self.SendBroadcast(uint8(tNode.Type), tNode.Content)
		}

		if nNowTime >= int64(tNode.DelTime) {
			if len(self.speakerSlice) == 1 {
				self.speakerSlice = []*speaker{}
				return
			} else {
				nLastIndex--
				self.speakerSlice[i], self.speakerSlice[nLastIndex] = self.speakerSlice[nLastIndex], self.speakerSlice[i]
				self.speakerSlice = self.speakerSlice[:nLastIndex]
			}
		}
	}
}
