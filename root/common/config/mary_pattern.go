package config

import (
	"root/core"
	"root/core/log"
	"encoding/json"
	"io/ioutil"
	"strconv"
)

type (
	mary_pattern_config map[int]interface{}
)

var Global_mary_pattern_config mary_pattern_config

func init() {
	Load_mary_pattern_Conf()
}

func Load_mary_pattern_Conf() {
	lock.Lock()
	defer lock.Unlock()

	Global_mary_pattern_config = mary_pattern_config{}
	data, e := ioutil.ReadFile(core.ConfigDir + "mary_pattern.json")
	if e != nil {
		log.Errorf("mary_pattern 错误:%v", e.Error())
		return
	}
	error := json.Unmarshal(data, &Global_mary_pattern_config)
	if error != nil {
		log.Errorf(" error %v", error.Error())
		return
	}
	log.Info("加载完成mary_pattern.json")
}

func Get_mary_pattern_Config(ID int, key string) string {
	roomConfig := Global_mary_pattern_config[ID]
	if roomConfig == nil {
		log.Panicf("找不到房间配置Global_mary_pattern_config[%v]", ID)
	}

	m := roomConfig.(map[string]interface{})
	if val,e := m[key];!e {
		log.Panicf("配置mary_pattern.json ID:%v 找不到字段：%v ",ID,key)
		return ""
	}else {
		return val.(string)
	}
}
func Get_mary_pattern_ConfigInt32(ID int, key string) int32 {
	ret :=  Get_mary_pattern_Config(ID, key)
	i64,e := strconv.Atoi(ret)
	if e != nil {
		log.Panicf("Global_mary_pattern_config配置不能转换成 int32 id:%v key:%v val:%v", ID,key,ret)
	}
	return int32(i64)
}
