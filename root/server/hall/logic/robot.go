package logic

import (
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/utils"
	"root/protomsg"
	"root/protomsg/inner"
	"root/server/hall/account"
	"root/server/hall/event"
	"root/server/hall/send_tools"
	"time"
)

var RobotMgr = NewRobotMgr()

type (
	robotMgr struct {
		Room_NumOfRobot_Limit map[uint32][]*time_frame // key roomid
	}
	// 时段结构
	time_frame struct {
		StartTime int64
		EndTime   int64
		Week      [7]bool
		Num       uint32
	}
)

func NewRobotMgr() *robotMgr {
	return &robotMgr{}
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
		self.Room_NumOfRobot_Limit[uint32(roomid)] = append(self.Room_NumOfRobot_Limit[uint32(roomid)], frame)
	}
}

func (self *robotMgr) NewRobot() *account.Account {
	acc := account.NewAccount(&protomsg.AccountStorageData{
		Name:      "robot" + utils.DateString(),
		Robot:     1,
		AccountId: 0,
		Money:     uint64(utils.Randx_y(1000, 500000)),
	})
	return acc
}
func (self *robotMgr) UpdateRobot(roomID uint32, robotCount uint32) {
	frames := self.Room_NumOfRobot_Limit[roomID]
	now := utils.MilliSecondTimeSince1970()
	nowWeek := utils.Week()
	for _, frame := range frames {
		if frame.StartTime <= now && now <= frame.EndTime && frame.Week[nowWeek] {
			// 命中时间范围判断人数是否需要加机器人
			acc := self.NewRobot()
			if robotCount < frame.Num {
				log.Infof("房间:%v 机器人数:%v 小于时段机器人数:%v ", roomID, robotCount, frame.Num)
				room := GameMgr.rooms[roomID]
				if room == nil {
					log.Warnf("机器人进入房间 找不到房间:%v", roomID)
					return
				}
				node := GameMgr.nodes[room.serverID]
				if node == nil {
					log.Warnf("找不到服务器节点 accID:%v roomID:%v, serverID:%v ", acc.GetAccountId(), roomID, room.serverID)
					return
				}
				sendPB := &inner.PLAYER_DATA_REQ{
					Account:     acc.AccountStorageData,
					AccountData: acc.AccountGameData,
					RoomID:      roomID,
					Reback:      true,
				}
				send_tools.Send2Game(inner.SERVERMSG_HG_PLAYER_DATA_REQ.UInt16(), sendPB, node.session)
				log.Infof("机器人:[%v] 请求进入房间:%v 给游戏:%v 发送数据 ", acc.GetAccountId(), roomID, room.serverID)
			}
		}
	}
}
func (self *robotMgr) OnEvent(ev core.Event, evt core.EventType) {
	switch evt {
	case event.EventType_RoomUpdate:
		tWrapEv := ev.(core.WrapEvent)
		roomUpdate := tWrapEv.Event.(event.RoomUpdate)
		log.Infof("处理房间更新人数事件:%+v ", roomUpdate)
		//self.UpdateRobot(roomUpdate.RoomID, roomUpdate.RobotCount)
	default:
		log.Warnf("事件:%v 未处理", evt)
	}
}
