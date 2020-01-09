package account

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"root/common"
	"root/common/config"
	"root/common/model/rediskey"
	"root/core"
	"root/core/db"
	"root/core/log"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/hall/send_tools"
	"root/server/hall/types"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
	"unsafe"
)

var AccountMgr = newAccountMgr()

type (
	login struct {
		LoginTime  string
		LogoutTime string
	}

	accountMgr struct {
		saveChange         []*Account
		accountbyUnDevice  map[string]*Account
		accountbyPhone     map[string]*Account
		accountbyWeiXin    map[string]*Account
		AccountbyID        map[uint32]*Account
		accountbySessionID map[int64]*Account
		playerTopCount     map[uint8]uint16
		sSaveRMBChangeLog  [10][]string // 元宝变更日志缓存表
		sTransferRMBLog    []string     // 赠送元宝日志缓存表
		mLoginLog          map[uint32]*login

		IDAssign []uint32
		Robots   []*Account
		Stamp    uint64
	}
)

func newAccountMgr() *accountMgr {

	mPlayerCount := make(map[uint8]uint16)
	mPlayerCount[0] = 0
	for nGameType := range common.GameTypeByID {
		mPlayerCount[uint8(nGameType)] = 0
	}

	ret := &accountMgr{
		saveChange:         make([]*Account, 0, 1000),
		accountbyUnDevice:  make(map[string]*Account),
		accountbyPhone:     make(map[string]*Account),
		accountbyWeiXin:    make(map[string]*Account),
		AccountbyID:        make(map[uint32]*Account),
		accountbySessionID: make(map[int64]*Account),
		IDAssign:           make([]uint32, 0, 100000),
		sTransferRMBLog: make([]string, 0, 100),
		mLoginLog:       make(map[uint32]*login),
		playerTopCount:  mPlayerCount,
	}
	for i, _ := range ret.sSaveRMBChangeLog {
		ret.sSaveRMBChangeLog[i] = make([]string, 0, 100)
	}
	return ret
}
// 每满50条或者60秒执行一次组装SQL和回存
func (self *accountMgr) AddRMBChangeLog(accid uint32, strLog string) {
	portion := accid % 10
	self.sSaveRMBChangeLog[portion] = append(self.sSaveRMBChangeLog[portion], strLog)

	RMB_CHANGE_LOG_COUNT := config.GetPublicConfig_Int64("RMB_CHANGE_LOG_COUNT")
	nCount := int64(len(self.sSaveRMBChangeLog[portion]))
	if nCount >= RMB_CHANGE_LOG_COUNT {
		self.SendRMBChangeLog(int(portion))
	}
}

func (self *accountMgr) UpdateRMBChangelog() {
	for i, _ := range self.sSaveRMBChangeLog {
		self.SendRMBChangeLog(i)
	}
}

// 每满50条或者60秒执行一次组装SQL和回存
func (self *accountMgr) SendRMBChangeLog(portion int) {
	if portion < 0 || portion > 9 {
		log.Errorf("不再0-9之间:%v", portion)
		return
	}

	logs := self.sSaveRMBChangeLog[portion]
	nLen := len(logs)
	if nLen <= 0 {
		return
	}

	strSQL := fmt.Sprintf("INSERT INTO log_rmb_%v (log_AccountID, log_ChangeValue, log_Value, log_Index, log_Operate, log_Time, log_RoomID, log_GameType, log_ClubID) VALUES ", portion)
	for i := 0; i < nLen; i++ {
		strNode := logs[i]
		strSQL += strNode
	}

	strSQL = strings.TrimRight(strSQL, ",")
	send_tools.SQLLog(strSQL)
	//log.Info(strSQL)

	self.sSaveRMBChangeLog[portion] = make([]string, 0, 100)
}

// 每5分钟回存一次
func (self *accountMgr) AddTransferRMBLog(strLog string) {
	self.sTransferRMBLog = append(self.sTransferRMBLog, strLog)
}

// 每5分钟回存一次
func (self *accountMgr) SendTransferRMBLog() {

	nLen := len(self.sTransferRMBLog)
	if nLen <= 0 {
		return
	}

	strSQL := "INSERT INTO log_transfer_rmb (log_AccountID,log_TargetID,log_Money,log_Time) VALUES "
	for i := 0; i < nLen; i++ {
		strNode := self.sTransferRMBLog[i]
		strSQL += strNode
	}

	strSQL = strings.TrimRight(strSQL, ",")
	send_tools.SQLLog( strSQL)
	self.sTransferRMBLog = make([]string, 0, 100)
}

// 所有玩家和机器人都初始化完成以后, 再将玩家和机器人的ID排除掉
func (self *accountMgr) CollatingIDAssign() {
	mCheckID := make(map[uint32]bool)
	for nID := range self.AccountbyID {
		mCheckID[nID] = true
	}

	for _, tNode := range config.GetRobotNameConfig() {
		if _, isExist := mCheckID[uint32(tNode.RobotID)]; isExist == false {
			mCheckID[uint32(tNode.RobotID)] = true
		} else {
			log.Errorf("机器人ID:%v 出现重复", tNode.RobotID)
		}
	}

	// 初始化帐号ID; 从100000--199999; 并将已使用的ID排除掉
	var nID uint32 = 100000
	for i := 0; i < 100000; i++ {
		if _, isExist := mCheckID[nID]; isExist == false {
			self.IDAssign = append(self.IDAssign, nID)
		}
		nID++
	}

	// 此时还没创建机器人, 帐号表里都是玩家
	nPlayer := len(self.AccountbyID)
	nRobot := len(config.GetRobotNameConfig())
	log.Infof("=========== 注册玩家帐号:%v, 机器人总数:%v, 可分配新帐号ID个数:%v", nPlayer, nRobot, len(self.IDAssign))
}

func (self *accountMgr) GetAccountByType(identifier string, loginType uint8) *Account {

	switch loginType {
	case types.LOGIN_TYPE_DEVICE.Value():
		return self.accountbyUnDevice[identifier]
	case types.LOGIN_TYPE_PHONE.Value():
		return self.accountbyPhone[identifier]
	case types.LOGIN_TYPE_WEIXIN.Value():
		return self.accountbyWeiXin[identifier]
	default:
		log.Panicf("不支持的登陆类型:%v", loginType)
		return nil
	}
}

func (self *accountMgr) GetAccountByID(id uint32) *Account {
	if id == 0 {
		return nil
	}
	return self.AccountbyID[id]
}

func (self *accountMgr) GetAccountBySessionID(session int64) *Account {
	if session == 0 {
		log.Panicf("找不到玩家:%v ",session)
		return nil
	}
	return self.accountbySessionID[session]
}

func (self *accountMgr) RemoveAccountBySessionID(session int64) {
	delete(self.accountbySessionID, session)
}

// 加载所有账号
func (self *accountMgr) LoadAllAccount(all_data []*protomsg.AccountStorageData) {
	for _, v := range all_data {
		new := NewAccount(v)
		_, exist := self.AccountbyID[new.AccountId]
		if exist {
			log.Fatalf("重复的account数据 AccountId:%v", new.AccountId)
			continue
		}

		if new.UnDevice != "" {
			_, exist = self.accountbyUnDevice[new.UnDevice]
			if exist {
				log.Fatalf("重复的account数据 UnDevice:%v", new.UnDevice)
				continue
			}
			self.accountbyUnDevice[new.UnDevice] = new
		}

		if new.Phone != "" {
			_, exist = self.accountbyPhone[new.Phone]
			if exist {
				log.Fatalf("重复的account数据 Phone:%v", new.Phone)
				continue
			}
			self.accountbyPhone[new.Phone] = new
		}

		if new.WeiXin != "" {
			_, exist = self.accountbyWeiXin[new.WeiXin]
			if exist {
				log.Fatalf("重复的account数据 WeiXin:%v", new.WeiXin)
				continue
			}
			self.accountbyWeiXin[new.WeiXin] = new
		}
		self.AccountbyID[new.AccountId] = new
	}
}

func (self *accountMgr) PrintMemoryUsageInfo() {
	var nTotalSize uintptr

	nTotalSize = 0
	for _, tNode := range self.ExchangeOrder {
		nTotalSize += unsafe.Sizeof(*tNode)
	}
	log.Infof("=========== 兑换订单共计:%v条, 占用内存:%v字节", len(self.ExchangeOrder), nTotalSize)
}

func (self *accountMgr) SaveChangeSlice() {

	nOldLen := len(self.saveChange)
	if nOldLen <= 0 {
		return
	}

	SAVE_CHANGE_COUNT := config.GetPublicConfig_Int64("SAVE_CHANGE_COUNT")
	nCount := int64(0)
	nLastIndex := 0
	for i, tAccount := range self.saveChange {
		nLastIndex = i
		if tAccount != nil && tAccount.IsChangeData == true && tAccount.Robot == 0 {
			nCount++
			tAccount.Save(false)
			if nCount >= SAVE_CHANGE_COUNT {
				break
			}
		}
	}

	self.saveChange = self.saveChange[nLastIndex+1:]
	log.Infof("SaveChangeSlice, 回存前数量:%v 剩余数量:%v", nOldLen, len(self.saveChange))
}

func (self *accountMgr) UpateSaveChangeSlice() {

	nCount := len(self.saveChange)
	if nCount <= 0 {
		for _, tAccount := range self.AccountbyID {
			if tAccount.IsChangeData == true && tAccount.Robot == 0 {
				self.saveChange = append(self.saveChange, tAccount)
			}
		}
	} else {
		for nAccountID, tAccount := range self.AccountbyID {
			if tAccount.IsChangeData == true && tAccount.Robot == 0 {
				isInSaveSlice := false
				for _, tSave := range self.saveChange {
					if tSave.AccountId == nAccountID {
						isInSaveSlice = true
						break
					}
				}
				if isInSaveSlice == false {
					self.saveChange = append(self.saveChange, tAccount)
				}
			}
		}
	}

	nCount = len(self.saveChange)
	if nCount > 0 {
		log.Infof("UpateSaveChangeSlice, 改变数量:%v", nCount)
	}
}

// 广播消息, 给所有在线玩家
// t类型1, 所有在大厅玩家接收
// t类型2, 所有在大厅和在房间的玩家接收
func (self *accountMgr) SendBroadcast(pack packet.IPacket, t uint8) {
	if t == 1 {
		for _, acc := range self.AccountbyID {
			if acc.IsOnline() && acc.Robot == 0 && acc.RoomID == 0 {
				send_tools.Send2Account(pack.GetData(), acc.SessionId)
			}
		}
	} else {
		for _, acc := range self.AccountbyID {
			if acc.IsOnline() && acc.Robot == 0 {
				send_tools.Send2Account(pack.GetData(), acc.SessionId)
			}
		}
	}
}

// 创建账号
func (self *accountMgr) CreateAccount(uniqueID string, nLoginType uint8, nChannelID uint16, strName string,strHeadURL string, nOSType uint8, strClientIP string, session int64, nRobot uint32) *Account {
	strNowTime := utils.DateString()
	var nNewAccountID uint32
	nLen := len(self.IDAssign)
	if nLen <= 0 {
		log.Fatalf("帐号ID已分配完 len:%v", nLen)
		return nil
	}
	// 随机分配一个id
	self.IDAssign, nNewAccountID = utils.RandomSliceAndRemoveReturn(self.IDAssign)

	tNewAccount := NewAccount(&protomsg.AccountStorageData{})
	switch nLoginType {
	case types.LOGIN_TYPE_DEVICE.Value():
		tNewAccount.UnDevice = uniqueID
		tNewAccount.Phone = ""
		tNewAccount.WeiXin = ""
		self.accountbyUnDevice[uniqueID] = tNewAccount
	case types.LOGIN_TYPE_PHONE.Value():
		tNewAccount.UnDevice = ""
		tNewAccount.Phone = uniqueID
		tNewAccount.WeiXin = ""
		self.accountbyPhone[uniqueID] = tNewAccount
	case types.LOGIN_TYPE_WEIXIN.Value():
		tNewAccount.UnDevice = ""
		tNewAccount.Phone = ""
		tNewAccount.WeiXin = uniqueID
		self.accountbyWeiXin[uniqueID] = tNewAccount
	}

	tNewAccount.AccountId = nNewAccountID
	tNewAccount.Name = strName
	tNewAccount.HeadURL = strHeadURL
	tNewAccount.Money = uint64(config.GetPublicConfig_Int64("INIT_RMB"))
	tNewAccount.SessionId = session
	tNewAccount.Robot = nRobot
	tNewAccount.ActiveIP = strClientIP
	tNewAccount.ActiveTime = strNowTime
	tNewAccount.OSType = uint32(nOSType)

	self.AccountbyID[nNewAccountID] = tNewAccount

	if nRobot == 0 {
		self.accountbySessionID[session] = tNewAccount
		// 新建账号，存数据库
		tNewAccount.Save()
		// 记录登录日志
		self.SendLoginLog(tNewAccount.AccountId)
	}

	// 登陆成功
	loginRet := &protomsg.LOGIN_HALL_RES{
		Ret:0,
		Account:tNewAccount.AccountStorageData,
		AccountData:tNewAccount.AccountGameData,
	}
	send_tools.Send2Account(protomsg.MSG_SC_LOGIN_HALL_RES.UInt16(),loginRet, session)

	return tNewAccount
}

func (self *accountMgr) LoginAccount(acc *Account, nLoginType uint8, clientIP string, session int64) {
	if acc.SessionId > 0 && acc.SessionId != session {
		send := packet.NewPacket(nil)
		send.SetMsgID(protomsg.MSG_SC_KICK_OUT_HALL.UInt16())
		send.WriteUInt8(1) // 顶号被踢出游戏
		send_tools.Send2Account(protomsg.MSG_SC_KICK_OUT_HALL.UInt16(),&protomsg.KICK_OUT_HALL{Ret:2}, acc.SessionId)

		// 取消之前的SessionID的关联
		delete(self.accountbySessionID, acc.SessionId)
		if acc.Robot == 0 {
			self.SendExitLog(acc.AccountId)
		}
	}
	acc.SessionId = session
	acc.LoginTime = utils.SecondTimeSince1970()

	// 建立新的关联
	self.accountbySessionID[session] = acc
	if acc.Robot == 0 {
		self.SendLoginLog(acc.AccountId)
		// 验证redis数据
		db.HGetKeyAll(rediskey.PlayerId(acc.AccountId), func(m map[string]string) {
			if m != nil {
				if s, e := m["Money"]; e {
					rmb, err := strconv.Atoi(s)
					if err != nil {
						log.Warnf("错误数据:%v err:%v", rmb, err.Error())
					} else {
						acc.Money = uint64(rmb)
					}
				}

				if s, e := m["SafeRMB"]; e {
					rmb, err := strconv.Atoi(s)
					if err != nil {
						log.Warnf("错误数据:%v err:%v", rmb, err.Error())
					} else {
						acc.SafeMoney = uint64(rmb)
					}
				}
			}

			// 登陆成功
			loginRet := &protomsg.LOGIN_HALL_RES{
				Ret:0,
				Account:acc.AccountStorageData,
				AccountData:acc.AccountGameData,
			}
			send_tools.Send2Account(protomsg.MSG_SC_LOGIN_HALL_RES.UInt16(),loginRet, session)
			log.Infof("login Player:%v, unique:%v, Money:%v, 登录类型:%v,  IP:%v, Session:%v", acc.AccountId, acc.Money, types.ELoginType(nLoginType), clientIP, session)
		})
	}
}

func (self *accountMgr) BindPhone(nAccountID uint32, strPhone string, isFromClient bool) uint8 {
	tAccount := self.GetAccountByID(nAccountID)

	if isFromClient == true {
		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(protomsg.Old_MSGID_BIND_PHONE.UInt16())
		if tAccount == nil {
			tSend.WriteUInt8(1)
			send_tools.Send2Account(tSend.GetData(), tAccount.SessionId)
			return 1
		}
		_, isExist := self.accountbyPhone[strPhone]
		if isExist == true || strPhone == "" || tAccount.Phone != "" || tAccount.Phone == strPhone {
			tSend.WriteUInt8(3)
			send_tools.Send2Account(tSend.GetData(), tAccount.SessionId)
			return 2
		}
	} else {
		if tAccount == nil {
			log.Warnf("Can't Find Account:%v, BindPhone:%v", nAccountID, strPhone)
			return 3
		}
		_, isExist := self.accountbyPhone[strPhone]
		if isExist == true || tAccount.Phone == strPhone {
			log.Warnf("Replace BindPhone, AccoutID:%v Phone:%v NewPhone:%v", nAccountID, tAccount.Phone, strPhone)
			return 4
		}
		// 允许后台重新绑定, 避免绑错了, 无法修改; 同时允许后台解绑;
		if tAccount.Phone != "" {
			if _, isExist := self.accountbyPhone[tAccount.Phone]; isExist == true {
				delete(self.accountbyPhone, tAccount.Phone)
			}
		}
	}

	if strPhone != "" {
		self.accountbyPhone[strPhone] = tAccount
	}
	tAccount.Phone = strPhone

	if isFromClient == true {
		BIND_PHONE_GIFT_RMB := config.GetPublicConfig_Int64("BIND_PHONE_GIFT_RMB")
		tAccount.AddMoney(BIND_PHONE_GIFT_RMB, common.EOperateType_BIND_PHONE)
	}
	tAccount.Save()

	strFrom := "Web"
	if isFromClient {
		strFrom = "Client"
	}
	log.Infof("绑定手机, AccountID:%v, Phone:%v, From:%v", nAccountID, strPhone, strFrom)

	if tAccount.IsOnline() == true {
		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(protomsg.Old_MSGID_BIND_PHONE.UInt16())
		tSend.WriteUInt8(0)
		tSend.WriteString(strPhone)
		send_tools.Send2Account(tSend.GetData(), tAccount.SessionId)
	}
	return 0
}

func (self *accountMgr) SetChannelID(nAccountID uint32, nNewChannelID uint32, strFrom string) uint8 {
	tAccount := self.GetAccountByID(nAccountID)
	if tAccount != nil && tAccount.ChannelID != nNewChannelID {
		nOldChannleID := tAccount.ChannelID
		tAccount.ChannelID = nNewChannelID
		tAccount.Save(true)
		log.Infof("%v设置帐号:%v的渠道ID; 旧渠道:%v, 新渠道:%v", strFrom, nAccountID, nOldChannleID, nNewChannelID)
		return 0
	}
	return 1
}

func (self *accountMgr) SetUpSalesmenType(nAccountID uint32, nSetSalesman uint32, strWebAccount string, strWebPassword string) uint8 {

	switch nSetSalesman {
	case types.SALESMAN_COMMON.Value():
	case types.SALESMAN_DA_QU.Value():
	case types.SALESMAN_CLUB.Value():
	default:
		log.Warnf("SetUpSalesmenType Illegal parameters; AccountID:%v, SetSalesman:%v", nAccountID, nSetSalesman)
		return 1
	}

	tAccount := self.GetAccountByID(nAccountID)
	if tAccount == nil {
		log.Warnf("SetUpSalesmenType AccountID:%v, 找不到帐号;  SetSalesman:%v", nAccountID, nSetSalesman)
		return 2
	}

	if tAccount.Salesman == types.SALESMAN_DA_QU.Value() {
		log.Warnf("SetUpSalesmenType AccountID:%v, 已经是大区代理身份, 不能提升代理身份, SetSalesman:%v;", nAccountID, nSetSalesman)
		return 3
	} else if tAccount.Salesman == types.SALESMAN_CLUB.Value() {
		log.Warnf("SetUpSalesmenType AccountID:%v, 已经是俱乐部代理身份, 不能提升代理身份, SetSalesman:%v;", nAccountID, nSetSalesman)
		return 4
	} else if tAccount.Salesman >= nSetSalesman {
		log.Warnf("SetUpSalesmenType 重复设置; AccountID:%v, Account.nSalesman:%v, SetSalesman:%v", nAccountID, tAccount.Salesman, nSetSalesman)
		return 5
	}

	if tAccount.BindCode > 0 {
		tOneUper := AccountMgr.GetAccountByID(tAccount.BindCode)
		if tOneUper != nil {
			if nSetSalesman == types.SALESMAN_CLUB.Value() && tOneUper.Salesman != types.SALESMAN_CLUB.Value() {
				log.Warnf("SetUpSalesmenType AccountID:%v, 俱乐部代理只能绑定俱乐部代理, 当前:%v SetSalesman:%v, UperSalesman:%v;", nAccountID, tAccount.Salesman, nSetSalesman, tOneUper.Salesman)
				return 6
			} else if nSetSalesman == types.SALESMAN_COMMON.Value() && tOneUper.Salesman != types.SALESMAN_COMMON.Value() && tOneUper.Salesman != types.SALESMAN_DA_QU.Value() {
				log.Warnf("SetUpSalesmenType AccountID:%v, 普通代理的上级只能是普通代理或大区代理, 当前:%v SetSalesman:%v, UperSalesman:%v;", nAccountID, tAccount.Salesman, nSetSalesman, tOneUper.Salesman)
				return 7
			} else if nSetSalesman == types.SALESMAN_DA_QU.Value() && tOneUper.Salesman != types.SALESMAN_DA_QU.Value() {
				log.Warnf("SetUpSalesmenType AccountID:%v, 大区代理的上级只能是大区代理, 当前:%v SetSalesman:%v, UperSalesman:%v;", nAccountID, tAccount.Salesman, nSetSalesman, tOneUper.Salesman)
				return 8
			}
		}
	}

	nOldSalesman := tAccount.Salesman
	tAccount.Salesman = nSetSalesman
	tAccount.Save(true)
	log.Infof("设置代理身份, AccountID:%v, OldSalesman:%v  NewSalesman:%v WebAccount:%v WebPassword:%v", nAccountID, nOldSalesman, nSetSalesman, strWebAccount, strWebPassword)

	// 发送通知成为代理的邮件
	if strWebAccount != "" && strWebPassword != "" {
		strContent := strWebAccount + "#$$#" + strWebPassword
		EmailMgr.AddMail(nAccountID, types.EMAIL_SALESMAN.Value(), strContent, 0)
	}

	if tAccount.IsOnline() == true {
		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(protomsg.Old_MSGID_UPDATE_SALESMAN.UInt16())
		tSend.WriteUInt8(uint8(tAccount.Salesman))
		send_tools.Send2Account(tSend.GetData(), tAccount.SessionId)
	}
	return 0
}

func (self *accountMgr) SetDownSalesmenType(nAccountID uint32, nSetSalesman uint32) uint8 {

	switch nSetSalesman {
	case types.SALESMAN_NULL.Value():
	case types.SALESMAN_COMMON.Value():
	default:
		log.Warnf("SetUpSalesmenType Illegal parameters; AccountID:%v, SetSalesman:%v", nAccountID, nSetSalesman)
		return 1
	}

	tAccount := self.GetAccountByID(nAccountID)
	if tAccount == nil {
		log.Warnf("SetDownSalesmenType AccountID:%v, 找不到帐号;  SetSalesman:%v", nAccountID, nSetSalesman)
		return 1
	}
	if tAccount.Salesman == types.SALESMAN_CLUB.Value() && nSetSalesman != types.SALESMAN_NULL.Value() {
		log.Warnf("SetDownSalesman AccountID:%v, 俱乐部代理只能降级成非代理:%v", nAccountID, nSetSalesman)
		return 2
	} else if tAccount.Salesman == types.SALESMAN_DA_QU.Value() && nSetSalesman != types.SALESMAN_NULL.Value() && nSetSalesman != types.SALESMAN_COMMON.Value() {
		log.Warnf("SetDownSalesman AccountID:%v, 大区代理只能降级成普通代理或者非代理:%v", nAccountID, nSetSalesman)
		return 3
	} else if tAccount.Salesman == types.SALESMAN_COMMON.Value() && nSetSalesman != types.SALESMAN_NULL.Value() {
		log.Warnf("SetDownSalesman AccountID:%v, 普通代理只能降级成非代理:%v", nAccountID, nSetSalesman)
		return 4
	}

	for _, tDown := range self.AccountbyID {
		if tDown.BindCode == nAccountID && tDown.Salesman > nSetSalesman {
			log.Warnf("SetDownSalesman 下级身份高于降级后自己的身份; AccountID:%d, 降级后身份:%d, 下级ID:%d, 下级身份:%d", nAccountID, nSetSalesman, tDown.AccountId, tDown.Salesman)
			return 5
		}
	}

	nOldSalesman := tAccount.Salesman
	tAccount.Salesman = nSetSalesman
	tAccount.Save(true)
	log.Infof("强制降级代理身份, AccountID:%v, OldSalesman:%v  NewSalesman:%v", nAccountID, nOldSalesman, nSetSalesman)

	if tAccount.Salesman == types.SALESMAN_NULL.Value() {
		EmailMgr.RemoveSalesmanTypeEmail(nAccountID)
	}

	if tAccount.IsOnline() == true {
		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(protomsg.Old_MSGID_UPDATE_SALESMAN.UInt16())
		tSend.WriteUInt8(uint8(tAccount.Salesman))
		send_tools.Send2Account(tSend.GetData(), tAccount.SessionId)
	}
	return 0
}

func (self *accountMgr) ModifyHeadURL(nAccountID uint32, strNewHeadURL string, isFromClient, isSave bool, nGameSessionID int64) uint8 {
	tAccount := self.GetAccountByID(nAccountID)
	if tAccount == nil {
		if isFromClient {
			tSend := packet.NewPacket(nil)
			tSend.SetMsgID(protomsg.Old_MSGID_MODIFY_HEADURL.UInt16())
			tSend.WriteUInt8(1)
			send_tools.Send2Account(tSend.GetData(), tAccount.SessionId)
		} else {
			log.Warnf("ModifyHeadURL Can't Find Account:%v, NewHeadURL:%v", nAccountID, strNewHeadURL)
		}
		return 1
	}

	nLen := len(strNewHeadURL)
	isHave := config.IsHaveBannedWords(strNewHeadURL)
	if isHave == true || nLen > 256 {
		if isFromClient {
			tSend := packet.NewPacket(nil)
			tSend.SetMsgID(protomsg.Old_MSGID_MODIFY_HEADURL.UInt16())
			tSend.WriteUInt8(2)
			send_tools.Send2Account(tSend.GetData(), tAccount.SessionId)
		} else {
			log.Warnf("ModifyHeadURL Have BannedWords, AccountID:%v, NewHeadURL:%v", nAccountID, strNewHeadURL)
		}
		return 2
	}

	tAccount.HeadURL = strNewHeadURL
	if isSave == true {
		tAccount.Save(true)
	}

	if nGameSessionID > 0 {
		tSendToGame := packet.NewPacket(nil)
		tSendToGame.SetMsgID(protomsg.Old_MSGID_CHANGE_PLAYER_INFO.UInt16())
		tSendToGame.WriteUInt32(nAccountID)
		tSendToGame.WriteUInt8(2) // 修改头像URL
		tSendToGame.WriteString(strNewHeadURL)
		send_tools.Send2Game(tSendToGame.GetData(), nGameSessionID)
	}

	if tAccount.IsOnline() == true {
		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(protomsg.Old_MSGID_MODIFY_HEADURL.UInt16())
		tSend.WriteUInt8(0)
		tSend.WriteString(strNewHeadURL)
		send_tools.Send2Account(tSend.GetData(), tAccount.SessionId)
	}
	return 0
}

func (self *accountMgr) ModifyName(nAccountID uint32, strNewName string, isFromClient, isSave bool, nGameSessionID, nSessionID int64) uint8 {
	tAccount := self.GetAccountByID(nAccountID)
	if tAccount == nil {
		if isFromClient {
			tSend := packet.NewPacket(nil)
			tSend.SetMsgID(protomsg.Old_MSGID_MODIFY_NAME.UInt16())
			tSend.WriteUInt8(1)
			send_tools.Send2Account(tSend.GetData(), nSessionID)
		} else {
			log.Warnf("ModifyName Can't Find Account:%v, NewName:%v", nAccountID, strNewName)
		}
		return 1
	}

	nLen := utf8.RuneCountInString(strNewName)
	isHave := config.IsHaveBannedWords(strNewName)
	if nLen <= 0 || nLen > 12 || isHave == true {
		if isFromClient {
			tSend := packet.NewPacket(nil)
			tSend.SetMsgID(protomsg.Old_MSGID_MODIFY_NAME.UInt16())
			tSend.WriteUInt8(2)
			send_tools.Send2Account(tSend.GetData(), nSessionID)
		} else {
			log.Warnf("ModifyName Have BannedWords, AccountID:%v, NewName:%v", nAccountID, strNewName)
		}
		return 2
	}

	tAccount.Name = strNewName
	if isSave == true {
		tAccount.Save()
	}

	if nGameSessionID > 0 {
		tSendToGame := packet.NewPacket(nil)
		tSendToGame.SetMsgID(protomsg.Old_MSGID_CHANGE_PLAYER_INFO.UInt16())
		tSendToGame.WriteUInt32(nAccountID)
		tSendToGame.WriteUInt8(1) // 修改名字
		tSendToGame.WriteString(strNewName)
		send_tools.Send2Game(tSendToGame.GetData(), nGameSessionID)
	}

	if tAccount.IsOnline() == true {
		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(protomsg.Old_MSGID_MODIFY_NAME.UInt16())
		tSend.WriteUInt8(0)
		tSend.WriteString(strNewName)
		send_tools.Send2Account(tSend.GetData(), nSessionID)
	}
	return 0
}

// 参数: 是否汇总; 每日凌晨24点时才汇总
func (self *accountMgr) UpdateOnlinePlayer(isGather bool, isPrint bool) {

	fSendOnlineLog := func(mPlayerCount map[uint8]uint16) {
		nNN := mPlayerCount[common.EGameTypeNIU_NIU.Value()]
		nJH := mPlayerCount[common.EGameTypeJIN_HUA.Value()]
		nCX := mPlayerCount[common.EGameTypeCHE_XUAN.Value()]
		nWZQ := mPlayerCount[common.EGameTypeWU_ZI_QI.Value()]
		nSSZZ := mPlayerCount[common.EGameTypeSHEN_SHOU_ZHI_ZHAN.Value()]
		nTTZ := mPlayerCount[common.EGameTypeTUI_TONG_ZI.Value()]
		nLHD := mPlayerCount[common.EGameTypeLONG_HU_DOU.Value()]
		nHHDZ := mPlayerCount[common.EGameTypeHONG_HEI_DA_ZHAN.Value()]
		nWHNN := mPlayerCount[common.EGameTypeWUHUA_NIUNIU.Value()]
		nDEH := mPlayerCount[common.EGameTypeDING_ER_HONG.Value()]
		nDGK := mPlayerCount[common.EGameTypeDGK.Value()]
		nXMMJ := mPlayerCount[common.EGameTypeXMMJ.Value()]
		nPDK := mPlayerCount[common.EGameTypePAO_DE_KUAI.Value()]
		nSSS := mPlayerCount[common.EGameTypeSHI_SAN_SHUI.Value()]
		nHB := mPlayerCount[common.EGameTypeHONG_BAO.Value()]
		nTNN := mPlayerCount[common.EGameTypeTEN_NIU_NIU.Value()]
		nFQZS := mPlayerCount[common.EGameTypeFQZS.Value()]
		nPDKHN := mPlayerCount[common.EGameTypePDK_HN.Value()]
		nSG := mPlayerCount[common.EGameTypeSAN_GONG.Value()]
		nAll := mPlayerCount[0]

		if isPrint == true {
			for nGameType, nCount := range mPlayerCount {
				if nGameType > 0 && nCount > 0 {
					fmt.Printf("==== %v 在线人数:%v\r\n", common.EGameType(nGameType), nCount)
				}
			}
		}

		strNowTime := time.Now().Format("2006-01-02")
		strSQL := fmt.Sprintf("REPLACE INTO log_online(log_Time, log_Online, log_NN, log_JH, log_CX, log_SSZZ, log_TTZ, log_WZQ, log_DEH, log_WHNN, log_SSS, log_LHD, log_HHDZ, log_HB, log_DGK, log_PDK, log_XMMJ, log_TNN, log_FQZS, log_PDKHN, log_SG) VALUES ('%v', %v, %v, %v, %v, %v, %v, %v, %v, %v, %v, %v, %v, %v, %v, %v, %v, %v, %v, %v, %v)", strNowTime, nAll, nNN, nJH, nCX, nSSZZ, nTTZ, nWZQ, nDEH, nWHNN, nSSS, nLHD, nHHDZ, nHB, nDGK, nPDK, nXMMJ, nTNN, nFQZS, nPDKHN, nSG)
		send_tools.SQLLog(self.StampNum(), strSQL)
	}

	if isGather == false {
		mPlayerCount := make(map[uint8]uint16)
		for nGameType := range self.playerTopCount {
			mPlayerCount[nGameType] = 0
		}
		for _, tAccount := range self.AccountbyID {
			if tAccount.Robot == 0 && tAccount.IsOnline() == true {
				mPlayerCount[0]++
				if tAccount.GameType > 0 {
					mPlayerCount[uint8(tAccount.GameType)]++
				}
			}
		}
		// 记录最高值
		for nGameType, nCount := range mPlayerCount {
			nTop := self.playerTopCount[nGameType]
			if nCount > nTop {
				self.playerTopCount[nGameType] = nCount
			}
		}
		fSendOnlineLog(mPlayerCount)

	} else {
		var nTotal uint16
		for nGameType, nCount := range self.playerTopCount {
			if nGameType > 0 {
				nTotal += nCount
			}
		}
		self.playerTopCount[0] = nTotal
		fSendOnlineLog(self.playerTopCount)
		for nGameType := range self.playerTopCount {
			self.playerTopCount[nGameType] = 0
		}
	}
}


func (self *accountMgr) SendLoginLog(accountId uint32) {
	if _, isExist := self.mLoginLog[accountId]; isExist == false {
		tLog := &login{LoginTime: utils.DateString()}
		self.mLoginLog[accountId] = tLog
	}
}

func (self *accountMgr) SendExitLog(accountId uint32) {
	if tNode, isExist := self.mLoginLog[accountId]; isExist == true {
		if tNode.LogoutTime == "" {
			tNode.LogoutTime = utils.DateString()
			strLog := fmt.Sprintf("INSERT INTO log_login(log_AccountID,log_LoginTime,log_LogoutTime) VALUES (%v, '%v','%v')", accountId, tNode.LoginTime, tNode.LogoutTime)
			send_tools.SQLLog(strLog)
		}
	}
}

// 每小时清空一次登录记录, 即每小时记录一次登录日志
func (self *accountMgr) ClearLoginLog() {
	self.mLoginLog = make(map[uint32]*login)
}

func CheckSession(accountId uint32, session int64) *Account {
	sacc := AccountMgr.GetAccountByID(accountId)
	var seAcc *Account
	seAccId := uint32(0)
	seAccName := ""
	if session != 0 {
		seAcc = AccountMgr.GetAccountBySessionID(session)
		if seAcc != nil {
			seAccId = seAcc.AccountId
			seAccName = seAcc.Name
		}
	}

	if sacc == nil {
		log.Errorf("作弊, session:%v 验证的accountId:%v session对应的玩家 accid:%v,name:%v", session, accountId, seAccId, seAccName)
		panic(nil)
	} else if sacc.SessionId != session {
		log.Errorf("作弊, session:%v accountID:%v 验证的session:%v accountId:%v Robot:%v", session, accountId, sacc.SessionId, sacc.AccountId, sacc.Robot)
		panic(nil)
	} else {
		return sacc
	}
}
