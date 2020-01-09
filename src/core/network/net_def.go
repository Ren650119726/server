package network

/* 常量定义 */
const (
	MAX_SESSION              = 10000
	HANDLE_HEARTBEAT_TIMEOUT = 30              //30s检查一次session，处理心跳超时
	HEARTBEAT_TIMEOUT        = 60              //心跳超时60s
	SEND_HEARTBEAT           = 30              //client发送心跳包时间间隔
	SOCKET_CACHE_SIZE        = 6 * 1024 * 1024 // 发送、接受 缓冲区大小
)

/* 网络层回调函数 非安全的 */
type NetCallBackIF interface {
	handle_input(session int64, data []byte)
}
