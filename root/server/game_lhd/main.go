package main

import (
	"root/common"
	"root/core"
	"root/server/game_lhd/logic"
)

func main() {
	// 创建server
	lhd := logic.NewLHD()
	msgchan := make(chan core.IMessage, 10000)
	actor := core.NewActor(common.EActorType_MAIN.Int32(), lhd, msgchan)
	core.CoreRegisteActor(actor)

	core.CoreStart()
}
