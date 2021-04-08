package internal

import (
	"caidaxiao/base"
	"github.com/name5566/leaf/module"
)

var (
	skeleton = base.NewSkeleton()
	ChanRPC  = skeleton.ChanRPCServer

	hall = NewHall()

	c4c = &Conn4Center{}
)

type Module struct {
	*module.Skeleton
}

func (m *Module) OnInit() {
	m.Skeleton = skeleton

	// 初始连接数据库  //todo
	InitMongoDB()

	// 大厅初始化
	hall.Init()

	// 中心服初始化并创建链接
	c4c.Init()
	c4c.CreatConnect()

	// 监听接口
	go StartHttpServer()

	test()
}

func (m *Module) OnDestroy() {

}
