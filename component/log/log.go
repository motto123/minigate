// 软件包log向其它模块和服务提供方便易用的日志接口，模块的主要功能如下：
//	1.区分日志级别，包括debug、info、warn、error、fatal五个级别；
//	2.支持随时调整日志打印级别，如在正常发布情况下只打印error级别以上的日志，
//	而当调试在线问题时可通过发送信号、调整配置等方式调整服务的打印级别为debug；
//	3.支持l2met日志格式，方便集成Graphite工具对服务的状态进行监控；
//	4.(TODO)可向指定日志管理服务器集中上报指定等级的日志，便于日志集中管理、查询；
package log

import (
	"bufio"
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Level uint8

// 日志级别
const (
	FATAL Level = Level(logrus.FatalLevel) // 致命错误，输出日志并调用`os.Exit(1)`
	ERROR       = Level(logrus.ErrorLevel) // 错误日志
	WARN        = Level(logrus.WarnLevel)  // 告警日志
	INFO        = Level(logrus.InfoLevel)  // 系统正常运行日志
	DEBUG       = Level(logrus.DebugLevel) // 调试日志
)

var (
	LogOutputFile    = "file"
	LogOutputConsole = "console"
	LogOutputAll     = "all"
)

// 日志实例类型
type Logger struct {
	maxSize   int             // 日志文件大小
	path      string          // 日志存放目录
	fileName  string          // 日志文件名称，默认直接使用src
	src       string          // 日志来源
	fp        *os.File        // 当前正在写入的日志文件
	logger    *logrus.Logger  // 底层封装日志库
	tagFilter map[string]bool // 日志过滤表
	closed    bool            //logger读写管道是否已经关闭
	Mytype    string          //FILE/TERM
}

const (
	minFileSize     = 1 << 20  //1MB
	defaultFileSize = 64 << 20 //64MB
	maxFileSize     = 1 << 30  //1GB
	maxBakFileNum   = 32       //日志备份数量
)

var (
	//记录所有已分配的日志实例，以src为key
	allLoggers     = make(map[string]*Logger, 10)
	logMutex       = new(sync.RWMutex)
	checkInterval  = time.Second * 5
	defaultBufSize = 1 << 20
)

//获取路径分割符
func getPathDel() string {
	return "/"
}

func init() {
	go func() {
		for {
			time.Sleep(checkInterval)

			// 检查日志文件是否存在，若不存在则创建
			logMutex.RLock()
			for _, log := range allLoggers {
				if log.logger.Out == nil {
					initLogFile(log)
					continue
				} else {
					logPath := log.path + getPathDel() + log.fileName
					_, err := os.Lstat(logPath)
					if err == nil {
						continue
					}
					initLogFile(log)
				}
			}
			logMutex.RUnlock()

			//TODO: 临时放这里
			FlushAll()
		}
	}()
}

func initLogFile(l *Logger) bool {
	var err error
	var f *os.File

	_, err = os.Lstat(l.path)
	if err != nil {
		err = os.MkdirAll(l.path, 0777)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			return false
		}
	}

	logPath := l.path + getPathDel() + l.fileName
	f, err = os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return false
	}
	if l.fp != nil {
		l.fp.Close()
	}
	l.fp = f
	writer, _ := rotatelogs.New(
		logPath+".%Y%m%d",
		rotatelogs.WithLinkName(logPath),
		//rotatelogs.WithMaxAge(time.Duration(180)*time.Second),
		rotatelogs.WithRotationCount(maxBakFileNum),
		rotatelogs.WithRotationTime(time.Duration(24)*time.Hour),
	)
	l.logger.SetOutput(writer)
	l.logger.ExitFunc = loggerExitFunc
	return true
}

func loggerExitFunc(int) {
	FlushAll()
	os.Exit(1)
}

type bakcollections []os.FileInfo

// 根据日志文件名前缀过滤出备份目录下所有的备份文件
func newBakCollections(logPrefix string, fis []os.FileInfo) bakcollections {
	baks := []os.FileInfo{}
	for _, fi := range fis {
		if strings.HasPrefix(fi.Name(), logPrefix) {
			baks = append(baks, fi)
		}
	}
	return bakcollections(baks)
}

func (baks bakcollections) Len() int {
	return len(baks)
}

func (baks bakcollections) Less(i, j int) bool {
	fields := strings.Split(baks[i].Name(), ".")
	t, _ := time.Parse(time.RFC3339, fields[len(fields)-1])
	t1 := t.Unix()
	fields = strings.Split(baks[j].Name(), ".")
	t, _ = time.Parse(time.RFC3339, fields[len(fields)-1])
	t2 := t.Unix()
	return t1 < t2
}

func (baks bakcollections) Swap(i, j int) {
	baks[i], baks[j] = baks[j], baks[i]
}

func bakLogFile(l *Logger) {
	var err error
	var fi os.FileInfo
	var dp *os.File

	if l.path == "" {
		return
	}

	logPath := l.path + getPathDel() + l.fileName
	fi, err = os.Lstat(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	if fi.Size() < int64(l.maxSize) {
		return
	}

	now := time.Now()
	bakPath := l.path + getPathDel() + "baklog"
	bakFile := fmt.Sprintf("%s/%s.%s", bakPath, l.fileName, now.Format(time.RFC3339))
	_, err = os.Lstat(bakPath)
	if err != nil {
		err = os.MkdirAll(bakPath, 0777)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			return
		}
	}
	err = os.Rename(logPath, bakFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	initLogFile(l)

	//删除多余的备份
	dp, err = os.Open(bakPath)
	if err == nil {
		fileInfos, _ := dp.Readdir(0)
		baks := newBakCollections(l.fileName, fileInfos)
		if len(baks) > maxBakFileNum {
			sort.Sort(baks)
			for i := 0; i < len(baks)-maxBakFileNum; i++ {
				os.Remove(bakPath + getPathDel() + baks[i].Name())
			}
		}
	}
}

//日志自定义格式
type LogFormatter struct{}

//格式详情
func (s *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05.999")

	data := entry.Data

	vTag := data["tag"]
	vFunc := data["func"]
	vLine := data["line"]
	//vFile := data["file"]

	vMsg := entry.Message
	vLevel := strings.ToUpper(entry.Level.String())

	var msg string
	if vTag == nil {
		msg = fmt.Sprintf("[%s %s] %v\n", timestamp, vLevel, vMsg)
	} else {
		msg = fmt.Sprintf("[%s %s] %v func=%v:%v tag=%v\n", timestamp, vLevel, vMsg, vFunc, vLine, vTag)
	}
	return []byte(msg), nil
}

// 分配新的日志实例，打印日志到文件，默认日志级别为ERROR
//	src：日志来源标识符，空字符串为非法参数
//	path：日志存放路径，空字符串为非法参数
//	fileName：日志文件名称，传入空字符串则默认使用src参数
func NewFileLogger(src string, path string, fileName string) *Logger {
	logMutex.Lock()
	defer logMutex.Unlock()
	if src == "" || path == "" {
		fmt.Fprintf(os.Stderr, "invalid params, %v %v %v\n", src, path, fileName)
		return nil
	}
	if allLoggers[src] != nil {
		fmt.Fprintf(os.Stderr, "same src %s already exists\n", src)
		return nil
	}

	l := new(Logger)
	l.src = src
	l.path = path
	l.maxSize = defaultFileSize
	if fileName != "" {
		l.fileName = fileName
	} else {
		l.fileName = src + ".log"
	}
	l.logger = logrus.New()
	l.logger.SetFormatter(&LogFormatter{})
	l.tagFilter = make(map[string]bool, 10)
	l.SetLevel(ERROR)
	l.closed = false
	l.Mytype = "FILE"
	if initLogFile(l) {
		allLoggers[src] = l
		return l
	} else {
		return nil
	}
}

// 分配新的日志实例，打印日志到终端，默认日志级别为ERROR
//	src：日志来源标识符，空字符串为非法参数
func NewTermLogger(src string) *Logger {
	logMutex.Lock()
	defer logMutex.Unlock()
	if src == "" {
		fmt.Fprintf(os.Stderr, "invalid params, %s\n", src)
		return nil
	}
	if allLoggers[src] != nil {
		fmt.Fprintf(os.Stderr, "same src %s already exists\n", src)
		return nil
	}

	l := new(Logger)
	l.src = src
	l.logger = logrus.New()
	l.logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:             true,
		ForceColors:               true,
		EnvironmentOverrideColors: true,
		TimestampFormat:           "2006-01-02 15:04:05.999",
	})
	l.logger.SetOutput(os.Stdout)
	l.tagFilter = make(map[string]bool, 10)
	l.SetLevel(ERROR)
	l.closed = false
	l.Mytype = "TERM"
	allLoggers[src] = l
	return l
}

// 删除日志实例
func Remove(l *Logger) {
	fmt.Fprintf(os.Stderr, "unsupport this function log:Remove")
	//if l != nil {
	//	logMutex.Lock()
	//	defer logMutex.Unlock()
	//	if !l.closed {
	//		l.logger.Del()
	//		l.closed = true
	//	}
	//	if l.fp != nil {
	//		l.fp.Close()
	//	}
	//	delete(allLoggers, l.src)
	//}
}

// 持久化日志缓存
func FlushAll() {
	logMutex.Lock()
	defer logMutex.Unlock()
	for _, v := range allLoggers {
		if v != nil {
			writer, ok := v.logger.Out.(*bufio.Writer)
			if ok {
				err := writer.Flush()
				if err != nil && err.Error() != "short write" {
					fmt.Fprintf(os.Stderr, "Flush failed, %v\n", err)
				}
			}
		}
	}

}

// 设置日志文件大小上限
func (l *Logger) SetMaxFileSize(s int) {
	if l.path == "" {
		return
	}
	if s < minFileSize {
		l.maxSize = minFileSize
	} else if s > maxFileSize {
		l.maxSize = maxFileSize
	} else {
		l.maxSize = s
	}
}

// 获取日志打印级别
func (l *Logger) GetLevel() Level {
	return Level(l.logger.Level)
}

// 设置日志打印级别，设置完成后系统只输出高于或等于当前等级的日志
func (l *Logger) SetLevel(lvl Level) {
	l.logger.SetLevel(logrus.Level(lvl))
}

// 增加日志过滤tag
func (l *Logger) AddTagFilter(tag string) {
	l.tagFilter[tag] = true
}

// 删除日志过滤tag
func (l *Logger) DelTagFilter(tag string) {
	delete(l.tagFilter, tag)
}

// 删除所有日志过滤tag
func (l *Logger) ClearTagFilter(tag string) {
	l.tagFilter = make(map[string]bool, 10)
}

// 记录相应等级的日志，按照format指定的格式输出日志，tag参数用于日志过滤
func (l *Logger) logf(skip int, lvl Level, tag string, simpleOutput bool, format string, args ...interface{}) bool {

	if Level(l.logger.Level) < lvl {
		return true
	}

	_, exists := l.tagFilter[tag]
	if len(l.tagFilter) != 0 && !exists {
		return false
	}

	var fields logrus.Fields
	if simpleOutput {
		fields = logrus.Fields{}
	} else {
		funcptr, file, line, ok := runtime.Caller(skip)
		if !ok {
			return false
		}
		split := strings.SplitAfter(file, "/")
		if len(split) >= 3 {
			file = split[len(split)-3] + split[len(split)-2] + split[len(split)-1]
		}
		//split := strings.SplitAfter(file, "/")
		//fmt.Println(len(split))
		//fmt.Println(split)
		//file := split[len(split)-3] + split[len(split)-2] + split[len(split)-1]

		if tag != "" {
			fields = logrus.Fields{
				"tag":  tag,
				"func": runtime.FuncForPC(funcptr).Name(),
				"line": line,
				"file": file,
			}
		} else {
			fields = logrus.Fields{
				"func": runtime.FuncForPC(funcptr).Name(),
				"line": line,
			}
		}
	}

	switch lvl {
	case FATAL:
		l.logger.WithFields(fields).Fatalf(format, args...)
	case ERROR:
		l.logger.WithFields(fields).Errorf(format, args...)
	case WARN:
		l.logger.WithFields(fields).Warnf(format, args...)
	case INFO:
		l.logger.WithFields(fields).Infof(format, args...)
	case DEBUG:
		l.logger.WithFields(fields).Debugf(format, args...)
	}

	//bakLogFile(l)
	return true
}

// 记录致命日志
func (l *Logger) Fatalf(tag string, format string, args ...interface{}) bool {
	return l.logf(3, FATAL, tag, false, format, args...)
}

// 记录错误日志
func (l *Logger) Errorf(tag string, format string, args ...interface{}) bool {
	return l.logf(3, ERROR, tag, false, format, args...)
}

// 记录告警日志
func (l *Logger) Warnf(tag string, format string, args ...interface{}) bool {
	return l.logf(3, WARN, tag, false, format, args...)
}

// 记录正常日志
func (l *Logger) Infof(tag string, format string, args ...interface{}) bool {
	return l.logf(3, INFO, tag, false, format, args...)
}

// 记录调试日志
func (l *Logger) Debugf(tag string, format string, args ...interface{}) bool {
	return l.logf(3, DEBUG, tag, false, format, args...)
}

// 记录致命日志(显示指定调用栈深度)
func (l *Logger) FatalfWithSkip(skip int, tag string, format string, args ...interface{}) bool {
	return l.logf(skip, FATAL, tag, false, format, args...)
}

// 记录错误日志(显示指定调用栈深度)
func (l *Logger) ErrorfWithSkip(skip int, tag string, format string, args ...interface{}) bool {
	return l.logf(skip, ERROR, tag, false, format, args...)
}

// 记录告警日志(显示指定调用栈深度)
func (l *Logger) WarnfWithSkip(skip int, tag string, format string, args ...interface{}) bool {
	return l.logf(skip, WARN, tag, false, format, args...)
}

// 记录正常日志(显示指定调用栈深度)
func (l *Logger) InfofWithSkip(skip int, tag string, format string, args ...interface{}) bool {
	return l.logf(skip, INFO, tag, false, format, args...)
}

// 记录调试日志(显示指定调用栈深度)
func (l *Logger) DebugfWithSkip(skip int, tag string, format string, args ...interface{}) bool {
	return l.logf(skip, DEBUG, tag, false, format, args...)
}

// 记录致命日志（不添加附加信息）
func (l *Logger) SimpleFatalf(tag string, format string, args ...interface{}) bool {
	return l.logf(0, FATAL, tag, true, format, args...)
}

// 记录错误日志（不添加附加信息）
func (l *Logger) SimpleErrorf(tag string, format string, args ...interface{}) bool {
	return l.logf(0, ERROR, tag, true, format, args...)
}

// 记录告警日志（不添加附加信息）
func (l *Logger) SimpleWarnf(tag string, format string, args ...interface{}) bool {
	return l.logf(0, WARN, tag, true, format, args...)
}

// 记录正常日志（不添加附加信息）
func (l *Logger) SimpleInfof(tag string, format string, args ...interface{}) bool {
	return l.logf(0, INFO, tag, true, format, args...)
}

// 记录调试日志（不添加附加信息）
func (l *Logger) SimpleDebugf(tag string, format string, args ...interface{}) bool {
	return l.logf(0, DEBUG, tag, true, format, args...)
}

// 重新设置日志目录，老的日志会被移动到新的位置，应尽量避免该操作
func (l *Logger) ChangeDir(dir string) bool {
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create path %s failed\n", dir)
		return false
	}
	os.Rename(l.path, dir)
	os.RemoveAll(l.path)
	l.path = dir
	return initLogFile(l)
}

// 获取日志保存路径
func (l *Logger) GetDir() string {
	return l.path
}

func GetCallerPrefix() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}
	//_, filename := path.Split(file)
	return "[MSG: " + file + ":" + strconv.Itoa(line) + "] "
}
