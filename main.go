package main

import (
	"caidaxiao/conf"
	"caidaxiao/game"
	"caidaxiao/gate"
	"caidaxiao/login"
	"time"

	"github.com/name5566/leaf"
	lconf "github.com/name5566/leaf/conf"
)

func main() {
	lconf.LogLevel = conf.Server.LogLevel
	lconf.LogPath = conf.Server.LogPath
	lconf.LogFlag = conf.LogFlag
	lconf.ConsolePort = conf.Server.ConsolePort
	lconf.ProfilePath = conf.Server.ProfilePath
	time.Sleep(3 * time.Second) // 延迟三秒启动(关闭服务时会等待三秒确认与中心服后续任物完成才断开ws)
	leaf.Run(

		login.Module,
		game.Module,
		gate.Module,
	)
}
