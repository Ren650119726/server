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

)

var typeStringify_operate = map[EOperateType]string{
	EOperateType_Unknown:         "NIL",             // 充值到帐通知可在游戏内领取充值邮件的操作原因
	EOperateType_CMD:             "GM",              // GM功能
	EOperateType_INIT:            "INIT",            // 初始化
	EOperateType_OFFLINE_CHARGE:  "CHARGE",          // VIP充值
	EOperateType_SAFE_MONEY_SAVE: "SAFE_MONEY_SAVE", // 存钱到保险箱
	EOperateType_SAFE_MONEY_GET:  "SAFE_MONEY_GET",  // 从保险箱取钱
}

func (e EOperateType) String() string {
	return typeStringify_operate[e]
}

func (e EOperateType) Int32() int32 {
	return int32(e)
}
