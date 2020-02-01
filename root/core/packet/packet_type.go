package packet

import (
	"github.com/golang/protobuf/proto"
	"root/core/log"
)

// 将2个Packet的协议体组装到一起, 返回一个新Packet, 协议头使用one Packet的协议头
func PacketMakeup(one IPacket, two IPacket) IPacket {
	packet := NewPacket(nil)
	packet.SetMsgID(one.GetMsgID())

	ret := packet.CatBody(one)
	if ret == false {
		return nil
	}

	ret = packet.CatBody(two)
	if ret == false {
		return nil
	}
	return packet
}
// 将2个Packet的协议体组装到一起, 返回一个新Packet, 协议头使用one Packet的协议头
func PBUnmarshal(bytes []byte, pb proto.Message) proto.Message {
	if err :=proto.Unmarshal(bytes, pb); err != nil{
		log.Panicf("解析错误:%v",err.Error())
	}

	return pb
}

type IPacket interface {
	Reset(msgid uint16)
	SetMsgID(msgid uint16)
	GetDataSize() uint16
	GetSpace() uint16
	GetMsgID() (ret uint16)
	GetData() []byte
	GetWritePos() uint16
	Rrevise(wpos uint16, data interface{}) bool
	CatBody(packet IPacket) bool
	GetBody() []byte
	ReadInt8() int8
	ReadInt16() int16
	ReadInt32() int32
	ReadInt64() int64
	ReadUInt8() uint8
	ReadUInt16() uint16
	ReadUInt32() uint32
	ReadUInt64() uint64
	ReadFloat32() float32
	ReadFloat64() float64
	ReadBool() bool
	ReadString() string
	ReadBytes() []byte
	WriteInt8(value int8) bool
	WriteInt16(value int16) bool
	WriteInt32(value int32) bool
	WriteInt64(value int64) bool
	WriteUInt8(value uint8) bool
	WriteUInt16(value uint16) bool
	WriteUInt32(value uint32) bool
	WriteUInt64(value uint64) bool
	WriteFloat32(value float32) bool
	WriteFloat64(value float64) bool
	WriteBool(value bool) bool
	WriteString(s string) bool
	WriteBytes(buff []byte) bool
}
