package utils

/* 网络固定消息包ID常量定义 */
const (

	/*******************************actor内部消息id*************************************/
	ID_CLIENT_CONNECT = uint16(1) // client连接到服务器
	ID_DISCONNECT     = uint16(2) // client断开连接
	ID_KICK_CLIENT    = uint16(3) // 剔除client
	ID_HEARTBEAT      = uint16(4) // 心跳
	ID_ACTOR_EXIT     = uint16(5) // actor退出
	ID_HEART_CLOSE    = uint16(6) // actor退出
)

/* 常量定义 */
const (
	MAX_ACTOR_NUMBER = 100
)
