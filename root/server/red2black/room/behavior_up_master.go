package room

import (
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"root/core/utils"
	"root/protomsg"
	"root/server/red2black/event"
)

type (
	Robot_UpMaster struct {
		Room      *Room
		conf_rate [][]int64
	}
)

func New_Behavior_UpMaster(room *Room) {
	conf_ := utils.SplitConf2Arr_ArrInt64(config.GetPublicConfig_String("ROBOT_UP_LIST_MIN"))
	obj := &Robot_UpMaster{Room: room, conf_rate: conf_}
	event.Dispatcher.AddEventListener(event.EventType_Update_UpMaster, obj)
}

func (self *Robot_UpMaster) OnEvent(e core.Event, t core.EventType) {
	switch t {
	case event.EventType_Update_UpMaster:
		wrap := e.(core.WrapEvent)
		data := wrap.Event.(*event.UpMaster)
		self.upmaster_logic(data)
	}
}

func (self *Robot_UpMaster) upmaster_logic(ev *event.UpMaster) {

	l := len(ev.Applist)
	rate := 0
	for _, r := range self.conf_rate {
		if l <= int(r[0]) {
			rate = int(r[1])
			break
		}
	}
	if rate == 0 {
		return
	}
	bazhuang := config.GetPublicConfig_Int64("ROBOT_DOMINATE_RATE")
	conf_share_val := config.GetPublicConfig_Int64("R2B_DOMINATE_MONEY")                // 1份的金额
	conf_ba_share := config.GetPublicConfig_Int64("R2B_DOMINATE_VAL")                   // 霸庄最低份额
	conf_down_master := config.GetPublicConfig_Int64("ROBOT_DOWN_MASTER_RATE")          // 下庄概率
	conf_down_master_limit := config.GetPublicConfig_String("ROBOT_MASTER_COUNT_LIMIT") // 上庄数量限制
	conf_down_master_limit_arr := utils.SplitConf2Arr_ArrInt64(conf_down_master_limit)

	count := self.Room.count()
	min := 0
	size_limit := 0
	for _, v := range conf_down_master_limit_arr {
		if min <= count && count < int(v[0]) {
			size_limit = int(v[1])
			break
		}
		min = int(v[0])
	}
	master_count := self.Room.master_count()
	if size_limit == 4 || (4-master_count <= size_limit) {
		return
	}

	for _, acc := range ev.Robots {
		masterIndex := self.Room.SeatMasterIndex(acc.AccountId)
		if masterIndex != -1 && utils.Probability(int(conf_down_master)) {
			msg := packet.NewPacket(nil)
			msg.SetMsgID(protomsg.Old_MSGID_R2B_UP_MASTER.UInt16())
			msg.WriteUInt32(acc.AccountId)
			msg.WriteUInt8(0)
			msg.WriteUInt64(0)
			core.CoreSend(0, int32(ev.RoomID), msg.GetData(), 0)
		} else {
			if !utils.Probability(rate) {
				continue
			}

			max_share := acc.GetMoney() / uint64(conf_share_val)
			if max_share == 0 {
				continue
			}
			ret_val := 0
			if utils.Probability(int(bazhuang)) {
				// 霸庄
				if max_share < uint64(conf_ba_share) {
					continue
				}

				ret_val = utils.Randx_y(int(conf_ba_share), int(max_share+1))
			} else {
				// 拼庄
				ret_val = utils.Randx_y(int(1), int(max_share+1))
			}

			msg := packet.NewPacket(nil)
			msg.SetMsgID(protomsg.Old_MSGID_R2B_UP_MASTER.UInt16())
			msg.WriteUInt32(acc.AccountId)
			msg.WriteUInt8(1)
			msg.WriteUInt64(uint64(ret_val))

			timer := utils.Randx_y(0, int(5000))
			roomid := ev.RoomID
			self.Room.owner.AddTimer(int64(timer), 1, func(dt int64) {
				core.CoreSend(0, int32(roomid), msg.GetData(), 0)
			})
			log.Debugf(colorized.Blue("机器人:[%v] 请求上庄 购买份额:[%v] 身上钱:%v"), acc.AccountId, ret_val, acc.GetMoney())
		}

	}

}
