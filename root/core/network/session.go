package network

import (
	"bytes"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"crypto/rc4"
	"net"
	"syscall"
	"time"
)

const (
	SESS_KEYEXCG = 0x1 // 是否已经交换完毕KEY
	SESS_ENCRYPT = 0x2 // 是否可以开始加密
)

/* 外部几口定义 */
type NetConnIF interface {
	SetReadDeadline(t time.Time) error
	Close() error
	SetWriteDeadline(t time.Time) error
	Write(b []byte) (n int, err error)
	RemoteAddr() net.Addr
	Read(p []byte) (n int, err error)
}

//type Packet struct {
//	msgid uint16
//	data  []byte
//}

/* 网络session */
type Session struct {
	id        int64       // SessionID
	conn      NetConnIF   // *net.TCPConn
	sendchan  chan []byte // 发送缓冲区(包含协议头)
	offchan   chan int64  // 离线的channel(离线上通知上层)
	httpchan  chan bool
	iskick    bool        // 离线标记
	exit_chan chan bool   // 读退出

	callback NetCallBackIF // 上层回调
	rdelay   time.Duration // 读超时时间
	wdalay   time.Duration // 写超时时间

	keyflag   int32
	encodekey []byte // 加密key
	decodekey []byte // 解密key
	heartbeat int64  // 心跳时间戳
}

/* 创建一个session */
func NewSession(sessionid int64, conn NetConnIF, offchan chan int64, callback NetCallBackIF, readdelay,
	writedelay time.Duration) *Session {
	sess := Session{}
	sess.id = sessionid
	sess.conn = conn
	sess.offchan = offchan
	sess.rdelay = readdelay
	sess.wdalay = writedelay
	sess.sendchan = make(chan []byte, 10000) // 10000个包的缓冲区
	sess.exit_chan = make(chan bool)
	sess.httpchan = make(chan bool)

	sess.callback = callback
	sess.iskick = false
	sess.keyflag = 0
	sess.heartbeat = utils.SecondTimeSince1970() // 新建session初始化心跳
	return &sess
}

// 创建(客户端用的)
func NewClientSession(sessionid int64, offchan chan int64, callback NetCallBackIF, readdelay,
	writedelay time.Duration) *Session {
	sess := Session{}
	sess.id = sessionid
	sess.offchan = offchan
	sess.rdelay = readdelay
	sess.wdalay = writedelay
	sess.sendchan = make(chan []byte, 10000) // 10000个包的缓冲区
	sess.exit_chan = make(chan bool)

	sess.callback = callback
	sess.iskick = false
	sess.heartbeat = utils.SecondTimeSince1970() // 新建session初始化心跳时间
	return &sess
}

// connector专用函数
func (self *Session) SetConn(conn NetConnIF) {
	self.conn = conn
	self.iskick = false
}

/* 开始工作 */
func (self *Session) DoWork() {
	go self.doread()
	go self.dowrite()
}

// RC4加密解密
func (self *Session) SetCipher(encodekey, decodekey []byte) error {
	if len(encodekey) < 1 || len(encodekey) > 256 {
		return rc4.KeySizeError(len(encodekey))
	}

	if len(decodekey) < 1 || len(decodekey) > 256 {
		return rc4.KeySizeError(len(decodekey))
	}

	self.encodekey = encodekey
	self.decodekey = decodekey
	self.keyflag |= SESS_KEYEXCG

	return nil
}

/* 远端的链接地址IP信息 */
func (self *Session) RemoteIP() string {
	addr := self.conn.RemoteAddr().String()
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	return net.ParseIP(host).String()
}

// 发送数据
func (self *Session) SyncSend(data []byte) bool {
	if self.iskick {
		return false
	}

	select {
	case self.sendchan <- data:
		return true
	default:
		// TODO:缓冲区满
		log.Warn("syncSend logcache full ", len(self.sendchan))
		return false
	}
	return true
}

// 设置kick标记
func (self *Session) Kick() {
	if self.iskick {
		return
	}
	self.iskick = true
	if self.conn != nil {
		if error := self.conn.Close(); error != nil {
			log.Errorf("关闭错误：%v", error.Error())
		} else {
			log.Infof(colorized.Gray("连接断开:[%v]"), self.RemoteIP())
		}
	}
}

// 通知owner
func (self *Session) notify() {
	self.offchan <- self.id
	self.httpchan <- true
}

/* 从网络层读取数据 */
func (self *Session) doread() {
	core.Gwg.Add(1)
	// 通知写协调退出
	defer func() {
		self.exit_chan <- true
		if !self.iskick { // 对端退出的
			if err := self.conn.Close(); err != nil {
				log.Warnf("tcp关闭报错:%v", err.Error())
			}
		}
		core.Gwg.Done()
	}()

	buffer := new(bytes.Buffer)
	readbuffer := make([]byte, packet.PACKET_BUFFER_LEN) //读取缓冲区

	// 开始读取网络层数据
	for {
		// 如果离线了就退出
		if self.iskick {
			log.Infof("Session 离线，主动断开")
			break
		}
		// 读取超时
		//self.conn.SetReadDeadline(time.Now().Add(self.rdelay * time.Second))

		// 从网络层读取数据
		n, err := self.conn.Read(readbuffer)
		if err != nil || n == 0 {
			if operr, ok := err.(*net.OpError); ok && operr != nil { // TODO:好像没必要，需要验证
				if operr.Err == syscall.EAGAIN || operr.Err == syscall.EWOULDBLOCK { // 异步操作(没数据了)
					continue
				}
			}
			return
		}

		// 将数据串起来,方便处理粘包
		buffer.Write(readbuffer[:n])

		// 处理粘包
		for {
			// 处理消息
			_, content, errcode := self.decode(buffer)
			if errcode != 0 {
				return
			}
			if content == nil {
				break // 没有包了
			}
			/*
				if msgid == 10063 {
					log.Debugf("pack.data:%v", content)
				}
			*/
			self.callback.handle_input(self.id, content)
		}
	}
}

// 解密&处理粘包
func (self *Session) decode(buffer *bytes.Buffer) (uint16, []byte, int) {
	buff_len := uint16(buffer.Len())
	if buff_len < packet.PACKET_HEAD_LEN { // 长度不够
		return 0, nil, 0
	}

	origin := buffer.Bytes()
	// 获取长度
	var msg_len uint16 = 0
	msg_len = uint16(origin[4])
	msg_len |= uint16(origin[5]) << 8
	if msg_len < packet.PACKET_HEAD_LEN || msg_len > packet.PACKET_BUFFER_LEN || msg_len > buff_len {
		return 0, nil, 0
	}

	content := buffer.Next(int(msg_len))
	//// 解密
	//if self.keyflag&SESS_ENCRYPT != 0 {
	//	decoder, _ := rc4.NewCipher(self.decodekey)
	//	decoder.XORKeyStream(content, content)
	//}

	newBuf := make([]byte, 0, msg_len)
	newBuf = append(newBuf, content...)

	// 解析msgid
	var msgid uint16 = 0
	msgid = uint16(newBuf[2])
	msgid |= uint16(newBuf[3]) << 8
	return msgid, newBuf, 0
}

// 发送到远端
func (self *Session) send(data []byte) error {
	//// 加密
	//if self.keyflag&SESS_ENCRYPT != 0 { // encryption is enabled
	//	encoder, _ := rc4.NewCipher(self.encodekey)
	//	data := pack[HEAD_LEN_SIZE:dlen]
	//	encoder.XORKeyStream(data, data)
	//} else if self.keyflag&SESS_KEYEXCG != 0 { // key is exchanged, encryption is not yet enabled
	//	self.keyflag &^= SESS_KEYEXCG
	//	self.keyflag |= SESS_ENCRYPT
	//}

	ssize := 0
	total_size := len(data)
	for {
		// 发送到远端
		//self.conn.SetWriteDeadline(time.Now().Add(self.wdalay * time.Second))
		n, err := self.conn.Write(data[ssize:total_size])
		if err != nil {
			if operr, ok := err.(*net.OpError); ok && operr != nil { // TODO:好像没必要，需要验证
				if operr.Err == syscall.EAGAIN || operr.Err == syscall.EWOULDBLOCK { // 异步操作(缓冲区满了)
					continue
				} else {
					return err
				}
			}
		}
		ssize += n
		if ssize >= total_size {
			break
		}
	}
	return nil
}

// 发送数据
func (self *Session) dowrite() {
	core.Gwg.Add(1)
	defer func() {
		core.Gwg.Done()
	}()
	for {
		select {
		case <-self.exit_chan: // 读协程退出
			self.notify()
			return
		default:
			select {
			case pack := <-self.sendchan:
				if !self.iskick { // 没有断线
					if err := self.send(pack); err != nil {
						//log.Errorf("DoWrite: %v", err.Error())
					}
				}
			case <-self.exit_chan: // 读协程退出
				self.notify()
				return
			}
		}
	}
}
