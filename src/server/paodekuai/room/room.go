package room

import (
	"root/common"
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"fmt"
	"math"
	"root/protomsg"
	"root/server/paodekuai/account"
	"root/server/paodekuai/send_tools"
	"root/server/paodekuai/types"
	"strconv"
	"strings"
)

const OP_NIL = 0       // 未操作
const OP_PASS = 1      // 要不起
const OP_OUT_CARD = 2  // 出牌
const OP_GUAN_CARD = 3 // 关牌

type (
	// 坐下以后的玩家
	GamePlayer struct {
		acc                  *account.Account
		status               types.EGameStatus  // 当前状态
		op                   uint8              // 操作状态
		hand_cards           []common.Card_info // 手牌
		last_out_card        []common.Card_info // 最后一次出的牌
		bomb_out_count       uint8              // 炸了几炸
		force_ready_time     int64              // 强制准备时间, 未准备会被踢到观战列表, 单位: 毫秒
		no_penalty_quit_time int64              // 无惩罚退出时间, 单位: 毫秒
	}

	Room struct {
		owner       *core.Actor
		room_status *utils.FSM

		roomId    uint32
		games     uint32
		clubID    uint32
		gameType  uint8
		matchType uint8 // 3人房=3 4人房=4

		param         string // 创建房间参数, 跑得快参数: 1底注 2入场 3离场 4人数(3人) 5炸弹数量(1炸,3炸) 6加锁
		bet           int64  // 解析创建房间参数, 第一参数
		sitdown_limit uint64 // 解析创建房间参数, 第二参数
		situp_limit   uint64 // 解析创建房间参数, 第三参数
		max_count     uint8  // 解析创建房间参数, 第四参数
		bomb_limit    uint8  // 解析创建房间参数, 第五参数
		lock          uint8  // 解析创建房间参数, 第六参数

		status_origin_timestamp int64 // 切换状态时刻 秒

		passwd   map[uint32]uint8            // 坐下是否需要密码
		accounts map[uint32]*account.Account // 进房间的所有人
		seats    []*GamePlayer               // 局坐下的人

		all_cards       []common.Card_info // 所有的牌, 每人10张, 3人总共30张
		last_out        []common.Card_info // 上一手出的牌, 最后结算要显示
		banker_index    uint8              // 庄家下标
		is_auto_close   bool
		auto_close_time int64 // 自动关闭房间时间, 单位: 秒
		kickPlayer      bool
		room_track      []string // 跟踪房间流程

		reward_conf map[int]string
	}
)

func NewRoom(id uint32) *Room {
	return &Room{
		roomId:          id,
		accounts:        make(map[uint32]*account.Account),
		passwd:          make(map[uint32]uint8),
		auto_close_time: 0,
	}
}
func (self *Room) Init(actor *core.Actor) bool {
	str := config.GetPublicConfig_String("PDK_REWARD_NAMES")
	self.reward_conf = utils.SplitConf2Mapis(str)

	self.owner = actor

	self.bet = int64(self.GetParamInt(0))
	self.sitdown_limit = uint64(self.GetParamInt(1))
	self.situp_limit = uint64(self.GetParamInt(2))
	self.max_count = uint8(self.GetParamInt(3))
	self.bomb_limit = uint8(self.GetParamInt(4))
	self.lock = uint8(self.GetParamInt(5))

	self.seats = make([]*GamePlayer, self.max_count, self.max_count)

	// 200ms 更新一次
	self.owner.AddTimer(utils.MILLISECONDS_OF_SECOND*0.2, -1, self.update)

	self.room_status = utils.NewFSM()
	self.room_status.Add(types.ERoomStatus_WAITING.Int32(), &watting{Room: self})
	self.room_status.Add(types.ERoomStatus_PLAYING.Int32(), &playing{Room: self})
	self.room_status.Add(types.ERoomStatus_SETTLEMENT.Int32(), &settlement{Room: self})
	self.room_status.Add(types.ERoomStatus_CLOSE.Int32(), &close{Room: self})

	log.Infof("房间初始化完成 param:%v match:%v", self.param, self.matchType)
	if err := self.room_status.Swtich(utils.SecondTimeSince1970(), types.ERoomStatus_WAITING.Int32()); err != nil {
		log.Errorf("房间状态初始化失败:%v", err.Error())
		return false
	}

	if self.is_auto_close == true {
		AUTO_CLOSE_TIME := config.GetPublicConfig_Int64("PDK_AUTO_CLOSE_TIME")
		self.auto_close_time = utils.SecondTimeSince1970() + AUTO_CLOSE_TIME
	}
	return true
}

func (self *Room) Stop() {

}

// 消息处理
func (self *Room) HandleMessage(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_ENTER_GAME.UInt16(): // 客户端链接进入游戏
		self.Old_MSGID_ENTER_GAME(actor, msg, session)
	case protomsg.Old_MSGID_LEAVE_GAME.UInt16(): // 客户端主动退出游戏
		self.Old_MSGID_LEAVE_GAME(actor, msg, session)
	case utils.ID_DISCONNECT: // 有连接断开
		self.Disconnect(session)
	case protomsg.Old_MSGID_PDK_WATCH_LIST.UInt16(): // 请求观战列表
		self.Old_MSGID_PDK_WATCH_LIST(actor, msg, session)
	case protomsg.Old_MSGID_SEND_EMOJI.UInt16(): // 发送魔法表情
		self.Old_MSGID_SEND_EMOJI(actor, msg, session)
	case protomsg.Old_MSGID_SEND_TEXT_SHORTCUTS.UInt16(): // 发送文字快捷聊天
		self.Old_MSGID_SEND_TEXT_SHORTCUTS(actor, msg, session)
	case protomsg.Old_MSGID_PDK_SIT_DOWN.UInt16(): // 观众请求坐下
		self.Old_MSGID_PDK_SIT_DOWN(actor, msg, session)
	case protomsg.Old_MSGID_PDK_PROFIT.UInt16(): // 个人盈利
		self.Old_MSGID_PDK_PROFIT(actor, msg, session)
	case protomsg.Old_MSGID_PDK_AWARD_HISTORY.UInt16(): // 历史记录
		self.Old_MSGID_PDK_AWARD_HISTORY(actor, msg, session)
	case protomsg.Old_MSGID_PDK_ALL_RECORD_INFO.UInt16(): // 所有人战绩
		self.Old_MSGID_PDK_ALL_RECORD_INFO(actor, msg, session)
	default:
		self.room_status.Handle(actor, msg, session)
	}
	return true
}

// 逻辑更新
func (self *Room) update(dt int64) {
	now := utils.SecondTimeSince1970()
	self.room_status.Update(now)
}

// 轮询获得下一个有效的玩家 座位号
func (self *Room) next_index(index uint8) uint8 {
	save_count := 0
	for {
		index++
		if index >= self.max_count {
			index = 0
		}
		if self.seats[index] != nil {
			return index
		}
		save_count++
		if save_count > 10 {
			log.Errorf("死循环了！！！！！！")
			break
		}
	}
	return math.MaxUint8
}

// 切换状态
func (self *Room) switch_room_status(now int64, next types.ERoomStatus) {
	self.status_origin_timestamp = now
	self.room_status.Swtich(now, int32(next))
}

// 进入房间条件校验
func (self *Room) can_enter_room(accountId uint32) (uint8, uint8) {

	index := self.get_seat_index(accountId)
	if index < self.max_count {
		return 0, index
	}

	return 0, math.MaxUint8
}

// 检测玩家准备状态是全开启, 还是全清除
func (self *Room) check_player_ready_status() {
	nCount := self.get_sit_down_count()
	if nCount == self.max_count {
		nNowTime := utils.MilliSecondTimeSince1970()
		FORCE_READY_TIME := config.GetPublicConfig_Int64("PDK_FORCE_READY_TIME") * 1000

		isHaveNoReady := false
		for _, tPlayer := range self.seats {
			if tPlayer != nil {
				if tPlayer.status == types.EGameStatus_SITDOWN {
					tPlayer.force_ready_time = nNowTime + FORCE_READY_TIME
					isHaveNoReady = true
				} else {
					tPlayer.force_ready_time = 0
				}
			}
		}

		if isHaveNoReady == true {
			tSend := packet.NewPacket(nil)
			tSend.SetMsgID(protomsg.Old_MSGID_PDK_CHECK_PLAYER_READY.UInt16())
			tSend.WriteInt64(nNowTime + FORCE_READY_TIME)
			self.SendBroadcast(tSend.GetData())
		}
	} else {
		for _, tPlayer := range self.seats {
			if tPlayer != nil {
				tPlayer.force_ready_time = 0
			}
		}
	}
}

// 进入房间条件校验
func (self *Room) sit_down(accountId uint32) uint8 {
	index := uint8(math.MaxUint8)
	for i := uint8(0); i < self.max_count; i++ {
		if self.seats[i] == nil {
			index = i
			break
		}
	}

	if index > self.max_count {
		return math.MaxUint8
	}

	acc := self.accounts[accountId]
	if acc == nil {
		return math.MaxUint8
	}

	no_penalty_quit_time := int64(0)
	MAX_QUIT_COUNT := config.GetPublicConfig_Int64("PDK_QUIT_COUNT")
	if self.room_status.State() == types.ERoomStatus_WAITING.Int32() && acc.Profit > 0 && acc.Games < int32(MAX_QUIT_COUNT) {
		nNowTime := utils.MilliSecondTimeSince1970()
		NO_PENALTY_QUIT_TIME := config.GetPublicConfig_Int64("PDK_NO_PENALTY_QUIT_TIME") * 1000
		no_penalty_quit_time = nNowTime + NO_PENALTY_QUIT_TIME
	}

	self.seats[index] = &GamePlayer{
		acc:                  acc,
		status:               types.EGameStatus_SITDOWN,
		hand_cards:           nil,
		last_out_card:        nil,
		bomb_out_count:       0,
		op:                   OP_NIL,
		force_ready_time:     0,
		no_penalty_quit_time: no_penalty_quit_time,
	}
	self.check_player_ready_status()
	self.set_need_passwd(accountId, common.ENTER_JOIN_IN_ROOM.Value())
	return index
}

// 进入房间
func (self *Room) enter_room(acc *account.Account) {
	acc.RoomID = self.roomId
	self.accounts[acc.AccountId] = acc
	self.auto_close_time = 0

	// 同步房间数量
	self.broadcast_watch_count() // 进入房间

	if acc.Robot == 0 {
		log.Infof(colorized.Cyan("-> In Player: AccountID:%v Name:%v Money:%v RoomID:%v Games:%v SessionID:%v"), acc.AccountId, acc.Name, acc.GetMoney(), self.roomId, acc.Games, acc.SessionId)
	} else {
		log.Infof(colorized.Cyan("-> In Robot: AccountID:%v Name:%v Money:%v RoomID:%v Games:%v SessionID:%v"), acc.AccountId, acc.Name, acc.GetMoney(), self.roomId, acc.Games, acc.SessionId)
	}
}

// 离开房间
func (self *Room) leave_room(acc *account.Account, is_penalty bool) uint8 {

	index := self.get_seat_index(acc.AccountId)
	// 观战可以随时退出, 坐下玩家有退出限制
	if index < self.max_count {
		if self.can_leave_room(acc.AccountId) == false {
			log.Warnf("当前状态:[%v] 不能离开房间", types.ERoomStatus(self.room_status.State()).String())
			return 3
		}
	}

	// 处理惩罚
	penalty_val := uint32(0)
	if is_penalty == true {
		penalty_val = self.calc_penalty_value(acc)
		if penalty_val > 0 {
			acc.AddMoney(-int64(penalty_val), 0, common.EOperateType_PENALTY)
			if self.clubID == 0 {
				RoomMgr.AddBonusPool(uint32(self.bet), int64(penalty_val))
			}
		}
	}

	// 删除座位上节点数据, 必须在处理完结算后进行
	if index < self.max_count {
		self.seats[index] = nil
		self.check_player_ready_status()
	}

	if acc.Robot == 0 {
		log.Infof(colorized.Cyan("-> Out Player: AccountID:%v Name:%v Money:%v RoomID:%v Games:%v SessionID:%v 盈利:%v 惩罚:%v"), acc.AccountId, acc.Name, acc.GetMoney(), self.roomId, acc.Games, acc.SessionId, acc.Profit, penalty_val)
	} else {
		log.Infof(colorized.Cyan("-> Out Robot: AccountID:%v Name:%v Money:%v RoomID:%v Games:%v SessionID:%v 盈利:%v 惩罚:%v"), acc.AccountId, acc.Name, acc.GetMoney(), self.roomId, acc.Games, acc.SessionId, acc.Profit, penalty_val)
	}

	delete(self.accounts, acc.AccountId)
	acc.Games = 0
	acc.Profit = 0

	if self.is_auto_close == true && self.total_count() == 0 {
		AUTO_CLOSE_TIME := config.GetPublicConfig_Int64("PDK_AUTO_CLOSE_TIME")
		self.auto_close_time = utils.SecondTimeSince1970() + AUTO_CLOSE_TIME
	}

	self.broadcast_watch_count() // 离开房间, 同步房间数量

	// 通知其他玩家离线
	leaveplayer := packet.NewPacket(nil)
	leaveplayer.SetMsgID(protomsg.Old_MSGID_PDK_DEL_PLAYER.UInt16())
	leaveplayer.WriteUInt8(uint8(index + 1))
	self.SendBroadcast(leaveplayer.GetData())

	// 2 hall
	nWatch := uint8(1)
	if index < self.max_count {
		nWatch = 0
	}
	send2hall := packet.NewPacket(nil)
	send2hall.SetMsgID(protomsg.Old_MSGID_UPDATE_LEAVE.UInt16())
	send2hall.WriteUInt32(acc.AccountId)
	send2hall.WriteUInt32(self.roomId)
	send2hall.WriteUInt16(uint16(self.get_sit_down_count()))
	send2hall.WriteUInt8(nWatch)
	send_tools.Send2Hall(send2hall.GetData())

	send2player := packet.NewPacket(nil)
	send2player.SetMsgID(protomsg.Old_MSGID_LEAVE_GAME.UInt16())
	send2player.WriteUInt8(0)
	send_tools.Send2Account(send2player.GetData(), acc.SessionId)

	core.LocalCoreSend(self.owner.Id, common.EActorType_MAIN.Int32(), func() {
		account.AccountMgr.DisconnectAccount(acc)
	})
	return 0
}

// 离开座位
func (self *Room) leave_seat(acc *account.Account, index uint8) {

	tCheck := self.seats[index]
	if tCheck == nil || tCheck.acc.AccountId != acc.AccountId {
		log.Warnf("!玩家离座时, 不匹配; 下标:%v  座位玩家ID:%v, 离座玩家ID:%v", index, tCheck.acc.AccountId, acc.AccountId)
		return
	}

	self.seats[index] = nil

	// 同步房间数量
	self.check_player_ready_status()
	self.broadcast_watch_count()

	if acc.Robot == 0 {
		log.Infof(colorized.Cyan("-> Leave Seat Player: accid:%v name:%v rmb:%v roomId:%v session:%v"), acc.AccountId, acc.Name, acc.GetMoney(), self.roomId, acc.SessionId)
	} else {
		log.Infof(colorized.Cyan("-> Leave Seat Robot: accid:%v name:%v rmb:%v roomId:%v session:%v"), acc.AccountId, acc.Name, acc.GetMoney(), self.roomId, acc.SessionId)
	}

	// 通知其他玩家删除
	leaveplayer := packet.NewPacket(nil)
	leaveplayer.SetMsgID(protomsg.Old_MSGID_PDK_DEL_PLAYER.UInt16())
	leaveplayer.WriteUInt8(index + 1)
	self.SendBroadcast(leaveplayer.GetData())

	// 发送离开消息到大厅
	tLeave := packet.NewPacket(nil)
	tLeave.SetMsgID(protomsg.Old_MSGID_UPDATE_LEAVE.UInt16())
	tLeave.WriteUInt32(acc.AccountId)
	tLeave.WriteUInt32(self.roomId)
	tLeave.WriteUInt16(uint16(self.get_sit_down_count()))
	tLeave.WriteUInt8(0) // 坐下标记
	send_tools.Send2Hall(tLeave.GetData())

	// 发送进入消息到大厅
	tEnter := packet.NewPacket(nil)
	tEnter.SetMsgID(protomsg.Old_MSGID_UPDATE_ENTER.UInt16())
	tEnter.WriteUInt32(acc.AccountId)
	tEnter.WriteUInt32(self.roomId)
	tEnter.WriteUInt16(uint16(self.get_sit_down_count()))
	tEnter.WriteUInt8(1) // 观战标记
	tEnter.WriteUInt8(0) // 下标0
	send_tools.Send2Hall(tEnter.GetData())
}

// 房间总人数
func (self *Room) total_count() uint8 {
	return uint8(len(self.accounts))
}

// 上座数量
func (self *Room) get_sit_down_count() uint8 {
	count := uint8(0)
	for _, v := range self.seats {
		if v != nil {
			count++
		}
	}
	return count
}

// 检测玩家在哪个座位上
func (self *Room) get_seat_index(accid uint32) uint8 {
	for k, v := range self.seats {
		if v != nil && v.acc.AccountId == accid {
			return uint8(k)
		}
	}
	return math.MaxUint8
}

func (self *Room) SendBroadcast(msg []byte) {
	for _, acc := range self.accounts {
		if acc.Robot == 0 && acc.SessionId > 0 {
			send_tools.Send2Account(msg, acc.SessionId)
		}
	}
}

// 第二参数: 排除指定帐号的玩家; 该玩家不会收到消息
func (self *Room) send_broadcast_excludeid(msg []byte, excludeID uint32) {
	for _, acc := range self.accounts {
		if acc.Robot == 0 && acc.SessionId > 0 && acc.AccountId != excludeID {
			send_tools.Send2Account(msg, acc.SessionId)
		}
	}
}

func (self *Room) calc_penalty_value(acc *account.Account) uint32 {

	_, nChouShui, _ := self.calc_fee()
	iRealProfit := acc.Profit - nChouShui*int64(acc.Games)
	MAX_QUIT_COUNT := config.GetPublicConfig_Int64("PDK_QUIT_COUNT")
	if acc.Games < int32(MAX_QUIT_COUNT) && iRealProfit > 0 {
		isNoPenaltyQuit := false
		nIndex := self.get_seat_index(acc.AccountId)
		if nIndex < self.max_count {
			tPlayer := self.seats[nIndex]
			if tPlayer != nil {
				nNowTime := utils.MilliSecondTimeSince1970()
				// 为避免网络延迟, 故此处时间少判断500毫秒
				if tPlayer.no_penalty_quit_time > 0 && nNowTime >= tPlayer.no_penalty_quit_time-500 {
					isNoPenaltyQuit = true
				}
			}
		}

		if isNoPenaltyQuit == true {
			return 0
		} else {
			QUIT_PENALTY := config.GetPublicConfig_Int64("PDK_QUIT_PENALTY")
			penalty := iRealProfit * QUIT_PENALTY / 100
			return uint32(penalty)
		}
	}
	return 0
}

func (self *Room) send_game_data(acc *account.Account) {
	index := self.get_seat_index(acc.AccountId)
	nWatch := uint8(1)
	// 有座位号，说明不是观战
	if index < self.max_count {
		nWatch = 0
	}

	nNeedPasswd := uint8(1)
	if self.lock == 0 {
		nNeedPasswd = 0
	} else if nNeedPWD, isExist := self.passwd[acc.AccountId]; isExist == true {
		nNeedPasswd = nNeedPWD
	}

	send2acc := packet.NewPacket(nil)
	send2acc.SetMsgID(protomsg.Old_MSGID_PDK_GAME_DATA.UInt16())
	send2acc.WriteUInt32(self.roomId)                    // 房间ID
	send2acc.WriteUInt32(self.clubID)                    // 俱乐部ID
	send2acc.WriteUInt8(self.gameType)                   // 匹配类型
	send2acc.WriteUInt8(self.matchType)                  // 匹配类型
	send2acc.WriteUInt8(uint8(self.room_status.State())) //房间状态
	send2acc.WriteUInt8(self.banker_index + 1)           //庄家下标
	send2acc.WriteUInt8(uint8(index + 1))                //自己下标
	send2acc.WriteString(self.param)
	send2acc.WriteUInt32(uint32(acc.Games))
	send2acc.WriteUInt8(nWatch)
	send2acc.WriteUInt32(uint32(self.total_count() - self.get_sit_down_count()))

	PDK_AUTO_CHECK_CREATE_ROOM_COUNT := config.GetPublicConfig_Mapi("PDK_AUTO_CHECK_CREATE_ROOM_COUNT")
	send2acc.WriteUInt16(uint16(len(PDK_AUTO_CHECK_CREATE_ROOM_COUNT)))
	for nBet := range PDK_AUTO_CHECK_CREATE_ROOM_COUNT {
		if nBonus, isExist := RoomMgr.Bonus[uint32(nBet)]; isExist == true {
			send2acc.WriteUInt32(uint32(nBonus))
			send2acc.WriteUInt32(uint32(nBet))
			send2acc.WriteString(self.reward_conf[int(nBet)])
		} else {
			send2acc.WriteUInt32(0)
			send2acc.WriteUInt32(uint32(nBet))
			send2acc.WriteString(self.reward_conf[int(nBet)])
		}
	}
	send2acc.WriteUInt8(nNeedPasswd)

	count := self.get_sit_down_count()
	send2acc.WriteUInt16(uint16(count))
	for i, player := range self.seats {
		if player != nil {
			send2acc.WriteUInt8(uint8(i + 1))
			send2acc.WriteUInt32(player.acc.AccountId)
			send2acc.WriteString(player.acc.Name)
			send2acc.WriteString(player.acc.HeadURL)
			send2acc.WriteInt64(int64(player.acc.GetMoney()))
			send2acc.WriteString(player.acc.Signature)
			send2acc.WriteUInt8(player.acc.IsOnline())
			send2acc.WriteUInt8(player.status.UInt8())
			// 操作状态
			send2acc.WriteUInt8(player.op)
			send2acc.WriteUInt8(uint8(len(player.hand_cards)))

			// 最后一次出的牌
			send2acc.WriteUInt16(uint16(len(player.last_out_card)))
			for _, card := range player.last_out_card {
				send2acc.WriteUInt8(card[0])
				send2acc.WriteUInt8(card[1])
			}

		}
	}

	// 自己的手牌和明细
	if index < self.max_count {
		tSelf := self.seats[index]
		send2acc.WriteUInt16(uint16(len(tSelf.hand_cards)))
		for _, card := range tSelf.hand_cards {
			send2acc.WriteUInt8(card[0])
			send2acc.WriteUInt8(card[1])
		}
	} else {
		send2acc.WriteUInt16(0)
	}

	self.room_status_build_packet(send2acc, acc)
	send_tools.Send2Account(send2acc.GetData(), acc.SessionId)
}

func (self *Room) room_status_build_packet(tPacket packet.IPacket, tAccount *account.Account) {
	curr := self.room_status.Current()
	if status, isOk := curr.(RoomStatusExInterface); isOk == true {
		status.BulidPacket(tPacket, tAccount)
	} else {
		log.Errorf("当前房间状态:%v 不是RoomStatusInterface", self.room_status.State())
	}
}

func (self *Room) can_leave_room(accountId uint32) bool {
	index := self.get_seat_index(accountId)
	if index > self.max_count {
		return true
	}

	nRoomState := self.room_status.State()
	if nRoomState == types.ERoomStatus_PLAYING.Int32() || nRoomState == types.ERoomStatus_SETTLEMENT.Int32() {
		return false
	}
	return true
}

func (self *Room) broadcast_watch_count() {
	send2other := packet.NewPacket(nil)
	send2other.SetMsgID(protomsg.Old_MSGID_PDK_UPDATE_WATCH_COUNT.UInt16())
	send2other.WriteUInt32(uint32(self.total_count() - self.get_sit_down_count()))
	self.SendBroadcast(send2other.GetData())
}

// 获得动态参数
func (self *Room) GetParamInt(index int) int {
	strs := strings.Split(self.param, "|")
	if index >= len(strs) {
		log.Errorf("索引越界 index:%v params:%v", index, self.param)
		return -1
	}

	number, err := strconv.Atoi(strs[index])
	if err != nil {
		log.Errorf("数据解析错误 index:%v params:%v", index, self.param)
		return -1
	}
	return number
}

func (self *Room) CloseRoom() {
	self.kickPlayer = true
}

////////////////////////////////////////////////////////////////////////////////////////////////////////

func (self *Room) broadcast_out_card(nPlayerIndex uint8, sRemove []common.Card_info) {
	tSend := packet.NewPacket(nil)
	tSend.SetMsgID(protomsg.Old_MSGID_PDK_DO_OPTION.UInt16())
	tSend.WriteUInt8(0)
	tSend.WriteUInt8(OP_OUT_CARD)
	tSend.WriteUInt8(nPlayerIndex + 1)
	tSend.WriteUInt16(uint16(len(sRemove)))
	for _, card := range sRemove {
		tSend.WriteUInt8(card[0])
		tSend.WriteUInt8(card[1])
	}
	self.SendBroadcast(tSend.GetData())
}

func (self *Room) broadcast_next_op(nNextTime int64, nNewRound, nPlyaerIndex, nNextOP uint8) {
	tSend := packet.NewPacket(nil)
	tSend.SetMsgID(protomsg.Old_MSGID_PDK_GAME_OPTION.UInt16())
	tSend.WriteInt64(nNextTime)
	tSend.WriteUInt8(nNewRound)
	tSend.WriteUInt8(nPlyaerIndex + 1)
	tSend.WriteUInt8(nNextOP)
	self.SendBroadcast(tSend.GetData())
}

// 第一返回: 单局每人增加奖金池金额; 结算时增加
// 第二返回: 单局每人抽水费用; 发牌时扣除
// 第三返回: 单局每人服务费; 发牌时存储到数据库
func (self *Room) calc_fee() (int64, int64, int64) {
	nExtractScale := int64(0)
	sExtractScale := config.GetPublicConfig_ArrInt64("PDK_EXTRACT_SCALE")
	for _, tNode := range sExtractScale {
		if self.bet <= tNode[0] {
			nExtractScale = tNode[1]
			break
		}
	}
	if nExtractScale == 0 {
		nExtractScale = sExtractScale[len(sExtractScale)-1][1]
	}

	nTax := config.GetPublicConfig_Int64("TAX")
	nBonusPoolScale := config.GetPublicConfig_Int64("PDK_BONUS_POOL_SCALE")
	if self.clubID > 0 {
		nBonusPoolScale = 0
	}

	nChouShui := self.bet * nExtractScale / 100
	nAfterTax := nChouShui * nTax / 100
	nAddBonus := nAfterTax * nBonusPoolScale / 100
	nServerFee := nAfterTax - nAddBonus
	return nAddBonus, nChouShui, nServerFee
}

func (self *Room) set_need_passwd(nAccountID uint32, nEnterType uint8) {

	nNeedPwd := uint8(0)
	if nEnterType == common.ENTER_BACK_TO_ROOM.Value() {
		return
	} else if nEnterType == common.ENTER_LIST_JOIN_IN.Value() {
		nNeedPwd = 1
	}

	if _, isExist := self.passwd[nAccountID]; isExist == false {
		// 记录是否需要密码
		self.passwd[nAccountID] = nNeedPwd
	} else if nNeedPwd == 0 {
		// 更新为不需要密码
		self.passwd[nAccountID] = 0
	}
}

func (self *Room) printRoom() {
	log.Infof("房间ID:%v 参数:%v 局数:%v 状态:%v 观战:%v 坐下:%v", self.roomId, self.param, self.games, types.ERoomStatus(self.room_status.State()), self.total_count()-self.get_sit_down_count(), self.get_sit_down_count())
	for nIndex, tPlayer := range self.seats {
		if tPlayer != nil {
			log.Infof("    下标:%v 玩家ID:%v 名字:%v 身上元宝:%v 参加局数:%v 盈利:%v", nIndex, tPlayer.acc.AccountId, tPlayer.acc.Name, tPlayer.acc.GetMoney(), tPlayer.acc.Games, tPlayer.acc.Profit)
		}
	}
}

func (self *Room) track_log(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	self.room_track = append(self.room_track, str)
	fmt.Println(utils.DateString(), str)
}
