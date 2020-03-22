package network

import (
	"net"
	"root/core/log"
)

/* 连接器 */
type Connector struct {
	offchan     chan int64    // 断线的channel
	session     *Session      // 在线链接
	remoteFun   func() string // 连接地址
	f_connected func()        // 连接上了回调
	isconnect   bool          // 连上服务器
}

/* 新建一个连接器 */
func NewConnector(callback NetCallBackIF) *Connector {
	connctor := &Connector{}
	connctor.offchan = make(chan int64, 1)
	connctor.session = NewClientSession(0, connctor.offchan, callback, 30, 30)
	connctor.isconnect = false
	return connctor
}

/* 连接服务器 */
func (self *Connector) Start(remote func() string, f_connected func()) {
	self.f_connected = f_connected
	self.remoteFun = remote
	remoteInfo := self.remoteFun()
	self.connect(remoteInfo)

}

func (self *Connector) Stop() {
	self.session.Kick()
}

// 连接远端
func (self *Connector) connect(remote string) {
	addr, err := net.ResolveTCPAddr("tcp", remote)
	if err != nil {
		log.Error("connect error:%v ", err.Error())
		return
	}

	// 执行逻辑
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Warnf("远端连接失败:%v", err.Error())
		return
	}
	self.isconnect = true

	// 设置缓冲区大小
	conn.SetReadBuffer(SOCKET_CACHE_SIZE)
	conn.SetWriteBuffer(SOCKET_CACHE_SIZE)
	self.f_connected()
	// 连接上了

	self.session.SetConn(conn)
	self.session.DoWork()
}

/* 检查session */
func (self *Connector) Update() {
	select {
	case <-self.offchan:
		self.isconnect = false
	default:
		if !self.isconnect {
			remoteInfo := self.remoteFun()
			self.connect(remoteInfo)
		}
		break
	}
}

/* 发送数据 */
func (self *Connector) Send(data []byte) bool {
	if !self.isconnect {
		return false
	}
	return self.session.SyncSend(data)
}

/* 关闭连接 */
func (self *Connector) Kick() {
	if !self.isconnect {
		return
	}
	self.session.Kick()
}
