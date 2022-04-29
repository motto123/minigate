package server

import (
	"testing"
)

type hi func(string) string

func TestUnmarshalForProto(t *testing.T) {
	teseaa(aaa)
}

func teseaa(h hi) {
	//fn := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	//fmt.Printf("fn: %+v\n", fn)
	println(GetFunctionName(h))
}

func aaa(string2 string) string {
	return ""
}
