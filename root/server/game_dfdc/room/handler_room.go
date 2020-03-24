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
	"root/server/game_dfdc/account"
	"root/server/game_dfdc/send_tools"
)

// 玩家进入游戏
func (self *Room) DFDCMSG_CS_ENTER_GAME_DFDC_REQ(actor int32, msg []byte, session int64) {
	enterPB := packet.PBUnmarshal(msg, &protomsg.ENTER_GAME_DFDC_REQ{}).(*protomsg.ENTER_GAME_DFDC_REQ)
	self.enterRoom(enterPB.GetAccountID())
}

// 玩家离开
func (self *Room) DFDCMSG_CS_LEAVE_GAME_DFDC_REQ(actor int32, msg []byte, session int64) {
	enterPB := packet.PBUnmarshal(msg, &protomsg.LEAVE_GAME_DFDC_REQ{}).(*protomsg.LEAVE_GAME_DFDC_REQ)
	ret := uint32(1)
	if self.canleave(enterPB.GetAccountID()) {
		ret = 0
	}
	send_tools.Send2Account(protomsg.DFDCMSG_SC_LEAVE_GAME_DFDC_RES.UInt16(), &protomsg.LEAVE_GAME_DFDC_RES{
		Ret:    ret,
		RoomID: self.roomId,
	}, session)
}

// 玩家请求开始游戏
func (self *Room) DFDCMSG_CS_START_DFDC_REQ(actor int32, msg []byte, session int64) {
	start := packet.PBUnmarshal(msg, &protomsg.START_DFDC_REQ{}).(*protomsg.START_DFDC_REQ)
	BetNum := start.GetBet()

	acc := account.AccountMgr.GetAccountBySessionIDAssert(session)
	freeCount := acc.FeeCount
	isFree := false
	if freeCount <= 0 {
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
	} else {
		isFree = true
		BetNum = acc.LastBet
		if BetNum < 1 {
			BetNum = self.bets[0]
		}
	}
	acc.LastBet = BetNum
	gameFun := func() {
		pArr := make([]int32, 0)
		gainFreeCount := 0
		sumOdds := int64(0) // 总倍数
		reward := int64(0)
		msgBouns := make(map[int32]int64)
		var showpos []*protomsg.DFDCPosition
		sumKillP := int32(config.Get_configInt("dfdc_room", int(self.roomId), "KillPersent")) + acc.GetKill()
		rNum := rand.Int31n(10000) + 1
		maxLoop := 1
		if rNum <= sumKillP {
			maxLoop = 30
		}
		for i := 0; i < maxLoop; i++ {
			if freeCount > 0 {
				pArr, gainFreeCount, sumOdds, showpos = self.selectWheel(self.freeWheel, int64(BetNum))
			} else {
				pArr, gainFreeCount, sumOdds, showpos = self.selectWheel(self.mainWheel, int64(BetNum))
			}
			if maxLoop > 1 && sumOdds > 0 && i < (maxLoop-1) {
				continue
			} else {
				break
			}
		}

		if BetNum >= uint64(self.Conf_JackpotBet) {
			probability := (BetNum / self.bets[0]) * uint64(self.Conf_Bet_Probability)
			if uint64(rand.Int63n(10000)+1) <= probability {
				lv := self.GetBounusLv()
				reward = self.bonus[lv]
				self.bonus[lv] = self.bounsInitGold[lv]
				msgBouns[lv] = reward
				log.Infof("中奖池 等级:%v 奖金:%v 恢复初始值:%v ", lv, reward, self.bounsInitGold[lv])
			}
		}

		val := reward + (sumOdds * int64(BetNum) / 100)
		acc.AddMoney(val, common.EOperateType_DFDC_WIN)
		if acc.OSType == 4 {
			self.owner.AddTimer(500, 1, func(dt int64) {
				asyn_addMoney(5, self.addr_url, acc.UnDevice, val, int32(self.roomId), "多福多财 中奖", nil, nil) //中奖
			})
		}

		log.Debugf("玩家:%v 结果->>>>>>> 身上的金币:%v 一维数组:%v 获得免费次数:%v 总赔率:%v 盈利:%v 获得奖金：%v",
			acc.GetAccountId(), acc.GetMoney(), pArr, gainFreeCount, sumOdds, val, reward)
		//log.Debugf("中奖坐标:%v", showpos)

		// 统计玩家本局获得金币
		if isFree {
			acc.FeeCount -= 1
		}

		if gainFreeCount > 0 {
			acc.StaticFee += int64(gainFreeCount)
			fmt.Printf("accid:%v 免费次数累计:%v \r\n", acc.AccountId, acc.StaticFee)
			acc.FeeCount += int32(gainFreeCount)
		}

		resultMsg := &protomsg.START_DFDC_RES{
			Ret:         0,
			PictureList: pArr,
			Bonus:       msgBouns,
			Money:       int64(acc.GetMoney()),
			FreeCount:   int64(acc.FeeCount),
			Shows:       showpos,
			TotalOdds:   sumOdds,
		}
		send_tools.Send2Account(protomsg.DFDCMSG_SC_START_DFDC_RES.UInt16(), resultMsg, session)

		for _, acc := range self.accounts {
			send_tools.Send2Account(protomsg.DFDCMSG_SC_UPDATE_DFDC_BONUS.UInt16(), &protomsg.UPDATE_DFDC_BONUS{Bonus: self.bonus}, acc.SessionId)
		}

		// 回存水池
		j, e := json.Marshal(self.bonus)
		if e != nil {
			log.Warnf("回存水池:%v 错误:%v", self.bonus, e.Error())
		}
		send_tools.Send2Hall(inner.SERVERMSG_GH_ROOM_BONUS_SAVE.UInt16(), &inner.ROOM_BONUS_SAVE{Value: string(j), RoomID: self.roomId})
	}

	// 免费的直接开始，押注的先走一趟平台扣钱，扣钱成功后再开始
	if !isFree {
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
				acc.AddMoney(int64(-(BetNum)), common.EOperateType_DFDC_BET)
			}
			for lv, _ := range self.bonus {
				add := int64(BetNum) * self.bounsRoller[lv] / 10000
				self.bonus[lv] += add
			}

			self.SendBroadcast(protomsg.DFDCMSG_SC_UPDATE_DFDC_BONUS.UInt16(), &protomsg.UPDATE_DFDC_BONUS{
				Bonus: self.bonus,
			})
			gameFun()
		}
		if acc.OSType == 4 {
			// 错误返回
			errback := func() {
				log.Warnf("http请求报错")
				resultMsg := &protomsg.START_DFDC_RES{
					Ret: 1,
				}
				send_tools.Send2Account(protomsg.DFDCMSG_SC_START_DFDC_RES.UInt16(), resultMsg, session)
			}
			asyn_addMoney(5, self.addr_url, acc.UnDevice, -int64(BetNum), int32(self.roomId), fmt.Sprintf("多福多财请求下注:%v", BetNum), back, errback)
		} else {
			back("", int64(acc.GetMoney()-BetNum), 0)
		}

	} else {
		gameFun()
	}
}

// 请求玩家列表
func (self *Room) DFDCMSG_CS_PLAYERS_DFDC_LIST_REQ(actor int32, msg []byte, session int64) {
	account.AccountMgr.GetAccountBySessionIDAssert(session)

	ret := &protomsg.PLAYERS_DFDC_LIST_RES{}
	ret.Players = make([]*protomsg.AccountStorageData, 0)
	for _, p := range self.accounts {
		ret.Players = append(ret.Players, p.AccountStorageData)
	}
	send_tools.Send2Account(protomsg.DFDCMSG_SC_PLAYERS_DFDC_LIST_RES.UInt16(), ret, session)
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
