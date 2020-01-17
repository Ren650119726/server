package main

import (
	"github.com/astaxie/beego"
	"golang.org/x/net/websocket"
	"net"
	"net/url"
	"root/common"
	"root/core"
)

type Client struct {
	Host string
	Path string
	ws net.Conn
	quit chan bool
}

func NewWebsocketClient(host, path string) *Client {
	return &Client{
		Host: host,
		Path: path,
		quit: make(chan bool),
	}
}

func (this *Client) SendMessage(body []byte) error {
	_, err := this.ws.Write(body)
	if err != nil {
		beego.Error(err)
		return err
	}

	return nil
}
func (this *Client) connect() error {
	u := url.URL{Scheme: "ws", Host: this.Host, Path: this.Path}

	ws, err := websocket.Dial(u.String(), "", "http://"+this.Host+"/")
	this.ws = ws
	if err != nil {
		beego.Error(err)
		return err
	}
	return nil
}

func main()  {
	// 创建server
	lo := NewLogic()
	msgchan := make(chan core.IMessage, 10000)
	actor := core.NewActor(common.EActorType_MAIN.Int32(), lo, msgchan)
	core.CoreRegisteActor(actor)

	core.CoreStart()
}