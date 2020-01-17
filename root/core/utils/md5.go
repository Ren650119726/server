package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

const LoginKey = "PcdJeO5il1iCGu9I"

func GetLoginMd5Token(ucaddr string, userid, tick int64) string {
	org := fmt.Sprintf("LoginKey=%s&ucaddr=%s&userid=%d&tick=%d", LoginKey, ucaddr, userid, tick)
	md5obj := md5.New()
	md5obj.Write([]byte(org))
	new_token := hex.EncodeToString(md5obj.Sum(nil))
	return new_token
}
