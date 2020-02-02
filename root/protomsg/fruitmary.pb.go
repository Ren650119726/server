// Code generated by protoc-gen-go. DO NOT EDIT.
// source: protobuf/fruitmary.proto

package protomsg

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// 网络消息
type FRUITMARYMSG int32

const (
	FRUITMARYMSG_UNKNOW_FRUITMARY            FRUITMARYMSG = 0
	FRUITMARYMSG_CS_ENTER_GAME_FRUITMARY_REQ FRUITMARYMSG = 20001
	FRUITMARYMSG_SC_ENTER_GAME_FRUITMARY_RES FRUITMARYMSG = 20002
	FRUITMARYMSG_CS_LEAVE_GAME_FRUITMARY_REQ FRUITMARYMSG = 20003
	FRUITMARYMSG_SC_LEAVE_GAME_FRUITMARY_RES FRUITMARYMSG = 20004
	FRUITMARYMSG_CS_START_MARY_REQ           FRUITMARYMSG = 20008
	FRUITMARYMSG_SC_START_MARY_RES           FRUITMARYMSG = 20009
	FRUITMARYMSG_SC_UPDATE_MARY_BONUS        FRUITMARYMSG = 20010
	FRUITMARYMSG_CS_START_MARY2_REQ          FRUITMARYMSG = 20011
	FRUITMARYMSG_SC_START_MARY2_RES          FRUITMARYMSG = 20012
	FRUITMARYMSG_CS_NEXT_MARY_RESULT         FRUITMARYMSG = 20013
)

var FRUITMARYMSG_name = map[int32]string{
	0:     "UNKNOW_FRUITMARY",
	20001: "CS_ENTER_GAME_FRUITMARY_REQ",
	20002: "SC_ENTER_GAME_FRUITMARY_RES",
	20003: "CS_LEAVE_GAME_FRUITMARY_REQ",
	20004: "SC_LEAVE_GAME_FRUITMARY_RES",
	20008: "CS_START_MARY_REQ",
	20009: "SC_START_MARY_RES",
	20010: "SC_UPDATE_MARY_BONUS",
	20011: "CS_START_MARY2_REQ",
	20012: "SC_START_MARY2_RES",
	20013: "CS_NEXT_MARY_RESULT",
}
var FRUITMARYMSG_value = map[string]int32{
	"UNKNOW_FRUITMARY":            0,
	"CS_ENTER_GAME_FRUITMARY_REQ": 20001,
	"SC_ENTER_GAME_FRUITMARY_RES": 20002,
	"CS_LEAVE_GAME_FRUITMARY_REQ": 20003,
	"SC_LEAVE_GAME_FRUITMARY_RES": 20004,
	"CS_START_MARY_REQ":           20008,
	"SC_START_MARY_RES":           20009,
	"SC_UPDATE_MARY_BONUS":        20010,
	"CS_START_MARY2_REQ":          20011,
	"SC_START_MARY2_RES":          20012,
	"CS_NEXT_MARY_RESULT":         20013,
}

func (x FRUITMARYMSG) String() string {
	return proto.EnumName(FRUITMARYMSG_name, int32(x))
}
func (FRUITMARYMSG) EnumDescriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

// 游戏1 图案枚举
type Fruit1ID int32

const (
	Fruit1ID_Fruit1Unknow     Fruit1ID = 0
	Fruit1ID_Fruit1Wild       Fruit1ID = 1
	Fruit1ID_Fruit1Bonus      Fruit1ID = 2
	Fruit1ID_Fruit1Scatter    Fruit1ID = 3
	Fruit1ID_Fruit1Bar        Fruit1ID = 4
	Fruit1ID_Fruit1Cherry     Fruit1ID = 5
	Fruit1ID_Fruit1Bell       Fruit1ID = 6
	Fruit1ID_Fruit1Pineapple  Fruit1ID = 7
	Fruit1ID_Fruit1Grap       Fruit1ID = 8
	Fruit1ID_Fruit1Mango      Fruit1ID = 9
	Fruit1ID_Fruit1Watermelon Fruit1ID = 10
	Fruit1ID_Fruit1Banana     Fruit1ID = 11
)

var Fruit1ID_name = map[int32]string{
	0:  "Fruit1Unknow",
	1:  "Fruit1Wild",
	2:  "Fruit1Bonus",
	3:  "Fruit1Scatter",
	4:  "Fruit1Bar",
	5:  "Fruit1Cherry",
	6:  "Fruit1Bell",
	7:  "Fruit1Pineapple",
	8:  "Fruit1Grap",
	9:  "Fruit1Mango",
	10: "Fruit1Watermelon",
	11: "Fruit1Banana",
}
var Fruit1ID_value = map[string]int32{
	"Fruit1Unknow":     0,
	"Fruit1Wild":       1,
	"Fruit1Bonus":      2,
	"Fruit1Scatter":    3,
	"Fruit1Bar":        4,
	"Fruit1Cherry":     5,
	"Fruit1Bell":       6,
	"Fruit1Pineapple":  7,
	"Fruit1Grap":       8,
	"Fruit1Mango":      9,
	"Fruit1Watermelon": 10,
	"Fruit1Banana":     11,
}

func (x Fruit1ID) String() string {
	return proto.EnumName(Fruit1ID_name, int32(x))
}
func (Fruit1ID) EnumDescriptor() ([]byte, []int) { return fileDescriptor1, []int{1} }

// 游戏2 图案枚举
type Fruit2ID int32

const (
	Fruit2ID_Fruit2Unknow     Fruit2ID = 0
	Fruit2ID_Fruit2Watermelon Fruit2ID = 1
	Fruit2ID_Fruit2Grap       Fruit2ID = 2
	Fruit2ID_Fruit2Mango      Fruit2ID = 3
	Fruit2ID_Fruit2Cherry     Fruit2ID = 4
	Fruit2ID_Fruit2Banana     Fruit2ID = 5
	Fruit2ID_Fruit2Orange     Fruit2ID = 6
	Fruit2ID_Fruit2Pineapple  Fruit2ID = 7
	Fruit2ID_Fruit2Bomb       Fruit2ID = 8
)

var Fruit2ID_name = map[int32]string{
	0: "Fruit2Unknow",
	1: "Fruit2Watermelon",
	2: "Fruit2Grap",
	3: "Fruit2Mango",
	4: "Fruit2Cherry",
	5: "Fruit2Banana",
	6: "Fruit2Orange",
	7: "Fruit2Pineapple",
	8: "Fruit2Bomb",
}
var Fruit2ID_value = map[string]int32{
	"Fruit2Unknow":     0,
	"Fruit2Watermelon": 1,
	"Fruit2Grap":       2,
	"Fruit2Mango":      3,
	"Fruit2Cherry":     4,
	"Fruit2Banana":     5,
	"Fruit2Orange":     6,
	"Fruit2Pineapple":  7,
	"Fruit2Bomb":       8,
}

func (x Fruit2ID) String() string {
	return proto.EnumName(Fruit2ID_name, int32(x))
}
func (Fruit2ID) EnumDescriptor() ([]byte, []int) { return fileDescriptor1, []int{2} }

// 请求进入房间
type ENTER_GAME_FRUITMARY_REQ struct {
	AccountID uint32 `protobuf:"varint,1,opt,name=AccountID" json:"AccountID,omitempty"`
	RoomID    uint32 `protobuf:"varint,2,opt,name=RoomID" json:"RoomID,omitempty"`
}

func (m *ENTER_GAME_FRUITMARY_REQ) Reset()                    { *m = ENTER_GAME_FRUITMARY_REQ{} }
func (m *ENTER_GAME_FRUITMARY_REQ) String() string            { return proto.CompactTextString(m) }
func (*ENTER_GAME_FRUITMARY_REQ) ProtoMessage()               {}
func (*ENTER_GAME_FRUITMARY_REQ) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

func (m *ENTER_GAME_FRUITMARY_REQ) GetAccountID() uint32 {
	if m != nil {
		return m.AccountID
	}
	return 0
}

func (m *ENTER_GAME_FRUITMARY_REQ) GetRoomID() uint32 {
	if m != nil {
		return m.RoomID
	}
	return 0
}

type ENTER_GAME_FRUITMARY_RES struct {
	RoomID       uint32                                         `protobuf:"varint,1,opt,name=RoomID" json:"RoomID,omitempty"`
	Basics       int64                                          `protobuf:"varint,2,opt,name=Basics" json:"Basics,omitempty"`
	Bonus        int64                                          `protobuf:"varint,3,opt,name=Bonus" json:"Bonus,omitempty"`
	LastBet      int64                                          `protobuf:"varint,4,opt,name=LastBet" json:"LastBet,omitempty"`
	Bets         []uint64                                       `protobuf:"varint,5,rep,packed,name=Bets" json:"Bets,omitempty"`
	Ratio        map[int32]*ENTER_GAME_FRUITMARY_RES_FruitRatio `protobuf:"bytes,6,rep,name=Ratio" json:"Ratio,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Mary2_Result *START_MARY2_RES                               `protobuf:"bytes,7,opt,name=Mary2_Result,json=Mary2Result" json:"Mary2_Result,omitempty"`
	FeeCount     int32                                          `protobuf:"varint,8,opt,name=FeeCount" json:"FeeCount,omitempty"`
}

func (m *ENTER_GAME_FRUITMARY_RES) Reset()                    { *m = ENTER_GAME_FRUITMARY_RES{} }
func (m *ENTER_GAME_FRUITMARY_RES) String() string            { return proto.CompactTextString(m) }
func (*ENTER_GAME_FRUITMARY_RES) ProtoMessage()               {}
func (*ENTER_GAME_FRUITMARY_RES) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{1} }

func (m *ENTER_GAME_FRUITMARY_RES) GetRoomID() uint32 {
	if m != nil {
		return m.RoomID
	}
	return 0
}

func (m *ENTER_GAME_FRUITMARY_RES) GetBasics() int64 {
	if m != nil {
		return m.Basics
	}
	return 0
}

func (m *ENTER_GAME_FRUITMARY_RES) GetBonus() int64 {
	if m != nil {
		return m.Bonus
	}
	return 0
}

func (m *ENTER_GAME_FRUITMARY_RES) GetLastBet() int64 {
	if m != nil {
		return m.LastBet
	}
	return 0
}

func (m *ENTER_GAME_FRUITMARY_RES) GetBets() []uint64 {
	if m != nil {
		return m.Bets
	}
	return nil
}

func (m *ENTER_GAME_FRUITMARY_RES) GetRatio() map[int32]*ENTER_GAME_FRUITMARY_RES_FruitRatio {
	if m != nil {
		return m.Ratio
	}
	return nil
}

func (m *ENTER_GAME_FRUITMARY_RES) GetMary2_Result() *START_MARY2_RES {
	if m != nil {
		return m.Mary2_Result
	}
	return nil
}

func (m *ENTER_GAME_FRUITMARY_RES) GetFeeCount() int32 {
	if m != nil {
		return m.FeeCount
	}
	return 0
}

type ENTER_GAME_FRUITMARY_RES_FruitRatio struct {
	ID     Fruit2ID `protobuf:"varint,1,opt,name=ID,enum=protomsg.Fruit2ID" json:"ID,omitempty"`
	Single int32    `protobuf:"varint,2,opt,name=Single" json:"Single,omitempty"`
	Same_2 int32    `protobuf:"varint,3,opt,name=Same_2,json=Same2" json:"Same_2,omitempty"`
	Same_3 int32    `protobuf:"varint,4,opt,name=Same_3,json=Same3" json:"Same_3,omitempty"`
	Same_4 int32    `protobuf:"varint,5,opt,name=Same_4,json=Same4" json:"Same_4,omitempty"`
}

func (m *ENTER_GAME_FRUITMARY_RES_FruitRatio) Reset()         { *m = ENTER_GAME_FRUITMARY_RES_FruitRatio{} }
func (m *ENTER_GAME_FRUITMARY_RES_FruitRatio) String() string { return proto.CompactTextString(m) }
func (*ENTER_GAME_FRUITMARY_RES_FruitRatio) ProtoMessage()    {}
func (*ENTER_GAME_FRUITMARY_RES_FruitRatio) Descriptor() ([]byte, []int) {
	return fileDescriptor1, []int{1, 0}
}

func (m *ENTER_GAME_FRUITMARY_RES_FruitRatio) GetID() Fruit2ID {
	if m != nil {
		return m.ID
	}
	return Fruit2ID_Fruit2Unknow
}

func (m *ENTER_GAME_FRUITMARY_RES_FruitRatio) GetSingle() int32 {
	if m != nil {
		return m.Single
	}
	return 0
}

func (m *ENTER_GAME_FRUITMARY_RES_FruitRatio) GetSame_2() int32 {
	if m != nil {
		return m.Same_2
	}
	return 0
}

func (m *ENTER_GAME_FRUITMARY_RES_FruitRatio) GetSame_3() int32 {
	if m != nil {
		return m.Same_3
	}
	return 0
}

func (m *ENTER_GAME_FRUITMARY_RES_FruitRatio) GetSame_4() int32 {
	if m != nil {
		return m.Same_4
	}
	return 0
}

// 请求退出房间
type LEAVE_GAME_FRUITMARY_REQ struct {
	AccountID uint32 `protobuf:"varint,1,opt,name=AccountID" json:"AccountID,omitempty"`
	RoomID    uint32 `protobuf:"varint,2,opt,name=RoomID" json:"RoomID,omitempty"`
}

func (m *LEAVE_GAME_FRUITMARY_REQ) Reset()                    { *m = LEAVE_GAME_FRUITMARY_REQ{} }
func (m *LEAVE_GAME_FRUITMARY_REQ) String() string            { return proto.CompactTextString(m) }
func (*LEAVE_GAME_FRUITMARY_REQ) ProtoMessage()               {}
func (*LEAVE_GAME_FRUITMARY_REQ) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{2} }

func (m *LEAVE_GAME_FRUITMARY_REQ) GetAccountID() uint32 {
	if m != nil {
		return m.AccountID
	}
	return 0
}

func (m *LEAVE_GAME_FRUITMARY_REQ) GetRoomID() uint32 {
	if m != nil {
		return m.RoomID
	}
	return 0
}

type LEAVE_GAME_FRUITMARY_RES struct {
	Ret    uint32 `protobuf:"varint,1,opt,name=Ret" json:"Ret,omitempty"`
	RoomID uint32 `protobuf:"varint,2,opt,name=RoomID" json:"RoomID,omitempty"`
}

func (m *LEAVE_GAME_FRUITMARY_RES) Reset()                    { *m = LEAVE_GAME_FRUITMARY_RES{} }
func (m *LEAVE_GAME_FRUITMARY_RES) String() string            { return proto.CompactTextString(m) }
func (*LEAVE_GAME_FRUITMARY_RES) ProtoMessage()               {}
func (*LEAVE_GAME_FRUITMARY_RES) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{3} }

func (m *LEAVE_GAME_FRUITMARY_RES) GetRet() uint32 {
	if m != nil {
		return m.Ret
	}
	return 0
}

func (m *LEAVE_GAME_FRUITMARY_RES) GetRoomID() uint32 {
	if m != nil {
		return m.RoomID
	}
	return 0
}

// //////////////////////////////////////////// 游戏1 /////////////////////////////////////////////
// 请求开始游戏1
type START_MARY_REQ struct {
	Bet uint64 `protobuf:"varint,1,opt,name=Bet" json:"Bet,omitempty"`
}

func (m *START_MARY_REQ) Reset()                    { *m = START_MARY_REQ{} }
func (m *START_MARY_REQ) String() string            { return proto.CompactTextString(m) }
func (*START_MARY_REQ) ProtoMessage()               {}
func (*START_MARY_REQ) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{4} }

func (m *START_MARY_REQ) GetBet() uint64 {
	if m != nil {
		return m.Bet
	}
	return 0
}

type START_MARY_RES struct {
	Ret         uint64              `protobuf:"varint,1,opt,name=Ret" json:"Ret,omitempty"`
	SumOdds     int64               `protobuf:"varint,2,opt,name=SumOdds" json:"SumOdds,omitempty"`
	Results     []*FRUITMARY_Result `protobuf:"bytes,3,rep,name=Results" json:"Results,omitempty"`
	PictureList []int32             `protobuf:"varint,4,rep,packed,name=PictureList" json:"PictureList,omitempty"`
	Bonus       int64               `protobuf:"varint,5,opt,name=Bonus" json:"Bonus,omitempty"`
	Money       int64               `protobuf:"varint,6,opt,name=Money" json:"Money,omitempty"`
	FreeCount   int64               `protobuf:"varint,7,opt,name=FreeCount" json:"FreeCount,omitempty"`
	MaryCount   int64               `protobuf:"varint,8,opt,name=MaryCount" json:"MaryCount,omitempty"`
}

func (m *START_MARY_RES) Reset()                    { *m = START_MARY_RES{} }
func (m *START_MARY_RES) String() string            { return proto.CompactTextString(m) }
func (*START_MARY_RES) ProtoMessage()               {}
func (*START_MARY_RES) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{5} }

func (m *START_MARY_RES) GetRet() uint64 {
	if m != nil {
		return m.Ret
	}
	return 0
}

func (m *START_MARY_RES) GetSumOdds() int64 {
	if m != nil {
		return m.SumOdds
	}
	return 0
}

func (m *START_MARY_RES) GetResults() []*FRUITMARY_Result {
	if m != nil {
		return m.Results
	}
	return nil
}

func (m *START_MARY_RES) GetPictureList() []int32 {
	if m != nil {
		return m.PictureList
	}
	return nil
}

func (m *START_MARY_RES) GetBonus() int64 {
	if m != nil {
		return m.Bonus
	}
	return 0
}

func (m *START_MARY_RES) GetMoney() int64 {
	if m != nil {
		return m.Money
	}
	return 0
}

func (m *START_MARY_RES) GetFreeCount() int64 {
	if m != nil {
		return m.FreeCount
	}
	return 0
}

func (m *START_MARY_RES) GetMaryCount() int64 {
	if m != nil {
		return m.MaryCount
	}
	return 0
}

type FRUITMARYPosition struct {
	Px int32 `protobuf:"varint,1,opt,name=px" json:"px,omitempty"`
	Py int32 `protobuf:"varint,2,opt,name=py" json:"py,omitempty"`
}

func (m *FRUITMARYPosition) Reset()                    { *m = FRUITMARYPosition{} }
func (m *FRUITMARYPosition) String() string            { return proto.CompactTextString(m) }
func (*FRUITMARYPosition) ProtoMessage()               {}
func (*FRUITMARYPosition) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{6} }

func (m *FRUITMARYPosition) GetPx() int32 {
	if m != nil {
		return m.Px
	}
	return 0
}

func (m *FRUITMARYPosition) GetPy() int32 {
	if m != nil {
		return m.Py
	}
	return 0
}

type FRUITMARY_Result struct {
	LineId    int32                `protobuf:"varint,1,opt,name=LineId" json:"LineId,omitempty"`
	Count     int32                `protobuf:"varint,2,opt,name=Count" json:"Count,omitempty"`
	Odds      int32                `protobuf:"varint,3,opt,name=Odds" json:"Odds,omitempty"`
	Positions []*FRUITMARYPosition `protobuf:"bytes,4,rep,name=Positions" json:"Positions,omitempty"`
}

func (m *FRUITMARY_Result) Reset()                    { *m = FRUITMARY_Result{} }
func (m *FRUITMARY_Result) String() string            { return proto.CompactTextString(m) }
func (*FRUITMARY_Result) ProtoMessage()               {}
func (*FRUITMARY_Result) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{7} }

func (m *FRUITMARY_Result) GetLineId() int32 {
	if m != nil {
		return m.LineId
	}
	return 0
}

func (m *FRUITMARY_Result) GetCount() int32 {
	if m != nil {
		return m.Count
	}
	return 0
}

func (m *FRUITMARY_Result) GetOdds() int32 {
	if m != nil {
		return m.Odds
	}
	return 0
}

func (m *FRUITMARY_Result) GetPositions() []*FRUITMARYPosition {
	if m != nil {
		return m.Positions
	}
	return nil
}

// 通知更新奖金池
type UPDATE_MARY_BONUS struct {
	Bonus int64 `protobuf:"varint,1,opt,name=Bonus" json:"Bonus,omitempty"`
}

func (m *UPDATE_MARY_BONUS) Reset()                    { *m = UPDATE_MARY_BONUS{} }
func (m *UPDATE_MARY_BONUS) String() string            { return proto.CompactTextString(m) }
func (*UPDATE_MARY_BONUS) ProtoMessage()               {}
func (*UPDATE_MARY_BONUS) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{8} }

func (m *UPDATE_MARY_BONUS) GetBonus() int64 {
	if m != nil {
		return m.Bonus
	}
	return 0
}

// //////////////////////////////////////////// 游戏2 /////////////////////////////////////////////
// 请求开始游戏2
type START_MARY2_REQ struct {
}

func (m *START_MARY2_REQ) Reset()                    { *m = START_MARY2_REQ{} }
func (m *START_MARY2_REQ) String() string            { return proto.CompactTextString(m) }
func (*START_MARY2_REQ) ProtoMessage()               {}
func (*START_MARY2_REQ) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{9} }

type START_MARY2_RES struct {
	Result         []*Mary2_Result `protobuf:"bytes,1,rep,name=Result" json:"Result,omitempty"`
	MarySpareCount int32           `protobuf:"varint,2,opt,name=MarySpareCount" json:"MarySpareCount,omitempty"`
}

func (m *START_MARY2_RES) Reset()                    { *m = START_MARY2_RES{} }
func (m *START_MARY2_RES) String() string            { return proto.CompactTextString(m) }
func (*START_MARY2_RES) ProtoMessage()               {}
func (*START_MARY2_RES) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{10} }

func (m *START_MARY2_RES) GetResult() []*Mary2_Result {
	if m != nil {
		return m.Result
	}
	return nil
}

func (m *START_MARY2_RES) GetMarySpareCount() int32 {
	if m != nil {
		return m.MarySpareCount
	}
	return 0
}

type Mary2_Result struct {
	IndexId int32   `protobuf:"varint,1,opt,name=IndexId" json:"IndexId,omitempty"`
	MaryId  []int32 `protobuf:"varint,2,rep,packed,name=MaryId" json:"MaryId,omitempty"`
	Profit1 int32   `protobuf:"varint,3,opt,name=Profit1" json:"Profit1,omitempty"`
	Profit2 int32   `protobuf:"varint,4,opt,name=Profit2" json:"Profit2,omitempty"`
}

func (m *Mary2_Result) Reset()                    { *m = Mary2_Result{} }
func (m *Mary2_Result) String() string            { return proto.CompactTextString(m) }
func (*Mary2_Result) ProtoMessage()               {}
func (*Mary2_Result) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{11} }

func (m *Mary2_Result) GetIndexId() int32 {
	if m != nil {
		return m.IndexId
	}
	return 0
}

func (m *Mary2_Result) GetMaryId() []int32 {
	if m != nil {
		return m.MaryId
	}
	return nil
}

func (m *Mary2_Result) GetProfit1() int32 {
	if m != nil {
		return m.Profit1
	}
	return 0
}

func (m *Mary2_Result) GetProfit2() int32 {
	if m != nil {
		return m.Profit2
	}
	return 0
}

type NEXT_MARY_RESULT struct {
}

func (m *NEXT_MARY_RESULT) Reset()                    { *m = NEXT_MARY_RESULT{} }
func (m *NEXT_MARY_RESULT) String() string            { return proto.CompactTextString(m) }
func (*NEXT_MARY_RESULT) ProtoMessage()               {}
func (*NEXT_MARY_RESULT) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{12} }

func init() {
	proto.RegisterType((*ENTER_GAME_FRUITMARY_REQ)(nil), "protomsg.ENTER_GAME_FRUITMARY_REQ")
	proto.RegisterType((*ENTER_GAME_FRUITMARY_RES)(nil), "protomsg.ENTER_GAME_FRUITMARY_RES")
	proto.RegisterType((*ENTER_GAME_FRUITMARY_RES_FruitRatio)(nil), "protomsg.ENTER_GAME_FRUITMARY_RES.FruitRatio")
	proto.RegisterType((*LEAVE_GAME_FRUITMARY_REQ)(nil), "protomsg.LEAVE_GAME_FRUITMARY_REQ")
	proto.RegisterType((*LEAVE_GAME_FRUITMARY_RES)(nil), "protomsg.LEAVE_GAME_FRUITMARY_RES")
	proto.RegisterType((*START_MARY_REQ)(nil), "protomsg.START_MARY_REQ")
	proto.RegisterType((*START_MARY_RES)(nil), "protomsg.START_MARY_RES")
	proto.RegisterType((*FRUITMARYPosition)(nil), "protomsg.FRUITMARY_position")
	proto.RegisterType((*FRUITMARY_Result)(nil), "protomsg.FRUITMARY_Result")
	proto.RegisterType((*UPDATE_MARY_BONUS)(nil), "protomsg.UPDATE_MARY_BONUS")
	proto.RegisterType((*START_MARY2_REQ)(nil), "protomsg.START_MARY2_REQ")
	proto.RegisterType((*START_MARY2_RES)(nil), "protomsg.START_MARY2_RES")
	proto.RegisterType((*Mary2_Result)(nil), "protomsg.Mary2_Result")
	proto.RegisterType((*NEXT_MARY_RESULT)(nil), "protomsg.NEXT_MARY_RESULT")
	proto.RegisterEnum("protomsg.FRUITMARYMSG", FRUITMARYMSG_name, FRUITMARYMSG_value)
	proto.RegisterEnum("protomsg.Fruit1ID", Fruit1ID_name, Fruit1ID_value)
	proto.RegisterEnum("protomsg.Fruit2ID", Fruit2ID_name, Fruit2ID_value)
}

func init() { proto.RegisterFile("protobuf/fruitmary.proto", fileDescriptor1) }

var fileDescriptor1 = []byte{
	// 1037 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x55, 0xdb, 0x6e, 0xe3, 0x44,
	0x18, 0x5e, 0x3b, 0x71, 0x92, 0xfe, 0xe9, 0x61, 0x3a, 0x5b, 0x16, 0x6f, 0xd9, 0x8b, 0xe0, 0x0b,
	0x14, 0x2a, 0xd1, 0x55, 0xdd, 0x5e, 0xa0, 0x15, 0x37, 0x39, 0xb5, 0x8a, 0x68, 0xda, 0xec, 0x4c,
	0xc2, 0x02, 0x37, 0x96, 0xdb, 0x4c, 0x8b, 0xb5, 0x89, 0x1d, 0xd9, 0x0e, 0x34, 0x0f, 0x81, 0xc4,
	0x03, 0xe4, 0x86, 0x83, 0x10, 0xe7, 0x37, 0xe1, 0x19, 0x78, 0x11, 0x2e, 0xd0, 0x1c, 0x9c, 0x71,
	0xc2, 0x46, 0xe2, 0x82, 0xab, 0xcc, 0xf7, 0xcd, 0x3f, 0xff, 0xe1, 0xf3, 0x37, 0x13, 0xb0, 0xa7,
	0x71, 0x94, 0x46, 0x37, 0xb3, 0xbb, 0xe7, 0x77, 0xf1, 0x2c, 0x48, 0x27, 0x7e, 0x3c, 0x3f, 0x16,
	0x14, 0xae, 0x88, 0x9f, 0x49, 0x72, 0xef, 0xf4, 0xc1, 0xee, 0x5c, 0x0d, 0x3a, 0xc4, 0xbb, 0x68,
	0xf4, 0x3a, 0xde, 0x39, 0x19, 0x76, 0x07, 0xbd, 0x06, 0xf9, 0xcc, 0x23, 0x9d, 0x97, 0xf8, 0x19,
	0x6c, 0x35, 0x6e, 0x6f, 0xa3, 0x59, 0x98, 0x76, 0xdb, 0xb6, 0x51, 0x33, 0xea, 0x3b, 0x44, 0x13,
	0xf8, 0x09, 0x94, 0x48, 0x14, 0x4d, 0xba, 0x6d, 0xdb, 0x14, 0x5b, 0x0a, 0x39, 0x3f, 0x16, 0x37,
	0xa6, 0xa4, 0xb9, 0x43, 0x46, 0xfe, 0x10, 0xe7, 0x9b, 0x7e, 0x12, 0xdc, 0x26, 0x22, 0x59, 0x81,
	0x28, 0x84, 0x0f, 0xc0, 0x6a, 0x46, 0xe1, 0x2c, 0xb1, 0x0b, 0x82, 0x96, 0x00, 0xdb, 0x50, 0xbe,
	0xf4, 0x93, 0xb4, 0xc9, 0x52, 0xbb, 0x28, 0xf8, 0x0c, 0x62, 0x0c, 0xc5, 0x26, 0x4b, 0x13, 0xdb,
	0xaa, 0x15, 0xea, 0x45, 0x22, 0xd6, 0xb8, 0x05, 0x16, 0xf1, 0xd3, 0x20, 0xb2, 0x4b, 0xb5, 0x42,
	0xbd, 0xea, 0x7e, 0x70, 0x9c, 0x0d, 0x7f, 0xbc, 0xa9, 0xcd, 0x63, 0x11, 0xdf, 0x09, 0xd3, 0x78,
	0x4e, 0xe4, 0x59, 0xfc, 0x11, 0x6c, 0xf7, 0xfc, 0x78, 0xee, 0x7a, 0x84, 0x25, 0xb3, 0x71, 0x6a,
	0x97, 0x6b, 0x46, 0xbd, 0xea, 0x3e, 0xd5, 0xb9, 0xe8, 0xa0, 0x41, 0x06, 0x1e, 0xcf, 0xe0, 0xf2,
	0x14, 0xa4, 0x2a, 0xc2, 0x65, 0x34, 0x3e, 0x84, 0xca, 0x39, 0x63, 0x2d, 0xae, 0x9c, 0x5d, 0xa9,
	0x19, 0x75, 0x8b, 0x2c, 0xf1, 0xe1, 0xd7, 0x06, 0xc0, 0x39, 0xff, 0x3e, 0xb2, 0x90, 0x03, 0xa6,
	0x52, 0x67, 0xd7, 0xc5, 0x3a, 0xbd, 0x88, 0x70, 0xbb, 0x6d, 0x62, 0x4a, 0xb5, 0x68, 0x10, 0xde,
	0x8f, 0x99, 0x50, 0xcb, 0x22, 0x0a, 0xe1, 0xb7, 0xa0, 0x44, 0xfd, 0x09, 0xf3, 0x5c, 0x21, 0x97,
	0x45, 0x2c, 0x8e, 0xdc, 0x25, 0x7d, 0x2a, 0xd4, 0x52, 0xf4, 0xe9, 0x92, 0x3e, 0xb3, 0x2d, 0x4d,
	0x9f, 0x1d, 0xde, 0x03, 0xe8, 0xf1, 0x31, 0x82, 0xc2, 0x6b, 0x36, 0x17, 0xfd, 0x58, 0x84, 0x2f,
	0xb9, 0x9c, 0x5f, 0xfa, 0xe3, 0x99, 0xac, 0xfd, 0xdf, 0xe4, 0xd4, 0xe3, 0x11, 0x79, 0xf6, 0x85,
	0xf9, 0xa1, 0xc1, 0xad, 0x77, 0xd9, 0x69, 0x7c, 0xd2, 0xf9, 0xff, 0xac, 0xd7, 0xde, 0x98, 0x91,
	0xf2, 0x41, 0x08, 0x4b, 0x55, 0x2e, 0xbe, 0xdc, 0x98, 0xc5, 0x81, 0x5d, 0xfd, 0x31, 0x45, 0x37,
	0x08, 0x0a, 0x4d, 0x75, 0xb6, 0x48, 0xf8, 0xd2, 0xf9, 0xdb, 0x58, 0x0b, 0x5a, 0x29, 0x50, 0x94,
	0x05, 0x6c, 0x28, 0xd3, 0xd9, 0xe4, 0x7a, 0x34, 0xca, 0x5c, 0x9d, 0x41, 0x7c, 0x06, 0x65, 0xe9,
	0x0c, 0x6e, 0x6c, 0x6e, 0xca, 0xc3, 0xdc, 0x97, 0xd6, 0x6d, 0x8b, 0x10, 0x92, 0x85, 0xe2, 0x1a,
	0x54, 0xfb, 0xc1, 0x6d, 0x3a, 0x8b, 0xd9, 0x65, 0x90, 0x70, 0xeb, 0x17, 0xea, 0x16, 0xc9, 0x53,
	0xfa, 0xba, 0x58, 0xf9, 0xeb, 0x72, 0x00, 0x56, 0x2f, 0x0a, 0xd9, 0xdc, 0x2e, 0x49, 0x56, 0x00,
	0x2e, 0xf1, 0x79, 0x9c, 0x99, 0xb2, 0x2c, 0x76, 0x34, 0xc1, 0x77, 0xb9, 0x81, 0xb5, 0x65, 0x0b,
	0x44, 0x13, 0xce, 0x19, 0x60, 0xdd, 0xe6, 0x34, 0x4a, 0x82, 0x34, 0x88, 0x42, 0xbc, 0x0b, 0xe6,
	0xf4, 0x41, 0x59, 0xc5, 0x9c, 0x3e, 0x08, 0x3c, 0x57, 0x16, 0x35, 0xa7, 0x73, 0xe7, 0x1b, 0x03,
	0xd0, 0xfa, 0x74, 0xfc, 0x2b, 0x5c, 0x06, 0x21, 0xeb, 0x8e, 0xd4, 0x41, 0x85, 0x78, 0xd3, 0xb2,
	0xb8, 0x3c, 0x2f, 0x01, 0xbf, 0xdf, 0x42, 0x4f, 0xe9, 0x6f, 0xb1, 0xc6, 0x2f, 0x60, 0xab, 0xaf,
	0x5a, 0x48, 0x84, 0x28, 0x55, 0xf7, 0xd9, 0x9b, 0xe4, 0xcc, 0xfa, 0x24, 0x3a, 0xdc, 0x79, 0x1f,
	0xf6, 0x87, 0xfd, 0x76, 0x63, 0xd0, 0x91, 0xdf, 0xb1, 0x79, 0x7d, 0x35, 0xa4, 0x5a, 0x45, 0x23,
	0xa7, 0xa2, 0xb3, 0x0f, 0x7b, 0xab, 0x77, 0xfc, 0xa5, 0x13, 0xac, 0x53, 0x14, 0x1f, 0x43, 0x49,
	0xbd, 0x10, 0x86, 0xe8, 0xe4, 0x89, 0xee, 0x24, 0xff, 0x7e, 0x10, 0x15, 0x85, 0xdf, 0x83, 0x5d,
	0xce, 0xd3, 0xa9, 0x1f, 0xb3, 0xfc, 0xbc, 0x6b, 0xac, 0x93, 0xae, 0xbe, 0x3f, 0xdc, 0x5b, 0xdd,
	0x70, 0xc4, 0x1e, 0x96, 0xba, 0x65, 0x90, 0x0b, 0xca, 0x23, 0xbb, 0x23, 0xdb, 0x14, 0x06, 0x51,
	0x88, 0x9f, 0xe8, 0xc7, 0xd1, 0x5d, 0x90, 0x9e, 0x28, 0xf5, 0x32, 0xa8, 0x77, 0x5c, 0xf5, 0x40,
	0x64, 0xd0, 0xc1, 0x80, 0xae, 0x3a, 0x9f, 0x6a, 0x93, 0x0f, 0x2f, 0x07, 0x47, 0x7f, 0x9a, 0xb0,
	0xbd, 0x14, 0xb5, 0x47, 0x2f, 0xf0, 0x01, 0xa0, 0xe1, 0xd5, 0xc7, 0x57, 0xd7, 0xaf, 0xf4, 0x8d,
	0x43, 0x8f, 0xf0, 0xbb, 0xf0, 0x4e, 0x8b, 0x7a, 0x9b, 0xfe, 0x5b, 0xd0, 0xb7, 0x0b, 0x83, 0x87,
	0xd0, 0xd6, 0xa6, 0x10, 0x8a, 0xbe, 0x93, 0x21, 0x2d, 0xea, 0x6d, 0x7a, 0x26, 0xd0, 0xf7, 0xcb,
	0x2c, 0x9b, 0xee, 0x3d, 0xfa, 0x61, 0x61, 0xe0, 0xb7, 0x61, 0xbf, 0x45, 0xbd, 0xd5, 0x4b, 0x8d,
	0x7e, 0x92, 0x1b, 0xb4, 0xb5, 0xba, 0x41, 0xd1, 0xcf, 0x0b, 0x03, 0x1f, 0xc2, 0x01, 0x6d, 0x79,
	0xff, 0xb2, 0x06, 0xfa, 0x65, 0x61, 0x60, 0x1b, 0xf0, 0x4a, 0x36, 0xe1, 0x05, 0xf4, 0xab, 0xdc,
	0x59, 0x49, 0x27, 0x2c, 0x81, 0x7e, 0x5b, 0x18, 0xf8, 0x29, 0x3c, 0x6e, 0x51, 0x6f, 0x5d, 0x4b,
	0xf4, 0xfb, 0xc2, 0x38, 0xfa, 0xcb, 0x80, 0x8a, 0x78, 0x20, 0x4f, 0xba, 0x6d, 0x8c, 0x60, 0x5b,
	0xae, 0x87, 0xe1, 0xeb, 0x30, 0xfa, 0x0a, 0x3d, 0xc2, 0xbb, 0xea, 0xdf, 0xe1, 0xe4, 0x55, 0x30,
	0x1e, 0x21, 0x03, 0xef, 0x41, 0x55, 0x62, 0xe1, 0x4a, 0x64, 0xe2, 0x7d, 0xd8, 0x91, 0x04, 0xbd,
	0xf5, 0xd3, 0x94, 0xc5, 0xa8, 0x80, 0x77, 0xf8, 0xd5, 0x16, 0x31, 0x7e, 0x8c, 0x8a, 0x3a, 0x69,
	0xeb, 0x0b, 0x16, 0xc7, 0x73, 0x64, 0xe9, 0xa4, 0x4d, 0x36, 0x1e, 0xa3, 0x12, 0x7e, 0x0c, 0x7b,
	0x12, 0xf7, 0x83, 0x90, 0xf9, 0xd3, 0xe9, 0x98, 0xa1, 0xb2, 0x0e, 0xba, 0x88, 0xfd, 0x29, 0xaa,
	0xe8, 0xca, 0x3d, 0x3f, 0xbc, 0x8f, 0xd0, 0x16, 0xff, 0xf0, 0xaa, 0x35, 0x3f, 0x65, 0xf1, 0x84,
	0x8d, 0xa3, 0x10, 0x81, 0xae, 0xd6, 0xf4, 0x43, 0x3f, 0xf4, 0x51, 0xf5, 0xe8, 0x8f, 0x6c, 0x42,
	0x37, 0x37, 0xa1, 0xbb, 0x9c, 0x30, 0x4b, 0xe3, 0xe6, 0xd2, 0x18, 0xcb, 0xea, 0xae, 0xa8, 0x6e,
	0x2e, 0xab, 0xbb, 0xb2, 0x7a, 0x41, 0x27, 0x52, 0x53, 0xe9, 0x39, 0x5d, 0x55, 0xd9, 0xd2, 0xcc,
	0x75, 0xec, 0x87, 0xf7, 0x2c, 0x37, 0xa9, 0xfb, 0xa6, 0x49, 0xdd, 0x66, 0x34, 0xb9, 0x41, 0x95,
	0xe6, 0xde, 0xe7, 0x3b, 0x71, 0x14, 0xa5, 0xcf, 0xb3, 0xab, 0x7b, 0x53, 0x12, 0xab, 0xd3, 0x7f,
	0x02, 0x00, 0x00, 0xff, 0xff, 0x9a, 0xab, 0xf6, 0xbd, 0x52, 0x09, 0x00, 0x00,
}
