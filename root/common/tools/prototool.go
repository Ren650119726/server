package tools

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"reflect"
	"root/core/log"
	"root/core/packet"
)

// 通过json 把相同的字段，从src中拷贝到dst中
func CopyProtoData(src proto.Message, dst proto.Message) proto.Message {
	bytes, err := json.Marshal(src)
	if err != nil {
		log.Errorf("Marshal pb json格式错误:%v", src.String())
		return nil
	}
	err = json.Unmarshal(bytes, dst)
	if err != nil {
		log.Errorf(" Unmarshal pb json格式错误 [%v]  %v", string(bytes), dst.String())
		return nil
	}
	return dst
}

// 将proto 组装成packet
func Proto2ByteStream(src interface{}) packet.IPacket {

	t := reflect.TypeOf(src)
	//v := reflect.ValueOf(src)

	for i := 0; i < t.NumField()-3; i++ {

		//value := v.Field(i).Interface()
		//fmt.Println(t.Field(i).Name, t.Field(i).Type)

		switch t.Field(i).Type.Kind() {
		//case reflect.Int32:
		//	fmt.Println(t.Field(i).Name, reflect.Int32)
		//case reflect.Uint32:
		//	fmt.Println(t.Field(i).Name, reflect.Uint32)
		//case reflect.String:
		//	fmt.Println(t.Field(i).Name, reflect.String)
		//case reflect.Struct:
		//	fmt.Println(t.Field(i).Name, reflect.Struct)
		//case reflect.Slice:
		//	fmt.Println(t.Field(i).Name, reflect.Slice)
		default:
			fmt.Println(t.Field(i).Name, t.Field(i).Type.Kind())
		}

	}

	return nil
}
