package server

import (
	"context"
	"google.golang.org/grpc"
	"net"
	"root/common"
	"root/common/model/inst"
	"root/common/tools"
	"root/core"
	"root/core/log"
	"root/protomsg"
)
type MysqlServer struct {}



func (s *MysqlServer)GetAccount(cxt context.Context,req *protomsg.GetAccountReq) ( *protomsg.AccountStorageData, error)  {
	rebackChannel := make(chan bool)
	retdata := &protomsg.AccountStorageData{}
	var err error
	core.LocalCoreSend(0,common.EActorType_MAIN.Int32(), func() {
		mod := &inst.AccountModel{}
		mod.AccountId = req.AccountID
		err = mod.GetAccount()
		if err != nil {
			return
		}
		retdata = &protomsg.AccountStorageData{}
		tools.CopyProtoData(mod, retdata) // 将grom model数据转换成proto数据
		log.Infof("获取数据 :%v",retdata.String())

		defer func() {
			rebackChannel <- true
		}()
	})
	<- rebackChannel
	return retdata,nil
}
func GRPC_SERVER(){
	lis, err := net.Listen("tcp", ":8028")  //监听所有网卡8028端口的TCP连接
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}
	rpc := grpc.NewServer()
	protomsg.RegisterMySQLServerServer(rpc, &MysqlServer{})
	log.Infof("started gRPC port:%v",8028)
	if e := rpc.Serve(lis);e != nil{
		log.Errorf("%v ",e)
	}
}