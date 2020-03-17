package logic

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"root/common"
	"root/common/config"
	"root/common/tools"
	"root/core"
	"root/core/log"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/hall/account"
	"root/server/hall/send_tools"
	"root/server/hall/types"
	"strconv"
)

// 客户端同步服务器时间
func (self *Hall) MSG_CS_SYNC_SERVER_TIME(actor int32, msg []byte, session int64) {
	//timeMSG := packet.PBUnmarshal(msg,&protomsg.SYNC_SERVER_TIME{}).(*protomsg.SYNC_SERVER_TIME)
	//log.Infof("同步时间:%v 数据:%v msg:%v",session,timeMSG.GetServerTimeStamp(),msg)
	nServerTime := utils.MilliSecondTimeSince1970()
	send_tools.Send2Account(protomsg.MSG_SC_SYNC_SERVER_TIME.UInt16(), &protomsg.SYNC_SERVER_TIME{ServerTimeStamp: uint64(nServerTime)}, session)
}

func (self *Hall) MSG_LOGIN_HALL(actor int32, msg []byte, session int64) {
	loginMSG := packet.PBUnmarshal(msg, &protomsg.LOGIN_HALL_REQ{}).(*protomsg.LOGIN_HALL_REQ)

	LOGIN_SIGN_KEY := "abcd1234"
	strCheckSign := fmt.Sprintf("%v%v%v%v", loginMSG.GetLoginType(), loginMSG.GetOSType(), loginMSG.GetUnique(), LOGIN_SIGN_KEY)
	strCheckSign = tools.MD5(strCheckSign)
	if strCheckSign != loginMSG.GetSign() {
		//log.Warnf("Error, not match sign, loginType:%v, OsType:%v unique:%v Session:%v",loginMSG.GetLoginType(), loginMSG.GetOSType(),loginMSG.GetUnique(), session)
		//return
	}

	strClientIP := core.GetRemoteIP(session)

	var loginFun func(Unique, Name string, LoginType uint8, Gold int64)
	loginFun = func(Unique, Name string, LoginType uint8, Gold int64) {
		openWhiteList := config.GetPublicConfig_Int64(1)
		acc := account.AccountMgr.GetAccountByType(Unique, LoginType)
		if acc == nil { // 注册新账号
			if openWhiteList == 1 {
				// 开启登录白名单功能后, 只允许特定帐号ID的玩家登录; 不允许注册
				log.Infof("登录白名单已开, 禁止登录; unique:%v, LoginType:%v, ClientIP:%v", Unique, LoginType, strClientIP)
				return
			}
			acc = account.AccountMgr.CreateAccount(Unique, uint8(loginMSG.GetLoginType()), Name, "", uint8(loginMSG.GetOSType()), strClientIP, session, 0, uint64(Gold))
		} else { // 登陆账号
			if openWhiteList == 1 {
				WHITE_LOGIN_LIST := config.GetPublicConfig_String(2)
				mWhiteList := utils.SplitConf2Mapii(WHITE_LOGIN_LIST)
				if _, isExist := mWhiteList[int(acc.AccountId)]; isExist == false {
					log.Infof("登录白名单已开, 禁止登录; Account:%v, LoginType:%v, ClientIP:%v", Unique, LoginType, strClientIP)
					return
				}
			}
			// account was frozen
			frozenTime := acc.GetFrozenTime()
			if utils.MilliSecondTimeSince1970() < int64(frozenTime) {
				// 账号被冻结
				return
			}
			acc.OSType = loginMSG.OSType
			if LoginType == 4 {
				acc.Money = uint64(Gold)
				if acc.RoomID != 0 {
					sendPB := &inner.PLAYER_DATA_REQ{
						Account:     acc.AccountStorageData,
						AccountData: acc.AccountGameData,
						RoomID:      acc.RoomID,
						Reback:      true,
					}
					GameMgr.Send2Game(inner.SERVERMSG_HG_PLAYER_DATA_REQ.UInt16(), sendPB, acc.RoomID)
				}
			}
			account.AccountMgr.LoginAccount(acc, LoginType, strClientIP, session)
		}
		// 发送游戏房间信息
		GameMgr.SendGameInfo(session)
	}

	switch loginMSG.LoginType {
	case uint32(types.LOGIN_TYPE_DEVICE.Value()):

	case uint32(types.LOGIN_TYPE_PHONE.Value()):
		match, err := regexp.MatchString(utils.PHONE_REG, loginMSG.GetUnique())
		if err != nil {
			log.Warnf("正则匹配错误 unique:%v :%v ", loginMSG.GetUnique(), err.Error())
			return
		}
		if !match {
			log.Warnf("手机号登陆,但格式不为手机号unique:%v ", loginMSG.GetUnique())
			return
		}
	case uint32(types.LOGIN_TYPE_OTHER.Value()): // 其他平台登陆
		getuserinfo_URL := beego.AppConfig.DefaultString("DEF::getuserinfo", "")
		go func() {
			log.Infof("请求平台登录userId:%v type:%v url:%v", loginMSG.Unique, loginMSG.LoginType, getuserinfo_URL)
			resp, err := http.PostForm(getuserinfo_URL,
				url.Values{"channelId": {"DDHYLC"}, "userId": {loginMSG.Unique}})

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
			errorCode := map[int]string{
				1001: "channel 填写错误",
				1002: "userId 填写错误",
				1003: "渠道不存在",
				1004: "渠道权限错误",
				1005: "找不到用户",
			}
			log.Infof("平台返回:%v", string(body))
			var jsonstr map[string]interface{}
			e := json.Unmarshal(body, &jsonstr)
			if e != nil {
				log.Warnf("json 解析错误:%v ", e.Error())
				return
			}
			if err, e := jsonstr["status"]; e && int(err.(float64)) != 0 {
				log.Warnf("平台返回错误码:%v ", errorCode[int(err.(float64))])
				return
			} else {
				data := jsonstr["data"].(map[string]interface{})
				core.LocalCoreSend(0, common.EActorType_MAIN.Int32(), func() {
					userID := data["userId"].(float64)
					name := data["nickName"].(string)
					gold := data["gold"].(float64)
					loginFun(strconv.Itoa(int(userID)), name, 4, int64(gold))
				})
			}
		}()
		return
	default:
		log.Panicf("不支持的登陆类型:%v", loginMSG.LoginType)
	}

	loginFun(loginMSG.GetUnique(), "", uint8(loginMSG.GetLoginType()), 0)
}
