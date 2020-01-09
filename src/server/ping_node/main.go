package main

import (
	"root/common"
	"root/core"
)

func main() {
	// 创建Main线程
	pdk := NewLogic()
	msgchan := make(chan core.IMessage, 10000)
	actor := core.NewActor(common.EActorType_MAIN.Int32(), pdk, msgchan)
	core.CoreRegisteActor(actor)
	core.CoreStart()
}
