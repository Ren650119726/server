package logic

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"root/common/config"
	"root/common/tools"
	"root/core"
	"root/core/log"
	"root/core/packet"
	"root/core/utils"
	"fmt"
	"root/protomsg"
	"root/server/hall/account"
	"root/server/hall/send_tools"
	"root/server/hall/types"
)

// 客户端同步服务器时间
func (self *Hall) MSG_CS_SYNC_SERVER_TIME(actor int32, msg []byte, session int64) {
	nServerTime := utils.MilliSecondTimeSince1970()
	send_tools.Send2Account(protomsg.MSG_SC_SYNC_SERVER_TIME.UInt16(),&protomsg.SYNC_SERVER_TIME{ServerTimeStamp:uint64(nServerTime)}, session)
}

func (self *Hall) MSG_LOGIN_HALL(actor int32, msg []byte, session int64) {
	loginMSG := packet.PBUnmarshal(msg,&protomsg.LOGIN_HALL_REQ{}).(*protomsg.LOGIN_HALL_REQ)

	LOGIN_SIGN_KEY := "abcd1234"
	strCheckSign := fmt.Sprintf("%v%v%v%v",  loginMSG.GetLoginType(), loginMSG.GetOSType(),loginMSG.GetUnique(), LOGIN_SIGN_KEY)
	strCheckSign = tools.MD5(strCheckSign)
	if strCheckSign != loginMSG.GetSign() {
		log.Warnf("Error, not match sign, loginType:%v, OsType:%v unique:%v Session:%v",loginMSG.GetLoginType(), loginMSG.GetOSType(),loginMSG.GetUnique(), session)
		//return
	}
	strClientIP := core.GetRemoteIP(session)
	switch loginMSG.LoginType {
	case uint32(types.LOGIN_TYPE_DEVICE.Value()):

	case uint32(types.LOGIN_TYPE_PHONE.Value()):
		match,err := regexp.MatchString(utils.PHONE_REG,loginMSG.GetUnique())
		if  err != nil {
			log.Warnf("正则匹配错误 unique:%v :%v ",loginMSG.GetUnique(), err.Error())
			return
		}
		if !match{
			log.Warnf("手机号登陆,但格式不为手机号unique:%v ",loginMSG.GetUnique())
			return
		}
	case uint32(types.LOGIN_TYPE_OTHER.Value()):	// 其他平台登陆
		go func() {
				resp, err := http.PostForm("http://47.244.119.129:1000/user/userInfo",
					url.Values{"channelId": {"GAME"}, "userId": {loginMSG.GetUnique()}})

				if err != nil {
					log.Warnf("三方平台，http 请求错误:%v", err.Error())
					return
				}

				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Warnf("三方平台，read 错误:%v", err.Error())
					return
				}
				log.Infof(string(body))
		}()
		return
	default:
		log.Panicf("不支持的登陆类型:%v", loginMSG.LoginType)
	}

	openWhiteList := config.GetPublicConfig_Int64("1")
	acc := account.AccountMgr.GetAccountByType(loginMSG.GetUnique(), uint8(loginMSG.GetLoginType()))
	if acc == nil {	// 注册新账号
		if openWhiteList == 1 {
			// 开启登录白名单功能后, 只允许特定帐号ID的玩家登录; 不允许注册
			log.Infof("登录白名单已开, 禁止登录; unique:%v, LoginType:%v, ClientIP:%v", loginMSG.GetUnique(), loginMSG.GetLoginType(), strClientIP)
			return
		}
		acc = account.AccountMgr.CreateAccount(loginMSG.GetUnique(), uint8(loginMSG.GetLoginType()), 0, "", "",uint8( loginMSG.GetOSType()), strClientIP, session, 0)
	} else { // 登陆账号
		if openWhiteList == 1 {
			WHITE_LOGIN_LIST := config.GetPublicConfig_String("2")
			mWhiteList := utils.SplitConf2Mapii(WHITE_LOGIN_LIST)
			if _, isExist := mWhiteList[int(acc.AccountId)]; isExist == false {
				log.Infof("登录白名单已开, 禁止登录; Account:%v, LoginType:%v, ClientIP:%v", loginMSG.GetUnique(), loginMSG.GetLoginType(), strClientIP)
				return
			}
		}
		// account was frozen
		frozenTime := acc.GetFrozenTime()
		if utils.MilliSecondTimeSince1970() < int64(frozenTime) {
			// 账号被冻结
			return
		}
		account.AccountMgr.LoginAccount(acc,uint8(loginMSG.LoginType), strClientIP, session)
	}
}
