package server

import (
	"root/core"
	"root/core/packet"
)

type redis_server struct {
	owner *core.Actor
}

func NewRedisHandler() *redis_server {
	dc := &redis_server{}
	return dc
}

// actor初始化(actor接口定义)
func (self *redis_server) Init(actor *core.Actor) bool {
	self.owner = actor
	return true
}

// 停止回收相关资源
func (self *redis_server) Stop() {

}

// actor消息处理
func (self *redis_server) HandleMessage(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	//case uint16(msgserver.ID_DC_LOAD_PLAYER_DATA_REQ): // 加载
	//self.handleLoadPlayerDataReq(actor, msgid, msg, session)

	//case uint16(msgserver.ID_DC_SAVE_PLAYER_DATA_REQ): // 回存
	//	//	self.handleSavePlayerDataReq(actor, pack.GetMsgID(), msg, session)
	default:

	}
	return true
}

// 获取玩家数据
/*func (self *redis_server) handleLoadPlayerDataReq(actor int32, msgid uint16, data []byte, session int64) {
	msg := &msgserver.DC_LOAD_Player_Data_Req{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		log.Error(err.Error())
		return
	}
	// 从redis里取出玩家数据，找到数据直接返回给center,如果redis中找不到，丢给mysql处理
	dataMap, ok := db.Redis.HGetAll(rediskey.PlayerId(msg.GetPlayerId()))
	if !ok {
		core.CoreSend(actor, types.EActorType_MYSQL.Int32(), msgid, data, session)
		return
	}

	// 组装数据
	sendData := &msgserver.DC_LOAD_Player_Data_Resp{}
	sendData.PlayerData = GetPlayerData2Redis(dataMap)
	sendData.CallbackId = msg.CallbackId

	byteData, err := proto.Marshal(sendData)
	if err != nil {
		log.Errorf("proto.Marshal err:%v", err.Error())
		return
	}
	core.CoreSend(self.owner.Id, types.EActorType_CENTER_CLIENT.Int32(), uint16(msgserver.ID_DC_LOAD_PLAYER_DATA_RESP), byteData, session)
}*/

// 回存玩家数据
func (self *redis_server) handleSavePlayerDataReq(actor int32, msgid uint16, data []byte, session int64) {
	//msg := &msgserver.DC_Save_Player_Data_Req{}
	//err := proto.Unmarshal(data, msg)
	//if err != nil {
	//	log.Error(err.Error())
	//	return
	//}
	//log.Infof("回存玩家数据:%v", msg.GetPlayerId())
	//
	//SavePlayerData2Redis(msg.GetPlayerData(), msg.GetPlayerId()) // redis 更新回存玩家数据
	//
	//// todo 这里先做成回存后立刻存入sql，后面考虑数据落地相关设计
	//
	//dataMap, ok := db.Redis.HGetAll(rediskey.PlayerId(msg.GetPlayerId()))
	//if ok {
	//	core.LocalCoreSend(self.owner.Id, types.EActorType_MYSQL.Int32(), func() {
	//		// 组装数据
	//		data := GetPlayerData2Redis(dataMap)
	//		playerModel := &model.PlayerModel{}
	//		playerModel.ConvertFromPB(data)
	//		playerModel.SaveData()
	//	})
	//}

}
