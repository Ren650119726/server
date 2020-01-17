package types

type EmailType uint8

// 邮件类型
const (
	EMAIL_SYSTEM            EmailType = 1
	EMAIL_EXCHANGE_RETURN   EmailType = 2
	EMAIL_ACTIVITY          EmailType = 3
	EMAIL_SALESMAN          EmailType = 4
	EMAIL_REBATE            EmailType = 6
	EMAIL_OFFLINE_CHARGE    EmailType = 8
	EMAIL_ONLINE_CHARGE     EmailType = 9
	EMAIL_TRANSFER_PROPERTY EmailType = 10
)

var strEmailType = map[EmailType]string{
	EMAIL_SYSTEM:            "系统邮件",   // 邮件内容: web后台输入
	EMAIL_ACTIVITY:          "活动邮件",   // 邮件内容: web后台输入
	EMAIL_SALESMAN:          "成为代理通知", //邮件内容: 后台帐号#$$#密码; "#$$#"分隔参数
	EMAIL_TRANSFER_PROPERTY: "转增元宝",   //邮件内容: 帐号ID#$$#名字#$$#金额; "#$$#"分隔参数
}

func (e EmailType) String() string {
	return strEmailType[e]
}

func (e EmailType) Value() uint32 {
	return uint32(e)
}
