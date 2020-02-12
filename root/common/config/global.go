package config

import (
	"root/core/log"
	"root/core/utils"
	"strconv"
	"strings"
	"sync"
)

var lock sync.RWMutex

func GetPublicConfig_Int64(key int) int64 {
	lock.RLock()
	defer lock.RUnlock()

	p, e := Config_Data["global"][key]
	if !e || p == "" {
		log.Errorf("全局配置找不到:[%v]", key)
		return 0
	}
	pp := p.(map[string]interface{})

	val, err := strconv.Atoi(pp["Value"].(string))
	if err != nil {
		log.Errorf("字符串不是整型:%v", p)
		return 0
	}

	return int64(val)
}

func GetPublicConfig_String(key int) string {
	lock.RLock()
	defer lock.RUnlock()

	p, e := Config_Data["global"][key]
	if !e || p == "" {
		log.Panicf("全局配置找不到:[%v]", key)
	}
	pp := p.(map[string]interface{})
	return pp["Value"].(string)
}

func GetPublicConfig_Slice(key int) []int {
	lock.RLock()
	defer lock.RUnlock()

	p, e := Config_Data["global"][key]
	if !e || p == "" {
		log.Panicf("全局配置找不到:[%v]", key)
	}
	pp := p.(map[string]interface{})

	valString := pp["Value"].(string)
	ret := make([]int, 0)
	s1 := strings.Split(valString, ",")
	for _, v := range s1 {
		value, err := strconv.Atoi(v)
		if err != nil {
			log.Errorf("无法解析配置:%v, 不能把%v转成整型", valString, v)
			return nil
		}
		ret = append(ret, value)
	}
	return ret
}

func GetPublicConfig_Mapi(key int) map[int]int {
	lock.RLock()
	defer lock.RUnlock()

	p, e := Config_Data["global"][key]
	if !e || p == "" {
		log.Panicf("全局配置找不到:[%v]", key)
	}
	pp := p.(map[string]interface{})
	ret := utils.SplitConf2Mapii(pp["Value"].(string))
	return ret
}

// 解析"1#100,2#200,3#300"为[[1,100],[2,200],[3,300]]
func GetPublicConfig_ArrInt64(key int) [][]int64 {
	lock.RLock()
	defer lock.RUnlock()
	p, e := Config_Data["global"][key]
	if !e || p == "" {
		log.Panicf("全局配置找不到:[%v]", key)
	}
	pp := p.(map[string]interface{})
	ret := utils.SplitConf2Arr_ArrInt64(pp["Value"].(string))
	return ret
}

// 解析"1#100,2#200,3#300"为{1:100,2:200,3:300}}
func GetPublicConfig_MapStrInt(key int) map[string]int {
	lock.RLock()
	defer lock.RUnlock()
	p, e := Config_Data["global"][key]
	if !e || p == "" {
		log.Panicf("全局配置找不到:[%v]", key)
	}
	pp := p.(map[string]interface{})
	ret := utils.SplitConf2MapStrInt(pp["Value"].(string))
	return ret
}

func GetPublicConfig_MapStrStr(key int) map[string]string {
	lock.RLock()
	defer lock.RUnlock()
	p, e := Config_Data["global"][key]
	if !e || p == "" {
		log.Panicf("全局配置找不到:[%v]", key)
	}
	pp := p.(map[string]interface{})
	ret := utils.SplitConf2MapStrStr(pp["Value"].(string), "|", "*")
	return ret
}