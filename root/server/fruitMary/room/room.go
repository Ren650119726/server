package room

import (
	"github.com/golang/protobuf/proto"
	"root/common"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/fruitMary/account"
	"root/server/fruitMary/send_tools"
)

type (
	Room struct {
		owner     *core.Actor
		status    *utils.FSM
		roomId    uint32
		accounts  map[uint32]*account.Account // 进房间的所有人
		Close     bool
	}
)

func NewRoom(id uint32) *Room {
	return &Room{
		accounts: 	make(map[uint32]*account.Account),
		roomId:   	id,
		Close:    	false,
	}
}

func (self *Room) Init(actor *core.Actor) bool {
	self.owner = actor
	self.status = utils.NewFSM()
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
	case protomsg.FRUITMARYMSG_CS_ENTER_GAME_FRUITMARY_REQ.UInt16(): // 请求进入小玛利房间
		self.FRUITMARYMSG_CS_ENTER_GAME_FRUITMARY_REQ(actor,msg,session)
	case protomsg.FRUITMARYMSG_CS_LEAVE_GAME_FRUITMARY_REQ.UInt16(): // 请求离开小玛利房间
		self.FRUITMARYMSG_CS_LEAVE_GAME_FRUITMARY_REQ(actor,msg,session)
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
func (self *Room) enterRoom(accountId uint32){
	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		log.Errorf("找不到acc:%v", accountId)
		return
	}


	acc.RoomID = self.roomId
	self.accounts[accountId] = acc

	if acc.Robot == 0 {
		log.Infof(colorized.Cyan("-> In roomid:%v Player:%v accid:%v name:%v money:%v %v %v"), self.roomId, acc.AccountId, acc.Name, acc.GetMoney(), ERoomStatus(self.status.State()).String(), acc.SessionId)
	} else {
		log.Infof(colorized.Cyan("-> In roomid:%v Robot:%v accid:%v name:%v money:%v %v %v"), self.roomId, acc.AccountId, acc.Name, acc.GetMoney(), ERoomStatus(self.status.State()).String(), acc.SessionId)
	}

	// 通知玩家进入游戏
	send_tools.Send2Account(protomsg.FRUITMARYMSG_SC_ENTER_GAME_FRUITMARY_RES.UInt16(),&protomsg.ENTER_GAME_FRUITMARY_REQ{
		AccountID: acc.GetAccountId(),
		RoomID:    self.roomId,
	},acc.SessionId)

	// 通知大厅 玩家进入房间
	send_tools.Send2Hall(inner.SERVERMSG_GH_PLAYER_ENTER_ROOM.UInt16(),&inner.PLAYER_ENTER_ROOM{
		AccountID: acc.GetAccountId(),
		RoomID:    self.roomId,
	})
	return
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
		log.Infof(colorized.Cyan("-> Out roomid:%v Player:%v name:%v money:%v %v"), self.roomId, acc.AccountId, acc.Name, acc.GetMoney(), acc.SessionId)
	} else {
		log.Infof(colorized.Cyan("-> Out roomid:%v Robot:%v name:%v money:%v %v"), self.roomId, acc.AccountId, acc.Name, acc.GetMoney(), acc.SessionId)
	}

	// 通知玩家 离开房间
	send_tools.Send2Hall(protomsg.FRUITMARYMSG_SC_LEAVE_GAME_FRUITMARY_RES.UInt16(),&protomsg.LEAVE_GAME_FRUITMARY_RES{
		RoomID:    self.roomId,
	})

	// 通知大厅 玩家离开房间
	send_tools.Send2Hall(inner.SERVERMSG_GH_PLAYER_LEAVE_ROOM.UInt16(),&inner.PLAYER_LEAVE_ROOM{
		AccountID: acc.GetAccountId(),
		RoomID:    self.roomId,
	})
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
