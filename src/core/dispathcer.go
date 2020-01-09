package core

import (
	"root/core/log"
	"fmt"
	"runtime"
)

type (
	EventType byte
	Event     interface{}
	Listener  interface {
		OnEvent(Event, EventType)
	}

	Dispatcher struct {
		listeners map[EventType][]Listener
	}

	/*Dispatcher ibase {
		Dispatch(event Event, t EventType)
		AddEventListener(eventType EventType, listener Listener)
		RemoveEventListener(eventType EventType, listener Listener)
		RemoveListener(listener Listener)
	}*/
)

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		listeners: make(map[EventType][]Listener),
	}
}

type WrapEvent struct {
	Event
	Caller string
}

func (dispatcher *Dispatcher) Dispatch(event Event, t EventType) {
	list, ok := dispatcher.listeners[t]
	if !ok {
		return
	}

	if true {
		wrap := WrapEvent{Event: event}
		_, file, line, _ := runtime.Caller(1)
		wrap.Caller = fmt.Sprintf("%s:%d", file, line)
		event = wrap
	}
	for i := range list {
		list[i].OnEvent(event, t)
	}
}

func (dispatcher *Dispatcher) AddEventListener(eventType EventType, listener Listener) {
	_, ok := dispatcher.listeners[eventType]
	if !ok {
		dispatcher.listeners[eventType] = make([]Listener, 0)
	}

	list := dispatcher.listeners[eventType]
	for i := range list {
		if list[i] == listener {
			log.Errorf("CEventDispatcher：AddEventListener：重复添加事件。eventType=%v", eventType)
			return
		}
	}

	dispatcher.listeners[eventType] = append(list, listener)
}

func (dispatcher *Dispatcher) RemoveEventListener(eventType EventType, listener Listener) {
	list, ok := dispatcher.listeners[eventType]

	if !ok {
		log.Errorf("CEventDispatcher：RemoveEventListener：当前事件id不包含任何回调结构体。eventType=%v", eventType)
		return
	}
	for i := range list {
		if list[i] == listener {
			list = append(list[:i], list[i+1:]...)
			break
		}
	}
	dispatcher.listeners[eventType] = list
}

func (dispatcher *Dispatcher) RemoveListener(listener Listener) {
	for typ := range dispatcher.listeners {
		dispatcher.RemoveEventListener(typ, listener)
	}
}
