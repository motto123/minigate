package main

import (
	"com.minigame.auth/server"
	"com.minigame.component/base"
)

func main() {
	fw := base.NewFrameWork()
	fw.SetServer(server.Srv)
	fw.Run()
}
