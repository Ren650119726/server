package main

import (
	"root/common"
	"root/core"
	"root/core/db"
	"root/server/hall/logic"
	"time"
)

func main() {
	ss := time.Now().Format("2006-01-02")
	println(ss)
	// 创建server
	hall := logic.NewHall()

	msgchan := make(chan core.IMessage, 10000)
	actor := core.NewActor(common.EActorType_MAIN.Int32(), hall, msgchan)
	core.CoreRegisteActor(actor)

	core.CoreRegisteActor(core.NewActor(common.EActorType_REDIS.Int32(), db.NewRedis(), make(chan core.IMessage, 10000)))
	core.CoreStart()
}
