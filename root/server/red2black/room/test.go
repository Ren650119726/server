package room

import (
	"root/common"
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/packet"
	"root/protomsg"
)

func init() {
	core.Cmd.Regist("tohall", tohall, false)
	core.Cmd.Regist("todb", todb, false)
	core.Cmd.Regist("reload", reload, false)
	core.Cmd.Regist("info", info, false)
	core.Cmd.Regist("closegame", Close, true)

}

func tohall(s []string) {
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_SS_TEST_NETWORK.UInt16())
	core.CoreSend(0, common.EActorType_CONNECT_HALL.Int32(), send.GetData(), 0)
}

func todb(s []string) {
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_SS_TEST_NETWORK.UInt16())

	core.CoreSend(0, common.EActorType_CONNECT_DB.Int32(), send.GetData(), 0)
}

func reload(s []string) {
	for roomID := range RoomMgr.roomActorId {
		core.LocalCoreSend(0, int32(roomID), func() {
			config.LoadPublic_Conf()
		})
	}
}

func info(s []string) {
	if RoomMgr.Global_room != nil {
		core.LocalCoreSend(0, int32(RoomMgr.Global_room.roomId), func() {
			max := config.GetPublicConfig_Int64("R2B_CEILING_LINE")
			min := config.GetPublicConfig_Int64("R2B_FLOOR_LINE")
			log.Infof("roomID:%v 总人数:%v 机器人数:%v 水位线:【floor:%v  current:%v  ceiling:%v】 param:%v",
				RoomMgr.Global_room.roomId, RoomMgr.Global_room.count(), RoomMgr.Global_room.RobotCount(), min, RoomMgr.Water_line, max, RoomMgr.Global_room.param)
			for accid, acc := range RoomMgr.Global_room.accounts {
				if acc.Robot == 0 {
					log.Infof("accid：%v name:%v money:%v master:%v", accid, acc.Name, acc.GetMoney(), RoomMgr.Global_room.SeatMasterIndex(acc.AccountId))
				}
			}
		})
	}

}
func Close(s []string) {
	RoomMgr.Global_room.Quit = true
	log.Infof("正在关闭房间.......")
}
