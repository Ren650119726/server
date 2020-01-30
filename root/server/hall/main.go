package main

import (
	"encoding/binary"
	"root/common"
	"root/core"
	"root/core/db"
	"root/core/log"
	"root/core/packet"
	"root/server/hall/logic"
)

func main() {
	var buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, 500)
	log.Infof("%v ",buf)


	msg := packet.NewPacket(nil)
	msg.SetMsgID(500)
	msg.WriteInt8(99)
	log.Infof("%v ",msg.GetData()[:6])
	log.Infof("%v ",msg.GetDataSize())
	// 创建server
	hall := logic.NewHall()

	msgchan := make(chan core.IMessage, 10000)
	actor := core.NewActor(common.EActorType_MAIN.Int32(), hall, msgchan)
	core.CoreRegisteActor(actor)

	core.CoreRegisteActor(core.NewActor(common.EActorType_REDIS.Int32(), db.NewRedis(), make(chan core.IMessage, 10000)))
	core.CoreStart()
}
