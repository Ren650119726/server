package common

type EActorType byte

// 服务器类别定义
const (
	EActorType_MAIN         EActorType = 1 // 主线程
	EActorType_SERVER       EActorType = 2 // 监听所有client的actor
	EActorType_CONNECT_HALL EActorType = 3 // 连接Hall
	EActorType_CONNECT_DB   EActorType = 4 // 连接DB
	EActorType_REDIS        EActorType = 5 // redis
	EActorType_CONNECT_DB2  EActorType = 6 //  连接DB2
)

var typeStringify = [...]string{
	EActorType_SERVER:       "Server",
	EActorType_CONNECT_HALL: "Connect_hall",
	EActorType_CONNECT_DB:   "Connect_db",
	EActorType_REDIS:        "redis",
}

func (e EActorType) String() string {
	return typeStringify[e]
}

func (e EActorType) Int32() int32 {
	return int32(e)
}
