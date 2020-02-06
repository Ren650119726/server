package main

import (
	"root/common"
	"root/core"
	"root/core/db"
	"root/core/log"
	"root/server/fruitMary/logic"
)

func main() {
	ii := []int{1,2,3,4,5}
	i := ii[0:1]
	log.Info(i)
	// 创建server
	hb := logic.NewFruitMary()
	msgchan := make(chan core.IMessage, 10000)
	actor := core.NewActor(common.EActorType_MAIN.Int32(), hb, msgchan)
	core.CoreRegisteActor(actor)

	core.CoreRegisteActor(core.NewActor(common.EActorType_REDIS.Int32(), db.NewRedis(), make(chan core.IMessage, 10000)))
	core.CoreStart()
}
