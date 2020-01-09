package room

import (
	"root/common"
	"root/common/config"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"math/rand"
	"root/protomsg"
	"root/server/hongbao/account"
	"root/server/hongbao/send_tools"
)

type (
	waitting struct {
		*Room
		s         ERoomStatus
		timestamp int64

		normal_deal []int64 // 普通红包
		bomb_deal   []int64 // 踩雷红包
		senddata    bool

		conf_floor_line   int64
		conf_ceiling_line int64
		conf_enforce      int64
	}
)

func (self *waitting) Enter(now int64) {
	self.conf_floor_line = config.GetPublicConfig_Int64("HB_FLOOR_LINE")
	self.conf_ceiling_line = config.GetPublicConfig_Int64("HB_CEILING_LINE")
	self.conf_enforce = config.GetPublicConfig_Int64("HB_ENFORCE")
	self.senddata = false
	duration := uint64(config.GetPublicConfig_Int64("HB_ROBING_TIME")) // 持续时间 秒

	log.Debugf(colorized.Blue("waitting enter duration:%v"), duration)
	self.timestamp = utils.MilliSecondTimeSince1970() + int64(duration*1000)

	if len(self.hongbao_list) == 0 {
		log.Errorf("红包列表数量为0")
		return
	}

	self.normal_deal = make([]int64, 0)
	self.bomb_deal = make([]int64, 0)
	self.profit = 0

	self.cur_hongbao = self.hongbao_list[0]
	self.hongbao_list = self.hongbao_list[1:]
	self.surplus_num = self.rob_num

	if self.cur_hongbao.acc.Robot == 0 {
		log.Debugf(colorized.Green("当前红包 ->>> 房间:%v 玩家:%v,红包金额:%v,红包数量:%v, 雷号:%v"), self.roomId, self.cur_hongbao.acc.AccountId, self.cur_hongbao.money, self.rob_num, self.cur_hongbao.bomb_num)
	} else {
		log.Debugf(colorized.Green("当前红包 ->>> 房间:%v 机器人:%v,红包金额:%v,红包数量:%v, 雷号:%v"), self.roomId, self.cur_hongbao.acc.AccountId, self.cur_hongbao.money, self.rob_num, self.cur_hongbao.bomb_num)
	}

	for _, acc := range self.accounts {
		if acc.IsOnline() == common.STATUS_OFFLINE.UInt8() {
			self.leaveRoom(acc.AccountId)
		}
	}
	begin_send := packet.NewPacket(nil)
	begin_send.SetMsgID(protomsg.Old_MSGID_HONGBAO_BEGIN.UInt16())
	begin_send.WriteInt64(self.timestamp)
	begin_send.WriteUInt32(self.cur_hongbao.acc.AccountId)
	begin_send.WriteString(self.cur_hongbao.acc.GetName())
	begin_send.WriteString(self.cur_hongbao.acc.GetHeadURL())
	begin_send.WriteInt64(self.cur_hongbao.money)   // 红包金额
	begin_send.WriteInt8(self.surplus_num)          // 可抢红包数量
	begin_send.WriteInt8(self.cur_hongbao.bomb_num) // 雷号
	self.SendBroadcast(begin_send.GetData())

	hongbaoNum := int64(self.rob_num)                                                       // 抢红包个数
	hongbao_money := int64(self.cur_hongbao.money)                                          // 抢红包总金额
	bomb_num := int8(self.cur_hongbao.bomb_num)                                             // 雷号
	newbomb_count := self.bomb_ratio_conf[utils.RandomWeight64(self.bomb_ratio_conf, 1)][0] // 生成几个红包雷

	// 水位线太低，固定2个雷
	if conf_val := self.conf_floor_line; RoomMgr.Water_line < conf_val {
		arr := utils.SplitConf2Arr_ArrInt64(config.GetPublicConfig_String("HB_EATING_RATIO"))
		index := utils.RandomWeight64(arr, 1)
		newbomb_count = arr[index][0]
	}

	rand.Seed(utils.MilliSecondTimeSince1970())
	bom, nor := self.hongbao_slice(hongbaoNum, hongbao_money, bomb_num)
	log.Debugf("房间:%v schedule 前 需要炸弹%v个 炸弹:%v,普通:%v", self.roomId, newbomb_count, bom, nor)
	bom, nor = self.schedule_bomb(bom, nor, int8(newbomb_count), bomb_num)
	self.bomb_deal = bom
	self.normal_deal = nor
	total := 0
	for _, v := range bom {
		total += int(v)
	}
	for _, v := range nor {
		total += int(v)
	}
	log.Debugf("房间:%v schedule 后 需要炸弹%v个 炸弹:%v,普通:%v  total :%v", self.roomId, newbomb_count, bom, nor, total)
	if int64(total) != hongbao_money {
		log.Warnf("出错了！！！！hongbao_money:%v", hongbao_money)
	}

	self.robot_rob_hongbao()
}

func (self *waitting) Tick(now int64) {
	curTime := utils.MilliSecondTimeSince1970()
	if curTime >= self.timestamp {
		if !self.senddata {
			self.settlement()
		}

		// 切换到等待状态
		self.switchStatus(0, ERoomStatus_STOP_BETTING)
		return
	}
}

// 结算
func (self *waitting) settlement() {

	iOldWater := RoomMgr.Water_line

	// 计算剩余红包量加入发红包人收益里
	surplus_val := int64(0)
	for _, v := range self.bomb_deal {
		surplus_val += v
	}
	for _, v := range self.normal_deal {
		surplus_val += v
	}
	self.profit += surplus_val
	self.cur_hongbao.acc.AddMoney(self.profit, 1, common.EOperateType_ROB_HONGBAO)

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_HONGBAO_SETTLEMENT.UInt16())
	send.WriteUInt16(uint16(len(self.rob_list)))

	servicepack := packet.NewPacket(nil)
	servicepack2 := packet.NewPacket(nil)
	playerCount := uint16(0)
	playerCount2 := uint16(0)
	amount_rob_val := int64(0)
	tax_scale := config.GetPublicConfig_Int64("TAX")
	conf_server_fee := config.GetPublicConfig_Int64("HB_ROB_SERVER_FEE")
	for _, v := range self.rob_list {
		amount_rob_val += v.money
		// 抽取抽水，实际获得的值
		fee_dot := conf_server_fee
		fee := (v.money * fee_dot / 100)
		if fee < 1 {
			fee = 1
		}
		final_val := v.money - fee
		if final_val < 0 {
			final_val = 0
		}
		change := final_val - v.loss
		v.acc.AddMoney(change, 2, common.EOperateType_ROB_HONGBAO)

		if fee > 1 && v.acc.Robot == 0 {
			playerCount++
			servicepack.WriteUInt32(uint32(v.acc.AccountId))
			servicepack.WriteUInt32(uint32(fee / 2 * tax_scale / 100))
		}
		if change != 0 {
			playerCount2++
			servicepack2.WriteUInt32(v.acc.AccountId)
			servicepack2.WriteInt64(int64(v.acc.GetMoney()))
			servicepack2.WriteInt64(int64(change))
			servicepack2.WriteString("")
		}
		if v.acc.Robot == 0 {
			log.Debugf(colorized.Yellow("房间:%v 玩家:%v 抢红包抢到:%v 服务费:%v fee_dot:%v final_val:%v 剩余红包:%v"), self.roomId, v.acc.AccountId, v.money, fee, fee_dot, final_val, self.surplus_num)
		} else {
			log.Debugf(colorized.Yellow("房间:%v 机器人:%v 抢红包抢到:%v 服务费:%v fee_dot:%v final_val:%v 剩余红包:%v"), self.roomId, v.acc.AccountId, v.money, fee, fee_dot, final_val, self.surplus_num)
		}

		if v.acc.Robot == 0 { // 抢钱的是玩家 水位扣除服务费
			RoomMgr.Water_line -= int64(fee / 2)
		}
	}

	if self.profit != 0 {
		playerCount2++
		servicepack2.WriteUInt32(self.cur_hongbao.acc.AccountId)
		servicepack2.WriteInt64(int64(self.cur_hongbao.acc.GetMoney()))
		servicepack2.WriteInt64(int64(self.profit))
		servicepack2.WriteString("")
	}

	if amount_rob_val != 0 && self.cur_hongbao.acc.Robot == 0 {
		playerCount++

		fee_dot := conf_server_fee
		fee := (amount_rob_val * fee_dot / 100)
		servicepack.WriteUInt32(uint32(self.cur_hongbao.acc.AccountId))
		servicepack.WriteUInt32(uint32(fee / 2 * tax_scale / 100))

		RoomMgr.Water_line -= int64(fee / 2)
	}
	for _, v := range self.rob_list {
		send.WriteUInt32(v.acc.AccountId)
		send.WriteString(v.acc.Name)
		send.WriteString(v.acc.HeadURL)
		send.WriteInt64(int64(v.acc.GetMoney()))
		send.WriteString(v.acc.Signature)
		send.WriteInt64(int64(v.money)) //抢到的金额
		send.WriteInt64(int64(v.loss))  //踩雷赔的钱
	}
	send.WriteInt64(int64(self.profit)) //发红包的人获得的钱
	self.SendBroadcast(send.GetData())

	if playerCount2 > 0 {
		updateAccount := packet.NewPacket(nil)
		updateAccount.SetMsgID(protomsg.Old_MSGID_UPDATE_ACCOUNT.UInt16())
		updateAccount.WriteUInt32(self.roomId)
		updateAccount.WriteUInt8(0)
		updateAccount.WriteUInt16(playerCount2)
		updateAccount.CatBody(servicepack2)
		send_tools.Send2Hall(updateAccount.GetData())
	}

	if playerCount > 0 {
		ser_fee := packet.NewPacket(nil)
		ser_fee.SetMsgID(protomsg.Old_MSGID_UPDATE_SERVICE_FEE.UInt16())
		ser_fee.WriteUInt8(uint8(self.gameType))
		ser_fee.WriteUInt32(uint32(self.roomId))
		ser_fee.WriteUInt16(playerCount)
		ser_fee.CatBody(servicepack)
		send_tools.Send2Hall(ser_fee.GetData())
	}

	if iOldWater != RoomMgr.Water_line {
		RoomMgr.SaveWaterLine()
	}
}
func (self *waitting) Leave(now int64) {

	log.Debugf(colorized.Blue("waitting leave\n"))
}

func (self *waitting) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.Old_MSGID_ENTER_GAME.UInt16(): // 客户端链接进入游戏
		self.Old_MSGID_ENTER_GAME(actor, msg, session)
	case protomsg.Old_MSGID_HONGBAO_ROB_HONGBAO.UInt16(): // 抢红包
		self.Old_MSGID_HONGBAO_ROB_HONGBAO(actor, msg, session)
	default:
		log.Warnf("waitting 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}

	return true
}

// 进入游戏
func (self *waitting) Old_MSGID_ENTER_GAME(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	_ = pack.ReadUInt32()

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_ENTER_GAME.UInt16())
	//if ret := self.canEnterRoom(accountId); ret > 0 {
	//	send.WriteUInt8(uint8(ret))
	//	send_tools.Send2Account(send.GetData(), session)
	//	return
	//}

	self.enterRoom(accountId)

	// 通知客户端，进入游戏成功
	acc := account.AccountMgr.GetAccountByID(accountId)
	send2c := packet.NewPacket(nil)
	send2c.SetMsgID(protomsg.Old_MSGID_ENTER_GAME.UInt16())
	send2c.WriteUInt8(0)
	send2c.WriteUInt32(self.roomId)
	send_tools.Send2Account(send2c.GetData(), acc.SessionId)

	send2acc := self.sendGameData(acc, int64(self.timestamp))
	send_tools.Send2Account(send2acc.GetData(), acc.SessionId)
}

func (self *waitting) get_nor_hongbao() int64 {
	var val int64
	if len(self.normal_deal) > 0 {
		index := utils.Randx_y(0, len(self.normal_deal))
		val = self.normal_deal[index]
		self.normal_deal = append(self.normal_deal[:index], self.normal_deal[index+1:]...)

	} else {
		index := utils.Randx_y(0, len(self.bomb_deal))
		val = self.bomb_deal[index]
		self.bomb_deal = append(self.bomb_deal[:index], self.bomb_deal[index+1:]...)

	}
	return val
}
func (self *waitting) get_bomb_hongbao() int64 {
	var val int64
	if len(self.bomb_deal) > 0 {
		index := utils.Randx_y(0, len(self.bomb_deal))
		val = self.bomb_deal[index]
		self.bomb_deal = append(self.bomb_deal[:index], self.bomb_deal[index+1:]...)

	} else {
		index := utils.Randx_y(0, len(self.normal_deal))
		val = self.normal_deal[index]
		self.normal_deal = append(self.normal_deal[:index], self.normal_deal[index+1:]...)

	}
	return val
}

// 计算获得红包
func (self *waitting) hongbao(acc *account.Account) int64 {
	val := int64(0)

	rand_fun := func() {
		bombNum := len(self.bomb_deal)
		NormalNum := len(self.normal_deal)
		bomb := 100 * bombNum / (bombNum + NormalNum)
		if utils.Probability(bomb) {
			val = self.get_bomb_hongbao()
		} else {
			val = self.get_nor_hongbao()
		}
	}
	// 水位线高于高位让玩家赢, 低于低位让玩家输
	if RoomMgr.Water_line > self.conf_ceiling_line {
		if acc.Robot == 0 {
			val = self.get_nor_hongbao()
		} else {
			val = self.get_bomb_hongbao()
		}

	} else if RoomMgr.Water_line < self.conf_floor_line {
		if acc.Robot == 0 {
			val = self.get_bomb_hongbao()
		} else {
			// 如果玩家发的普通红包都被抢完了，又需要收水的时候，尝试去调整红包数量，让机器人尽量不去踩雷
			if self.cur_hongbao.acc.Robot == 0 && len(self.normal_deal) == 0 && len(self.bomb_deal) > 1 {
				log.Debugf("水位线低于地位，玩家发的红包，普通红包领完，剩余：%v颗雷，调整2颗雷，避免机器人踩雷", len(self.bomb_deal))
				self.bomb_deal[0] -= 1
				self.bomb_deal[1] += 1
				self.normal_deal = append(self.normal_deal, self.bomb_deal[:2]...)
				self.bomb_deal = self.bomb_deal[2:]

				log.Debugf("调整后: 普通红包:%v  雷红包:%v", self.normal_deal, self.bomb_deal)
				val = self.get_nor_hongbao()

			} else if self.cur_hongbao.acc.Robot != 0 && utils.Probability(int(100-self.conf_enforce)) {
				rand_fun()
			} else {
				val = self.get_nor_hongbao()
			}

		}
	} else {
		rand_fun()
	}

	if self.cur_hongbao.acc.Robot == 0 && acc.Robot != 0 { // 发钱的是玩家，抢钱的机器人，水位线涨
		RoomMgr.Water_line += val
	} else if self.cur_hongbao.acc.Robot != 0 && acc.Robot == 0 { // 发钱的是机器人，抢钱的玩家，水位线降
		RoomMgr.Water_line -= val
	}

	return val
}

// 抢红包操作
func (self *waitting) Old_MSGID_HONGBAO_ROB_HONGBAO(actor int32, msg []byte, session int64) {
	pack := packet.NewPacket(msg)
	accountId := pack.ReadUInt32()
	acc := account.AccountMgr.GetAccountByID(accountId)
	if acc == nil {
		return
	}

	account.CheckSession(accountId, session)

	send := packet.NewPacket(nil)
	send.SetMsgID(protomsg.Old_MSGID_HONGBAO_ROB_HONGBAO.UInt16())
	if self.surplus_num <= 0 {
		send.WriteUInt8(1)
		send_tools.Send2Account(send.GetData(), session)
		return
	}

	if accountId == self.cur_hongbao.acc.AccountId {
		return
	}
	ratio := self.GetParamInt(5)
	loss := (self.cur_hongbao.money * int64(ratio)) / 100
	// 钱不够赔
	if acc.GetMoney() < uint64(loss) {
		send.WriteUInt8(2)
		send_tools.Send2Account(send.GetData(), session)
		return
	}
	for _, v := range self.rob_list {
		if v.acc.AccountId == accountId {
			return
		}
	}

	send.WriteUInt8(0)

	self.surplus_num--

	// 计算抢到的红包值
	val := self.hongbao(acc)
	send.WriteInt64(val)

	rob := &Rob{acc: acc, money: val}
	// 计算是否踩雷
	if int8(val%10) == self.cur_hongbao.bomb_num {
		rob.loss = loss

		if self.cur_hongbao.acc.Robot == 0 && acc.Robot != 0 { // 发钱的是玩家，抢钱的机器人，水位线涨
			RoomMgr.Water_line -= loss
		} else if self.cur_hongbao.acc.Robot != 0 && acc.Robot == 0 { // 发钱的是机器人，抢钱的玩家，水位线降
			RoomMgr.Water_line += loss
		}

		if acc.Robot == 0 {
			log.Debugf(colorized.White("玩家:%v 踩雷, loss:%v 水位线:%v "), acc.AccountId, loss, RoomMgr.Water_line)
		} else {
			log.Debugf(colorized.White("机器人:%v 踩雷, loss:%v 水位线:%v "), acc.AccountId, loss, RoomMgr.Water_line)
		}
		send.WriteInt64(loss)
		self.profit += loss
	} else {
		send.WriteInt64(0)
	}

	send_tools.Send2Account(send.GetData(), session)

	// 玩家赔的钱计算进入发红包人的收益里

	send_broadcast := packet.NewPacket(nil)
	send_broadcast.SetMsgID(protomsg.Old_MSGID_HONGBAO_BROADCAST_HONGBAO.UInt16())
	send_broadcast.WriteUInt32(acc.AccountId)
	send_broadcast.WriteString(acc.Name)
	send_broadcast.WriteString(acc.HeadURL)
	send_broadcast.WriteInt64(int64(acc.GetMoney()))
	send_broadcast.WriteString(acc.Signature)
	send_broadcast.WriteInt8(self.surplus_num)
	self.SendBroadcast(send_broadcast.GetData())

	auto_qiang := false
	self.rob_list = append(self.rob_list, rob)
	if self.surplus_num == 0 {
		self.settlement()
		self.senddata = true
	} else if self.surplus_num == 2 && RoomMgr.Water_line < self.conf_floor_line { // 如果红包还有2个，并且都是雷,随机概率的把雷消除掉

		// 直接消除2颗雷
		if len(self.bomb_deal) == 2 {
			self.bomb_deal[0] += 2
			self.bomb_deal[1] -= 2
			auto_qiang = true
		} else if len(self.bomb_deal) == 1 {
			auto_qiang = true
			self.bomb_deal[0] += 1
			for k, v := range self.normal_deal {
				if int8((v+1)%10) != self.cur_hongbao.bomb_num {
					self.normal_deal[k] += 1
					break
				}
			}
		}

		self.normal_deal = append(self.normal_deal, self.bomb_deal...)
		self.bomb_deal = make([]int64, 0)
	}

	if auto_qiang {
		count := 2
		for _, robot := range self.Robots() {
			if count == 0 {
				break
			}
			ex := false
			for _, v := range self.rob_list {
				if v.acc.AccountId == robot.AccountId {
					ex = true
					break
				}
			}
			if ex {
				continue
			}

			robmsg := packet.NewPacket(nil)
			robmsg.SetMsgID(protomsg.Old_MSGID_HONGBAO_ROB_HONGBAO.UInt16())
			robmsg.WriteUInt32(robot.AccountId)
			self.Old_MSGID_HONGBAO_ROB_HONGBAO(actor, robmsg.GetData(), 0)
			count--
		}
	}

}
