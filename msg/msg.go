package msg

import (
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/network/protobuf"
)

// 使用默认的 Json 消息处理器 (默认还提供了 ProtoBuf 消息处理器)
var Processor = protobuf.NewProcessor()

func init() {
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
	Processor.Register(&ChangeRoomType_S2C{})
}


