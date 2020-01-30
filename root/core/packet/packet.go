package packet

import (
	"root/core/log"
	"encoding/binary"
	"math"
)

/* 常量定义 */
const (
	PACKET_HEAD_LEN   = 6     //包头长度
	PACKET_BUFFER_LEN = 65535 //包缓冲区长度
	PACKET_CAP_LEN    = 64    //包缓冲区动态增加长度
)

type (
	// 协议头结构, 字段顺序不能变, 与客户端保持一致
	//PacketHead struct {
	//	reserved uint16 //保留字段 2字节, 未使用
	//	msgid    uint16 //消息编号 2字节
	//	size     uint16 //消息总长度 2字节 (消息头6字节长度 + 消息体总长度)
	//}

	packet struct {
		data []byte
		rpos uint16
		wpos uint16
	}
)

func NewPacket(data []byte) *packet {
	packet := &packet{}
	if data == nil {
		packet.data = make([]byte, PACKET_HEAD_LEN, PACKET_CAP_LEN)
		packet.rpos = PACKET_HEAD_LEN
		packet.wpos = PACKET_HEAD_LEN
	} else {
		packet.data = data
		packet.rpos = PACKET_HEAD_LEN
		packet.wpos = uint16(len(data))
	}
	return packet
}

func (self *packet) Reset(msgid uint16) {
	self.wpos = PACKET_HEAD_LEN
	self.rpos = PACKET_HEAD_LEN
	self.data = self.data[:PACKET_HEAD_LEN]
	self.SetMsgID(msgid)
}

func (self *packet) SetMsgID(msgid uint16) {
	self.data[2] = uint8(msgid)
	self.data[3] = uint8(msgid >> 8)
}
func (self *packet) HeadByte() []byte {
	return self.data[:PACKET_HEAD_LEN]
}

func (self *packet) SetHeadByte(h []byte){
	 self.data = h
}

func (self *packet) Rrevise(wpos uint16, data interface{}) bool {

	if wpos >= self.wpos {
		return false
	}

	switch data.(type) {
	case int8:
		if wpos+1 >= self.wpos {
			return false
		}
		value := data.(int8)
		self.data[wpos] = uint8(value)
		return true
	case uint8:
		if wpos+1 >= self.wpos {
			return false
		}
		value := data.(uint8)
		self.data[wpos] = value
		return true
	case int16:
		if wpos+2 >= self.wpos {
			return false
		}
		value := data.(int16)
		self.data[wpos] = uint8(value)
		self.data[wpos+1] = uint8(value >> 8)
		return true
	case uint16:
		if wpos+2 >= self.wpos {
			return false
		}
		value := data.(uint16)
		self.data[wpos] = uint8(value)
		self.data[wpos+1] = uint8(value >> 8)
		return true
	case int32:
		if wpos+4 >= self.wpos {
			return false
		}

		value := data.(int32)
		self.data[wpos] = uint8(value)
		self.data[wpos+1] = uint8(value >> 8)
		self.data[wpos+2] = uint8(value >> 16)
		self.data[wpos+3] = uint8(value >> 24)
		return true
	case uint32:
		if wpos+4 >= self.wpos {
			return false
		}

		value := data.(uint32)
		self.data[wpos] = uint8(value)
		self.data[wpos+1] = uint8(value >> 8)
		self.data[wpos+2] = uint8(value >> 16)
		self.data[wpos+3] = uint8(value >> 24)
		return true
	case int64:
		if wpos+8 >= self.wpos {
			return false
		}
		value := data.(int64)
		self.data[wpos] = uint8(value)
		self.data[wpos+1] = uint8(value >> 8)
		self.data[wpos+2] = uint8(value >> 16)
		self.data[wpos+3] = uint8(value >> 24)
		self.data[wpos+4] = uint8(value >> 32)
		self.data[wpos+5] = uint8(value >> 40)
		self.data[wpos+6] = uint8(value >> 48)
		self.data[wpos+7] = uint8(value >> 56)
		return true
	case uint64:
		if wpos+8 >= self.wpos {
			return false
		}
		value := data.(uint64)
		self.data[wpos] = uint8(value)
		self.data[wpos+1] = uint8(value >> 8)
		self.data[wpos+2] = uint8(value >> 16)
		self.data[wpos+3] = uint8(value >> 24)
		self.data[wpos+4] = uint8(value >> 32)
		self.data[wpos+5] = uint8(value >> 40)
		self.data[wpos+6] = uint8(value >> 48)
		self.data[wpos+7] = uint8(value >> 56)
		return true
	}

	log.Error("Unallowed data type, Only allow integer types")
	return false
}

func (self *packet) GetDataSize() uint16 {
	return self.wpos
}

func (self *packet) GetSpace() uint16 {
	return PACKET_BUFFER_LEN - self.wpos
}

func (self *packet) GetMsgID() (ret uint16) {
	ret = uint16(self.data[2])
	ret |= uint16(self.data[3]) << 8
	return ret
}

func (self *packet) GetMsgSize() (ret uint16) {
	ret = uint16(self.data[4])
	ret |= uint16(self.data[5]) << 8
	return ret
}

func (self *packet) GetData() []byte {
	self.data[4] = uint8(self.wpos)
	self.data[5] = uint8(self.wpos >> 8)
	return self.data
}

func (self *packet) GetBody() []byte {
	body := self.data[PACKET_HEAD_LEN:self.wpos]
	return body
}

func (self *packet) CatBody(packet IPacket) bool {

	buff := packet.GetBody()
	buff_len := uint16(len(buff))
	size := self.wpos + buff_len
	if size >= PACKET_BUFFER_LEN {
		return false
	}

	self.data = append(self.data, buff...)
	self.wpos = size
	return true
}

//func (self *packet) GetReadPos() uint16 {
//	return self.rpos
//}
//
func (self *packet) GetWritePos() uint16 {
	return self.wpos
}

func (self *packet) ReadInt8() (ret int8) {
	i := self.rpos + 1
	if i < 0 || i > self.wpos || i > uint16(len(self.data)) {
		return 0
	}
	self.rpos = i
	ret = int8(self.data[i-1])
	return ret
}

func (self *packet) ReadInt16() (ret int16) {
	i := self.rpos + 2
	if i < 0 || i > self.wpos || i > uint16(len(self.data)) {
		return 0
	}
	self.rpos = i
	ret = int16(self.data[i-2])
	ret |= int16(self.data[i-1]) << 8
	return ret
}

func (self *packet) ReadInt32() (ret int32) {
	i := self.rpos + 4
	if i < 0 || i > self.wpos || i > uint16(len(self.data)) {
		return 0
	}
	self.rpos = i
	ret = int32(self.data[i-4])
	ret |= int32(self.data[i-3]) << 8
	ret |= int32(self.data[i-2]) << 16
	ret |= int32(self.data[i-1]) << 24
	return ret
}

func (self *packet) ReadInt64() (ret int64) {
	i := self.rpos + 8
	if i < 0 || i > self.wpos || i > uint16(len(self.data)) {
		return 0
	}
	self.rpos = i
	ret = int64(self.data[i-8])
	ret |= int64(self.data[i-7]) << 8
	ret |= int64(self.data[i-6]) << 16
	ret |= int64(self.data[i-5]) << 24
	ret |= int64(self.data[i-4]) << 32
	ret |= int64(self.data[i-3]) << 40
	ret |= int64(self.data[i-2]) << 48
	ret |= int64(self.data[i-1]) << 56
	return ret
}

func (self *packet) ReadUInt8() (ret uint8) {
	i := self.rpos + 1
	if i < 0 || i > self.wpos || i > uint16(len(self.data)) {
		return 0
	}
	self.rpos = i
	ret = uint8(self.data[i-1])
	return ret
}

func (self *packet) ReadUInt16() (ret uint16) {
	i := self.rpos + 2
	if i < 0 || i > self.wpos || i > uint16(len(self.data)) {
		return 0
	}
	self.rpos = i
	ret = uint16(self.data[i-2])
	ret |= uint16(self.data[i-1]) << 8
	return ret
}

func (self *packet) ReadUInt32() (ret uint32) {
	i := self.rpos + 4
	if i < 0 || i > self.wpos || i > uint16(len(self.data)) {
		return 0
	}
	self.rpos = i
	ret = uint32(self.data[i-4])
	ret |= uint32(self.data[i-3]) << 8
	ret |= uint32(self.data[i-2]) << 16
	ret |= uint32(self.data[i-1]) << 24
	return ret
}

func (self *packet) ReadUInt64() (ret uint64) {
	i := self.rpos + 8
	if i < 0 || i > self.wpos || i > uint16(len(self.data)) {
		return 0
	}
	self.rpos = i
	ret = uint64(self.data[i-8])
	ret |= uint64(self.data[i-7]) << 8
	ret |= uint64(self.data[i-6]) << 16
	ret |= uint64(self.data[i-5]) << 24
	ret |= uint64(self.data[i-4]) << 32
	ret |= uint64(self.data[i-3]) << 40
	ret |= uint64(self.data[i-2]) << 48
	ret |= uint64(self.data[i-1]) << 56
	return ret
}

func (self *packet) ReadFloat32() (ret float32) {
	i := self.rpos + 4
	if i < 0 || i > self.wpos || i > uint16(len(self.data)) {
		return 0
	}
	bits := binary.LittleEndian.Uint32(self.data[self.rpos:i])
	ret = math.Float32frombits(bits)
	self.rpos = i
	return ret
}

func (self *packet) ReadFloat64() (ret float64) {
	i := self.rpos + 8
	if i < 0 || i > self.wpos || i > uint16(len(self.data)) {
		return 0
	}
	bits := binary.LittleEndian.Uint64(self.data[self.rpos:i])
	ret = math.Float64frombits(bits)
	self.rpos = i
	return ret
}

func (self *packet) ReadBool() (ret bool) {
	i := self.rpos + 1
	if i < 0 || i > self.wpos || i > uint16(len(self.data)) {
		return false
	}
	self.rpos = i
	v := uint8(self.data[i-1])
	ret = false
	if v != 0 {
		ret = true
	}
	return ret
}
func (self *packet) ReadString() (ret string) {

	size := self.ReadUInt16()
	if size <= 0 {
		return ""
	}

	i := self.rpos + size
	if i < 0 || i > self.wpos || i > uint16(len(self.data)) {
		return "nil"
	}

	ret = string(self.data[self.rpos:i])
	self.rpos = i
	return ret
}

func (self *packet) ReadBytes() (ret []byte) {
	size := self.GetMsgSize() - 6
	if size <= 0 {
		return nil
	}

	i := self.rpos + size
	if i < 0 || i > self.wpos || i > uint16(len(self.data)) {
		return nil
	}

	ret = self.data[self.rpos:i]
	self.rpos = i
	return ret
}

func (self *packet) WriteInt8(value int8) bool {
	size := self.wpos + 1
	if size >= PACKET_BUFFER_LEN {
		return false
	}
	self.data = append(self.data,
		uint8(value))
	self.wpos = size
	return true
}

func (self *packet) WriteInt16(value int16) bool {
	size := self.wpos + 2
	if size >= PACKET_BUFFER_LEN {
		return false
	}
	self.data = append(self.data,
		uint8(value),
		uint8(value>>8))
	self.wpos = size
	return true
}

func (self *packet) WriteInt32(value int32) bool {
	size := self.wpos + 4
	if size >= PACKET_BUFFER_LEN {
		return false
	}
	self.data = append(self.data,
		uint8(value),
		uint8(value>>8),
		uint8(value>>16),
		uint8(value>>24))
	self.wpos = size
	return true
}

func (self *packet) WriteInt64(value int64) bool {
	size := self.wpos + 8
	if size >= PACKET_BUFFER_LEN {
		return false
	}
	self.data = append(self.data,
		uint8(value),
		uint8(value>>8),
		uint8(value>>16),
		uint8(value>>24),
		uint8(value>>32),
		uint8(value>>40),
		uint8(value>>48),
		uint8(value>>56))
	self.wpos = size
	return true
}

func (self *packet) WriteUInt8(value uint8) bool {
	size := self.wpos + 1
	if size >= PACKET_BUFFER_LEN {
		return false
	}
	self.data = append(self.data,
		uint8(value))
	self.wpos = size
	return true
}

func (self *packet) WriteUInt16(value uint16) bool {
	size := self.wpos + 2
	if size >= PACKET_BUFFER_LEN {
		return false
	}
	self.data = append(self.data,
		uint8(value),
		uint8(value>>8))
	self.wpos = size
	return true
}

func (self *packet) WriteUInt32(value uint32) bool {
	size := self.wpos + 4
	if size >= PACKET_BUFFER_LEN {
		return false
	}
	self.data = append(self.data,
		uint8(value),
		uint8(value>>8),
		uint8(value>>16),
		uint8(value>>24))
	self.wpos = size
	return true
}

func (self *packet) WriteUInt64(value uint64) bool {
	size := self.wpos + 8
	if size >= PACKET_BUFFER_LEN {
		return false
	}
	self.data = append(self.data,
		uint8(value),
		uint8(value>>8),
		uint8(value>>16),
		uint8(value>>24),
		uint8(value>>32),
		uint8(value>>40),
		uint8(value>>48),
		uint8(value>>56))
	self.wpos = size
	return true
}

func (self *packet) WriteFloat32(value float32) bool {
	size := self.wpos + 4
	if size >= PACKET_BUFFER_LEN {
		return false
	}

	bits := math.Float32bits(value)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)
	self.data = append(self.data, bytes...)
	self.wpos = size
	return true
}

func (self *packet) WriteFloat64(value float64) bool {
	size := self.wpos + 8
	if size >= PACKET_BUFFER_LEN {
		return false
	}
	bits := math.Float64bits(value)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	self.data = append(self.data, bytes...)
	self.wpos = size
	return true
}

func (self *packet) WriteBool(value bool) bool {
	size := self.wpos + 1
	if size >= PACKET_BUFFER_LEN {
		return false
	}

	x := 0
	if value == true {
		x = 1
	}

	self.data = append(self.data,
		uint8(x))
	self.wpos = size
	return true
}

func (self *packet) WriteString(s string) bool {

	str_len := uint16(len(s))
	ret := self.WriteUInt16(str_len)
	if ret == false {
		return false
	}

	size := self.wpos + str_len
	if size >= PACKET_BUFFER_LEN {
		return false
	}

	//var strbytes []byte = []byte(s)
	//self.data = append(self.data, strbytes...)
	self.data = append(self.data, s...)
	self.wpos = size
	return true
}

func (self *packet) WriteBytes(buff []byte) bool {
	buff_len := uint16(len(buff))
	//ret := self.WriteUInt16(buff_len)
	//if ret == false {
	//	return false
	//}

	size := self.wpos + buff_len
	if size >= PACKET_BUFFER_LEN {
		return false
	}

	self.data = append(self.data, buff...)
	self.wpos = size
	return true
}
