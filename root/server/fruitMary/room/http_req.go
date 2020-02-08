package room

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"root/core"
	"root/core/log"
	"strconv"
)

func asyn_addMoney(unique string,num int64,roomID int32, back func(backunique string,backmoney int64))  {
	go func() {
		send := url.Values{"channelId": {"DDHYLC"},
			"gameId": {"fruitMary"},
			"userId": {unique},
			"num": {strconv.Itoa(int(num))},
			"desc": {fmt.Sprintf("水果小玛利请求下注:%v",num)},
		}
		resp, err := http.PostForm("http://47.244.119.129:1000/user/updateGold",
			send)
		log.Infof("小玛利请求下注:%v",send)

		if err != nil {
			log.Errorf("三方平台，http 请求错误:%v", err.Error())
			return
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