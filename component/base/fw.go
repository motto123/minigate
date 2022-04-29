package base

import (
	"fmt"
)

type Framework struct {
	server Server
}

func NewFrameWork() *Framework {
	return new(Framework)
}

func (fw *Framework) SetServer(s Server) {
	fw.server = s
}

func (fw *Framework) Run() {
	// 初始化服务接口
	err := fw.server.Init(fw)
	if err != nil {
		str := fmt.Sprintf("err: %+v", err)
		panic(str)
	}

	// 调用服务主循环函数
	GoMain(fw.server.MainLoop)

	println("------fw run")
}
