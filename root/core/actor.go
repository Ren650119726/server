package core

import (
	"root/common"
	"root/core/log"
	"root/core/utils"
	"errors"
	"time"
)

type (
	/* acotr 处理接口 */
	IFHandler interface {
		Init(actor *Actor) bool
		Stop()
		HandleMessage(actor int32, msg []byte, session int64) bool
	}

	// Messge接口
	IMessage interface{}
)

/* 内部消息结构 */
type (
	CoreMessage struct {
		Source  int32
		Target  int32
		Session int64
		Data    []byte
	}

	// 主要用于本地Actor方便处理多线程逻辑
	LocalMessage struct {
		Source     int32
		Target     int32
		FunHandler func()
	}

	// 运行单元
	Actor struct {
		Id int32
		// 消息处理
		Handler      IFHandler
		MessageCache chan IMessage
		TimerMgr     *utils.TimerMgr
		IsSuspend    bool
	}
)

// 新创建一个actor
func NewActor(id int32, handler IFHandler, msgchan chan IMessage) *Actor {
	actor := &Actor{Id: id, Handler: handler, MessageCache: msgchan}
	actor.TimerMgr = utils.NewTimerMgr(100)
	actor.Resume()
	return actor
}

// 获取actorID
func (self *Actor) GetID() int32 {
	return self.Id
}

// Push一个消息
func (self *Actor) Push(msg IMessage) {
	if msg == nil {
		return
	}
	l := len(self.MessageCache)
	c := cap(self.MessageCache)
	if l > c*2/3 {
		log.Warnf("警告！队列消息超过三分之二了，actor:%v len:%v cap:%v", self.Id, l, c)
	}
	self.MessageCache <- msg
}

// 增加一个timer(interval 单位毫秒)
func (self *Actor) AddTimer(interval int64, trigger_times int32, callback utils.FuncTimeOut) int64 {
	now := utils.MilliSecondTime()
	timer := utils.NewTimer(now, now+interval, interval, trigger_times, callback)
	return self.TimerMgr.AddTimer(timer)
}

// 增加一个timer 固定某个日期更新(interval 单位毫秒) "2006-01-02 15:04:05"
func (self *Actor) AddScheduleTimer(date string, callback utils.FuncTimeOut) (timerId int64, err error) {
	time := utils.String2UnixStamp(date)
	now := utils.MilliSecondTime()
	if time < now {
		return -1, errors.New("传入的时间，已经比当前时间小")
	}
	timer := utils.NewTimer(now, time, time-now, 1, callback)
	return self.TimerMgr.AddTimer(timer), nil
}

// 增加一个timer 固定每日某个时刻更新(interval 单位毫秒) "15:04:05"
func (self *Actor) AddEverydayTimer(date string, callback utils.FuncTimeOut) (timerId int64, err error) {
	var timer *utils.MyTimer
	now := utils.MilliSecondTime()
	timeStamp := utils.GetStamp(date)
	// 这个的固定间隔时间为24小时
	timer = utils.NewTimer(now, timeStamp, int64(24*time.Hour/time.Millisecond), -1, callback)
	return self.TimerMgr.AddTimer(timer), nil
}

func (self *Actor) CancelTimer(timeid int64) {
	self.TimerMgr.CancelTimer(timeid)
}

// 逻辑层主动退出
func (self *Actor) Suspend() {
	self.IsSuspend = true
	self.TimerMgr.Reset()
}

// 唤醒
func (self *Actor) Resume() {
	self.IsSuspend = false
}

func GetRemoteIP(nSessionID int64) string {
	actor := GetActor(common.EActorType_SERVER.Int32())
	if actor == nil {
		return "-1.-1.-1.-1"
	}

	type IServer interface {
		GetSessionIP(sesseionId int64) string
	}
	tcpserver, b := actor.Handler.(IServer)
	if b == false {
		return "0.0.0.0"
	}

	ip := tcpserver.GetSessionIP(nSessionID)
	return ip
}
