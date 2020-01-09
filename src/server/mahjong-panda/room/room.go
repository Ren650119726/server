package room

import (
	"root/common"
	ca "root/common/algorithm"
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"fmt"
	"github.com/golang/protobuf/proto"
	"root/protomsg"
	"root/server/mahjong-panda/account"
	"root/server/mahjong-panda/send_tools"
	"root/server/mahjong-panda/types"
	"strconv"
	"strings"
)

const Param_max_count = 3          // 房间人数的配置参数
const reward_animation_time = 3000 // 3000毫秒

type (
	CardGroup struct {
		hand       []common.EMaJiangType
		peng       [][]common.EMaJiangType
		gang       [][]common.EMaJiangType
		last_index int
	}

	Showcard struct {
		card common.EMaJiangType
		t    uint8 //1 直杠 2 暗杠 3 弯杠4 碰
	}
	inheritData struct {
		games       int32
		profit      int64
		fee         int64
		extractBoun int64 // 奖金池
	}
	// 坐下以后的玩家
	GamePlayer struct {
		acc          *account.Account
		status       types.EGameStatus // 当前状态
		time_of_join int64             // 坐下的时刻 秒
		cards        *CardGroup
		card_time    int
		//jiao         []algorithm.Jiao_Card // 报叫以后缺的牌和胡类型
		trash_cards []common.EMaJiangType // 所有打出去的牌

		hu     common.EMaJiangHu
		hut    int8 // 1自摸 2 炮胡
		huCard common.EMaJiangType

		gang_score   map[int]int64                // 杠了以后，收到了哪些人的钱
		gang_score_z map[int]int64                // 杠 转雨 处理用
		exclude_hu   int                          // 限制胡牌 番数
		exclude_peng map[common.EMaJiangType]bool // 限制碰

		show_card []Showcard // 客户端显示peng\gang 显示数据

		money_before      uint64
		money_after       uint64
		timeout_times     int // 超时托管
		trusteeship       int // 1 托管 0 没托管
		safe_quit_timeout int64

		decide_t int8 // 定缺类型 1筒 2条 3万
	}
	Room struct {
		t      bool
		owner  *core.Actor
		status *utils.FSM

		creater                 uint32
		passwd                  map[uint32]uint8 // 坐下是否需要密码
		roomId                  uint32
		games                   uint32
		clubID                  uint32
		gameType                uint8
		matchType               uint8  // 123456
		param                   string // 参数: 底注、入场、离场、人数、加锁
		status_origin_timestamp int64  // 切换状态时刻 秒
		master                  int    // 庄家
		next_master             int    // 下一把庄家

		accounts         map[uint32]*account.Account // 进房间的所有人
		seats            []*GamePlayer               // 局坐下的人
		lastBanker_index int                         // 最后一次庄家座位

		// 结算用
		all_ting    map[int]ca.Majiang_Hu // 听牌
		all_no_ting []int                 // 未听牌
		pigs        []int                 // 花猪

		room_track []string // 跟踪房间流程
		kickPlayer bool
		hu_fan     []int32
		extra_fan  []int32

		settle_hu            packet.IPacket // 结算信息 胡牌
		settle_hu_count      uint16
		settle_gang          packet.IPacket // 结算信息 杠牌
		settle_gang_count    uint16
		settle_zy            packet.IPacket // 结算信息 转雨
		settle_zy_count      uint16
		settle_zy_count_wpos uint16
		settle_ty            packet.IPacket // 结算信息 退雨

		settle_ting packet.IPacket // 结算信息 听牌
		settle_pig  packet.IPacket // 结算信息 花猪赔钱

		settle_total_profit packet.IPacket // 结算总输赢

		reward_pool_pack       packet.IPacket // 奖金池中奖
		reward_pool_pack_count uint16         // 数量

		liuju        bool
		destory_time int64

		reward_conf map[int]string

		dispatcher *core.Dispatcher

		inherits map[uint32]inheritData
	}
)

func NewRoom(id uint32) *Room {
	return &Room{
		roomId:           id,
		lastBanker_index: -1,
		accounts:         make(map[uint32]*account.Account),
		passwd:           make(map[uint32]uint8),
		inherits:         make(map[uint32]inheritData),
	}
}

func (self *Room) SettleMsg() packet.IPacket {
	pack := packet.NewPacket(nil)

	pack.WriteUInt16(self.settle_hu_count)
	pack.CatBody(self.settle_hu)
	pack.WriteUInt16(self.settle_gang_count)
	pack.CatBody(self.settle_gang)
	self.settle_zy.Rrevise(self.settle_zy_count_wpos, self.settle_zy_count)
	pack.CatBody(self.settle_zy)
	if self.settle_ty == nil {
		pack.WriteInt8(0)   // 杠的人下标
		pack.WriteUInt16(0) // 数量
	} else {
		pack.CatBody(self.settle_ty)
	}

	pack.CatBody(self.settle_ting)
	pack.CatBody(self.settle_pig)
	pack.CatBody(self.settle_total_profit)

	// 中奖信息
	pack.WriteUInt16(self.reward_pool_pack_count)
	pack.CatBody(self.reward_pool_pack)
	return pack
}

func (self *Room) Init(actor *core.Actor) bool {
	strname := config.GetPublicConfig_String("PANDA_REWARD_NAMES")
	self.reward_conf = utils.SplitConf2Mapis(strname)
	self.owner = actor

	str := config.GetPublicConfig_String("PANDA_HU_FAN")
	self.hu_fan = utils.SplitConf2ArrInt32(str, ",")
	if len(self.hu_fan) != 21 {
		log.Errorf("配置不21个胡的番 :%v", len(self.hu_fan))
		return false
	}

	str = config.GetPublicConfig_String("PANDA_EXTRA_FAN")
	self.extra_fan = utils.SplitConf2ArrInt32(str, ",")
	if len(self.extra_fan) != 15 {
		log.Errorf("配置不21个额外的番 :%v", len(self.extra_fan))
		return false
	}

	self.dispatcher = core.NewDispatcher()

	max_count := self.GetParamInt(Param_max_count)
	self.seats = make([]*GamePlayer, max_count, max_count)
	// 200ms 更新一次
	self.owner.AddTimer(utils.MILLISECONDS_OF_SECOND*0.2, -1, self.update)

	self.status = utils.NewFSM()
	self.status.Add(types.ERoomStatus_WAITING.Int32(), &watting{Room: self, s: types.ERoomStatus_WAITING})
	self.status.Add(types.ERoomStatus_PLAYING.Int32(), &playing{Room: self, s: types.ERoomStatus_PLAYING})
	self.status.Add(types.ERoomStatus_SETTLEMENT.Int32(), &settlement{Room: self, s: types.ERoomStatus_SETTLEMENT})
	self.status.Add(types.ERoomStatus_CLOSE.Int32(), &close{Room: self, s: types.ERoomStatus_CLOSE})

	log.Infof("房间:%v 初始化完成 param:%v match:%v", self.roomId, self.param, self.matchType)
	if err := self.status.Swtich(utils.SecondTimeSince1970(), types.ERoomStatus_WAITING.Int32()); err != nil {
		log.Errorf("房间状态初始化失败:%v", err.Error())
		return false
	}
	if self.count() == 0 && self.creater > 0 {
		self.destory_time = utils.SecondTimeSince1970() + config.GetPublicConfig_Int64("DGK_DESTORY_TIME")
	} else {
		self.destory_time = -1
	}

	New_Behavior(self)
	return true
}

func (self *Room) GetBounsType() int {
	conf_arr := config.GetPublicConfig_Slice("PANDA_REWARD_POOL")
	b := self.GetParamInt(0)

	index := 0
	min := 0
	for i, v := range conf_arr {
		max := v

		if min <= b && b < max {
			index = i + 1
			break
		}

		min = max
	}

	return index
}

func (self *Room) Stop() {

}
func (self *Room) Close() {
	self.kickPlayer = true
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
	case protomsg.Old_MSGID_CX_UPDATE_MONEY.UInt16(): // 更新剩余资产
		self.SendBroadcast(msg)

	case protomsg.Old_MSGID_PANDA_AUDIENCE_LIST.UInt16(): // 请求玩家列表
		self.Old_MSGID_PANDA_AUDIENCE_LIST(actor, msg, session)
	case protomsg.Old_MSGID_SEND_EMOJI.UInt16(): // 发送魔法表情
		self.Old_MSGID_SEND_EMOJI(actor, msg, session)
	case protomsg.Old_MSGID_SEND_TEXT_SHORTCUTS.UInt16(): // 发送文字快捷聊天
		self.Old_MSGID_SEND_TEXT_SHORTCUTS(actor, msg, session)
	case protomsg.Old_MSGID_PANDA_SIT_DOWN.UInt16(): // 观众请求坐下
		self.Old_MSGID_PANDA_SIT_DOWN(actor, msg, session)
	case protomsg.Old_MSGID_PANDA_PRESON_INFO.UInt16(): // 战绩
		self.Old_MSGID_PANDA_PRESON_INFO(actor, msg, session)
	case protomsg.Old_MSGID_PANDA_GAME_STRUSATEESHIP_CANCEL.UInt16(): // 取消托管
		self.Old_MSGID_PANDA_GAME_STRUSATEESHIP_CANCEL(actor, msg, session)
	case protomsg.Old_MSGID_PANDA_PROFIT.UInt16():
		self.Old_MSGID_PANDA_PROFIT(actor, msg, session)
	case protomsg.Old_MSGID_PANDA_GAME_REWARD_HISTORY.UInt16():
		self.PANDA_GAME_REWARD_HISTORY(actor, msg, session)

	case protomsg.MSGID_HG_REENTER_OTHER_GAME.UInt16():
		self.MSGID_HG_REENTER_OTHER_GAME(actor, msg, session)
	default:
		self.status.Handle(actor, msg, session)
	}
	return true
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

func (self *Room) getRobot() *account.Account {
	for _, p := range self.accounts {
		if p.Robot == 0 {
			continue
		}
		if self.seatIndex(p.AccountId) != -1 {
			continue
		}
		return p
	}

	return nil
}

// 逻辑更新
func (self *Room) update(dt int64) {
	now := utils.SecondTimeSince1970()

	self.status.Update(now)
}

// 逻辑更新
func (self *Room) setInheritAccInfo(accid uint32, g int32, p, f, b int64) {
	self.inherits[accid] = inheritData{games: g, profit: p, fee: f, extractBoun: b}
}

// 逻辑更新
func (self *Room) trusateeship(index uint8) {
	gamePlayer := self.seats[index]
	gamePlayer.timeout_times++
	if gamePlayer.timeout_times == 3 {
		gamePlayer.trusteeship = 1

		send := packet.NewPacket(nil)
		send.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_STRUSATEESHIP.UInt16())
		send.WriteUInt8(index + 1)
		send.WriteUInt8(1)
		self.SendBroadcast(send.GetData())
	}
}

// 轮询获得下一个有效的玩家 座位号
func (self *Room) nextIndex(index int) int {
	save_count := 0
	max_count := self.GetParamInt(Param_max_count)
	for {
		index++
		if index >= max_count {
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

// 进入房间条件校验
func (self *Room) sitDown(accountId uint32) int {
	index := -1
	max_count := self.GetParamInt(Param_max_count)
	for i := 0; i < max_count; i++ {
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
		acc:          acc,
		status:       types.EGameStatus_SITDOWN,
		time_of_join: -1, //utils.SecondTimeSince1970() + config.GetPublicConfig_Int64("PANDA_READY_TIME"),
		cards:        NewCardGroup(),
	}
	self.set_need_passwd(accountId, common.ENTER_JOIN_IN_ROOM.Value())

	send2hall := packet.NewPacket(nil)
	send2hall.SetMsgID(protomsg.Old_MSGID_UPDATE_LEAVE.UInt16())
	send2hall.WriteUInt32(acc.AccountId)
	send2hall.WriteUInt32(self.roomId)
	send2hall.WriteUInt16(uint16(self.sitDownCount()))
	send2hall.WriteUInt8(uint8(1))
	send_tools.Send2Hall(send2hall.GetData())
	return index
}

// 进入房间
func (self *Room) enterRoom(accountId uint32) {
	acc := account.AccountMgr.GetAccountByID(accountId)
	acc.RoomID = self.roomId
	self.accounts[accountId] = acc

	if inheritAcc, exist := self.inherits[accountId]; exist {
		acc.Games = inheritAcc.games
		acc.Profit = inheritAcc.profit
		acc.Fee = inheritAcc.fee
		acc.ExtractBoun = inheritAcc.extractBoun
		delete(self.inherits, accountId)
	}

	// 同步房间数量
	self.broadcast_count()

	if acc.Robot == 0 {
		log.Infof(colorized.Cyan("-> In Player: Accid:%v Name:%v rmb:%v %v roomid:%v games:%v"), acc.AccountId, acc.Name, acc.GetMoney(), types.ERoomStatus(self.status.State()).String(), self.roomId, acc.Games)
	} else {
		log.Infof(colorized.Cyan("-> In Robot: Accid:%v Name:%v rmb:%v %v roomid:%v games:%v"), acc.AccountId, acc.Name, acc.GetMoney(), types.ERoomStatus(self.status.State()).String(), self.roomId, acc.Games)
	}
	self.destory_time = -1
}

// 离开房间
func (self *Room) leaveRoom(accountId uint32, penalty bool) bool {
	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		return false
	}

	if _, exist := self.accounts[accountId]; !exist {
		return false
	}

	index := -1
	for k, player := range self.seats {
		if player != nil && player.acc.AccountId == accountId {
			index = k
			break
		}
	}

	if index >= 0 && self.status.State() != types.ERoomStatus_WAITING.Int32() {
		//log.Warnf("座位上的玩家:%v 在非watting状态退出", accountId)
		return false
	}

	if penalty {
		if int64(acc.Games) < config.GetPublicConfig_Int64("PANDA_REWARD_QUIT_COUNT") {
			if acc.Profit > 0 {
				s := self.status.Current().(IPANDAStatus_universal)
				if !s.SaveQuit(accountId) {
					log.Infof("退出惩罚accid:%v, profit:%v fee:%v ", acc.AccountId, acc.Profit, acc.Fee)
					penalty_val := (acc.Profit - acc.Fee - acc.ExtractBoun) * config.GetPublicConfig_Int64("PANDA_PENALTY_RATIO") / 100
					acc.AddMoney(-penalty_val, 0, common.EOperateType_PENALTY)
					if self.clubID == 0 {
						RoomMgr.Add_bonus(uint32(self.GetParamInt(0)), uint64(penalty_val))
					}
				}
			}
		}
	}

	if index >= 0 {
		for _, player := range self.seats {
			if player != nil && player.status == types.EGameStatus_SITDOWN {
				player.time_of_join = -1 // 有人主动离开，所有人不进入到倒计时
			}
		}
		self.seats[index] = nil
	}

	send2player := packet.NewPacket(nil)
	send2player.SetMsgID(protomsg.Old_MSGID_LEAVE_GAME.UInt16())
	send2player.WriteUInt8(0)

	send_tools.Send2Account(send2player.GetData(), acc.SessionId)
	delete(self.accounts, acc.AccountId)
	acc.Games = 0
	acc.Profit = 0
	acc.Fee = 0
	acc.ExtractBoun = 0

	// 同步房间数量
	self.broadcast_count() // 离开房间

	if acc.Robot == 0 {
		log.Infof(colorized.Cyan("-> Out Player: accid:%v name:%v rmb:%v roomId:%v session:%v"), acc.AccountId, acc.Name, acc.GetMoney(), self.roomId, acc.SessionId)
	} else {
		log.Infof(colorized.Cyan("-> Out Robot: accid:%v name:%v rmb:%v roomId:%v session:%v"), acc.AccountId, acc.Name, acc.GetMoney(), self.roomId, acc.SessionId)
	}

	// 通知其他玩家离线
	leaveplayer := packet.NewPacket(nil)
	leaveplayer.SetMsgID(protomsg.Old_MSGID_PANDA_LEAVE_PLAYER.UInt16())
	leaveplayer.WriteUInt8(uint8(index + 1))
	self.SendBroadcast(leaveplayer.GetData())

	// 2 hall
	audience := 1
	if index != -1 {
		audience = 0
		self.games = 0
	}
	send2hall := packet.NewPacket(nil)
	send2hall.SetMsgID(protomsg.Old_MSGID_UPDATE_LEAVE.UInt16())
	send2hall.WriteUInt32(acc.AccountId)
	send2hall.WriteUInt32(self.roomId)
	send2hall.WriteUInt16(uint16(self.sitDownCount()))
	send2hall.WriteUInt8(uint8(audience))
	send_tools.Send2Hall(send2hall.GetData())

	core.LocalCoreSend(self.owner.Id, common.EActorType_MAIN.Int32(), func() {
		account.AccountMgr.DisconnectAccount(accountId)
	})

	if self.count() == 0 && self.creater > 0 {
		self.destory_time = utils.SecondTimeSince1970() + config.GetPublicConfig_Int64("DGK_DESTORY_TIME")
	} else {
		self.destory_time = -1
	}

	msg := packet.NewPacket(nil)
	msg.SetMsgID(protomsg.MSGID_GH_LEAVE_MATCH_NEW.UInt16())
	data, _ := proto.Marshal(&protomsg.GH_LEAVE_MATCH{AccountId: acc.AccountId})
	msg.WriteBytes(data)
	send_tools.Send2Hall(msg.GetData())
	return true
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

// 检测玩家在哪个座位上
func (self *Room) seatIndex(accid uint32) int {
	for k, v := range self.seats {
		if v != nil && v.acc.AccountId == accid {
			return k
		}
	}

	return -1
}

func (self *Room) SendBroadcast(msg []byte) {
	for _, acc := range self.accounts {
		if acc.Robot == 0 && acc.SessionId > 0 {
			send_tools.Send2Account(msg, acc.SessionId)
		}
	}
}

func (self *Room) status_obj() IPANDAStatus_universal {
	status := self.status.Current()
	check, ok := status.(IPANDAStatus_universal)
	if !ok {
		log.Errorf("IPANDAStatus_universal 请检查代码 current_status:%v", self.status.State())
		return nil
	}

	return check
}

func (self *Room) sendGameData(acc *account.Account) packet.IPacket {

	nNeedPasswd := uint8(1)
	if self.GetParamInt(4) == 0 {
		nNeedPasswd = 0
	} else if nNeedPWD, isExist := self.passwd[acc.AccountId]; isExist == true {
		nNeedPasswd = nNeedPWD
	}

	send2acc := packet.NewPacket(nil)
	send2acc.SetMsgID(protomsg.Old_MSGID_PANDA_GAME_DATA.UInt16())
	send2acc.WriteUInt32(self.roomId)               // 房间ID
	send2acc.WriteUInt32(self.clubID)               // 俱乐部ID
	send2acc.WriteUInt8(self.gameType)              // 匹配类型
	send2acc.WriteUInt8(self.matchType)             // 匹配类型
	send2acc.WriteUInt32(0)                         // 房主ID
	send2acc.WriteUInt8(uint8(self.status.State())) //房间状态
	send2acc.WriteUInt8(uint8(self.master + 1))     // 庄家位置
	index := self.seatIndex(acc.AccountId)
	send2acc.WriteUInt8(uint8(index + 1)) //  视角？？？？
	send2acc.WriteString(self.param)      //  动态参数
	if index == -1 {
		send2acc.WriteUInt32(uint32(acc.Games))
		send2acc.WriteUInt8(1)
	} else {
		gamePlayer := self.seats[index]
		send2acc.WriteUInt32(uint32(gamePlayer.acc.Games))
		send2acc.WriteUInt8(0)
	}
	send2acc.WriteUInt32(uint32(self.count() - self.sitDownCount())) // 观战人数

	PANDA_AUTO_CHECK_CREATE_ROOM_COUNT_3 := config.GetPublicConfig_Mapi("PANDA_AUTO_CHECK_CREATE_ROOM_COUNT_3")
	send2acc.WriteUInt16(uint16(len(PANDA_AUTO_CHECK_CREATE_ROOM_COUNT_3)))
	for nBet := range PANDA_AUTO_CHECK_CREATE_ROOM_COUNT_3 {
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
	send2acc.WriteUInt16(uint16(self.sitDownCount()))
	for index, player := range self.seats {
		if player != nil {
			send2acc.WriteUInt8(uint8(index + 1))
			send2acc.WriteUInt8(uint8(player.trusteeship))
			send2acc.WriteUInt32(uint32(player.acc.AccountId))
			send2acc.WriteString(player.acc.Name)
			send2acc.WriteString(player.acc.HeadURL)
			send2acc.WriteInt64(int64(player.acc.GetMoney()))
			send2acc.WriteString(player.acc.GetSignature())
			send2acc.WriteUInt8(player.acc.IsOnline())
		}
	}

	self.status_obj().CombineMSG(send2acc, acc)
	return send2acc
}

func (self *Room) broadcast_count() {
	send2other := packet.NewPacket(nil)
	send2other.SetMsgID(protomsg.Old_MSGID_PANDA_UPDATE_AUDIENCE.UInt16())
	send2other.WriteUInt32(uint32(self.count() - self.sitDownCount()))
	self.SendBroadcast(send2other.GetData())
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
	acc.State = account.STATUS_OFFLINE
	index := self.seatIndex(acc.AccountId)
	if index != -1 {
		offline := packet.NewPacket(nil)
		offline.SetMsgID(protomsg.Old_MSGID_PANDA_OFFLINE.UInt16())
		offline.WriteUInt8(uint8(index + 1))
		self.SendBroadcast(offline.GetData())
	} //else {
	//self.leaveRoom(acc.AccountId, false)
	//}
}

// 获得动态 参数: 底注、入场、离场、人数、加锁、是否换三张
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

////////////////////////////////////////////////////////////////////////////////////////////////////////

func (self *Room) track_log(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	s := fmt.Sprintf("roomID:%v ", self.roomId)
	self.room_track = append(self.room_track, s+str)
	log.Infof(str)
}
