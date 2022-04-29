package main

import (
	"com.example.client/client"
	"com.minigame.component/log"
)

func main() {
	log.InitLogger("a", "./")
	ci, err := client.NewClient("127.0.0.1", "6601")
	if err != nil {
		panic(err)
	}
	ci.Do()
}
