package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"root/core"
	"root/core/log"
	"root/core/utils"
	"strconv"
)

type (
	Config_map  map[int]interface{}
	json_config map[string]Config_map // 不带日期
)

var Config_Data json_config

func init() {
	Load_Conf()
}

func Load_Conf() {
	lock.Lock()
	defer lock.Unlock()

	Config_Data = make(json_config)
	dir_list, e := ioutil.ReadDir(core.ConfigDir)
	if e != nil {
		fmt.Println("read dir error")
		return
	}
	DataStrLen := len(utils.STD_NUMBER_FORMAT) + 1 // 日期的长度，多了一个下划线_ 所以需要+1
	for _, file := range dir_list {
		b, _ := regexp.Match("/*.json", []byte(file.Name()))
		if b {
			ret := regexp.MustCompile(`json`).FindStringIndex(file.Name())
			if ret != nil {
				data, e := ioutil.ReadFile(core.ConfigDir + file.Name())
				if e != nil {
					log.Errorf("文件读取错误:%v 错误:%v", file.Name(), e.Error())
					return
				}

				jsonname := file.Name()[:ret[0]-1-DataStrLen]
				cmap := make(Config_map)
				error := json.Unmarshal(data, &cmap)
				if error != nil {
					log.Errorf(" error %v file:%v ", error.Error(), file.Name())
					return
				}
				Config_Data[jsonname] = cmap
				log.Infof("加载完成 %v", file.Name())
			}
		}
	}
	log.Infof("")
}

func Get_config(table string) Config_map {
	lock.RLock()
	defer lock.RUnlock()
	tb, e := Config_Data[table]
	if !e {
		log.Panicf("找不到配置文件:%v", table)
		return nil
	}
	return tb
}

func Get_configString(table string, ID int, key string) string {
	lock.RLock()
	defer lock.RUnlock()
	tb, e := Config_Data[table]
	if !e {
		log.Panicf("找不到配置文件:%v", table)
		return ""
	}

	config := tb[ID]
	if config == nil {
		log.Panicf("配置文件:%v 找不到 ID:%v ", table, ID)
	}

	m := config.(map[string]interface{})
	if val, e := m[key]; !e {
		log.Panicf("配置文件:%v ID:%v 找不到字段:%v", table, ID, key)
		return ""
	} else {
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

func Get_configInt(table string, ID int, key string) int {
	lock.RLock()
	defer lock.RUnlock()
	tb, e := Config_Data[table]
	if !e {
		log.Panicf("找不到配置文件:%v", table)
		return 0
	}

	config := tb[ID]
	if config == nil {
		log.Panicf("配置文件:%v 找不到 ID:%v ", table, ID)
	}

	m := config.(map[string]interface{})
	if val, e := m[key]; !e {
		log.Panicf("配置文件:%v ID:%v 找不到字段:%v", table, ID, key)
		return 0
	} else {
		switch val.(type) {
		case string:
			i, e := strconv.Atoi(val.(string))
			if e != nil {
				log.Errorf("字段不是int:%v", i)
			}
			return 0
		case float64:
			return int(val.(float64))
		default:
			return val.(int)
		}
		return val.(int)
	}
}

func Get_JsonDataString(jsonData interface{}, ID int, key string) string {
	lock.RLock()
	defer lock.RUnlock()

	m := jsonData.(map[string]interface{})
	if val, e := m[key]; !e {
		log.Panicf("数据:%v ID:%v 找不到字段:%v", jsonData, ID, key)
		return ""
	} else {
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
func Get_JsonDataInt(jsonData interface{}, ID int, key string) int {
	lock.RLock()
	defer lock.RUnlock()

	m := jsonData.(map[string]interface{})
	if val, e := m[key]; !e {
		log.Panicf("数据:%v ID:%v 找不到字段:%v", jsonData, ID, key)
		return 0
	} else {
		switch val.(type) {
		case string:
			i, e := strconv.Atoi(val.(string))
			if e != nil {
				log.Errorf("字段不是int:%v", i)
			}
			return 0
		case float64:
			return int(val.(float64))
		default:
			return val.(int)
		}
		return val.(int)
	}
}
