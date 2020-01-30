package config

import (
	"root/core"
	"root/core/log"
	"encoding/json"
	"io/ioutil"
	"strconv"
)

type (
	mary_room_config map[int]interface{}
)

var Global_mary_room_config mary_room_config

func init() {
	Load_mary_room_Conf()
}

func Load_mary_room_Conf() {
	lock.Lock()
	defer lock.Unlock()

	Global_mary_room_config = mary_room_config{}
	data, e := ioutil.ReadFile(core.ConfigDir + "mary_room.json")
	if e != nil {
		log.Errorf("fruitMary 错误:%v", e.Error())
		return
	}
	error := json.Unmarshal(data, &Global_mary_room_config)
	if error != nil {
		log.Errorf(" error %v", error.Error())
		return
	}
	log.Info("加载完成mary_room.json")
}

func Get_mary_room_Config(roomID int, key string) string {
	roomConfig := Global_mary_room_config[roomID]
	if roomConfig == nil {
		log.Panicf("找不到房间配置Global_fruitmary_room_config[%v]", roomID)
	}

	m := roomConfig.(map[string]interface{})
	if val,e := m[key];!e {
		log.Panicf("配置mary_room.json roomID:%v 找不到字段：%v ",roomID,key)
		return ""
	}else {
		return val.(string)
	}
}
func Get_mary_room_ConfigInt64(roomID int, key string) int64 {
	ret :=  Get_mary_room_Config(roomID, key)
	i64,_ := strconv.Atoi(ret)
	return int64(i64)
}
