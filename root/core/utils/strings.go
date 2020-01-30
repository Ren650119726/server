package utils

import (
	"root/core/log"
	"strconv"
	"strings"
	"unicode/utf8"
)

const EMAIL_REG = "^[_a-z0-9-]+(\\.[_a-z0-9-]+)*@[a-z0-9-]+(\\.[a-z0-9-]+)*(\\.[a-z]{2,4})$"
const PHONE_REG = "^(13[0-9]|14[579]|15[0-3,5-9]|16[6]|17[0135678]|18[0-9]|19[89])\\d{8}$"

// 过滤大于3字节的字符(比如说emoji)
func FilterEmoji(content string) string {
	new_content := ""
	for _, value := range content {
		_, size := utf8.DecodeRuneInString(string(value))
		if size <= 3 {
			new_content += string(value)
		}
	}
	return new_content
}

// 解析"1,2,3,4,5"为[1,2,3,4,5]
func SplitConf2ArrInt32(value string, sep string) []int32 {
	strs := strings.Split(value, sep)
	arrInt32 := make([]int32, 0, 2)

	for _, v := range strs {
		iv, err := strconv.Atoi(v)
		if err != nil {
			log.Errorf("解析错误 :%v", value)
			return nil
		}
		arrInt32 = append(arrInt32, int32(iv))
	}
	return arrInt32
}

// 解析"1#100,2#200,3#300"为[[1,100],[2,200],[3,300]]
func SplitConf2Arr_ArrInt32(value string) [][]int32 {
	strs1 := strings.Split(value, ",")
	arrInt32 := make([][]int32, 0, 2)

	for _, v := range strs1 {
		strs2 := strings.Split(v, "#")
		arr := make([]int32, 0, 2)
		for _, v := range strs2 {
			iv, err := strconv.Atoi(v)
			if err != nil {
				log.Errorf("解析错误 :%v", value)
				return nil
			}
			arr = append(arr, int32(iv))
		}
		arrInt32 = append(arrInt32, arr)
	}
	return arrInt32
}

// 解析"1,2,3,4,5"为[1,2,3,4,5]
func SplitConf2ArrInt64(value string) []int64 {
	strs := strings.Split(value, ",")
	arrInt64 := make([]int64, 0, 2)

	for _, v := range strs {
		iv, err := strconv.Atoi(v)
		if err != nil {
			log.Errorf("解析错误 :%v", value)
			return nil
		}
		arrInt64 = append(arrInt64, int64(iv))
	}
	return arrInt64
}
// 解析"1,2,3,4,5"为[1,2,3,4,5]
func SplitConf2ArrUInt64(value string) []uint64 {
	strs := strings.Split(value, ",")
	arrInt64 := make([]uint64, 0, 2)

	for _, v := range strs {
		iv, err := strconv.Atoi(v)
		if err != nil {
			log.Errorf("解析错误 :%v", value)
			return nil
		}
		arrInt64 = append(arrInt64, uint64(iv))
	}
	return arrInt64
}

// 解析"1#100,2#200,3#300"为[[1,100],[2,200],[3,300]]
func SplitConf2Arr_ArrInt64(value string) [][]int64 {
	strs1 := strings.Split(value, ",")
	arrInt64 := make([][]int64, 0, 2)

	for _, v := range strs1 {
		strs2 := strings.Split(v, "#")
		arr := make([]int64, 0, 2)
		for _, v := range strs2 {
			iv, err := strconv.Atoi(v)
			if err != nil {
				log.Errorf("解析错误 :%v", value)
				return nil
			}
			arr = append(arr, int64(iv))
		}
		arrInt64 = append(arrInt64, arr)
	}
	return arrInt64
}

// 解析"1#100,2#200,3#300"为map{1:100,2:200,3:300}
func SplitConf2map(value string) map[int32]float32 {
	strs1 := strings.Split(value, ",")
	k_vMap := make(map[int32]float32)

	for _, v := range strs1 {
		strs2 := strings.Split(v, "#")
		if len(strs2) != 2 {
			log.Errorf("解析错误，当前字段数量不为2个", value)
			return nil
		}
		key, err := strconv.Atoi(strs2[0])
		if err != nil {
			log.Errorf("解析错误，当前字段不是数字", value)
			return nil
		}
		value, err := strconv.ParseFloat(strs2[1], 32)
		if err != nil {
			log.Errorf("解析错误，当前字段不是float", value)
			return nil
		}

		k_vMap[int32(key)] = float32(value)
	}
	return k_vMap
}

func CountMapByMapii(mMap map[int]int) uint16 {
	nCount := uint16(0)
	for _, value := range mMap {
		nCount += uint16(value)
	}
	return nCount
}

// 解析"1*6|2*10|3*8|7*2"为map{1:6,2:10,3:8,7:2}
func SplitConf2Mapii(value string) map[int]int {
	ret := make(map[int]int)
	s1 := strings.Split(value, "|")
	for _, v := range s1 {
		s2 := strings.Split(v, "*")
		key, err := strconv.Atoi(s2[0])
		if err != nil {
			log.Errorf("无法解析配置:%v, 不能把%v转成整型", value, s2[0])
			return nil
		}
		val, err := strconv.Atoi(s2[1])
		if err != nil {
			log.Errorf("无法解析配置:%v, 不能把%v转成整型", value, s2[1])
			return nil
		}
		ret[key] = val
	}
	return ret
}

// 解析"1*6|2*10|3*8|7*2"为map{1:6,2:10,3:8,7:2}
func SplitConf2Mapis(value string) map[int]string {
	ret := make(map[int]string)
	s1 := strings.Split(value, "|")
	for _, v := range s1 {
		s2 := strings.Split(v, "*")
		key, err := strconv.Atoi(s2[0])
		if err != nil {
			log.Errorf("无法解析配置:%v, 不能把%v转成整型", value, s2[0])
			return nil
		}
		ret[key] = s2[1]
	}
	return ret
}

// 解析"1#1*6|2*10@3*8|7*2"为map{1:"1:6,2:10", 3:"3:8,7:2"}
// @分隔大组, #前面的数值表示大组的key
// |分隔小组, *前面的数值表示小组的key
func SplitConf2Mapistr(value string) map[int]string {
	ret := make(map[int]string)
	s1 := strings.Split(value, "@")
	for _, v := range s1 {
		s2 := strings.Split(v, "#")
		key, err := strconv.Atoi(s2[0])
		if err != nil {
			log.Errorf("无法解析配置:%v, 不能把%v转成整型", value, s2[0])
			return nil
		}
		ret[key] = s2[1]
	}
	return ret
}

// 解析"192.168.2.30*1|192.168.2.72*7"为map{192.168.2.30:1, 192.168.2.72:7}
func SplitConf2MapStrInt(value string) map[string]int {
	ret := make(map[string]int)
	s1 := strings.Split(value, "|")
	for _, v := range s1 {
		s2 := strings.Split(v, "*")
		key := s2[0]
		val, err := strconv.Atoi(s2[1])
		if err != nil {
			log.Errorf("无法解析配置:%v, 不能把%v转成整型", value, s2[1])
			return nil
		}
		ret[key] = val
	}
	return ret
}

func SplitConf2MapStrStr(value string, cutter1, cutter2 string) map[string]string {
	ret := make(map[string]string)
	s1 := strings.Split(value, cutter1)
	for _, v := range s1 {
		s2 := strings.Split(v, cutter2)
		key := s2[0]
		ret[key] = s2[1]
	}
	return ret
}

// 第一参数: 配置内容, 格式: 0*30|1*40|2*30; 根据概率返回0,1,2之一
// 第二参数: 失败默认返回值
// 随到0的概率30%  随到1概率40  随到2概率30
// 概率总和可以是100% 也可以超过100%
func CalcRandomReturnKey(param string, default_ret int) int {
	ret := SplitConf2Mapii(param)
	var sKey []int
	var sVal []int
	nTotalVal := 0
	for key, value := range ret {
		nTotalVal += value
		sKey = append(sKey, key)
		sVal = append(sVal, nTotalVal)
	}
	if sKey == nil || len(sKey) == 0 {
		return default_ret
	}

	nRandom := Randx_y(1, nTotalVal)
	for i := 0; i < len(sKey); i++ {
		if nRandom < sVal[i] {
			return sKey[i]
		}
	}
	return default_ret
}
