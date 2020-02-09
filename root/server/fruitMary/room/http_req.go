package room

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"root/common/config"
	"root/core"
	"root/core/log"
	"strconv"
	"time"
)

func asyn_addMoney(unique string,num int64,roomID int32, desc string,back func(backunique string,backmoney int64),errback func())  {
	go func() {
		send := url.Values{"channelId": {"DDHYLC"},
			"gameId": {"fruitMary"},
			"userId": {unique},
			"num": {strconv.Itoa(int(num))},
			"desc": {desc},
		}
		resp, err := http.PostForm(config.ALTERUSERGOLD_URL,
			send)
		log.Infof("小玛利请求下注:%v",send)

		if err != nil {
			log.Errorf("三方平台，http 请求错误:%v", err.Error())
			for i := 0;i < 10;i++{
				time.Sleep(1*time.Second)
				resp, err = http.PostForm(config.ALTERUSERGOLD_URL,
					send)
				if err == nil {
					break
				}
			}

			// 10次请求以后，如果还有错，就直接返回了
			if err != nil {
				if errback != nil {
					core.LocalCoreSend(0,roomID, func() {
						errback()
					})
				}
				return
			}
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("三方平台，read 错误:%v", err.Error())
			return
		}
		log.Infof("小玛利请求下注，平台返回:%v",string(body))
		var jsonstr map[string]interface{}
		e := json.Unmarshal(body,&jsonstr)
		if e != nil {
			log.Errorf("json 解析错误:%v ",e.Error())
			return
		}
		if err,e := jsonstr["status"];e && int(err.(float64)) != 0{
			log.Errorf("平台返回错误码:%v ",int(err.(float64)))
			return
		}else{
			data := jsonstr["data"].(map[string]interface{})
			gold := data["gold"].(float64)
			if back != nil {
				core.LocalCoreSend(0,roomID, func() {
					back(unique,int64(gold))
				})
			}
		}
	}()
}