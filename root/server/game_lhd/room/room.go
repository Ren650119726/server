package room

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"math/rand"
	"root/common"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/game_lhd/account"
	"root/server/game_lhd/send_tools"
)

type (
	Room struct {
		owner     *core.Actor
		status    *utils.FSM
		roomId    uint32
		accounts  map[uint32]*account.Account // 进房间的所有人
		Close     bool
		RoomCards []*protomsg.Card

		history         []*protomsg.ENTER_GAME_LHD_RES_Winner // 历史结果
		status_duration map[ERoomStatus]int64                 // 每个状态的持续时间 (毫秒)
		betPlayers      map[uint32]map[int32]int64            // 玩家每个区域的押注
		bets_conf       []int64                               // 房间可押注筹码值
		odds_conf       map[protomsg.LHDAREA]int64            // 区域赔率
		pump_conf       map[protomsg.LHDAREA]int64            // 区域抽水比例
		interval_conf   int64                                 // 两次下注间隔时间
		profit          int64                                 // 房间盈利
		showNum         int                                   // 开局显示的牌数
		betlimit        int64                                 //
		GameCards       []*protomsg.Card                      // 本局随机牌组
		logStack        []string                              // 房间日志堆栈
	}
)

func NewRoom(id uint32) *Room {
	return &Room{
		accounts:        make(map[uint32]*account.Account),
		roomId:          id,
		Close:           false,
		status_duration: make(map[ERoomStatus]int64),
		odds_conf:       make(map[protomsg.LHDAREA]int64),
		pump_conf:       make(map[protomsg.LHDAREA]int64),
		history:         make([]*protomsg.ENTER_GAME_LHD_RES_Winner, 0, 70),
		logStack:        make([]string, 0, 40),
	}
}

func (self *Room) Init(actor *core.Actor) bool {
	self.owner = actor
	self.LoadConfig()

	self.status = utils.NewFSM()
	self.status.Add(ERoomStatus_WAITING_TO_START.Int32(), &waitting{Room: self, s: ERoomStatus_WAITING_TO_START})
	self.status.Add(ERoomStatus_START_BETTING.Int32(), &betting{Room: self, s: ERoomStatus_START_BETTING})
	self.status.Add(ERoomStatus_STOP_BETTING.Int32(), &stop{Room: self, s: ERoomStatus_STOP_BETTING})
	self.status.Add(ERoomStatus_SETTLEMENT.Int32(), &settlement{Room: self, s: ERoomStatus_SETTLEMENT})

	self.switchStatus(0, ERoomStatus_WAITING_TO_START)

	// 200ms 更新一次
	self.owner.AddTimer(utils.MILLISECONDS_OF_SECOND*0.2, -1, self.update)

	rand.Seed(utils.SecondTimeSince1970())
	// 初始化，获取房间盈利
	return true
}

func (self *Room) log(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	s := fmt.Sprintf("roomID:%v ", self.roomId)
	self.logStack = append(self.logStack, s+str)
	log.Infof(str)
}

func (self *Room) Stop() {
	log.Infof("房间:%v 关闭", self.roomId)
}
func (self *Room) close() {
	log.Infof("房间:%v 开始关闭", self.roomId)
	self.Close = true
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
	case inner.SERVERMSG_HG_ROOM_WATER_PROFIT.UInt16(): // 房间盈利
		self.SERVERMSG_HG_ROOM_WATER_PROFIT(actor, pack.ReadBytes(), session)
	case protomsg.LHDMSG_CS_ENTER_GAME_LHD_REQ.UInt16(): // 请求进入房间
		self.LHDMSG_CS_ENTER_GAME_LHD_REQ(actor, pack.ReadBytes(), session)
	case protomsg.LHDMSG_CS_LEAVE_GAME_LHD_REQ.UInt16(): // 请求离开房间
		self.LHDMSG_CS_LEAVE_GAME_LHD_REQ(actor, pack.ReadBytes(), session)
	case protomsg.LHDMSG_CS_PLAYERS_LHD_LIST_REQ.UInt16(): // 请求玩家列表
		self.LHDMSG_CS_PLAYERS_LHD_LIST_REQ(actor, pack.ReadBytes(), session)
	default:
		self.status.Handle(actor, msg, session)
	}
	return true
}

// 逻辑更新
func (self *Room) update(dt int64) {
	now := utils.MilliSecondTimeSince1970()
	self.status.Update(now)
}

// 切换状态
func (self *Room) switchStatus(now int64, next ERoomStatus) {
	self.status.Swtich(now, int32(next))
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

	type IStatusEnterData interface {
		enterData(accountId uint32) *protomsg.StatusMsgLHD
	}
	statusEnter, b := self.status.Current().(IStatusEnterData)
	if !b {
		log.Panicf("当前状态没有处理 enterData函数 :%v ", self.status.State())
	}
	enterRoom := &protomsg.ENTER_GAME_LHD_RES{
		RoomID:         self.roomId,
		HistoryWinners: self.history,
		Bets:           self.bets_conf,
		ShowNum:        uint32(self.showNum),
		BetLimit:       uint64(self.betlimit),
		Status:         statusEnter.enterData(accountId),
	}
	if enterRoom.Status == nil {
		log.Errorf("进入房间:%v 没有Status！！！！！！！！！！！！！", self.roomId)
	}
	// 通知玩家进入游戏
	send_tools.Send2Account(protomsg.LHDMSG_SC_ENTER_GAME_LHD_RES.UInt16(), enterRoom, acc.SessionId)

	// 通知大厅 玩家进入房间
	send_tools.Send2Hall(inner.SERVERMSG_GH_PLAYER_ENTER_ROOM.UInt16(), &inner.PLAYER_ENTER_ROOM{
		AccountID: acc.GetAccountId(),
		RoomID:    self.roomId,
	})
	return
}

func (self *Room) canleave(accountId uint32) bool {
	acc := self.accounts[accountId]
	if acc == nil {
		log.Warnf("找不到玩家:%v ", accountId)
		return false
	}

	type ILeave interface {
		leave(accid uint32) bool
	}
	iLeave, b := self.status.Current().(ILeave)
	if b {
		return iLeave.leave(accountId)
	} else {
		log.Errorf("当前状态没有处理leave roomid:%v  玩家不能退出 状态:%v ", self.roomId, self.status.State())
	}
	return false
}

// 离开房间
func (self *Room) leaveRoom(accountId uint32) {
	acc := self.accounts[accountId]
	if acc == nil {
		log.Debugf("离开房间%v 找不到玩家:%v", self.roomId, accountId)
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

	// 通知大厅 玩家离开房间
	send_tools.Send2Hall(inner.SERVERMSG_GH_PLAYER_LEAVE_ROOM.UInt16(), &inner.PLAYER_LEAVE_ROOM{
		AccountID: acc.GetAccountId(),
		RoomID:    self.roomId,
	})
}

// 房间总人数
func (self *Room) count() int {
	return len(self.accounts)
}

// 玩家三个区域的总押注
func (self *Room) playerAreaBetVal(accID uint32) map[int32]int64 {
	ret := make(map[int32]int64)
	pb, e := self.betPlayers[accID]
	if !e {
		return ret
	} else {
		return pb
	}
	return ret
}

// 分别获得3个区域的总押注 robot 是否计算机器人
func (self *Room) areaBetVal(robot bool) map[int32]int64 {
	ret := make(map[int32]int64)
	if robot {
		for _, bet := range self.betPlayers {
			for area, val := range bet {
				ret[area] += val
			}

		}
	} else {
		for accid, bet := range self.betPlayers {
			acc := self.accounts[accid]
			if acc.Robot == 0 {
				for area, val := range bet {
					ret[area] += val
				}
			}
		}
	}

	return ret
}

func (self *Room) SendBroadcast(msgID uint16, pb proto.Message) {
	var bytes []byte
	if pb == nil {
		bytes = []byte{}
	} else {
		data, error := proto.Marshal(pb)
		if error != nil {
			log.Errorf("发送数据出错 roomid:%v err:%v", self.roomId, error.Error())
			return
		}
		bytes = data
	}

	for _, acc := range self.accounts {
		if acc.Robot == 0 && acc.SessionId > 0 {
			send_tools.Send2AccountBytes(msgID, bytes, acc.SessionId)
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
