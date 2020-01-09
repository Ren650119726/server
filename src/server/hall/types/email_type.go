package types

type EmailType uint8

// 邮件类型
const (
	EMAIL_SYSTEM            EmailType = 1
	EMAIL_EXCHANGE_RETURN   EmailType = 2
	EMAIL_ACTIVITY          EmailType = 3
	EMAIL_SALESMAN          EmailType = 4
	EMAIL_PASSWORD          EmailType = 5
	EMAIL_REBATE            EmailType = 6
	EMAIL_SUBORDINATE       EmailType = 7
	EMAIL_OFFLINE_CHARGE    EmailType = 8
	EMAIL_ONLINE_CHARGE     EmailType = 9
	EMAIL_TRANSFER_PROPERTY EmailType = 10
	EMAIL_SALESMAN_DISCOUNT EmailType = 11
	EMAIL_MANUAL_1          EmailType = 12
	EMAIL_MANUAL_2          EmailType = 13
	EMAIL_MANUAL_3          EmailType = 14
)

var strEmailType = map[EmailType]string{
	EMAIL_SYSTEM:            "系统邮件",   // 邮件内容: web后台输入
	EMAIL_EXCHANGE_RETURN:   "兑换返还",   // 邮件内容: web后台输入
	EMAIL_ACTIVITY:          "活动邮件",   // 邮件内容: web后台输入
	EMAIL_SALESMAN:          "成为代理通知", //邮件内容: 后台帐号#$$#密码; "#$$#"分隔参数
	EMAIL_PASSWORD:          "通知网站密码", //邮件内容: 密码字符串
	EMAIL_REBATE:            "返利通知",   //邮件内容: 无
	EMAIL_SUBORDINATE:       "新增下属",   //邮件内容: 帐号ID#$$#名字; "#$$#"分隔参数
	EMAIL_OFFLINE_CHARGE:    "人工充值邮件", //邮件内容:
	EMAIL_ONLINE_CHARGE:     "在线充值邮件", //邮件内容:
	EMAIL_TRANSFER_PROPERTY: "转增元宝",   //邮件内容: 帐号ID#$$#名字#$$#金额; "#$$#"分隔参数
	EMAIL_SALESMAN_DISCOUNT: "代理补点",   //邮件内容: 帐号ID#$$#名字#$$#金额; "#$$#"分隔参数
	EMAIL_MANUAL_1:          "手动补偿充值", //邮件内容:
	EMAIL_MANUAL_2:          "手动补偿掉分", //邮件内容:
	EMAIL_MANUAL_3:          "手动补偿兑换", //邮件内容:
}

func (e EmailType) String() string {
	return strEmailType[e]
}

func (e EmailType) Value() uint32 {
	return uint32(e)
}
