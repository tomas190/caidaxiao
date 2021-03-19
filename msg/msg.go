package msg

import (
	"fmt"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/network/protobuf"
	"sort"
)

// 使用默认的 Json 消息处理器 (默认还提供了 ProtoBuf 消息处理器)
var Processor = protobuf.NewProcessor()

func init() {
	test2()
	log.Debug("msg init ~~~")
	Processor.Register(&Ping{})
	Processor.Register(&Pong{})
	Processor.Register(&Login_C2S{})
	Processor.Register(&Login_S2C{})
	Processor.Register(&Logout_C2S{})
	Processor.Register(&Logout_S2C{})
	Processor.Register(&JoinRoom_C2S{})
	Processor.Register(&JoinRoom_S2C{})
	Processor.Register(&EnterRoom_S2C{})
	Processor.Register(&LeaveRoom_C2S{})
	Processor.Register(&LeaveRoom_S2C{})
	Processor.Register(&ActionTime_S2C{})
	Processor.Register(&PlayerAction_C2S{})
	Processor.Register(&PlayerAction_S2C{})
	Processor.Register(&PotChangeMoney_S2C{})
	Processor.Register(&UptPlayerList_S2C{})
	Processor.Register(&ResultData_S2C{})
	Processor.Register(&BankerData_C2S{})
	Processor.Register(&BankerData_S2C{})
	Processor.Register(&EmojiChat_C2S{})
	Processor.Register(&EmojiChat_S2C{})
	Processor.Register(&SendActTime_S2C{})
}


// 结构体定义
type test struct {
	value int
	str   string
}

func test2() {
	s := make([]test, 7)
	s[0] = test{value: 2, str: "2021-03-19 18:15:10"}
	s[1] = test{value: 4, str: "2021-03-19 18:16:10"}
	s[2] = test{value: 1, str: "2021-03-19 18:17:10"}
	s[3] = test{value: 5, str: "2021-03-19 18:18:10"}
	s[4] = test{value: 3, str: "2021-03-19 18:19:10"}
	s[5] = test{value: 3, str: "2021-03-19 19:19:10"}
	s[6] = test{value: 3, str: "2021-03-19 19:20:10"}
	fmt.Println("初始化结果:")
	fmt.Println(s)

	// 从小到大排序(不稳定排序)
	//sort.Slice(s, func(i, j int) bool {
	//	if s[i].value < s[j].value {
	//		return true
	//	}
	//	return false
	//})


	// 从小到大排序(稳定排序)
	//sort.SliceStable(s, func(i, j int) bool {
	//	if s[i].str < s[j].str {
	//		return true
	//	}
	//	return false
	//})
	//fmt.Println("\n从小到大排序结果:")
	//fmt.Println(s)

	//// 是否从小到大排序
	//bLess := sort.SliceIsSorted(s, func(i, j int) bool {
	//	if s[i].value < s[j].value {
	//		return true
	//	}
	//	return false
	//})
	//fmt.Printf("数组s是否从小到大排序,bLess:%v\n", bLess)
	//
	// 从大到小排序(不稳定排序)
	sort.Slice(s, func(i, j int) bool {
		if s[i].str > s[j].str {
			return true
		}
		return false
	})
	fmt.Println("\n从大到小排序结果:")
	fmt.Println(s)
	//
	//// 是否从大到小排序
	//bMore := sort.SliceIsSorted(s, func(i, j int) bool {
	//	if s[i].value > s[j].value {
	//		return true
	//	}
	//	return false
	//})
	//fmt.Printf("数组s是否从大到小排序,bMore:%v\n", bMore)

}
