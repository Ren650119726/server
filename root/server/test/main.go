package main

import (
	"root/core/log"
	"runtime"
)
type test_stru struct {
	val int
	val2 int
	val3 int
	val4 int
	val5 int
	val6 int
	val8 int
	valfe int
	valf32f int
	valf325 int

}

func testObj() []*test_stru{
	arr := make([]*test_stru,0,0)
	for i := 0;i < 100000000; i++{
		arr = append(arr, &test_stru{val:i})
	}
	log.Infof("for over")
	s := &runtime.MemStats{}
	runtime.ReadMemStats(s)
	log.Infof("alloc %+v",s)
	return  arr
}
func main() {
	s := &runtime.MemStats{}
	runtime.ReadMemStats(s)
	log.Infof("init %+v",s)

	arr := testObj()
	runtime.ReadMemStats(s)
	log.Infof("gc %+v",s)
	log.Infof("NumGC %v",s.NumGC)
	log.Infof("GCCPUFraction %v",s.GCCPUFraction)
	log.Infof("PauseTotalNs %v",s.PauseTotalNs)
	log.Infof("LastGC %v",s.LastGC)
	log.Infof("PauseEnd %v",s.PauseEnd)
	log.Infof("over %v", len(arr))
}
