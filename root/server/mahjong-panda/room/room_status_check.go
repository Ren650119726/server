package room

import (
	"root/core/packet"
	"root/server/mahjong-panda/account"
)

type IPANDAStatus_universal interface {
	CombineMSG(packet packet.IPacket, acc *account.Account)
	SaveQuit(accid uint32) bool
}
type IPANDAStatus_Game_universal interface {
	Combine_Game_MSG(packet packet.IPacket, acc *account.Account)
}
