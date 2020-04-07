// Code generated by protoc-gen-go. DO NOT EDIT.
// source: protobuf/777.proto

/*
Package protomsg is a generated protocol buffer package.

It is generated from these files:
	protobuf/777.proto
	protobuf/data.proto
	protobuf/dfdc.proto
	protobuf/fruitmary.proto
	protobuf/hall.proto
	protobuf/jpm.proto
	protobuf/lhd.proto
	protobuf/luckfruit.proto
	protobuf/red2black.proto

It has these top-level messages:
	ENTER_GAME_S777_REQ
	ENTER_GAME_S777_RES
	LEAVE_GAME_S777_REQ
	LEAVE_GAME_S777_RES
	START_S777_REQ
	START_S777_RES
	S777Position
	UPDATE_S777_BONUS
	PLAYERS_S777_LIST_RES
	AccountStorageData
	AccountGameData
	Email
	RoomInfo
	GameInfo
	Card
	ENTER_GAME_DFDC_REQ
	ENTER_GAME_DFDC_RES
	LEAVE_GAME_DFDC_REQ
	LEAVE_GAME_DFDC_RES
	START_DFDC_REQ
	START_DFDC_RES
	DFDCPosition
	UPDATE_DFDC_BONUS
	PLAYERS_DFDC_LIST_RES
	ENTER_GAME_FRUITMARY_REQ
	ENTER_GAME_FRUITMARY_RES
	LEAVE_GAME_FRUITMARY_REQ
	LEAVE_GAME_FRUITMARY_RES
	START_MARY_REQ
	START_MARY_RES
	FRUITMARYPosition
	FRUITMARY_Result
	UPDATE_MARY_BONUS
	START_MARY2_REQ
	START_MARY2_RES
	Mary2_Result
	NEXT_MARY_RESULT
	PLAYERS_LIST_RES
	SYNC_SERVER_TIME
	KICK_OUT_HALL
	LOGIN_HALL_REQ
	LOGIN_HALL_RES
	SAFEMONEY_OPERATE_REQ
	SAFEMONEY_OPERATE_RES
	BIND_PHONE_REQ
	BIND_PHONE_RES
	ENTER_ROOM_REQ
	ENTER_ROOM_RES
	EMAILS_REQ
	EMAILS_RES
	EMAIL_READ_REQ
	EMAIL_READ_RES
	EMAIL_REWARD_REQ
	EMAIL_REWARD_RES
	EMAIL_DEL_REQ
	EMAIL_DEL_RES
	UPDATE_MONEY
	EMAIL_NEW
	BROADCAST_MSG
	UPDATE_ROOMLIST
	ENTER_GAME_JPM_REQ
	ENTER_GAME_JPM_RES
	LEAVE_GAME_JPM_REQ
	LEAVE_GAME_JPM_RES
	START_JPM_REQ
	START_JPM_RES
	JPMPosition
	JPM_Result
	UPDATE_JPM_BONUS
	PLAYERS_JPM_LIST_RES
	ENTER_GAME_LHD_REQ
	ENTER_GAME_LHD_RES
	LEAVE_GAME_LHD_REQ
	LEAVE_GAME_LHD_RES
	SWITCH_GAME_STATUS_BROADCAST_LHD
	StatusMsgLHD
	Status_Wait_LHD
	Status_Bet_LHD
	Status_Stop_LHD
	Status_Settle_LHD
	BET_LHD_REQ
	BET_LHD_RES
	CLEAN_BET_LHD_REQ
	CLEAN_BET_LHD_RES
	PLAYERS_LHD_LIST_RES
	ENTER_GAME_LUCKFRUIT_REQ
	ENTER_GAME_LUCKFRUIT_RES
	LEAVE_GAME_LUCKFRUIT_REQ
	LEAVE_GAME_LUCKFRUIT_RES
	START_LUCKFRUIT_REQ
	START_LUCKFRUIT_RES
	LUCKFRUITPosition
	LUCKFRUIT_Result
	UPDATE_LUCKFRUIT_BONUS
	PLAYERS_LUCKFRUIT_LIST_RES
	ENTER_GAME_RED2BLACK_REQ
	ENTER_GAME_RED2BLACK_RES
	LEAVE_GAME_RED2BLACK_REQ
	LEAVE_GAME_RED2BLACK_RES
	SWITCH_GAME_STATUS_BROADCAST
	StatusMsg
	Status_Wait
	Status_Bet
	Status_Stop
	Status_Settle
	BET_RED2BLACK_REQ
	BET_RED2BLACK_RES
	CLEAN_BET_RED2BLACK_REQ
	CLEAN_BET_RED2BLACK_RES
	PLAYERS_RED2BLACK_LIST_RES
*/
package protomsg

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// 网络消息
type S777MSG int32

const (
	S777MSG_UNKNOW_S777              S777MSG = 0
	S777MSG_CS_ENTER_GAME_S777_REQ   S777MSG = 18001
	S777MSG_SC_ENTER_GAME_S777_RES   S777MSG = 18002
	S777MSG_CS_LEAVE_GAME_S777_REQ   S777MSG = 18003
	S777MSG_SC_LEAVE_GAME_S777_RES   S777MSG = 18004
	S777MSG_CS_START_S777_REQ        S777MSG = 18008
	S777MSG_SC_START_S777_RES        S777MSG = 18009
	S777MSG_SC_UPDATE_S777_BONUS     S777MSG = 18010
	S777MSG_CS_PLAYERS_S777_LIST_REQ S777MSG = 18015
	S777MSG_SC_PLAYERS_S777_LIST_RES S777MSG = 18016
)

var S777MSG_name = map[int32]string{
	0:     "UNKNOW_S777",
	18001: "CS_ENTER_GAME_S777_REQ",
	18002: "SC_ENTER_GAME_S777_RES",
	18003: "CS_LEAVE_GAME_S777_REQ",
	18004: "SC_LEAVE_GAME_S777_RES",
	18008: "CS_START_S777_REQ",
	18009: "SC_START_S777_RES",
	18010: "SC_UPDATE_S777_BONUS",
	18015: "CS_PLAYERS_S777_LIST_REQ",
	18016: "SC_PLAYERS_S777_LIST_RES",
}
var S777MSG_value = map[string]int32{
	"UNKNOW_S777":              0,
	"CS_ENTER_GAME_S777_REQ":   18001,
	"SC_ENTER_GAME_S777_RES":   18002,
	"CS_LEAVE_GAME_S777_REQ":   18003,
	"SC_LEAVE_GAME_S777_RES":   18004,
	"CS_START_S777_REQ":        18008,
	"SC_START_S777_RES":        18009,
	"SC_UPDATE_S777_BONUS":     18010,
	"CS_PLAYERS_S777_LIST_REQ": 18015,
	"SC_PLAYERS_S777_LIST_RES": 18016,
}

func (x S777MSG) String() string {
	return proto.EnumName(S777MSG_name, int32(x))
}
func (S777MSG) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// 1-3槽 图案枚举
type S777ID int32

const (
	S777ID_S777Unknow S777ID = 0
	S777ID_S7771      S777ID = 1
	S777ID_S7772      S777ID = 2
	S777ID_S7773      S777ID = 3
	S777ID_S7774      S777ID = 4
	S777ID_S7775      S777ID = 5
	S777ID_S7776      S777ID = 6
	S777ID_S7777      S777ID = 7
)

var S777ID_name = map[int32]string{
	0: "S777Unknow",
	1: "S7771",
	2: "S7772",
	3: "S7773",
	4: "S7774",
	5: "S7775",
	6: "S7776",
	7: "S7777",
}
var S777ID_value = map[string]int32{
	"S777Unknow": 0,
	"S7771":      1,
	"S7772":      2,
	"S7773":      3,
	"S7774":      4,
	"S7775":      5,
	"S7776":      6,
	"S7777":      7,
}

func (x S777ID) String() string {
	return proto.EnumName(S777ID_name, int32(x))
}
func (S777ID) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

// 4槽 图案枚举
type JackPotID int32

const (
	JackPotID_JackPotUnknow JackPotID = 0
	JackPotID_JackPot1      JackPotID = 1
	JackPotID_JackPot2      JackPotID = 2
	JackPotID_JackPot3      JackPotID = 3
	JackPotID_JackPot4      JackPotID = 4
	JackPotID_JackPot5      JackPotID = 5
	JackPotID_JackPot6      JackPotID = 6
)

var JackPotID_name = map[int32]string{
	0: "JackPotUnknow",
	1: "JackPot1",
	2: "JackPot2",
	3: "JackPot3",
	4: "JackPot4",
	5: "JackPot5",
	6: "JackPot6",
}
var JackPotID_value = map[string]int32{
	"JackPotUnknow": 0,
	"JackPot1":      1,
	"JackPot2":      2,
	"JackPot3":      3,
	"JackPot4":      4,
	"JackPot5":      5,
	"JackPot6":      6,
}

func (x JackPotID) String() string {
	return proto.EnumName(JackPotID_name, int32(x))
}
func (JackPotID) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

// 请求进入房间
type ENTER_GAME_S777_REQ struct {
	AccountID uint32 `protobuf:"varint,1,opt,name=AccountID" json:"AccountID,omitempty"`
	RoomID    uint32 `protobuf:"varint,2,opt,name=RoomID" json:"RoomID,omitempty"`
}

func (m *ENTER_GAME_S777_REQ) Reset()                    { *m = ENTER_GAME_S777_REQ{} }
func (m *ENTER_GAME_S777_REQ) String() string            { return proto.CompactTextString(m) }
func (*ENTER_GAME_S777_REQ) ProtoMessage()               {}
func (*ENTER_GAME_S777_REQ) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *ENTER_GAME_S777_REQ) GetAccountID() uint32 {
	if m != nil {
		return m.AccountID
	}
	return 0
}

func (m *ENTER_GAME_S777_REQ) GetRoomID() uint32 {
	if m != nil {
		return m.RoomID
	}
	return 0
}

type ENTER_GAME_S777_RES struct {
	RoomID  uint32          `protobuf:"varint,1,opt,name=RoomID" json:"RoomID,omitempty"`
	Basics  int64           `protobuf:"varint,2,opt,name=Basics" json:"Basics,omitempty"`
	Bonus   map[int32]int64 `protobuf:"bytes,3,rep,name=Bonus" json:"Bonus,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"varint,2,opt,name=value"`
	LastBet int64           `protobuf:"varint,4,opt,name=LastBet" json:"LastBet,omitempty"`
	Bets    []uint64        `protobuf:"varint,5,rep,packed,name=Bets" json:"Bets,omitempty"`
}

func (m *ENTER_GAME_S777_RES) Reset()                    { *m = ENTER_GAME_S777_RES{} }
func (m *ENTER_GAME_S777_RES) String() string            { return proto.CompactTextString(m) }
func (*ENTER_GAME_S777_RES) ProtoMessage()               {}
func (*ENTER_GAME_S777_RES) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *ENTER_GAME_S777_RES) GetRoomID() uint32 {
	if m != nil {
		return m.RoomID
	}
	return 0
}

func (m *ENTER_GAME_S777_RES) GetBasics() int64 {
	if m != nil {
		return m.Basics
	}
	return 0
}

func (m *ENTER_GAME_S777_RES) GetBonus() map[int32]int64 {
	if m != nil {
		return m.Bonus
	}
	return nil
}

func (m *ENTER_GAME_S777_RES) GetLastBet() int64 {
	if m != nil {
		return m.LastBet
	}
	return 0
}

func (m *ENTER_GAME_S777_RES) GetBets() []uint64 {
	if m != nil {
		return m.Bets
	}
	return nil
}

// 请求退出房间
type LEAVE_GAME_S777_REQ struct {
	AccountID uint32 `protobuf:"varint,1,opt,name=AccountID" json:"AccountID,omitempty"`
	RoomID    uint32 `protobuf:"varint,2,opt,name=RoomID" json:"RoomID,omitempty"`
}

func (m *LEAVE_GAME_S777_REQ) Reset()                    { *m = LEAVE_GAME_S777_REQ{} }
func (m *LEAVE_GAME_S777_REQ) String() string            { return proto.CompactTextString(m) }
func (*LEAVE_GAME_S777_REQ) ProtoMessage()               {}
func (*LEAVE_GAME_S777_REQ) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *LEAVE_GAME_S777_REQ) GetAccountID() uint32 {
	if m != nil {
		return m.AccountID
	}
	return 0
}

func (m *LEAVE_GAME_S777_REQ) GetRoomID() uint32 {
	if m != nil {
		return m.RoomID
	}
	return 0
}

type LEAVE_GAME_S777_RES struct {
	Ret    uint32 `protobuf:"varint,1,opt,name=Ret" json:"Ret,omitempty"`
	RoomID uint32 `protobuf:"varint,2,opt,name=RoomID" json:"RoomID,omitempty"`
}

func (m *LEAVE_GAME_S777_RES) Reset()                    { *m = LEAVE_GAME_S777_RES{} }
func (m *LEAVE_GAME_S777_RES) String() string            { return proto.CompactTextString(m) }
func (*LEAVE_GAME_S777_RES) ProtoMessage()               {}
func (*LEAVE_GAME_S777_RES) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *LEAVE_GAME_S777_RES) GetRet() uint32 {
	if m != nil {
		return m.Ret
	}
	return 0
}

func (m *LEAVE_GAME_S777_RES) GetRoomID() uint32 {
	if m != nil {
		return m.RoomID
	}
	return 0
}

// //////////////////////////////////////////// 游戏 /////////////////////////////////////////////
// 请求开始游戏1
type START_S777_REQ struct {
	Bet uint64 `protobuf:"varint,1,opt,name=Bet" json:"Bet,omitempty"`
}

func (m *START_S777_REQ) Reset()                    { *m = START_S777_REQ{} }
func (m *START_S777_REQ) String() string            { return proto.CompactTextString(m) }
func (*START_S777_REQ) ProtoMessage()               {}
func (*START_S777_REQ) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *START_S777_REQ) GetBet() uint64 {
	if m != nil {
		return m.Bet
	}
	return 0
}

type START_S777_RES struct {
	Ret         uint64          `protobuf:"varint,1,opt,name=Ret" json:"Ret,omitempty"`
	PictureList []int32         `protobuf:"varint,2,rep,packed,name=PictureList" json:"PictureList,omitempty"`
	Bonus       map[int32]int64 `protobuf:"bytes,3,rep,name=Bonus" json:"Bonus,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"varint,2,opt,name=value"`
	Money       int64           `protobuf:"varint,4,opt,name=Money" json:"Money,omitempty"`
	TotalOdds   int64           `protobuf:"varint,5,opt,name=TotalOdds" json:"TotalOdds,omitempty"`
	Id          JackPotID       `protobuf:"varint,7,opt,name=id,enum=protomsg.JackPotID" json:"id,omitempty"`
}

func (m *START_S777_RES) Reset()                    { *m = START_S777_RES{} }
func (m *START_S777_RES) String() string            { return proto.CompactTextString(m) }
func (*START_S777_RES) ProtoMessage()               {}
func (*START_S777_RES) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *START_S777_RES) GetRet() uint64 {
	if m != nil {
		return m.Ret
	}
	return 0
}

func (m *START_S777_RES) GetPictureList() []int32 {
	if m != nil {
		return m.PictureList
	}
	return nil
}

func (m *START_S777_RES) GetBonus() map[int32]int64 {
	if m != nil {
		return m.Bonus
	}
	return nil
}

func (m *START_S777_RES) GetMoney() int64 {
	if m != nil {
		return m.Money
	}
	return 0
}

func (m *START_S777_RES) GetTotalOdds() int64 {
	if m != nil {
		return m.TotalOdds
	}
	return 0
}

func (m *START_S777_RES) GetId() JackPotID {
	if m != nil {
		return m.Id
	}
	return JackPotID_JackPotUnknow
}

type S777Position struct {
	Px int32 `protobuf:"varint,1,opt,name=px" json:"px,omitempty"`
	Py int32 `protobuf:"varint,2,opt,name=py" json:"py,omitempty"`
}

func (m *S777Position) Reset()                    { *m = S777Position{} }
func (m *S777Position) String() string            { return proto.CompactTextString(m) }
func (*S777Position) ProtoMessage()               {}
func (*S777Position) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *S777Position) GetPx() int32 {
	if m != nil {
		return m.Px
	}
	return 0
}

func (m *S777Position) GetPy() int32 {
	if m != nil {
		return m.Py
	}
	return 0
}

// 通知更新奖金池
type UPDATE_S777_BONUS struct {
	Bonus map[int32]int64 `protobuf:"bytes,1,rep,name=Bonus" json:"Bonus,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"varint,2,opt,name=value"`
}

func (m *UPDATE_S777_BONUS) Reset()                    { *m = UPDATE_S777_BONUS{} }
func (m *UPDATE_S777_BONUS) String() string            { return proto.CompactTextString(m) }
func (*UPDATE_S777_BONUS) ProtoMessage()               {}
func (*UPDATE_S777_BONUS) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *UPDATE_S777_BONUS) GetBonus() map[int32]int64 {
	if m != nil {
		return m.Bonus
	}
	return nil
}

// 请求S777玩家列表
type PLAYERS_S777_LIST_RES struct {
	Players []*AccountStorageData `protobuf:"bytes,1,rep,name=players" json:"players,omitempty"`
}

func (m *PLAYERS_S777_LIST_RES) Reset()                    { *m = PLAYERS_S777_LIST_RES{} }
func (m *PLAYERS_S777_LIST_RES) String() string            { return proto.CompactTextString(m) }
func (*PLAYERS_S777_LIST_RES) ProtoMessage()               {}
func (*PLAYERS_S777_LIST_RES) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *PLAYERS_S777_LIST_RES) GetPlayers() []*AccountStorageData {
	if m != nil {
		return m.Players
	}
	return nil
}

func init() {
	proto.RegisterType((*ENTER_GAME_S777_REQ)(nil), "protomsg.ENTER_GAME_S777_REQ")
	proto.RegisterType((*ENTER_GAME_S777_RES)(nil), "protomsg.ENTER_GAME_S777_RES")
	proto.RegisterType((*LEAVE_GAME_S777_REQ)(nil), "protomsg.LEAVE_GAME_S777_REQ")
	proto.RegisterType((*LEAVE_GAME_S777_RES)(nil), "protomsg.LEAVE_GAME_S777_RES")
	proto.RegisterType((*START_S777_REQ)(nil), "protomsg.START_S777_REQ")
	proto.RegisterType((*START_S777_RES)(nil), "protomsg.START_S777_RES")
	proto.RegisterType((*S777Position)(nil), "protomsg.S777_position")
	proto.RegisterType((*UPDATE_S777_BONUS)(nil), "protomsg.UPDATE_S777_BONUS")
	proto.RegisterType((*PLAYERS_S777_LIST_RES)(nil), "protomsg.PLAYERS_S777_LIST_RES")
	proto.RegisterEnum("protomsg.S777MSG", S777MSG_name, S777MSG_value)
	proto.RegisterEnum("protomsg.S777ID", S777ID_name, S777ID_value)
	proto.RegisterEnum("protomsg.JackPotID", JackPotID_name, JackPotID_value)
}

func init() { proto.RegisterFile("protobuf/777.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 710 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x54, 0xdd, 0x4e, 0x13, 0x41,
	0x14, 0x66, 0x77, 0xbb, 0x2d, 0x1c, 0x6c, 0x19, 0x06, 0xc4, 0x0d, 0x21, 0xa6, 0x29, 0x89, 0x69,
	0xb8, 0x28, 0x11, 0x94, 0xa2, 0x31, 0x9a, 0x76, 0xbb, 0x21, 0x95, 0xd2, 0x96, 0x99, 0x56, 0xa3,
	0x37, 0x9b, 0xa5, 0x5d, 0x49, 0x03, 0x74, 0x9a, 0xee, 0x54, 0xe9, 0x33, 0xd8, 0x2b, 0x9f, 0xc0,
	0x57, 0xf2, 0xe7, 0x42, 0xbd, 0xd1, 0x17, 0xf0, 0x1d, 0xcc, 0xcc, 0xee, 0x76, 0xa9, 0xac, 0x57,
	0x78, 0xc5, 0xf7, 0x9d, 0xef, 0x9c, 0xef, 0x1c, 0xce, 0x99, 0x2d, 0xe0, 0xc1, 0x90, 0x71, 0x76,
	0x32, 0x7a, 0xb3, 0x5d, 0x2c, 0x16, 0x0b, 0x92, 0xe0, 0x79, 0xf9, 0xe7, 0xc2, 0x3b, 0x5d, 0x5f,
	0x99, 0xaa, 0x5d, 0x87, 0x3b, 0xbe, 0x9c, 0x3b, 0x84, 0x15, 0xab, 0xde, 0xb2, 0x88, 0x7d, 0x50,
	0x3a, 0xb2, 0x6c, 0x5a, 0x2c, 0x16, 0x6d, 0x62, 0x1d, 0xe3, 0x0d, 0x58, 0x28, 0x75, 0x3a, 0x6c,
	0xd4, 0xe7, 0xd5, 0x8a, 0xa1, 0x64, 0x95, 0x7c, 0x9a, 0x44, 0x01, 0xbc, 0x06, 0x49, 0xc2, 0xd8,
	0x45, 0xb5, 0x62, 0xa8, 0x52, 0x0a, 0x58, 0xee, 0xb7, 0x12, 0xe7, 0x46, 0xaf, 0xe4, 0x2b, 0x57,
	0xf3, 0x45, 0xbc, 0xec, 0x78, 0xbd, 0x8e, 0x27, 0x7d, 0x34, 0x12, 0x30, 0xfc, 0x14, 0xf4, 0x32,
	0xeb, 0x8f, 0x3c, 0x43, 0xcb, 0x6a, 0xf9, 0xc5, 0x9d, 0x7c, 0x21, 0xfc, 0x1f, 0x0a, 0x31, 0xee,
	0x05, 0x99, 0x6a, 0xf5, 0xf9, 0x70, 0x4c, 0xfc, 0x32, 0x6c, 0x40, 0xaa, 0xe6, 0x78, 0xbc, 0xec,
	0x72, 0x23, 0x21, 0x8d, 0x43, 0x8a, 0x31, 0x24, 0xca, 0x2e, 0xf7, 0x0c, 0x3d, 0xab, 0xe5, 0x13,
	0x44, 0xe2, 0xf5, 0x7d, 0x80, 0xc8, 0x02, 0x23, 0xd0, 0xce, 0xdc, 0xb1, 0x1c, 0x54, 0x27, 0x02,
	0xe2, 0x55, 0xd0, 0xdf, 0x3a, 0xe7, 0x23, 0x37, 0x18, 0xd2, 0x27, 0x8f, 0xd5, 0x7d, 0x45, 0x2c,
	0xaf, 0x66, 0x95, 0x5e, 0x58, 0xff, 0x65, 0x79, 0xcf, 0xe2, 0xcc, 0xa8, 0x98, 0x87, 0xb8, 0x3c,
	0xb0, 0x11, 0xf0, 0x9f, 0x06, 0x39, 0xc8, 0xd0, 0x56, 0x89, 0xb4, 0xa2, 0x41, 0x10, 0x68, 0xe5,
	0xa0, 0x36, 0x41, 0x04, 0xcc, 0x7d, 0x50, 0xff, 0x4a, 0x9a, 0x69, 0x90, 0xf0, 0x1b, 0x64, 0x61,
	0xb1, 0xd9, 0xeb, 0xf0, 0xd1, 0xd0, 0xad, 0xf5, 0x3c, 0x6e, 0xa8, 0x59, 0x2d, 0xaf, 0x93, 0xab,
	0x21, 0xfc, 0x68, 0xf6, 0x40, 0x9b, 0xd1, 0x81, 0x66, 0xcd, 0x63, 0x6e, 0xb3, 0x0a, 0xfa, 0x11,
	0xeb, 0xbb, 0xe3, 0xe0, 0x32, 0x3e, 0x11, 0x2b, 0x6b, 0x31, 0xee, 0x9c, 0x37, 0xba, 0x5d, 0x71,
	0x1c, 0xa1, 0x44, 0x01, 0xbc, 0x09, 0x6a, 0xaf, 0x6b, 0xa4, 0xb2, 0x4a, 0x3e, 0xb3, 0xb3, 0x12,
	0xf5, 0x7a, 0xee, 0x74, 0xce, 0x9a, 0x8c, 0x57, 0x2b, 0x44, 0xed, 0x75, 0x6f, 0x70, 0xc6, 0x6d,
	0x48, 0xcb, 0x81, 0x07, 0xcc, 0xeb, 0xf1, 0x1e, 0xeb, 0xe3, 0x0c, 0xa8, 0x83, 0xcb, 0xa0, 0x56,
	0x1d, 0x5c, 0x4a, 0x3e, 0x96, 0x75, 0x82, 0x8f, 0x73, 0xef, 0x15, 0x58, 0x6e, 0x37, 0x2b, 0xa5,
	0x56, 0x70, 0xa7, 0x72, 0xa3, 0xde, 0xa6, 0xf8, 0x49, 0xb8, 0x14, 0x45, 0x2e, 0xe5, 0x5e, 0x34,
	0xe8, 0xb5, 0xdc, 0xeb, 0x7b, 0xb9, 0xc1, 0xf8, 0x0d, 0xb8, 0xdd, 0xac, 0x95, 0x5e, 0x59, 0x84,
	0xfa, 0x1d, 0x6a, 0x55, 0xda, 0x92, 0x97, 0xdd, 0x83, 0xd4, 0xe0, 0xdc, 0x19, 0xbb, 0xc3, 0x70,
	0xa4, 0x8d, 0x68, 0xa4, 0xe0, 0x3d, 0x52, 0xce, 0x86, 0xce, 0xa9, 0x5b, 0x71, 0xb8, 0x43, 0xc2,
	0xe4, 0xad, 0x8f, 0x2a, 0xa4, 0x84, 0xd3, 0x11, 0x3d, 0xc0, 0x4b, 0xb0, 0xd8, 0xae, 0x1f, 0xd6,
	0x1b, 0x2f, 0xa5, 0x37, 0x9a, 0xc3, 0x1b, 0xb0, 0x66, 0x52, 0x3b, 0xe6, 0x37, 0x03, 0x7d, 0x9a,
	0x28, 0x42, 0xa5, 0x66, 0x8c, 0x4a, 0xd1, 0x67, 0x5f, 0x35, 0xa9, 0x1d, 0xf3, 0xc9, 0xa0, 0x2f,
	0xd3, 0xda, 0x98, 0x6f, 0x00, 0x7d, 0x9d, 0x28, 0xf8, 0x0e, 0x2c, 0x9b, 0xd4, 0x9e, 0x7d, 0xe0,
	0xe8, 0x9b, 0x2f, 0x50, 0x73, 0x56, 0xa0, 0xe8, 0xfb, 0x44, 0xc1, 0xeb, 0xb0, 0x4a, 0x4d, 0xfb,
	0xda, 0xee, 0xd1, 0x8f, 0x89, 0x82, 0xef, 0x82, 0x61, 0x52, 0x3b, 0x6e, 0x6d, 0xc7, 0xe8, 0xa7,
	0xaf, 0x53, 0x33, 0x56, 0xa7, 0xe8, 0xd7, 0x44, 0xd9, 0x72, 0x21, 0x29, 0x82, 0xd5, 0x0a, 0xce,
	0x00, 0x08, 0xd4, 0xee, 0x9f, 0xf5, 0xd9, 0x3b, 0x34, 0x87, 0x17, 0x40, 0x17, 0xfc, 0x3e, 0x52,
	0x42, 0xb8, 0x83, 0xd4, 0x10, 0xee, 0x22, 0x2d, 0x84, 0x0f, 0x50, 0x22, 0x84, 0x0f, 0x91, 0x1e,
	0xc2, 0x3d, 0x94, 0x0c, 0x61, 0x11, 0xa5, 0xb6, 0x86, 0xb0, 0x30, 0x7d, 0xe4, 0x78, 0x19, 0xd2,
	0x01, 0x99, 0x36, 0xbb, 0x05, 0xf3, 0x41, 0x48, 0xf4, 0x8b, 0x98, 0x68, 0x19, 0x31, 0xd1, 0x35,
	0x62, 0xa2, 0x71, 0xc4, 0x44, 0xef, 0x88, 0xed, 0xa1, 0x64, 0x79, 0xe9, 0x75, 0x7a, 0xc8, 0x18,
	0xdf, 0x0e, 0x9f, 0xca, 0x49, 0x52, 0xa2, 0xdd, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0xea, 0x72,
	0xbf, 0xff, 0x5e, 0x06, 0x00, 0x00,
}
