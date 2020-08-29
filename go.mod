module root

go 1.15

require (
	github.com/360EntSecGroup-Skylar/excelize v1.4.1
	github.com/astaxie/beego v1.12.0
	github.com/gin-gonic/gin v1.4.0
	github.com/go-redis/redis v6.15.6+incompatible
	github.com/go-sql-driver/mysql v1.4.1
	github.com/golang/protobuf v1.4.1
	github.com/google/gopacket v1.1.17
	github.com/jinzhu/gorm v1.9.11
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mattn/go-isatty v0.0.8 // indirect
	github.com/shiena/ansicolor v0.0.0-20151119151921-a422bbe96644 // indirect
	github.com/yuin/gopher-lua v0.0.0-20190514113301-1cd887cd7036
	golang.org/x/net v0.0.0-20190503192946-f4e77d36d62c
	google.golang.org/grpc v1.27.0
	google.golang.org/protobuf v1.25.0
)

replace google.golang.org/grpc => github.com/grpc/grpc-go v1.31.1
