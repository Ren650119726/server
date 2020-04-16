package room

import (
	"fmt"
	"root/common"
	"root/common/config"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/game_hongbao/account"
	"root/server/game_hongbao/send_tools"
	"root/server/platform"
)

// 玩家进入游戏
func (self *Room) HBMSG_CS_ENTER_GAME_HB_REQ(actor int32, msg []byte, session int64) {
	enterPB := packet.PBUnmarshal(msg, &protomsg.ENTER_GAME_HB_REQ{}).(*protomsg.ENTER_GAME_HB_REQ)
	self.enterRoom(enterPB.GetAccountID())
}

// 玩家离开
func (self *Room) HBMSG_CS_LEAVE_GAME_HB_REQ(actor int32, msg []byte, session int64) {
	enterPB := packet.PBUnmarshal(msg, &protomsg.LEAVE_GAME_HB_REQ{}).(*protomsg.LEAVE_GAME_HB_REQ)
	ret := uint32(1)
	if self.canleave(enterPB.GetAccountID()) {
		ret = 0
	}
	send_tools.Send2Account(protomsg.HBMSG_SC_LEAVE_GAME_HB_RES.UInt16(), &protomsg.LEAVE_GAME_HB_RES{
		Ret:    ret,
		RoomID: self.roomId,
	}, session)
}

// 玩家请求发红包
func (self *Room) HBMSG_CS_ASSIGN_HB_REQ(actor int32, msg []byte, session int64) {
	assignHB := packet.PBUnmarshal(msg, &protomsg.ASSIGN_HB_REQ{}).(*protomsg.ASSIGN_HB_REQ)
	var acc *account.Account
	if session == 0 {
		acc = account.AccountMgr.GetAccountByID(assignHB.GetAccountID())
		if acc == nil {
			return
		}
	} else {
		acc = account.AccountMgr.GetAccountBySessionID(session)
	}

	if _, e := self.Red_Odds[assignHB.Count]; !e {
		log.Warnf("%v %v 请求发红包,但是 包数:%v 不在配置中:%v ", acc.GetAccountId(), acc.UnDevice, assignHB.Count, self.Red_Odds)
		return
	}

	if assignHB.GetNum() > uint32(self.Red_Count) {
		log.Warnf("%v %v 请求发红包,但是 同时发包数%v 操作配置大小", acc.GetAccountId(), acc.UnDevice, assignHB.GetNum(), self.Red_Count)
		return
	}

	totalVal := assignHB.GetValue() * uint64(assignHB.GetNum())
	if acc.GetMoney() < totalVal {
		send_tools.Send2Account(protomsg.HBMSG_SC_PLAYERS_HB_LIST_RES.UInt16(), &protomsg.ASSIGN_HB_RES{Ret: 1}, session)
		return
	}
	send_tools.Send2Account(protomsg.HBMSG_SC_PLAYERS_HB_LIST_RES.UInt16(), &protomsg.ASSIGN_HB_RES{Ret: 0}, session)

	if acc.Robot == 0 {
		log.Infof("玩家:%v uid:%v 请求发红包 金额:%v 红包包数:%v 雷号:%v 同时发包数:%v ",
			acc.GetAccountId(), acc.GetUnDevice(), assignHB.GetValue(), assignHB.GetCount(), assignHB.BombNumber, assignHB.Num)
	}

	newHBLogic := func() {
		self.hongbaoID++
		if self.hongbaoID >= 99999999 {
			self.hongbaoID = 1
		}

		bombCount := 0
		if acc.Robot != 0 {
			i := utils.RandomWeight32(self.Send_Thunder, 1)
			bombCount = int(self.Send_Thunder[i][0])
		}
		new := &hongbao{
			hbID:         self.hongbaoID,
			assignerID:   acc.GetAccountId(),
			assignerName: acc.GetName(),
			value:        int64(assignHB.GetValue()),
			bombNumber:   int64(assignHB.GetBombNumber()),
			arr:          hongbao_slice(int64(assignHB.GetValue()), int64(assignHB.GetCount()), int64(self.Rand_Point), bombCount, int(assignHB.GetBombNumber())),
			count:        int64(assignHB.GetCount()),
			time:         utils.DateString(),
			grabs:        make(map[uint32]unit),
			bombs:        make(map[uint32]unit),
		}
		self.hbList = append(self.hbList, new) // 新红包加入红包列表
		if len(self.hbList) > int(self.Red_Max) {
			lasthb := self.hbList[0] // 处理多余的红包
			acc := account.AccountMgr.GetAccountByID(lasthb.assignerID)
			backVal := int64(0)
			c := 0
			for _, v := range lasthb.arr {
				backVal += v
				c++
			}
			if backVal > 0 {
				lasthb.arr = lasthb.arr[:0]
				acc.AddMoney(backVal, common.EOperateType_HB_BACK)
				if acc.GetOSType() == 4 {
					platform.Asyn_addMoney(5, self.addr_url, acc.UnDevice, backVal, int32(self.roomId), "game_hb",
						fmt.Sprintf("退还没抢完的红包 roomid:%v 红包id:%v 发包人:%v 红包金额:%v 包数:%v 剩余包数:%v 退还金额:%v ", self.roomId, lasthb.hbID, lasthb.assignerName, lasthb.value, lasthb.count, c, backVal),
						nil, nil)
				} else {
					log.Infof("roomID:%v 红包没抢完，退还钱 id:%v 发包人:%v 红包金额:%v 包数:%v 剩余包数:%v 退还金额:%v ", self.roomId, lasthb.hbID, lasthb.assignerName, lasthb.value, lasthb.count, c, backVal)
				}
			}

			self.hbList = self.hbList[1:]

			// 真实玩家，保存最后20个红包记录
			if acc.Robot == 0 {
				arr := self.players[acc.GetAccountId()]
				if arr == nil {
					self.players[acc.GetAccountId()] = make([]*hongbao, 0)
					arr = self.players[acc.GetAccountId()]
				}
				arr = append(arr, new)
				if len(arr) > int(self.Red_Max) {
					arr = arr[1:]
				}
				self.players[acc.GetAccountId()] = arr
			}
		}

		broadcast := &protomsg.BROADCAST_NEW_HB{New: &protomsg.HONGBAO{
			ID:            uint32(new.hbID),
			AssignerAccID: acc.GetAccountId(),
			AssignerName:  acc.GetName(),
			Value:         assignHB.GetValue(),
			Count:         uint64(assignHB.GetCount()),
			Spare:         uint64(assignHB.GetCount()),
			BombNumber:    uint64(assignHB.GetBombNumber()),
			Time:          new.time,
			Profits:       make(map[uint32]int64),
		}}

		for _, acc := range self.accounts {
			if acc.SessionId != 0 {
				send_tools.Send2Account(protomsg.HBMSG_SC_BROADCAST_NEW_HB.UInt16(), broadcast, acc.SessionId)
			}
		}
		if acc.Robot == 0 {
			log.Infof(colorized.Cyan("roomID:%v 玩家:%v uid:%v 发的红包 新红包:%+v"), self.roomId, acc.GetAccountId(), acc.GetUnDevice(), new)
		} else {
			log.Infof(colorized.Cyan("roomID:%v 机器人:%v uid:%v 发的红包 有:%v 颗雷 新红包:%+v"), self.roomId, acc.GetAccountId(), acc.GetUnDevice(), bombCount, new)
		}

	}

	if acc.Robot != 0 { // 机器人直接发
		for i := 0; i < int(assignHB.Num); i++ {
			newHBLogic()
		}
	} else {
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
			if acc.GetMoney()-totalVal != uint64(backmoney) {
				log.Warnf("数据错误  ->>>>>> userID:%v money:%v totalVal:%v gold:%v", acc.GetUnDevice(), acc.GetMoney(), totalVal, backmoney)
				acc.AddMoney(backmoney-int64(acc.GetMoney()), common.EOperateType_INIT)
			} else {
				acc.AddMoney(int64(-(totalVal)), common.EOperateType_HB_ASSIGN)
			}

			for i := 0; i < int(assignHB.Num); i++ {
				newHBLogic()
			}
		}
		if acc.GetOSType() == 4 {
			// 错误返回
			errback := func() {
				log.Warnf("http请求报错")
				resultMsg := &protomsg.ASSIGN_HB_RES{
					Ret: 3,
				}
				send_tools.Send2Account(protomsg.HBMSG_SC_ASSIGN_HB_RES.UInt16(), resultMsg, session)
			}
			platform.Asyn_addMoney(5, self.addr_url, acc.UnDevice, -int64(totalVal), int32(self.roomId), "game_hb", fmt.Sprintf("玩家:%v 身上钱:%v roomID:%v 请求发红包 hbID:%v 金额:%v 子包:%v 同时发包:%v",
				acc.GetAccountId(), acc.GetMoney(), self.roomId, self.hongbaoID, assignHB.GetValue(), assignHB.Count, assignHB.GetNum()), back, errback)
		} else {
			back("", int64(acc.GetMoney()-totalVal), 0)
		}
	}
}

// 请求抢红包
func (self *Room) HBMSG_CS_GRAB_HB_REQ(actor int32, msg []byte, session int64) {
	grab := packet.PBUnmarshal(msg, &protomsg.GRAB_HB_REQ{}).(*protomsg.GRAB_HB_REQ)
	hbID := grab.GetID()
	var acc *account.Account
	if session == 0 {
		acc = account.AccountMgr.GetAccountByID(grab.GetAccountID())
		if acc == nil {
			return
		}
	} else {
		acc = account.AccountMgr.GetAccountBySessionID(session)
	}

	var hb *hongbao
	for _, v := range self.hbList {
		if v.hbID == int32(hbID) {
			hb = v
			break
		}
	}
	if hb == nil {
		log.Warnf("不存在的红包实例ID:%v", hbID)
		return
	}

	// 不能重复抢
	if _, e := hb.grabs[acc.GetAccountId()]; e {
		send_tools.Send2Account(protomsg.HBMSG_SC_GRAB_HB_RES.UInt16(), &protomsg.GRAB_HB_RES{
			Ret: 3,
		}, session)
		return
	}
	// 红包已经被抢完了
	if len(hb.arr) == 0 {
		send_tools.Send2Account(protomsg.HBMSG_SC_GRAB_HB_RES.UInt16(), &protomsg.GRAB_HB_RES{
			Ret: 1,
		}, session)
		return
	}

	// 钱不够赔
	bombValue := uint64(self.Red_Odds[uint32(hb.count)] * hb.value / 100)
	if acc.GetMoney() < bombValue {
		log.Infof("玩家%v %v 身上的钱:%v 不够赔 不能抢 红包金额:%v 红包包数:%v 赔率:%v", acc.GetAccountId(), acc.GetUnDevice(), acc.GetMoney(), hb.value, hb.count, self.Red_Odds[uint32(hb.count)])
		send_tools.Send2Account(protomsg.HBMSG_SC_GRAB_HB_RES.UInt16(), &protomsg.GRAB_HB_RES{
			Ret: 2,
		}, session)
		return
	}
	send_tools.Send2Account(protomsg.HBMSG_SC_GRAB_HB_RES.UInt16(), &protomsg.GRAB_HB_RES{
		Ret: 0,
	}, session)

	val := hb.arr[0]
	hb.arr = hb.arr[1:]
	totalVal := val // 总盈利
	bomb := int64(0)
	if val%10 == hb.bombNumber {
		bomb = int64(bombValue)
	}
	totalVal -= bomb
	hb.grabs[acc.AccountId] = unit{
		name: acc.GetName(),
		val:  val,
	}

	// 先处理逻辑，再通知平台改变金币
	acc.AddMoney(totalVal, common.EOperateType_HB_ASSIGN)
	self.SendBroadcast(protomsg.HBMSG_SC_BROADCAST_UPDATE_GRAB.UInt16(), &protomsg.BROADCAST_UPDATE_GRAB{
		AccountID: acc.GetAccountId(),
		HbID:      uint32(hb.hbID),
		Profit:    totalVal,
	})
	if acc.GetOSType() == 4 {
		platform.Asyn_addMoney(5, self.addr_url, acc.UnDevice, totalVal, int32(self.roomId), "game_hb",
			fmt.Sprintf("玩家:%v roomID:%v 抢红包id:%v 红包金额:%v 包数:%v 雷号:%v 抢到:%v 总金额:%v", acc.AccountId, self.roomId, hb.hbID, hb.value, hb.count, hb.bombNumber, val, totalVal),
			nil, nil)
	}

	var profit int64
	if bomb != 0 {
		hb.bombs[acc.AccountId] = unit{
			name: acc.GetName(),
			val:  bomb,
		}
		profit_acc := account.AccountMgr.GetAccountByID(hb.assignerID)
		profit = bomb - (bomb * int64(self.Pump) / 10000)
		if profit_acc != nil {
			profit_acc.AddMoney(profit, common.EOperateType_HB_BOMB_WIN)
			self.SendBroadcast(protomsg.HBMSG_SC_BROADCAST_UPDATE_BOMB.UInt16(), &protomsg.BROADCAST_UPDATE_BOMB{
				AccountID: hb.assignerID,
				HbID:      uint32(hb.hbID),
				Profit:    profit,
			})
			if profit_acc.GetOSType() == 4 {
				platform.Asyn_addMoney(5, self.addr_url, profit_acc.UnDevice, profit, int32(self.roomId), "game_hb",
					fmt.Sprintf("玩家:%v %v roomID:%v 中雷 发红包人:%v 红包ID:%v 获得盈利:%v",
						profit_acc.AccountId, profit_acc.UnDevice, self.roomId, acc.UnDevice, hb.hbID, profit),
					nil, nil)
			}
		} else {
			log.Warnf("给玩家赔，但是找不到玩家了 %v", hb.assignerID)
		}
	}
	log.Infof(colorized.Blue("抢红包成功 玩家:%v uid:%v roomID:%v 红包ID:%v 抢得:%v 雷号:%v 赔付玩家:%v 炸了:%v 抽水后:%v"),
		acc.GetAccountId(), acc.GetUnDevice(), self.roomId, hb.hbID, val, hb.bombNumber, hb.assignerID, bomb, profit)

	broadcast := &protomsg.BROADCAST_UPDATE_HB{
		ID:    uint32(hb.hbID),
		Spare: uint32(len(hb.arr)),
	}
	self.SendBroadcast(protomsg.HBMSG_SC_BROADCAST_UPDATE_HB.UInt16(), broadcast)
	//
	//if acc.Robot != 0 {
	//	Logic()
	//} else {
	//	back := func(backunique string, backmoney int64, bwType int32) { // 押注
	//		if int64(acc.GetMoney())+totalVal != backmoney {
	//			log.Warnf("数据错误  ->>>>>> userID:%v money:%v totalVal:%v gold:%v", acc.GetUnDevice(), acc.GetMoney(), totalVal, backmoney)
	//			acc.AddMoney(backmoney-int64(acc.GetMoney()), common.EOperateType_INIT)
	//		} else {
	//			acc.AddMoney(totalVal, common.EOperateType_HB_ASSIGN)
	//			self.SendBroadcast(protomsg.HBMSG_SC_BROADCAST_UPDATE_GRAB.UInt16(), &protomsg.BROADCAST_UPDATE_GRAB{
	//				AccountID: acc.GetAccountId(),
	//				HbID:      uint32(hb.hbID),
	//				Profit:    totalVal,
	//			})
	//		}
	//		Logic()
	//	}
	//	if acc.GetOSType() == 4 {
	//		// 错误返回
	//		errback := func() {
	//			log.Warnf("acc:%v http请求报错", acc.GetUnDevice())
	//			resultMsg := &protomsg.ASSIGN_HB_RES{
	//				Ret: 3,
	//			}
	//			send_tools.Send2Account(protomsg.HBMSG_SC_ASSIGN_HB_RES.UInt16(), resultMsg, session)
	//			hb.arr = append(hb.arr, val)
	//			log.Warnf("平台返回错误，抢红包失败，重新把红包丢入红包列表里 红包ID:", hb.hbID, hb.arr)
	//		}
	//
	//	} else {
	//		back("", int64(acc.GetMoney())+totalVal, 0)
	//	}
	//}
}

func (self *Room) HBMSG_CS_HB_LIST_REQ(actor int32, msg []byte, session int64) {
	acc := account.AccountMgr.GetAccountBySessionIDAssert(session)
	arr := self.players[acc.GetAccountId()]
	send := &protomsg.HB_LIST_RES{List: make([]*protomsg.HONGBAO, 0)}
	if arr != nil {
		for _, hb := range arr {
			send.List = append(send.List, &protomsg.HONGBAO{
				ID:            uint32(hb.hbID),
				AssignerAccID: uint32(hb.assignerID),
				AssignerName:  hb.assignerName,
				Value:         uint64(hb.value),
				Count:         uint64(hb.count),
				Spare:         uint64(len(hb.arr)),
				BombNumber:    uint64(hb.bombNumber),
				Profits:       nil,
			})
		}
	}

	send_tools.Send2Account(protomsg.HBMSG_SC_HB_INFO_RES.UInt16(), send, session)
}
func (self *Room) HBMSG_CS_HB_INFO_REQ(actor int32, msg []byte, session int64) {
	grab := packet.PBUnmarshal(msg, &protomsg.HB_INFO_REQ{}).(*protomsg.HB_INFO_REQ)
	hbID := grab.GetHbID()
	acc := account.AccountMgr.GetAccountBySessionIDAssert(session)

	arr := self.players[acc.GetAccountId()]
	if arr == nil {
		log.Infof("房间:%v 没有玩家:%v 发红包记录", self.roomId, acc.GetAccountId())
		return
	}
	var hb *hongbao
	for _, h := range arr {
		if h.hbID == int32(hbID) {
			hb = h
			break
		}
	}
	if hb == nil {
		log.Infof("房间:%v 没有玩家:%v 发红包记录 请求的记录:%v  玩家红包记录:%v ", self.roomId, acc.GetAccountId(), hbID, arr)
		return
	}

	send := &protomsg.HB_INFO_RES{
		HbID: hbID,
		List: []*protomsg.HB_INFO_RES_HBGrabInfo{},
	}

	for accid, u := range hb.grabs {
		ub, e := hb.bombs[accid]
		ubval := int64(0)
		if e {
			ubval = ub.val
		}
		send.List = append(send.List, &protomsg.HB_INFO_RES_HBGrabInfo{
			AccountID: accid,
			Name:      u.name,
			Profit:    ub.val - ubval,
		})
	}

	send_tools.Send2Account(protomsg.HBMSG_SC_HB_INFO_RES.UInt16(), send, session)
}

func (self *Room) HBMSG_CS_PLAYERS_HB_LIST_REQ(actor int32, msg []byte, session int64) {
	account.AccountMgr.GetAccountBySessionIDAssert(session)

	ret := &protomsg.PLAYERS_HB_LIST_RES{}
	ret.Players = make([]*protomsg.AccountStorageData, 0)
	for _, p := range self.accounts {
		ret.Players = append(ret.Players, p.AccountStorageData)
	}
	send_tools.Send2Account(protomsg.HBMSG_SC_PLAYERS_HB_LIST_RES.UInt16(), ret, session)
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
