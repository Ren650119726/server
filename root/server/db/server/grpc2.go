package server

import (
	"google.golang.org/grpc"
	"io"
	"net"
	"root/core/log"
	"root/protomsg"
)
type GRPC_Service struct {}
func (s *GRPC_Service) Route(stream protomsg.GRPC_SERVICE_RouteServer) error {
	for {
		d,e := stream.Recv()
		if e == io.EOF{
			log.Infof("????????????")
			return e
		}
		if e != nil {
			log.Infof("eee:%v",e.Error())
			return e
		}
		//if e != nil {
		//	eee := stream.SendAndClose(&protomsg.Close{
		//		Ret: 123123,
		//	})
		//	if eee != nil {
		//		log.Error("%v",eee.Error())
		//	}
		//
		//	return eee
		//}
		log.Infof("%v %v ",stream.Context(),d.String())
	}
}
func GRPC_SERVER2(){
	lis, err := net.Listen("tcp", ":8028")  //监听所有网卡8028端口的TCP连接
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}
	rpc := grpc.NewServer()
	protomsg.RegisterGRPC_SERVICEServer(rpc, &GRPC_Service{})
	log.Infof("started gRPC port:%v",8028)
	if e := rpc.Serve(lis);e != nil{
		log.Errorf("%v ",e)
	}
}