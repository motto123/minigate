package utils

import "reflect"

//GetStructName 获取的Struct的名字,argument:st is struct or pointer or slice, slice不能为空,否则返回空字符
func GetStructName(st interface{}) string {
	v := reflect.ValueOf(st)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	if t.Kind() == reflect.Struct {
		return t.Name()
	}
	if t.Kind() == reflect.Slice {
		if v.Len() == 0 {
			return ""
		}
		childV := v.Index(0)
		if childV.Kind() == reflect.Ptr {
			childV = childV.Elem()
		}
		childT := childV.Type()
		return childT.Name()
	}
	return ""
}
