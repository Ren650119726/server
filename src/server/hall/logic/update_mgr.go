package logic

import (
	"fmt"
	"math"
	"root/common"
	"root/common/config"
	"root/core/log"
	"root/core/utils"
	"root/server/hall/account"
	"root/server/hall/send_tools"
	"root/server/hall/server"
	"root/server/hall/speaker"
	"sort"
)

// 小喇叭更新
func OnSpeakerUpdate(dt int64) {
	speaker.SpeakerMgr.Update()
}

// 每分钟更新
func OneMinuteUpdate(dt int64) {

	account.AccountMgr.UpdateRMBChangelog()
	account.AccountMgr.ResetErrorRobot(false)
}

// 玩家实例数据定时回存
func OnSaveAccount(dt int64) {
	account.AccountMgr.SaveChangeSlice()
}

// 每30秒更新一次
func OnThirtySecondsUpdate(dt int64) {
	HallMgr.UpdateServiceFeeLog()
}

// 每5分钟更新
func FiveMinuteUpdate(dt int64) {

	// 定时更新数据有变化的玩家, 将定时回存
	account.AccountMgr.UpateSaveChangeSlice()
	account.AccountMgr.UpdateOnlinePlayer(false, false)
	account.AccountMgr.SendTransferRMBLog()
}

// 每小时更新
func OneHourUpdate(dt int64) {
	account.EmailMgr.Update()
	account.AccountMgr.ClearLoginLog()
}

// 每秒更新
func SecondUpdate(dt int64) {

}

// 每日0点更新更新
func ZeroUpdate(dt int64) {
	strNowTime := utils.DateString()
	send_tools.SQLLog(fmt.Sprintf("UPDATE log_login SET log_LogoutTime='%v' WHERE log_LogoutTime IS NULL", strNowTime))
	account.AccountMgr.UpdateOnlinePlayer(true, false)
	var nTotalRMB uint64
	var nTotalSafeRMB uint64
	for _, tAccount := range account.AccountMgr.AccountbyID {
		if tAccount.Robot == 0 {
			nTotalRMB += tAccount.Money
			nTotalSafeRMB += tAccount.SafeMoney
		}
	}
	send_tools.SQLLog(fmt.Sprintf("INSERT INTO log_money_daily(log_RMB, log_SafeRMB, log_Time) VALUES (%v, %v, '%v')", nTotalRMB, nTotalSafeRMB, strNowTime))
}