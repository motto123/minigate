package base

import (
	"com.minigame.component/log"
	"fmt"
)

//Framework 管理server的框架,打印panic信息到日志
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
	log.Infof("Server", "=========== end ============")

	// 调用服务主循环函数
	GoMain(fw.server.MainLoop)

}
