package logic

import (
	"root/common"
	"root/core/log"
	"root/core/packet"
	"root/protomsg/inner"
	"root/server/hall/account"
	"root/server/hall/logcache"
)

// 游戏连接hall
func (self *Hall) SERVERMSG_GH_GAME_CONNECT_HALL(actor int32, msg []byte, session int64) {
	gamecon := packet.PBUnmarshal(msg,&inner.GAME_CONNECT_HALL{}).(*inner.GAME_CONNECT_HALL)
	GameMgr.GameConnectHall(gamecon.GetServerID(),gamecon.GetGameType(),session)
}

// 游戏通知大厅，房间信息
/*
 	房间由游戏主动创建，通知大厅展示
 */
func (self *Hall) SERVERMSG_GH_ROOM_INFO(actor int32, msg []byte, session int64) {
	roomInfos := packet.PBUnmarshal(msg,&inner.ROOM_INFO{}).(*inner.ROOM_INFO)
	for _,id := range roomInfos.GetRoomsID(){
		if _,e := GameMgr.rooms[id];!e{
			GameMgr.rooms[id] = &roomInfo{
				roomID:id,
				serverID:roomInfos.GetServerID(),
				PlayerCount:0,
			}
		}
	}
	log.Infof("收到 游戏 房间信息 sid:%v rooms:%v ",roomInfos.GetServerID(),roomInfos.GetRoomsID())
	// 发送游戏房间信息
	GameMgr.SendGameInfo(0)
}

// 游戏服务费
func (self *Hall) SERVERMSG_GH_SERVERFEE_LOG(actor int32, msg []byte, session int64) {
	data := packet.PBUnmarshal(msg,&inner.SERVERFEE_LOG{}).(*inner.SERVERFEE_LOG)
	logcache.LogCache.AddServiceFeeLog(data)
}

// 游戏金币改变
func (self *Hall) SERVERMSG_GH_MONEYCHANGE(actor int32, msg []byte, session int64) {
	data := packet.PBUnmarshal(msg,&inner.MONEYCHANGE{}).(*inner.MONEYCHANGE)
	logcache.LogCache.AddMoneyChangeLog(data) // 游戏通知回存金币改变日志
	acc := account.AccountMgr.GetAccountByIDAssert(data.GetAccountID())
	acc.AddMoney(data.GetChangeValue(),common.EOperateType(data.GetOperate()))
}
