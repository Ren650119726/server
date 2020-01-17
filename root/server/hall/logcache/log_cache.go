package logcache

import (
	"fmt"
	"root/core/log"
	"root/protomsg/inner"
	"root/server/hall/send_tools"
	"strings"
)

var LogCache  = newLogCacheLog()
type (
	LogCacheLog struct {
		saveServiceFeeLog [10][]string   // 服务费日志缓存表
		saveMoneyChangeLog [10][]string // 玩家金币变动日志缓存
	}
)

func newLogCacheLog() *LogCacheLog {
	ret := &LogCacheLog{}
	for i, _ := range ret.saveServiceFeeLog {
		ret.saveServiceFeeLog[i] = make([]string, 0, 100)
	}
	for i, _ := range ret.saveServiceFeeLog {
		ret.saveMoneyChangeLog[i] = make([]string, 0, 100)
	}
	return ret
}

// 每30秒执行一次组装SQL和回存
func (self *LogCacheLog) UpdateServiceFeeLog() {
	for i, _ := range self.saveServiceFeeLog {
		self.SendServiceFeeLog(i)
	}
}

// 每30秒执行一次组装SQL和回存
func (self *LogCacheLog) UpdateMoneyChangeLog() {
	for i, _ := range self.saveMoneyChangeLog {
		self.SendMoneyChangeLog(i)
	}
}

func (self *LogCacheLog) AddServiceFeeLog(serverfee_log *inner.SERVERFEE_LOG) {
	portion := int(serverfee_log.GetAccountID() % 10)
	strLog := fmt.Sprintf("(%v, %v, %v, %v, %v),",serverfee_log.GetAccountID(), serverfee_log.GetServerFee(), serverfee_log.GetGameType(),serverfee_log.GetTime(),serverfee_log.GetRoomID())
	self.saveServiceFeeLog[portion] = append(self.saveServiceFeeLog[portion], strLog) //
}

// 每30秒执行一次组装SQL和回存
func (self *LogCacheLog) SendServiceFeeLog(portion int) {
	if portion < 0 || portion > 9 {
		log.Errorf("不在0-9 之间portion:%v", portion)
		return
	}
	logs := self.saveServiceFeeLog[portion]
	nLen := len(logs)
	if nLen <= 0 {
		return
	}
	strSQL := fmt.Sprintf("INSERT INTO log_servicefee_%v (log_AccountID,log_ServiceFee,log_GameType,log_Time,log_RoomID) VALUES ", portion)
	for i := 0; i < nLen; i++ {
		strNode := logs[i]
		strSQL += strNode
	}
	strSQL = strings.TrimRight(strSQL, ",")
	send_tools.SQLLog(strSQL)
	log.Info(strSQL)
	self.saveServiceFeeLog[portion] = make([]string, 0, 100)
}


func (self *LogCacheLog) AddMoneyChangeLog(serverfee_log *inner.MONEYCHANGE) {
	portion := int(serverfee_log.GetAccountID() % 10)
	strLog := fmt.Sprintf("(%v, %v, %v, %v, '%v',%v),",serverfee_log.GetAccountID(), serverfee_log.GetChangeValue(), serverfee_log.GetValue(),serverfee_log.GetOperate(),serverfee_log.GetTime(),serverfee_log.GetRoomID())
	self.saveMoneyChangeLog[portion] = append(self.saveMoneyChangeLog[portion], strLog) //
}

// 每30秒执行一次组装SQL和回存
func (self *LogCacheLog) SendMoneyChangeLog(portion int) {
	if portion < 0 || portion > 9 {
		log.Errorf("不在0-9 之间portion:%v", portion)
		return
	}
	logs := self.saveMoneyChangeLog[portion]
	nLen := len(logs)
	if nLen <= 0 {
		return
	}
	strSQL := fmt.Sprintf("INSERT INTO log_money_%v (log_AccountID,log_ChangeValue,log_Value,log_Operate, log_Time, log_RoomID) VALUES ", portion)
	for i := 0; i < nLen; i++ {
		strNode := logs[i]
		strSQL += strNode
	}
	strSQL = strings.TrimRight(strSQL, ",")
	send_tools.SQLLog(strSQL)
	log.Info(strSQL)
	self.saveMoneyChangeLog[portion] = make([]string, 0, 100)
}
