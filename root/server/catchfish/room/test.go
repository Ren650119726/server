package room

import (
	"root/common"
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/packet"
	"root/protomsg/inner"
	"time"
)

func init() {
	core.Cmd.Regist("tohall", tohall, true)
	core.Cmd.Regist("todb", todb, true)
	core.Cmd.Regist("reload", reload, true)
	core.Cmd.Regist("info", info, true)
	core.Cmd.Regist("closegame", Close, true)

}

func tohall(s []string) {
	send := packet.NewPacket(nil)
	send.SetMsgID(inner.SERVERMSG_SS_TEST_NETWORK.UInt16())
	core.CoreSend(0, common.EActorType_CONNECT_HALL.Int32(), send.GetData(), 0)
}

func todb(s []string) {
	send := packet.NewPacket(nil)
	send.SetMsgID(inner.SERVERMSG_SS_TEST_NETWORK.UInt16())

	core.CoreSend(0, common.EActorType_CONNECT_DB.Int32(), send.GetData(), 0)
}

func reload(s []string) {
	config.LoadPublic_Conf()
}

func info(s []string) {
	log.Infof("当前水位线:%v", RoomMgr.Water_line)
	for actorId, room := range RoomMgr.rooms {
		r := room
		core.LocalCoreSend(0, actorId, func() {
			log.Infof("房间[%v] param:%v 总人数:%v 机器人:%v 剩余红包:%v ", r.roomId, r.param, r.count())
			for _, v := range r.accounts {
				if v.Robot == 0 {
					log.Infof("房间[%v]  玩家:%v name:%v 身上的钱:%v ", r.roomId, v.AccountId, v.Name, v.GetMoney())
				}
			}
			log.Infof("")
		})
		time.Sleep(200 * time.Millisecond)
	}
}
func Close(s []string) {
	for _, room := range RoomMgr.rooms {
		room.close()
	}

	a := core.GetActor(int32(common.EActorType_SERVER))
	if a != nil {
		a.Suspend()
	}
}
