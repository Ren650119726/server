package main

import (
	"context"
	"google.golang.org/grpc"
	"net"
	"root/core/log"
	"root/protomsg"
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
type Server struct{}
func (s *Server)DoMD5(ctx context.Context, in *protomsg.Req) (*protomsg.Res, error){
	log.Infof("收到了 grpc消息！！！！")
	return &protomsg.Res{BackJson:"json hahahahah "},nil
}


func gRPC(){
	lis, err := net.Listen("tcp", ":8028")  //监听所有网卡8028端口的TCP连接
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}
	rpc := grpc.NewServer()
	protomsg.RegisterWaiterServer(rpc, &Server{})
	rpc.Serve(lis)

}
func test(){
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
func main() {

}
