package config

import (
	"root/core"
	"root/core/log"
	"encoding/json"
	"io/ioutil"
)

type (
	rooms_config map[int]interface{}
)

var Global_fruitmary_room_config rooms_config

func init() {
	LoadStore_Conf()
}

func LoadStore_Conf() {
	lock.Lock()
	defer lock.Unlock()

	Global_fruitmary_room_config = rooms_config{}
	data, e := ioutil.ReadFile(core.ConfigDir + "fruitMary.json")
	if e != nil {
		log.Errorf("fruitMary 错误:%v", e.Error())
		return
	}
	error := json.Unmarshal(data, &Global_fruitmary_room_config)
	if error != nil {
		log.Errorf(" error %v", error.Error())
		return
	}
	log.Info("加载完成mary_room.json")
}