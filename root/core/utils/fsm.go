package utils

import (
	"errors"
	"fmt"
)

// 状态函数
type StateHandler interface {
	Enter(sn int64)
	Tick(sn int64)
	Leave(sn int64)
	Handle(actor int32, msg []byte, session int64) bool
}

// 状态机
type FSM struct {
	curr_state int32                  // 当前状态
	states     map[int32]StateHandler // 状态处理对象
}

// 创建一个状态机
func NewFSM() *FSM {
	fsm := &FSM{}
	fsm.curr_state = -1 // 此状态无效
	fsm.states = make(map[int32]StateHandler, 2)
	return fsm
}

// 新增一个状态
func (self *FSM) Add(state int32, handler StateHandler) error {
	if _, exit := self.states[state]; exit {
		return errors.New(fmt.Sprintf("重复的状态:%v", state))
	}
	self.states[state] = handler
	return nil
}

// 状态迁移
func (self *FSM) Swtich(tp int64, new_state int32) error {
	if _, exit := self.states[new_state]; !exit {
		return errors.New(fmt.Sprintf("找不到状态:%v", new_state))
	}

	if self.State() != -1 {
		// 当前状态是否有效
		self.states[self.State()].Leave(tp)
	}

	self.curr_state = new_state

	self.states[self.State()].Enter(tp)

	return nil
}

// update
func (self *FSM) Update(now int64) error {
	if _, exit := self.states[self.State()]; !exit {
		return errors.New(fmt.Sprintf("找不到状态:%v", self.State()))
	}

	self.states[self.State()].Tick(now)
	return nil
}

func (self *FSM) State() int32 {
	return self.curr_state
}

func (self *FSM) Current() StateHandler {
	return self.states[self.curr_state]
}

func (self *FSM) Handle(actor int32, msg []byte, session int64) bool {
	return self.states[self.curr_state].Handle(actor, msg, session)
}
