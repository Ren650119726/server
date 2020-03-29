package room

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"root/common"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/game_lhd/account"
	"root/server/game_lhd/algorithm"
	"root/server/platform"
)

type (
	betting struct {
		*Room
		s                        ERoomStatus
		start_timestamp          int64
		end_timestamp            int64
		enterMsg                 *protomsg.StatusMsgLHD
		interval_broadcast_timer int64 // 间隔广播下注缓存
		betTimer                 int64

		bets_cache      []*protomsg.BET_LHD_RES_BetPlayer // 下注缓存
		cd              map[uint32]int64                  // 玩家最后一次下注时间戳
		forbidBetplayer map[uint32]bool                   // 禁止下注的玩家
		robots          map[uint32]*behavior
	}
)

func (self *betting) Enter(now int64) {
	duration := self.status_duration[self.s]
	self.start_timestamp = utils.MilliSecondTimeSince1970()
	self.end_timestamp = self.start_timestamp + duration
	self.log(colorized.Yellow("betting enter duration:%v"), duration)

	self.cd = make(map[uint32]int64)
	self.forbidBetplayer = make(map[uint32]bool)
	self.robots = make(map[uint32]*behavior)
	// 如果没有牌靴了，就随机一组牌靴
	if len(self.GameCards) <= 2 {
		num := utils.Randx_y(100, 350)
		if num%2 == 1 {
			num++
		}
		self.GameCards = algorithm.GetRandom_Card(self.RoomCards, num)
		self.history = self.history[:0]
		self.SendBroadcast(protomsg.LHDMSG_SC_CLEAN_HISTORY_BROADCAST_LHD.UInt16(), nil)
	} else {
		self.GameCards = self.GameCards[2:]
	}

	self.log("开始下注 靴牌剩余:%v张，本局牌:%+v ", len(self.GameCards), self.GameCards[:2])

	self.bets_cache = make([]*protomsg.BET_LHD_RES_BetPlayer, 0)
	// 广播房间玩家，切换状态
	bet, err := proto.Marshal(&protomsg.Status_Bet{
		//todo .....................................................
	})
	if err != nil {
		log.Panicf("错误:%v ", err.Error())
	}

	betval := self.areaBetVal(true)
	for _, acc := range self.accounts {
		acc.Betcount = 0
		if acc.GetMoney() < uint64(self.betlimit_conf) {
			self.forbidBetplayer[acc.AccountId] = true
		}
		self.initRobotBehavior(acc)
	}

	self.enterMsg = &protomsg.StatusMsgLHD{
		Status:           protomsg.LHDGAMESTATUS(self.s),
		Status_StartTime: uint64(self.start_timestamp),
		Status_EndTime:   uint64(self.end_timestamp),
		AreaBetVal:       betval,
		AreaBetVal_Own:   map[int32]int64{},
		Status_Data:      bet,
	}
	self.SendBroadcast(protomsg.LHDMSG_SC_SWITCH_GAME_STATUS_BROADCAST_LHD.UInt16(), &protomsg.SWITCH_GAME_STATUS_BROADCAST_LHD{self.enterMsg})

	self.interval_broadcast_timer = self.owner.AddTimer(500, -1, self.updateBetPlayers)
	self.betTimer = self.owner.AddTimer(100, -1, self.robotbet)
}

func (self *betting) updateBetPlayers(now int64) {
	if len(self.bets_cache) != 0 {
		betval := self.areaBetVal(true)
		self.SendBroadcast(protomsg.LHDMSG_SC_BET_LHD_RES.UInt16(), &protomsg.BET_LHD_RES{
			Players:    self.bets_cache,
			AreaBetVal: betval,
		})
		self.bets_cache = self.bets_cache[:0]
	}
}
func (self *betting) initRobotBehavior(acc *account.Account) {
	// 初始化机器人行为表
	if acc.Robot != 0 {
		var (
			area       protomsg.LHDAREA
			areacount  int32
			peace      bool
			peacecount int32
		)
		beti := utils.RandomWeight32(self.robot_conf.BetWeight, 1)
		areai := utils.RandomWeight32(self.robot_conf.AreaWeight, 1)
		area = protomsg.LHDAREA(self.robot_conf.AreaWeight[areai][0])
		if area == protomsg.LHDAREA_LHD_AREA_DRAGON {
			areacount = int32(utils.Randx_y(int(self.robot_conf.DragonRandCount[0]), int(self.robot_conf.DragonRandCount[1])))
		} else if area == protomsg.LHDAREA_LHD_AREA_TIGER {
			areacount = int32(utils.Randx_y(int(self.robot_conf.TigerRandCount[0]), int(self.robot_conf.TigerRandCount[1])))
		}
		peace = utils.Probability10000(self.robot_conf.PeaceRatio)
		if peace {
			peacecount = int32(utils.Randx_y(int(self.robot_conf.PeaceCount[0]), int(self.robot_conf.PeaceCount[1])))
		}

		self.robots[acc.AccountId] = &behavior{
			bet:         uint64(self.bets_conf[uint64(self.robot_conf.BetWeight[beti][0])]),
			area:        area,
			areacount:   areacount,
			peace:       peace,
			peacecount:  peacecount,
			nextBetTime: utils.MilliSecondTimeSince1970() + int64(utils.Randx_y(int(self.robot_conf.BetFrequencies[0]), int(self.robot_conf.BetFrequencies[1]))),
		}
		log.Infof("初始化机器人行为 %v %+v:", acc.AccountId, self.robots[acc.AccountId])
	}
}
func (self *betting) Tick(now int64) {
	if now >= self.end_timestamp {
		self.owner.CancelTimer(self.interval_broadcast_timer)
		self.updateBetPlayers(now) // 把剩余的发给客户端
		self.switchStatus(now, ERoomStatus_STOP_BETTING)
		return
	}
}

func (self *betting) leave(accid uint32) bool {
	_, exist := self.betPlayers[accid]
	// 如果 玩家有押注，不能退出游戏
	if exist {
		return false
	}
	return true
}

func (self *betting) enterData(accountId uint32) *protomsg.StatusMsgLHD {
	betval_own := self.playerAreaBetVal(accountId)
	self.enterMsg.AreaBetVal_Own = betval_own
	self.enterMsg.AreaBetVal = self.areaBetVal(true)
	return self.enterMsg
}

func (self *betting) Leave(now int64) {
	self.owner.CancelTimer(self.betTimer)
	self.log(colorized.Yellow("betting leave\n"))
	self.log(colorized.Blue(""))
}

func (self *betting) Handle(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.LHDMSG_CS_BET_LHD_REQ.UInt16(): // 请求下注
		self.LHDMSG_CS_BET_LHD_REQ(actor, pack.ReadBytes(), session)
	case protomsg.LHDMSG_CS_CLEAN_BET_LHD_REQ.UInt16(): // 请求清空下注
		self.LHDMSG_CS_CLEAN_BET_LHD_REQ(actor, pack.ReadBytes(), session)
	default:
		log.Warnf("betting 状态 没有处理消息msgId:%v", pack.GetMsgID())
		return false
	}

	return true
}

func (self *betting) LHDMSG_CS_BET_LHD_REQ(actor int32, msg []byte, session int64) {
	betdata := packet.PBUnmarshal(msg, &protomsg.BET_LHD_REQ{}).(*protomsg.BET_LHD_REQ)
	var acc *account.Account
	now := utils.MilliSecondTimeSince1970()
	if session == 0 {
		acc = account.AccountMgr.GetAccountByIDAssert(betdata.GetAccountID())
	} else {
		acc = account.AccountMgr.GetAccountBySessionIDAssert(session)
	}

	if acc.GetMoney() < betdata.GetBet() {
		log.Warnf("acc:%v room:%v 钱不够下注 身上钱:%v 请求下注:%v ", acc.AccountId, self.roomId, acc.GetMoney(), betdata.GetBet())
		return
	}
	if self.forbidBetplayer[acc.AccountId] {
		log.Warnf("acc:%v room:%v 钱不够下注 身上钱:%v 低于bet_limit 请求下注失败:%v ", acc.AccountId, self.roomId, acc.GetMoney(), betdata.GetBet())
		return
	}
	if last := self.cd[acc.GetAccountId()]; now-last < self.interval_conf && betdata.BetType == 0 {
		log.Warnf("还未到押注cd")
		return
	}
	if now-acc.CLeanTime < 3000 {
		log.Warnf("acc:%v 还未收到三方的清除消息，不能押注:%v ", acc.GetAccountId(), now-acc.CLeanTime)
		return
	}

	check := false
	for _, betVal := range self.bets_conf {
		if uint64(betVal) == betdata.Bet {
			check = true
			break
		}
	}
	if !check {
		log.Warnf("acc:%v room:%v 钱不够下注 请求下注不在下注区域内:%v ", acc.AccountId, self.roomId, betdata.GetBet())
		return
	}

	self.cd[acc.GetAccountId()] = now
	back := func(backunique string, backmoney int64,bwType int32) {
		if acc.GetMoney()-betdata.GetBet() != uint64(backmoney) {
			log.Warnf("数据错误  ->>>>>> userID:%v money:%v Bet:%v gold:%v", acc.GetUnDevice(), acc.GetMoney(), betdata.GetBet(), backmoney)
			acc.AddMoney(backmoney-int64(acc.GetMoney()), common.EOperateType_INIT)
		} else {
			acc.AddMoney(int64(-(betdata.GetBet())), common.EOperateType_LHD_BET)
		}
		playerBets, e := self.betPlayers[acc.AccountId]
		if !e {
			self.betPlayers[acc.AccountId] = make(map[int32]int64)
			playerBets, _ = self.betPlayers[acc.AccountId]
		}
		playerBets[int32(betdata.Area)] += int64(betdata.Bet)
		self.bets_cache = append(self.bets_cache, &protomsg.BET_LHD_RES_BetPlayer{
			AccountID: acc.GetAccountId(),
			Area:      betdata.GetArea(),
			Bet:       betdata.Bet,
		})
		self.log("acc:%v下注成功,下注区域:%v 金额:%v", acc.GetAccountId(), betdata.Area, betdata.Bet)
		acc.Betcount--
	}

	acc.Betcount++
	log.Infof("收到消息 acc：%v robot:%v OSType:%v 开始请求下注 金额:%v 区域:%v ", acc.GetAccountId(), acc.Robot, acc.GetOSType(), betdata.GetBet(), betdata.GetArea())
	if acc.Robot != 0 || acc.GetOSType() != 4 {
		back(acc.UnDevice, int64(acc.GetMoney()-betdata.GetBet()),0)
	} else {
		// 错误返回
		errback := func() {
			log.Panicf("http请求报错 玩家:%v roomID:%v  下注:%v 失败", acc.GetAccountId(), self.roomId, betdata.GetBet())
		}
		self.log("acc:%v unique:%v 请求下注,下注区域:%v 金额:%v", acc.GetAccountId(), acc.UnDevice, betdata.Area, betdata.Bet)
		platform.Asyn_addMoney(5, self.addr_url, acc.UnDevice, -int64(betdata.GetBet()), int32(self.roomId),"game_lhd", fmt.Sprintf("龙虎斗 请求下注:%v", betdata.GetBet()), back, errback)
	}
}
func (self *betting) LHDMSG_CS_CLEAN_BET_LHD_REQ(actor int32, msg []byte, session int64) {
	betdata := packet.PBUnmarshal(msg, &protomsg.CLEAN_BET_LHD_REQ{}).(*protomsg.CLEAN_BET_LHD_REQ)
	var acc *account.Account
	now := utils.MilliSecondTimeSince1970()
	if session == 0 {
		acc = account.AccountMgr.GetAccountByIDAssert(betdata.GetAccountID())
	} else {
		acc = account.AccountMgr.GetAccountBySessionIDAssert(session)
	}
	if acc.Betcount > 0 {
		log.Warnf("玩家:%v 下注计数为:%v 需要等到所有下注都成功，才能清除", acc.GetAccountId(), acc.Betcount)
		return
	}

	// 先统计一下玩家的总下注
	totalBets := self.betPlayers[acc.GetAccountId()]
	totalVal := uint64(0)
	for _, v := range totalBets {
		totalVal += uint64(v)
	}

	if totalVal == 0 {
		log.Warnf("玩家:%v 下注 为0 区域下注为:%v 不需要清除", acc.GetAccountId(), totalBets)
		return
	}
	// 不管是异步押注，还是直接押注，只要通过了押注校验，就现添加到betPlayers里，防止玩家退出游戏
	if self.betPlayers[acc.AccountId] == nil {
		self.betPlayers[acc.AccountId] = make(map[int32]int64)
	}

	acc.CLeanTime = now
	self.cd[acc.GetAccountId()] = now
	back := func(backunique string, backmoney int64) {
		if acc.GetMoney()+totalVal != uint64(backmoney) {
			self.log("数据错误  ->>>>>> roomid:%v userID:%v money:%v Bet:%v gold:%v", self.roomId, acc.GetUnDevice(), acc.GetMoney(), totalVal, backmoney)
			acc.AddMoney(backmoney-int64(acc.GetMoney()), common.EOperateType_INIT)
		} else {
			acc.AddMoney(int64(totalVal), common.EOperateType_LHD_BET_CLEAN)
		}

		acc.CLeanTime = 0
		bet := self.betPlayers[acc.AccountId]
		delete(self.betPlayers, acc.AccountId)
		// 通知玩家更新下注区域
		total := self.areaBetVal(true)
		msg := &protomsg.CLEAN_BET_LHD_RES{
			AccountID:        acc.AccountId,
			AreaBetVal:       total,
			PlayerAreaBetVal: bet,
		}
		self.SendBroadcast(protomsg.LHDMSG_SC_CLEAN_BET_LHD_RES.UInt16(), msg)

	}

	log.Infof("玩家:%v  请求清除下注", acc.GetAccountId())
	if acc.Robot == 0 || acc.OSType != 4 {
		back(acc.UnDevice, int64(acc.GetMoney()+totalVal))
	} else {
		// 错误返回
		errback := func() {
			log.Panicf("http请求报错 玩家:%v roomID:%v  下注:%v 失败", acc.GetAccountId(), self.roomId, totalVal)
		}
		platform.Asyn_addMoney(5, self.addr_url, acc.UnDevice, int64(totalVal), int32(self.roomId),"game_lhd", fmt.Sprintf("龙虎斗请求清除下注:%v ", totalVal), back, errback)
	}
}
