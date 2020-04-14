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
	"root/server/game_hongbao/account"
	"root/server/game_hongbao/send_tools"
)

type (
	hongbao struct {
		hbID         int32            // 红包实例ID
		assignerID   uint32           // 发红包的账号ID
		assignerName string           // 发红包人的名字
		value        int64            // 红包金额
		bombNumber   int64            // 雷号
		arr          []int64          // 剩余的红包
		count        int64            // 红包总数
		grabs        map[uint32]int64 // key 抢红包的人 value 抢到的金额
	}

	conf struct {
		Min_Red             int
		Max_Red             int
		Red_Count           int
		Pump                int
		Robot_Num           int
		Robot_Send_Interval string
		Robot_Send_Count    int
		Robot_Send_Value    string
		Rand_Point          int
	}

	Room struct {
		owner    *core.Actor
		roomId   uint32
		accounts map[uint32]*account.Account // 进房间的所有人
		Close    bool
		hbList   []*hongbao // 红包列表
		*conf
		addr_url string
	}
)

func NewRoom(id uint32) *Room {
	return &Room{
		accounts: make(map[uint32]*account.Account),
		roomId:   id,
		Close:    false,
		hbList:   make([]*hongbao, 10),
	}
}

func (self *Room) Init(actor *core.Actor) bool {
	self.owner = actor
	// 200ms 更新一次
	self.owner.AddTimer(utils.MILLISECONDS_OF_SECOND*3, -1, self.updateRobot)

	self.LoadConfig()

	// 请求水池金额
	send_tools.Send2Hall(inner.SERVERMSG_GH_ROOM_BONUS_REQ.UInt16(), &inner.ROOM_BONUS_REQ{
		RoomID: self.roomId,
	})
	return true
}

func (self *Room) Stop() {
	log.Infof("房间:%v 关闭", self.roomId)
}
func (self *Room) close() {
	log.Infof("房间:%v 正在关闭 剩余红包数量:%v", self.roomId, len(self.hbList))
	roomId := self.roomId
	core.LocalCoreSend(0, common.EActorType_MAIN.Int32(), func() {
		delete(RoomMgr.Rooms, roomId)
	})
	self.Close = true
	self.owner.Suspend()
}

// 消息处理
func (self *Room) HandleMessage(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case inner.SERVERMSG_SS_CLOSE_SERVER.UInt16(): //关服通知
		self.close()
	case inner.SERVERMSG_SS_RELOAD_CONFIG.UInt16(): // 更新配置
		self.LoadConfig()
	case inner.SERVERMSG_HG_NOTIFY_ALTER_DATE.UInt16(): // 大厅通知修改玩家数据
		self.SERVERMSG_HG_NOTIFY_ALTER_DATE(actor, pack.ReadBytes(), session)
	case utils.ID_DISCONNECT: // 有连接断开
		self.Disconnect(session)
	case protomsg.HBMSG_CS_ENTER_GAME_HB_REQ.UInt16(): // 请求进入房间
		self.HBMSG_CS_ENTER_GAME_HB_REQ(actor, pack.ReadBytes(), session)
	case protomsg.HBMSG_CS_LEAVE_GAME_HB_REQ.UInt16(): // 请求离开房间
		self.HBMSG_CS_LEAVE_GAME_HB_REQ(actor, pack.ReadBytes(), session)
	case protomsg.HBMSG_CS_ASSIGN_HB_REQ.UInt16(): // 请求发红包
		self.HBMSG_CS_ASSIGN_HB_REQ(actor, pack.ReadBytes(), session)
	case protomsg.HBMSG_CS_GRAB_HB_REQ.UInt16(): // 请求抢红包
		self.HBMSG_CS_GRAB_HB_REQ(actor, pack.ReadBytes(), session)
	case protomsg.HBMSG_CS_PLAYERS_HB_LIST_REQ.UInt16(): // 请求玩家列表
		self.HBMSG_CS_PLAYERS_HB_LIST_REQ(actor, pack.ReadBytes(), session)
	}
	return true
}

// 进入房间条件校验
func (self *Room) canEnterRoom(accountId uint32) int {
	if _, exit := self.accounts[accountId]; !exit {
		return 0
	}

	return 20
}

// 房间总人数 玩家人数，机器人人数
func (self *Room) countStatis() (playerc int, robotc int) {
	pc := 0
	rc := 0
	for _, acc := range self.accounts {
		if acc.Robot == 0 {
			pc++
		} else {
			rc++
		}
	}
	return pc, rc
}

// 进入房间
func (self *Room) enterRoom(accountId uint32) {
	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		log.Errorf("找不到acc:%v", accountId)
		return
	}

	acc.RoomID = self.roomId
	self.accounts[accountId] = acc

	if acc.Robot == 0 {
		log.Infof(colorized.Cyan("-> In roomid:%v Player:%v name:%v money:%v kill:%v %v session:%v"), self.roomId, acc.AccountId, acc.Name, acc.GetMoney(), acc.GetKill(), acc.SessionId)
	} else {
		log.Infof(colorized.Cyan("-> In roomid:%v Robot:%v name:%v money:%v kill:%v %v session:%v"), self.roomId, acc.AccountId, acc.Name, acc.GetMoney(), acc.GetKill(), acc.SessionId)
	}

	// 通知玩家进入游戏
	send_tools.Send2Account(protomsg.HBMSG_SC_ENTER_GAME_HB_RES.UInt16(), &protomsg.ENTER_GAME_HB_RES{
		RoomID: self.roomId,
		// todo .............
	}, acc.SessionId)

	pc, rc := self.countStatis()
	// 通知大厅 玩家进入房间
	send_tools.Send2Hall(inner.SERVERMSG_GH_PLAYER_ENTER_ROOM.UInt16(), &inner.PLAYER_ENTER_ROOM{
		AccountID:   acc.GetAccountId(),
		RoomID:      self.roomId,
		PlayerCount: uint32(pc),
		RobotCount:  uint32(rc),
	})
	return
}

func (self *Room) canleave(accountId uint32) bool {
	acc := self.accounts[accountId]
	if acc == nil {
		log.Warnf("找不到玩家:%v ", accountId)
		return false
	}

	// 如果玩家还有红包没有被抢完，不能退出房间
	for _, hb := range self.hbList {
		if hb.assignerID == accountId && len(hb.arr) != 0 {
			return false
		}
	}
	return true
}

// 离开房间
func (self *Room) leaveRoom(accountId uint32) {
	acc := self.accounts[accountId]
	if acc == nil {
		log.Debugf("离开房间找不到玩家:%v", accountId)
		return
	}

	delete(self.accounts, accountId)

	if acc.Robot == 0 {
		log.Infof(colorized.Cyan("-> Out roomid:%v Player:%v name:%v money:%v %v"), self.roomId, acc.AccountId, acc.Name, acc.GetMoney(), acc.SessionId)
	} else {
		log.Infof(colorized.Cyan("-> Out roomid:%v Robot:%v name:%v money:%v %v"), self.roomId, acc.AccountId, acc.Name, acc.GetMoney(), acc.SessionId)
	}

	core.LocalCoreSend(self.owner.Id, common.EActorType_MAIN.Int32(), func() {
		account.AccountMgr.DisconnectAccount(acc)
	})

	pc, rc := self.countStatis()
	// 通知大厅 玩家离开房间
	send_tools.Send2Hall(inner.SERVERMSG_GH_PLAYER_LEAVE_ROOM.UInt16(), &inner.PLAYER_LEAVE_ROOM{
		AccountID:   acc.GetAccountId(),
		RoomID:      self.roomId,
		PlayerCount: uint32(pc),
		RobotCount:  uint32(rc),
	})
}

// 房间总人数
func (self *Room) count() int {
	return len(self.accounts)
}
func (self *Room) SendBroadcast(msgID uint16, pb proto.Message) {
	for _, acc := range self.accounts {
		if acc.Robot == 0 && acc.SessionId > 0 {
			send_tools.Send2Account(msgID, pb, acc.SessionId)
		}
	}
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

	if self.canleave(acc.GetAccountId()) {
		self.leaveRoom(acc.AccountId)
	}
}
