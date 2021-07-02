package internal

import (
	common "caidaxiao/base"

	"github.com/name5566/leaf/module"
)

var (
	skeleton = common.NewSkeleton()
	ChanRPC  = skeleton.ChanRPCServer
	c4c      = &Conn4Center{}
)

type Module struct {
	*module.Skeleton
}

func (m *Module) OnInit() {
	m.Skeleton = skeleton
	c4c.S2CS_Conn_init()
	initLoginChanRPC()
	common.GetInstance().SetLoginChanRpc(ChanRPC) //單例模式

}

func (m *Module) OnDestroy() {
	common.Debug_log("loginModule OnDestroy")
	// common.GetInstance().Game.Go("LogoutAllUser")
}
