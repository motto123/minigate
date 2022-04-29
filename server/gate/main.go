package main

import (
	"com.minigame.component/base"
	"com.minigame.server.gate/server"
)

func main() {
	fw := base.NewFrameWork()
	fw.SetServer(server.Srv)
	fw.Run()
}
