package rediskey

import "fmt"

const (
	BaseData = "BaseData"
)

func PlayerId(playerId uint32) string {
	return fmt.Sprintf("BaseData:players:%v", playerId)
}

func ClubId(clubId uint32) string {
	return fmt.Sprintf("ClubData:club:%v", clubId)
}
