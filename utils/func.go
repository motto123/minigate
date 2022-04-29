package utils

import (
	"reflect"
	"runtime"
	"strings"
)

//GetFunctionName 获取函数名称
func GetFunctionName(i interface{}) string {
	fn := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	splits := strings.SplitAfter(fn, ".")
	return splits[len(splits)-1]
}
