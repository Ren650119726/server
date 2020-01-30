package core

import (
	"context"
	"root/core/log"
	"root/core/utils"
	"fmt"
	"github.com/astaxie/beego"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"
)

// 全局变量
var (
	Gctx        context.Context
	Gwg         *sync.WaitGroup
	gshutdown   context.CancelFunc
	newchan     chan *Actor
	closechan   chan int32
	msg_chan    chan IMessage
	cmd_chan    chan func()
	actor_cache map[int32]*Actor
	logfile     *os.File
	ConfigDir   string
	Appname     string // 该进程的Appname 配置主键
	ScriptDir   string // 脚本路径
	lock        sync.Mutex
	SID         int
)

func init() {
	argc := len(os.Args)
	if argc < 3 {
		panic("param num must 3, but it's less then")
	}
	ConfigDir = os.Args[2]
	coreInit(os.Args[1], ConfigDir+"app.conf")

	// 第4个是脚本路径
	if argc == 4 {
		ScriptDir = os.Args[3]
	}
}

// 初始化主配置
func initConfig(section, path string) {
	// 基础配置(服务器网络相关)
	err := beego.LoadAppConfig("ini", path)
	if err != nil {
		panic(err)
		return
	}
	Appname = section
}

func initLogger() {
	logLv := beego.AppConfig.DefaultString("DEF::level", "debug")

	// 设置日志相关配置
	log.SetLevel(logLv)

	// 通过编译参数控制日志输出
	var outputs []io.Writer

	// 开启了控制台输出
	enableconsole := beego.AppConfig.DefaultBool("DEF::enableconsole", false)
	if enableconsole {
		outputs = append(outputs, os.Stderr)
	}

	directory := beego.AppConfig.DefaultString("DEF::dir", "../log")
	//指定日志文件备份方式为日期的方式
	//第一个参数为日志文件存放目录
	//第二个参数为日志文件命名
	prefix := beego.AppConfig.DefaultString("DEF::prefix", "log")

	logname := beego.AppConfig.DefaultString(fmt.Sprintf("%v::logname", Appname), string(Appname))

	directory = directory + "/" + logname
	os.MkdirAll(directory, os.ModePerm)
	logfilename := fmt.Sprintf("%s_%s_%s.log", prefix, logname, time.Now().Format("2006-01-02"))
	logfilename = filepath.Join(directory, logfilename)
	file, err := os.OpenFile(logfilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		panic(fmt.Errorf("日志文件打开错误, Error=%s", err.Error()))
	}
	outputs = append(outputs, file)
	// 没有输出 则啥也没有
	cnt := len(outputs)
	if cnt <= 0 {
		return
	}
	log.SetOutput(io.MultiWriter(outputs...))
	if logfile != nil {
		logfile.Close()
	}
	logfile = file
}

/* 初始化 日志等 */
func coreInit(section, path string) {
	// 配置初始化
	initConfig(section, path)
	// 初始化时区
	utils.InitLocalTime("Asia/Shanghai")
	utils.ResetTime(0)
	SID = beego.AppConfig.DefaultInt(Appname+"::sid", -1)
	// 日志初始化
	initLogger()

	// 设置运行的cpunum
	cpu := runtime.NumCPU()
	runtime.GOMAXPROCS(cpu)
	log.Infof("当前服务器CPU 数量%v",cpu)

	// actor相关初始化
	newchan = make(chan *Actor, utils.MAX_ACTOR_NUMBER)
	closechan = make(chan int32, utils.MAX_ACTOR_NUMBER)
	msg_chan = make(chan IMessage, utils.MAX_ACTOR_NUMBER*500) // 1个actor 500个应该够了吧
	cmd_chan = make(chan func(), 10)

	// 构建actor map
	actor_cache = make(map[int32]*Actor)
}

// 执行业务逻辑执行函数-内部函数
func coreExecLogic(actor *Actor) {
	Gwg.Add(1)

	utils.Try(func() {
		actor.Handler.Init(actor)
	})

	//log.Debugf("root/coreExecLogic start %v", actor.Id)

	up_timer := time.NewTimer(time.Millisecond * 1)

	// 计算下一个零点
	now := time.Now()
	next := now.Add(time.Hour * 24)
	next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
	t_redirect_log := time.NewTimer(next.Sub(now))

	defer func() {
		up_timer.Stop()
		actor.Handler.Stop()
		closechan <- actor.GetID()
		Gwg.Done()
	}()

	// 执行逻辑
	for {
		select {
		case <-Gctx.Done():
			return
		case <-t_redirect_log.C:
			initLogger()
			t_redirect_log.Reset(time.Hour * 24)
		default:
			// 处理消息
			select {
			case <-Gctx.Done():
				return
			case <-up_timer.C:
				actor.TimerMgr.Update(utils.MilliSecondTime())
				up_timer.Reset(time.Millisecond * 1)
				break
			case msg := <-actor.MessageCache:

				l := len(actor.MessageCache)
				if l > 1000 {
					c := cap(actor.MessageCache)
					log.Warnf("actor:%v 目前堆积消息:%v 条 capacity:%v !!!!!!", actor.Id, l, c)
				}

				switch message := msg.(type) {
				case *CoreMessage:
					utils.Try(func() {
						actor.Handler.HandleMessage(message.Source, message.Data, message.Session) // 一次只处理一条消息

					})
				case *LocalMessage:
					utils.Try(func() { message.FunHandler() })
				default:
					return
				}
				break
			}
		}

		// 逻辑层关闭了
		if actor.IsSuspend {
			break
		}
	}
}

func CoreAppConfString(key string) string {
	lock.Lock()
	defer func() {
		lock.Unlock()
	}()
	return beego.AppConfig.DefaultString(Appname+"::"+key, "")
}

// 注册actor
func CoreRegisteActor(actor *Actor) {
	newchan <- actor
}

func GetActor(Id int32) *Actor {
	return actor_cache[Id]
}

/* 开始执行 阻塞调用 */
func CoreStart() {
	// 启动后即阻塞
	Gwg = &sync.WaitGroup{}
	Gctx, gshutdown = context.WithCancel(context.Background())

	// 启动后即阻塞
	defer func() {
		// 通知所有go退出
		gshutdown()
		Gwg.Wait()
	}()

	// 信号处理
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	// 主循环处理
	for {
		select {
		case sys_signal := <-ch:
			switch sys_signal {
			case syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT:
				return
			}
		case actor := <-newchan:
			actor_cache[actor.GetID()] = actor // 新的actor
			go coreExecLogic(actor)
		case id := <-closechan:
			delete(actor_cache, id)
		default:
			select {
			case sys_signal := <-ch:
				switch sys_signal {
				case syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT:
					return
				}
			case actor := <-newchan:
				actor_cache[actor.GetID()] = actor // 新的actor
				go coreExecLogic(actor)
			case id := <-closechan:
				delete(actor_cache, id)
			case msg := <-msg_chan:
				targetId := int32(0)
				switch message := msg.(type) {
				case *CoreMessage: // 主要用于处理带msg和data的消息
					targetId = message.Target
				case *LocalMessage: // 主要用于方便线程间进行闭包逻辑处理
					targetId = message.Target
				}

				if actor, ok := actor_cache[targetId]; ok {
					actor.Push(msg)
				} else {
					msg_chan <- msg // 再还回去
				}
			case cmd := <-cmd_chan:
				cmd()
			}
		}
	}
}

/* Core之间发送消息 */
func CoreSend(sourceid int32, targetid int32, data []byte, session int64) {
	msg := &CoreMessage{Source: sourceid, Target: targetid, Session: session, Data: data}
	msg_chan <- msg
}

/* Core之间发送消息 */
func LocalCoreSend(sourceid int32, targetid int32, f func()) {
	msg := &LocalMessage{Source: sourceid, Target: targetid, FunHandler: f}
	// 组织一个message
	msg_chan <- msg
}

func Exe_cmd(f func()) {
	// 组织一个message
	cmd_chan <- f
}
