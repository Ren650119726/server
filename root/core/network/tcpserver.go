package network

import (
	"fmt"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"strconv"
)

/*
 * 针对客户端连接的server, 负责逻辑和client之间的消息中转，不处理任何消息
 */
type NetworkServer struct {
	acceptor        *Acceptor
	customer        []*core.Actor
	owner           *core.Actor
	hashring        *utils.HashRing
	listen_TCPaddr  string // 监听的tcp
	listen_HTTPaddr string // 监听的http
}

// 创建一个TCPServer
func NewNetworkServer(customer []*core.Actor, laddr, haddr string) *NetworkServer {
	server := &NetworkServer{}
	server.customer = append(server.customer, customer...)
	server.acceptor = NewAcceptor(server)
	server.listen_TCPaddr = laddr
	server.listen_HTTPaddr = haddr

	// 构建400个虚拟节点
	server.hashring = utils.NewHashRing(400)
	for i := 0; i < len(server.customer); i++ {
		key := fmt.Sprintf("%d", server.customer[i].GetID())
		server.hashring.AddNode(key, 50)
	}

	if server.acceptor == nil {
		log.Errorf("accept 为 nil ")
		return nil
	}
	return server
}

// actor初始化(actor接口定义)
func (self *NetworkServer) Init(owner *core.Actor) bool {
	if err := self.acceptor.Start(self.listen_TCPaddr, self.listen_HTTPaddr); err != nil {
		panic(err)
	}
	self.owner = owner
	// 启动定时器(执行update逻辑)
	self.owner.AddTimer(1, -1, self.update)

	log.Infof(colorized.Green("actor:[%v]  listenTCP:[%v] listenHTTP:[%v]"), self.owner.Id, self.listen_TCPaddr, self.listen_HTTPaddr)
	return true
}

// 资源清理
func (self *NetworkServer) Stop() {
	self.acceptor.Stop()
}

// actor消息处理 网络actor收到逻辑actor发送的消息
func (self *NetworkServer) HandleMessage(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case utils.ID_KICK_CLIENT: //踢玩家下线，断连接
		self.acceptor.Kick(session)
	case utils.ID_HEARTBEAT:
		//网络actor收到逻辑actor转发的client心跳
		self.handleHeartbeat(msg, session)
	default:
		// 这里的消息统统是需要发送出去
		// 这里可以做加密工作
		self.acceptor.Send(session, msg)
	}
	return true
}

/* 网络层回调接口 */
func (self *NetworkServer) handle_input(session int64, data []byte) {
	pack := packet.NewPacket(data)

	switch pack.GetMsgID() {
	case utils.ID_HEARTBEAT:
		core.CoreSend(self.owner.GetID(), self.owner.GetID(), data, session)

	default:
		switch pack.GetMsgID() {
		case utils.ID_DISCONNECT: // 有连接断开
			//log.Debugf("连接断开：%v", pack.ReadString())
		}

		// 通过一致性hash算法进行分配
		key := fmt.Sprintf("%d", session)
		node := self.hashring.GetNode(key)
		customer, err := strconv.Atoi(node)
		if err != nil {
			log.Error(err, session, pack.GetMsgID())
			return
		}

		core.CoreSend(self.owner.GetID(), int32(customer), data, session)
	}
}

/* 网络层回调接口 */
func (self *NetworkServer) update(dt int64) {
	self.acceptor.Update()
}

func (self *NetworkServer) doHeartbeat(dt int64) {
	self.acceptor.timingHandleHeartbeatTimeout()
}

func (self *NetworkServer) handleHeartbeat(buf []byte, session int64) {
	self.acceptor.updateHeartbeat(buf, session)
}

/* 网络层回调接口 */
func (self *NetworkServer) GetSessionIP(sesseionId int64) string {
	self.acceptor.sessions.RLock()
	defer self.acceptor.sessions.RUnlock()

	sess := self.acceptor.sessions.m[sesseionId]
	if sess != nil {
		return sess.RemoteIP()
	}
	return "err ip"
}
