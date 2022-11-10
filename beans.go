package middleware

import (
	"reflect"
	"strings"
)

// GetProperty
// 获取bean实例的属性值
//
// bean 应为 struct
func GetProperty(bean interface{}, name string) interface{} {
	if bean == nil {
		return nil
	}
	name = strings.TrimSpace(name)
	if len(name) <= 0 {
		return nil
	}
	tp := reflect.ValueOf(bean)
	if tp.Kind() != reflect.Struct {
		return nil
	}
	beanVal := reflect.ValueOf(bean)
	fieldVal := beanVal.FieldByName(name)
	return fieldVal.Interface()
}

// SetProperty
// 设置bean实例的属性值
//
// bean 应为 指向 struct指针
func SetProperty(bean interface{}, name string, value interface{}) {
	if bean == nil {
		return
	}
	name = strings.TrimSpace(name)
	if len(name) <= 0 {
		return
	}
	tp := reflect.ValueOf(bean)
	if tp.Kind() != reflect.Ptr {
		return
	}
	beanVal := reflect.ValueOf(bean).Elem()
	valVal := reflect.ValueOf(value)
	fieldVal := beanVal.FieldByName(name)
	if fieldVal.Kind() != valVal.Kind() {
		return
	}
	fieldVal.Set(reflect.ValueOf(value))
	return
}
