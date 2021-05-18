package gate

import (
	"caidaxiao/game"
	"caidaxiao/msg"
)

func init() {
	msg.Processor.SetRouter(&msg.Ping{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.Login_C2S{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Logout_C2S{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.JoinRoom_C2S{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.LeaveRoom_C2S{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.PlayerAction_C2S{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.BankerData_C2S{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.EmojiChat_C2S{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.ShowTableInfo_C2S{}, game.ChanRPC)
}
