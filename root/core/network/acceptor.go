package network

import (
	"golang.org/x/net/websocket"
	"net"
	"net/http"
	"root/core"
	"root/core/log"
	"root/core/packet"
	"root/core/utils"
	"sync"
)

/* TCP 接受器 */
type Acceptor struct {
	listener *net.TCPListener // 监听socket
	httpser  *http.Server
	offchan  chan int64       // 断线的channel
	newchan  chan net.Conn    // 新的链接channel
	sessions struct {
		sync.RWMutex
		m map[int64]*Session
	}
	//sessions            map[int64]*Session // 在线链接
	callback            NetCallBackIF // 网络上层回调
	curr_sessionid      int64         // 当前的sessionID
	last_handle_session int64         //上次处理心跳超时的session时间戳
}

// 新建acceptor
func NewAcceptor(callback NetCallBackIF) *Acceptor {
	acceptor := &Acceptor{}
	acceptor.offchan = make(chan int64, MAX_SESSION)
	acceptor.newchan = make(chan net.Conn, MAX_SESSION)
	acceptor.sessions = struct {
		sync.RWMutex
		m map[int64]*Session
	}{m: make(map[int64]*Session)}
	acceptor.callback = callback
	acceptor.curr_sessionid = int64(0)
	acceptor.last_handle_session = utils.SecondTimeSince1970()
	return acceptor
}

// 启动
func (self *Acceptor) Start(addr,httpAddr string) error {
	// 解析IP地址
	ip_addr, err := net.ResolveTCPAddr("tcp", addr)
	if nil != err {
		return err
	}

	// 监听
	self.listener, err = net.ListenTCP("tcp", ip_addr)
	if err != nil {
		return err
	}

	go self.DoListen()
	if httpAddr != ""{
		go self.DoListenHttp(httpAddr)
	}

	return nil
}

func (self *Acceptor) Stop() {
	self.sessions.RLock()
	defer self.sessions.RUnlock()

	self.listener.Close()
	if self.httpser != nil {
		self.httpser.Shutdown(nil)
	}


	for _, sess := range self.sessions.m {
		sess.Kick()
	}
}

// 监听go程
func (self *Acceptor) DoListen() {
	core.Gwg.Add(1)
	defer func() {
		core.Gwg.Done()
	}()

	for {
		conn, err := self.listener.AcceptTCP()
		if err != nil {
			return
		}
		// 产生一个新的链接
		// 设置缓冲区大小
		conn.SetReadBuffer(SOCKET_CACHE_SIZE)
		conn.SetWriteBuffer(SOCKET_CACHE_SIZE)
		self.newchan <- conn
	}
}

func (self *Acceptor) DoListenHttp(httpAddr string) {
	self.httpser = &http.Server{Addr:httpAddr}

	http.Handle("/connect", websocket.Handler(func(ws *websocket.Conn) {
		self.sessions.Lock()
		self.curr_sessionid++
		ws.PayloadType = websocket.BinaryFrame // 此行解决前端收到报错:Could not decode a text frame as UTF-8
		session := NewSession(self.curr_sessionid, ws, self.offchan, self.callback, 30, 30)
		//log.Infof("new websocket connect:%v", ws.LocalAddr())
		if session != nil {
			self.sessions.m[self.curr_sessionid] = session
			session.DoWork()
		}
		self.sessions.Unlock()
		select{
		case <-session.httpchan:
			//log.Infof("断开连接：%v",ws.RemoteAddr())
			return
		}
	}))
	log.Infof("监听websocket:%v",httpAddr)
	if err := self.httpser.ListenAndServe(); err != nil {
		log.Infof("http监听失败: err:%v ",err.Error())
	}
}

/* 检查session */
func (self *Acceptor) Update() {

	// 处理关闭的连接
	for {
		if len(self.offchan) <= 0 {
			break
		}
		sessionid := <-self.offchan

		self.sessions.Lock()
		if ses, ok := self.sessions.m[sessionid]; ok {
			pack := packet.NewPacket(nil)
			pack.WriteString(ses.RemoteIP())
			pack.SetMsgID(utils.ID_DISCONNECT)
			self.callback.handle_input(sessionid, pack.GetData())
			delete(self.sessions.m, sessionid)
		}
		self.sessions.Unlock()
	}

	// 处理新连接
	for {
		if len(self.newchan) <= 0 {
			break
		}

		conn := <-self.newchan
		self.sessions.Lock()
		if len(self.sessions.m) >= MAX_SESSION {
			// 超过最大连接数，就直接不让连接进来
			if err := conn.Close(); err != nil {
				log.Warnf("超过最大连接数:%v 关闭连接失败:%v", len(self.sessions.m), err.Error())
			}
		} else {
			self.curr_sessionid++
			session := NewSession(self.curr_sessionid, conn, self.offchan, self.callback, 30, 30)
			if session != nil {
				self.sessions.m[self.curr_sessionid] = session
				session.DoWork()
			}
		}

		self.sessions.Unlock()
	}

}

/* 发送数据 */
func (self *Acceptor) Send(sessionid int64, data []byte) {
	self.sessions.RLock()
	defer self.sessions.RUnlock()
	if session, ok := self.sessions.m[sessionid]; ok {
		session.SyncSend(data)
	}

}

/* 剔除 */
func (self *Acceptor) Kick(sessionid int64) {
	self.sessions.RLock()
	defer self.sessions.RUnlock()
	if session, ok := self.sessions.m[sessionid]; ok {
		session.Kick()
	}
}

//定时处理心跳包超时的session
func (self *Acceptor) timingHandleHeartbeatTimeout() {
	self.sessions.RLock()
	defer self.sessions.RUnlock()

	now := utils.SecondTimeSince1970()
	self.last_handle_session = now
	for _, v := range self.sessions.m {
		if now-v.heartbeat > HEARTBEAT_TIMEOUT {
			// todo
		}
	}
}

func (self *Acceptor) updateHeartbeat(buf []byte, sessid int64) {
	self.sessions.RLock()
	defer self.sessions.RUnlock()

	sess, ok := self.sessions.m[sessid]
	if ok {
		sess.heartbeat = utils.SecondTimeSince1970() // 服务器更新session心跳时间
		heartbeat := packet.NewPacket(nil)
		heartbeat.SetMsgID(utils.ID_HEARTBEAT)
		sess.send(heartbeat.GetData())
	} else {
		log.Warnf("心跳 找不到session：%v", sessid)
	}
}
