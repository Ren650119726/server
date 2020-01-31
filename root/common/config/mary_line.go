package config

import (
	"root/core"
	"root/core/log"
	"encoding/json"
	"io/ioutil"
	"strconv"
)

type (
	mary_lines_config map[int]interface{}
)

var Global_mary_lines_config mary_lines_config

func init() {
	Load_mary_lines_Conf()
}

func Load_mary_lines_Conf() {
	lock.Lock()
	defer lock.Unlock()

	Global_mary_lines_config = mary_lines_config{}
	data, e := ioutil.ReadFile(core.ConfigDir + "mary_lines.json")
	if e != nil {
		log.Errorf("mary_lines 错误:%v", e.Error())
		return
	}
	error := json.Unmarshal(data, &Global_mary_lines_config)
	if error != nil {
		log.Errorf(" error %v", error.Error())
		return
	}
	log.Info("加载完成mary_lines.json")
}

func Get_mary_lines_config(ID int, key string) string {
	lock.Lock()
	defer lock.Unlock()

	roomConfig := Global_mary_lines_config[ID]
	if roomConfig == nil {
		log.Panicf("找不到房间配置Global_mary_lines_config[%v]", ID)
	}

	m := roomConfig.(map[string]interface{})
	if val,e := m[key];!e {
		log.Panicf("配置mary_lines.json ID:%v 找不到字段：%v ",ID,key)
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
func Get_mary_lines_configInt(ID int, key string) int {
	ret :=  Get_mary_lines_config(ID, key)
	i64,e := strconv.Atoi(ret)
	if e != nil {
		log.Panicf("Global_mary_lines_config配置不能转换成 int32 id:%v key:%v val:%v", ID,key,ret)
	}
	return int(i64)
}
