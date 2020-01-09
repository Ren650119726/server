package logic

import (
	"fmt"
	"root/common"
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg/inner"
	"root/server/hall/account"
	"root/server/hall/send_tools"
	"root/server/hall/server"
	"root/server/hall/speaker"
	"root/server/hall/types"
	"strconv"
	"unicode/utf8"
)

func CMD_Help(sParam []string) {
	fmt.Println("=========================================================================")
	fmt.Println("=========================================================================")
	fmt.Println("=========================================================================")
	fmt.Println(" 重要: 关服更新或维护流程说明!!!!")
	fmt.Println(" 步骤1. Hall    执行stopall命令; 参数: 停服维护时间(分钟)")
	fmt.Println(" 步骤2. Hall    等待所有游戏通知大厅可以关闭后, 会有提示")
	fmt.Println(" 步骤3. Game    异常若长时间都未通知大厅可以关闭, 在游戏服务器, 手动踢玩家和关闭游戏进程")
	fmt.Println(" 步骤4. Hall    执行kickall命令, 必须在所有游戏通知大厅可以关闭后, ")
	fmt.Println(" 步骤5. Hall    执行saveall命令, 等待提示ctrl+c")
	fmt.Println(" 步骤6. Supervisord执行stop all命令, 开始执行更新exe和config等流程")
	fmt.Println("=========================================================================")
	fmt.Println("=========================================================================")
	fmt.Println("=========================================================================")
}

func (self *Hall) CMD_LoadConfig(sParam []string) {
	strCMD := ""
	if len(sParam) > 0 {
		strCMD = sParam[0]
	}
	if strCMD == "" || strCMD == "all" {
		config.LoadPublic_Conf()
		config.LoadRobot_Conf()
		config.LoadStore_Conf()
		self.SetNode()
	} else if strCMD == "public" {
		config.LoadPublic_Conf()
	} else if strCMD == "robot" {
		config.LoadRobot_Conf()
	} else if strCMD == "store" {
		config.LoadStore_Conf()
	}

	strServerIP := utils.GetLocalIP()
	HallMgr.PrintSign(strServerIP)
	fmt.Printf("====== Hall 命令执行成功 ======\r\n")
}

func CMD_StopAll(sParam []string) {
	if len(sParam) <= 0 {
		fmt.Printf("× 参数错误, 请输入维护时间 (单位:分钟)\r\n")
		return
	}

	nTime, err := strconv.Atoi(sParam[0])
	if err != nil {
		fmt.Printf("× 参数错误, 请输入维护时间 (单位:分钟)\r\n")
		return
	}
	HallMgr.setMaintenanceTime(uint32(nTime))
	fmt.Printf("====== Hall 命令执行成功 ======\r\n")
}

func saveAll(isForceSave bool) {
	// 将玩家重要数据写入日志
	for _, account := range account.AccountMgr.AccountbyID {
		if account.Robot == 0 {
			log.Infof("UPDATE gd_account SET gd_UnDevice='%v', gd_Phone='%v', gd_WeiXin='%v', gd_RMB=%v, gd_SafeRMB=%v, gd_Salesman=%v WHERE gd_AccountID=%v;", account.UnDevice, account.Phone, account.WeiXin, account.Money, account.SafeMoney, account.Salesman, account.AccountId)
		}
	}

	// 将邮件信息写入日志
	account.EmailMgr.LogAllEmail()

	log.Info("数据备份到日志完成, 请等待数据库回存完毕提示")
	log.Info("注意! 在数据库回存完毕之前, 请不要执行任何操作!!")

	// 回存所有奖金池和水位线
	HallMgr.SaveAllWaterLine()
	HallMgr.SaveAllBonusPool()
	HallMgr.SaveWebData()
	account.AccountMgr.SendTransferRMBLog()

	// 停服更新回存方式, 只回存有变化的玩家数据; 邮件不回存,有改变时已回存
	if isForceSave == false {

		// 回存所有有改变的玩家数据
		nSavePlayerCount := 0
		for _, tAccount := range account.AccountMgr.AccountbyID {
			if tAccount.Robot == 0 && tAccount.Change == true {
				tAccount.Save()
				nSavePlayerCount++
			}
		}
		log.Infof("=========回存玩家人数:%v", nSavePlayerCount)

		// 通知DB所有数据回存完毕
		send_tools.Send2DB(inner.SERVERMSG_HD_SAVE_ALL.UInt16(), nil)

		// 关闭心跳
		pack := packet.NewPacket(nil)
		pack.SetMsgID(utils.ID_HEART_CLOSE)
		core.CoreSend(0, common.EActorType_CONNECT_DB.Int32(), pack.GetData(), 0)

	} else {
		// 强制回存所有玩家数据, 无论是否数据有改变
		mPlayer := make(map[uint32]bool)
		FORCE_SAVE_PLAYER_TIME := config.GetPublicConfig_Int64("FORCE_SAVE_PLAYER_TIME")
		nNowTime := utils.SecondTimeSince1970()
		for nAccountID, tAccount := range account.AccountMgr.AccountbyID {
			if tAccount.Robot == 0 {
				if nNowTime < tAccount.LoginTime+FORCE_SAVE_PLAYER_TIME {
					tAccount.Save()
					mPlayer[nAccountID] = true
				}
			}
		}

		// 强制回存指定玩家的邮件数据
		account.EmailMgr.SaveAll(mPlayer)

		// 通知DB所有数据回存完毕
		send_tools.Send2DB(inner.SERVERMSG_HD_SAVE_ALL.UInt16(), nil)
	}
}

func CMD_SaveAll(sParam []string) {
	if HallMgr.nMaintenanceTime > 0 {
		saveAll(false)
	} else {
		fmt.Printf(" × 错误: 请执行help命令查看说明\r\n")
	}
	fmt.Printf("====== Hall 命令执行成功 ======\r\n")
}
func CMD_KickAll(sParam []string) {
	HallMgr.ListenActor.Suspend()
	fmt.Printf("====== Hall  关闭监听链接 开始踢人... ======\r\n")
	CMD_On(nil)
}

func CMD_ForceSaveAll(sParam []string) {
	log.Info("!!!!!!!!!!!!!!!!!强制回存所有玩家数据  开始 !!!!!!!!!!!!!!!!!")
	saveAll(true)
	log.Info("!!!!!!!!!!!!!!!!!强制回存所有玩家数据  结束 !!!!!!!!!!!!!!!!!")
	fmt.Printf("====== Hall 命令执行成功 ======\r\n")
}

func CMD_Count(sParam []string) {
	nRegCount := len(account.AccountMgr.AccountbyID)
	nRobot := 0
	nOnline := 0
	nOffline := 0
	for _, tAccount := range account.AccountMgr.AccountbyID {
		if tAccount.Robot > 0 {
			nRobot++
		} else if tAccount.IsOnline() == true {
			nOnline++
		} else {
			nOffline++
		}
	}
	fmt.Printf("总注册:%v  当前在线玩家:%v  当前离线玩家:%v  机器人个数:%v  可分配帐号ID个数:%v\r\n", nRegCount, nOnline, nOffline, nRobot, len(account.AccountMgr.IDAssign))
	fmt.Printf("====== Hall 命令执行成功 ======\r\n")
}

func CMD_Player(sParam []string) {
	if len(sParam) <= 0 {
		fmt.Printf("× 参数错误, 请输入玩家帐号ID\r\n")
		return
	}

	nAccountID, err := strconv.Atoi(sParam[0])
	if err != nil {
		fmt.Printf("× 参数错误, 请输入正确的玩家帐号ID\r\n")
		return
	}

	tAccount := account.AccountMgr.GetAccountByID(uint32(nAccountID))
	if tAccount == nil {
		fmt.Printf("× 找不到指定ID的玩家, 请输入正确的玩家帐号ID\r\n")
		return
	} else {
		var strType string
		if tAccount.IsOnline() == true {
			strType = "在线"
		} else {
			strType = "离线"
		}
		if tAccount.Robot > 0 {
			fmt.Printf("%v 机器人 %v %v 所在:%v 房间ID:%v 元宝:%v 保险箱:%v 特殊:%v \r\n", strType, tAccount.AccountId, tAccount.Name,  tAccount.RoomID, tAccount.Money, tAccount.SafeMoney,tAccount.Special)
		} else {
			fmt.Printf("%v 玩家 %v %v 所在:%v 房间ID:%v  元宝:%v 保险箱:%v 代理:%v 特殊:%v 系统:%v\r\n", strType, tAccount.AccountId, tAccount.Name,  tAccount.RoomID, tAccount.Money, tAccount.SafeMoney, types.ESalesmanType(tAccount.Salesman), tAccount.Special, tAccount.OSType)
		}

		fmt.Printf("%v 头像URL:%v\r\n", strType, tAccount.HeadURL)
	}
	fmt.Printf("====== Hall 命令执行成功 ======\r\n")
}

func CMD_To(sParam []string) {
	if len(sParam) <= 0 {
		fmt.Printf("× 参数错误, 请输入游戏名字\r\n")
		return
	}

	strGameType := sParam[0]
	if strGameType == "all" {
		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(inner.SERVERMSG_SS_TEST_NETWORK.UInt16())
		for _, tNode := range server.ServerMgr.GetAllServerList() {
			send_tools.Send2Game(tSend.GetData(), tNode.SessionID)
		}
	} else {
		nGameType, isExist := common.GameTypeByString[strGameType]
		if isExist == false {
			fmt.Printf("× 参数错误, 请输入正确的游戏类型\r\n")
			return
		}
		nSessionID := server.ServerMgr.GetBySessionID(uint8(nGameType))
		if nSessionID > 0 {
			tSend := packet.NewPacket(nil)
			tSend.SetMsgID(inner.SERVERMSG_SS_TEST_NETWORK.UInt16())
			send_tools.Send2Game(tSend.GetData(), nSessionID)
			fmt.Printf("====== Hall 命令执行成功 ======\r\n")
		} else {
			fmt.Printf("× 参数错误, 请输入正确的游戏类型\r\n")
		}
	}
}

func CMD_On(sParam []string) {
	nCount := 0
	var nTotalRMB uint64
	var nTotalSafeRMB uint64
	for _, tAccount := range account.AccountMgr.AccountbyID {
		if tAccount.Robot == 0 {
			if tAccount.IsOnline() == true {
				fmt.Printf("在线玩家 %v %v 房间ID:%v  元宝:%v 保险箱:%v 代理:%v 特殊:%v 系统:%v\r\n", tAccount.AccountId, tAccount.Name,  tAccount.RoomID, tAccount.Money, tAccount.SafeMoney, types.ESalesmanType(tAccount.Salesman), tAccount.Special, tAccount.OSType)
				nCount++
			}
			nTotalRMB += tAccount.Money
			nTotalSafeRMB += tAccount.SafeMoney
		}
	}
	fmt.Printf("%v 总在线:%v 全服玩家身上元宝:%v, 保险箱元宝:%v, 总计:%v\r\n", utils.DateString(), nCount, nTotalRMB, nTotalSafeRMB, (nTotalRMB+nTotalSafeRMB))
	account.AccountMgr.UpdateOnlinePlayer(false, true)
	fmt.Printf("====== Hall 命令执行成功 ======\r\n")
}
func CMD_Off(sParam []string) {

	nCount := 0
	for _, tAccount := range account.AccountMgr.AccountbyID {
		if tAccount.IsOnline() == false && tAccount.Robot == 0 {
			fmt.Printf("离线玩家 %v %v 房间ID:%v  元宝:%v 保险箱:%v 代理:%v 特殊:%v 系统:%v\r\n", tAccount.AccountId, tAccount.Name,  tAccount.RoomID, tAccount.Money, tAccount.SafeMoney, types.ESalesmanType(tAccount.Salesman), tAccount.Special, tAccount.OSType)
			nCount++
		}
	}
	fmt.Printf("离线玩家人数:%v\r\n", nCount)
	fmt.Printf("====== Hall 命令执行成功 ======\r\n")
}

func CMD_Room(sParam []string) {
	if len(sParam) <= 0 {
		fmt.Printf("× 参数错误, 请输入房间ID 或 游戏类型字符串\r\n")
		return
	}

	nRoomID, err := strconv.Atoi(sParam[0])
	if err != nil {
		strGameType := sParam[0]
		nRoomCount := 0
		if strGameType == "all" {
			for _, tRoom := range HallMgr.mRoomTable {
				fmt.Printf("房间信息:%+v\r\n\r\n", tRoom)
				nRoomCount++
			}
		} else {
			nGameType, isExist := common.GameTypeByString[strGameType]
			if isExist == false {
				fmt.Printf("× 参数错误, 请输入正确的游戏类型\r\n")
				return
			}
			mGameRoom := HallMgr.getGameMap(uint8(nGameType))
			for _, tRoom := range mGameRoom {
				fmt.Printf("房间信息:%+v\r\n\r\n", tRoom)
				nRoomCount++
			}
		}
		fmt.Printf("%v共计:%v个房间\r\n", strGameType, nRoomCount)
		fmt.Printf("====== Hall 命令执行成功 ======\r\n")

	} else {
		tRoom := HallMgr.GetRoom(uint32(nRoomID))
		if tRoom == nil {
			fmt.Printf("× 找不到指定ID的房间, 请输入正确的房间ID\r\n")
			return
		}
		fmt.Printf("房间信息:%+v\r\n", tRoom)
		fmt.Printf("====== Hall 命令执行成功 ======\r\n")
	}
}

func CMD_Print_Speaker(sParam []string) {

	speaker.SpeakerMgr.PrintAll()
	fmt.Printf("====== Hall 命令执行成功 ======\r\n")
}
func CMD_Del_Speaker(sParam []string) {

	speaker.SpeakerMgr.PrintAll()
	speaker.SpeakerMgr.RemoveSpeaker(-1)
	speaker.SpeakerMgr.PrintAll()
	fmt.Printf("====== Hall 命令执行成功 ======\r\n")
}

func CMD_Add_Speaker(sParam []string) {

	if len(sParam) <= 0 {
		fmt.Printf("× 参数错误, 参数: 小喇叭内容, 字数100字以内\r\n")
		return
	}

	strContent := sParam[0]
	if utf8.RuneCountInString(strContent) > 100 {
		fmt.Printf("× 参数错误, 参数: 小喇叭内容, 字数100字以内\r\n")
		return
	}

	// 临时使用添加小喇叭功能, 为了简单. 只传小喇叭内容
	nStartTime := utils.SecondTimeSince1970()
	nDelTime := nStartTime + 86400
	nIntervalTime := uint16(20)

	speaker.SpeakerMgr.AddSpeaker(nStartTime, nDelTime, nIntervalTime, 2, strContent)
	fmt.Printf("====== Hall 命令执行成功 ======\r\n")
}

func CMD_Print_Email(sParam []string) {

	if len(sParam) < 1 {
		fmt.Printf("× 参数错误, 参数: 玩家ID\r\n")
		return
	}

	nAccountID, err := strconv.Atoi(sParam[0])
	if err != nil {
		fmt.Printf("× 参数错误, 请输入正确的玩家帐号ID\r\n")
		return
	}

	tAccount := account.AccountMgr.GetAccountByID(uint32(nAccountID))
	if tAccount == nil {
		fmt.Printf("× 找不到指定ID的玩家, 请输入正确的玩家帐号ID\r\n")
		return
	}

	account.EmailMgr.PrintEmail(uint32(nAccountID))
	fmt.Printf("====== Hall 命令执行成功 ======\r\n")
}

func CMD_Del_Email(sParam []string) {

	if len(sParam) < 2 {
		fmt.Printf("× 参数错误, 参数1:玩家ID; 参数2:邮件ID\r\n")
		return
	}

	nAccountID, err := strconv.Atoi(sParam[0])
	if err != nil {
		fmt.Printf("× 参数错误, 参数1:玩家ID; 参数2:邮件ID\r\n")
		return
	}

	nEmailID, err := strconv.Atoi(sParam[1])
	if err != nil {
		fmt.Printf("× 参数错误, 参数1:玩家ID; 参数2:邮件ID\r\n")
		return
	}

	tAccount := account.AccountMgr.GetAccountByID(uint32(nAccountID))
	if tAccount == nil {
		fmt.Printf("× 找不到指定ID的玩家, 请输入正确的玩家帐号ID\r\n")
		return
	}

	nRet := account.EmailMgr.ResetEmailRMB(uint32(nAccountID), uint32(nEmailID))
	if nRet == 0 {
		nRet = account.EmailMgr.RemoveMail(uint32(nAccountID), uint32(nEmailID))
		if nRet == 0 {
			fmt.Printf("====== 删除邮件成功 ======\r\n")
		} else {
			fmt.Printf("====== 删除邮件失败 ====== 错误码:%v\r\n", nRet)
		}
	} else {
		fmt.Printf("====== 删除邮件失败 ====== 错误码:%v\r\n", nRet)
	}
	fmt.Printf("====== Hall 命令执行成功 ======\r\n")
}

func CMD_SetUp_SalesmenType(sParam []string) {
	if len(sParam) < 1 {
		fmt.Printf("× 参数错误, 参数1: 玩家ID; 参数2: 代理身份类型\r\n")
		return
	}

	nAccountID, err := strconv.Atoi(sParam[0])
	if err != nil {
		fmt.Printf("× 参数错误, 参数1: 玩家ID; 参数2: 代理身份类型\r\n")
		return
	}

	nSetSalesman, err := strconv.Atoi(sParam[1])
	if err != nil {
		fmt.Printf("× 参数错误, 参数1: 玩家ID; 参数2: 代理身份类型\r\n")
		return
	}

	tAccount := account.AccountMgr.GetAccountByID(uint32(nAccountID))
	if tAccount == nil {
		fmt.Printf("× 找不到指定ID的玩家, 请输入正确的玩家ID\r\n")
		return
	}
	account.AccountMgr.SetUpSalesmenType(uint32(nAccountID), uint32(nSetSalesman), "", "")
	fmt.Printf("====== Hall 命令执行成功 ======\r\n")
}

func CMD_SetDown_SalesmenType(sParam []string) {
	if len(sParam) < 1 {
		fmt.Printf("× 参数错误, 参数1: 玩家ID; 参数2: 代理身份类型\r\n")
		return
	}

	nAccountID, err := strconv.Atoi(sParam[0])
	if err != nil {
		fmt.Printf("× 参数错误, 参数1: 玩家ID; 参数2: 代理身份类型\r\n")
		return
	}

	nSetSalesman, err := strconv.Atoi(sParam[1])
	if err != nil {
		fmt.Printf("× 参数错误, 参数1: 玩家ID; 参数2: 代理身份类型\r\n")
		return
	}

	tAccount := account.AccountMgr.GetAccountByID(uint32(nAccountID))
	if tAccount == nil {
		fmt.Printf("× 找不到指定ID的玩家, 请输入正确的玩家ID\r\n")
		return
	}
	account.AccountMgr.SetDownSalesmenType(uint32(nAccountID), uint32(nSetSalesman))
	fmt.Printf("====== Hall 命令执行成功 ======\r\n")
}

func CMD_Set_ChannelID(sParam []string) {
	if len(sParam) < 1 {
		fmt.Printf("× 参数错误, 参数1: 玩家ID; 参数2: 渠道ID\r\n")
		return
	}

	nAccountID, err := strconv.Atoi(sParam[0])
	if err != nil {
		fmt.Printf("× 参数错误, 参数1: 玩家ID; 参数2: 渠道ID\r\n")
		return
	}

	nNewChannelID, err := strconv.Atoi(sParam[1])
	if err != nil {
		fmt.Printf("× 参数错误, 参数1: 玩家ID; 参数2: 渠道ID\r\n")
		return
	}

	tAccount := account.AccountMgr.GetAccountByID(uint32(nAccountID))
	if tAccount == nil {
		fmt.Printf("× 找不到指定ID的玩家, 请输入正确的玩家ID\r\n")
		return
	}
	account.AccountMgr.SetChannelID(uint32(nAccountID), uint32(nNewChannelID), "CMD")
	fmt.Printf("====== Hall 命令执行成功 ======\r\n")
}

func CMD_Check(sParam []string) {

	fmt.Printf("====== Hall 命令执行成功 ======\r\n")
}

func CMD_Add_RMB(sParam []string) {

	if len(sParam) < 1 {
		fmt.Printf("× 参数错误, 参数1: 玩家ID; 参数2: 改变元宝数量\r\n")
		return
	}

	iChangeRMB, err := strconv.Atoi(sParam[1])
	if err != nil {
		fmt.Printf("× 参数错误, 参数1: 玩家ID; 参数2: 改变元宝数量\r\n")
		return
	}

	iAccountID, err := strconv.Atoi(sParam[0])
	if err != nil || iAccountID < 0 {
		fmt.Printf("× 参数错误, 参数1: 玩家ID; 参数2: 改变元宝数量\r\n")
		return
	}

	if HallMgr.isTestCharge == false {
		fmt.Printf("× 正式环境, 不允许加元宝\r\n")
		return
	}

	if iAccountID == 0 {
		for _, tAccount := range account.AccountMgr.AccountbyID {
			if tAccount.Robot == 0 {
				tAccount.AddMoney(int64(iChangeRMB), common.EOperateType_GM)
			}
		}
	} else {
		tAccount := account.AccountMgr.GetAccountByID(uint32(iAccountID))
		if tAccount == nil {
			fmt.Printf("× 找不到指定ID的玩家, 请输入正确的玩家ID\r\n")
			return
		}
		if tAccount.Robot == 0 {
			tAccount.AddMoney(int64(iChangeRMB), common.EOperateType_GM)
		}
	}
	fmt.Printf("====== Hall 命令执行成功 ======\r\n")
}

func CMD_Get_Robot_ID(sParam []string) {

	if len(sParam) < 1 {
		fmt.Printf("× 参数错误, 参数1: 需要机器人ID数量\r\n")
		return
	}

	nGetLen, err := strconv.Atoi(sParam[0])
	if err != nil {
		fmt.Printf("× 参数错误, 参数1: 需要机器人ID数量\r\n")
		return
	}

	var nNewID uint32
	for i := 0; i < nGetLen; i++ {
		account.AccountMgr.IDAssign, nNewID = utils.RandomSliceAndRemoveReturn(account.AccountMgr.IDAssign)
		log.Infof("%v", nNewID)
	}
	fmt.Printf("====== Hall 命令执行成功 ======\r\n")
}

func CMD_Test(sParam []string) {

	if len(sParam) < 1 {
		fmt.Printf("× 参数错误, 参数1: 1开启测试; 0关闭测试;\r\n")
		return
	}

	nOpen, err := strconv.Atoi(sParam[0])
	if err != nil {
		fmt.Printf("× 参数错误, 参数1: 1开启测试; 0关闭测试;\r\n")
		return
	}

	isTestServer, strLocalIP, _ := config.IsTestServer()
	if isTestServer == true {
		HallMgr.isTestCharge = (nOpen == 1)
	} else {
		HallMgr.isTestCharge = false
	}
	HallMgr.PrintSign(strLocalIP)
	fmt.Printf("====== Hall 命令执行成功 ======\r\n")
}

func CMD_ToDB(s []string) {
	send_tools.Send2DB(inner.SERVERMSG_SS_TEST_NETWORK.UInt16(), nil)
}

func (self *Hall) CMD_Node(s []string) {
	self.SetNode()
}

var exe = false

func (self *Hall) Open(s []string) {
	if len(s) != 1 {
		fmt.Printf("× 参数错误, 参数:  1 开放 0 关闭\r\n")
		return
	}

	num, _ := strconv.Atoi(s[0])
	HallMgr.OpenDesk = uint32(num)
}

func (self *Hall) WeiHuGame(s []string) {
	if len(s) != 2 {
		fmt.Printf("× 参数错误, 参数1:游戏类型, 参数2:开启1, 关闭0\r\n")
		return
	}

	strGameType := s[0]
	nGameType, isExist := common.GameTypeByString[strGameType]
	if isExist == false {
		fmt.Printf("× 参数错误, 请输入正确的游戏类型\r\n")
		return
	}
	nOpen, _ := strconv.Atoi(s[1])
	isOpen := nOpen == 1
	sList := server.ServerMgr.GetServerList(uint8(nGameType))
	if sList != nil {
		for _, tNode := range sList {
			tNode.IsMaintenance = isOpen
		}
		log.Infof(" ====> CMD 游戏:%v 设置维护标记:%v", nGameType, isOpen)
	}
}
