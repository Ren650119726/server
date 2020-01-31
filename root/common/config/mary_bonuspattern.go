package config

import (
	"root/core"
	"root/core/log"
	"encoding/json"
	"io/ioutil"
	"strconv"
)

type (
	mary_bonuspattern_config map[int]interface{}
)

var Global_mary_bonuspattern_config mary_bonuspattern_config

func init() {
	Load_mary_bonuspattern_Conf()
}

func Load_mary_bonuspattern_Conf() {
	lock.Lock()
	defer lock.Unlock()

	Global_mary_bonuspattern_config = mary_bonuspattern_config{}
	data, e := ioutil.ReadFile(core.ConfigDir + "mary_bonuspattern.json")
	if e != nil {
		log.Errorf("mary_bonuspattern 错误:%v", e.Error())
		return
	}
	error := json.Unmarshal(data, &Global_mary_bonuspattern_config)
	if error != nil {
		log.Errorf(" error %v", error.Error())
		return
	}
	log.Info("加载完成mary_bonuspattern.json")
}

func Get_mary_bonuspattern_Config(ID int, key string) string {
	lock.Lock()
	defer lock.Unlock()

	roomConfig := Global_mary_bonuspattern_config[ID]
	if roomConfig == nil {
		log.Panicf("找不到房间配置Global_mary_pattern_config[%v]", ID)
	}

	m := roomConfig.(map[string]interface{})
	if val,e := m[key];!e {
		log.Panicf("配置mary_bonuspattern.json ID:%v 找不到字段：%v ",ID,key)
		return ""
	}else {
		switch val.(type) {
		case string:
			return val.(string)
		case float64:
			return strconv.Itoa(int(val.(float64)))
		default:
			return strconv.Itoa(val.(int))
		}
		return val.(string)
	}
}
func Get_mary_bonuspattern_ConfigInt32(ID int, key string) int32 {
	ret :=  Get_mary_bonuspattern_Config(ID, key)
	i64,_ := strconv.Atoi(ret)
	return int32(i64)
}
