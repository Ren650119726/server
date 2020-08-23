package utils

type FuncTimeOut func(dt int64)

type MyTimer struct {
	timeid           int64	// 计时器id
	interval         int64  // 时间间隔(单位毫秒)
	next_triggertime int64  // 下一次执行的时间戳(单位毫秒)
	last_triggertime int64
	trigger_times    int32       // 执行次数
	func_timeout     FuncTimeOut // 超时回调
	disabled         bool        // 失效
}

type TimerMgr struct {
	timers     *Heap // 小顶堆
	id2timer   map[int64]*MyTimer
	cur_timeid int64
}

// 创建一个timer管理器
func NewTimerMgr() *TimerMgr {
	timer_mgr := TimerMgr{}
	timer_mgr.cur_timeid = 0
	timer_mgr.id2timer = make(map[int64]*MyTimer)
	timer_mgr.timers = NewHeap(nil, 1)
	return &timer_mgr
}

// 创建一个timer
func NewTimer(now, next_triggertime, interval int64, trigger_times int32, callback FuncTimeOut) *MyTimer {
	return &MyTimer{
		interval:         interval,
		next_triggertime: next_triggertime,
		last_triggertime: now,
		trigger_times:    trigger_times,
		func_timeout:     callback,
		disabled:         false,
	}
}

func (self *TimerMgr) Reset() {
	self.timers = NewHeap(nil, 2)
	self.id2timer = make(map[int64]*MyTimer)
	self.cur_timeid = 0
}

// 增加一个timer
func (self *TimerMgr) AddTimer(timer *MyTimer) int64 {
	self.cur_timeid += 1
	timer.timeid = self.cur_timeid
	self.timers.Push(timer)
	self.id2timer[timer.timeid] = timer
	return timer.timeid
}

func (self *TimerMgr) CancelTimer(timeid int64) {
	if timer, ok := self.id2timer[timeid]; ok {
		timer.disabled = true // 这里只是记录标记，统一处理
	}
}

// 获取遍历
func (self *TimerMgr) Update(now int64) {
	for {
		intf := self.timers.Peek()
		if intf == nil {
			break
		}
		timer := intf.(*MyTimer)

		if !timer.disabled {
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

			// 还有次数,继续加入优先队列
			if timer.trigger_times != 0 {
				timer.last_triggertime = now
				timer.next_triggertime += timer.interval - (dt % timer.interval) // 下一次到期时间扣除这次多跑的时差
				self.timers.Push(timer)
			}
		}
		self.timers.Pop()
	}
}

func (self *MyTimer) Priority() int64 {
	return self.next_triggertime
}
