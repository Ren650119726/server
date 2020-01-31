package logic

import (
	"root/common"
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/utils"
	"root/protomsg"
	"root/server/hall/account"
	"root/server/hall/event"
	"root/server/hall/send_tools"
)

var GameMgr = newGameMgr()

type (
	nodeInfo struct {
		gameType uint32	// 游戏类型
		session  int64
	}

	roomInfo struct {
		roomID 		uint32 // 房间ID
		serverID 	uint32 // 服务器ID
		PlayerCount int    // 房间人数
	}

	gameMgr struct {
		nodes map[uint32]*nodeInfo // 游戏节点 key:sid
		rooms map[uint32]*roomInfo // 所有房间 key:roomID
	}
)

func newGameMgr() *gameMgr {
	hall := &gameMgr{
		nodes: make(map[uint32]*nodeInfo),
		rooms: make(map[uint32]*roomInfo),
	}
	event.Dispatcher.AddEventListener(event.EventType_UpdateCharge, hall)
	return hall
}

func (self *gameMgr)GameConnectHall(sid, gameType uint32, session int64)  {
	if game,e := self.nodes[sid];e{
		log.Infof("游戏:%v sid:%v session:%v 重新连接 新session:%v",common.EGameType(gameType).String(), sid, game.session, session)
		game.session = session
	}else{
		log.Infof("游戏:%v sid:%v 连接成功 session:%v",common.EGameType(gameType).String(), sid, session)
		self.nodes[sid] = &nodeInfo{
			gameType: gameType,
			session:  session,
		}
	}
}
func (self *gameMgr)GameDisconnect(session int64)  {
	for sid,node := range self.nodes{
		if node.session == session{
			log.Infof("服务器断开连接 sid:%v session:%v ",sid, session)
			for roomId,room := range self.rooms{
				if room.serverID == sid{
					log.Infof("清除房间:%v", roomId)
					delete(self.rooms,roomId)
				}
			}
			break
		}
	}
}

func (self *gameMgr)SendGameInfo(session int64) {
	games := make(map[uint32]*protomsg.GameInfo)
	for roomid,room := range GameMgr.rooms{
		gamenode,e := GameMgr.nodes[room.serverID]
		if !e{
			if games[gamenode.gameType].Rooms == nil {
				games[gamenode.gameType].Rooms = make([]*protomsg.RoomInfo,0)
				games[gamenode.gameType].GameType = gamenode.gameType
			}
			minMoney,t,bets := GameMgr.GetBaseInfo(roomid)
			games[gamenode.gameType].Rooms = append(games[gamenode.gameType].Rooms,&protomsg.RoomInfo{
				RoomID:roomid,
				MinMoney:minMoney,
				Type:t,
				Bets:bets,
			})
		}
	}

	if session != 0{
		// 发送房间列表
		send_tools.Send2Account(protomsg.MSG_SC_UPDATE_ROOMLIST.UInt16(),&protomsg.UPDATE_ROOMLIST{Games:games},session)
	}else {
		account.AccountMgr.SendBroadcast(protomsg.MSG_SC_UPDATE_ROOMLIST.UInt16(),&protomsg.UPDATE_ROOMLIST{Games:games},1)
	}

}
func (self *gameMgr)GetBaseInfo(roomID uint32) (minMoney uint64,t uint32, bet []uint64) {
	room := self.rooms[roomID]
	if room == nil {
		log.Warn("找不到房间:%v",roomID)
		return 0,0,nil
	}
	game := self.nodes[room.serverID]
	if game == nil {
		log.Warn("房间%v 找不到链接:%v",roomID,room.serverID)
		return 0,0,nil
	}
	switch common.EGameType(game.gameType) {
	case common.EGameTypeCATCHFISH:
	case common.EGameTypeFRUITMARY:
		minMoney = uint64(config.Get_mary_room_ConfigInt64(int(roomID),"GlodNeed"))
		t = uint32(config.Get_mary_room_ConfigInt64(int(roomID),"Type"))
		betstr := config.Get_mary_room_Config(int(roomID),"Bet")
		bet = utils.SplitConf2ArrUInt64(betstr)
		return
	default:
		log.Warn("GetBaseInfo 找不到的游戏类型:%v ",game.gameType)
		return 0,0,nil
	}
	return 0,0,nil
}


func (self *gameMgr) PrintSign(strServerIP string) {
	//if config.GetPublicConfig_Int64("APP_STORE") == 1 {
	//	log.Infof("=========== 审核标志:审核版")
	//} else {
	//	log.Infof("=========== 审核标志:正式版")
	//}
	//if config.GetPublicConfig_Int64("WHITE_LIST_OPEN") == 1 {
	//	log.Infof("=========== 白名单功能:已开启;          ServerIP:%v\r\n", strServerIP)
	//} else {
	//	log.Infof("=========== 白名单功能:已关闭;          ServerIP:%v\r\n", strServerIP)
	//}
}


func (self *gameMgr) OnEvent(e core.Event, t core.EventType) {
	switch t {
	case event.EventType_UpdateCharge:
		tWrapEv := e.(core.WrapEvent)
		tUpdateCharge := tWrapEv.Event.(event.UpdateCharge)
		log.Infof("充值邮件到账:%+v ",tUpdateCharge)
	default:

	}
}
