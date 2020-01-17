package logic

import (
	"root/common"
	"root/core"
	"root/core/log"
	"root/server/hall/event"
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
