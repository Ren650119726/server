package main

import (
	"root/common"
	"root/core"
	"root/server/db/server"
)

// 主函数入口
func main() {
	// dcServer逻辑处理actor
	dcServer := server.NewDBServer()
	actor := core.NewActor(common.EActorType_MAIN.Int32(), dcServer, make(chan core.IMessage, 10000))
	core.CoreRegisteActor(actor)

	// 启动core
	core.CoreStart()
}
