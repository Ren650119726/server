package speaker

import (
	"root/core/log"
	"root/core/utils"
	"root/protomsg"
	"root/server/hall/account"
	"root/server/hall/send_tools"
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

// t值0表示全服播放
// t值1表示只在大厅播放
// t值2表示只在游戏播放
func (self *speakerMgr) SendBroadcast(t uint8, Content string) {
	peoples := account.AccountMgr.GetAllAccount()
	for _,acc := range peoples{
		if( t == 1 && acc.GetRoomID() != 0) || (t == 2 && acc.GetRoomID() == 0){
			continue
		}
		send_tools.Send2Account(protomsg.MSG_SC_BROADCAST_MSG.UInt16(),&protomsg.BROADCAST_MSG{Content:Content},acc.SessionId)
	}
}

func (self *speakerMgr) Update() {
	nNowTime := utils.SecondTimeSince1970()
	nLastIndex := len(self.speakerSlice)
	for i := nLastIndex - 1; i >= 0; i-- {
		tNode := self.speakerSlice[i]
		if nNowTime >= tNode.NextTime{
			tNode.NextTime += int64(tNode.IntervalTime)
			self.SendBroadcast(tNode.Type, tNode.Content)
		}

		if nNowTime >= tNode.DelTime{
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
