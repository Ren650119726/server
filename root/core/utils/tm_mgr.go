package utils

import (
	"sort"
)

type FuncTimeOut func(dt int64)

type MyTimer struct {
	timeid           int64
	interval         int64 // 时间间隔(单位毫秒)
	next_triggertime int64 // 下一次执行的时间点(单位毫秒)
	last_triggertime int64
	trigger_times    int32       // 执行次数
	func_timeout     FuncTimeOut // 超时回调
	needdel          bool        // 需要删除
}

type TimerMgr struct {
	timers     []*MyTimer
	id2timer   map[int64]*MyTimer
	pending    []*MyTimer
	limit      int32
	cur_timeid int64
}

// 创建一个timer管理器
func NewTimerMgr(limit int32) *TimerMgr {
	timer_mgr := &TimerMgr{}
	timer_mgr.limit = limit
	timer_mgr.cur_timeid = 0
	timer_mgr.id2timer = make(map[int64]*MyTimer)
	return timer_mgr
}

// 创建一个timer
func NewTimer(now, next_triggertime, interval int64, trigger_times int32, callback FuncTimeOut) *MyTimer {
	return &MyTimer{
		interval:         interval,
		next_triggertime: next_triggertime,
		last_triggertime: now,
		trigger_times:    trigger_times,
		func_timeout:     callback,
		needdel:          false,
	}
}

func (self *TimerMgr) Reset() {
	self.timers = []*MyTimer{}
	self.pending = []*MyTimer{}
	self.id2timer = make(map[int64]*MyTimer)
	self.cur_timeid = 0
}
func (self *TimerMgr) Len() int {
	return len(self.timers)
}
func (self *TimerMgr) Swap(i, j int) {
	self.timers[i], self.timers[j] = self.timers[j], self.timers[i]
}
func (self *TimerMgr) Less(i, j int) bool {
	return self.timers[i].next_triggertime < self.timers[j].next_triggertime
}

// 增加一个timer
func (self *TimerMgr) AddTimer(timer *MyTimer) int64 {
	cnt := len(self.timers)
	if int32(cnt+1) > self.limit {
		return -1
	}

	// 先加入到pending列表，统一update处理
	self.pending = append(self.pending, timer)
	self.cur_timeid += 1
	timer.timeid = self.cur_timeid
	return timer.timeid
}

func (self *TimerMgr) CancelTimer(timeid int64) {
	if timer, ok := self.id2timer[timeid]; ok {
		timer.needdel = true       // 这里只是记录标记，统一处理
		timer.next_triggertime = 0 // 目前就是排序时靠前
	}
}

// 获取遍历
func (self *TimerMgr) Update(now int64) {
	// 先处理pending状态的timer
	for _, timer := range self.pending {
		self.timers = append(self.timers, timer)
		self.id2timer[timer.timeid] = timer
	}

	needsort := false
	if len(self.pending) != 0 {
		self.pending = self.pending[:0]
		needsort = true
		// sort.Sort(self) // TODO:这里需要排序么？先不排了吧 不排序的话就会延后一帧执行，需要的话能避免刚刚加入的timer能得到当次的有效执行
	}

	// 处理定时器
	cnt := len(self.timers)
	if cnt == 0 {
		return
	}

	for idx := 0; idx < len(self.timers); {
		timer := self.timers[idx]
		if timer.needdel { // 需要删除的就干掉吧
			self.timers = append(self.timers[:idx], self.timers[idx+1:]...)
			delete(self.id2timer, timer.timeid)
			continue
		}

		// 检查执行时间是否到了
		dt := now - timer.next_triggertime
		if dt < 0 {
			break
		}

		// 执行timer回调
		Try(func() { timer.func_timeout(now - timer.last_triggertime) })

		if timer.trigger_times > 0 {
			timer.trigger_times--
		}
		if timer.trigger_times == 0 {
			// 删除timer
			self.timers = append(self.timers[:idx], self.timers[idx+1:]...)
			delete(self.id2timer, timer.timeid)
			continue
		} else {
			idx++
			needsort = true
			timer.last_triggertime = now
			timer.next_triggertime += timer.interval - (dt % timer.interval) // 下一次到期时间扣除这次多跑的时差
		}
	}

	// 定时器排序序列发生变化时就需要排序了
	if needsort {
		sort.Sort(self)
	}
}
