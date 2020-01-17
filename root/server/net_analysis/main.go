package main

import (
	"root/common"
	"root/core"
	"root/core/utils"
	"fmt"
)

func main() {
	fmt.Println(utils.GetTimeFormatString(utils.SecondTimeSince1970()))
	// 创建server
	lo := NewLogic()
	msgchan := make(chan core.IMessage, 10000)
	actor := core.NewActor(common.EActorType_MAIN.Int32(), lo, msgchan)
	core.CoreRegisteActor(actor)

	core.CoreStart()
}
