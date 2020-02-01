package logic

import (
	"fmt"
	"root/common"
	"root/common/config"
	"root/core/log"
	"root/core/utils"
	"root/protomsg/inner"
	"root/server/hall/account"
	"root/server/hall/send_tools"
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
	} else if strCMD == "public" {
		config.LoadPublic_Conf()
	}

	strServerIP := utils.GetLocalIP()
	GameMgr.PrintSign(strServerIP)
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
	ret := account.EmailMgr.RemoveMail(uint32(nAccountID), uint32(nEmailID))
	if ret == 0 {
		fmt.Printf("====== 删除邮件成功 ======\r\n")
	} else {
		fmt.Printf("====== 删除邮件失败 ====== 错误码:%v\r\n", ret)
	}
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

	tAccount := account.AccountMgr.GetAccountByID(uint32(iAccountID))
	if tAccount == nil {
		fmt.Printf("× 找不到指定ID的玩家, 请输入正确的玩家ID\r\n")
		return
	}
	if tAccount.Robot == 0 {
		tAccount.AddMoney(int64(iChangeRMB), common.EOperateType_CMD)
	}
	fmt.Printf("====== Hall 命令执行成功 玩家:%v 当前金币:%v ======\r\n", tAccount.GetAccountId(),tAccount.GetMoney())
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
	log.Infof("====== Hall 命令执行成功 ======")
}

func CMD_ToDB(s []string) {
	send_tools.Send2DB(inner.SERVERMSG_SS_TEST_NETWORK.UInt16(), nil)
}
func CMD_Save(s []string) {
	account.AccountMgr.ArchiveAll()
	log.Infof("====== Hall 命令执行成功 ======")
}
func (self *Hall)CMD_Stop(s []string) {
	self.ListenActor.Suspend()
}