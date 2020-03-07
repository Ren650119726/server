package room

import (
	"fmt"
	"math/rand"
	"root/common"
	"root/common/config"
	"root/core/log"
	"root/core/packet"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/game_luckfruit/account"
	"root/server/game_luckfruit/send_tools"
	"strconv"
)

// 玩家进入游戏
func (self *Room) LUCKFRUITMSG_CS_ENTER_GAME_LUCKFRUIT_REQ(actor int32, msg []byte, session int64) {
	enterPB := packet.PBUnmarshal(msg, &protomsg.ENTER_GAME_LUCKFRUIT_REQ{}).(*protomsg.ENTER_GAME_LUCKFRUIT_REQ)
	self.enterRoom(enterPB.GetAccountID())
}

// 玩家离开
func (self *Room) LUCKFRUITMSG_CS_LEAVE_GAME_LUCKFRUIT_REQ(actor int32, msg []byte, session int64) {
	enterPB := packet.PBUnmarshal(msg, &protomsg.LEAVE_GAME_LUCKFRUIT_REQ{}).(*protomsg.LEAVE_GAME_LUCKFRUIT_REQ)
	ret := uint32(1)
	if self.canleave(enterPB.GetAccountID()) {
		ret = 0
	}
	send_tools.Send2Account(protomsg.LUCKFRUITMSG_SC_LEAVE_GAME_LUCKFRUIT_RES.UInt16(), &protomsg.LEAVE_GAME_LUCKFRUIT_RES{
		Ret:    ret,
		RoomID: self.roomId,
	}, session)
}

// 玩家请求开始游戏
func (self *Room) LUCKFRUITMSG_CS_START_LUCKFRUIT_REQ(actor int32, msg []byte, session int64) {
	start := packet.PBUnmarshal(msg, &protomsg.START_LUCKFRUIT_REQ{}).(*protomsg.START_LUCKFRUIT_REQ)
	msgBetNum := start.GetBet()
	BetNum := msgBetNum * 9
	acc := account.AccountMgr.GetAccountBySessionIDAssert(session)

	freeCount := acc.FeeCount
	isFree := false
	if freeCount <= 0 {
		isFind := false
		for i := 0; i < len(self.bets); i++ {
			if msgBetNum == self.bets[i] {
				isFind = true
				break
			}
		}
		if !isFind {
			log.Warnf("没有该档次%v", msgBetNum)
			return
		}

		if acc.GetMoney() < BetNum {
			log.Warnf("玩家积分不够了")
			return
		}
	} else {
		isFree = true
		BetNum = acc.LastBet
		if BetNum < 1 {
			BetNum = self.bets[0] * 9
		}
	}
	acc.LastBet = BetNum
	gameFun := func() {
		resluts := make([]*protomsg.LUCKFRUIT_Result, 0)
		pArr := make([]int32, 0)
		gainFreeCount := int(0)
		sumOdds := int64(0)
		reward := int64(0)
		var feepos []*protomsg.LUCKFRUITPosition

		sumKillP := int32(config.Get_configInt("luckfruit_room", int(self.roomId), "KillPersent")) + acc.GetKill()
		//log.Debugf("玩家的 杀数为: %d", sumKillP)
		rNum := rand.Int31n(10000) + 1
		isKill := false
		maxLoop := 1
		if rNum <= sumKillP {
			isKill = true
			maxLoop = 30
		}
		for i := 0; i < maxLoop; i++ {
			if freeCount > 0 {
				resluts, pArr, gainFreeCount, _, sumOdds, reward = self.selectWheel(self.freeWheel, int64(BetNum), isKill, false)
			} else {
				resluts, pArr, gainFreeCount, _, sumOdds, reward = self.selectWheel(self.mainWheel, int64(BetNum), isKill, false)
			}
			if maxLoop > 1 && sumOdds > 0 && i < (maxLoop-1) {
				continue
			} else {
				break
			}
		}
		if msgBetNum < uint64(self.jackLimit) {
			reward = 0
		}

		val := reward + (sumOdds * int64(BetNum) / 9)
		acc.AddMoney(val, common.EOperateType_LUCKFRUIT_WIN)
		if acc.GetOSType() == 4 {
			asyn_addMoney(self.addr_url, acc.UnDevice, val, int32(self.roomId), "小玛利游戏1 中奖", nil, nil) //中奖
		}

		log.Debugf("玩家:%v 结果->>>>>>> 身上的金币:%v 所有中奖线:%+v 一维数组:%v 获得免费次数:%v 总赔率:%v 获得奖金：%v",
			acc.GetAccountId(), acc.GetMoney(), resluts, pArr, gainFreeCount, sumOdds, reward)

		sub := self.bonus - reward
		if sub < 0 {
			self.bonus = 0
		} else {
			self.bonus = sub
		}

		// 统计玩家本局获得金币
		if isFree {
			acc.FeeCount -= 1
		}

		if gainFreeCount > 0 {
			acc.FeeCount += int32(gainFreeCount)
		}

		resultMsg := &protomsg.START_LUCKFRUIT_RES{
			Ret:          0,
			SumOdds:      sumOdds,
			Results:      resluts,
			PictureList:  pArr,
			Bonus:        reward,
			Money:        int64(acc.GetMoney()),
			FreeCount:    int64(acc.FeeCount),
			FeePositions: feepos,
		}
		send_tools.Send2Account(protomsg.LUCKFRUITMSG_SC_START_LUCKFRUIT_RES.UInt16(), resultMsg, session)

		for _, acc := range self.accounts {
			send_tools.Send2Account(protomsg.LUCKFRUITMSG_SC_UPDATE_LUCKFRUIT_BONUS.UInt16(), &protomsg.UPDATE_LUCKFRUIT_BONUS{Bonus: self.bonus}, acc.SessionId)
		}

		// 回存水池
		send_tools.Send2Hall(inner.SERVERMSG_GH_ROOM_BONUS_SAVE.UInt16(), &inner.ROOM_BONUS_SAVE{Value: strconv.Itoa(int(self.bonus)), RoomID: self.roomId})
	}

	//抽水的分加进水池
	if !isFree {
		back := func(backunique string, backmoney int64) { // 押注
			a := BetNum * self.jackpotRate / 10000
			self.bonus += int64(a)
			if acc.GetMoney()-BetNum != uint64(backmoney) {
				log.Warnf("数据错误  ->>>>>> userID:%v money:%v Bet:%v gold:%v", acc.GetUnDevice(), acc.GetMoney(), BetNum, backmoney)
				acc.AddMoney(backmoney-int64(acc.GetMoney()), common.EOperateType_INIT)
			} else {
				acc.AddMoney(int64(-(BetNum)), common.EOperateType_LUCKFRUIT_BET)
			}
			gameFun()
		}
		if acc.GetOSType() == 4 {
			// 错误返回
			errback := func() {
				log.Warnf("http请求报错")
				resultMsg := &protomsg.START_LUCKFRUIT_RES{
					Ret: 1,
				}
				send_tools.Send2Account(protomsg.LUCKFRUITMSG_SC_START_LUCKFRUIT_RES.UInt16(), resultMsg, session)
			}
			asyn_addMoney(self.addr_url, acc.UnDevice, -int64(BetNum), int32(self.roomId), fmt.Sprintf("金瓶梅请求下注:%v", BetNum), back, errback)
		} else {
			back("", int64(acc.GetMoney()-BetNum))
		}
	} else {
		gameFun()
	}
}

// 请求玩家列表
func (self *Room) LUCKFRUITMSG_CS_PLAYERS_LUCKFRUIT_LIST_REQ(actor int32, msg []byte, session int64) {
	account.AccountMgr.GetAccountBySessionIDAssert(session)

	ret := &protomsg.PLAYERS_LUCKFRUIT_LIST_RES{}
	ret.Players = make([]*protomsg.AccountStorageData, 0)
	for _, p := range self.accounts {
		ret.Players = append(ret.Players, p.AccountStorageData)
	}
	send_tools.Send2Account(protomsg.LUCKFRUITMSG_SC_PLAYERS_LUCKFRUIT_LIST_RES.UInt16(), ret, session)
}

// 大厅返回水池金额
func (self *Room) SERVERMSG_HG_ROOM_BONUS_RES(actor int32, msg []byte, session int64) {
	data := packet.PBUnmarshal(msg, &inner.ROOM_BONUS_RES{}).(*inner.ROOM_BONUS_RES)
	log.Infof("大厅返回房间:[%v] 水池金额:[%v]", self.roomId, data.GetValue())
	if data.GetValue() == "" {
		self.bonus = 0
	} else {
		v, e := strconv.Atoi(data.GetValue())
		if e != nil {
			log.Errorf("解析错误%v", e.Error())
		}
		self.bonus = int64(v)
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
