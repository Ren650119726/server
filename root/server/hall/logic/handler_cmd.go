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
	fmt.Println("指令表 参数用一个英文空格隔开，不需要输入[]")
	fmt.Println("rmb [账号id] [修改金币]     说明：增加 减少金币(负数为减少),(8位数以内)")
	fmt.Println("reload                     说明：热更新所有配置表")
	fmt.Println("kill [账号id] [杀数]        说明：设置玩家杀数")
	fmt.Println("=========================================================================")
	fmt.Println("=========================================================================")
	fmt.Println("=========================================================================")
}

func (self *Hall) CMD_LoadConfig(sParam []string) {
	config.Load_Conf()

	strServerIP := utils.GetLocalIP()
	GameMgr.PrintSign(strServerIP)
	for _,v := range GameMgr.nodes{
		send_tools.Send2Game(inner.SERVERMSG_SS_RELOAD_CONFIG.UInt16(),nil,v.session)
	}
	fmt.Printf("====== 命令执行成功 ======\r\n")
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
	fmt.Printf("====== 命令执行成功 ======\r\n")
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
	fmt.Printf("====== 命令执行成功 ======\r\n")
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
	fmt.Printf("====== 命令执行成功 ======\r\n")
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
	fmt.Printf("====== 命令执行成功 ======\r\n")
}

func CMD_Print_Speaker(sParam []string) {

	speaker.SpeakerMgr.PrintAll()
	fmt.Printf("====== 命令执行成功 ======\r\n")
}
func CMD_Del_Speaker(sParam []string) {

	speaker.SpeakerMgr.PrintAll()
	speaker.SpeakerMgr.RemoveSpeaker(-1)
	speaker.SpeakerMgr.PrintAll()
	fmt.Printf("====== 命令执行成功 ======\r\n")
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
	fmt.Printf("====== 命令执行成功 ======\r\n")
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
	fmt.Printf("====== 命令执行成功 ======\r\n")
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
	fmt.Printf("====== 命令执行成功 ======\r\n")
}
func CMD_Add_Money(sParam []string) {
	if len(sParam) < 1 {
		fmt.Printf("× 参数错误, 参数1: 玩家ID; 参数2: 改变元宝数量\r\n")
		return
	}
	changeValue, err := strconv.Atoi(sParam[1])
	if err != nil {
		fmt.Printf("× 参数错误, 参数1: 玩家ID; 参数2: 改变元宝数量\r\n")
		return
	}

	accID, err := strconv.Atoi(sParam[0])
	if err != nil || accID < 0 {
		fmt.Printf("× 参数错误, 参数1: 玩家ID; 参数2: 改变元宝数量\r\n")
		return
	}

	acc := account.AccountMgr.GetAccountByID(uint32(accID))
	if acc == nil {
		fmt.Printf("× 找不到指定ID的玩家, 请输入正确的玩家ID\r\n")
		return
	}
	m := acc.GetMoney()
	if acc.RoomID == 0{
		if changeValue < 0 && -changeValue > int(acc.GetMoney()){
			changeValue = int(-acc.GetMoney())
		}
		acc.AddMoney(int64(changeValue), common.EOperateType_CMD)
		fmt.Printf("====== 命令执行成功 玩家:%v 金币:%v+(%v)=%v ======\r\n", acc.GetAccountId(),m, changeValue,acc.GetMoney())
	}else{
		GameMgr.Send2Game(inner.SERVERMSG_HG_NOTIFY_ALTER_DATE.UInt16(),&inner.NOTIFY_ALTER_DATE{
			AccountID:  acc.GetAccountId(),
			Type:       1,
			AlterValue: int64(changeValue),
			RoomID:     acc.RoomID,
			OperateType:common.EOperateType_CMD,
		}, acc.RoomID)
		fmt.Printf("====== 命令执行成功 玩家:%v 金币:%v+(%v) 请在玩家房间内查看金币 ======\r\n", acc.GetAccountId(),m, changeValue)
	}

}

func CMD_Kill(sParam []string) {
	if len(sParam) < 1 {
		fmt.Printf("× 参数错误, 参数1: 玩家ID; 参数2: 杀数\r\n")
		return
	}
	changeValue, err := strconv.Atoi(sParam[1])
	if err != nil {
		fmt.Printf("× 参数错误, 参数1: 玩家ID; 参数2: 杀数\r\n")
		return
	}

	accID, err := strconv.Atoi(sParam[0])
	if err != nil || accID < 0 {
		fmt.Printf("× 参数错误, 参数1: 玩家ID; 参数2: 杀数\r\n")
		return
	}

	acc := account.AccountMgr.GetAccountByID(uint32(accID))
	if acc == nil {
		fmt.Printf("× 找不到指定ID的玩家, 请输入正确的玩家ID\r\n")
		return
	}

	if acc.RoomID != 0{
		GameMgr.Send2Game(inner.SERVERMSG_HG_NOTIFY_ALTER_DATE.UInt16(),&inner.NOTIFY_ALTER_DATE{
			AccountID:  acc.GetAccountId(),
			Type:       2,
			AlterValue: int64(changeValue),
			RoomID:     acc.RoomID,
		}, acc.RoomID)
	}

	k := acc.Kill
	acc.Kill = int32(changeValue)


	fmt.Printf("====== 命令执行成功 玩家:%v 当前杀数:%v 修改为:%v ======\r\n", acc.GetAccountId(),k, changeValue)
}


func CMD_ToDB(s []string) {
	send_tools.Send2DB(inner.SERVERMSG_SS_TEST_NETWORK.UInt16(), nil)
}
func CMD_Save(s []string) {
	account.AccountMgr.ArchiveAll()
	GameMgr.Save()
	log.Infof("====== 回存命令执行成功 ======")
}
func (self *Hall)CMD_Stop(s []string) {
	CMD_Save(nil)
	self.ListenActor.Suspend()
}