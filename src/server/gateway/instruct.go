package main

import (
	"root/common"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/network"
)

func init() {
	core.Cmd.Regist("help", h, true)
	core.Cmd.Regist("check", check, true)
	core.Cmd.Regist("test", test, true)
}

// 创建房间
func h([]string) {
	log.Infof("help: \n" +
		"al 检测所有连接映射，分析是否有异常 \n" +
		"test 测试和服务器之间的链接是否能正常建立")
}

func check(sParam []string) {
	log.Infof("开始检查...")
	// 检查两个映射表是否有不对应的情况
	actor_map := make(map[int64]bool, 0)
	count := 0
	allcount := 0
	for k, g := range G.cmap {
		allcount++
		s_actor := g.s_actorId
		if s_actor == -1 {
			continue
		}
		actor_map[s_actor] = true
		count++

		_, exit := G.smap[s_actor]
		if !exit {
			log.Warnf(colorized.Cyan("出现不匹配的情况，cmap中存在，smap中不存在 映射组:%v"), *g)
		} else {
			actor := core.GetActor(int32(g.s_actorId))
			client, _ := actor.Handler.(*network.TCPClient)
			if client == nil {
				log.Warnf("映射服务器 异常！！！！！")
			} else {
				log.Infof(colorized.White("cmap key:%v c_session[%v]:[%v]  >>>>  s_actorId[%v]:[%v]"), k, g.c_session, core.GetRemoteIP(g.c_session), g.s_actorId, client.Remote())
			}
		}
	}

	// 检查smap表，查看是否有多余的
	for actorId, g := range G.smap {
		if actor_map[actorId] == false {
			log.Warnf(colorized.Cyan("smap 中存在 cmap 没有的映射组 actorId:%v g:%v"), actorId, *g)
		} else {
			log.Warnf(colorized.Gray("smap key:%v g:%v 状态正常"), actorId, *g)
		}
	}

	log.Infof("完成检查 有效连接:%v 所有连接:%v", count, allcount)
}

func test(sParam []string) {
	//conIP := beego.AppConfig.DefaultString(core.Appname+"::listen", "")
	var actor *core.Actor
	connect_actor := network.NewTCPClient(G_ACTOR,
		func() string {
			return core.CoreAppConfString("listen")
		},
		func() {
			if actor != nil {
				actor.Suspend()
			}
		})

	actor = core.NewActor(int32(common.EActorType_CONNECT_HALL), connect_actor, make(chan core.IMessage, 1000))
	core.CoreRegisteActor(actor)
}
