package main

import (
	"caidaxiao/conf"
	"caidaxiao/game"
	"caidaxiao/gate"
	"caidaxiao/login"

	"github.com/name5566/leaf"
	lconf "github.com/name5566/leaf/conf"
)

func main() {
	lconf.LogLevel = conf.Server.LogLevel
	lconf.LogPath = conf.Server.LogPath
	lconf.LogFlag = conf.LogFlag
	lconf.ConsolePort = conf.Server.ConsolePort
	lconf.ProfilePath = conf.Server.ProfilePath

	leaf.Run(
		login.Module,
		game.Module,
		gate.Module,
	)
}
