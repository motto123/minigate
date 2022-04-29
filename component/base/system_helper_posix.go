//go:build linux || darwin || unix
// +build linux darwin unix

/*
posix类操作系统的实现
*/

package base

import (
	"os"
	"runtime"
	"syscall"
)

func daemonize() int {
	var ret, ret2 uintptr
	var err syscall.Errno

	darwin := runtime.GOOS == "darwin"

	// already a daemon
	if syscall.Getppid() == 1 || darwin {
		return 0
	}

	// fork off the parent process
	ret, ret2, err = syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)
	if err != 0 {
		return -1
	}

	// failure
	if ret2 < 0 {
		os.Exit(-1)
	}

	// handle exception for darwin
	if darwin && ret2 == 1 {
		ret = 0
	}

	// if we got a good PID, then we call exit the parent process.
	if ret > 0 {
		os.Exit(0)
	}
	/* Change the file mode mask */
	_ = syscall.Umask(0)

	// create a new SID for the child process
	s_ret, _ := syscall.Setsid()

	if s_ret < 0 {
		return -1
	}

	return 0
}

func flock(fd int) error {
	return syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)
}

//从路径中获取文件名
func base(path string) string {
	return baseimpl(path, "/")
}

//从路径中获取目录名
func dir(path string) string {
	return dirimpl(path, "/")
}

//获取路径分割符
func getPathDel() string {
	return "/"
}
