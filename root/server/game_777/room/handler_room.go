package room

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"root/common"
	"root/common/config"
	"root/core/log"
	"root/core/packet"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/game_777/account"
	"root/server/game_777/send_tools"
	"root/server/platform"
)

// 玩家进入游戏
func (self *Room) S777MSG_CS_ENTER_GAME_S777_REQ(actor int32, msg []byte, session int64) {
	enterPB := packet.PBUnmarshal(msg, &protomsg.ENTER_GAME_S777_REQ{}).(*protomsg.ENTER_GAME_S777_REQ)
	self.enterRoom(enterPB.GetAccountID())
}

// 玩家离开
func (self *Room) S777MSG_CS_LEAVE_GAME_S777_REQ(actor int32, msg []byte, session int64) {
	enterPB := packet.PBUnmarshal(msg, &protomsg.LEAVE_GAME_S777_REQ{}).(*protomsg.LEAVE_GAME_S777_REQ)
	ret := uint32(1)
	if self.canleave(enterPB.GetAccountID()) {
		ret = 0
	}
	send_tools.Send2Account(protomsg.S777MSG_SC_LEAVE_GAME_S777_RES.UInt16(), &protomsg.LEAVE_GAME_S777_RES{
		Ret:    ret,
		RoomID: self.roomId,
	}, session)
}

// 玩家请求开始游戏
func (self *Room) S777MSG_CS_START_S777_REQ(actor int32, msg []byte, session int64) {
	start := packet.PBUnmarshal(msg, &protomsg.START_S777_REQ{}).(*protomsg.START_S777_REQ)
	BetNum := start.GetBet()

	acc := account.AccountMgr.GetAccountBySessionIDAssert(session)
	isFind := false
	for i := 0; i < len(self.bets); i++ {
		if BetNum == self.bets[i] {
			isFind = true
			break
		}
	}
	if !isFind {
		log.Warnf("没有该档次%v", BetNum)
		return
	}

	if acc.GetMoney() < BetNum {
		log.Warnf("玩家积分不够了")
		return
	}
	acc.LastBet = BetNum
	gameFun := func() {
		pArr := make([]int32, 0)
		sumOdds := int64(0) // 总倍数
		msgBouns := make(map[int32]int64)
		jackpotlv := 0
		bigwinID := int32(0)
		sumKillP := int32(config.Get_configInt("777_room", int(self.roomId), "KillPersent")) + acc.GetKill()
		rNum := rand.Int31n(10000) + 1
		maxLoop := 1
		if rNum <= sumKillP {
			maxLoop = 30
		}
		for i := 0; i < maxLoop; i++ {
			pArr, sumOdds, bigwinID, jackpotlv = self.selectWheel(self.mainWheel, int64(BetNum))
			if maxLoop > 1 && sumOdds > 0 && i < (maxLoop-1) {
				continue
			} else {
				break
			}
		}

		val := sumOdds * int64(BetNum) / 10000
		var jackpotval int64
		if BetNum >= uint64(self.Conf_JackpotBet) && jackpotlv != 0 {
			jackpotval = self.bonus[int32(jackpotlv)]
			jackpotval += self.bounsInitGold[int32(jackpotlv)] * int64(BetNum)
			val += jackpotval
			self.bonus[int32(jackpotlv)] = 0
			log.Infof("中奖池 lv:%v val:%v 奖池初始化:%v ", jackpotlv, jackpotval, self.bonus[int32(jackpotlv)])
		}
		acc.AddMoney(val, common.EOperateType_S777_WIN)
		if acc.OSType == 4 {
			self.owner.AddTimer(500, 1, func(dt int64) {
				platform.Asyn_addMoney(5, self.addr_url, acc.UnDevice, val, int32(self.roomId), "game_777", "777 中奖", nil, nil) //中奖
			})
		}

		log.Debugf("玩家:%v 结果->>>>>>> 身上的金币:%v 一维数组:%v 总赔率:%v 盈利:%v bigwinID:%v jackpotVal:%v",
			acc.GetAccountId(), acc.GetMoney(), pArr, sumOdds, val, bigwinID, jackpotval)

		resultMsg := &protomsg.START_S777_RES{
			Ret:         0,
			PictureList: pArr,
			Bonus:       msgBouns,
			Money:       int64(acc.GetMoney()),
			TotalOdds:   sumOdds,
			Id:          protomsg.JackPotID(bigwinID),
		}
		send_tools.Send2Account(protomsg.S777MSG_SC_START_S777_RES.UInt16(), resultMsg, session)

		for _, acc := range self.accounts {
			send_tools.Send2Account(protomsg.S777MSG_SC_UPDATE_S777_BONUS.UInt16(), &protomsg.UPDATE_S777_BONUS{Bonus: self.bonus}, acc.SessionId)
		}

		// 回存水池
		j, e := json.Marshal(self.bonus)
		if e != nil {
			log.Warnf("回存水池:%v 错误:%v", self.bonus, e.Error())
		}
		send_tools.Send2Hall(inner.SERVERMSG_GH_ROOM_BONUS_SAVE.UInt16(), &inner.ROOM_BONUS_SAVE{Value: string(j), RoomID: self.roomId})
	}

	// 免费的直接开始，押注的先走一趟平台扣钱，扣钱成功后再开始
	back := func(backunique string, backmoney int64, bwType int32) { // 押注
		if bwType == 1 {
			acc.Kill = int32(config.GetPublicConfig_Int64(4))
			log.Infof("acc:%v 三方黑名单 杀数为:%v ", acc.GetAccountId(), acc.Kill)
		} else if bwType == 2 {
			acc.Kill = int32(config.GetPublicConfig_Int64(5))
			log.Infof("acc:%v 三方白名单 杀数为:%v ", acc.GetAccountId(), acc.Kill)
		} else if bwType == 0 {
			acc.Kill = 0
			log.Infof("acc:%v bwType:0 ", acc.GetAccountId())
		}

		//log.Infof("玩家:%v 下注成功 扣除:%v ", acc.GetUnDevice(), BetNum)
		if acc.GetMoney()-BetNum != uint64(backmoney) {
			log.Warnf("数据错误  ->>>>>> userID:%v money:%v Bet:%v gold:%v", acc.GetUnDevice(), acc.GetMoney(), BetNum, backmoney)
			acc.AddMoney(backmoney-int64(acc.GetMoney()), common.EOperateType_INIT)
		} else {
			acc.AddMoney(int64(-(BetNum)), common.EOperateType_S777_BET)
		}
		for lv, _ := range self.bonus {
			add := int64(BetNum) * self.bounsRoller[lv] / 10000
			self.bonus[lv] += add
		}

		self.SendBroadcast(protomsg.S777MSG_SC_UPDATE_S777_BONUS.UInt16(), &protomsg.UPDATE_S777_BONUS{
			Bonus: self.bonus,
		})
		gameFun()
	}
	if acc.OSType == 4 {
		// 错误返回
		errback := func() {
			log.Warnf("accid:%v http请求报错！！！！！！！！！！！", acc.AccountId)
			resultMsg := &protomsg.START_S777_RES{
				Ret: 1,
			}
			send_tools.Send2Account(protomsg.S777MSG_SC_START_S777_RES.UInt16(), resultMsg, session)
		}
		platform.Asyn_addMoney(5, self.addr_url, acc.UnDevice, -int64(BetNum), int32(self.roomId), "game_777", fmt.Sprintf("多福多财请求下注:%v", BetNum), back, errback)
	} else {
		back("", int64(acc.GetMoney()-BetNum), 0)
	}
}

// 请求玩家列表
func (self *Room) S777MSG_CS_PLAYERS_S777_LIST_REQ(actor int32, msg []byte, session int64) {
	account.AccountMgr.GetAccountBySessionIDAssert(session)

	ret := &protomsg.PLAYERS_S777_LIST_RES{}
	ret.Players = make([]*protomsg.AccountStorageData, 0)
	for _, p := range self.accounts {
		ret.Players = append(ret.Players, p.AccountStorageData)
	}
	send_tools.Send2Account(protomsg.S777MSG_SC_PLAYERS_S777_LIST_RES.UInt16(), ret, session)
}

// 大厅返回水池金额
func (self *Room) SERVERMSG_HG_ROOM_BONUS_RES(actor int32, msg []byte, session int64) {
	data := packet.PBUnmarshal(msg, &inner.ROOM_BONUS_RES{}).(*inner.ROOM_BONUS_RES)
	log.Infof("大厅返回房间:[%v] 水池金额:[%v]", self.roomId, data.GetValue())
	if data.GetValue() == "" {
		return
	}
	new := map[int32]int64{}
	e := json.Unmarshal([]byte(data.GetValue()), &new)
	if e != nil {
		log.Errorf("解析错误%v", e.Error())
	} else {
		self.bonus = new
	}

}

// 大厅请求修改玩家数据
func (self *Room) SERVERMSG_HG_NOTIFY_ALTER_DATE(actor int32, msg []byte, session int64) {
	if session != 0 {
		log.Warnf("此消息只能大厅发送 %v", session)
		return
	}
	data := packet.PBUnmarshal(msg, &inner.NOTIFY_ALTER_DATE{}).(*inner.NOTIFY_ALTER_DATE)
	acc := account.AccountMgr.GetAccountByIDAssert(data.GetAccountID())
	if data.GetType() == 1 { // 修改金币
		changeValue := int(data.GetAlterValue())
		if changeValue < 0 && -changeValue > int(acc.GetMoney()) {
			changeValue = int(-acc.GetMoney())
		}
		acc.AddMoney(int64(changeValue), common.EOperateType(data.GetOperateType()))
	} else if data.GetType() == 2 { // 修改杀数
		acc.Kill = int32(data.GetAlterValue())
	}
}
