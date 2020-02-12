package room

import (
	"fmt"
	"math/rand"
	"root/common"
	"root/common/config"
	"root/core/log"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/game_fruitMary/account"
	"root/server/game_fruitMary/send_tools"
	"strconv"
	"time"
)

// 玩家进入游戏
func (self *Room) FRUITMARYMSG_CS_ENTER_GAME_FRUITMARY_REQ(actor int32, msg []byte, session int64) {
	enterPB := packet.PBUnmarshal(msg,&protomsg.ENTER_GAME_FRUITMARY_REQ{}).(*protomsg.ENTER_GAME_FRUITMARY_REQ)
	self.enterRoom(enterPB.GetAccountID())
}

// 玩家离开
func (self *Room) FRUITMARYMSG_CS_LEAVE_GAME_FRUITMARY_REQ(actor int32, msg []byte, session int64) {
	enterPB := packet.PBUnmarshal(msg,&protomsg.LEAVE_GAME_FRUITMARY_REQ{}).(*protomsg.LEAVE_GAME_FRUITMARY_REQ)
	ret := uint32(1)
	if self.canleave(enterPB.GetAccountID()){
		ret = 0
	}
	send_tools.Send2Account(protomsg.FRUITMARYMSG_SC_LEAVE_GAME_FRUITMARY_RES.UInt16(),&protomsg.LEAVE_GAME_FRUITMARY_RES{
		Ret:ret,
		RoomID:    self.roomId,
	},session)
}

// 玩家请求开始游戏
func (self *Room) FRUITMARYMSG_CS_START_MARY_REQ(actor int32, msg []byte, session int64) {
	start := packet.PBUnmarshal(msg,&protomsg.START_MARY_REQ{}).(*protomsg.START_MARY_REQ)
	msgBetNum := start.GetBet()
	BetNum :=msgBetNum*9
	acc := account.AccountMgr.GetAccountBySessionIDAssert(session)
	if acc.MaryCount > 0{
		log.Warnf("doBet 玩家[%v]还有小玛丽剩余次数 %v次，不能下注", acc.GetAccountId(),acc.MaryCount)
		return
	}

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
			log.Warnf("没有该档次%v",msgBetNum)
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
			BetNum = self.bets[0]*9
		}
	}
	acc.LastBet = BetNum
	gameFun := func() {
		resluts := make([]*protomsg.FRUITMARY_Result, 0)
		pArr := make([]int32, 0)
		gainFreeCount := int(0)
		sumOdds := int64(0)
		reward := int64(0)
		var feepos []*protomsg.FRUITMARYPosition
		maryCount := 0

		sumKillP := int32( config.Get_configInt("mary_room",int(self.roomId),"KillPersent")) + acc.GetKill()
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
				resluts, pArr, gainFreeCount, maryCount,sumOdds, reward,feepos = self.selectWheel(self.freeWheel, int64(BetNum), isKill,false)
			} else {
				resluts, pArr, gainFreeCount, maryCount,sumOdds, reward,feepos = self.selectWheel(self.mainWheel, int64(BetNum), isKill,false)
			}
			if maxLoop > 1 && sumOdds > 0 && i < (maxLoop-1) {
				continue
			} else {
				break
			}
		}
		if msgBetNum < uint64(self.jackLimit){
			reward = 0
		}


		acc.MaryCount = int32(maryCount)
		val := reward+sumOdds*int64(BetNum/9)
		acc.AddMoney(val, common.EOperateType_FRUIT_MARY_WIN)
		asyn_addMoney(acc.UnDevice,val,int32(self.roomId), "小玛利游戏1 中奖",nil,nil) //中奖

		log.Debugf("玩家:%v 结果->>>>>>> 身上的金币:%v 所有中奖线:%+v 一维数组:%v 获得免费次数:%v 触发小玛丽次数:%v 总赔率:%v 获得奖金：%v",
			acc.GetAccountId(),acc.GetMoney(),resluts, pArr, gainFreeCount, maryCount,sumOdds, reward)
		log.Debugf("免费和小玛利 中奖坐标:%v",feepos)

		sub := self.bonus - reward
		if sub < 0{
			self.bonus = 0
		}else{
			self.bonus = sub
		}

		// 统计玩家本局获得金币
		if isFree {
			acc.FeeCount -= 1
		}

		if gainFreeCount > 0 {
			acc.FeeCount += int32(gainFreeCount)
		}

		resultMsg := &protomsg.START_MARY_RES{
			Ret:0,
			SumOdds:sumOdds,
			Results:resluts,
			PictureList:pArr,
			Bonus:reward,
			Money:int64(acc.GetMoney()),
			FreeCount:int64(acc.FeeCount),
			MaryCount:int64(acc.MaryCount),
			FeePositions:feepos,
		}
		send_tools.Send2Account(protomsg.FRUITMARYMSG_SC_START_MARY_RES.UInt16(),resultMsg,session)

		for _,acc := range self.accounts {
			send_tools.Send2Account(protomsg.FRUITMARYMSG_SC_UPDATE_MARY_BONUS.UInt16(),&protomsg.UPDATE_MARY_BONUS{Bonus:self.bonus},acc.SessionId)
		}

		// 回存水池
		send_tools.Send2Hall(inner.SERVERMSG_GH_ROOM_BONUS_SAVE.UInt16(),&inner.ROOM_BONUS_SAVE{Value:strconv.Itoa(int(self.bonus)),RoomID:self.roomId})
	}

	//抽水的分加进水池
	if !isFree {
		back := func(backunique string, backmoney int64) { // 押注
			a := BetNum * self.jackpotRate/10000
			self.bonus += int64(a)
			if acc.GetMoney() - BetNum != uint64(backmoney){
				log.Warnf("数据错误  ->>>>>> userID:%v money:%v Bet:%v gold:%v",acc.GetUnDevice(),acc.GetMoney(),BetNum,backmoney)
				acc.AddMoney(backmoney - int64(acc.GetMoney()),common.EOperateType_INIT)
			}else{
				acc.AddMoney(int64(-(BetNum)), common.EOperateType_FRUIT_MARY_BET)
			}
			gameFun()
		}
		// 错误返回
		errback := func() {
			log.Warnf("http请求报错")
			resultMsg := &protomsg.START_MARY_RES{
				Ret:1,
			}
			send_tools.Send2Account(protomsg.FRUITMARYMSG_SC_START_MARY_RES.UInt16(),resultMsg,session)
		}
		asyn_addMoney(acc.UnDevice,-int64(BetNum),int32(self.roomId),fmt.Sprintf("水果小玛利请求下注:%v",BetNum),back,errback)
	}else{
		gameFun()
	}


}

// 玩家请求开始游戏2
func (self *Room) FRUITMARYMSG_CS_START_MARY2_REQ(actor int32, msg []byte, session int64) {
	acc := account.AccountMgr.GetAccountBySessionIDAssert(session)
	if acc.MaryCount <= 0{
		log.Warnf("maryStart 玩家[%v]小玛丽次数用完:%v ",acc.GetAccountId(), acc.MaryCount)
		return
	}

	resultList := &protomsg.START_MARY2_RES{}
	resultList.Result = make([]*protomsg.Mary2_Result,0)

	rand.Seed(time.Now().UnixNano())
	totalCount := 0
	for true {
		// 先随机一个bingo水果图案
		index := utils.RandomWeight32(self.weight_ratio,1)
		id := self.weight_ratio[index][0]
		if id > 8{
			log.Warnf("错误的随机数:%v ",id)
			return
		}
		result := &protomsg.Mary2_Result{}
		result.IndexId = id
		maryID := id
		totalCount++

		// 随机内部4个水果图案 计算4个中出现bingo水果的次数
		fourvalue := []int32{0,0,0,0}
		for i:=0;i < 4;i++{
			r := rand.Int31n(int32(len(self.maryWheel)))
			fourvalue[i] = int32(self.maryWheel[int(r)].ids[i])

		}
		result.MaryId = fourvalue
		if maryID == 8 {
			acc.MaryCount--
			zero := int32(0)
			result.Profit1 = zero
			result.Profit2 = zero
			resultList.MarySpareCount = acc.MaryCount
			resultList.Result = append(resultList.Result,result)
			break
		}

		bprofit1 := false
		for i:=0 ;i < 4;i++{
			if fourvalue[i] == maryID {
				bprofit1 = true
				break
			}
		}

		firstID := fourvalue[0]
		count := 0
		for i:=0 ;i < 4;i++{
			if fourvalue[i] == firstID{
				count++
			}else{
				break
			}
		}

		profit1 := int32(0)
		if bprofit1{
			profit1 = int32(acc.LastBet) * self.FruitRatio[maryID].Single
		}

		cr := int32(0)
		if count > 0 && firstID == maryID{
			if count < 3{
				cr = self.FruitRatio[maryID].Same_2
			}else if count == 3{
				cr = self.FruitRatio[maryID].Same_3
			}else if count == 4{
				cr = self.FruitRatio[maryID].Same_4
			}
		}
		profit2 := acc.LastBet * uint64(cr)
		result.Profit1 = profit1
		result.Profit2 = int32(profit2)
		acc.AddMoney(int64(uint64(profit1)+profit2),common.EOperateType_FRUIT_MARY2_WIN)
		asyn_addMoney(acc.UnDevice,int64(uint64(profit1)+profit2),int32(self.roomId), "小玛利游戏2 中奖",nil,nil) //中奖
		resultList.Result = append(resultList.Result,result)
	}
	acc.ResultList = resultList.Result
	send_tools.Send2Account(protomsg.FRUITMARYMSG_SC_START_MARY2_RES.UInt16(),resultList,session)
}

// 请求玩家列表
func (self *Room) FRUITMARYMSG_CS_PLAYERS_LIST_REQ(actor int32, msg []byte, session int64) {
	account.AccountMgr.GetAccountBySessionIDAssert(session)

	ret := &protomsg.PLAYERS_LIST_RES{}
	ret.Players = make([]*protomsg.AccountStorageData,0)
	for _,p := range self.accounts{
		ret.Players = append(ret.Players,p.AccountStorageData)
	}
	send_tools.Send2Account(protomsg.FRUITMARYMSG_SC_PLAYERS_LIST_RES.UInt16(),ret,session)
}

// 大厅返回水池金额
func (self *Room) SERVERMSG_HG_ROOM_BONUS_RES(actor int32, msg []byte, session int64) {
	data := packet.PBUnmarshal(msg,&inner.ROOM_BONUS_RES{}).(*inner.ROOM_BONUS_RES)
	log.Infof("大厅返回房间:[%v] 水池金额:[%v]",self.roomId,data.GetValue())
	if data.GetValue() == ""{
		self.bonus = 0
	}else{
		v,e := strconv.Atoi(data.GetValue())
		if e != nil {
			log.Errorf("解析错误%v",e.Error())
		}
		self.bonus = int64(v)
	}

}

// 大厅请求修改玩家数据
func (self *Room) SERVERMSG_HG_NOTIFY_ALTER_DATE(actor int32, msg []byte, session int64) {
	if session != 0{
		log.Warnf("此消息只能大厅发送 %v",session)
		return
	}
	data := packet.PBUnmarshal(msg,&inner.NOTIFY_ALTER_DATE{}).(*inner.NOTIFY_ALTER_DATE)
	acc := account.AccountMgr.GetAccountByIDAssert(data.GetAccountID())
	if data.GetType() == 1{ // 修改金币
		changeValue := int(data.GetAlterValue())
		if changeValue < 0 && -changeValue > int(acc.GetMoney()){
			changeValue = int(-acc.GetMoney())
		}
		acc.AddMoney(int64(changeValue),common.EOperateType(data.GetOperateType()))
	}else if data.GetType() == 2{ // 修改杀数
		acc.Kill = int32(data.GetAlterValue())
	}
}
