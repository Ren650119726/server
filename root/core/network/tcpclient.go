package network

import (
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/log/tags"
	"root/core/packet"
	"root/core/utils"
	"time"
)

/*
 * 负责与center连接 收发消息
 */
type TCPClient struct {
	conn        *Connector
	parent      *core.Actor
	owner       *core.Actor
	remoteFun   func() string // 不直接传参的主要考虑因素是希望短线重连的时候可以每次都读取配置,热更新
	f_connected func()
	heart_timer int64
	heartbeat   bool
	heart_close bool
}

// 新创建一个center client 连接成功后回调onCallback
func NewTCPClient(parent *core.Actor, f_remote func() string, f_connected func()) *TCPClient {
	cli := &TCPClient{f_connected: f_connected, heart_close: false, heart_timer: 0}
	cli.parent = parent
	cli.conn = NewConnector(cli)
	cli.remoteFun = f_remote
	if cli.conn == nil {
		return nil
	}
	return cli
}

// 初始化
func (self *TCPClient) Init(actor *core.Actor) bool {
	self.owner = actor
	self.conn.Start(self.remoteFun, func() {
		log.Infof(colorized.Yellow("连接:[%v] 成功"), self.remoteFun())
		self.heartbeat = true
		core.LocalCoreSend(self.owner.Id, self.parent.Id, self.f_connected)
		if !tags.DEBUG {
			if self.heart_timer == 0 {
				self.heart_timer = self.owner.AddTimer(SEND_HEARTBEAT*1000, -1, self.sendHeartbeat)
			}
		}
	})

	self.owner.AddTimer(int64(time.Microsecond*1), -1, func(dt int64) {
		if self.owner.IsSuspend {
			return
		}
		self.Update(0)
	})

	return true
}

func (self *TCPClient) Stop() {
	self.conn.Stop()
}

// actor消息处理
func (self *TCPClient) HandleMessage(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case utils.ID_HEART_CLOSE:
		self.heart_close = true
	default:
		// 这里的消息统统是需要发送出去
		// 这里可以做加密工作
		self.conn.Send(msg)
	}

	return true
}

/* 网络层回调接口 */
func (self *TCPClient) handle_input(session int64, data []byte) {
	pack := packet.NewPacket(data)
	switch pack.GetMsgID() {
	case utils.ID_HEART_CLOSE:
		self.heart_close = true
	case utils.ID_HEARTBEAT:
		self.handleHeartbeat()
	case utils.ID_DISCONNECT: // server主动断开连接
		self.conn.Kick()
	default:
		/* 将消息交给逻辑层(自身不处理消息)
		 * 这里可以做解密工作(如果需要的话)
		 */
		core.CoreSend(self.owner.GetID(), self.parent.GetID(), data, session)
	}

}

/* 网络层回调接口 */
func (self *TCPClient) Update(dt int64) {
	self.conn.Update()
}

// 激活心跳
func (self *TCPClient) handleHeartbeat() {
	self.heartbeat = true
}

// 发送心跳包
func (self *TCPClient) sendHeartbeat(dt int64) {
	if !self.conn.isconnect || self.heart_close {
		return
	}

	if !self.heartbeat {
		// 心跳超时
		log.Warnf("心跳超时 ")
		self.conn.Kick()
		return
	}

	self.heartbeat = false
	pack := packet.NewPacket(nil)
	pack.SetMsgID(utils.ID_HEARTBEAT)
	self.conn.Send(pack.GetData())
}

func (self *TCPClient) Remote() string {
	return self.conn.remoteFun()
}
