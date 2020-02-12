package main

import (
	"root/common"
	"root/core"
	"root/core/db"
	"root/server/game_fruitMary/logic"
)

func main() {
	// 创建server
	hb := logic.NewFruitMary()
	msgchan := make(chan core.IMessage, 10000)
	actor := core.NewActor(common.EActorType_MAIN.Int32(), hb, msgchan)
	core.CoreRegisteActor(actor)

	core.CoreRegisteActor(core.NewActor(common.EActorType_REDIS.Int32(), db.NewRedis(), make(chan core.IMessage, 10000)))
	core.CoreStart()
}
