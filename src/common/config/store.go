package config

import (
	"root/core"
	"root/core/log"
	"encoding/json"
	"io/ioutil"
)

type (
	store struct {
		Price      int `json:"Price"`
		MoneyValue int `json:"MoneyValue"`
	}

	store_type map[int]*store
)

var global_store_config store_type

func init() {
	LoadStore_Conf()
}

func LoadStore_Conf() {
	lock.Lock()
	defer lock.Unlock()

	global_store_config = store_type{}
	data, e := ioutil.ReadFile(core.ConfigDir + "store.json")
	if e != nil {
		log.Errorf("public 错误:%v", e.Error())
		return
	}
	error := json.Unmarshal(data, &global_store_config)
	if error != nil {
		log.Errorf(" error %v", error.Error())
		return
	}
	log.Info("加载完成store.json")
}

// 第一返回值 商品对应的价格RMB
// 第二返回值 商品对应的元宝数量, 13元等于1300返回
func GetStoreConfig(nGoodsID int) (int, int) {
	lock.Lock()
	defer lock.Unlock()

	tNode := global_store_config[nGoodsID]
	if tNode != nil {
		return tNode.Price, tNode.MoneyValue
	}
	return 0, 0
}
