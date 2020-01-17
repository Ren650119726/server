package room

import (
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"root/common"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/server/catchfish/account"
	"root/server/catchfish/send_tools"
)

type (
	Room struct {
		owner     *core.Actor
		status    *utils.FSM
		roomId    uint32
		GameType  uint32
		ServerID  uint32
		jsonParam map[string]interface{} // config json
		accounts  map[uint32]*account.Account // 进房间的所有人
		Close     bool
	}
)

func NewRoom(id,gameType,serverID uint32,jsonParam string) *Room {
	jsonData := make(map[string]interface{})
	if e := json.Unmarshal([]byte(jsonParam),jsonData);e != nil {
		log.Panicf("解析json 错误:%v ", e.Error())
	}

	return &Room{
		accounts: 	make(map[uint32]*account.Account),
		roomId:   	id,
		GameType:   gameType,
		ServerID:   serverID,
		jsonParam:  jsonData,
		Close:    	false,
	}
}

func (self *Room) Init(actor *core.Actor) bool {
	self.owner = actor
	self.status = utils.NewFSM()
	self.status.Add(ERoomStatus_WAITING_TO_START.Int32(), &waitting{Room: self, s: ERoomStatus_WAITING_TO_START})
	self.status.Add(ERoomStatus_GAME.Int32(), &stop{Room: self, s: ERoomStatus_GAME})

	self.switchStatus(0, ERoomStatus_GAME)
	// 200ms 更新一次
	self.owner.AddTimer(utils.MILLISECONDS_OF_SECOND*0.2, -1, self.update)
	return true
}

func (self *Room) Stop() {
	RoomMgr.SaveWaterLine()
}
func (self *Room) close() {
	self.Close = true
}

// 消息处理
func (self *Room) HandleMessage(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case utils.ID_DISCONNECT: // 有连接断开
		self.Disconnect(session)

	default:
		self.status.Handle(actor, msg, session)
	}
	return true
}

// 逻辑更新
func (self *Room) update(dt int64) {
	now := utils.SecondTimeSince1970()
	self.status.Update(now)
}

// 切换状态
func (self *Room) switchStatus(now int64, next ERoomStatus) {
	self.status.Swtich(now, int32(next))
}

// 进入房间条件校验
func (self *Room) canEnterRoom(accountId uint32) int {
	if _, exit := self.accounts[accountId]; !exit {
		return 0
	}

	return 20
}

// 进入房间
func (self *Room) enterRoom(accountId uint32) {
	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		log.Errorf("找不到acc:%v", accountId)
		return
	}


	acc.RoomID = uint32(self.roomId)
	self.accounts[accountId] = acc

	if acc.Robot == 0 {
		log.Infof(colorized.Cyan("-> In roomid:%v Player:%v accid:%v name:%v money:%v %v %v"), self.roomId, utils.DateString(), acc.AccountId, acc.Name, acc.GetMoney(), ERoomStatus(self.status.State()).String(), acc.SessionId)
	} else {
		log.Infof(colorized.Cyan("-> In roomid:%v Robot:%v accid:%v name:%v money:%v %v %v"), self.roomId, utils.DateString(), acc.AccountId, acc.Name, acc.GetMoney(), ERoomStatus(self.status.State()).String(), acc.SessionId)
	}

	//update_count := packet.NewPacket(nil)
	//update_count.SetMsgID(protomsg.Old_MSGID_HONGBAO_UPDATE_COUNT.UInt16())
	//update_count.WriteUInt16(uint16(len(self.accounts)))
	//self.SendBroadcast(update_count.GetData())
}

// 离开房间
func (self *Room) leaveRoom(accountId uint32) {
	acc := self.accounts[accountId]
	if acc == nil {
		log.Debugf("离开房间找不到玩家:%v", accountId)
		return
	}

	core.LocalCoreSend(self.owner.Id, common.EActorType_MAIN.Int32(), func() {
		account.AccountMgr.DisconnectAccount(acc)
	})

	delete(self.accounts, accountId)

	if acc.Robot == 0 {
		log.Infof(colorized.Cyan("-> Out time:%v roomid:%v Player:%v name:%v money:%v %v"), utils.DateString(), self.roomId, acc.AccountId, acc.Name, acc.GetMoney(), acc.SessionId)
	} else {
		log.Infof(colorized.Cyan("-> Out time:%v roomid:%v Robot:%v name:%v money:%v %v"), utils.DateString(), self.roomId, acc.AccountId, acc.Name, acc.GetMoney(), acc.SessionId)
	}
	//
	//update_count := packet.NewPacket(nil)
	//update_count.SetMsgID(protomsg.Old_MSGID_HONGBAO_UPDATE_COUNT.UInt16())
	//update_count.WriteUInt16(uint16(len(self.accounts)))
	//self.SendBroadcast(update_count.GetData())
}

// 房间总人数
func (self *Room) count() int {
	return len(self.accounts)
}
func (self *Room) SendBroadcast(msgID uint16, pb proto.Message) {
	for _, acc := range self.accounts {
		if acc.Robot == 0 && acc.SessionId > 0 {
			send_tools.Send2Account(msgID, pb,acc.SessionId)
		}
	}
}

func (self *Room) sendGameData(acc *account.Account, status_duration int64) packet.IPacket {
	dataMSG := packet.NewPacket(nil)
	return dataMSG
}

// 连接断开处理
func (self *Room) Disconnect(session int64) {
	acc := account.AccountMgr.GetAccountBySessionID(session)
	if acc == nil {
		return
	}
	acc = self.accounts[acc.AccountId]
	if acc == nil {
		return
	}
	// 如果玩家发了红包，暂时不能离开游戏todo
	self.leaveRoom(acc.AccountId)
}
