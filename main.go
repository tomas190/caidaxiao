package main

import (
	"caidaxiao/conf"
	"caidaxiao/game"
	"caidaxiao/gate"
	"fmt"
	"github.com/name5566/leaf"
	lconf "github.com/name5566/leaf/conf"
)

func main() {
	test()
	lconf.LogLevel = conf.Server.LogLevel
	lconf.LogPath = conf.Server.LogPath
	lconf.LogFlag = conf.LogFlag
	lconf.ConsolePort = conf.Server.ConsolePort
	lconf.ProfilePath = conf.Server.ProfilePath

	leaf.Run(
		game.Module,
		gate.Module,
	)
}

func test()  {
	var a int64 = 9
	fmt.Println("test")
	fmt.Println(uint16(a))
}