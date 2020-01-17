package tools

import (
	"root/core/log"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/golang/protobuf/proto"
)

const EnterGameKey = "PcdJeO5il1iCGuf323f239I"

func MarshalToken(src proto.Message) string {
	bytes, err := proto.Marshal(src)
	if err != nil {
		log.Error("Marshal pb 格式错误: ", src.String(), " error: ", err.Error())
		return ""
	}
	return base64.StdEncoding.EncodeToString(bytes)
}

func UnmarshalToken(token string, dst proto.Message) proto.Message {
	tokenData, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		log.Warn("UnmarshalToken error, ", err.Error())
		return nil
	}

	err = proto.Unmarshal(tokenData, dst)
	if err != nil {
		log.Error(" UnmarshalToken pb 错误 ", string(tokenData), err.Error())
		return nil
	}
	return dst
}

// 玩家进入游戏服的签名信息
func MarshalEnterGameSign(roomId int64, playerId, timestamp int64) string {
	str := fmt.Sprintf("%s%v%v%v", EnterGameKey, roomId, playerId, timestamp)
	return MD5(str)
}

func MD5(str string) string {
	md5obj := md5.New()
	md5obj.Write([]byte(str))
	return hex.EncodeToString(md5obj.Sum(nil))
}
