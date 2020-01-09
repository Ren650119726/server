package logic

import (
	"root/common"
	"root/common/config"
	"root/common/tools"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"root/protomsg"
	"root/server/hall/account"
	"root/server/hall/event"
	"root/server/hall/send_tools"
	"root/server/hall/server"
	"root/server/hall/types"
	"sort"
	"strings"
)

var HallMgr = newHallMgr()

type (
	auto_create_node struct {
		nBet       uint32
		nMax       uint16
		nFreeSeat  uint16
		nMatchType uint8
	}

	game_node struct {
		nNodeID   uint16
		nUseCount uint16
	}

	// 创建信息节点
	createNode struct {
		nNewRoomID      uint32
		nGameType       uint8
		nMatchType      uint8
		strParam        string
		nOwnerID        uint32
		nServerID       uint16
		nAnswerProtocol uint16
		nClubID         uint32
		nClubmgr        uint32 // 是否是俱乐部管理员创建
	}

	// 房间结构
	room struct {
		nRoomID     uint32          // 房间ID
		nGameType   uint8           // 游戏类型
		nMax        uint16          // 当前最大人数
		nServerID   uint16          // 房间所在ServerID
		nMatchType  uint8           // 匹配类型
		iCreateTime int64           // 创建时间
		strParam    string          // 房间参数
		isLock      bool            // 是否为加密房间
		nOwnerID    uint32          // 创建房间者, 为0也表示系统创建
		nBet        uint32          // 底注金额
		nEnter      uint32          // 入场条件
		nLeave      uint32          // 离场条件
		mSeatID     map[uint32]bool // 坐下玩家列表
		mWatchID    map[uint32]bool // 观战玩家列表
		nGameNodeID uint16          // 游戏节点ID
		clubID      uint32          // 所属俱乐部 0 公共房间
		nClubmgr    uint32          // 是否是俱乐部管理员创建
	}

	bonusPool struct {
		BonusPool  string // 奖金池json
		HistoryTop string // 历史最高记录json
	}

	waterLine struct {
		GameType  uint8  // 游戏类型
		WaterLine string // 水位线
	}

	accMap  map[uint32]*account.Account // map[帐号ID][玩家指针]
	gameMap map[uint8][]accMap          // map[游戏类型][档次列表]
	roomMap map[uint32]*room            // map[房间ID][房间指针]

	hallMgr struct {
		mCreateTable            map[uint32]*createNode // 所有游戏的创建列表;  key:游戏房间ID value:创建信息节点
		mRoomTable              roomMap                // 所有房间元数据列表;  key:游戏房间ID value:房间数据
		mUpdateDesk             map[uint32]gameMap     // 更新类对应请求玩家;  key:俱乐部ID value: gameMap
		mBonusPool              map[uint16]*bonusPool  // 奖金池map; key:nServerID value: 奖金池对象
		mWaterLine              map[uint16]*waterLine  // 水位线map; key:nServerID value: 水位线对象
		ipNodes                 map[int]string         //
		mSaveBonusPoolTime      map[uint16]int64       // 回存奖金池的时间
		mSaveWaterLineTime      map[uint16]int64       // 回存水位线的时间
		sSaveServiceFeeLog      [10][]string           // 服务费日志缓存表
		sExchangeChan           chan string            // 兑换请求缓存队列
		nMaintenanceTime        uint32
		nLieExchangeNotify      int64
		nNextEnterRoomTime      int64
		nNextEnterMatchRoomTime int64
		isTestCharge            bool
		webConfig               string
		payChannel              map[uint32]string
		exchangeConfig          string
		ListenActor             *core.Actor

		OpenDesk    uint32 // 0 关闭 1 开放
	}
)

func newHallMgr() *hallMgr {

	hall := &hallMgr{
		mCreateTable: make(map[uint32]*createNode),
		mRoomTable:   make(roomMap),
		mUpdateDesk:  make(map[uint32]gameMap),
		mBonusPool:   make(map[uint16]*bonusPool),
		mWaterLine:   make(map[uint16]*waterLine),
		//sSaveServiceFeeLog:      make([]string, 0, 100),
		mSaveBonusPoolTime:      make(map[uint16]int64),
		mSaveWaterLineTime:      make(map[uint16]int64),
		sExchangeChan:           make(chan string, 2000),
		nMaintenanceTime:        0,
		nLieExchangeNotify:      0,
		nNextEnterRoomTime:      0,
		nNextEnterMatchRoomTime: 0,
		isTestCharge:            false,
		OpenDesk:                1,
	}

	for i, _ := range hall.sSaveServiceFeeLog {
		hall.sSaveServiceFeeLog[i] = make([]string, 0, 100)
	}
	event.Dispatcher.AddEventListener(event.EventType_UpdateCharge, hall)
	return hall
}

func (self *hallMgr) GetRoom(nRoomID uint32) *room {
	return self.mRoomTable[nRoomID]
}

func (self *room) GetSeatCount() uint16 {
	return uint16(len(self.mSeatID))
}
func (self *room) GetWatchCount() uint16 {
	return uint16(len(self.mWatchID))
}

func (self *room) FillRoomSeatInfo(tSend packet.IPacket) {

	var nWritePlayerCount uint16
	nWritePlayerPos := tSend.GetWritePos()
	tSend.WriteUInt16(0)

	isWanRenChang := HallMgr.isWanRenChang(self.nGameType)
	if isWanRenChang == true {
		WRC_SHOW_SEAT_COUNT := config.GetPublicConfig_Int64("WRC_SHOW_SEAT_COUNT")
		for key := range self.mSeatID {
			tAccount := account.AccountMgr.GetAccountByID(key)
			if tAccount != nil {
				if tAccount.Index < 7 {
					if WRC_SHOW_SEAT_COUNT <= 0 {
						continue
					} else {
						WRC_SHOW_SEAT_COUNT--
					}
				}
				nWritePlayerCount++
				tSend.WriteUInt8(uint8(tAccount.Index))
				tSend.WriteString(tAccount.HeadURL)
			}
		}
	} else {
		for key := range self.mSeatID {
			tAccount := account.AccountMgr.GetAccountByID(key)
			if tAccount != nil {
				nWritePlayerCount++
				tSend.WriteUInt8(uint8(tAccount.Index))
				tSend.WriteString(tAccount.HeadURL)
			}
		}
	}
	tSend.Rrevise(nWritePlayerPos, nWritePlayerCount)
}

func (self *room) GetRobotSeatCount() uint16 {
	nRobotCount := uint16(0)
	for nAccountID := range self.mSeatID {
		tAccount := account.AccountMgr.GetAccountByID(nAccountID)
		if tAccount != nil && tAccount.Robot > 0 {
			nRobotCount++
		}
	}
	return nRobotCount
}
func (self *room) GetRobotWatchCount() uint16 {
	nRobotCount := uint16(0)
	for nAccountID := range self.mWatchID {
		tAccount := account.AccountMgr.GetAccountByID(nAccountID)
		if tAccount != nil && tAccount.Robot > 0 {
			nRobotCount++
		}
	}
	return nRobotCount
}

func (self *hallMgr) IsInRoom(nRoomID, nAccountID uint32) bool {
	tRoom := self.GetRoom(nRoomID)
	if tRoom == nil {
		return false
	} else {
		_, isInSeat := tRoom.mSeatID[nAccountID]
		_, isInWatch := tRoom.mWatchID[nAccountID]
		if isInSeat == false && isInWatch == false {
			return false
		}
	}
	return true
}

// 每30秒执行一次组装SQL和回存
func (self *hallMgr) AddServiceFeeLog(accid uint32, strLog string) {
	portion := int(accid % 10)
	self.sSaveServiceFeeLog[portion] = append(self.sSaveServiceFeeLog[portion], strLog) //
}

// 每30秒执行一次组装SQL和回存
func (self *hallMgr) UpdateServiceFeeLog() {
	for i, _ := range self.sSaveServiceFeeLog {
		self.SendServiceFeeLog(i)
	}
}

// 每30秒执行一次组装SQL和回存
func (self *hallMgr) SendServiceFeeLog(portion int) {
	if portion < 0 || portion > 9 {
		log.Errorf("不在0-9 之间portion:%v", portion)
		return
	}
	logs := self.sSaveServiceFeeLog[portion]
	nLen := len(logs)
	if nLen <= 0 {
		return
	}

	strSQL := fmt.Sprintf("INSERT INTO log_servicefee_%v (log_AccountID,log_ServiceFee,log_GameType,log_Time,log_RoomID, log_ClubID) VALUES ", portion)
	for i := 0; i < nLen; i++ {
		strNode := logs[i]
		strSQL += strNode
	}

	strSQL = strings.TrimRight(strSQL, ",")
	send_tools.SQLLog(account.AccountMgr.StampNum(), strSQL)
	log.Info(strSQL)

	self.sSaveServiceFeeLog[portion] = make([]string, 0, 100)
}

func (self *hallMgr) SaveAllWaterLine() {

	for nServerID, tNode := range self.mWaterLine {
		tLine := &protomsg.WaterLine{ServerID: uint32(nServerID), GameType: uint32(tNode.GameType), WaterLine: tNode.WaterLine}
		send_tools.Send2DB(account.AccountMgr.StampNum(), protomsg.MSGID_HG_SAVE_WATER_LINE.UInt16(), &protomsg.SAVE_WATER_LINE{Line: tLine}, false)
		log.Infof("SQL_INST: UPDATE gd_water_line SET gd_WaterLine='%v' WHERE gd_ServerID=%v;", tNode.WaterLine, nServerID)
	}
}
func (self *hallMgr) SaveAllBonusPool() {

	for nServerID, tNode := range self.mBonusPool {
		strHistoryTop := tNode.HistoryTop
		if tNode.HistoryTop == "" {
			strHistoryTop = "{}"
		}
		tPool := &protomsg.BonusPool{ServerID: uint32(nServerID), BonusPool: tNode.BonusPool, HistoryTop: strHistoryTop}
		send_tools.Send2DB(account.AccountMgr.StampNum(), protomsg.MSGID_HG_SAVE_BONUS_POOL.UInt16(), &protomsg.SAVE_BONUS_POOL{Pool: tPool}, false)
		log.Infof("SQL_INST: UPDATE gd_bonus_pool SET gd_BonusPool='%v', gd_HistoryTop='%v' WHERE gd_ServerID=%v;", tNode.BonusPool, tNode.HistoryTop, nServerID)
	}
}

func (self *hallMgr) SaveWebData() {
	send_tools.Send2DB(account.AccountMgr.StampNum(), protomsg.MSGID_HG_SAVE_ALL_WEB_DATA.UInt16(), &protomsg.WebInfo{
		Webconfig:      self.webConfig,
		Exchangeconfig: self.exchangeConfig,
		PayChannel:     self.payChannel,
	}, false)
}

func (self *hallMgr) PrintSign(strServerIP string) {
	if config.GetPublicConfig_Int64("APP_STORE") == 1 {
		fmt.Println("=========== 审核标志:审核版")
	} else {
		fmt.Println("=========== 审核标志:正式版")
	}
	if config.GetPublicConfig_Int64("WHITE_LIST_OPEN") == 1 {
		fmt.Printf("=========== 白名单功能:已开启;          ServerIP:%v\r\n", strServerIP)
	} else {
		fmt.Printf("=========== 白名单功能:已关闭;          ServerIP:%v\r\n", strServerIP)
	}
	if HallMgr.isTestCharge == true {
		log.Infof(colorized.Green("=========== 登录加钱功能:已开启, 若要关闭请在控制台输入test 0"))
		log.Infof(colorized.Green("=========== 登录加钱功能:已开启, 若要关闭请在控制台输入test 0"))
		log.Infof(colorized.Green("=========== 登录加钱功能:已开启, 若要关闭请在控制台输入test 0"))
	} else {
		log.Infof(colorized.Green("=========== 登录加钱功能:已关闭, 请谨慎开启此功能, 正式服务器不允许开启此功能"))
		log.Infof(colorized.Green("=========== 登录加钱功能:已关闭, 请谨慎开启此功能, 正式服务器不允许开启此功能"))
		log.Infof(colorized.Green("=========== 登录加钱功能:已关闭, 请谨慎开启此功能, 正式服务器不允许开启此功能"))
	}
}

func (self *hallMgr) runExchangeRequest() {
	core.Gwg.Add(1)
	defer func() {
		core.Gwg.Done()
	}()

	EXCHANGE_URL := ""

	const CONTENT_TYPE = "application/x-www-form-urlencoded"
	for {
		select {
		case strJson := <-self.sExchangeChan:
			_, strlocalIP, _ := config.IsTestServer()
			DD_HALL_IP := config.GetPublicConfig_String("DD_HALL_IP")
			if strlocalIP == DD_HALL_IP {
				EXCHANGE_URL = config.GetPublicConfig_String("DD_EXCHANGE_URL")
			} else {
				EXCHANGE_URL = config.GetPublicConfig_String("HH_EXCHANGE_URL")
			}

			// 组装签名字段
			SIGN_KEY := config.GetPublicConfig_String("SIGN_KEY")
			strSign := fmt.Sprintf("%v%v", strJson, SIGN_KEY)
			strSign = tools.MD5(strSign)

			strParam := fmt.Sprintf("param=%v&sign=%v", strJson, strSign)
			resp, err := http.Post(EXCHANGE_URL, CONTENT_TYPE, strings.NewReader(strParam))
			if err != nil {
				log.Fatalf("兑现请求异常, 发送不成功, 提现:%v 错误:%v; 转为人工处理", strParam, err.Error())
				pborder := &protomsg.ExchangeOrder{}
				jerr := json.Unmarshal([]byte(strJson), pborder)
				if jerr == nil {
					// 发送失败, 转为人工审核
					core.LocalCoreSend(0, common.EActorType_MAIN.Int32(), func() {
						account.AccountMgr.UpdateExchangeOrder(pborder.Order, common.EXCHANGE_STATE_MANUAL_REVIEW.Value(), "", "")
					})
				}
				break
			}
			eclose := resp.Body.Close()
			if eclose != nil {
				log.Fatalf("Error, 关闭https resps失败, 错误:%v", eclose.Error())
			}
			if resp.StatusCode != 200 {
				log.Fatalf("兑现请求异常, 返回失败, 提现:%v 状态码:%v", strParam, resp.Status)
				break
			}

			log.Infof("兑现请求成功, 内容:%v", strParam)
		}
	}
}

func (self *hallMgr) sendRoomList(nGameType uint8, nClubID uint32, nSessionID int64) {
	mGameMap := self.getGameMap(nGameType)
	tSend := packet.NewPacket(nil)
	tSend.SetMsgID(protomsg.Old_MSGID_GET_ROOM_LIST.UInt16())
	tSend.WriteUInt8(0)
	tSend.WriteUInt8(nGameType)
	nWritePos := tSend.GetWritePos()
	var nWriteCount uint16
	tSend.WriteUInt16(nWriteCount)
	for _, tRoom := range mGameMap {
		if tRoom.nMatchType == 0 && tRoom.nGameType == nGameType && tRoom.clubID == nClubID {
			nWriteCount++
			tSend.WriteUInt32(tRoom.nRoomID)
			tSend.WriteUInt16(tRoom.GetSeatCount())
			tSend.WriteInt64(tRoom.iCreateTime)
			tSend.WriteString(tRoom.strParam)
		}
	}
	tSend.Rrevise(nWritePos, nWriteCount)
	send_tools.Send2Account(tSend.GetData(), nSessionID)
}

func (self *hallMgr) createRoom(tAccount *account.Account, nGameType uint8, nMatchType uint8, strParam string, nAnswerProtocol uint16, nClubID uint32) uint8 {

	if self.nMaintenanceTime > 0 {
		return 4
	}

	isWanRenChang := self.isWanRenChang(nGameType)
	if isWanRenChang == true && nClubID > 0 {
		return 6
	}

	isSystemCreate := nAnswerProtocol == protomsg.Old_MSGID_SYSTEM_CREATE_ROOM.UInt16()
	if tAccount != nil {
		nCheckBindCode := config.GetPublicConfig_Int64("CHECK_BIND_CODE")
		if nCheckBindCode == 1 {
			if tAccount.BindCode <= 0 && tAccount.Robot == 0 {
				return 5
			}
		}

		if tAccount.RoomID > 0 && isSystemCreate == false {
			return 2
		}
	} else {
		if isSystemCreate == false {
			return 1
		}
	}

	sServerList := server.ServerMgr.GetServerList(nGameType)
	nServerLen := len(sServerList)
	if nServerLen <= 0 || sServerList[0].IsMaintenance == true {
		return 4
	}

	var nOwnerID uint32
	if tAccount != nil {
		nOwnerID = tAccount.AccountId // 用于标记一个玩家只能创建2个房间
	}
	if nOwnerID > 0 && isSystemCreate == false {
		// 系统创建房间, 不限制创建数量
		nCanCreateRoom := self.canCreateRoom(tAccount, nGameType, nMatchType, nClubID, strParam, nAnswerProtocol)
		if nCanCreateRoom > 0 {
			return nCanCreateRoom
		}
	}

	nClubMgr := uint32(0)
	if c, e := ClubMgr.Clubs[nClubID]; e && tAccount != nil {
		if tNode, isExist := c.Member[tAccount.AccountId]; isExist == true {
			nClubMgr = tNode.Manager
		}
	}
	nRandIndex := utils.Randx_y(0, nServerLen)
	tServerNode := sServerList[nRandIndex]
	nServerID := tServerNode.ServerID
	nNewRoomID := self.getNewRoomID()
	tCreateNode := &createNode{
		nNewRoomID:      nNewRoomID,
		nGameType:       nGameType,
		nMatchType:      nMatchType,
		strParam:        strParam,
		nOwnerID:        nOwnerID,
		nServerID:       nServerID,
		nClubID:         nClubID,
		nClubmgr:        nClubMgr,
		nAnswerProtocol: nAnswerProtocol,
	}
	self.mCreateTable[nNewRoomID] = tCreateNode

	tSend := packet.NewPacket(nil)
	tSend.SetMsgID(protomsg.Old_MSGID_CREATE_ROOM.UInt16())
	tSend.WriteUInt32(nOwnerID)
	tSend.WriteUInt32(nNewRoomID)
	tSend.WriteUInt8(uint8(nGameType))
	tSend.WriteString(strParam)
	tSend.WriteUInt8(uint8(nMatchType))
	tSend.WriteUInt32(nClubID)
	tSend.WriteUInt32(nClubMgr)
	send_tools.Send2Game(tSend.GetData(), tServerNode.SessionID)
	return 0
}

func (self *hallMgr) destroyRoom(tRoom *room) {

	self.updateDeskList(tRoom, 3)

	for key := range tRoom.mSeatID {
		tAccount := account.AccountMgr.GetAccountByID(key)
		if tAccount != nil && tAccount.RoomID == tRoom.nRoomID {
			tAccount.RoomID = 0
			tAccount.Index = 0
			tAccount.GameType = 0
			tAccount.MatchType = 0
			if tAccount.Robot > 0 {
				tAccount.IsUse = false
				tAccount.RMB = 0
			}
		}
	}
	for key := range tRoom.mWatchID {
		tAccount := account.AccountMgr.GetAccountByID(key)
		if tAccount != nil && tAccount.RoomID == tRoom.nRoomID {
			tAccount.RoomID = 0
			tAccount.Index = 0
			tAccount.GameType = 0
			tAccount.MatchType = 0
			if tAccount.Robot > 0 {
				tAccount.IsUse = false
				tAccount.RMB = 0
			}
		}
	}

	delete(self.mRoomTable, tRoom.nRoomID)

	if tRoom.clubID != 0 {
		club := ClubMgr.Clubs[tRoom.clubID]
		if club == nil {
			log.Warnf("销毁房间:%v 失败，找不到俱乐部:%v", tRoom.nRoomID, tRoom.clubID)
			return
		}
		ClubMgr.Clubs[tRoom.clubID].UnRelateRoom(tRoom.nRoomID)
	}
}

// 该列表里是允许开启桌子的游戏列表
var desk_list = map[uint8]bool{
	common.EGameTypeDGK.Value():                true,
	common.EGameTypeXMMJ.Value():               true,
	common.EGameTypePDK_HN.Value():             true,
	common.EGameTypeDING_ER_HONG.Value():       true,
	common.EGameTypeCHE_XUAN.Value():           true,
	common.EGameTypeTEN_NIU_NIU.Value():        true,
	common.EGameTypeSAN_GONG.Value():           true,
	common.EGameTypeWUHUA_NIUNIU.Value():       true,
	common.EGameTypeLONG_HU_DOU.Value():        true,
	common.EGameTypeFQZS.Value():               true,
	common.EGameTypeTUI_TONG_ZI.Value():        true,
	common.EGameTypeSHEN_SHOU_ZHI_ZHAN.Value(): true,
	common.EGameTypeHONG_BAO.Value():           true,
	//common.EGameTypeNIU_NIU.Value(): true,
	//common.EGameTypePAO_DE_KUAI.Value(): true,
}

func (self *hallMgr) clearDeskList(nGameType uint8, nAccountID uint32, nClubID uint32) {

	for nCheckGameType, _ := range desk_list {
		if self.mUpdateDesk[nClubID] == nil || self.mUpdateDesk[nClubID][nCheckGameType] == nil {
			continue
		}
		tCheckGame := self.mUpdateDesk[nClubID][nCheckGameType]
		for _, tCheckMap := range tCheckGame {
			if tCheckMap != nil {
				delete(tCheckMap, nAccountID)
			}
		}
	}
}

func (self *hallMgr) openDeskUpdate(nGameType uint8, nMatchType uint8, tAccount *account.Account, nClubID uint32) uint8 {

	if tAccount == nil {
		return 1
	}

	if nMatchType < 1 || nMatchType > 10 {
		return 2
	}

	if _, isExist := desk_list[nGameType]; isExist == false {
		return 3
	}

	switch nGameType {
	case common.EGameTypeDGK.Value(),
		common.EGameTypeXMMJ.Value(),
		common.EGameTypeCHE_XUAN.Value():
		if nClubID == 0 {
			MatchList.CleanMatchInfo(tAccount.AccountId, common.EGameType(nGameType), tAccount.SessionId)
		}
	}

	if self.mUpdateDesk[nClubID] == nil {
		self.mUpdateDesk[nClubID] = gameMap{}
	}
	if self.mUpdateDesk[nClubID][nGameType] == nil {
		// 客户端使用的是1~10的下标  所以服务器初始化11个
		self.mUpdateDesk[nClubID][nGameType] = make([]accMap, config.MAX_MATCH_TYPE, config.MAX_MATCH_TYPE)
	}
	if self.mUpdateDesk[nClubID][nGameType][nMatchType] == nil {
		self.mUpdateDesk[nClubID][nGameType][nMatchType] = make(accMap)
	}

	// 客户端一个页面会打开多个游戏类型列表, 故不能清理其他游戏类型; 进房间后再清理
	//self.clearDeskList(nGameType, tAccount.AccountId, nClubID)

	self.mUpdateDesk[nClubID][nGameType][nMatchType][tAccount.AccountId] = tAccount
	self.sendDeskList(nGameType, nMatchType, nClubID, tAccount.SessionId)

	//MatchList.JoinList(tAccount.AccountId, common.EGameType(nGameType)) // 玩家在大厅打开界面
	return 0
}

func (self *hallMgr) closeDeskUpdate(nGameType uint8, nMatchType uint8, tAccount *account.Account, nClubID uint32) uint8 {

	if tAccount == nil {
		return 1
	}

	if nMatchType < 1 || nMatchType > 10 {
		return 2
	}

	if _, isExist := desk_list[nGameType]; isExist == false {
		return 3
	}

	if self.mUpdateDesk[nClubID] == nil || self.mUpdateDesk[nClubID][nGameType] == nil || self.mUpdateDesk[nClubID][nGameType][nMatchType] == nil {
		return 4
	}

	delete(self.mUpdateDesk[nClubID][nGameType][nMatchType], tAccount.AccountId)

	// 进入匹配
	MatchList.LeaveList(tAccount.AccountId)
	return 0
}

func (self *hallMgr) buildAllDeskPlayerMsg(nGameType uint8, nClubID uint32) packet.IPacket {

	sGameMap := self.getGameMap(nGameType)
	if sGameMap == nil {
		return nil
	}

	mCount := make(map[uint8]uint16)
	for _, tNode := range sGameMap {
		if tNode.clubID == nClubID {
			if _, isExist := mCount[tNode.nMatchType]; isExist == false {
				mCount[tNode.nMatchType] = tNode.GetSeatCount()
			} else {
				mCount[tNode.nMatchType] += tNode.GetSeatCount()
			}
		}
	}

	pBuild := packet.NewPacket(nil)
	pBuild.SetMsgID(protomsg.MSGID_ALL_DESK_PLAYER_COUNT.UInt16())
	pBuild.WriteUInt8(nGameType)
	pBuild.WriteUInt16(uint16(len(mCount)))
	for nMatch, nCount := range mCount {
		pBuild.WriteUInt8(nMatch)
		pBuild.WriteUInt16(nCount)
	}
	return pBuild
}

func (self *hallMgr) sendDeskList(nGameType uint8, nMatchType uint8, nClubID uint32, nSessionID int64) {

	tSend := packet.NewPacket(nil)
	tSend.SetMsgID(protomsg.Old_MSGID_ROOM_DESK_LIST.UInt16())
	tSend.WriteUInt8(1) // 1所有房间列表, 2更新房间, 3销毁房间, 4新增房间
	tSend.WriteUInt8(nGameType)
	tSend.WriteUInt8(nMatchType)
	tSend.WriteUInt32(nClubID)

	mMatchMap := self.getMatchMap(nGameType, nMatchType)
	nRoomCount := uint16(0)
	nRoomWPos := tSend.GetWritePos()
	tSend.WriteUInt16(nRoomCount)
	for _, tRoom := range mMatchMap {
		if tRoom.clubID == nClubID {
			nRoomCount++
			tSend.WriteUInt32(tRoom.nRoomID)
			tSend.WriteString(tRoom.strParam)
			tSend.WriteUInt32(tRoom.nBet)
			tSend.WriteInt64(tRoom.iCreateTime)
			tRoom.FillRoomSeatInfo(tSend)
		}
	}
	tSend.Rrevise(nRoomWPos, nRoomCount)
	send_tools.Send2Account(tSend.GetData(), nSessionID)

	// 发送所有档次对应的玩家人数给客户端
	tAllPlayerCount := self.buildAllDeskPlayerMsg(nGameType, nClubID)
	send_tools.Send2Account(tAllPlayerCount.GetData(), nSessionID)
}

// nType: 2更新房间, 3销毁房间, 4新增房间
func (self *hallMgr) updateDeskList(tRoom *room, nUpdateType uint8) {
	nGameType := tRoom.nGameType
	switch nGameType {
	case common.EGameTypeDGK.Value():
	case common.EGameTypeXMMJ.Value():
	case common.EGameTypePDK_HN.Value():
	case common.EGameTypeDING_ER_HONG.Value():
	case common.EGameTypeCHE_XUAN.Value():
	case common.EGameTypeTEN_NIU_NIU.Value():
	case common.EGameTypeSAN_GONG.Value():
	case common.EGameTypeWUHUA_NIUNIU.Value():
	case common.EGameTypeLONG_HU_DOU.Value():
	case common.EGameTypeFQZS.Value():
	case common.EGameTypeHONG_BAO.Value():
	case common.EGameTypeTUI_TONG_ZI.Value():
	//case common.EGameTypeNIU_NIU.Value():
	//case common.EGameTypePAO_DE_KUAI.Value():
	default:
		// 其他类型不允许
		return
	}

	sUpdateNode := self.getUpdateDeskSlice(nGameType, tRoom.clubID)
	if sUpdateNode == nil {
		return
	}

	// 发送所有档次对应的玩家人数给客户端
	tAllPlayerCount := self.buildAllDeskPlayerMsg(nGameType, tRoom.clubID)
	for _, sUpdate := range sUpdateNode {
		for _, tAccount := range sUpdate {
			if tAccount.IsOnline() == true && tAccount.IsInClub(tRoom.clubID) == true {
				send_tools.Send2Account(tAllPlayerCount.GetData(), tAccount.SessionId)
			}
		}
	}

	sUpdateMatch := sUpdateNode[tRoom.nMatchType]
	if sUpdateMatch == nil || len(sUpdateMatch) <= 0 {
		return
	}

	tSend := packet.NewPacket(nil)
	tSend.SetMsgID(protomsg.Old_MSGID_ROOM_DESK_LIST.UInt16())
	if nUpdateType == 3 {
		tSend.WriteUInt8(nUpdateType)
		tSend.WriteUInt8(tRoom.nGameType)
		tSend.WriteUInt8(tRoom.nMatchType)
		tSend.WriteUInt32(tRoom.clubID)
		tSend.WriteUInt32(tRoom.nRoomID)
	} else {
		tSend.WriteUInt8(nUpdateType)
		tSend.WriteUInt8(tRoom.nGameType)
		tSend.WriteUInt8(tRoom.nMatchType)
		tSend.WriteUInt32(tRoom.clubID)
		tSend.WriteUInt16(1)
		tSend.WriteUInt32(tRoom.nRoomID)
		tSend.WriteString(tRoom.strParam)
		tSend.WriteUInt32(tRoom.nBet)
		tSend.WriteInt64(tRoom.iCreateTime)
		tRoom.FillRoomSeatInfo(tSend)
	}

	for _, tAccount := range sUpdateMatch {
		if tAccount.IsOnline() == true && tAccount.IsInClub(tRoom.clubID) == true {
			send_tools.Send2Account(tSend.GetData(), tAccount.SessionId)
		}
	}
}

func (self *hallMgr) mapServerNode(nHaveRoomCount uint32, nServerID uint16, nSessionID int64) {

	nGameType := server.ServerMgr.AddServerNode(nServerID, nSessionID)
	if nHaveRoomCount <= 0 {
		nDestoryRoomCount := 0
		for _, tRoom := range self.mRoomTable {
			if tRoom.nServerID == nServerID {
				self.destroyRoom(tRoom)
				log.Infof("Mapping DestroyRoom, ServerID:%v Game:%v RoomID:%v", nServerID, common.EGameType(tRoom.nGameType), tRoom.nRoomID)
				nDestoryRoomCount++
			}
		}
		self.preCreateRoom(nGameType, 0)
		ClubMgr.preCreateRoom(nGameType)
	}
}

func (self *hallMgr) UnMapServerNode(nSessionID int64) {
	tNode := server.ServerMgr.SetServerNodeUnavailable(nSessionID)
	if tNode != nil {
		// 维护标记开启; 重新建立连接时恢复关闭
		tNode.IsMaintenance = true

		var nDestoryRoomCount uint8

		// 启服时预先创建的房间不主动销毁
		switch tNode.GameType {
		case common.EGameTypeDGK.Value():
		case common.EGameTypeXMMJ.Value():
		case common.EGameTypePAO_DE_KUAI.Value():
		case common.EGameTypePDK_HN.Value():
		case common.EGameTypeDING_ER_HONG.Value():
		case common.EGameTypeHONG_BAO.Value():
		case common.EGameTypeCHE_XUAN.Value():
		case common.EGameTypeNIU_NIU.Value():
		case common.EGameTypeWUHUA_NIUNIU.Value():
		case common.EGameTypeTEN_NIU_NIU.Value():
		case common.EGameTypeSAN_GONG.Value():
		default:
			for _, tRoom := range self.mRoomTable {
				if tRoom.nServerID == tNode.ServerID {
					self.destroyRoom(tRoom)
					log.Infof("Unlink-1 DestroyRoom, ServerID:%v Game:%v RoomID:%v", tRoom.nServerID, common.EGameType(tRoom.nGameType), tRoom.nRoomID)
					nDestoryRoomCount++
				}
			}
		}
		log.Infof("Unlink-1 the GameServer mapping, nServerID:%v SessionID:%v, Destory:%v Room", tNode.ServerID, nSessionID, nDestoryRoomCount)
	}
}

// 判断俱乐部内指定游戏是否有房间
func (self *hallMgr) isHaveClubRoom(nGameType uint8, nClubID uint32) bool {
	mGameMap := self.getGameMap(nGameType)
	if mGameMap == nil {
		return false
	}
	for _, tRoom := range mGameMap {
		if tRoom.clubID == nClubID {
			return true
		}
	}
	return false
}

func (self *hallMgr) isWanRenChang(nGameType uint8) bool {
	switch nGameType {
	case common.EGameTypeSHEN_SHOU_ZHI_ZHAN.Value():
		return true
	case common.EGameTypeTUI_TONG_ZI.Value():
		return true
	case common.EGameTypeLONG_HU_DOU.Value():
		return true
	case common.EGameTypeHONG_HEI_DA_ZHAN.Value():
		return true
	case common.EGameTypeFQZS.Value():
		return true
	case common.EGameTypeHONG_BAO.Value():
		return true
	default:
		return false
	}
}

func (self *hallMgr) preCreateRoom(nGameType uint8, nClubID uint32) {

	// 麻将类房间创建; 字符串参数顺序一致才能使用
	fCreateRoomByMaJiang := func(nCreateIndex uint8, CREATE_ROOM_COUNT map[int]int, nCountIndex uint8) {
		nMatchType := uint8(0)
		strParam := config.GetAutoCreateRoomParam(nGameType, nCreateIndex)
		sParam := utils.SplitConf2ArrInt32(strParam, "|")
		if len(sParam) >= 4 {
			// 房间档次按照房间参数中人数来确定; 3人档次3  4人档次4
			nMatchType = uint8(sParam[3])
		} else {
			log.Warnf("预创建 %v 房间时, 读取参数错误; 参数:%v", nGameType, strParam)
		}

		// 每个配置创建多个房间
		nCreateCount := CREATE_ROOM_COUNT[int(nCountIndex)]
		for n := 0; n < nCreateCount; n++ {
			nRet := self.createRoom(nil, nGameType, nMatchType, strParam, protomsg.Old_MSGID_SYSTEM_CREATE_ROOM.UInt16(), nClubID)
			if nRet > 0 {
				log.Warnf("PerCreateRoom Error, nGameType:%v, nParamIndex:%v, ErrCode:%v", nGameType, nMatchType, nRet)
			}
		}

		log.Infof("预先创建游戏:%v, 房间档次:%v, 房间参数:%v, 创建数量:%v, 俱乐部ID:%v", common.EGameType(nGameType).String(), nMatchType, strParam, nCreateCount, nClubID)
	}

	// 除万人场和麻将类游戏; 通用的房间创建函数
	// 第一参数: 游戏类型,
	// 第二参数: 创建房间参数表; [档次] 创建几个房间
	// 第三参数: 1: 固定匹配档次模式  2: 解析字符串参数中的值为匹配档次 3: 参数下标为匹配档次
	// 第四参数: 对应第三参数的值
	// 第五参数: 俱乐部ID
	fCreateRoomByCommon := func(nGameType uint8, CREATE_ROOM_COUNT map[int]int, eModeType types.ECreateMode, nModeValue uint8, nClubID uint32) {

		isWanRenChang := self.isWanRenChang(nGameType)
		if isWanRenChang == true && nClubID > 0 {
			return
		}

		if CREATE_ROOM_COUNT == nil {
			log.Warnf("预创建 %v 房间时, 创建房间参数为空, CREATE_ROOM_COUNT", nGameType)
			return
		}

		nLen := config.GetAutoCreateRoomParamLen(nGameType)
		nMatchType := uint8(0)
		if eModeType == types.MODE_FIXED_MATCH_TYPE {
			// 固定匹配档次模式
			nMatchType = nModeValue
		}

		for nParamIndex := uint8(1); nParamIndex < nLen; nParamIndex++ {
			strParam := config.GetAutoCreateRoomParam(nGameType, nParamIndex)

			if eModeType == types.MODE_PARS_STRING_PARAM {
				// 固定匹配档次, 设置为字符串参数中解析出来的值
				sParam := utils.SplitConf2ArrInt32(strParam, "|")
				if len(sParam) >= int(nModeValue)+1 {
					// 房间档次按照房间参数中人数来确定; 2人档次2 3人档次3  4人档次4
					nMatchType = uint8(sParam[nModeValue])
				} else {
					log.Warnf("预创建 %v 房间时, 读取参数错误; 参数:%v", nGameType, strParam)
				}
			} else if eModeType == types.MODE_ACCORD_PARAM_INDEX {
				// 不固定匹配档次; 设置为参数下标
				nMatchType = nParamIndex
			}

			// 每个配置创建多个房间
			nCreateRoomCount := CREATE_ROOM_COUNT[int(nParamIndex)]
			for n := 0; n < int(nCreateRoomCount); n++ {
				nRet := self.createRoom(nil, nGameType, nMatchType, strParam, protomsg.Old_MSGID_SYSTEM_CREATE_ROOM.UInt16(), nClubID)
				if nRet > 0 {
					log.Warnf("PerCreateRoom Error, nGameType:%v, nParamIndex:%v, ErrCode:%v", nGameType, nMatchType, nRet)
				}
			}
			log.Infof("预先创建游戏:%v, 房间档次:%v, 房间参数:%v, 创建数量:%v, 俱乐部ID:%v", common.EGameType(nGameType).String(), nMatchType, strParam, nCreateRoomCount, nClubID)
		}
	}

	// 万人场在连接服务器时, 自动创建房间; 万人场只创建非俱乐部的房间
	isWanRenChang := self.isWanRenChang(nGameType)
	if isWanRenChang == true && nClubID == 0 {
		if nGameType != common.EGameTypeHONG_BAO.Value() {
			nMatchType := uint8(1)
			strParam := config.GetAutoCreateRoomParam(nGameType, nMatchType)
			nRet := self.createRoom(nil, nGameType, nMatchType, strParam, protomsg.Old_MSGID_SYSTEM_CREATE_ROOM.UInt16(), 0)
			if nRet > 0 {
				log.Warnf("PerCreateRoom Error, nGameType:%v, nParamIndex:%v, ErrCode:%v", nGameType, nMatchType, nRet)
			} else {
				log.Infof("预先创建游戏:%v, 房间档次:%v, 房间参数:%v, 创建数量:1, 俱乐部ID:%v", common.EGameType(nGameType).String(), nMatchType, strParam, nClubID)
			}
		} else {
			CREATE_ROOM_COUNT := config.GetPublicConfig_Mapi("HB_AUTO_CREATE_ROOM_COUNT")
			fCreateRoomByCommon(nGameType, CREATE_ROOM_COUNT, types.MODE_ACCORD_PARAM_INDEX, 0, 0)
		}
		return
	}

	switch nGameType {
	case common.EGameTypeDGK.Value(), common.EGameTypeXMMJ.Value():

		var CREATE_ROOM_COUNT_2 map[int]int
		var CREATE_ROOM_COUNT_3 map[int]int
		if nGameType == common.EGameTypeDGK.Value() {
			CREATE_ROOM_COUNT_2 = config.GetPublicConfig_Mapi("DGK_AUTO_CREATE_ROOM_COUNT_2")
			CREATE_ROOM_COUNT_3 = config.GetPublicConfig_Mapi("DGK_AUTO_CREATE_ROOM_COUNT_3")
		} else {
			CREATE_ROOM_COUNT_2 = config.GetPublicConfig_Mapi("PANDA_AUTO_CREATE_ROOM_COUNT_2")
			//CREATE_ROOM_COUNT_3 = config.GetPublicConfig_Mapi("PANDA_AUTO_CREATE_ROOM_COUNT_3")
		}
		if CREATE_ROOM_COUNT_2 == nil {
			log.Warnf("获取创建房间档次对应房间数量参数失败, 游戏:%v", common.ECardType(nGameType))
			return
		}

		// 创建两人房间
		CREATE_ROOM_2_LEN := uint8(len(CREATE_ROOM_COUNT_2))
		for i := uint8(1); i <= CREATE_ROOM_2_LEN; i++ {
			fCreateRoomByMaJiang(i, CREATE_ROOM_COUNT_2, i)
		}

		// 创建三人房间
		if CREATE_ROOM_COUNT_3 != nil {
			nLen := config.GetAutoCreateRoomParamLen(nGameType)
			for j := CREATE_ROOM_2_LEN + 1; j < nLen; j++ {
				fCreateRoomByMaJiang(j, CREATE_ROOM_COUNT_3, j-CREATE_ROOM_2_LEN)
			}
		}

	case common.EGameTypePDK_HN.Value():
		CREATE_ROOM_COUNT := config.GetPublicConfig_Mapi("PDK_HN_AUTO_CREATE_ROOM_COUNT")
		fCreateRoomByCommon(nGameType, CREATE_ROOM_COUNT, types.MODE_FIXED_MATCH_TYPE, config.PDK_HN_MAX_PLAYER, nClubID)

	//case common.EGameTypePAO_DE_KUAI.Value():
	//	CREATE_ROOM_COUNT := config.GetPublicConfig_Mapi("PDK_AUTO_CREATE_ROOM_COUNT")
	//	fCreateRoomByCommon(nGameType, CREATE_ROOM_COUNT, types.MODE_PARS_STRING_PARAM, 3, nClubID)

	case common.EGameTypeNIU_NIU.Value(), common.EGameTypeWUHUA_NIUNIU.Value():
		CREATE_ROOM_COUNT := config.GetPublicConfig_Mapi("NN_AUTO_CREATE_ROOM_COUNT")
		//fCreateRoomByCommon(common.EGameTypeNIU_NIU.Value(), CREATE_ROOM_COUNT, types.MODE_FIXED_MATCH_TYPE, config.NN_MAX_PLAYER, nClubID)
		fCreateRoomByCommon(common.EGameTypeWUHUA_NIUNIU.Value(), CREATE_ROOM_COUNT, types.MODE_FIXED_MATCH_TYPE, config.NN_MAX_PLAYER, nClubID)

	case common.EGameTypeTEN_NIU_NIU.Value():
		CREATE_ROOM_COUNT := config.GetPublicConfig_Mapi("TNN_AUTO_CREATE_ROOM_COUNT")
		fCreateRoomByCommon(nGameType, CREATE_ROOM_COUNT, types.MODE_FIXED_MATCH_TYPE, config.TNN_MAX_PLAYER, nClubID)

	case common.EGameTypeSAN_GONG.Value():
		CREATE_ROOM_COUNT := config.GetPublicConfig_Mapi("SG_AUTO_CREATE_ROOM_COUNT")
		fCreateRoomByCommon(nGameType, CREATE_ROOM_COUNT, types.MODE_FIXED_MATCH_TYPE, config.SG_MAX_PLAYER, nClubID)

	case common.EGameTypeSHI_SAN_SHUI.Value():
		CREATE_ROOM_COUNT := config.GetPublicConfig_Mapi("SSS_AUTO_CREATE_ROOM_COUNT")
		fCreateRoomByCommon(nGameType, CREATE_ROOM_COUNT, types.MODE_FIXED_MATCH_TYPE, 0, nClubID)

	case common.EGameTypeDING_ER_HONG.Value():
		CREATE_ROOM_COUNT := config.GetPublicConfig_Mapi("DEH_AUTO_CREATE_ROOM_COUNT")
		fCreateRoomByCommon(nGameType, CREATE_ROOM_COUNT, types.MODE_FIXED_MATCH_TYPE, config.DEH_MAX_PLAYER, nClubID)

	case common.EGameTypeCHE_XUAN.Value():
		CREATE_ROOM_COUNT := config.GetPublicConfig_Mapi("CX_AUTO_CREATE_ROOM_COUNT")
		fCreateRoomByCommon(nGameType, CREATE_ROOM_COUNT, types.MODE_FIXED_MATCH_TYPE, config.CX_MAX_PLAYER, nClubID)

	default:
		// 其他类型不允许
		return
	}
}

func (self *hallMgr) checkPreCreateRoomCount(nGameType, nMatchType uint8, tAccount *account.Account, nClubID uint32) {

	// 必须有创建者
	if tAccount == nil {
		return
	}

	var CHECK_CREATE_ROOM_COUNT map[int]int
	switch nGameType {
	case common.EGameTypeCHE_XUAN.Value():
		CHECK_CREATE_ROOM_COUNT = config.GetPublicConfig_Mapi("CX_AUTO_CHECK_CREATE_ROOM_COUNT")
	case common.EGameTypeDING_ER_HONG.Value():
		CHECK_CREATE_ROOM_COUNT = config.GetPublicConfig_Mapi("DEH_AUTO_CHECK_CREATE_ROOM_COUNT")
	//case common.EGameTypePAO_DE_KUAI.Value():
	//	CHECK_CREATE_ROOM_COUNT = config.GetPublicConfig_Mapi("PDK_AUTO_CHECK_CREATE_ROOM_COUNT")
	case common.EGameTypePDK_HN.Value():
		CHECK_CREATE_ROOM_COUNT = config.GetPublicConfig_Mapi("PDK_HN_AUTO_CHECK_CREATE_ROOM_COUNT")
	case common.EGameTypeSHI_SAN_SHUI.Value():
		CHECK_CREATE_ROOM_COUNT = config.GetPublicConfig_Mapi("SSS_AUTO_CHECK_CREATE_ROOM_COUNT")
	case common.EGameTypeNIU_NIU.Value():
		// 六人抢庄牛牛不创建房间
		return
	case common.EGameTypeWUHUA_NIUNIU.Value():
		CHECK_CREATE_ROOM_COUNT = config.GetPublicConfig_Mapi("NN_AUTO_CHECK_CREATE_ROOM_COUNT")
	case common.EGameTypeTEN_NIU_NIU.Value():
		CHECK_CREATE_ROOM_COUNT = config.GetPublicConfig_Mapi("TNN_AUTO_CHECK_CREATE_ROOM_COUNT")
	case common.EGameTypeSAN_GONG.Value():
		CHECK_CREATE_ROOM_COUNT = config.GetPublicConfig_Mapi("SG_AUTO_CHECK_CREATE_ROOM_COUNT")
	case common.EGameTypeDGK.Value():
		if nMatchType == 2 {
			CHECK_CREATE_ROOM_COUNT = config.GetPublicConfig_Mapi("DGK_AUTO_CHECK_CREATE_ROOM_COUNT_2")
		} else {
			CHECK_CREATE_ROOM_COUNT = config.GetPublicConfig_Mapi("DGK_AUTO_CHECK_CREATE_ROOM_COUNT_3")
		}
	case common.EGameTypeXMMJ.Value():
		if nMatchType == 2 {
			CHECK_CREATE_ROOM_COUNT = config.GetPublicConfig_Mapi("PANDA_AUTO_CHECK_CREATE_ROOM_COUNT_2")
		} else {
			// 3人桌子不允许自动创建房间
			return
		}
	default:
		// 其他类型不允许
		return
	}

	if CHECK_CREATE_ROOM_COUNT == nil {
		log.Warnf("获取检查创建房间档次对应房间数量参数失败, 游戏:%v", common.EGameType(nGameType))
		return
	}

	mGameMap := self.getMatchMap(nGameType, nMatchType)
	if mGameMap == nil {
		return
	}

	mAutoCreateMap := make(map[string]*auto_create_node)
	for _, tRoom := range mGameMap {
		if tRoom != nil && tRoom.isLock == false && tRoom.clubID == nClubID {
			if mAutoCreateMap[tRoom.strParam] == nil {
				mAutoCreateMap[tRoom.strParam] = &auto_create_node{nBet: tRoom.nBet, nMax: tRoom.nMax, nFreeSeat: 0, nMatchType: tRoom.nMatchType}
			}

			tNode := mAutoCreateMap[tRoom.strParam]
			tNode.nFreeSeat += tRoom.nMax - tRoom.GetSeatCount()
		}
	}

	for strParam, tNode := range mAutoCreateMap {
		if tNode.nFreeSeat <= 1 {
			// 每个配置创建多个房间: 根据底注找到对应创建房间数
			nCheckCreateCount := CHECK_CREATE_ROOM_COUNT[int(tNode.nBet)]
			for i := 0; i < nCheckCreateCount; i++ {
				nRet := self.createRoom(tAccount, nGameType, tNode.nMatchType, strParam, protomsg.Old_MSGID_SYSTEM_CREATE_ROOM.UInt16(), nClubID)
				if nRet > 0 {
					log.Infof("System Auto createRoom Error, nGameType:%v, nMatchType:%v, strParam:%v ErrCode:%v", common.EGameType(nGameType), nMatchType, strParam, nRet)
				}
			}
		}
	}
}

func (self *hallMgr) getNewRoomID() uint32 {
	var nNewRoomID uint32
	for true {
		nNewRoomID = uint32(utils.Randx_y(100000, 999999))
		tRoom := self.GetRoom(nNewRoomID)
		if tRoom == nil {
			return nNewRoomID
		}
	}
	return 0
}

func (self *hallMgr) getGameMap(nGameType uint8) roomMap {
	mGameMap := make(roomMap)
	for key, value := range self.mRoomTable {
		if value.nGameType == nGameType {
			mGameMap[key] = value
		}
	}
	return mGameMap
}

func (self *hallMgr) getMatchMap(nGameType uint8, nMatchType uint8) roomMap {
	mMatchMap := make(roomMap)
	for key, value := range self.mRoomTable {
		if value.nGameType == nGameType && value.nMatchType == nMatchType {
			mMatchMap[key] = value
		}
	}
	return mMatchMap
}

func (self *hallMgr) getUpdateDeskSlice(nGameType uint8, nClubID uint32) []accMap {
	if self.mUpdateDesk[nClubID] == nil || self.mUpdateDesk[nClubID][nGameType] == nil {
		return nil
	}
	sUpdateSlice := self.mUpdateDesk[nClubID][nGameType]
	return sUpdateSlice
}

// 万人场不需要分配游戏节点IP; 使用时随机选择一个IP即可
func (self *hallMgr) assignmentGameNodeIP(tNewRoom *room) {
	switch uint8(tNewRoom.nGameType) {
	case common.EGameTypeNIU_NIU.Value():
	case common.EGameTypeWUHUA_NIUNIU.Value():
	case common.EGameTypeTEN_NIU_NIU.Value():
	case common.EGameTypeSAN_GONG.Value():
	case common.EGameTypeJIN_HUA.Value():
	case common.EGameTypeSHI_SAN_SHUI.Value():
	case common.EGameTypeCHE_XUAN.Value():
	case common.EGameTypeDING_ER_HONG.Value():
	case common.EGameTypeDGK.Value():
	case common.EGameTypeXMMJ.Value():
	case common.EGameTypePAO_DE_KUAI.Value():
	case common.EGameTypePDK_HN.Value():
	default:
		return
	}

	if len(self.ipNodes) == 0 {
		return
	}

	//mNodeUse := make(map[uint16]uint16)
	//mGame := self.getGameMap(tNewRoom.nGameType)
	//for _, tRoom := range mGame {
	//	if tRoom.nGameNodeID > 0 {
	//		if _, isExist := mNodeUse[tRoom.nGameNodeID]; isExist == false {
	//			mNodeUse[tRoom.nGameNodeID] = 1
	//		} else {
	//			mNodeUse[tRoom.nGameNodeID]++
	//		}
	//	}
	//}
	//
	//sGameNode := make([]*game_node, 0, 10)
	//for nNodeID := range self.ipNodes {
	//	if nUseCount, isExist := mNodeUse[uint16(nNodeID)]; isExist == true {
	//		sGameNode = append(sGameNode, &game_node{nNodeID: uint16(nNodeID), nUseCount: nUseCount})
	//	} else {
	//		sGameNode = append(sGameNode, &game_node{nNodeID: uint16(nNodeID), nUseCount: 0})
	//	}
	//}
	//
	//// 未使用的排最前面, 已使用中使用数量少的排前面
	//sort.Slice(sGameNode, func(i, j int) bool {
	//	tOne := sGameNode[i]
	//	tTwo := sGameNode[j]
	//	if tOne.nUseCount < tTwo.nUseCount {
	//		return true
	//	} else if tOne.nUseCount == tTwo.nUseCount {
	//		if tOne.nNodeID < tTwo.nNodeID {
	//			return true
	//		}
	//	}
	//	return false
	//})
	//
	//// 游戏房间分配一个节点ID
	//tNewRoom.nGameNodeID = sGameNode[0].nNodeID
}

func (self *hallMgr) newRoom(tAccount *account.Account, nGameType uint8, nNewRoomID uint32, nServerID uint16, nMatchType uint8, strParam string, nAnswerProtocol uint16, clubID, clubMgr uint32) *room {

	var nOwnerID uint32
	if tAccount == nil {
		nOwnerID = 0
	} else {
		nOwnerID = tAccount.AccountId
	}

	sParam := utils.SplitConf2ArrInt32(strParam, "|")
	if sParam == nil || len(sParam) < 3 {
		log.Fatalf("newRoom失败, 参数异常, 游戏:%v, NewRoomID:%v, 创建者:%v, ServerID:%v, 匹配档次:%v, 参数:%v, 响应消息:%v", common.EGameType(nGameType), nNewRoomID, nOwnerID, nServerID, nMatchType, strParam, nAnswerProtocol)
		return nil
	}

	var nMax uint16
	nBet := uint32(sParam[0])
	nEnter := uint32(sParam[1])
	nLeave := uint32(sParam[2])
	isLock := false

	// 最大人数不采用配置方式; 最大人数不允许修改;
	switch uint8(nGameType) {
	case common.EGameTypeNIU_NIU.Value():
		return nil
	case common.EGameTypeWUHUA_NIUNIU.Value():
		nMax = config.NN_MAX_PLAYER
		isLock = (sParam[len(sParam)-1] == 1) // 必须是最后一个参数; 机器人要组装此参数
	case common.EGameTypeSAN_GONG.Value():
		nMax = config.SG_MAX_PLAYER
		isLock = (sParam[len(sParam)-1] == 1) // 必须是最后一个参数; 机器人要组装此参数
	case common.EGameTypeTEN_NIU_NIU.Value():
		nMax = config.TNN_MAX_PLAYER
		isLock = (sParam[len(sParam)-1] == 1) // 必须是最后一个参数; 机器人要组装此参数
		nMatchType = uint8(nMax)              // 根据房间最多人数确定档次10人档次10
	case common.EGameTypeJIN_HUA.Value():
		nMax = config.JH_MAX_PLAYER
	case common.EGameTypeSHI_SAN_SHUI.Value():
		nMax = config.JH_MAX_PLAYER
		isLock = (sParam[len(sParam)-1] == 1) // 必须是最后一个参数; 机器人要组装此参数
	case common.EGameTypeDING_ER_HONG.Value():
		nMax = config.DEH_MAX_PLAYER
	case common.EGameTypeCHE_XUAN.Value():
		nMax = config.CX_MAX_PLAYER
	case common.EGameTypeSHEN_SHOU_ZHI_ZHAN.Value():
		nMax = config.SSZZ_MAX_PLAYER
	case common.EGameTypeFQZS.Value():
		nMax = config.FQZS_MAX_PLAYER
	case common.EGameTypeTUI_TONG_ZI.Value():
		nMax = config.TTZ_MAX_PLAYER
	case common.EGameTypeLONG_HU_DOU.Value():
		nMax = config.LHD_MAX_PLAYER
	case common.EGameTypeHONG_BAO.Value():
		nMax = config.HB_MAX_PLAYER
	case common.EGameTypeHONG_HEI_DA_ZHAN.Value():
		nMax = config.HHDZ_MAX_PLAYER
	case common.EGameTypeWU_ZI_QI.Value():
		nMax = config.WZQ_MAX_PLAYER
	case common.EGameTypeDGK.Value(), common.EGameTypeXMMJ.Value():
		isLock = (sParam[4] == 1)
		nMax = uint16(sParam[3]) // 房间最多人数由配置确定
		nMatchType = uint8(nMax) // 根据房间最多人数确定档次2人档次2, 3人档次3
		if nMax != 2 && nMax != 3 {
			if nGameType == common.EGameTypeDGK.Value() {
				log.Errorf("创建断勾卡房间异常, 房间人数只允许2人和3人, 当前设置人数:%v", nMax)
			} else {
				log.Errorf("创建熊猫麻将房间异常, 房间人数只允许2人和3人, 当前设置人数:%v", nMax)
			}
			return nil
		}
	case common.EGameTypePDK_HN.Value():
		nMax = config.PDK_HN_MAX_PLAYER
		isLock = (sParam[len(sParam)-1] == 1) // 必须是最后一个参数; 机器人要组装此参数
	case common.EGameTypePAO_DE_KUAI.Value():
		return nil
		//// 跑得快参数: 1底注 2入场 3离场 4人数(3人) 5炸弹数量(1炸,3炸) 6加锁
		//isLock = (sParam[5] == 1)
		//nMax = uint16(sParam[3]) // 房间最多人数由配置确定
		//nMatchType = uint8(nMax) // 根据房间最多人数确定档次3人档次3 4人4档次
		//if nMax != 3 && nMax != 4 {
		//	log.Errorf("创建跑得快房间异常, 房间人数只允许3人和4人, 当前设置人数:%v", nMax)
		//	return nil
		//}
	default:
		log.Errorf("错误的游戏类型:%v", nGameType)
		return nil
	}

	tNewRoom := &room{
		nRoomID:     nNewRoomID,
		nGameType:   nGameType,
		nMax:        nMax,
		nServerID:   nServerID,
		nMatchType:  nMatchType,
		iCreateTime: utils.MilliSecondTimeSince1970(),
		strParam:    strParam,
		isLock:      isLock,
		nOwnerID:    nOwnerID,
		nBet:        nBet,
		nEnter:      nEnter,
		nLeave:      nLeave,
		mSeatID:     make(map[uint32]bool),
		mWatchID:    make(map[uint32]bool),
		clubID:      clubID,
		nClubmgr:    clubMgr,
	}

	self.mRoomTable[nNewRoomID] = tNewRoom
	if clubID != 0 {
		club := ClubMgr.Clubs[clubID]
		if club == nil {
			log.Warnf("俱乐部桌子创建失败，找不到俱乐部:%v ", clubID)
			return nil
		}
		ClubMgr.Clubs[clubID].RelateRoom(tNewRoom)
	}
	return tNewRoom
}

func (self *hallMgr) canCreateRoom(tAccount *account.Account, nGameType uint8, nMatchType uint8, nClubID uint32, strParam string, nAnswerProtocol uint16) uint8 {
	switch nGameType {
	case common.EGameTypeNIU_NIU.Value():
		return 29
	case common.EGameTypePAO_DE_KUAI.Value():
		return 29
	case common.EGameTypeWUHUA_NIUNIU.Value():
	case common.EGameTypeTEN_NIU_NIU.Value():
	case common.EGameTypeSAN_GONG.Value():
	case common.EGameTypeJIN_HUA.Value():
	case common.EGameTypeSHI_SAN_SHUI.Value():
	case common.EGameTypeDGK.Value():
	case common.EGameTypeXMMJ.Value():
	case common.EGameTypePDK_HN.Value():
		// 创建需要判断元宝是否足够

	case common.EGameTypeSHEN_SHOU_ZHI_ZHAN.Value():
		return 29
	case common.EGameTypeFQZS.Value():
		return 29
	case common.EGameTypeTUI_TONG_ZI.Value():
		return 29
	case common.EGameTypeLONG_HU_DOU.Value():
		return 29
	case common.EGameTypeHONG_BAO.Value():
		return 29
	case common.EGameTypeHONG_HEI_DA_ZHAN.Value():
		return 29
	}

	if tAccount == nil {
		log.Errorf("判断能否创建房间时错误, 帐号对象为空; Game:%v MatchType:%v ClubID:%v strParam:%v", common.EGameType(nGameType), nMatchType, nClubID, strParam)
		return 0
	}

	isClubManager := false
	if nClubID > 0 && tAccount != nil {
		club := ClubMgr.Clubs[nClubID]
		if club != nil {
			nManageType := club.GetManageType(tAccount.AccountId)
			if nManageType > 0 {
				isClubManager = true
			}
		}
	}
	if isClubManager == true && nAnswerProtocol != protomsg.Old_MSGID_RMB_MATCH.UInt16() {
		// 俱乐部管理者不限制创建房间个数和金币要求
		return 0
	}

	if nMatchType == 0 ||
		nGameType == common.EGameTypeNIU_NIU.Value() ||
		nGameType == common.EGameTypeWUHUA_NIUNIU.Value() ||
		nGameType == common.EGameTypeTEN_NIU_NIU.Value() ||
		nGameType == common.EGameTypeSAN_GONG.Value() ||
		nGameType == common.EGameTypeDGK.Value() ||
		nGameType == common.EGameTypePAO_DE_KUAI.Value() ||
		nGameType == common.EGameTypePDK_HN.Value() ||
		nGameType == common.EGameTypeSHI_SAN_SHUI.Value() ||
		nGameType == common.EGameTypeXMMJ.Value() {
		mGameMap := self.getGameMap(nGameType)
		nHave := 0
		for _, tRoom := range mGameMap {
			if tRoom != nil && tRoom.nMatchType == nMatchType && tRoom.nOwnerID == tAccount.AccountId && tRoom.clubID == nClubID {
				nHave++
				if nHave >= 2 {
					// 同一类型只允许创建2个房间
					return 29
				}
			}
		}
	}

	sParam := utils.SplitConf2ArrInt32(strParam, "|")
	var nBet, nEnter, nLeave uint32
	nBet = uint32(sParam[0])   // 底注
	nEnter = uint32(sParam[1]) // 入场
	nLeave = uint32(sParam[2]) // 离场

	if nBet <= 0 || nEnter < nLeave || nBet >= nEnter {
		return 25
	}

	nHaveMoney := tAccount.GetMoney()
	if nHaveMoney < uint64(nBet) || nHaveMoney < uint64(nEnter) {
		return 25
	}
	return 0
}

func (self *hallMgr) canEnterRoom(tAccount *account.Account, tRoom *room) uint8 {
	nGameType := tRoom.nGameType
	if nGameType != common.EGameTypeJIN_HUA.Value() {
		return 0
	}

	if tAccount.RoomID > 0 && tAccount.RoomID != tRoom.nRoomID {
		return 20
	}

	if tRoom.nEnter < tRoom.nLeave || tRoom.nBet >= tRoom.nEnter {
		return 25
	}

	nHaveMoney := tAccount.GetMoney()
	if nHaveMoney < uint64(tRoom.nBet) || nHaveMoney < uint64(tRoom.nEnter) {
		return 25
	}

	if tRoom.clubID > 0 {
		isInClub := tAccount.IsInClub(tRoom.clubID)
		if isInClub == false {
			return 27
		}
	}
	return 0
}

// 获取指定房间所在游戏进程的SessionID
func (self *hallMgr) getGameSessionID(nRoomID uint32) int64 {
	tRoom := self.GetRoom(nRoomID)
	if tRoom != nil {
		tServerNode := server.ServerMgr.GetServerNode(tRoom.nServerID)
		if tServerNode != nil {
			return tServerNode.SessionID
		}
	}
	return 0
}

func (self *hallMgr) setMaintenanceTime(nTime uint32) {
	if self.nMaintenanceTime == 0 {
		self.nMaintenanceTime = nTime
		log.Infof("MaintenanceNotice, Time:%v", nTime)

		tSend := packet.NewPacket(nil)
		tSend.SetMsgID(protomsg.Old_MSGID_MAINTENANCE_NOTICE.UInt16())
		tSend.WriteUInt32(nTime)
		for _, tGame := range server.ServerMgr.GetAllServerList() {
			send_tools.Send2Game(tSend.GetData(), tGame.SessionID)
		}
	} else {
		log.Warnf("Replace Set MaintenanceNotice, Time:%v", self.nMaintenanceTime)
	}
}

func (self *hallMgr) OnEvent(e core.Event, t core.EventType) {
	switch t {
	case event.EventType_UpdateCharge:
		tWrapEv := e.(core.WrapEvent)
		tUpdateCharge := tWrapEv.Event.(event.UpdateCharge)

		tAccount := account.AccountMgr.GetAccountByID(tUpdateCharge.AccountID)
		if tAccount != nil {
			if tAccount.RoomID > 0 && (tAccount.GameType == uint32(common.EGameTypeDING_ER_HONG) || tAccount.GameType == uint32(common.EGameTypeCHE_XUAN)) {
				nGameSessionID := self.getGameSessionID(tAccount.RoomID)
				if nGameSessionID > 0 {
					tSend := packet.NewPacket(nil)
					tSend.SetMsgID(protomsg.Old_MSGID_UPDATE_CHARGE.UInt16())
					tSend.WriteUInt32(tAccount.AccountId)
					tSend.WriteInt64(tUpdateCharge.RMB)
					send_tools.Send2Game(tSend.GetData(), nGameSessionID)
				}
			}
		}
	default:

	}
}
