package log

import (
	"os/exec"
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestMe(t *testing.T) {
	log1 := NewTermLogger("termlog_test1")
	defer FlushAll()
	if !assert.NotNil(t, log1, "allocate logger failed!") {
		return
	}
	log1.SetLevel(DEBUG)

	log1.Errorf("log1", "1. This is a log1")
	log1.Infof("log1", "1. This is a log2")
	log1.Warnf("log1", "1. This is a log3")
	log1.Errorf("log1", "1. This is a log4")

	m := make(map[int]int)
	m[100] = 100
	log1.Infof("log1", "sdfasssssssssfasdfasdf %p %v", m, m)

	cmd := exec.Command("bash", "-c", "rm -rf *.log")
	cmd.Run()
}

func TestTermLogger(t *testing.T) {
	log1 := NewTermLogger("termlog_test1")
	log2 := NewTermLogger("termlog_test2")
	defer FlushAll()
	if !assert.NotNil(t, log1, "allocate logger failed!") {
		return
	}
	if !assert.NotNil(t, log2, "allocate logger failed!") {
		return
	}
	log3 := NewTermLogger("termlog_test1")
	if !assert.Nil(t, log3, "allocate logger failed!") {
		return
	}

	log1.Errorf("log1", "1. This is a log1")
	log1.Infof("log1", "1. This is a log2")
	log1.Warnf("log1", "1. This is a log3")
	log1.Errorf("log1", "1. This is a log4")

	log2.SetLevel(DEBUG)
	log2.Errorf("log2", "2. This is a log1")
	log2.Infof("log2", "2. This is a log2")
	log2.Warnf("log2", "2. This is a log3")
	log2.Errorf("log2", "2. This is a log4")

	log2.AddTagFilter("log2")
	log2.Errorf("log2", "3. This is a log1")
	log2.Infof("log2", "3. This is a log2")

	log2.ClearTagFilter("log2")
	log2.Warnf("log2", "4. This is a log3")
	log2.Errorf("log2", "4. This is a log4")
	log2.Errorf("", "4. This is a log4 without tag")

	log2.AddTagFilter("whatever")
	log2.Errorf("", "5. This is a log4 without tag")
	log2.DelTagFilter("whatever")
	log2.Errorf("", "5. This is a log4 without tag")

	cmd := exec.Command("bash", "-c", "rm -rf *.log")
	cmd.Run()
}

func TestFileLogger(t *testing.T) {
	log := NewFileLogger("filelog_test", ".", "")
	log1 := NewFileLogger("filelog_test1", ".", "filelog.log")
	defer FlushAll()
	if !assert.NotNil(t, log, "allocate logger failed!") {
		return
	}
	if !assert.NotNil(t, log1, "allocate logger failed!") {
		return
	}

	log.Errorf("tmp", "This is a log1")
	log.Infof("tmp", "This is a log2")
	log.Warnf("tmp", "This is a log3")
	log.Errorf("tmp", "This is a log4")
	log1.Errorf("tmp", "This is a log1")
	log1.Infof("tmp", "This is a log2")
	log1.Warnf("tmp", "This is a log3")
	log1.Errorf("tmp", "This is a log4")

	log.SetLevel(ERROR)
	log1.SetLevel(DEBUG)
	log.Errorf("tmp", "This is a log1")
	log.Infof("tmp", "This is a log2")
	log.Warnf("tmp", "This is a log3")
	log.Errorf("tmp", "This is a log4")
	log1.Errorf("tmp", "This is a log1")
	log1.Infof("tmp", "This is a log2")
	log1.Warnf("tmp", "This is a log3")
	log1.Errorf("tmp", "This is a log4")

	cmd := exec.Command("bash", "-c", "rm -rf *.log")
	cmd.Run()
}

func TestSimple(t *testing.T) {
	InitLogger("a", "./")
	Debugf("tag", "hi %d", 1111111111)
	SetLevel(ERROR)
	Debugf("tag", "hi %d", 2222222222)

	ChooseLog(FILE)
	Debugf("tag", "hi %d", 3333333333)
	ChooseLog(TERM)
	Debugf("tag", "hi %d", 4444444444)
	SetLevel(DEBUG)
	Debugf("tag", "hi %d", 5555555555)
	ChooseLog("")
	Debugf("tag", "double %d", 8888888885)
	Infof("tag", "double %d", 8888888885)
	Errorf("tag", "double %d", 8888888885)
	ChooseLog(TERM)
	Debugf("tag", "hi %d", 4444444444)

}

/*
func TestFatal(t *testing.T) {
	log := NewFileLogger("fatal_log", ".", "")
	if !assert.NotNil(t, log, "allocate logger failed!") {
		return
	}

	log.Fatalf("tmp", "**** fatal log!!!!")
}
*/

/*
func TestFileLoggerBak(t *testing.T) {
	log := NewFileLogger("filelog_test_bak", ".", "")
	defer FlushAll()
	if !assert.NotNil(t, log, "allocate logger failed!") {
		return
	}

	for {
		log.Errorf("This is a log4")
	}
}
*/

var bench_counter int32 = 1

//func BenchmarkTermLogger(b *testing.B) {
//	log := NewTermLogger(fmt.Sprintf("filelog_benchmark%d", atomic.LoadInt32(&bench_counter)))
//	bench_counter = atomic.AddInt32(&bench_counter, 1)
//	defer FlushAll()
//	if !assert.NotNil(b, log, "allocate logger failed!") {
//		return
//	}
//
//	for i := 0; i < b.N; i++ {
//		log.Infof("benchmark", "This is a log2")
//	}
//
//	cmd := exec.Command("bash", "-c", "rm -rf *.log")
//	cmd.Run()
//}
//
//func BenchmarkFileLogger(b *testing.B) {
//	log := NewFileLogger(fmt.Sprintf("filelog_benchmark%d", atomic.LoadInt32(&bench_counter)), ".", "")
//	bench_counter = atomic.AddInt32(&bench_counter, 1)
//	defer FlushAll()
//	if !assert.NotNil(b, log, "allocate logger failed!") {
//		return
//	}
//
//	for i := 0; i < b.N; i++ {
//		log.Errorf("benchmark", "This is a log4")
//	}
//
//	cmd := exec.Command("bash", "-c", "rm -rf *.log")
//	cmd.Run()
//}
