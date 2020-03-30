package main

import (
	"root/common"
	"root/core"
)

func main()  {
	G := NewLogic()
	msgchan := make(chan core.IMessage, 10000)
	G_ACTOR := core.NewActor(common.EActorType_MAIN.Int32(), G, msgchan)
	core.CoreRegisteActor(G_ACTOR)

	core.CoreStart()
}


