package main

import (
	"root/common/config"
	"root/core"
	"root/core/log"
	"root/core/network"
	"root/core/utils"
	"fmt"
	"os"
	"sort"
)

type (
	logic struct {
		owner *core.Actor
	}
)

func NewLogic() *logic {

	return &logic{}
}

func (self *logic) Init(actor *core.Actor) bool {
	self.owner = actor

	argc := len(os.Args)
	if argc < 3 {
		panic(fmt.Sprintf("param至少需要3个参数; param:%v", os.Args))
	}
	strListName := os.Args[1]
	if strListName == "dd" {
		conf := config.GetPublicConfig_String("DD_ALL_NODE_LIST")
		self.SetNode(conf)
	} else if strListName == "hh" {
		conf := config.GetPublicConfig_String("HH_ALL_NODE_LIST")
		self.SetNode(conf)
	} else {
		panic(fmt.Sprintf("启动exe需传递节点IP名(dd or hh); param:%v", os.Args))
	}
	return true
}

func (self *logic) Stop() {

}

func (self *logic) HandleMessage(actor int32, msg []byte, session int64) bool {

	return true
}

func (self *logic) SetNode(nodestr string) {
	log.Infof("node:%v ", nodestr)
	m := utils.SplitConf2Mapis(nodestr)

	enable_node := make(map[int]string)
	for i, ip := range m {
		node := i
		actorID := node * 1000
		remoteIP := fmt.Sprintf("%v:50000", ip)
		connectDB_actor := network.NewTCPClient(self.owner, func() string {
			return remoteIP
		}, func() {
			enable_node[node] = remoteIP
			a := core.GetActor(int32(actorID))
			if a != nil {
				a.Suspend()
			}

		})
		nodeActor := core.NewActor(int32(actorID), connectDB_actor, make(chan core.IMessage, 10000))
		core.CoreRegisteActor(nodeActor)
	}

	self.owner.AddTimer(3000, 1, func(dt int64) {
		sNetworkFailure := make([]int, 0)
		sNetworkOK := make([]int, 0)
		log.Infof("ping node ...")
		for i := range m {
			if _, e := enable_node[i]; !e {
				sNetworkFailure = append(sNetworkFailure, i)
			} else {
				sNetworkOK = append(sNetworkOK, i)
			}
			a := core.GetActor(int32(i * 1000))
			if a != nil {
				a.Suspend()
			}
		}

		sort.Slice(sNetworkFailure, func(i, j int) bool {
			if sNetworkFailure[i] < sNetworkFailure[j] {
				return true
			}
			return false
		})
		for _, nID := range sNetworkFailure {
			log.Infof("!!!!! %v 号节点无法连接 !!!!!", nID)
		}
		log.Infof("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		sort.Slice(sNetworkOK, func(i, j int) bool {
			if sNetworkOK[i] < sNetworkOK[j] {
				return true
			}
			return false
		})
		for _, nID := range sNetworkOK {
			log.Infof("!!!!! %v 号节点连接正常 !!!!!", nID)
		}
		log.Infof("ping node over")
	})
}
