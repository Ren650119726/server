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
	"root/server/game_dfdc/account"
	"root/server/game_dfdc/send_tools"
)

type (
	Room struct {
		owner                *core.Actor
		status               *utils.FSM
		roomId               uint32
		accounts             map[uint32]*account.Account // 进房间的所有人
		Close                bool
		bonus                map[int32]int64 // key等级 val水池钱
		bonusRatio           [][]int32
		bounsRoller          map[int32]int64 // 等级对应滚动率
		bounsInitGold        map[int32]int64 // 等级对应的奖池初始值
		Conf_JackpotBet      int64
		Conf_Bet_Probability int64

		bets []uint64

		mapPictureNodes map[int32]*protomsg.ENTER_GAME_DFDC_RES_DfdcRatio
		mainWheel       []*wheelNode
		freeWheel       []*wheelNode
		bonus_pattern   map[int]int
		addr_url        string
	}
)

func NewRoom(id uint32) *Room {
	return &Room{
		accounts:      make(map[uint32]*account.Account),
		roomId:        id,
		Close:         false,
		bonus:         make(map[int32]int64),
		bounsRoller:   make(map[int32]int64),
		bounsInitGold: make(map[int32]int64),
	}
}

func (self *Room) Init(actor *core.Actor) bool {
	self.owner = actor
	self.status = utils.NewFSM()
	self.status.Add(ERoomStatus_GAME.Int32(), &game{Room: self, s: ERoomStatus_GAME})

	self.switchStatus(0, ERoomStatus_GAME)
	// 200ms 更新一次
	self.owner.AddTimer(utils.MILLISECONDS_OF_SECOND*0.2, -1, self.update)
	self.owner.AddTimer(utils.MILLISECONDS_OF_SECOND*3, -1, self.updateRobot)

	self.LoadConfig()

	// 请求水池金额
	send_tools.Send2Hall(inner.SERVERMSG_GH_ROOM_BONUS_REQ.UInt16(), &inner.ROOM_BONUS_REQ{
		RoomID: self.roomId,
	})
	return true
}
func (self *Room) GetBounusLv() int32 {
	index := utils.RandomWeight32(self.bonusRatio, 1)
	return self.bonusRatio[index][0]
}

func (self *Room) Stop() {
	log.Infof("房间:%v 关闭，回存房间水池:%v ", self.roomId, self.bonus)
}
func (self *Room) close() {
	log.Infof("房间:%v 正在关闭", self.roomId)
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
	case inner.SERVERMSG_SS_RELOAD_CONFIG.UInt16(): // 重新读取配置文件
		self.LoadConfig()
	case inner.SERVERMSG_HG_NOTIFY_ALTER_DATE.UInt16(): // 大厅通知修改玩家数据
		self.SERVERMSG_HG_NOTIFY_ALTER_DATE(actor, pack.ReadBytes(), session)
	case utils.ID_DISCONNECT: // 有连接断开
		self.Disconnect(session)
	case protomsg.DFDCMSG_CS_ENTER_GAME_DFDC_REQ.UInt16(): // 请求进入房间
		self.DFDCMSG_CS_ENTER_GAME_DFDC_REQ(actor, pack.ReadBytes(), session)
	case protomsg.DFDCMSG_CS_LEAVE_GAME_DFDC_REQ.UInt16(): // 请求离开房间
		self.DFDCMSG_CS_LEAVE_GAME_DFDC_REQ(actor, pack.ReadBytes(), session)
	case protomsg.DFDCMSG_CS_START_DFDC_REQ.UInt16(): // 请求开始游戏
		self.DFDCMSG_CS_START_DFDC_REQ(actor, pack.ReadBytes(), session)
	case protomsg.DFDCMSG_CS_PLAYERS_DFDC_LIST_REQ.UInt16(): // 请求玩家列表
		self.DFDCMSG_CS_PLAYERS_DFDC_LIST_REQ(actor, pack.ReadBytes(), session)
	case inner.SERVERMSG_HG_ROOM_BONUS_RES.UInt16(): // 水池金额
		self.SERVERMSG_HG_ROOM_BONUS_RES(actor, pack.ReadBytes(), session)
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

	acc.RoomID = self.roomId
	self.accounts[accountId] = acc

	if acc.Robot == 0 {
		log.Infof(colorized.Cyan("-> In roomid:%v Player:%v name:%v money:%v kill:%v %v session:%v"), self.roomId, acc.AccountId, acc.Name, acc.GetMoney(), acc.GetKill(), ERoomStatus(self.status.State()).String(), acc.SessionId)
	} else {
		log.Infof(colorized.Cyan("-> In roomid:%v Robot:%v name:%v money:%v kill:%v %v session:%v"), self.roomId, acc.AccountId, acc.Name, acc.GetMoney(), acc.GetKill(), ERoomStatus(self.status.State()).String(), acc.SessionId)
	}

	// 通知玩家进入游戏
	send_tools.Send2Account(protomsg.DFDCMSG_SC_ENTER_GAME_DFDC_RES.UInt16(), &protomsg.ENTER_GAME_DFDC_RES{
		RoomID:   self.roomId,
		Bonus:    self.bonus,
		LastBet:  int64(acc.LastBet),
		Bets:     self.bets,
		Ratio:    self.mapPictureNodes,
		FeeCount: acc.FeeCount,
	}, acc.SessionId)

	pc, rc := self.countStatis()
	// 通知大厅 玩家进入房间
	send_tools.Send2Hall(inner.SERVERMSG_GH_PLAYER_ENTER_ROOM.UInt16(), &inner.PLAYER_ENTER_ROOM{ // dfdc
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
	if acc.FeeCount > 0 {
		return false
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

func (self *Room) SendBroadcast(msgID uint16, pb proto.Message) {
	for _, acc := range self.accounts {
		if acc.Robot == 0 && acc.SessionId > 0 {
			send_tools.Send2Account(msgID, pb, acc.SessionId)
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

	if self.canleave(acc.GetAccountId()) {
		self.leaveRoom(acc.AccountId)
	}
}
