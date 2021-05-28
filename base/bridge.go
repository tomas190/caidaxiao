package base

import (
	"sync"

	"github.com/name5566/leaf/chanrpc"
)

type chanrpcManage struct {
	Game  *chanrpc.Server
	Login *chanrpc.Server
}

var m *chanrpcManage
var once sync.Once

func GetInstance() *chanrpcManage {
	// Debug_log("bridge_chan : GetInstance *chanrpcManage ")
	once.Do(func() {
		m = &chanrpcManage{}
	})
	return m
}

//初始化的时候把下面的方法执行下。
func (m *chanrpcManage) SetGameChanRpc(server *chanrpc.Server) {
	// Debug_log("bridge_chan : *chanrpcManage SetGameChanRpc  ")
	m.Game = server
	// Debug_log("m.Game：", m.Game)
}
func (m *chanrpcManage) SetLoginChanRpc(server *chanrpc.Server) {
	// Debug_log("bridge_chan : *chanrpcManage SetLoginChanRpc  ")
	m.Login = server
	// Debug_log("&m.Login：", m.Login)
}
