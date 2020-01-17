package room

import (
	"root/core/packet"
	"root/server/mahjong-dgk/account"
)

type IDGKStatus_universal interface {
	CombineMSG(packet packet.IPacket, acc *account.Account)
	SaveQuit(accid uint32) bool
}
type IDGKStatus_Game_universal interface {
	Combine_Game_MSG(packet packet.IPacket, acc *account.Account)
}
