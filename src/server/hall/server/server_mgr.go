package server

import (
	"root/common"
	"root/common/config"
	"root/core/log"
	"fmt"
	"github.com/astaxie/beego"
)

var ServerMgr = newServerMgr()

type (
	ServerNode struct {
		GameType      uint8  // 游戏类型
		ServerID      uint16 // 游戏服务器ID
		SessionID     int64  // 游戏服务器的通信ID
		Unavailable   bool   // 游戏节点是否可使用, 由于增加游戏到大厅通信的安全效验, 故不能该节点; 只能标记为不可用;
		CanClose      bool   // 游戏是否可关闭, 停服维护时标记是否所有游戏都关闭了;
		IsMaintenance bool   // 游戏节点是否维护标记; true维护中
	}

	serverMgr struct {
		mServerMap map[uint16]*ServerNode // gameType:sessionId
	}
)

func newServerMgr() *serverMgr {
	return &serverMgr{
		mServerMap: make(map[uint16]*ServerNode),
	}
}

// 判断指定SessionID是否是游戏进程的SessionID
func (self *serverMgr) IsGameServerSession(nSessionID int64) bool {
	GAME_TO_HALL_CHECK_SESSION := config.GetPublicConfig_Int64("GAME_TO_HALL_CHECK_SESSION")
	if GAME_TO_HALL_CHECK_SESSION != 1 {
		return true
	}

	for _, tNode := range self.mServerMap {
		if tNode.SessionID == nSessionID {
			return true
		}
	}
	return false
}

func (self *serverMgr) GetByGameType(nSessionID int64) uint8 {
	for _, tNode := range self.mServerMap {
		if tNode.SessionID == nSessionID {
			return tNode.GameType
		}
	}
	return 0
}

func (self *serverMgr) GetBySessionID(nGameType uint8) int64 {
	for _, tNode := range self.mServerMap {
		if tNode.GameType == nGameType {
			return tNode.SessionID
		}
	}
	return 0
}

func (self *serverMgr) AddServerNode(nServerID uint16, nSessionID int64) uint8 {

	nGameType := uint8(beego.AppConfig.DefaultInt(fmt.Sprintf("%v", nServerID)+"::gametype", 0))
	if nGameType == 0 {
		log.Errorf("找不到对应的 nGameType 配置 nServerID:%v", nServerID)
		return 0
	}

	var nOldSessionID int64
	node := self.mServerMap[nServerID]
	if node != nil {
		nOldSessionID = node.SessionID
		node.GameType = nGameType
		node.ServerID = nServerID
		node.SessionID = nSessionID
		node.CanClose = false
		node.Unavailable = true
		node.IsMaintenance = false

		log.Infof("======> 与游戏 %v 服务器ID:%v 重新建立连接, 旧SessionID:%v 新SessionID:%v", common.EGameType(nGameType).String(), nServerID, nOldSessionID, nSessionID)
	} else {
		node = &ServerNode{
			GameType:      nGameType,
			ServerID:      nServerID,
			SessionID:     nSessionID,
			CanClose:      false,
			Unavailable:   true,
			IsMaintenance: false,
		}
		self.mServerMap[nServerID] = node
		log.Infof("======> 与游戏 %v 服务器ID:%v 建立连接, 关联SessionID:%v", common.EGameType(nGameType).String(), nServerID, nSessionID)
	}
	return node.GameType
}

// 删除指定SessionID的服务器节点, 并返回该节点的ServerID
func (self *serverMgr) SetServerNodeUnavailable(nSessionID int64) *ServerNode {
	for key, value := range self.mServerMap {
		if value.SessionID == nSessionID {
			value.Unavailable = false
			tNode := value
			delete(self.mServerMap, key) // 由于增加游戏到大厅通信的安全效验, 故不能删除该节点; 只能标记为不可用
			return tNode
		}
	}
	return nil
}

func (self *serverMgr) GetServerNode(nServerID uint16) *ServerNode {
	tNode, isExist := self.mServerMap[nServerID]
	if isExist == true {
		if tNode.Unavailable == false {
			return nil
		}
		return tNode
	} else {
		return nil
	}
}

func (self *serverMgr) GetServerList(nGameType uint8) []*ServerNode {
	sServerList := []*ServerNode{}
	for _, value := range self.mServerMap {
		if value.GameType == nGameType {
			sServerList = append(sServerList, value)
		}
	}
	return sServerList
}

func (self *serverMgr) GetAllServerList() map[uint16]*ServerNode {
	return self.mServerMap
}
