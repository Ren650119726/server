package logic

import (
	"github.com/golang/protobuf/proto"
	"root/common"
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/utils"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/hall/account"
	"root/server/hall/event"
	"root/server/hall/send_tools"
	"root/server/hall/types"
	"time"
)

var RobotMgr = NewRobotMgr()

type (
	robotMgr struct {
		Room_NumOfRobot_Limit map[uint32][]*time_frame // key roomid
		NameTableIndex        int
	}
	// 时段结构
	time_frame struct {
		StartTime int64
		EndTime   int64
		Week      [7]bool
		Num       uint32
		MoneyMin  int
		MoneyMax  int
	}
)

func NewRobotMgr() *robotMgr {
	obj := &robotMgr{}
	event.Dispatcher.AddEventListener(event.EventType_RoomUpdate, obj)
	return obj
}

// 加载配置
func (self *robotMgr) Load() {
	self.Room_NumOfRobot_Limit = make(map[uint32][]*time_frame)
	robotTime_conf := config.Get_config("robot_time")
	DateString := time.Now().Format(utils.STD_TIMEFORMAT2)
	for id, timeFrame := range robotTime_conf {
		startTime := config.Get_JsonDataString(timeFrame, id, "StartTime")
		endTime := config.Get_JsonDataString(timeFrame, id, "EndTime")
		Week := config.Get_JsonDataString(timeFrame, id, "Week")
		num := config.Get_JsonDataInt(timeFrame, id, "Num")
		roomid := config.Get_JsonDataInt(timeFrame, id, "RoomID")
		weekbool := [7]bool{}
		weekarr := utils.SplitConf2ArrInt32(Week, ",")
		for _, v := range weekarr {
			if v == 7 {
				v = 0
			}
			weekbool[v] = true
		}
		frame := &time_frame{StartTime: utils.String2UnixStamp(DateString + " " + startTime), EndTime: utils.String2UnixStamp(DateString + " " + endTime), Week: weekbool, Num: uint32(num)}
		if self.Room_NumOfRobot_Limit[uint32(roomid)] == nil {
			self.Room_NumOfRobot_Limit[uint32(roomid)] = make([]*time_frame, 0)
		}
		moneystr := config.Get_JsonDataString(timeFrame, id, "Gold")
		str_arr := utils.SplitConf2ArrInt32(moneystr, ",")
		frame.MoneyMin = int(str_arr[0])
		frame.MoneyMax = int(str_arr[1])
		self.Room_NumOfRobot_Limit[uint32(roomid)] = append(self.Room_NumOfRobot_Limit[uint32(roomid)], frame)
	}
}

func (self *robotMgr) update() {
	for _, room := range GameMgr.rooms {
		self.UpdateRobot(room.roomID, room.RobotCount)
	}
}

func (self *robotMgr) FreeRobot() *account.Account {
	var acc *account.Account
	for _, robot := range account.AccountMgr.AccountbyID {
		if robot.Robot != 0 && robot.RoomID == 0 {
			acc = robot
			break
		}
	}
	if acc == nil {
		money := uint64(utils.Randx_y(1000, 100000))
		self.NameTableIndex++
		name := config.Get_configString("robot_name", self.NameTableIndex, "Name")
		acc = account.AccountMgr.CreateAccount(name, types.LOGIN_TYPE_ROBOT.Value(), name, "", 1, "", 0, 1, money)
	}
	return acc
}
func (self *robotMgr) UpdateRobot(roomID uint32, robotCount uint32) {
	frames := self.Room_NumOfRobot_Limit[roomID]
	now := utils.MilliSecondTimeSince1970()
	nowWeek := utils.Week()
	for _, frame := range frames {
		if frame.StartTime <= now && now <= frame.EndTime && frame.Week[nowWeek] {
			// 命中时间范围判断人数是否需要加机器人
			for robotCount < frame.Num {
				log.Infof("房间:%v 机器人数:%v 小于时段机器人数:%v ", roomID, robotCount, frame.Num)
				room := GameMgr.rooms[roomID]
				if room == nil {
					log.Warnf("机器人进入房间 找不到房间:%v", roomID)
					return
				}
				node := GameMgr.nodes[room.serverID]
				if node == nil {
					log.Warnf("找不到服务器节点 roomID:%v, serverID:%v ", roomID, room.serverID)
					return
				}
				acc := self.FreeRobot()
				acc.RoomID = roomID
				m := int64(utils.Randx_y(frame.MoneyMin, frame.MoneyMax)) - int64(acc.GetMoney())
				acc.AddMoney(m, common.EOperateType_INIT, 0)
				sendPB := &inner.PLAYER_DATA_REQ{
					Account:     acc.AccountStorageData,
					AccountData: acc.AccountGameData,
					RoomID:      roomID,
					Reback:      true,
				}
				send_tools.Send2Game(inner.SERVERMSG_HG_PLAYER_DATA_REQ.UInt16(), sendPB, node.session)
				roomidMSG := uint16(0)
				var pbMessage proto.Message
				switch common.EGameType(node.gameType) {
				case common.EGameTypeFRUITMARY:
					roomidMSG = protomsg.FRUITMARYMSG_CS_ENTER_GAME_FRUITMARY_REQ.UInt16()
					pbMessage = &protomsg.ENTER_GAME_FRUITMARY_REQ{AccountID: acc.GetAccountId(), RoomID: roomID}
				case common.EGameTypeDFDC:
					roomidMSG = protomsg.DFDCMSG_CS_ENTER_GAME_DFDC_REQ.UInt16()
					pbMessage = &protomsg.ENTER_GAME_DFDC_REQ{AccountID: acc.GetAccountId(), RoomID: roomID}
				case common.EGameTypeJPM:
					roomidMSG = protomsg.JPMMSG_CS_ENTER_GAME_JPM_REQ.UInt16()
					pbMessage = &protomsg.ENTER_GAME_JPM_REQ{AccountID: acc.GetAccountId(), RoomID: roomID}
				case common.EGameTypeLUCKFRUIT:
					roomidMSG = protomsg.LUCKFRUITMSG_CS_ENTER_GAME_LUCKFRUIT_REQ.UInt16()
					pbMessage = &protomsg.ENTER_GAME_LUCKFRUIT_REQ{AccountID: acc.GetAccountId(), RoomID: roomID}
				case common.EGameTypeRED2BLACK:
					roomidMSG = protomsg.RED2BLACKMSG_CS_ENTER_GAME_RED2BLACK_REQ.UInt16()
					pbMessage = &protomsg.ENTER_GAME_RED2BLACK_REQ{AccountID: acc.GetAccountId(), RoomID: roomID}
				case common.EGameTypeLHD:
					roomidMSG = protomsg.LHDMSG_CS_ENTER_GAME_LHD_REQ.UInt16()
					pbMessage = &protomsg.ENTER_GAME_LHD_REQ{AccountID: acc.GetAccountId(), RoomID: roomID}
				default:

				}
				send_tools.Send2Game(roomidMSG, pbMessage, node.session)
				log.Infof("机器人:[%v] 请求进入房间:%v 给游戏:%v 发送数据 ", acc.GetAccountId(), roomID, room.serverID)
				robotCount++
			}
			break
		}
	}
}
func (self *robotMgr) OnEvent(ev core.Event, evt core.EventType) {
	switch evt {
	case event.EventType_RoomUpdate:
		//tWrapEv := ev.(core.WrapEvent)
		//roomUpdate := tWrapEv.Event.(event.RoomUpdate)
		//log.Infof("处理房间更新人数事件:%+v ", roomUpdate)
		//self.UpdateRobot(roomUpdate.RoomID, roomUpdate.RobotCount)
	default:
		log.Warnf("事件:%v 未处理", evt)
	}
}