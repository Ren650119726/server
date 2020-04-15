package main

import (
	"bytes"
	"fmt"
	"root/common"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/protomsg"
)

type (
	Logic struct {
		owner       *core.Actor
		init        bool // 重新建立连接是否需要拉取所有数据
		ListenActor *core.Actor
	}
)

var AccountID = uint32(0)

//var addr = "47.108.87.29"

var addr = "192.168.8.111"

func NewLogic() *Logic {
	return &Logic{}
}

var client_Global *Client

func (self *Logic) Init(actor *core.Actor) bool {
	self.owner = actor

	client_Global = NewWebsocketClient(addr+":41000", "/connect")
	client_Global.connect()
	fmt.Println("connected success :", client_Global.ws.RemoteAddr())
	go func() {
		for {
			recv := make([]byte, 65535)
			n, err := client_Global.ws.Read(recv)
			if err != nil {
				log.Warnf("err:%v", err.Error())
				continue
			}
			//log.Infof("读出%v个字节",n)
			recv = recv[0:n]
			buffer := new(bytes.Buffer)
			buffer.Write(recv)
			_, content, errcode := self.decode(buffer)
			if errcode != 0 {
				log.Warnf("错误:%v", errcode)
			}
			self.HandleMessage(0, content, 0)
		}
	}()

	self.owner.AddTimer(30000, -1, func(dt int64) {
		Send2Hall(protomsg.MSG_CLIENT_KEEPALIVE.UInt16(), nil)
	})
	//login([]string{"aabbcc"})
	return true
}

// 解密&处理粘包
func (self *Logic) decode(buffer *bytes.Buffer) (uint16, []byte, int) {
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

	newBuf := make([]byte, 0, msg_len)
	newBuf = append(newBuf, content...)

	// 解析msgid
	var msgid uint16 = 0
	msgid = uint16(newBuf[2])
	msgid |= uint16(newBuf[3]) << 8
	return msgid, newBuf, 0
}

func (self *Logic) registerHall() {

}

func (self *Logic) Stop() {

}

func (self *Logic) HandleMessage(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	case protomsg.MSG_SC_LOGIN_HALL_RES.UInt16():
		pb := packet.PBUnmarshal(pack.ReadBytes(), &protomsg.LOGIN_HALL_RES{}).(*protomsg.LOGIN_HALL_RES)
		AccountID = pb.GetAccount().AccountId
		log.Infof(colorized.Blue("登陆成功：%+v"), pb)
		if pb.AccountData.RoomID != 0 {
			game := NewGame()
			msgchan := make(chan core.IMessage, 10000)
			actor := core.NewActor(common.EActorType_MAIN.Int32(), game, msgchan)
			core.CoreRegisteActor(actor)
		}

	case protomsg.MSG_SC_SYNC_SERVER_TIME.UInt16():
		pb := packet.PBUnmarshal(pack.ReadBytes(), &protomsg.SYNC_SERVER_TIME{}).(*protomsg.SYNC_SERVER_TIME)
		log.Infof(colorized.Blue("同步服务器时间：%+v"), pb)

	case protomsg.MSG_SC_UPDATE_ROOMLIST.UInt16():
		pb := packet.PBUnmarshal(pack.ReadBytes(), &protomsg.UPDATE_ROOMLIST{}).(*protomsg.UPDATE_ROOMLIST)
		for _, v := range pb.GetGames() {
			log.Infof(colorized.Blue("服务器更新房间 游戏:%v 房间:%+v"), common.EGameType(v.GetGameType()), v.GetRooms())
		}
	case protomsg.MSG_SC_ENTER_ROOM_RES.UInt16():
		pb := packet.PBUnmarshal(pack.ReadBytes(), &protomsg.ENTER_ROOM_RES{}).(*protomsg.ENTER_ROOM_RES)
		log.Infof(colorized.Blue("可以进入房间 房间:%+v"), pb)

		game := NewGame()
		game.roomID = pb.GetRoomID()
		msgchan := make(chan core.IMessage, 10000)
		actor := core.NewActor(common.EActorType_MAIN.Int32(), game, msgchan)
		core.CoreRegisteActor(actor)
	}

	return false
}
