package common

type EOperateType byte

// 服务器类别定义
const (
	EOperateType_Unknown         EOperateType = 0 // 充值到帐通知可在游戏内领取充值邮件的操作原因
	EOperateType_CMD                          = 1 // 命令
	EOperateType_INIT                         = 2 // 初始化
	EOperateType_OFFLINE_CHARGE               = 3 // 人工充值
	EOperateType_SAFE_MONEY_SAVE              = 4 // 存钱到保险箱
	EOperateType_SAFE_MONEY_GET               = 5 // 从保险箱取钱
	EOperateType_FRUIT_MARY_BET               = 21 // 水果小玛利押注扣除金币
	EOperateType_FRUIT_MARY_WIN               = 22 // 水果小玛利 游戏1 获得钱
	EOperateType_FRUIT_MARY2_WIN              = 23 // 水果小玛利 游戏2 获得钱
	EOperateType_DFDC_BET               	  = 24 // 多福多财 押注扣除金币
	EOperateType_DFDC_WIN               	  = 25 // 多福多财 获得钱

)

var typeStringify_operate = map[EOperateType]string{
	EOperateType_Unknown:         "NIL",             // 充值到帐通知可在游戏内领取充值邮件的操作原因
	EOperateType_CMD:             "GM",              // GM功能
	EOperateType_INIT:            "INIT",            // 初始化
	EOperateType_OFFLINE_CHARGE:  "CHARGE",          // VIP充值
	EOperateType_SAFE_MONEY_SAVE: "SAFE_MONEY_SAVE", // 存钱到保险箱
	EOperateType_SAFE_MONEY_GET:  "SAFE_MONEY_GET",  // 从保险箱取钱
	EOperateType_FRUIT_MARY_WIN:  "FRUIT_MARY_WIN",  // 水果小玛利 游戏1 获得钱
	EOperateType_FRUIT_MARY2_WIN: "FRUIT_MARY2_WIN", // 水果小玛利 游戏1 获得钱
	EOperateType_FRUIT_MARY_BET:  "FRUIT_MARY_BET",  // 水果小玛利押注扣除金币
	EOperateType_DFDC_BET:  	  "FRUIT_DFDC_BET",  // 多福多财
	EOperateType_DFDC_WIN:  	  "FRUIT_DFDC_WIN",  // 多福多财
}

func (e EOperateType) String() string {
	return typeStringify_operate[e]
}

func (e EOperateType) Int32() int32 {
	return int32(e)
}
