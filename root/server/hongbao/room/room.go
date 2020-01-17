package room

import (
	"root/common"
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/hongbao/account"
	"root/server/hongbao/send_tools"
	"strconv"
	"strings"
)

type (
	HongBao struct {
		acc      *account.Account // 发红包的人
		money    int64            // 红包金额
		bomb_num int8             // 雷号
	}
	Rob struct {
		acc   *account.Account
		money int64 // 抢到的金额
		loss  int64 // 踩雷赔钱
	}
	Room struct {
		owner     *core.Actor
		status    *utils.FSM
		roomId    uint32
		gameType  uint8
		matchType uint8
		param     string
		clubID    uint32

		accounts     map[uint32]*account.Account // 进房间的所有人
		hongbao_list []*HongBao                  // 红包列表
		rob_list     []*Rob                      // 抢到红包的所有人

		rob_num     int8     // 可抢红包数量
		profit      int64    // 本局发红包的人获利
		cur_hongbao *HongBao // 当前发的红包
		surplus_num int8     // 当前红包剩余数量

		bomb_ratio_conf [][]int64
		Close           bool
	}
)

func NewRoom(id uint32) *Room {
	return &Room{
		accounts: make(map[uint32]*account.Account),
		roomId:   id,
		Close:    false,
	}
}

func (self *Room) Init(actor *core.Actor) bool {
	self.owner = actor
	self.status = utils.NewFSM()
	self.status.Add(ERoomStatus_WAITING_TO_START.Int32(), &waitting{Room: self, s: ERoomStatus_WAITING_TO_START})
	self.status.Add(ERoomStatus_STOP_BETTING.Int32(), &stop{Room: self, s: ERoomStatus_STOP_BETTING})

	self.rob_num = int8(self.GetParamInt(4))
	self.bomb_ratio_conf = utils.SplitConf2Arr_ArrInt64(config.GetPublicConfig_String("HB_BOMB_RATIO_10"))
	if self.GetParamInt(4) == 7 {
		self.bomb_ratio_conf = utils.SplitConf2Arr_ArrInt64(config.GetPublicConfig_String("HB_BOMB_RATIO_7"))
	}
	self.switchStatus(0, ERoomStatus_STOP_BETTING)

	self.owner.AddTimer(3000, -1, self.auto_push_hongbao)
	self.owner.AddTimer(10000, -1, self.robot_quit)

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
	case protomsg.Old_MSGID_LEAVE_GAME.UInt16(): // 客户端主动退出游戏
		self.Old_MSGID_LEAVE_GAME(actor, msg, session)
	case utils.ID_DISCONNECT: // 有连接断开
		self.Disconnect(session)

	case protomsg.Old_MSGID_HONGBAO_POST_HONGBAO.UInt16(): // 请求发红包
		self.Old_MSGID_HONGBAO_POST_HONGBAO(actor, msg, session)
	case protomsg.Old_MSGID_HONGBAO_PLAYER_LIST.UInt16(): // 请求房间玩家列表
		self.Old_MSGID_HONGBAO_PLAYER_LIST(actor, msg, session)
	case protomsg.Old_MSGID_SEND_EMOJI.UInt16(): // 发送魔法表情
		self.Old_MSGID_SEND_EMOJI(actor, msg, session)
	case protomsg.Old_MSGID_SEND_TEXT_SHORTCUTS.UInt16(): // 发送文字快捷聊天
		self.Old_MSGID_SEND_TEXT_SHORTCUTS(actor, msg, session)
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

// 获得动态参数 参数下标: 0底注红包金额 1入场 2离场 3最高红包是底注红包的多少倍 4一个红包几人抢 5触雷赔付红包金额倍数(放大100倍填写)
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

// 把红包切分成n份
func (self *Room) hongbao_slice(hongbao_num, hongbao_money int64, bomb_num int8) (bombdeal, normaldeal []int64) {
	bombdeal = make([]int64, 0)
	normaldeal = make([]int64, 0)

	var random_recursion func(num, money int64)
	random_recursion = func(num, money int64) {
		if num == 0 {
			return
		}
		calculate_val := money / num // 均等值
		if calculate_val != money {
			// 随机化
			rand_conf := config.GetPublicConfig_Int64("HB_RANDOM")
			rand_val := utils.Randx_y(0, int(money*rand_conf/100))
			calculate_val += int64(rand_val)
		}

		//做踩雷处理
		if int8(calculate_val%10) == bomb_num {
			bombdeal = append(bombdeal, calculate_val)
		} else {
			normaldeal = append(normaldeal, calculate_val)
		}

		random_recursion(num-1, money-calculate_val)
	}

	random_recursion(hongbao_num, hongbao_money)
	return bombdeal, normaldeal
}

// 雷数调控
func (self *Room) schedule_bomb(bomb, normal []int64, need_count, bomb_num int8) (bombdeal, normaldeal []int64) {
	surplus_bomb := int8(len(bomb)) - need_count
	if surplus_bomb > 0 { // 消除雷
		total := 0
		n := int(surplus_bomb)
		for i := 0; i < n; i++ {
			total += 1
			bomb[i] -= 1
		}

		normal = append(normal, bomb[:n]...)
		bomb = bomb[n:]

		for k, v := range normal {
			if int8((v+int64(total))%10) != bomb_num {
				normal[k] += int64(total)
				break
			}
		}
	} else { // 加雷
		total := int64(0)
		n := int(-surplus_bomb)
		for i := 0; i < n; i++ {
			t := (normal[i] % 10) - int64(bomb_num)
			normal[i] -= t
			total += t
		}

		bomb = append(bomb, normal[:n]...)
		normal = normal[n:]

		for k, v := range normal {
			if int8((v+int64(total))%10) != bomb_num {
				normal[k] += int64(total)
				break
			}
		}
	}
	return bomb, normal
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

	// 通知大厅，更新账号房间信息
	self.updateEnter(acc.AccountId)

	acc.RoomID = self.roomId
	self.accounts[accountId] = acc

	if acc.Robot == 0 {
		log.Infof(colorized.Cyan("-> In roomid:%v Player:%v accid:%v name:%v money:%v %v %v"), self.roomId, utils.DateString(), acc.AccountId, acc.Name, acc.GetMoney(), ERoomStatus(self.status.State()).String(), acc.SessionId)
	} else {
		log.Infof(colorized.Cyan("-> In roomid:%v Robot:%v accid:%v name:%v money:%v %v %v"), self.roomId, utils.DateString(), acc.AccountId, acc.Name, acc.GetMoney(), ERoomStatus(self.status.State()).String(), acc.SessionId)
	}

	update_count := packet.NewPacket(nil)
	update_count.SetMsgID(protomsg.Old_MSGID_HONGBAO_UPDATE_COUNT.UInt16())
	update_count.WriteUInt16(uint16(len(self.accounts)))
	self.SendBroadcast(update_count.GetData())
}

// 离开房间
func (self *Room) leaveRoom(accountId uint32) {
	acc := self.accounts[accountId]
	if acc == nil {
		log.Debugf("离开房间找不到玩家:%v", accountId)
		return
	}
	acc.Quit_flag = false

	for _, v := range self.rob_list {
		if v.acc.AccountId == accountId {
			return
		}
	}

	for _, v := range self.hongbao_list {
		if v.acc.AccountId == accountId {
			return
		}
	}

	if self.cur_hongbao != nil && self.cur_hongbao.acc.AccountId == accountId {
		return
	}

	core.LocalCoreSend(self.owner.Id, common.EActorType_MAIN.Int32(), func() {
		account.AccountMgr.DisconnectAccount(acc)
	})

	// 通知大厅，更新账号房间信息
	self.updateLeave(accountId)
	delete(self.accounts, accountId)

	if acc.Robot == 0 {
		log.Infof(colorized.Cyan("-> Out time:%v roomid:%v Player:%v name:%v money:%v %v"), utils.DateString(), self.roomId, acc.AccountId, acc.Name, acc.GetMoney(), acc.SessionId)
	} else {
		log.Infof(colorized.Cyan("-> Out time:%v roomid:%v Robot:%v name:%v money:%v %v"), utils.DateString(), self.roomId, acc.AccountId, acc.Name, acc.GetMoney(), acc.SessionId)
	}

	update_count := packet.NewPacket(nil)
	update_count.SetMsgID(protomsg.Old_MSGID_HONGBAO_UPDATE_COUNT.UInt16())
	update_count.WriteUInt16(uint16(len(self.accounts)))
	self.SendBroadcast(update_count.GetData())
}

// 房间总人数
func (self *Room) count() int {
	return len(self.accounts)
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

func (self *Room) sendGameData(acc *account.Account, status_duration int64) packet.IPacket {
	dataMSG := packet.NewPacket(nil)
	dataMSG.SetMsgID(protomsg.Old_MSGID_HONGBAO_GAME_DATA.UInt16())
	dataMSG.WriteUInt32(self.roomId)
	dataMSG.WriteUInt8(uint8(self.status.State()))
	dataMSG.WriteInt64(status_duration)
	dataMSG.WriteInt64(int64(acc.GetMoney()))
	dataMSG.WriteString(self.param)
	dataMSG.WriteUInt32(uint32(self.count()))
	dataMSG.WriteUInt32(self.cur_hongbao.acc.AccountId)
	dataMSG.WriteString(self.cur_hongbao.acc.Name)
	dataMSG.WriteString(self.cur_hongbao.acc.HeadURL)
	dataMSG.WriteInt64(self.cur_hongbao.money)
	dataMSG.WriteInt8(self.surplus_num)
	dataMSG.WriteInt8(self.cur_hongbao.bomb_num)

	dataMSG.WriteUInt16(uint16(len(self.rob_list)))
	for _, v := range self.rob_list {
		dataMSG.WriteUInt32(v.acc.AccountId)
		dataMSG.WriteString(v.acc.Name)
		dataMSG.WriteString(v.acc.HeadURL)
		dataMSG.WriteInt64(int64(v.acc.GetMoney()))
		dataMSG.WriteString(v.acc.Signature)
		dataMSG.WriteInt64(int64(v.money))
		dataMSG.WriteInt64(int64(v.loss))
	}

	dataMSG.WriteInt64(int64(self.profit))

	dataMSG.WriteUInt16(uint16(len(self.hongbao_list)))
	for _, v := range self.hongbao_list {
		dataMSG.WriteUInt32(v.acc.AccountId)
		dataMSG.WriteString(v.acc.Name)
		dataMSG.WriteString(v.acc.HeadURL)
		dataMSG.WriteInt64(int64(v.acc.GetMoney()))
		dataMSG.WriteInt64(v.money)
	}
	return dataMSG
}

func (self *Room) updateEnter(accountId uint32) {
	if _, exist := self.accounts[accountId]; exist == false {
		send2hall := packet.NewPacket(nil)
		send2hall.SetMsgID(protomsg.Old_MSGID_UPDATE_ENTER.UInt16())
		send2hall.WriteUInt32(accountId)
		send2hall.WriteUInt32(self.roomId)
		send2hall.WriteUInt16(uint16(self.count()))
		send2hall.WriteUInt8(0)
		send_tools.Send2Hall(send2hall.GetData())
	}
}
func (self *Room) updateLeave(accountId uint32) {
	if _, exist := self.accounts[accountId]; exist == true {
		send2hall := packet.NewPacket(nil)
		send2hall.SetMsgID(protomsg.Old_MSGID_UPDATE_LEAVE.UInt16())
		send2hall.WriteUInt32(accountId)
		send2hall.WriteUInt32(self.roomId)
		send2hall.WriteUInt16(uint16(self.count()))
		send2hall.WriteUInt8(0)
		send_tools.Send2Hall(send2hall.GetData())
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

	// 如果玩家发了红包，暂时不能离开游戏todo
	self.leaveRoom(acc.AccountId)
}
