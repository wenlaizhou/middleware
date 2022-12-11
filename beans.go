package middleware

import (
	"errors"
	"reflect"
	"strings"
)

// GetProperty
// 获取bean实例的属性值
//
// bean 应为 struct
func GetProperty(bean interface{}, name string) (interface{}, error) {
	if bean == nil {
		return nil, errors.New("bean为空")
	}
	name = strings.TrimSpace(name)
	if len(name) <= 0 {
		return nil, errors.New("属性名name为空")
	}
	tp := reflect.ValueOf(bean)
	if tp.Kind() != reflect.Struct {
		return nil, errors.New("bean类型不为struct")
	}
	beanVal := reflect.ValueOf(bean)
	fieldVal := beanVal.FieldByName(name)
	return fieldVal.Interface(), nil
}

// SetProperty
// 设置bean实例的属性值
//
// bean 应为 指向 struct指针
func SetProperty(bean interface{}, name string, value interface{}) error {
	if bean == nil {
		return errors.New("bean为空")
	}
	name = strings.TrimSpace(name)
	if len(name) <= 0 {
		return errors.New("属性name为空")
	}
	tp := reflect.ValueOf(bean)
	if tp.Kind() != reflect.Ptr {
		return errors.New("bean应为指向struct指针")
	}
	beanVal := reflect.ValueOf(bean).Elem()
	valVal := reflect.ValueOf(value)
	fieldVal := beanVal.FieldByName(name)
	if fieldVal.Kind() != valVal.Kind() {
		return errors.New("值类型与bean属性定义类型不符")
	}
	fieldVal.Set(reflect.ValueOf(value))
	return nil
}
