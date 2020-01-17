package types

type ELoginType uint8

// 登录类型
const (
	LOGIN_TYPE_DEVICE ELoginType = 1	// 游客登陆
	LOGIN_TYPE_PHONE  ELoginType = 2	// 手机登陆
	LOGIN_TYPE_WEIXIN ELoginType = 3	// 微信登陆

)

var strLoginType = map[ELoginType]string{
	LOGIN_TYPE_DEVICE: "设备码",
	LOGIN_TYPE_PHONE:  "手机",
	LOGIN_TYPE_WEIXIN: "微信",
}

func (e ELoginType) String() string {
	return strLoginType[e]
}

func (e ELoginType) Value() uint8 {
	return uint8(e)
}
