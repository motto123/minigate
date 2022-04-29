package log

import (
	"fmt"
	"github.com/pkg/errors"
)

// 初始化2个logger,terminal logger and file logger,可以选一个logger或者两个logger打印日志

const (
	TERM = "TERM"
	FILE = "FILE"
	ALL  = "ALL"
)

var (
	logbaks  map[string]*Logger
	curTag   string
	curLevel Level
)

func init() {
	logbaks = make(map[string]*Logger)
	curTag = TERM
}

func InitLogger(procName string, defaultLogDir string) error {
	// 输出到terminal
	termLog := NewTermLogger(procName + "TERM")
	if termLog == nil {
		return errors.New("init TERM logger failed")
	}
	termLog.SetLevel(DEBUG)
	// cmd := exec.Command("bash", "-c", "rm -rf *.log")
	// cmd.Run()
	logbaks[TERM] = termLog

	// 输出到文件
	fileLog := NewFileLogger(procName, defaultLogDir, "")
	if fileLog == nil {
		return errors.New("init FILE logger failed")
	}
	fileLog.SetLevel(DEBUG)
	logbaks[FILE] = fileLog
	curLevel = DEBUG
	return nil
}

func SetLevel(level Level) (err error) {
	switch curTag {
	case TERM:
		logger := logbaks[curTag]
		if logger == nil {
			err = errors.Errorf("curTag %s logger is nil", curTag)
		}
		logger.SetLevel(level)
	case FILE:
		logger := logbaks[curTag]
		if logger == nil {
			err = errors.Errorf("curTag %s logger is nil", curTag)
		}
		logger.SetLevel(level)
	default:
		for _, v := range logbaks {
			v.SetLevel(level)
		}
	}
	curLevel = level
	return
}

func GetLevel() Level {
	return curLevel
}

func ChooseLog(tag string) error {
	b := tag == ALL || tag == TERM || tag == FILE
	if !b {
		return errors.Errorf("%s tag is illgal", tag)
	}
	curTag = tag
	return nil
}

func getCurLogger() (loggers []*Logger) {
	switch curTag {
	case TERM:
		logger := logbaks[curTag]
		if logger == nil {
			errors.Errorf("curTag %s logger is nil", curTag)
		}
		loggers = append(loggers, logger)
	case FILE:
		logger := logbaks[curTag]
		if logger == nil {
			errors.Errorf("curTag %s logger is nil", curTag)
		}
		loggers = append(loggers, logger)
	default:
		for _, v := range logbaks {
			loggers = append(loggers, v)
		}
	}
	return
}

// Fatalf 记录致命日志
func Fatalf(tag string, format string, args ...interface{}) bool {
	for _, logger := range getCurLogger() {
		if !logger.Fatalf(tag, format, args...) {
			return false
		}
	}
	return true
}

// Errorf 记录错误日志
func Errorf(tag string, format string, args ...interface{}) bool {
	for _, logger := range getCurLogger() {
		if !logger.Errorf(tag, format, args...) {
			return false
		}
	}
	return true
}

func ErrorfAndRetErr(tag string, format string, args ...interface{}) (bool, error) {
	for _, logger := range getCurLogger() {
		if !logger.Errorf(tag, format, args...) {
			return false, nil
		}
	}
	err := fmt.Errorf(format, args...)
	return true, err
}

// Warnf 记录告警日志
func Warnf(tag string, format string, args ...interface{}) bool {
	for _, logger := range getCurLogger() {
		if !logger.Warnf(tag, format, args...) {
			return false
		}
	}
	return true
}

// Infof 记录正常日志
func Infof(tag string, format string, args ...interface{}) bool {
	for _, logger := range getCurLogger() {
		if !logger.Infof(tag, format, args...) {
			return false
		}
	}
	return true
}

// Debugf 记录调试日志
func Debugf(tag string, format string, args ...interface{}) bool {
	for _, logger := range getCurLogger() {
		if !logger.Debugf(tag, format, args...) {
			return false
		}
	}
	return true
}
