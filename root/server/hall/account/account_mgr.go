package account

import (
	"github.com/golang/protobuf/proto"
	"root/common"
	"root/common/config"
	"root/common/model/rediskey"
	"root/core/db"
	"root/core/log"
	"root/core/utils"
	"root/protomsg"
	"root/server/hall/send_tools"
	"root/server/hall/types"
	"strconv"
)

var AccountMgr = newAccountMgr()

type (
	accountMgr struct {
		saveChange         []*Account
		accountbyUnDevice  map[string]*Account
		accountbyPhone     map[string]*Account
		accountbyWeiXin    map[string]*Account
		AccountbyID        map[uint32]*Account
		accountbySessionID map[int64]*Account
		IDAssign []uint32
	}
)

func newAccountMgr() *accountMgr {
	ret := &accountMgr{
		saveChange:         make([]*Account, 0, 1000),
		accountbyUnDevice:  make(map[string]*Account),
		accountbyPhone:     make(map[string]*Account),
		accountbyWeiXin:    make(map[string]*Account),
		AccountbyID:        make(map[uint32]*Account),
		accountbySessionID: make(map[int64]*Account),
		IDAssign:           make([]uint32, 0, 100000),
	}
	return ret
}
// 所有玩家和机器人都初始化完成以后, 再将玩家和机器人的ID排除掉
func (self *accountMgr) CollatingIDAssign() {
	mCheckID := make(map[uint32]bool)
	for nID := range self.AccountbyID {
		mCheckID[nID] = true
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
	log.Infof("=========== 注册玩家帐号:%v, 可分配新帐号ID个数:%v", nPlayer, len(self.IDAssign))
}

func (self *accountMgr) GetAllAccount() map[uint32]*Account {
	return self.AccountbyID
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
func (self *accountMgr) GetAccountByIDAssert(id uint32) *Account {
	acc := self.AccountbyID[id]
	if acc == nil {
		log.Panicf("找不到玩家:%v ", id)
	}
	return acc
}

func (self *accountMgr) GetAccountBySessionIDAssert(session int64) *Account {
	if session == 0 {
		log.Panicf("找不到玩家:%v ",session)
		return nil
	}
	return self.accountbySessionID[session]
}

func (self *accountMgr) RemoveAccountBySessionID(session int64) {
	delete(self.accountbySessionID, session)
}


func (self *accountMgr) ArchiveAll() {
	count := 0
	for _,acc := range self.AccountbyID{
		if acc.Store && acc.Robot == 0{
			acc.Save()
			count++
		}
	}

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

// 广播消息, 给所有在线玩家
// t类型1, 所有在大厅玩家接收
// t类型2, 所有在大厅和在房间的玩家接收
func (self *accountMgr) SendBroadcast(msgID uint16,pb proto.Message, t uint8) {
	if t == 1 {
		for _, acc := range self.AccountbyID {
			if acc.IsOnline() && acc.Robot == 0 && acc.RoomID == 0 {
				send_tools.Send2Account(msgID,pb, acc.SessionId)
			}
		}
	} else {
		for _, acc := range self.AccountbyID {
			if acc.IsOnline() && acc.Robot == 0 {
				send_tools.Send2Account(msgID,pb, acc.SessionId)
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
	tNewAccount.Money = 0
	tNewAccount.SessionId = session
	tNewAccount.Robot = nRobot
	tNewAccount.ActiveIP = strClientIP
	tNewAccount.ActiveTime = strNowTime
	tNewAccount.OSType = uint32(nOSType)

	self.AccountbyID[nNewAccountID] = tNewAccount
	tNewAccount.AddMoney(config.GetPublicConfig_Int64("3"),common.EOperateType_INIT)

	if nRobot == 0 {
		self.accountbySessionID[session] = tNewAccount
		tNewAccount.Save() 		// 新建账号，存数据库
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

func (self *accountMgr) LoginAccount(acc *Account, nLoginType uint8, IP string, session int64) {
	if acc.SessionId > 0 && acc.SessionId != session {
		send_tools.Send2Account(protomsg.MSG_SC_KICK_OUT_HALL.UInt16(),&protomsg.KICK_OUT_HALL{Ret:2}, acc.SessionId)
		// 取消之前的SessionID的关联
		delete(self.accountbySessionID, acc.SessionId)
	}
	acc.SessionId = session
	acc.LoginTime = utils.SecondTimeSince1970()

	// 建立新的关联
	self.accountbySessionID[session] = acc
	if acc.Robot == 0 {
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
			log.Infof("login Player:%v, unique:%v, Money:%v, 登录类型:%v,IP:%v Session:%v", acc.AccountId,acc.UnDevice, acc.Money, types.ELoginType(nLoginType),IP, session)
		})
	}
}

func CheckSession(accountId uint32, session int64) *Account {
	sacc := AccountMgr.GetAccountByID(accountId)
	var seAcc *Account
	seAccId := uint32(0)
	seAccName := ""
	if session != 0 {
		seAcc = AccountMgr.GetAccountBySessionIDAssert(session)
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
