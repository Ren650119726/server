package room

import (
	"root/common"
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"fmt"
	"root/protomsg"
	"root/server/dehgame/account"
	"root/server/dehgame/send_tools"
	"root/server/dehgame/types"
	"strconv"
)

func init() {
	core.Cmd.Regist("room", createRoom, false)
	core.Cmd.Regist("join", joinPlayer, false)
	core.Cmd.Regist("info", infoRoom, false)
	core.Cmd.Regist("list", roomList, false)
	core.Cmd.Regist("closegame", CloseServer, false)

	core.Cmd.Regist("sit", sit, false)
	core.Cmd.Regist("ready", ready, false)
	core.Cmd.Regist("bobo", bobo, false)
	core.Cmd.Regist("hanhua", hanhua, false)
	core.Cmd.Regist("tohall", tohall, false)
	core.Cmd.Regist("todb", todb, false)
	core.Cmd.Regist("setbonus", setbouns, false)
	core.Cmd.Regist("reload", reload, false)
	core.Cmd.Regist("on", online, false)

	//cards1 := []common.Card_info{
	//	{common.ECardType_HONGTAO.UInt8(), 7},
	//	{common.ECardType_FANGKUAI.UInt8(), 7},
	//	{common.ECardType_JKEOR.UInt8(), 6},
	//	{common.ECardType_MEIHUA.UInt8(), 7},
	//}
	//cards2 := []common.Card_info{
	//	{common.ECardType_FANGKUAI.UInt8(), 12},
	//	{common.ECardType_MEIHUA.UInt8(), 8},
	//	{common.ECardType_HEITAO.UInt8(), 7},
	//	{common.ECardType_HONGTAO.UInt8(), 8},
	//}
	//_, taili, _ := algorithm.CalcOnePlayerCardType(cards1, 0, true)
	//_, tailj, _ := algorithm.CalcOnePlayerCardType(cards2, 0, true)
	//
	//ret := algorithm.CompareTouWei(
	//	cards1,
	//	0,
	//	cards2,
	//	0,
	//	0,
	//	true)
	//
	//print(ret)
}

// 创建房间
func createRoom([]string) {
	p := packet.NewPacket(nil)
	p.SetMsgID(protomsg.Old_MSGID_CREATE_ROOM.UInt16())
	p.WriteUInt32(0)
	p.WriteUInt32(999)
	p.WriteUInt8(1)
	p.WriteUInt8(2)
	p.WriteString("1|0|100|200|5000|400|0")
	p.WriteUInt8(1)
	p.WriteUInt32(0)
	core.CoreSend(0, common.EActorType_MAIN.Int32(), p.GetData(), 0)
}

// 房间信息
func infoRoom(s []string) {
	roomid, _ := strconv.Atoi(s[0])
	Room := RoomMgr.Room(uint32(roomid))
	if Room == nil {
		log.Warnf("找不到房间%v", roomid)
	} else {
		log.Infof(colorized.White("房间状态:[%v] 当前庄家:[%v] 房间局数:[%v] 休芒次数:[%v] param:[%v]"), types.ERoomStatus(Room.status.State()).String(), Room.lastBanker_index, Room.games, Room.mangoCount, Room.param)
		for index, player := range Room.seats {
			str := "nil"
			if player != nil {
				str = fmt.Sprintf("Accid:%v 名字:%v 状态:%v 身上金额:%v 簸簸:%v 押注:%v 押芒果:%v 游戏局数:%v",
					player.acc.AccountId,
					player.acc.Name,
					player.status.String(),
					player.acc.GetMoney(),
					player.bobo,
					player.bet,
					player.mangoVal,
					player.acc.Games)
			}
			log.Infof("座位号:[%v] 玩家信息:(%v)", index, str)
		}
	}
}

// 房间信息
func roomList(s []string) {
	for _, room := range RoomMgr.roomActor {
		if room != nil {
			log.Infof("rooms:%v 坐下的人:%v 总人数:%v 当前服务费总额:%v", room.roomId, room.sitDownCount(), len(room.accounts), RoomMgr.Fee)
		}

	}

}

// 加入玩家房间
func joinPlayer(s []string) {
	accid, _ := strconv.Atoi(s[0])
	roomid, _ := strconv.Atoi(s[1])
	data := &protomsg.AccountStorageData{
		AccountId: uint32(accid),
		RMB:       10000,
	}
	account.AccountMgr.RecvAccount(data, &protomsg.AccountGameData{RoomID: uint32(roomid)})

	j := packet.NewPacket(nil)
	j.SetMsgID(protomsg.Old_MSGID_ENTER_GAME.UInt16())
	j.WriteUInt32(uint32(accid))
	j.WriteUInt32(uint32(roomid))
	core.CoreSend(0, common.EActorType_MAIN.Int32(), j.GetData(), 0)

	j2 := packet.NewPacket(nil)
	j2.SetMsgID(protomsg.Old_MSGID_CX_SIT_DOWN.UInt16())
	j2.WriteUInt32(uint32(accid))
	core.CoreSend(0, common.EActorType_MAIN.Int32(), j2.GetData(), 0)

	j3 := packet.NewPacket(nil)
	j3.SetMsgID(protomsg.Old_MSGID_CX_READY.UInt16())
	j3.WriteUInt32(uint32(accid))
	core.CoreSend(0, common.EActorType_MAIN.Int32(), j3.GetData(), 0)
}

// 加入玩家房间
func sit(s []string) {
	accid, _ := strconv.Atoi(s[0])
	j := packet.NewPacket(nil)
	j.SetMsgID(protomsg.Old_MSGID_CX_SIT_DOWN.UInt16())
	j.WriteUInt32(uint32(accid))
	core.CoreSend(0, common.EActorType_MAIN.Int32(), j.GetData(), 0)
}

// 准备
func ready(s []string) {

	accid, _ := strconv.Atoi(s[0])
	j := packet.NewPacket(nil)
	j.SetMsgID(protomsg.Old_MSGID_CX_READY.UInt16())
	j.WriteUInt32(uint32(accid))
	core.CoreSend(0, common.EActorType_MAIN.Int32(), j.GetData(), 0)
}

// 准备
func bobo(s []string) {
	accid, _ := strconv.Atoi(s[0])
	bobo, _ := strconv.Atoi(s[1])
	j := packet.NewPacket(nil)
	j.SetMsgID(protomsg.Old_MSGID_CX_SET_BOBO.UInt16())
	j.WriteUInt32(uint32(accid))
	j.WriteUInt32(uint32(bobo))
	core.CoreSend(0, common.EActorType_MAIN.Int32(), j.GetData(), 0)
}

// 喊话
func hanhua(s []string) {
	accid, _ := strconv.Atoi(s[0])
	hanhua, _ := strconv.Atoi(s[1])
	val, _ := strconv.Atoi(s[2])
	j := packet.NewPacket(nil)
	j.WriteUInt32(uint32(accid))
	switch hanhua {
	case 1:
		j.SetMsgID(protomsg.Old_MSGID_CX_HANHUA_XIU.UInt16())
	case 2:
		j.SetMsgID(protomsg.Old_MSGID_CX_HANHUA_DIU.UInt16())
	case 3:
		j.SetMsgID(protomsg.Old_MSGID_CX_HANHUA_DA.UInt16())
		j.WriteInt64(int64(val))
	case 4:
		j.SetMsgID(protomsg.Old_MSGID_CX_HANHUA_QIAO.UInt16())
	case 5:
		j.SetMsgID(protomsg.Old_MSGID_CX_HANHUA_GEN.UInt16())
		j.WriteInt64(int64(val))
	}
	core.CoreSend(0, common.EActorType_MAIN.Int32(), j.GetData(), 0)
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

func CloseServer(s []string) {

	for _, room := range RoomMgr.roomActor {
		room.Close()
	}
	RoomMgr.IsMaintenance = true

	nServerID, _ := strconv.Atoi(core.Appname)
	tSend := packet.NewPacket(nil)
	tSend.SetMsgID(protomsg.MSGID_GH_SET_MAINTENANCE.UInt16())
	tSend.WriteUInt16(uint16(nServerID))
	tSend.WriteUInt8(1) // 设置进入维护状态
	send_tools.Send2Hall(tSend.GetData())

	a := core.GetActor(int32(common.EActorType_SERVER))
	if a != nil {
		a.Suspend()
	}

}

func setbouns(s []string) {
	matchType, _ := strconv.Atoi(s[0])
	addition, _ := strconv.Atoi(s[1])
	RoomMgr.Set_bonus(uint32(matchType), uint64(addition))
}

func reload(s []string) {
	config.LoadPublic_Conf()
}

func online(s []string) {
	RoomMgr.OnlineStatics()
}
