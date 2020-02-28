package logic

import (
	"github.com/golang/protobuf/proto"
	"root/common"
	"root/core/log"
	"root/core/packet"
	"root/protomsg/inner"
	"root/server/hall/account"
	"root/server/hall/logcache"
	"root/server/hall/send_tools"
)

// 游戏连接hall
func (self *Hall) SERVERMSG_GH_GAME_CONNECT_HALL(actor int32, msg []byte, session int64) {
	gamecon := packet.PBUnmarshal(msg, &inner.GAME_CONNECT_HALL{}).(*inner.GAME_CONNECT_HALL)
	GameMgr.GameConnectHall(gamecon.GetServerID(), gamecon.GetGameType(), session)
}

// 游戏通知大厅，房间信息
/*
	房间由游戏主动创建，通知大厅展示
*/
func (self *Hall) SERVERMSG_GH_ROOM_INFO(actor int32, msg []byte, session int64) {
	roomInfos := packet.PBUnmarshal(msg, &inner.ROOM_INFO{}).(*inner.ROOM_INFO)
	for _, id := range roomInfos.GetRoomsID() {
		if _, e := GameMgr.rooms[id]; !e {
			GameMgr.rooms[id] = &roomInfo{
				roomID:      id,
				serverID:    roomInfos.GetServerID(),
				PlayerCount: 0,
			}
		}

		profit, e := GameMgr.room_profit[id]
		if e {
			send_tools.Send2Game(inner.SERVERMSG_HG_ROOM_WATER_PROFIT.UInt16(), &inner.SAVE_WATER_LINE{
				RoomID:    id,
				WaterLine: profit,
			}, session)
		}
	}
	log.Infof("收到 游戏 房间信息 sid:%v rooms:%v ", roomInfos.GetServerID(), roomInfos.GetRoomsID())
	// 发送游戏房间信息
	GameMgr.SendGameInfo(0)

}

// 游戏服务费
func (self *Hall) SERVERMSG_GH_SERVERFEE_LOG(actor int32, msg []byte, session int64) {
	data := packet.PBUnmarshal(msg, &inner.SERVERFEE_LOG{}).(*inner.SERVERFEE_LOG)
	logcache.LogCache.AddServiceFeeLog(data)
}

// 游戏金币改变
func (self *Hall) SERVERMSG_GH_MONEYCHANGE(actor int32, msg []byte, session int64) {
	data := packet.PBUnmarshal(msg, &inner.MONEYCHANGE{}).(*inner.MONEYCHANGE)
	logcache.LogCache.AddMoneyChangeLog(data) // 游戏通知回存金币改变日志
	acc := account.AccountMgr.GetAccountByIDAssert(data.GetAccountID())
	acc.AddMoney(data.GetChangeValue(), common.EOperateType(data.GetOperate()))
}

// 游戏请求水池金额
func (self *Hall) SERVERMSG_GH_ROOM_BONUS_REQ(actor int32, msg []byte, session int64) {
	data := packet.PBUnmarshal(msg, &inner.ROOM_BONUS_REQ{}).(*inner.ROOM_BONUS_REQ)
	v := GameMgr.room_bonus[data.GetRoomID()]
	log.Infof("房间:%v 请求水池金额:%v ", data.RoomID, v)
	send_tools.Send2Game(inner.SERVERMSG_HG_ROOM_BONUS_RES.UInt16(), &inner.ROOM_BONUS_RES{Value: v, RoomID: data.GetRoomID()}, session)
}

// 游戏请求回存水池金额
func (self *Hall) SERVERMSG_GH_ROOM_BONUS_SAVE(actor int32, msg []byte, session int64) {
	data := packet.PBUnmarshal(msg, &inner.ROOM_BONUS_SAVE{}).(*inner.ROOM_BONUS_SAVE)
	GameMgr.room_bonus[data.GetRoomID()] = data.GetValue()
	GameMgr.savebounus = true
}

// db返回的所有房间水池
func (self *Hall) SERVERMSG_DH_ALL_ROOM_BONUS(actor int32, msg []byte, session int64) {
	if session != 0 {
		log.Infof("Error: 不是来自于DB服务器的消息, MSGID_GH_ALL_EMAIL, SessionID:%v", session)
		return
	}
	all_bonus := &inner.ALL_ROOM_BONUS{}
	err := proto.Unmarshal(msg, all_bonus)
	if err != nil {
		log.Errorf("房间水池数据读取错误:%v", err)
		return
	}
	for _, b := range all_bonus.Bonus {
		GameMgr.room_bonus[b.GetRoomID()] = b.GetValue()
		log.Infof("初始化房间:%v 水池:%v", b.GetRoomID(), b.GetValue())
	}
}

// db返回的所有房间盈利
func (self *Hall) SERVERMSG_DH_ALL_WATER_LINE(actor int32, msg []byte, session int64) {
	if session != 0 {
		log.Infof("Error: 不是来自于DB服务器的消息, MSGID_GH_ALL_EMAIL, SessionID:%v", session)
		return
	}
	all_bonus := &inner.ALL_WATER_LINE{}
	err := proto.Unmarshal(msg, all_bonus)
	if err != nil {
		log.Errorf("房间盈利数据读取错误:%v", err)
		return
	}
	for _, b := range all_bonus.Line {
		GameMgr.room_profit[b.GetRoomID()] = b.WaterLine
		log.Infof("初始化房间:%v 盈利:%v", b.GetRoomID(), b.WaterLine)
	}
}
