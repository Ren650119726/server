package types

type EActorType byte

// 服务器类别定义
const (
	EActorType_Unknown EActorType = 0
	EActorType_MYSQL   EActorType = 20 // mysql逻辑actor
	EActorType_REDIS   EActorType = 30 // redis逻辑actor
	EActorType_LOG     EActorType = 40 // log逻辑actor
)

var typeStringify = [...]string{
	EActorType_Unknown: "Unknown",
	EActorType_MYSQL:   "mysql",
	EActorType_REDIS:   "redis",
	EActorType_LOG:     "log",
}

func (e EActorType) String() string {
	return typeStringify[e]
}

func (e EActorType) Int32() int32 {
	return int32(e)
}
