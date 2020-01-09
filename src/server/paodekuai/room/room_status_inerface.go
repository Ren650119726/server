package room

import (
	"root/core/packet"
	"root/server/paodekuai/account"
)

type RoomStatusExInterface interface {
	BulidPacket(tPacket packet.IPacket, tAccount *account.Account)
}
