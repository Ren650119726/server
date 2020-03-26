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
	"strconv"
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

	RobotMgr.Load()
	strServerIP := utils.GetLocalIP()
	GameMgr.PrintSign(strServerIP)
	for _, v := range GameMgr.nodes {
		send_tools.Send2Game(inner.SERVERMSG_SS_RELOAD_CONFIG.UInt16(), nil, v.session)
	}
	fmt.Printf("====== 命令执行成功 ======\r\n")
}


func CMD_On(sParam []string) {
	nCount := 0
	var nTotalRMB uint64
	var nTotalSafeRMB uint64
	now := utils.SecondTimeSince1970()
	for _, tAccount := range account.AccountMgr.AccountbyID {
		if tAccount.Robot == 0 {
			if tAccount.IsOnline() == true {
				online := now - tAccount.LoginTime
				totalminsec := (online / 60) * 60
				sec := online - totalminsec
				totalmin := totalminsec / 60
				hour := totalmin / 60
				min := totalmin % 60
				str := fmt.Sprintf("%-2v时 %-2v分 %-2v秒", hour, min, sec)
				fmt.Printf("在线玩家 %v %-15v 房间ID:%-5v 元宝:%-10v 保险箱:%-5v OSType:%v 在线时长:%v \r\n", tAccount.AccountId, tAccount.Name, tAccount.RoomID, tAccount.Money, tAccount.SafeMoney, tAccount.OSType, str)
				nCount++
			}
			nTotalRMB += tAccount.Money
			nTotalSafeRMB += tAccount.SafeMoney
		}
	}
	fmt.Printf("%v 总在线:%v 全服玩家身上元宝:%v, 保险箱元宝:%v, 总计:%v\r\n", utils.DateString(), nCount, nTotalRMB, nTotalSafeRMB, (nTotalRMB + nTotalSafeRMB))
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
	if acc.RoomID == 0 {
		if changeValue < 0 && -changeValue > int(acc.GetMoney()) {
			changeValue = int(-acc.GetMoney())
		}
		acc.AddMoney(int64(changeValue), common.EOperateType_CMD, 0)
		fmt.Printf("====== 命令执行成功 玩家:%v 金币:%v+(%v)=%v ======\r\n", acc.GetAccountId(), m, changeValue, acc.GetMoney())
	} else {
		GameMgr.Send2Game(inner.SERVERMSG_HG_NOTIFY_ALTER_DATE.UInt16(), &inner.NOTIFY_ALTER_DATE{
			AccountID:   acc.GetAccountId(),
			Type:        1,
			AlterValue:  int64(changeValue),
			RoomID:      acc.RoomID,
			OperateType: common.EOperateType_CMD,
		}, acc.RoomID)
		fmt.Printf("====== 命令执行成功 玩家:%v 金币:%v+(%v) 请在玩家房间内查看金币 ======\r\n", acc.GetAccountId(), m, changeValue)
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

	if acc.RoomID != 0 {
		GameMgr.Send2Game(inner.SERVERMSG_HG_NOTIFY_ALTER_DATE.UInt16(), &inner.NOTIFY_ALTER_DATE{
			AccountID:  acc.GetAccountId(),
			Type:       2,
			AlterValue: int64(changeValue),
			RoomID:     acc.RoomID,
		}, acc.RoomID)
	}

	k := acc.Kill
	acc.Kill = int32(changeValue)

	fmt.Printf("====== 命令执行成功 玩家:%v 当前杀数:%v 修改为:%v ======\r\n", acc.GetAccountId(), k, changeValue)
}

func CMD_ToDB(s []string) {
	send_tools.Send2DB(inner.SERVERMSG_SS_TEST_NETWORK.UInt16(), nil)
}

func (self *Hall) CMD_RoomInfo(s []string) {
	for roomid, room := range GameMgr.rooms {
		log.Infof(" 房间:%v 房间信息:%+v ", roomid,room)
	}
}

func (self *Hall) CMD_Stop(s []string) {
	for _, node := range GameMgr.nodes {
		send_tools.Send2Game(inner.SERVERMSG_SS_CLOSE_SERVER.UInt16(), nil, int64(node.session))
		log.Infof("通知游戏服关闭 游戏:%v", common.EGameType(node.gameType))
	}
}

func (self *Hall) CMD_Close(s []string) {
	self.ListenActor.Suspend()
	// 无论什么情况，关服时，玩家都要下线
	for _, acc := range account.AccountMgr.AccountbyID {
		if acc.LoginTime > acc.LogoutTime {
			acc.LogoutTime = utils.SecondTimeSince1970()
		}
	}
	//CMD_Save(nil)
}


func CMD_Save(s []string) {
	account.AccountMgr.ArchiveAll(true)
	GameMgr.Save()
	send_tools.Send2DB(inner.SERVERMSG_HD_SAVE_ALL.UInt16(), nil)
	log.Infof("====== 回存命令执行成功 ======")
}