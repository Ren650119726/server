package common

type EOperateType byte

// 服务器类别定义
const (
	EOperateType_Unknown         EOperateType = 0  // 充值到帐通知可在游戏内领取充值邮件的操作原因
	EOperateType_GM                           = 1  // GM功能
	EOperateType_INIT                         = 2  // 初始化
	EOperateType_OFFLINE_CHARGE               = 3  // 人工充值
	EOperateType_SAFE_MONEY_SAVE              = 4  // 存钱到保险箱
	EOperateType_SAFE_MONEY_GET               = 5  // 从保险箱取钱
	EOperateType_ROBOT_CREATE                 = 6  // 创造机器人
	EOperateType_ROBOT_DESTROY                = 7  // 创造机器人
	EOperateType_TRANSFER                     = 8  // 转让
	EOperateType_DAILY_SHARE                  = 9  // 每日分享
	EOperateType_ONLINE_CHARGE                = 10 // 在线充值
	EOperateType_BIND_PHONE                   = 19 // 绑定手机
	EOperateType_CREATE_ROOM                  = 20 // 创建房间
	EOperateType_ENTER_ROOM                   = 21 // 进入房间
	EOperateType_COST_RETURN                  = 22 // 返回房卡
	EOperateType_SERVICE_FEE                  = 23 // 服务费
	EOperateType_SETTLEMENT                   = 24 // 结算
	EOperateType_ERROR_RETURN                 = 25 // 异常返还
	EOperateType_BETTING                      = 26 // 下注
	EOperateType_DIVVIDEND                    = 27 // 带彩奖金扣除
	EOperateType_DISSOLVE_RETURN              = 28 // 解散返回
	EOperateType_BU_FEN                       = 29 // 补分
	EOperateType_UN_BETTING                   = 30 // 取消下注
	EOperateType_PENALTY                      = 31 // 惩罚
	EOperateType_BOMB                         = 32 // 炸弹
	EOperateType_EMAILL                       = 50 // EMail赠送
	EOperateType_GIFT_CODE                    = 51 // 礼品码领取
	EOperateType_EXCHANGE                     = 52 // 兑换
	EOperateType_REBATE                       = 53 // 返利
	EOperateType_HONGBAO                      = 54 // 发红包
	EOperateType_ROB_HONGBAO                  = 56 // 抢红包
	EOperateType_HONGBAO_PRIFIT               = 57 // 发红包获利红包
	EOperateType_DGK_HU                       = 58 // 赔胡
	EOperateType_DGK_GANG                     = 59 // 赔杠
	EOperateType_DGK_FEE                      = 60 // dgk 抽水
	EOperateType_DGK_REWARD                   = 61 // 奖金
	EOperateType_PANDA_HU                     = 78 // 赔胡
	EOperateType_PANDA_GANG                   = 79 // 赔杠
	EOperateType_PANDA_FEE                    = 80 // dgk 抽水
	EOperateType_PANDA_REWARD                 = 81 // 奖金
	EOperateType_PANDA_PIG                    = 82 // 花猪赔钱

)

var typeStringify_operate = map[EOperateType]string{
	EOperateType_Unknown:         "NIL",              // 充值到帐通知可在游戏内领取充值邮件的操作原因
	EOperateType_GM:              "GM",               // GM功能
	EOperateType_INIT:            "INIT",             // 初始化
	EOperateType_OFFLINE_CHARGE:  "CHARGE",           // VIP充值
	EOperateType_SAFE_MONEY_SAVE: "SAFE_MONEY_SAVE",  // 存钱到保险箱
	EOperateType_SAFE_MONEY_GET:  "SAFE_MONEY_GET",   // 从保险箱取钱
	EOperateType_ROBOT_CREATE:    "ROBOT_CREATE",     // 创造机器人
	EOperateType_ROBOT_DESTROY:   "ROBOT_DESTROY",    // 创造机器人
	EOperateType_TRANSFER:        "TRANSFER",         // 转让
	EOperateType_DAILY_SHARE:     "DAILY_SHARE",      // 每日分享
	EOperateType_ONLINE_CHARGE:   "ONLINE_CHARGE",    // 在线充值
	EOperateType_BIND_PHONE:      "BIND_PHONE",       // 绑定手机
	EOperateType_CREATE_ROOM:     "CREATE_ROOM0",     // 创建房间
	EOperateType_ENTER_ROOM:      "ENTER_ROOM1",      // 进入房间
	EOperateType_COST_RETURN:     "COST_RETURN2",     // 返回房卡
	EOperateType_SERVICE_FEE:     "SERVICE_FEE3",     // 服务费
	EOperateType_SETTLEMENT:      "SETTLEMENT4",      // 结算
	EOperateType_ERROR_RETURN:    "ERROR_RETURN5",    // 异常返还
	EOperateType_BETTING:         "BETTING6",         // 下注
	EOperateType_DIVVIDEND:       "DIVVIDEND7",       // 带彩奖金扣除
	EOperateType_DISSOLVE_RETURN: "DISSOLVE_RETURN8", // 解散返回
	EOperateType_BU_FEN:          "BU_FEN9",          // 补分
	EOperateType_UN_BETTING:      "UN_BETTING0",      // 取消下注
	EOperateType_PENALTY:         "PENALTY",          // 惩罚
	EOperateType_BOMB:            "BOMB",             // 炸弹
	EOperateType_EMAILL:          "EMAILL0",          // EMail赠送
	EOperateType_GIFT_CODE:       "GIFT_CODE1",       // 礼品码领取
	EOperateType_EXCHANGE:        "EXCHANGE",         // 兑换
	EOperateType_HONGBAO:         "HONGBAO",          // 红包
	EOperateType_ROB_HONGBAO:     "ROB_HONGBAO",      // 抢红包
	EOperateType_HONGBAO_PRIFIT:  "HONGBAO_PRIFIT",   // 发红包获利红包
	EOperateType_DGK_HU:          "DGK_HU",           // 赔胡
	EOperateType_DGK_GANG:        "DGK_GANG",         // 赔杠
	EOperateType_DGK_FEE:         "DGK_FEE",          // 赔杠
	EOperateType_DGK_REWARD:      "DGK_REWARD",       // 赔杠
	EOperateType_PANDA_HU:        "PANDA_HU",         // 赔胡
	EOperateType_PANDA_GANG:      "PANDA_GANG",       // 赔杠
	EOperateType_PANDA_FEE:       "PANDA_FEE",        // 赔杠
	EOperateType_PANDA_REWARD:    "PANDA_REWARD",     // 赔杠
	EOperateType_PANDA_PIG:       "PANDA_PIG",        // 花猪赔钱
}

func (e EOperateType) String() string {
	return typeStringify_operate[e]
}

func (e EOperateType) Int32() int32 {
	return int32(e)
}
