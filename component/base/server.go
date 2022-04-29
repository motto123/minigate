package base

type Server interface {
	Init(fw *Framework) error
	MainLoop()
	OnReload() // 重载
	OnExit()   // 退出
}
