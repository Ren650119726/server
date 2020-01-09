package main

import (
	"root/common"
	"root/core"
	"root/core/db"
	"root/server/baseprocess/logic"
)

func main() {
	// 创建server
	process := logic.NewProcess()
	msgchan := make(chan core.IMessage, 10000)
	actor := core.NewActor(common.EActorType_MAIN.Int32(), process, msgchan)
	core.CoreRegisteActor(actor)

	core.CoreRegisteActor(core.NewActor(common.EActorType_REDIS.Int32(), db.NewRedis(), make(chan core.IMessage, 10000)))
	// 启动
	core.CoreStart()
}
