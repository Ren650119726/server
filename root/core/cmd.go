package core

import (
	"bufio"
	"root/common"
	"root/core/log"
	"root/core/utils"
	"fmt"
	"github.com/astaxie/beego"
	"os"
	"sort"
	"strings"
)

type (
	callBack struct {
		f    func([]string)
		main bool
	}
	cmd_callback struct {
		callback map[string]*callBack
	}
)

var Cmd = cmd_callback{callback: make(map[string]*callBack, 0)}

func init() {
	go loop()
	Cmd.Regist("actor", actorInfo, false)
	Cmd.Regist("reapp", reapp, false)
	Cmd.Regist("h", Cmd.help, false)
}

func actorInfo(s []string) {
	Exe_cmd(func() {
		for id, a := range actor_cache {
			ato := a
			LocalCoreSend(0, id, func() {
				l := len(ato.MessageCache)
				c := cap(ato.MessageCache)
				log.Infof("actor:%v len:%v cap:%v", ato.Id, l, c)
			})
		}
	})
}
func reapp(s []string) {
	Exe_cmd(func() {
		beego.LoadAppConfig("ini", ConfigDir+"app.conf")
	})
}

func loop() {

	for {
		utils.Try(func() {
			reader := bufio.NewReader(os.Stdin)

			result, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("read error:", err)
			}

			result = result[:len(result)-1]
			args := strings.Split(result, " ")
			cmdName := args[0]
			node := Cmd.callback[cmdName]
			if node == nil {
				log.Printf("无效的命令:%v\r\n", cmdName)
			} else {
				f := node.f
				main := node.main
				if f != nil {
					if main {
						LocalCoreSend(0, common.EActorType_MAIN.Int32(), func() {
							f(args[1:])
						})
					} else {
						f(args[1:])
					}
				} else {
					log.Printf("无效的命令:%v\r\n", cmdName)
				}
			}
		})
	}
}

func (self *cmd_callback) help(sp []string) {
	s := []string{}
	for key, _ := range self.callback {
		s = append(s, key)
	}

	sort.SliceIsSorted(s, func(i, j int) bool {
		return uint32(s[i][0]) < uint32(s[j][0])
	})

	p := "所有命令:\n"
	for _, str := range s {
		p += "   " + str + "\n"
	}
	log.Infof("%v", p)
}

func (self *cmd_callback) Regist(cmd string, f func([]string), deal_main bool) {
	if _, exist := self.callback[cmd]; exist {
		log.Errorf("重复注册命令:%v ", cmd)
		return
	}
	c := &callBack{}
	c.f = f
	c.main = deal_main
	self.callback[cmd] = c
}
