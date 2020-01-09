package room

import (
	"root/common"
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/packet"
	"root/protomsg"
	"root/server/paodekuai/account"
	"strconv"
)

func init() {
	core.Cmd.Regist("info", infoRoom, false)
	core.Cmd.Regist("list", roomList, false)
	core.Cmd.Regist("closegame", CloseServer, false)
	core.Cmd.Regist("setbonus", setbonus, false)

	core.Cmd.Regist("tohall", tohall, false)
	core.Cmd.Regist("reload", reload, false)
	core.Cmd.Regist("on", online, false)
	core.Cmd.Regist("off", offline, false)
	core.Cmd.Regist("remove", remove_player, false)
	core.Cmd.Regist("leave", leave_seat, false)
}

// 房间信息
func infoRoom(s []string) {
	if len(s) <= 0 {
		log.Warn("请输入正确的房间ID")
		return
	}

	roomid, _ := strconv.Atoi(s[0])
	room := RoomMgr.GetRoom(uint32(roomid))
	if room == nil {
		log.Warnf("找不到房间%v", roomid)
	} else {
		room.printRoom()
	}
}

func setbonus(s []string) {

	if len(s) <= 1 {
		log.Warn("请输入匹配类型和奖金池值")
		return
	}

	bet, _ := strconv.Atoi(s[0])
	addition, _ := strconv.Atoi(s[1])
	RoomMgr.SetBonusPool(uint32(bet), int64(addition), true)
}

// 房间信息
func roomList(s []string) {
	for nBet, nValue := range RoomMgr.Bonus {
		log.Infof("底注:%v 奖金池金额:%v", nBet, nValue)
	}

	log.Infof("历史最高：%+v", RoomMgr.Bonus_h)

	for _, room := range RoomMgr.RoomActor {
		if room != nil {
			room.printRoom()
		}
	}
}
func tohall(s []string) {
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_SS_TEST_NETWORK.UInt16())
	core.CoreSend(0, common.EActorType_CONNECT_HALL.Int32(), send.GetData(), 0)
}

func CloseServer(s []string) {
	for _, room := range RoomMgr.RoomActor {
		room.CloseRoom()
	}
	a := core.GetActor(int32(common.EActorType_SERVER))
	a.Suspend()
}

func reload(s []string) {
	config.LoadPublic_Conf()
}

func online(s []string) {
	RoomMgr.OnlineStatics(common.STATUS_ONLINE.UInt32())
}

func offline(s []string) {
	RoomMgr.OnlineStatics(common.STATUS_OFFLINE.UInt32())
}

func remove_player(s []string) {
	if len(s) <= 0 {
		log.Warn("请输入需要从房间删除的玩家ID")
		return
	}
	nAccountID, _ := strconv.Atoi(s[0])
	tAccount := account.AccountMgr.GetAccountByID(uint32(nAccountID))
	if tAccount == nil || tAccount.RoomID <= 0 {
		log.Warn("请输入需要从房间删除的玩家ID")
		return
	}
	tRoom := RoomMgr.GetRoom(tAccount.RoomID)
	if tRoom == nil {
		log.Warn("请输入需要从房间删除的玩家ID")
		return
	}
	core.LocalCoreSend(0, int32(tAccount.RoomID), func() {
		tRoom.leave_room(tAccount, false)
	})
}

func leave_seat(s []string) {

	if len(s) <= 0 {
		log.Warn("请输入需要离座的玩家ID")
		return
	}
	nAccountID, _ := strconv.Atoi(s[0])
	tAccount := account.AccountMgr.GetAccountByID(uint32(nAccountID))
	if tAccount == nil || tAccount.RoomID <= 0 {
		log.Warn("请输入需要正确的离座玩家ID")
		return
	}
	tRoom := RoomMgr.GetRoom(tAccount.RoomID)
	if tRoom == nil {
		log.Warn("请输入需要正确的离座玩家ID")
		return
	}
	core.LocalCoreSend(0, int32(tAccount.RoomID), func() {

		nIndex := tRoom.get_seat_index(tAccount.AccountId)
		if nIndex > tRoom.max_count {
			log.Warn("请输入需要正确的离座玩家ID")
			return
		} else {
			tRoom.leave_seat(tAccount, nIndex)
		}
	})
}
