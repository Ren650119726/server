package network

import (
	"root/core/log"
	"github.com/gin-gonic/gin"
	"os"
)

/* web server */
type WebServer struct {
	router *gin.Engine
}

// 创建一个webserver
func NewWebServer(runmode string) *WebServer {
	web := &WebServer{}
	web.router = gin.New()
	gin.SetMode(runmode)
	return web
}

/* 启动 http server */
func (self *WebServer) RunHttpServer(url string) {
	go func() {
		defer func() {
			if e := recover(); e != nil {
				log.Error("recover error", e)
			}
		}()

		if err := self.router.Run(url); err != nil {
			os.Exit(1)
		}
	}()
}

// 启动httpsserver
func (self *WebServer) RunHttpsServer(url, certfile, keyfile string) {
	go func() {
		defer func() {
			if e := recover(); e != nil {
				log.Error("recover error", e)
			}
		}()

		if err := self.router.RunTLS(url, certfile, keyfile); err != nil {
			os.Exit(2)
		}
	}()
}

// 注册url接口
func (self *WebServer) RegisteUrlAPI(apiname string, handler gin.HandlerFunc) {
	self.router.Any(apiname, handler)
}
