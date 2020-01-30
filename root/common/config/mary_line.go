package config

import (
	"root/core"
	"root/core/log"
	"encoding/json"
	"io/ioutil"
	"strconv"
)

type (
	mary_line_config map[int]interface{}
)

var Global_mary_line_config mary_line_config

func init() {
	Load_mary_line_Conf()
}

func Load_mary_line_Conf() {
	lock.Lock()
	defer lock.Unlock()

	Global_mary_line_config = mary_line_config{}
	data, e := ioutil.ReadFile(core.ConfigDir + "mary_line.json")
	if e != nil {
		log.Errorf("mary_line 错误:%v", e.Error())
		return
	}
	error := json.Unmarshal(data, &Global_mary_line_config)
	if error != nil {
		log.Errorf(" error %v", error.Error())
		return
	}
	log.Info("加载完成mary_line.json")
}

func Get_mary_line_config(ID int, key string) string {
	roomConfig := Global_mary_line_config[ID]
	if roomConfig == nil {
		log.Panicf("找不到房间配置Global_mary_line_config[%v]", ID)
	}

	m := roomConfig.(map[string]interface{})
	if val,e := m[key];!e {
		log.Panicf("配置mary_line.json ID:%v 找不到字段：%v ",ID,key)
		return ""
	}else {
		return val.(string)
	}
}
func Get_mary_line_configInt(ID int, key string) int {
	ret :=  Get_mary_line_config(ID, key)
	i64,e := strconv.Atoi(ret)
	if e != nil {
		log.Panicf("Global_mary_line_config配置不能转换成 int32 id:%v key:%v val:%v", ID,key,ret)
	}
	return int(i64)
}
