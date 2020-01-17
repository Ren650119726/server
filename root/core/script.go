package core

import (
	"github.com/yuin/gopher-lua"
	"root/common"
	"root/common/tools"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"strconv"
	"unicode/utf8"
)

type ELuaType int

// 服务器类别定义
const (
	LUA_INT8   ELuaType = 1
	LUA_INT16  ELuaType = 2
	LUA_INT32  ELuaType = 3
	LUA_UINT8  ELuaType = 4
	LUA_UINT16 ELuaType = 5
	LUA_UINT32 ELuaType = 6
	LUA_STRING ELuaType = 7
	LUA_FLOAT  ELuaType = 8
	LUA_TABLE  ELuaType = 9
	LUA_DOUBLE ELuaType = 10
	LUA_INT64  ELuaType = 11
)

var typeStringELuaType = [...]string{
	LUA_INT8:   "int8",
	LUA_INT16:  "int16",
	LUA_INT32:  "int32",
	LUA_UINT8:  "uint8",
	LUA_UINT16: "uint16",
	LUA_UINT32: "uint32",
	LUA_STRING: "string",
	LUA_FLOAT:  "float",
	LUA_TABLE:  "table",
	LUA_DOUBLE: "double",
	LUA_INT64:  "int64",
}

func (e ELuaType) String() string {
	return typeStringELuaType[e]
}

func (e ELuaType) Int32() int32 {
	return int32(e)
}

var (
	Global_Lua *lua.LState
)

func InitScript(path string) {
	Global_Lua = lua.NewState()
	Global_Lua.OpenLibs()
	Global_Lua.Register("c_ParserPacket", ParserPacketForLuaDefine)
	Global_Lua.Register("c_SendMsgToClient", SendMsgToClient)
	Global_Lua.Register("c_SendMsgToDB", SendMsgToDB)
	Global_Lua.Register("c_SendMsgToHall", SendMsgToHall)
	Global_Lua.Register("c_GetTimeToMillisecond", GetTimeToMillisecond)
	Global_Lua.Register("c_GetClientIP", GetClientIP)
	Global_Lua.Register("c_Loglog", Debuglog)
	Global_Lua.Register("c_Infolog", Infolog)
	Global_Lua.Register("c_Errorlog", Errorlog)
	Global_Lua.Register("c_ServerCloseSocket", ServerCloseSocket)
	Global_Lua.Register("c_Utf8_len", Utf8_len)
	Global_Lua.Register("c_RegistryTimer", RegistryTimer)
	Global_Lua.Register("c_RegistryTimerFun", RegistryTimerFun)
	Global_Lua.Register("c_CancelTimer", CancelTimer)

	Global_Lua.Register("c_StringToTime", StringToTime)
	Global_Lua.Register("c_GetLocalIP", GetLocalIP)
	Global_Lua.Register("c_band", Band)
	Global_Lua.Register("c_bor", Bor)
	Global_Lua.Register("c_serverID", ServerID)
	Global_Lua.Register("c_Sleep", Sleep)
	Global_Lua.Register("c_MD5", CalcMD5)
	Global_Lua.Register("c_CloseAllClientSocket", CloseAllClientSocket)
	Global_Lua.Register("c_CalcRandomReturnKey", CalcRandomReturnKey)
	Global_Lua.SetGlobal("INT8", lua.LNumber(LUA_INT8))
	Global_Lua.SetGlobal("INT16", lua.LNumber(LUA_INT16))
	Global_Lua.SetGlobal("INT32", lua.LNumber(LUA_INT32))
	Global_Lua.SetGlobal("UINT8", lua.LNumber(LUA_UINT8))
	Global_Lua.SetGlobal("UINT16", lua.LNumber(LUA_UINT16))
	Global_Lua.SetGlobal("UINT32", lua.LNumber(LUA_UINT32))
	Global_Lua.SetGlobal("STRING", lua.LNumber(LUA_STRING))
	Global_Lua.SetGlobal("FLOAT", lua.LNumber(LUA_FLOAT))
	Global_Lua.SetGlobal("TABLE", lua.LNumber(LUA_TABLE))
	Global_Lua.SetGlobal("SERVER_ID", lua.LNumber(SID))
	Global_Lua.SetGlobal("DOUBLE", lua.LNumber(LUA_DOUBLE))
	Global_Lua.SetGlobal("INT64", lua.LNumber(LUA_INT64))

	err := Global_Lua.DoFile(path + "init.lua")
	if err != nil {
		log.Debug(err.Error())
	} else {
		if err := Global_Lua.CallByParam(lua.P{
			Fn:      Global_Lua.GetGlobal("File_Load"),
			NRet:    0,
			Protect: true,
		}, lua.LString(path)); err != nil {
			log.Errorf("%v", err.Error())
			return
		}
	}

	Cmd.Regist("lua", cmd_Load, true)
	Cmd.Regist("cmd", lua_cmd_process, true)
}

func cmd_Load(params []string) {
	if len(params) != 1 {
		log.Errorf("参数个数不为1 错误：%v", len(params))
		return
	}
	param := params[0]
	if param == "all" {
		err := Global_Lua.DoFile(ScriptDir + "/init.lua")
		if err != nil {
			log.Debug(err.Error())
		} else {
			Global_Lua.SetTop(0)
			if err := Global_Lua.CallByParam(lua.P{
				Fn:      Global_Lua.GetGlobal("File_Load"),
				NRet:    0,
				Protect: true,
			}, lua.LString(ScriptDir)); err != nil {
				log.Errorf("%v", err.Error())
				return
			}
		}
	} else {
		err := Global_Lua.DoFile(ScriptDir + "/" + param)
		if err != nil {
			log.Debug(err.Error())
		}
		Global_Lua.SetTop(0)
	}

}

func lua_cmd_process(params []string) {
	if len(params) < 1 {
		log.Errorf("参数个数不为1 错误：%v", len(params))
		return
	}
	var (
		str_cmd string
		inum1   int
		inum2   int
		inum3   int
	)
	if len(params) >= 1 {
		str_cmd = params[0]
	}

	if len(params) >= 2 {
		inum1, _ = strconv.Atoi(params[1])
	}

	if len(params) >= 3 {
		inum2, _ = strconv.Atoi(params[2])
	}

	if len(params) >= 4 {
		inum3, _ = strconv.Atoi(params[3])
	}

	if err := Global_Lua.CallByParam(
		lua.P{
			Fn:      Global_Lua.GetGlobal("lua_TestMsg"),
			NRet:    0,
			Protect: true,
		},
		lua.LString(str_cmd),
		lua.LNumber(inum1),
		lua.LNumber(inum2),
		lua.LNumber(inum3),
	); err != nil {
		log.Errorf("%v", err.Error())
	}

}

func Error(L *lua.LState) int {
	var (
		nProtocol uint16
		packet    = packet.NewPacket(nil)
	)

	nProtocol = uint16(L.ToInt(-1))
	L.Pop(1)

	packet.SetMsgID(nProtocol)
	if msgData := L.CheckTable(1); msgData != nil {
		// 解析数据包到packet中
		ParserTalbeToPacket(packet, msgData)

		CoreSend(0, common.EActorType_CONNECT_DB.Int32(), packet.GetData(), 0)
	} else {
		log.Error("SendMsgToDB error")
	}
	return 0
}

func MsgProcess(msgid uint16, msg []byte, session int64) bool {
	if Global_Lua == nil {
		return false
	}
	userData := Global_Lua.NewUserData()
	userData.Value = packet.NewPacket(msg)
	if err := Global_Lua.CallByParam(
		lua.P{
			Fn:      Global_Lua.GetGlobal("lua_MsgHandle"),
			NRet:    1,
			Protect: true,
		},
		lua.LNumber(msgid),
		userData,
		lua.LNumber(session),
	); err != nil {
		log.Errorf("msgId:%v error:%v", msgid, err.Error())
	}

	Global_Lua.Pop(1)
	return true
}

//********************************************************************
//函数功能: 解析消息包，并封装成table
//lua第一参数: 解析顺序的table
//lua第二参数: cPacket 消息包指针
//返回说明:
//备注说明: 标准的c->lua 接口函数，返回值lua返回值个数(栈个数)
//********************************************************************
func ParserPacketForLuaDefine(L *lua.LState) int {
	var (
		rec  func(t *lua.LTable)
		rec2 func(t *lua.LTable, size, begin, end int) int
	)

	PacketVec := []int{}
	msg := L.ToUserData(2) // 获得第二个参数

	packet := msg.Value.(packet.IPacket)

	L.Pop(1)
	table := L.ToTable(1)

	rec = func(t *lua.LTable) {
		t.ForEach(func(k, v lua.LValue) {
			if v.Type() == lua.LTNumber {
				PacketVec = append(PacketVec, int(v.(lua.LNumber)))
			} else if v.Type() == lua.LTTable {
				t := v.(*lua.LTable)
				PacketVec = append(PacketVec, t.Len())
				PacketVec = append(PacketVec, int(LUA_TABLE))
				rec(t)
			}
		})
	}

	rec(table)
	L.Pop(1)

	retTable := L.NewTable()

	rec2 = func(t *lua.LTable, size, begin, end int) int {
		for i := 0; i < size; i++ {
			arrayTable := L.NewTable()
			for i := begin; i <= end; i++ {
				if i+2 < end && PacketVec[i+2] == int(LUA_TABLE) {
					attributeTable := L.NewTable()
					count := PacketVec[i+1]
					nextSize := packet.ReadUInt16()

					i = rec2(attributeTable, int(nextSize), i+3, i+2+count)
					arrayTable.Append(attributeTable)
				} else {
					switch ELuaType(PacketVec[i]) {
					case LUA_INT8:

						arrayTable.Append(lua.LNumber(packet.ReadInt8()))
					case LUA_INT16:
						arrayTable.Append(lua.LNumber(packet.ReadInt16()))
					case LUA_INT32:
						arrayTable.Append(lua.LNumber(packet.ReadInt32()))
					case LUA_UINT8:
						arrayTable.Append(lua.LNumber(packet.ReadUInt8()))
					case LUA_UINT16:
						arrayTable.Append(lua.LNumber(packet.ReadUInt16()))
					case LUA_UINT32:
						arrayTable.Append(lua.LNumber(packet.ReadUInt32()))
					case LUA_INT64:
						arrayTable.Append(lua.LNumber(packet.ReadInt64()))
					case LUA_FLOAT:
						arrayTable.Append(lua.LNumber(packet.ReadFloat32()))
					case LUA_STRING:
						arrayTable.Append(lua.LString(packet.ReadString()))
					case LUA_DOUBLE:
						arrayTable.Append(lua.LNumber(packet.ReadFloat64()))

					default:
						continue
					}
				}
			}

			t.Append(arrayTable)
		}
		return end
	}

	rec2(retTable, 1, 0, len(PacketVec)-1)

	L.Push(retTable.RawGetInt(1))
	return 1
}

func ParserTalbeToPacket(pack packet.IPacket, table *lua.LTable) {
	table.ForEach(func(k, v lua.LValue) {
		if v.Type() != lua.LTTable {
			log.Error(" 错误的类型  v.type :%v", v.Type())
			return
		}
		field := v.(*lua.LTable)

		t := field.RawGetInt(1)
		if t.Type() != lua.LTNumber {
			log.Warnf(" 错误的类型  t.type :%v", t.Type())
			for i := 1; i <= 3; i++ {
				log.Warnf("stack trace:%v", Global_Lua.Where(i))
			}
			return
		}

		tt := t.(lua.LNumber)
		vv := field.RawGetInt(2)
		switch ELuaType(tt) {
		case LUA_INT8:
			if vv.Type() != lua.LTNumber {
				log.Errorf("table 里的值不是LUA_INT8   %d", vv.Type())
				return
			}
			pack.WriteInt8(int8(vv.(lua.LNumber)))

		case LUA_INT16:
			if vv.Type() != lua.LTNumber {
				log.Errorf("table LUA_INT16   %d", vv.Type())
				return
			}
			pack.WriteInt16(int16(vv.(lua.LNumber)))
		case LUA_INT32:
			if vv.Type() != lua.LTNumber {
				log.Errorf("table LUA_INT32   %d", vv.Type())
				return
			}
			pack.WriteInt32(int32(vv.(lua.LNumber)))
		case LUA_UINT8:
			if vv.Type() != lua.LTNumber {
				log.Errorf("table LUA_UINT8   %d", vv.Type())
				return
			}
			pack.WriteUInt8(uint8(vv.(lua.LNumber)))
		case LUA_UINT16:
			if vv.Type() != lua.LTNumber {
				log.Errorf("table LUA_UINT16   %d", vv.Type())
				return
			}
			pack.WriteUInt16(uint16(vv.(lua.LNumber)))
		case LUA_UINT32:
			if vv.Type() != lua.LTNumber {
				log.Errorf("table LUA_UINT32   %d", vv.Type())
				return
			}
			pack.WriteUInt32(uint32(vv.(lua.LNumber)))
		case LUA_INT64:
			if vv.Type() != lua.LTNumber {
				log.Errorf("table LUA_INT64   %d", vv.Type())
				return
			}
			pack.WriteInt64(int64(vv.(lua.LNumber)))
		case LUA_FLOAT:
			if vv.Type() != lua.LTNumber {
				log.Errorf("table LUA_FLOAT   %d", vv.Type())
				return
			}
			pack.WriteFloat32(float32(vv.(lua.LNumber)))
		case LUA_DOUBLE:
			if vv.Type() != lua.LTNumber {
				log.Errorf("table LUA_FLOAT64   %d", vv.Type())
				return
			}
			pack.WriteFloat64(float64(vv.(lua.LNumber)))
		case LUA_STRING:
			if vv.Type() != lua.LTString {
				log.Errorf("table LUA_STRING   %d", vv.Type())
				return
			}
			pack.WriteString(string(vv.(lua.LString)))
		case LUA_TABLE:
			if vv.Type() != lua.LTTable {
				log.Errorf("table LUA_TABLE   %d", vv.Type())
				return
			}
			nexttable := vv.(*lua.LTable)
			len := uint16(nexttable.Len())
			pack.WriteUInt16(len)
			if len > 0 {
				nexttable.ForEach(func(k, v lua.LValue) {
					if v.Type() != lua.LTTable {
						log.Errorf(" error :%v", v.Type())
						return
					}
					ParserTalbeToPacket(pack, v.(*lua.LTable))
				})

			}
		}
	})
}

//********************************************************************
//函数功能: 解析lua数据table，封装成数据包并且发送Client
//lua第一参数: table    发送数据table
//lua第二参数: Procotol 协议编号
//lua第三参数: sockid

//返回说明:
//备注说明: 标准的c->lua 接口函数，返回值lua返回值个数(栈个数)
//********************************************************************
func SendMsgToClient(L *lua.LState) int {
	var (
		nProtocol uint16
		sessionID int64
		packet    = packet.NewPacket(nil)
	)

	sessionID = int64(L.ToInt(-1))
	if sessionID == 0 {
		L.Pop(3)
		return 0
	}

	nProtocol = uint16(L.ToInt(-2))
	L.Pop(2)

	packet.SetMsgID(nProtocol)
	if msgData := L.CheckTable(1); msgData != nil {
		// 解析数据包到packet中
		ParserTalbeToPacket(packet, msgData)

		CoreSend(0, common.EActorType_SERVER.Int32(), packet.GetData(), sessionID)
	} else {
		log.Error("SendMsgToClient error")
	}
	return 0
}

//********************************************************************
//函数功能: 解析lua数据table，封装成数据包并且发送Server
//lua第一参数: table    发送数据table
//lua第二参数: Procotol 协议编号

//返回说明:
//备注说明: 标准的c->lua 接口函数，返回值lua返回值个数(栈个数)
//********************************************************************
func SendMsgToDB(L *lua.LState) int {
	var (
		nProtocol uint16
		packet    = packet.NewPacket(nil)
	)

	nProtocol = uint16(L.ToInt(-1))
	if nProtocol == 0 {
		log.Warnf("未组装msgID")
		for i := 1; i <= 3; i++ {
			log.Warnf("stack trace:%v", L.Where(i))
		}

	}
	L.Pop(1)

	packet.SetMsgID(nProtocol)
	if msgData := L.CheckTable(1); msgData != nil {
		// 解析数据包到packet中
		ParserTalbeToPacket(packet, msgData)

		CoreSend(0, common.EActorType_CONNECT_DB.Int32(), packet.GetData(), 0)
	} else {
		log.Error("SendMsgToDB error")
	}
	return 0
}

//********************************************************************
//函数功能: 解析lua数据table，封装成数据包并且发送Server
//lua第一参数: table    发送数据table
//lua第二参数: Procotol 协议编号

//返回说明:
//备注说明: 标准的c->lua 接口函数，返回值lua返回值个数(栈个数)
//********************************************************************
func SendMsgToHall(L *lua.LState) int {
	var (
		nProtocol uint16
		packet    = packet.NewPacket(nil)
	)
	nProtocol = uint16(L.ToInt(-1))
	if nProtocol == 0 {
		log.Warnf("未组装msgID")
		for i := 1; i <= 3; i++ {
			log.Warnf("stack trace:%v", L.Where(i))
		}

	}
	L.Pop(1)

	packet.SetMsgID(nProtocol)

	if msgData := L.CheckTable(1); msgData != nil {
		// 解析数据包到packet中
		ParserTalbeToPacket(packet, msgData)

		CoreSend(0, common.EActorType_CONNECT_HALL.Int32(), packet.GetData(), 0)
	} else {
		log.Error("SendMsgToHall error")
	}

	return 0
}

func GetTimeToMillisecond(L *lua.LState) int {
	ms := utils.MilliSecondTime()
	L.Push(lua.LNumber(ms))
	return 1
}

func GetClientIP(L *lua.LState) int {
	nSessionId := L.ToInt(-1)
	L.Pop(1)

	ip := GetRemoteIP(int64(nSessionId))
	L.Push(lua.LString(ip))
	return 1
}

func Infolog(L *lua.LState) int {

	logString := L.ToString(1)
	L.Pop(1)

	log.Info(colorized.Yellow(logString))
	return 0
}
func Debuglog(L *lua.LState) int {

	logString := L.ToString(1)
	L.Pop(1)

	log.Debug(colorized.Yellow(logString))
	return 0
}

func Errorlog(L *lua.LState) int {

	logString := L.ToString(1)
	L.Pop(1)

	log.Warn(colorized.Green(logString))
	return 0
}

func ServerCloseSocket(L *lua.LState) int {

	return 0
}

func Utf8_len(L *lua.LState) int {
	str := L.ToString(1)
	L.Pop(1)
	count := utf8.RuneCountInString(str)
	L.Push(lua.LNumber(count))
	return 1
}

func RegistryTimer(L *lua.LState) int {
	var (
		nInterval = L.ToInt(-1)
		iCount    = L.ToInt(-2)
		nProtocol = L.ToInt(-3)
	)
	L.Pop(3)
	pack := packet.NewPacket(nil)

	if v := L.Get(1); v.Type() == lua.LTTable {
		ParserTalbeToPacket(pack, v.(*lua.LTable))
		L.Pop(1)
	}
	pack.SetMsgID(uint16(nProtocol))
	if iCount == 0 {
		iCount = 1
	}

	actor := GetActor(common.EActorType_MAIN.Int32())
	if actor == nil {
		log.Error("找不到启动线程")
		return 0
	}

	timerId := actor.AddTimer(int64(nInterval), int32(iCount), func(dt int64) {
		CoreSend(common.EActorType_MAIN.Int32(), common.EActorType_MAIN.Int32(), pack.GetData(), 0)
	})

	L.Push(lua.LNumber(timerId))
	return 1
}

func RegistryTimerFun(L *lua.LState) int {
	var (
		nInterval = L.ToInt(-1)
		iCount    = L.ToInt(-2)
		fun       = L.ToFunction(-3)
	)
	L.Pop(3)
	if iCount == 0 {
		iCount = 1
	}

	actor := GetActor(common.EActorType_MAIN.Int32())
	if actor == nil {
		log.Error("找不到启动线程")
		return 0
	}

	timerId := actor.AddTimer(int64(nInterval), int32(iCount), func(dt int64) {
		if err := Global_Lua.CallByParam(lua.P{
			Fn:      fun,
			NRet:    0,
			Protect: true,
		}); err != nil {
			log.Errorf("%v", err.Error())
			return
		}
	})

	L.Push(lua.LNumber(timerId))
	return 1
}

func CancelTimer(L *lua.LState) int {
	var (
		timerId = L.ToInt(1)
	)
	L.Pop(1)

	actor := GetActor(common.EActorType_MAIN.Int32())
	if actor == nil {
		log.Error("找不到启动线程")
		return 0
	}

	actor.CancelTimer(int64(timerId))

	return 0
}

func StringToTime(L *lua.LState) int {
	var (
		strTime = L.ToString(1)
	)
	L.Pop(1)

	rettime := utils.String2UnixStamp(strTime)

	L.Push(lua.LNumber(rettime / 1000))
	return 1
}

func GetLocalIP(L *lua.LState) int {
	strIP := utils.GetLocalIP()

	L.Push(lua.LString(strIP))
	return 1
}

func Band(L *lua.LState) int {
	a := L.ToInt(1)
	b := L.ToInt(2)

	L.Push(lua.LNumber(a & b))
	return 1
}

func Bor(L *lua.LState) int {
	a := L.ToInt(1)
	b := L.ToInt(2)

	L.Push(lua.LNumber(a | b))
	return 1
}
func ServerID(L *lua.LState) int {
	id, _ := strconv.Atoi(Appname)
	L.Push(lua.LNumber(id))
	return 1
}
func Sleep(L *lua.LState) int {
	_ = L.ToInt(1)
	return 0
}

func CalcMD5(L *lua.LState) int {
	strContext := L.ToString(1)
	L.Pop(1)

	strMD5 := tools.MD5(strContext)
	L.Push(lua.LString(strMD5))
	return 1
}

func CloseAllClientSocket(L *lua.LState) int {
	return 0
}

func CalcRandomReturnKey(L *lua.LState) int {
	param := L.ToString(1)
	def := L.ToInt(2)
	L.Pop(2)

	val := utils.CalcRandomReturnKey(param, def)
	L.Push(lua.LNumber(val))
	return 1
}
