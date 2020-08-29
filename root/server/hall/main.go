package main

import (
	_ "net/http/pprof"
	"root/common"
	"root/core"
	"root/core/db"
	"root/server/hall/logic"
)

func main() {
	// 创建server
	hall := logic.NewHall()

	msgchan := make(chan core.IMessage, 10000)
	actor := core.NewActor(common.EActorType_MAIN.Int32(), hall, msgchan)
	core.CoreRegisteActor(actor)

	core.CoreRegisteActor(core.NewActor(common.EActorType_REDIS.Int32(), db.NewRedis(), make(chan core.IMessage, 10000)))
	core.CoreStart()
}
