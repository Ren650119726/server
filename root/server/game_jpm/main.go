package main

import (
	"root/common"
	"root/core"
	"root/core/db"
	"root/server/game_hongbao/logic"
)

func greed(prices []int, fee int) int {
	preBuy := prices[0]
	lastSell := 0
	total := 0
	n := len(prices)
	for i := 1; i < n; i++ {
		c := prices[i] - prices[i-1]
		// 1 0 2 -fee 3   // 3 可以买入. 2 不值得买再，继续观望. 1 更好的卖出机会
		if c >= 0 { // 更好的卖出机会
			lastSell = max(lastSell, prices[i]-preBuy-fee)
		} else if c <= -fee { // 可以确定卖出，并且重新买入
			total += lastSell // 买入前，先累计盈利
			preBuy = prices[i]
			lastSell = 0 // 买入后，重置卖出
		} else { // 更好的买入机会
			preBuy = min(preBuy, prices[i])
		}
	}
	total += lastSell
	return total
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	greed([]int{2, 2, 1, 1, 5, 5, 3, 1, 5, 4}, 2)
	// 创建server
	hb := logic.Newjpm()
	msgchan := make(chan core.IMessage, 10000)
	actor := core.NewActor(common.EActorType_MAIN.Int32(), hb, msgchan)
	core.CoreRegisteActor(actor)

	core.CoreRegisteActor(core.NewActor(common.EActorType_REDIS.Int32(), db.NewRedis(), make(chan core.IMessage, 10000)))
	core.CoreStart()
}
