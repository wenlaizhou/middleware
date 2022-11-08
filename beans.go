package middleware

import (
	"reflect"
	"strings"
)

// Properties
// 获取bean实例的属性值
func Properties(bean interface{}, name string) interface{} {
	if bean == nil {
		return nil
	}
	name = strings.TrimSpace(name)
	if len(name) <= 0 {
		return nil
	}
	beanVal := reflect.ValueOf(bean)
	fieldVal := beanVal.FieldByName(name)
	return fieldVal.Interface()
}

// SetProperties
// 设置bean实例的属性值
func SetProperties(bean interface{}, name string, value interface{}) {
	if bean == nil {
		return
	}
	name = strings.TrimSpace(name)
	if len(name) <= 0 {
		return
	}
	beanVal := reflect.ValueOf(bean)
	fieldVal := beanVal.FieldByName(name)
	fieldVal.Set(reflect.ValueOf(value))
	return
}
