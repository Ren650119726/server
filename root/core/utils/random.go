package utils

import (
	"math/rand"
	"reflect"
	"root/core/log"
)

// [[],[],[],[],[],[]] i是权重下标，随机一个[]
func RandomWeight64(rands [][]int64, i int) (index int) {
	total := int64(0)
	for _, v := range rands {
		total += v[i]
	}
	if total == 0 {
		log.Errorf("total ==0 rands:%v", rands)
		return 0
	}
	retValue := rand.Int63n(total)

	addValue := int64(0)
	for k, v := range rands {
		if addValue <= retValue && retValue < (addValue+v[i]) {
			return k
		}
		addValue += v[i]
	}

	log.Errorf("没有筛选出权重值 i:%v  :%v  retValue:%v", i, rands, retValue)
	return -1
}

func RandomWeight32(rands [][]int32, i int) (index int) {
	total := int32(0)
	for _, v := range rands {
		total += v[i]
	}
	retValue := rand.Int31n(total)

	addValue := int32(0)
	for k, v := range rands {
		if addValue <= retValue && retValue < (addValue+v[i]) {
			return k
		}
		addValue += v[i]
	}

	log.Errorf("没有筛选出权重值 i:%v  :%v, retValue;%v", i, rands, retValue)
	return -1
}

// 不包含y，[x,y)
func Randx_y(x, y int) int {
	if x > y {
		log.Errorf("随机范围值x>y; x:%v y:%v", x, y)
		return 0
	} else if x == y {
		return x
	}
	return rand.Intn(y-x) + x
}

// 函数作用: 将切片元素顺序打乱
func RandomSlice(sSlice interface{}) {
	t := reflect.TypeOf(sSlice)
	if t.Kind() != reflect.Slice {
		log.Error("传入的参数不是切片!!!")
		return
	}

	v := reflect.ValueOf(sSlice)
	fSwapper := reflect.Swapper(sSlice)

	nLen := v.Len()
	nEnd := nLen - 1
	nLastIndex := nLen
	for i := 0; i < nEnd; i++ {
		nLastIndex--
		nIndex := Randx_y(0, nLastIndex)
		fSwapper(nIndex, nLastIndex)
	}
}

// 从切片中随机获取一个int值, 并返回新切片, 新切片中将去掉返回的值
func RandomSliceAndRemoveReturn(slice []uint32) ([]uint32, uint32) {
	last_idx := len(slice) - 1
	rand_idx := Randx_y(0, last_idx)
	value := slice[rand_idx]
	slice[rand_idx], slice[last_idx] = slice[last_idx], slice[rand_idx]

	slice = slice[:last_idx]
	return slice, value
}

// 简单的概率ratio 1-100
func Probability(ratio int) bool {
	v := Randx_y(0, 100)
	return v < ratio
}

// 简单的概率ratio 1-10000
func Probability10000(ratio int) bool {
	v := Randx_y(0, 10000)
	return v < ratio
}
