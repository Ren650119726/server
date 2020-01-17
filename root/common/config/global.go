package config

import (
	"encoding/json"
	"io/ioutil"
	"root/core"
	"root/core/log"
	"root/core/utils"
	"strconv"
	"strings"
	"sync"
)

type global_type map[string]interface{}

//val 只有两种，int64 string
var global_public_config global_type

var lock sync.Mutex

func init() {
	LoadPublic_Conf()
}

func LoadPublic_Conf() {
	lock.Lock()
	defer lock.Unlock()

	global_public_config = global_type{}
	data, e := ioutil.ReadFile(core.ConfigDir + "global.json")
	if e != nil {
		log.Errorf("global 错误:%v", e.Error())
		return
	}
	err := json.Unmarshal(data, &global_public_config)
	if err != nil {
		log.Errorf(" error %v", err.Error())
		return
	}

	log.Info("加载完成global.json")
}

func GetPublicConfig_Int64(key string) int64 {
	lock.Lock()
	defer lock.Unlock()

	p, e := global_public_config[key]
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

func GetPublicConfig_Float32(key string) float32 {
	lock.Lock()
	defer lock.Unlock()

	p, e := global_public_config[key]
	if !e || p == "" {
		log.Errorf("全局配置找不到:[%v]", key)
		return 0
	}
	pp := p.(map[string]interface{})

	val, err := strconv.ParseFloat(pp["Value"].(string), 32)
	if err != nil {
		log.Errorf("字符串不是浮点型:%v", p)
		return 0
	}

	return float32(val)
}

func GetPublicConfig_String(key string) string {
	lock.Lock()
	defer lock.Unlock()

	p, e := global_public_config[key]
	if !e || p == "" {
		log.Panicf("全局配置找不到:[%v]", key)
	}
	pp := p.(map[string]interface{})
	return pp["Value"].(string)
}

func GetPublicConfig_Slice(key string) []int {
	lock.Lock()
	defer lock.Unlock()

	p, e := global_public_config[key]
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

func GetPublicConfig_Mapi(key string) map[int]int {
	lock.Lock()
	defer lock.Unlock()
	p, _ := global_public_config[key]
	p, e := global_public_config[key]
	if !e || p == "" {
		log.Panicf("全局配置找不到:[%v]", key)
	}
	pp := p.(map[string]interface{})
	ret := utils.SplitConf2Mapii(pp["Value"].(string))
	return ret
}

// 解析"1#100,2#200,3#300"为[[1,100],[2,200],[3,300]]
func GetPublicConfig_ArrInt64(key string) [][]int64 {
	lock.Lock()
	defer lock.Unlock()
	p, e := global_public_config[key]
	if !e || p == "" {
		log.Panicf("全局配置找不到:[%v]", key)
	}
	pp := p.(map[string]interface{})
	ret := utils.SplitConf2Arr_ArrInt64(pp["Value"].(string))
	return ret
}

// 解析"1#100,2#200,3#300"为{1:100,2:200,3:300}}
func GetPublicConfig_MapStrInt(key string) map[string]int {
	lock.Lock()
	defer lock.Unlock()
	p, e := global_public_config[key]
	if !e || p == "" {
		log.Panicf("全局配置找不到:[%v]", key)
	}
	pp := p.(map[string]interface{})
	ret := utils.SplitConf2MapStrInt(pp["Value"].(string))
	return ret
}

func GetPublicConfig_MapStrStr(key string) map[string]string {
	lock.Lock()
	defer lock.Unlock()
	p, e := global_public_config[key]
	if !e || p == "" {
		log.Panicf("全局配置找不到:[%v]", key)
	}
	pp := p.(map[string]interface{})
	ret := utils.SplitConf2MapStrStr(pp["Value"].(string), "|", "*")
	return ret
}