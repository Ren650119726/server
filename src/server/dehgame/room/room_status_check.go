package room

import (
	"root/core/packet"
	"root/server/dehgame/account"
)

type IDEHStatus_universal interface {
	CanQuit(accId uint32) bool
	CombineMSG(packet packet.IPacket, acc *account.Account)
	ShowCard(player *GamePlayer, show_self bool) packet.IPacket
}
