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
	"root/protomsg"
	"root/server/red2black/account"
	"root/server/red2black/algorithm"
	"root/server/red2black/event"
	"root/server/red2black/send_tools"
	"sort"
)

type (
	history_ret struct {
		ret int8 // 1红 2黑 3特
		t   int8 // 特殊牌型
	}
	master_sort struct {
		List []*account.Master
	}

	Room struct {
		owner  *core.Actor
		status *utils.FSM

		roomId                  uint32
		clubID                  uint32
		gameType                uint8
		matchType               uint8
		param                   string
		game_count              uint32      // 房间局数
		status_origin_timestamp int64       // 切换状态时刻 秒
		status_duration         map[int]int // 每个状态的持续时间

		accounts map[uint32]*account.Account // 进房间的所有人
		seats    [6]*account.Account         // 坐下的人
		statis   []history_ret               // 统计输赢

		master_seats    [4]*account.Master // 庄家位
		apply_list      []*account.Master  // 申请上庄列表
		dominated_times int                // -1 拼装 >0 霸庄剩余次数

		red_cards   []algorithm.Card_info // 红方 牌型
		black_cards []algorithm.Card_info // 黑方 牌型

		Quit bool

		pack packet.IPacket

		downMasterMSG map[uint32]packet.IPacket

		total_bet_player_val int64
	}
)

func NewRoom(id uint32) *Room {
	return &Room{
		roomId:          id,
		accounts:        make(map[uint32]*account.Account),
		statis:          make([]history_ret, 0),
		red_cards:       make([]algorithm.Card_info, 0, 3),
		black_cards:     make([]algorithm.Card_info, 0, 3),
		apply_list:      make([]*account.Master, 0, 0),
		downMasterMSG:   make(map[uint32]packet.IPacket),
		dominated_times: -1,
	}
}

func (self *Room) Init(actor *core.Actor) bool {
	self.owner = actor
	self.status_duration = config.GetPublicConfig_Mapi("R2B_ROOM_TIME")
	if self.status_duration == nil {
		log.Errorf("配置错误！！！！！  请检查 R2B_ROOM_TIME")
		return false
	}
	// 200ms 更新一次
	self.owner.AddTimer(utils.MILLISECONDS_OF_SECOND*0.2, -1, self.update)
	self.owner.AddTimer(utils.MILLISECONDS_OF_SECOND*10, -1, self.up_master_update)

	self.status = utils.NewFSM()
	self.status.Add(ERoomStatus_WAITING_TO_START.Int32(), &waitting{Room: self, s: ERoomStatus_WAITING_TO_START})
	self.status.Add(ERoomStatus_GRAB_MASTER.Int32(), &master{Room: self, s: ERoomStatus_GRAB_MASTER})
	self.status.Add(ERoomStatus_START_BETTING.Int32(), &betting{Room: self, s: ERoomStatus_START_BETTING})
	self.status.Add(ERoomStatus_STOP_BETTING.Int32(), &stop{Room: self, s: ERoomStatus_STOP_BETTING})
	self.status.Add(ERoomStatus_SETTLEMENT.Int32(), &settlement{Room: self, s: ERoomStatus_SETTLEMENT})

	log.Infof("房间初始化完成 roomid:%v", self.roomId)
	if err := self.status.Swtich(utils.SecondTimeSince1970(), ERoomStatus_WAITING_TO_START.Int32()); err != nil {
		log.Errorf("房间状态初始化失败:%v", err.Error())
		return false
	}

	for i := 0; i < 20; i++ {
		self.addStatList(int8(utils.Randx_y(1, 3)), int8(utils.Randx_y(1, 7)))
	}

	New_Behavior_Bet(self)
	New_Behavior_UpMaster(self)
	New_Behavior_SeatDown(self)
	New_Behavior_Quit(self)
	New_Behavior_Emotion(self)
	return true
}

func (self *Room) Stop() {
	RoomMgr.SaveWaterLine()
}

// 消息处理
func (self *Room) HandleMessage(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_LEAVE_GAME.UInt16(): // 客户端主动退出游戏
		self.Old_MSGID_R2B_LEAVE_GAME(actor, msg, session)
	case protomsg.Old_MSGID_R2B_GAME_LEAVE_GAME.UInt16(): // 客户端主动退出游戏
		self.Old_MSGID_R2B_GAME_LEAVE_GAME(actor, msg, session)
	case utils.ID_DISCONNECT: // 有连接断开
		self.Disconnect(session)

	case protomsg.Old_MSGID_R2B_PLAYER_LIST.UInt16(): // 请求玩家列表
		self.Old_MSGID_R2B_PLAYER_LIST(actor, msg, session)
	case protomsg.Old_MSGID_R2B_STATISTICS_LIST.UInt16(): // 请求最近比赛结果
		self.Old_MSGID_R2B_STATISTICS_LIST(actor, msg, session)
	case protomsg.Old_MSGID_R2B_UP_SEAT.UInt16(): // 请求上座/换座
		self.Old_MSGID_R2B_UP_SEAT(actor, msg, session)
	case protomsg.Old_MSGID_SEND_EMOJI.UInt16(): // 发送魔法表情
		self.Old_MSGID_SEND_EMOJI(actor, msg, session)
	case protomsg.Old_MSGID_SEND_TEXT_SHORTCUTS.UInt16(): // 发送文字快捷聊天
		self.Old_MSGID_SEND_TEXT_SHORTCUTS(actor, msg, session)

	case protomsg.Old_MSGID_R2B_UP_MASTER.UInt16(): // 上庄申请
		self.Old_MSGID_R2B_UP_MASTER(actor, msg, session)
	case protomsg.Old_MSGID_R2B_UPMASTER_LIST.UInt16(): // 上庄列表申请
		self.Old_MSGID_R2B_UPMASTER_LIST(actor, msg, session)
	default:
		self.status.Handle(actor, msg, session)
	}
	return true
}

// 逻辑更新
func (self *Room) printInfo(dt int64) {
	now := utils.SecondTimeSince1970()
	self.status.Update(now)
}

// 庄家空位
func (self *Room) master_fee() int {
	for index, p := range self.master_seats {
		if p == nil {
			return index
		}
	}

	return -1
}

// 庄家空位
func (self *Room) master_count() int {
	count := 0
	for _, p := range self.master_seats {
		if p != nil {
			count++
		}
	}

	return count
}

// 逻辑更新
func (self *Room) update(dt int64) {
	now := utils.SecondTimeSince1970()
	self.status.Update(now)
}

// 逻辑更新
func (self *Room) up_master_update(dt int64) {
	event.Dispatcher.Dispatch(&event.UpMaster{
		RoomID:         self.roomId,
		Robots:         self.Robots(),
		MasterSeats:    self.master_seats,
		Applist:        self.apply_list,
		Dominate_times: self.dominated_times,
	}, event.EventType_Update_UpMaster)
}

// 切换状态
func (self *Room) switchStatus(now int64, next ERoomStatus) {
	self.status_origin_timestamp = now
	self.status.Swtich(now, int32(next))
}

// 进入房间条件校验
func (self *Room) canEnterRoom(accountId uint32) int {
	max_count := config.GetPublicConfig_Int64("R2B_MAX_PLAYER")
	if self.count() < int(max_count) {
		return 0
	}

	if _, exit := self.accounts[accountId]; !exit {
		return 0
	}

	return 20
}

func (self *Room) throw_PlayerCountChange() {
	if self.status.State() != int32(ERoomStatus_SETTLEMENT) {
		event.Dispatcher.Dispatch(&event.PlayerCountChange{
			RoomID:     self.roomId,
			TotalCount: self.count(),
			Robots:     self.Robots(),
			Seats:      self.seats,
		}, event.EventType_PlayerCountChange)
	}

}

// 进入房间
func (self *Room) enterRoom(accountId uint32) {
	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		log.Errorf("找不到acc:%v", accountId)
		return
	}
	acc.Games = 0
	acc.RoomID = self.roomId
	self.accounts[accountId] = acc

	// 通知大厅，更新账号房间信息
	self.updateEnter(acc.AccountId)

	// 同步房间数量
	self.broadcast_count()

	if acc.Robot == 0 {
		log.Infof(colorized.Cyan("-> In Player:%v accid:%v name:%v money:%v %v %v"), utils.DateString(), acc.AccountId, acc.Name, acc.GetMoney(), ERoomStatus(self.status.State()).String(), acc.SessionId)
	} else {
		log.Infof(colorized.Cyan("-> In Robot:%v accid:%v name:%v money:%v %v %v"), utils.DateString(), acc.AccountId, acc.Name, acc.GetMoney(), ERoomStatus(self.status.State()).String(), acc.SessionId)
	}

	self.throw_PlayerCountChange()
}

// 离开房间
func (self *Room) leaveRoom(accountId uint32, force bool) {
	acc := self.accounts[accountId]
	if acc == nil {
		log.Debugf("离开房间找不到玩家:%v", accountId)
		return
	}

	master_index := self.SeatMasterIndex(accountId)
	if !force && master_index != -1 {
		return
	}

	if master_index != -1 {
		for i, v := range self.master_seats {
			if v == nil {
				continue
			}
			if v.AccountId == acc.AccountId {
				self.master_seats[i] = nil
				self.dominated_times = -1
				break
			}
		}
	}
	// 下座处理
	self.downSeat(accountId)

	// 如果在申请列表，删除
	for i, app := range self.apply_list {
		if app.AccountId == acc.AccountId {
			self.apply_list = append(self.apply_list[:i], self.apply_list[i+1:]...)
			self.update_applist_count()
			break
		}
	}

	delete(self.accounts, accountId)

	core.LocalCoreSend(self.owner.Id, common.EActorType_MAIN.Int32(), func() {
		account.AccountMgr.DisconnectAccount(acc.SessionId)

		core.LocalCoreSend(0, int32(self.roomId), func() {
			// 通知大厅，更新账号房间信息
			self.updateLeave(accountId)
		})
	})
	// 同步房间数量
	self.broadcast_count()
	if acc.Robot == 0 {
		log.Infof(colorized.Cyan("-> Out Player:%v %v %v %v %v"), utils.DateString(), acc.AccountId, acc.Name, acc.RMB/10000, acc.SessionId)
	} else {
		log.Infof(colorized.Cyan("-> Out Robot:%v %v %v %v %v"), utils.DateString(), acc.AccountId, acc.Name, acc.RMB/10000, acc.SessionId)
	}
}

// 房间总人数
func (self *Room) count() int {
	return len(self.accounts)
}

// 上座数量
func (self *Room) count_seat() int {
	count := 0
	for _, v := range self.seats {
		if v != nil {
			count++
		}
	}
	return count
}

// 检测玩家在哪个座位上
func (self *Room) seatIndex(accid uint32) int {
	for k, v := range self.seats {
		if v != nil && v.AccountId == accid {
			return k
		}
	}

	return -1
}

// 检测玩家在哪个庄家座位上
func (self *Room) SeatMasterIndex(accid uint32) int {
	for k, v := range self.master_seats {
		if v != nil && v.AccountId == accid {
			return k
		}
	}

	return -1
}

// 检测玩家是否在申请列表中
func (self *Room) check_apply_list(accid uint32) bool {
	for _, v := range self.apply_list {
		if v.AccountId == accid {
			return true
		}
	}

	return false
}

// 房间机器人数量
func (self *Room) RobotCount() int {
	len := 0
	for _, v := range self.accounts {
		if v.Robot != 0 {
			len++
		}
	}
	return len
}

// 筛选所有房间机器人
func (self *Room) Robots() []*account.Account {
	ret := make([]*account.Account, 0)
	for _, acc := range self.accounts {
		if acc.Robot != 0 {
			ret = append(ret, acc)
		}
	}
	return ret
}

func (self *Room) SendBroadcast(msg []byte) {
	for _, acc := range self.accounts {
		if acc.Robot == 0 && acc.SessionId > 0 {
			send_tools.Send2Account(msg, acc.SessionId)
		}
	}
}

func (self *Room) sendGameData(acc *account.Account, status_duration uint32) packet.IPacket {
	send2acc := packet.NewPacket(nil)
	send2acc.SetMsgID(protomsg.Old_MSGID_R2B_GAME_DATA.UInt16())
	send2acc.WriteUInt32(self.roomId)
	send2acc.WriteUInt8(uint8(self.status.State()))
	send2acc.WriteUInt32(status_duration * 1000)
	send2acc.WriteUInt8(2)
	send2acc.WriteInt64(int64(acc.GetMoney()))
	send2acc.WriteUInt8(uint8(self.seatIndex(acc.AccountId) + 1))
	send2acc.WriteUInt16(uint16(len(self.accounts)))
	send2acc.WriteUInt64(uint64(config.GetPublicConfig_Int64("R2B_DOMINATE_MONEY"))) //1份对应金额
	send2acc.WriteUInt16(uint16(len(self.apply_list)))
	send2acc.WriteInt32(int32(self.dominated_times))
	temp := packet.NewPacket(nil)
	count := uint16(0)
	for i, master := range self.master_seats {
		if master != nil {
			count++
			temp.WriteUInt8(uint8(i) + 1)
			temp.WriteUInt32(master.AccountId)
			temp.WriteString(master.Name)
			temp.WriteString(master.HeadURL)
			temp.WriteInt64(int64(master.GetMoney()))
			temp.WriteString(master.Signature)
			temp.WriteUInt64(uint64(master.Share))
		}
	}
	send2acc.WriteUInt16(count)
	send2acc.CatBody(temp)

	totalbet := self.total_bet()
	send2acc.WriteUInt16(uint16(len(totalbet) - 1))
	for k, v := range totalbet {
		if k > 0 {
			send2acc.WriteUInt8(uint8(k))
			send2acc.WriteUInt32(uint32(v))
			send2acc.WriteUInt32(uint32(acc.BetVal[k]))

		}
	}

	num := self.count_seat()
	send2acc.WriteUInt16(uint16(num))
	for k, v := range self.seats {
		if v != nil {
			send2acc.WriteUInt8(uint8(k + 1))
			send2acc.WriteUInt32(v.AccountId)
			send2acc.WriteString(v.Name)
			send2acc.WriteString(fmt.Sprintf("%v", v.HeadURL))
			send2acc.WriteInt64(int64(v.GetMoney()))
			send2acc.WriteString(v.Signature)
		}
	}
	send2acc.WriteUInt16(uint16(len(self.statis)))
	for _, v := range self.statis {
		send2acc.WriteUInt8(uint8(v.ret))
		send2acc.WriteUInt8(uint8(v.t))
	}
	return send2acc
}

func (self *Room) updateEnter(accountId uint32) {
	send2hall := packet.NewPacket(nil)
	send2hall.SetMsgID(protomsg.Old_MSGID_UPDATE_ENTER.UInt16())
	send2hall.WriteUInt32(accountId)
	send2hall.WriteUInt32(self.roomId)
	send2hall.WriteUInt16(uint16(self.count()))
	send2hall.WriteUInt8(0)
	send_tools.Send2Hall(send2hall.GetData())
}
func (self *Room) updateLeave(accountId uint32) {
	send2hall := packet.NewPacket(nil)
	send2hall.SetMsgID(protomsg.Old_MSGID_UPDATE_LEAVE.UInt16())
	send2hall.WriteUInt32(accountId)
	send2hall.WriteUInt32(self.roomId)
	send2hall.WriteUInt16(uint16(self.count()))
	send2hall.WriteUInt8(0)
	send_tools.Send2Hall(send2hall.GetData())
}

func (self *Room) broadcast_count() {
	send2other := packet.NewPacket(nil)
	send2other.SetMsgID(protomsg.Old_MSGID_R2B_CHANGE_PLAYER.UInt16())
	send2other.WriteUInt16(uint16(self.count()))
	self.SendBroadcast(send2other.GetData())
}

// 下座
func (self *Room) downSeat(accountId uint32) {
	index := 0
	for i, v := range self.seats {
		if v != nil && v.AccountId == accountId {
			self.seats[i] = nil
			index = i + 1
			break
		}
	}

	if index != 0 {
		send2other := packet.NewPacket(nil)
		send2other.SetMsgID(protomsg.Old_MSGID_R2B_DOWN_SEAT.UInt16())
		send2other.WriteUInt8(uint8(index))
		self.SendBroadcast(send2other.GetData())
		log.Debugf("玩家:%v 下座 index:[%v]", accountId, index)
	}
}

// 上座
func (self *Room) UpSeat(accountId uint32, seatIndex uint8, send packet.IPacket) int8 {
	if int(seatIndex) >= len(self.seats) {
		log.Errorf("数组越界 seatindex;%v", seatIndex)
		return 12
	}

	// 位置上已经存在玩家
	if self.seats[seatIndex] != nil {
		acc := self.seats[seatIndex]
		log.Errorf("位置:[%v] 已经有玩家: name %v, accid:%v", seatIndex, acc.Name, acc.AccountId)
		return 15
	}

	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		log.Errorf("玩家不存在，%v", accountId)
		return 11
	}

	if _, exit := self.accounts[accountId]; !exit {
		log.Warnf("玩家不在房间内 %v", accountId)
		return 11
	}

	if acc.GetMoney() < uint64(config.GetPublicConfig_Int64("R2B_UP_SEAT_MONEY")) {
		return 16
	}

	before_seat_index := -1
	for i, v := range self.seats {
		if v != nil && v.AccountId == acc.AccountId {
			before_seat_index = i
			self.seats[i] = nil
			break
		}
	}

	self.seats[seatIndex] = acc

	send.WriteUInt8(0)
	send.WriteUInt8(uint8(before_seat_index + 1))
	send.WriteUInt8(seatIndex + 1)
	send.WriteUInt32(acc.AccountId)
	send.WriteString(acc.Name)
	send.WriteString(acc.HeadURL)
	send.WriteInt64(int64(acc.GetMoney()))
	send.WriteString(acc.Signature)
	self.SendBroadcast(send.GetData())
	log.Debugf("玩家:%v 上座成功 index:[%v]", acc.AccountId, seatIndex)
	return 0
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

	// 如果玩家有下注，先不能离开房间
	if acc.GetTotalBetVal() > 0 {
		acc.State = common.STATUS_OFFLINE.UInt32()
		return
	}

	if self.SeatMasterIndex(acc.AccountId) != -1 {
		acc.State = common.STATUS_OFFLINE.UInt32()
		return
	}

	self.leaveRoom(acc.AccountId, false)
}

// 添加一次结果统计
func (self *Room) addStatList(result int8, t int8) {
	max := int(config.GetPublicConfig_Int64("LHD_TREND_MAX"))
	if len(self.statis) >= max {
		self.statis = self.statis[max-20:]
	}
	self.statis = append(self.statis, history_ret{ret: result, t: t})
}

// 获得三方押注分别的总金额
func (self *Room) total_player_bet() (player_bet []uint32, robot_bet []uint32) {
	ret := make([]uint32, account.BET_KIND+1, account.BET_KIND+1)
	ret_robot := make([]uint32, account.BET_KIND+1, account.BET_KIND+1)

	for _, acc := range self.accounts {
		if acc.Robot == 0 {
			ret[1] += acc.BetVal[1]
			ret[2] += acc.BetVal[2]
			ret[3] += acc.BetVal[3]
		} else {
			ret_robot[1] += acc.BetVal[1]
			ret_robot[2] += acc.BetVal[2]
			ret_robot[3] += acc.BetVal[3]
		}
	}
	return ret, ret_robot
}

// 获得三方押注分别的总金额
func (self *Room) total_bet() []uint32 {
	ret := make([]uint32, account.BET_KIND+1, account.BET_KIND+1)

	for _, acc := range self.accounts {
		ret[1] += acc.BetVal[1]
		ret[2] += acc.BetVal[2]
		ret[3] += acc.BetVal[3]
	}
	return ret
}

// 判断是否已经操过可押注值
func (self *Room) check_bet(bets []uint32, total_share int64, choushui int) bool {
	val := int(bets[1]) - int(bets[2])
	val = utils.Abs(val)
	val += int(bets[3] * uint32(algorithm.Rate_type(common.ECardType_BAOZI)))
	val += choushui

	return val < int(total_share*config.GetPublicConfig_Int64("R2B_DOMINATE_MONEY"))
}

// 获得庄家总份额
func (self *Room) total_master_val() int64 {
	ret := int64(0)
	for _, master := range self.master_seats {
		if master != nil {
			ret += master.Share
		}
	}

	return ret
}

// 获得庄家玩家总份额
func (self *Room) total_master_player_val() int64 {
	ret := int64(0)
	for _, master := range self.master_seats {
		if master != nil && master.Robot == 0 {
			ret += master.Share
		}
	}

	return ret
}

// 更新申请列表
func (self *Room) update_applist_count() {
	response := packet.NewPacket(nil)
	response.SetMsgID(protomsg.Old_MSGID_R2B_UPDATE_APP_LIST_COUNT.UInt16())
	response.WriteUInt16(uint16(len(self.apply_list)))
	self.SendBroadcast(response.GetData())
}

func (self *Room) update_applist_sort() {

	// 排序，按照购买金额 从 大 -> 小
	so := &master_sort{List: self.apply_list}
	sort.Sort(so)
}

// 更新庄家位置
func (self *Room) update_master_list() {
	response := packet.NewPacket(nil)
	response.SetMsgID(protomsg.Old_MSGID_R2B_UPDATE_MASTER_LIST.UInt16())
	temp := packet.NewPacket(nil)
	count := uint16(0)
	for i, master := range self.master_seats {
		if master != nil {
			count++
			temp.WriteUInt8(uint8(i) + 1)
			temp.WriteUInt32(master.AccountId)
			temp.WriteString(master.Name)
			temp.WriteString(master.HeadURL)
			temp.WriteInt64(int64(master.GetMoney()))
			temp.WriteString(master.Signature)
			temp.WriteUInt64(uint64(master.Share))
		}
	}

	response.WriteUInt32(uint32(self.dominated_times))
	response.WriteUInt16(count)
	response.CatBody(temp)
	self.SendBroadcast(response.GetData())
}

func (self *master_sort) Len() int {
	return len(self.List)
}
func (self *master_sort) Less(i, j int) bool {
	return self.List[i].Share > self.List[j].Share
}
func (self *master_sort) Swap(i, j int) {
	self.List[i], self.List[j] = self.List[j], self.List[i]
}
