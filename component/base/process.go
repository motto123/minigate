package base

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

type Process struct {
	Name          string //进程名
	Version       string //版本描述
	DefaultRunDir string //默认运行目录
	DefaultLogDir string //默认日志目录
	DefaultCfgDir string //默认配置目录
	inited        bool   //是否已经初始化过
	otherTypeRun  bool
}

//初始化实例
//参数defaultLogDir:默认日志路径(传入空值，由组件自动设置，值应该为相对运行路径的相对路径)
//参数defaultCfgDir:默认配置路径(传入空值，由组件自动设置，值应该为相对运行路径的相对路径)
//参数defaultDeployDir:默认部署路径(传入空值，由组件自动设置，值应该为相对运行路径的相对路径)
//参数cmd:响应信号和http命令的接口
func (p *Process) Initialize(defaultLogDir string, defaultCfgDir string) error {

	//判断是否已经初始化过
	if p.inited {
		return errors.New("already inited")
	}

	//解析程序名和路径
	n, path := GetProcessNameAndPath()
	//_ = path

	//设置程序名
	//特殊处理通过goland运行的进程名字
	p.Name = n

	//设置版本号
	p.Version = ProcessVersion()

	//设置默认运行路径
	p.DefaultRunDir = path
	p.DefaultLogDir = defaultLogDir
	p.DefaultCfgDir = defaultCfgDir

	if strings.Contains(p.DefaultRunDir, "/tmp/go-build") {
		p.runByGoCommand()
	} else if strings.Contains(p.DefaultRunDir, "/tmp/GoLand") {
		p.runByGoland()
	} else {
		p.runByDefault()
	}

	//标记为已经初始化过
	p.inited = true
	return nil
}

func (p *Process) runByDefault() {
	//设置默认日志路径
	if p.DefaultLogDir == "" {
		p.DefaultLogDir = p.DefaultRunDir +
			GetPathDel() + "../../../log" + GetPathDel() + p.Name
	}

	//设置默认配置路径
	if p.DefaultCfgDir == "" {
		p.DefaultCfgDir = p.DefaultRunDir + "./conf"
	}
}

func (p *Process) runByGoland() {
	//特殊处理通过goland运行的进程名字和路径
	// ___go_build_com_minigame_server_gate
	///tmp/GoLand
	split := strings.Split(p.Name, "_")
	if len(split) >= 3 {
		p.Name = split[len(split)-1:][0]
	}
	//设置默认运行路径
	p.DefaultRunDir = "."

	//设置默认日志路径
	if p.DefaultLogDir == "" {
		p.DefaultLogDir = p.DefaultRunDir +
			GetPathDel() + "../../log" + GetPathDel() + p.Name
	}
	//设置默认配置路径
	if p.DefaultCfgDir == "" {
		p.DefaultCfgDir = p.DefaultRunDir + GetPathDel() + "conf"
	}
}

func (p *Process) runByGoCommand() {
	//特殊处理通过go run运行的进程名字和路径
	//main
	///tmp/go-build2496642274/b001/exe

	//设置默认运行路径
	p.DefaultRunDir = "."

	//设置默认日志路径
	if p.DefaultLogDir == "" {
		p.DefaultLogDir = p.DefaultRunDir +
			GetPathDel() + "../../log" + GetPathDel() + p.Name
	}
	//设置默认配置路径
	if p.DefaultCfgDir == "" {
		p.DefaultCfgDir = p.DefaultRunDir + GetPathDel() + "conf"
	}

}

//输出字符串信息
func (p *Process) String() string {
	s := fmt.Sprintf("Name: %s\nVersion: %s\nDefaultRunDir: %s\nDefaultLogDir: %s\nDefaultCfgDir: %s\nDefaultDeployDir: %s\n\n",
		p.Name,
		p.Version,
		p.DefaultRunDir,
		p.DefaultLogDir,
		p.DefaultCfgDir,
	)

	return s
}
