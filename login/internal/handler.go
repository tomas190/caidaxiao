package internal

import (
	"reflect"
)

// protocbuf結構體，透過router註冊給LoginModule
func init() {
	// 範例Login_handler(&msg.Login{}, Login) //註冊一個結構體所對應的方法

}

func Login_handler(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}
