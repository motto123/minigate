package main

import (
	"com.example.client/client"
	"com.minigame.component/log"
)

// main 聊天demo, 实现登录,注册，1对1聊天
func main() {
	log.InitLogger("a", "./")
	ci, err := client.NewClient("127.0.0.1", "6601")
	if err != nil {
		panic(err)
	}
	ci.Do()
}
