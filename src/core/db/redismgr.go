package db

import (
	"root/common"
	"root/core"
	"root/core/log"
	"root/core/log/colorized"
	"root/core/packet"
	"github.com/astaxie/beego"
	"github.com/go-redis/redis"
)

type RedisCon struct {
	*redis.Client
	addr  string
	owner *core.Actor
}

var r *RedisCon

func NewRedis() *RedisCon {
	r = &RedisCon{}
	return r
}

// actor初始化(actor接口定义)
func (self *RedisCon) Init(actor *core.Actor) bool {
	self.owner = actor
	conf_redis := beego.AppConfig.DefaultString(core.Appname+"::redis", "")
	if conf_redis == "" {
		log.Warnf(colorized.Gray("没有配置redis连接"))
		return false
	}

	r.addr = conf_redis

	connect := func() bool {
		con, err := self.dialDefaultServer(r.addr)
		if err != nil {
			log.Errorf("redis connect error :%v", err.Error())
			return false
		}
		r.Client = con
		log.Infof(colorized.Cyan("redis连接成功:%v"), conf_redis)
		return true
	}

	self.owner.AddTimer(5000, -1, func(dt int64) {
		pong, err := r.Client.Ping().Result()
		if err != nil {
			log.Infof("redis reconnect %v %v ", pong, err.Error())
			connect()
		}
	})

	return connect()
}

func (self *RedisCon) dialDefaultServer(addr string) (*redis.Client, error) {
	redis_password := beego.AppConfig.DefaultString(core.Appname+"::redispassword", "")
	if redis_password == "" {
		log.Warnf(colorized.Gray("redis 没有设置密码"))
	}

	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: redis_password,
		DB:       0,
	})

	pong, err := c.Ping().Result()
	log.Infof("%v %v ", pong, err)
	return c, err
}

// 停止回收相关资源
func (self *RedisCon) Stop() {
	self.Client.Close()
}

// actor消息处理
func (self *RedisCon) HandleMessage(actor int32, msg []byte, session int64) bool {
	pack := packet.NewPacket(msg)
	switch pack.GetMsgID() {
	default:
	}
	return true
}

func HSet(key, field string, value interface{}) {
	core.LocalCoreSend(0, common.EActorType_REDIS.Int32(), func() {
		if r.Client == nil {
			return
		}
		r.HSet(key, field, value)
	})
}

func Del(key string) {
	core.LocalCoreSend(0, common.EActorType_REDIS.Int32(), func() {
		if r.Client == nil {
			return
		}
		ret := r.Del(key)
		if ret.Val() != 1 {
			log.Warnf("删除redis 数据失败 key:%v err:%v ", key, ret.Err().Error())
		}
	})
}

func HGetKeyAll(key string, f func(map[string]string)) {
	core.LocalCoreSend(0, common.EActorType_REDIS.Int32(), func() {
		if r.Client == nil {
			f(nil)
			return
		}
		m, e := r.HGetAll(key).Result()
		if e != nil {
			log.Warnf("get redis data error:%v ", e.Error())
			f(nil)
			return
		}
		f(m)
	})
}
