package main

import (
	"root/common"
	"root/core"
	"root/server/game_red2black/logic"
)

func main() {
	// 创建server
	r2b := logic.NewRed2Black()
	msgchan := make(chan core.IMessage, 10000)
	actor := core.NewActor(common.EActorType_MAIN.Int32(), r2b, msgchan)
	core.CoreRegisteActor(actor)

	core.CoreStart()
}
