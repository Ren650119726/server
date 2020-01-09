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
	"root/server/dehgame/account"
	"root/server/dehgame/send_tools"
	"root/server/dehgame/types"
	"sort"
	"strconv"
	"strings"
)

const MAX_PLAYER = 6

type (
	property_sorte struct {
		S []*GamePlayer
	}

	// 坐下以后的玩家
	GamePlayer struct {
		acc           *account.Account
		status        types.EGameStatus // 当前状态
		time_of_join  int64             // 坐下的时刻 秒
		timeout_count int8              // 超时未加入次数

		cards         []common.Card_info // 牌
		showcards     int8
		bobo          int64               // 簸簸里的钱
		last_speech   types.ESpeechStatus // 最后一次喊话
		last_speech_c types.ESpeechStatus // 最后一次喊话(不会清除)
		last_bet      int64               // 最后一次喊话下注金额

		bet      int64 // 下注金额
		mangoVal int64 // 下注芒果金额

		profit      int64 // 获利的钱，负数表示输
		extractDec  int64 // 抽水
		extractBoun int64 // 奖金池获得
	}
	Room struct {
		owner  *core.Actor
		status *utils.FSM

		roomId                  uint32
		games                   uint32
		clubID                  uint32
		gameType                uint8
		matchType               uint8  // 123456
		param                   string // 参数: 1小皮 2入场 3离场 4大皮 5最小簸簸 6特殊牌型 7地九王
		status_origin_timestamp int64  // 切换状态时刻 秒
		max_bet                 int64  // 最大下注金额
		show_count              int    // 显示牌的数量

		accounts         map[uint32]*account.Account // 进房间的所有人
		seats            [MAX_PLAYER]*GamePlayer     // 局坐下的人
		lastBanker_index int                         // 最后一次庄家座位

		continues  []*GamePlayer           // 可以喊话的人
		diu        []*GamePlayer           // 丢的人列表
		qiao       []*GamePlayer           // 敲的人列表
		xiu        []*GamePlayer           // 休得人列表，临时记录，用于发牌
		mangoCount int8                    // 芒果次数
		overType   types.ESettlementStatus // 结算结果
		show_card  bool                    // 是否需要亮牌

		after_playing_pack packet.IPacket
		after_playing_bobo []int64 // 开始游戏前的bobo金额

		room_track []string // 跟踪房间流程
		permanent  bool     // 是否永久存在
		kickPlayer bool

		pipool int64

		next3cards []common.Card_info

		reward_conf map[int]string
	}
)

func NewRoom(id uint32) *Room {
	return &Room{
		roomId:           id,
		lastBanker_index: -1,
		overType:         types.ESettlementStatus_nil,
		accounts:         make(map[uint32]*account.Account),
		room_track:       make([]string, 0, 10),
		kickPlayer:       false,
	}
}

func (self *Room) Init(actor *core.Actor) bool {
	str := config.GetPublicConfig_String("DEH_REWARD_NAMES")
	self.reward_conf = utils.SplitConf2Mapis(str)
	self.owner = actor
	// 200ms 更新一次
	self.owner.AddTimer(utils.MILLISECONDS_OF_SECOND*0.2, -1, self.update)

	self.status = utils.NewFSM()
	self.status.Add(types.ERoomStatus_WAITING.Int32(), &watting_new{Room: self, s: types.ERoomStatus_WAITING})
	//self.status.Add(types.ROOM_DENG_DAI.Int32(), &watting{Room: self, s: types.ROOM_DENG_DAI})
	//self.status.Add(types.ERoomStatus_SETBOBO.Int32(), &setBoBo{Room: self, s: types.ERoomStatus_SETBOBO})
	self.status.Add(types.ERoomStatus_PLAYING.Int32(), &playing{Room: self, s: types.ERoomStatus_PLAYING})
	self.status.Add(types.ERoomStatus_SETTLEMENT.Int32(), &settlement{Room: self, s: types.ERoomStatus_SETTLEMENT})
	self.status.Add(types.ERoomStatus_ARRANGEMENT.Int32(), &arrangement{Room: self, s: types.ERoomStatus_ARRANGEMENT})
	self.status.Add(types.ERoomStatus_CLOSE.Int32(), &close{Room: self, s: types.ERoomStatus_CLOSE})

	log.Infof("房间初始化完成 param:%v match:%v", self.param, self.matchType)
	if err := self.status.Swtich(utils.SecondTimeSince1970(), types.ERoomStatus_WAITING.Int32()); err != nil {
		log.Errorf("房间状态初始化失败:%v", err.Error())
		return false
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
		self.Old_MSGID_CX_ENTER_GAME(actor, msg, session)
	case protomsg.Old_MSGID_LEAVE_GAME.UInt16(): // 客户端主动退出游戏
		self.Old_MSGID_CX_LEAVE_GAME(actor, msg, session)
	case utils.ID_DISCONNECT: // 有连接断开
		self.Disconnect(session)
	case protomsg.Old_MSGID_CX_UPDATE_MONEY.UInt16(): // 更新剩余资产
		self.SendBroadcast(msg)

	case protomsg.Old_MSGID_CX_PLAYER_LIST.UInt16(): // 请求玩家列表
		self.Old_MSGID_CX_PLAYER_LIST(actor, msg, session)
	case protomsg.Old_MSGID_SEND_EMOJI.UInt16(): // 发送魔法表情
		self.Old_MSGID_SEND_EMOJI(actor, msg, session)
	case protomsg.Old_MSGID_SEND_TEXT_SHORTCUTS.UInt16(): // 发送文字快捷聊天
		self.Old_MSGID_SEND_TEXT_SHORTCUTS(actor, msg, session)
	case protomsg.Old_MSGID_CX_SIT_DOWN.UInt16(): // 观众请求坐下
		self.Old_MSGID_CX_SIT_DOWN(actor, msg, session)

	case protomsg.Old_MSGID_CX_AWARD_HISTORY.UInt16(): // 历史记录
		self.Old_MSGID_CX_AWARD_HISTORY(actor, msg, session)

	case protomsg.Old_MSGID_CX_PROFIT_VAL.UInt16():
		self.Old_MSGID_CX_PROFIT_VAL(actor, msg, session)
	case protomsg.Old_MSGID_CX_SHOW_PERSON_INFO.UInt16():
		self.Old_MSGID_CX_SHOW_PERSON_INFO(actor, msg, session)
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

// 轮询获得下一个有效的玩家 座位号
func (self *Room) nextIndex(index int) int {
	save_count := 0
	for {
		index++
		if index >= MAX_PLAYER {
			index = 0
		}
		if self.seats[index] != nil {
			return index
		}
		if save_count > 99 {
			log.Errorf("死循环了！！！！！！")
			break
		}
	}
	return -1
}

// 切换状态
func (self *Room) switchStatus(now int64, next types.ERoomStatus) {
	self.status_origin_timestamp = now
	self.status.Swtich(now, int32(next))
}

// 进入房间条件校验
func (self *Room) canEnterRoom(accountId uint32) int {
	max_count := config.GetPublicConfig_Int64("DEH_MAX_PLAYER")
	if self.count() < int(max_count) {
		return 0
	}

	if _, exit := self.accounts[accountId]; !exit {
		return 0
	}

	return 20
}

// 更新玩家总下注，簸簸值，芒果
func (self *Room) update_bet_bobo_mango(accountId uint32) {
	index := self.seatIndex(accountId)
	if index == -1 {
		log.Errorf("更新 找不到玩家:%v", accountId)
		return
	}

	player := self.seats[index]
	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_CX_SET_PI.UInt16())
	send.WriteUInt8(uint8(index + 1))
	send.WriteUInt64(uint64(player.bobo))
	send.WriteUInt64(uint64(player.bet))
	send.WriteUInt64(uint64(player.mangoVal))
	send.WriteInt64(self.pipool)

	self.SendBroadcast(send.GetData())
}

// 进入房间条件校验
func (self *Room) sitDown(accountId uint32) int {
	index := -1
	for i := 0; i < MAX_PLAYER; i++ {
		if self.seats[i] == nil {
			index = i
			break
		}
	}

	if index == -1 {
		return -1
	}

	acc := self.accounts[accountId]
	if acc == nil {
		return -1
	}

	self.seats[index] = &GamePlayer{
		acc:           acc,
		status:        types.EGameStatus_GIVE_UP,
		time_of_join:  utils.SecondTimeSince1970() + config.GetPublicConfig_Int64("READY_TIME"),
		cards:         nil,
		bobo:          0,
		last_speech:   types.NIL,
		last_speech_c: types.NIL,
		bet:           0,
	}

	if self.status.State() == types.ERoomStatus_WAITING.Int32() {
		self.seats[index].status = types.EGameStatus_SITDOWN
	}
	return index
}

// 进入房间
func (self *Room) enterRoom(accountId uint32) {
	acc := account.AccountMgr.GetAccountByID(accountId)
	acc.RoomID = self.roomId
	self.accounts[accountId] = acc

	// 同步房间数量
	self.broadcast_count() // 进入房间

	if acc.Robot == 0 {
		log.Infof(colorized.Cyan("-> In %v Player: Accid:%v Name:%v rmb:%v %v roomid:%v games:%v"), utils.DateString(), acc.AccountId, acc.Name, acc.GetMoney(), types.ERoomStatus(self.status.State()).String(), self.roomId, acc.Games)
	} else {
		log.Infof(colorized.Cyan("-> In %v Player: Accid:%v Name:%v rmb:%v %v roomid:%v games:%v"), utils.DateString(), acc.AccountId, acc.Name, acc.GetMoney(), types.ERoomStatus(self.status.State()).String(), self.roomId, acc.Games)
	}
}

// 离开房间
func (self *Room) leaveRoom(accountId uint32, penalty bool) bool {
	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		log.Debugf("离开房间找不到玩家:%v", accountId)
		return false
	}

	index := self.seatIndex(accountId)
	// audience 观众可以随时退出
	if index != -1 {
		check_obj := self.status_obj()
		if check_obj == nil || !check_obj.CanQuit(accountId) {
			log.Warnf("当前状态:[%v] 不能离开房间", types.ERoomStatus(self.status.State()).String())
			return false
		}
	}

	for k, player := range self.seats {
		if player != nil && player.acc.AccountId == accountId {
			max_count := config.GetPublicConfig_Int64("DEH_MAX_QUIT_COUNT")
			if player.acc.Games < int32(max_count) && penalty {
				self.penalty(player)
			}
			player.acc.AddMoney(player.bobo, 0, common.EOperateType_SETTLEMENT)
			player.bobo = 0
			self.seats[k] = nil
			break
		}
	}

	send2player := packet.NewPacket(nil)
	send2player.SetMsgID(protomsg.Old_MSGID_LEAVE_GAME.UInt16())
	send2player.WriteUInt8(0)

	send_tools.Send2Account(send2player.GetData(), acc.SessionId)
	delete(self.accounts, acc.AccountId)
	acc.Games = 0

	// 同步房间数量
	self.broadcast_count() // 离开房间

	if acc.Robot == 0 {
		log.Infof(colorized.Cyan("-> Out Player: accid:%v name:%v rmb:%v roomId:%v session:%v"), acc.AccountId, acc.Name, acc.GetMoney(), self.roomId, acc.SessionId)
	} else {
		log.Infof(colorized.Cyan("-> Out Robot: accid:%v name:%v rmb:%v roomId:%v session:%v"), acc.AccountId, acc.Name, acc.GetMoney(), self.roomId, acc.SessionId)
	}

	// 通知其他玩家离线
	leaveplayer := packet.NewPacket(nil)
	leaveplayer.SetMsgID(protomsg.Old_MSGID_CX_PLAYER_LEAVE.UInt16())
	leaveplayer.WriteUInt8(uint8(index + 1))
	self.SendBroadcast(leaveplayer.GetData())

	// 2 hall
	audience := 1
	if index != -1 {
		audience = 0
	}
	send2hall := packet.NewPacket(nil)
	send2hall.SetMsgID(protomsg.Old_MSGID_UPDATE_LEAVE.UInt16())
	send2hall.WriteUInt32(acc.AccountId)
	send2hall.WriteUInt32(self.roomId)
	send2hall.WriteUInt16(uint16(self.sitDownCount()))
	send2hall.WriteUInt8(uint8(audience))
	send_tools.Send2Hall(send2hall.GetData())

	return true
}

func (self *Room) penalty(player *GamePlayer) {
	if player.profit > 0 && self.sitDownCount() != 1 {
		punish_val := math.Floor(float64(player.profit * int64(config.GetPublicConfig_Int64("QUIT_PENALTY")) / 100))
		player.bobo -= int64(punish_val)
		if self.clubID == 0 {
			RoomMgr.Add_bonus(uint32(self.GetParamInt(0)), uint64(punish_val))
		}
	}
}

func (self *Room) status_obj() IDEHStatus_universal {
	status := self.status.Current()
	check, ok := status.(IDEHStatus_universal)
	if !ok {
		log.Errorf("当前状态不是IDEHStatusCheck 请检查代码 current_status:%v", self.status.State())
		return nil
	}

	return check
}

// 从大到小，排序所有下注
func (self *Room) allBet() []*GamePlayer {
	players := []*GamePlayer{}
	for _, p := range self.seats {
		if p != nil && p.status == types.EGameStatus_PLAYING {
			players = append(players, p)
		}
	}

	sor := &property_sorte{S: players}
	sort.Sort(sor)

	return sor.S
}

// 房间总人数
func (self *Room) count() int {
	return len(self.accounts)
}

// 上座数量
func (self *Room) sitDownCount() int {
	count := 0
	for _, v := range self.seats {
		if v != nil {
			count++
		}
	}

	return count
}

// 参与者人数
func (self *Room) playerCount() int {
	count := 0
	for _, v := range self.seats {
		if v != nil && (v.status == types.EGameStatus_PLAYING || v.status == types.EGameStatus_PREPARE || v.status == types.EGameStatus_JOIN) {
			count++
		}
	}

	return count
}

// 检测玩家在哪个座位上
func (self *Room) seatIndex(accid uint32) int {
	for k, v := range self.seats {
		if v != nil && v.acc.AccountId == accid {
			return k
		}
	}

	return -1
}

func (self *Room) mango() (mango uint64) {
	_ = uint64(self.GetParamInt(0)) // 小皮
	maxP := self.GetParamInt(3)     // 大皮

	switch self.mangoCount {
	case 0:
		mango = uint64(maxP)
	case 1:
		mango = uint64(maxP) + uint64(maxP)
	case 2:
		mango = uint64(maxP) + uint64(maxP*2)
	case 3:
		mango = uint64(maxP) + uint64(maxP*4)
	}

	return mango
}

func (self *Room) minboboShow(player *GamePlayer) uint64 {
	ret := int(self.mango() * uint64(self.playerCount()+1))
	min_val := self.GetParamInt(4)
	if ret > min_val {
		min_val = ret
	}
	max_count := config.GetPublicConfig_Int64("DEH_MAX_QUIT_COUNT")

	if ret <= int(player.bobo) {
		if player.acc.Games < int32(max_count) {
			ret = int(player.bobo)
		} else {
			vv := 50 * self.GetParamInt(1)
			if ret < vv {
				ret = vv
			}
		}
	} else {
		if total := player.acc.GetMoney() + uint64(player.bobo); total < uint64(ret) {
			ret = int(player.acc.GetMoney() + uint64(player.bobo))
		} else if uint64(player.bobo) < uint64(ret) {
			if player.acc.GetMoney() >= uint64(min_val) {
				ret = int(int64(min_val) + player.bobo)
			} else {
				ret = int(player.acc.GetMoney() + uint64(player.bobo))
			}
		}
	}

	return uint64(ret)
}
func (self *Room) minbobo() uint64 {
	return uint64(self.mango() * uint64(self.playerCount()+1))

}

func (self *Room) Da_Val(accid uint32) uint32 {
	index := self.seatIndex(accid)
	if index == -1 {
		log.Errorf("错误的玩家:%v", accid)
		return 0
	}
	player := self.seats[index]
	if player == nil {
		log.Errorf("玩家不再房间内:%v", accid)
		return 0
	}

	var ret = uint32(0)
	if self.max_bet == 0 {
		ret = uint32(self.mango()) * uint32(self.playerCount()) // 总共的芒果分
	} else {
		ret = uint32(self.max_bet * 2)
	}

	return ret
}

func (self *Room) isInXIU(accid uint32) bool {
	if self.xiu == nil {
		return false
	}

	for _, player := range self.xiu {
		if player.acc.AccountId == accid {
			return true
		}
	}

	return false
}

func (self *Room) SendBroadcast(msg []byte) {
	for _, acc := range self.accounts {
		if acc.Robot == 0 && acc.SessionId > 0 {
			send_tools.Send2Account(msg, acc.SessionId)
		}
	}
}

func (self *Room) sendGameData(acc *account.Account) packet.IPacket {
	index := self.seatIndex(acc.AccountId)
	audience := 1
	// 有座位号，说明不是观战
	if index != -1 {
		audience = 0
	}
	send2acc := packet.NewPacket(nil)
	send2acc.SetMsgID(protomsg.Old_MSGID_CX_GAME_DATA.UInt16())
	send2acc.WriteUInt32(self.roomId)                     // 房间ID
	send2acc.WriteUInt32(self.clubID)                     // 俱乐部ID
	send2acc.WriteUInt8(self.gameType)                    // 匹配类型
	send2acc.WriteUInt8(self.matchType)                   // 匹配类型
	send2acc.WriteUInt32(0)                               // 房主ID
	send2acc.WriteUInt8(uint8(self.status.State()))       //房间状态
	send2acc.WriteUInt8(uint8(self.lastBanker_index + 1)) // 庄家座位号
	send2acc.WriteUInt8(uint8(index + 1))
	send2acc.WriteInt64(int64(acc.GetMoney()))
	send2acc.WriteString(self.param)
	send2acc.WriteUInt32(uint32(acc.Games))
	send2acc.WriteUInt8(uint8(audience))
	send2acc.WriteInt64(self.pipool)
	send2acc.WriteUInt8(uint8(self.mangoCount))
	send2acc.WriteUInt32(uint32(self.count() - self.sitDownCount()))

	DEH_AUTO_CHECK_CREATE_ROOM_COUNT := config.GetPublicConfig_Mapi("DEH_AUTO_CHECK_CREATE_ROOM_COUNT")
	send2acc.WriteUInt16(uint16(len(DEH_AUTO_CHECK_CREATE_ROOM_COUNT)))
	for nBet := range DEH_AUTO_CHECK_CREATE_ROOM_COUNT {
		RoomMgr.Bonus.RLock()
		if nBonus, isExist := RoomMgr.Bonus.M[uint32(nBet)]; isExist == true {
			send2acc.WriteUInt32(uint32(nBonus))
			send2acc.WriteUInt32(uint32(nBet))
			send2acc.WriteString(self.reward_conf[int(nBet)])
		} else {
			send2acc.WriteUInt32(0)
			send2acc.WriteUInt32(uint32(nBet))
			send2acc.WriteString(self.reward_conf[int(nBet)])
		}
		RoomMgr.Bonus.RUnlock()
	}

	confstr := config.GetPublicConfig_String(fmt.Sprintf("EXPEND_") + strconv.Itoa(int(self.matchType)))
	arrInt := utils.SplitConf2ArrInt32(confstr, ",")
	send2acc.WriteInt64(int64(arrInt[0]))
	send2acc.WriteInt64(int64(arrInt[1]))

	count := self.sitDownCount()
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
			send2acc.WriteInt64(player.bobo)
			send2acc.WriteInt64(player.bet)
			send2acc.WriteInt64(player.mangoVal)
			send2acc.WriteUInt8(0) // 是否特牌

			send2acc.WriteUInt8(player.status.UInt8())
			send2acc.WriteUInt8(player.last_speech_c.UInt8())

			show_self := false
			if i == index {
				show_self = true
			}
			if player.status == types.EGameStatus_GIVE_UP {
				send2acc.WriteUInt16(0)
			} else {
				attach_block := self.status_obj().ShowCard(player, show_self)
				send2acc.CatBody(attach_block)
			}

		}
	}

	return send2acc
}

func (self *Room) broadcast_count() {
	send2other := packet.NewPacket(nil)
	send2other.SetMsgID(protomsg.Old_MSGID_CX_AUDIENCE_COUNT.UInt16())
	send2other.WriteUInt32(uint32(self.count() - self.sitDownCount()))
	self.SendBroadcast(send2other.GetData())
}
func (self *Room) Close() {
	self.kickPlayer = true
}

// 连接断开处理
func (self *Room) Disconnect(session int64) {
	acc := account.AccountMgr.GetAccountBySessionID(session)
	if acc == nil {
		log.Warnf("找不到玩家:%v", session)
		return
	}

	acc = self.accounts[acc.AccountId]
	if acc == nil {
		return
	}
	acc.State = common.STATUS_OFFLINE.UInt32()
	index := self.seatIndex(acc.AccountId)
	if index != -1 {
		offline := packet.NewPacket(nil)
		offline.SetMsgID(protomsg.Old_MSGID_CX_PLAYER_OFFLINE.UInt16())
		offline.WriteUInt8(uint8(index + 1))
		self.SendBroadcast(offline.GetData())
	} else {
		self.leaveRoom(acc.AccountId, false)
	}
}

// 获得动态参数 参数下标: 0特殊牌型(三花十三花六1表示开启) 1地九王算大牌(1算大牌) 2小皮 3大皮 4最小簸簸 5入场 6离场 !!!!!ruins!!!!!!!!!!!!!!!!
//1小皮 2入场 3离场 4大皮 5最小簸簸 6特殊牌型 7地九王
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

// 获得奖金池奖励 scale 100 获得100%
func (self *Room) Get_bonus(scales []int32, accountIds []uint32, cardTypes []string, pack packet.IPacket, back func()) {
	if len(scales) != len(accountIds) {
		log.Errorf("参数不对，%v,%v ", len(scales), len(accountIds))
		return
	}
	core.LocalCoreSend(int32(self.roomId), common.EActorType_MAIN.Int32(), func() {
		bet := uint32(self.GetParamInt(0))
		RoomMgr.Bonus.RLock()
		bonusVal := RoomMgr.Bonus.M[bet]
		RoomMgr.Bonus.RUnlock()
		if bonusVal < uint64(config.GetPublicConfig_Int64("BOUNS_MIN_OPEN_AWARD")) {
			core.LocalCoreSend(common.EActorType_MAIN.Int32(), int32(self.roomId), func() {
				back()
			})
			return
		}
		total_scale := int32(0)
		for _, val := range scales {
			total_scale += val
		}

		award_bonus := uint64(math.Floor(float64(bonusVal*uint64(total_scale)/100)/100) * 100)
		RoomMgr.Bonus.Lock()
		RoomMgr.Bonus.M[bet] = bonusVal - award_bonus
		RoomMgr.Bonus.Unlock()
		core.LocalCoreSend(common.EActorType_MAIN.Int32(), int32(self.roomId), func() {
			for i, accountId := range accountIds {
				single_award := award_bonus * uint64(scales[i]) / uint64(total_scale)
				single_award = uint64(math.Floor(float64(single_award/100)) * 100)
				index := self.seatIndex(accountId)
				if index == -1 {
					acc := account.AccountMgr.GetAccountByID(accountId)
					if acc != nil {
						acc.AddMoney(int64(single_award), 0, common.EOperateType_SETTLEMENT)
					} else {
						log.Warnf("给玩家:[%v] 兑换奖励池金额:[%v]，玩家已经 不在房间内，出现此情况，请协商如何处理!!!!!!!!!!!!", accountId, award_bonus)
						return
					}
				} else {
					player := self.seats[index]
					player.bobo += int64(single_award)
					player.extractBoun += int64(single_award)
					self.track_log("玩家:[%v] 座位号:[%v] 牌型:[%v] 奖池:[%v] 获得金额:[%v] ", accountId, index, cardTypes[i], bonusVal, single_award)

					RoomMgr.Add_award_hisotry(player.acc.AccountId, player.acc.Name, uint32(single_award), cardTypes[i], bet)
				}

				// 玩家中奖
				pack.WriteUInt32(accountId)
				pack.WriteUInt8(uint8(index + 1))
				pack.WriteUInt32(uint32(single_award))
				pack.WriteString(cardTypes[i])
				// 广播给每个房间
				RoomMgr.Broadcast_update_value(bet) // 有人中奖
			}
			back()
		})
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////////
func (self *property_sorte) Len() int {
	return len(self.S)
}
func (self *property_sorte) Less(i, j int) bool {
	return self.S[i].bet > self.S[j].bet
}
func (self *property_sorte) Swap(i, j int) {
	self.S[i], self.S[j] = self.S[j], self.S[i]
}

func (self *Room) track_log(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	self.room_track = append(self.room_track, str)
	log.Infof(str)
}
