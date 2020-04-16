package common

type EOperateType byte

// 服务器类别定义
const (
	EOperateType_Unknown             EOperateType = 0  // 充值到帐通知可在游戏内领取充值邮件的操作原因
	EOperateType_CMD                              = 1  // 命令
	EOperateType_INIT                             = 2  // 初始化
	EOperateType_OFFLINE_CHARGE                   = 3  // 人工充值
	EOperateType_SAFE_MONEY_SAVE                  = 4  // 存钱到保险箱
	EOperateType_SAFE_MONEY_GET                   = 5  // 从保险箱取钱
	EOperateType_FRUIT_MARY_BET                   = 21 // 水果小玛利押注扣除金币
	EOperateType_FRUIT_MARY_WIN                   = 22 // 水果小玛利 游戏1 获得钱
	EOperateType_FRUIT_MARY2_WIN                  = 23 // 水果小玛利 游戏2 获得钱
	EOperateType_DFDC_BET                         = 24 // 多福多财 押注扣除金币
	EOperateType_DFDC_WIN                         = 25 // 多福多财 获得钱
	EOperateType_JPM_BET                          = 26 // 金瓶梅 押注扣除金币
	EOperateType_JPM_WIN                          = 27 // 金瓶梅 获得钱
	EOperateType_LUCKFRUIT_WIN                    = 28 // 幸运水果机 获得钱
	EOperateType_LUCKFRUIT_BET                    = 29 // 幸运水果机 押注扣除金币
	EOperateType_RED2BLACK_BET                    = 31 // 红黑大战 押注扣除金币
	EOperateType_RED2BLACK_BET_CLEAN              = 32 // 红黑大战 押注清除
	EOperateType_RED2BLACK_WIN                    = 33 // 红黑大战 赢的钱
	EOperateType_LHD_WIN                          = 34 // 龙虎斗 赢的钱
	EOperateType_LHD_BET                          = 35 // 龙虎斗 押注扣除金币
	EOperateType_LHD_BET_CLEAN                    = 36 // 龙虎斗 押注清除
	EOperateType_S777_WIN                         = 37 // 777 赢钱
	EOperateType_S777_BET                         = 38 // 777 押注
	EOperateType_HB_ASSIGN                        = 39 // 发红包
	EOperateType_HB_WIN                           = 40 // 红包赔的钱
	EOperateType_HB_BACK                          = 41 // 红包未抢完，退还的钱
	EOperateType_HB_BOMB_WIN                      = 42 // 发红包，赢得钱(有人中炸弹)

)

var typeStringify_operate = map[EOperateType]string{
	EOperateType_Unknown:             "NIL",                 // 充值到帐通知可在游戏内领取充值邮件的操作原因
	EOperateType_CMD:                 "GM",                  // GM功能
	EOperateType_INIT:                "INIT",                // 初始化
	EOperateType_OFFLINE_CHARGE:      "CHARGE",              // VIP充值
	EOperateType_SAFE_MONEY_SAVE:     "SAFE_MONEY_SAVE",     // 存钱到保险箱
	EOperateType_SAFE_MONEY_GET:      "SAFE_MONEY_GET",      // 从保险箱取钱
	EOperateType_FRUIT_MARY_WIN:      "FRUIT_MARY_WIN",      // 水果小玛利 游戏1 获得钱
	EOperateType_FRUIT_MARY2_WIN:     "FRUIT_MARY2_WIN",     // 水果小玛利 游戏1 获得钱
	EOperateType_FRUIT_MARY_BET:      "FRUIT_MARY_BET",      // 水果小玛利押注扣除金币
	EOperateType_DFDC_BET:            "DFDC_BET",            // 多福多财
	EOperateType_DFDC_WIN:            "DFDC_WIN",            // 多福多财
	EOperateType_JPM_BET:             "JPM_WIN",             // 金瓶梅
	EOperateType_JPM_WIN:             "JPM_WIN",             // 金瓶梅
	EOperateType_LUCKFRUIT_WIN:       "LUCKFRUIT_WIN",       // 幸运水果机
	EOperateType_LUCKFRUIT_BET:       "LUCKFRUIT_BET",       // 幸运水果机
	EOperateType_RED2BLACK_BET:       "RED2BLACK_BET",       // 红黑大战 押注扣除金币
	EOperateType_RED2BLACK_WIN:       "RED2BLACK_WIN",       // 红黑大战 赢的钱
	EOperateType_RED2BLACK_BET_CLEAN: "RED2BLACK_BET_CLEAN", // 红黑大战 押注清除
	EOperateType_LHD_WIN:             "LHD_WIN",             // 龙虎斗 赢的钱
	EOperateType_LHD_BET_CLEAN:       "LHD_BET_CLEAN",       // 龙虎斗 押注清除
	EOperateType_S777_WIN:            "777_WIN",             // 777 赢钱
	EOperateType_S777_BET:            "777_BET",             // 777 押注
	EOperateType_HB_ASSIGN:           "HONGBAO",             // 发红包
	EOperateType_HB_WIN:              "HONGBAO",             // 红包赔的钱
	EOperateType_HB_BACK:             "BACK",                // 红包退的钱
	EOperateType_HB_BOMB_WIN:         "BOMB_WIN",            // 发红包，赢得钱(有人中炸弹)

}

func (e EOperateType) String() string {
	return typeStringify_operate[e]
}

func (e EOperateType) Int32() int32 {
	return int32(e)
}
