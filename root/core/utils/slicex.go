package utils

// 从切片中删除指定下标的值, 并返回新切片和删除的值, 新切片中将去掉删除的值
func RemoveSliceIndex(sSlice []uint32, nIndex uint8) ([]uint32, uint32) {
	nLastIndex := len(sSlice) - 1
	nValue := sSlice[nIndex]
	sSlice[nIndex], sSlice[nLastIndex] = sSlice[nLastIndex], sSlice[nIndex]
	sSlice = sSlice[:nLastIndex]
	return sSlice, nValue
}

func StatElementCount(sSlice []uint8) map[uint8]int {
	mCount := make(map[uint8]int, 0)
	for _, value := range sSlice {
		if _, isExist := mCount[value]; isExist == false {
			mCount[value] = 1
		} else {
			mCount[value]++
		}
	}
	return mCount
}
