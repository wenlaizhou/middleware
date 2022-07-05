package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

func ParseToMap(param interface{}) (map[string]string, error) {
	if param == nil {
		return nil, errors.New("param is nil")
	}
	value := reflect.ValueOf(param)
	if value.Kind() != reflect.Map {
		return nil, errors.New("param is not map")
	}
	res := map[string]string{}
	mapRange := value.MapRange()
	for mapRange.Next() {
		res[ParseValueToString(mapRange.Key())] = ParseValueToString(mapRange.Value())
	}
	return res, nil
}

func ParseValueToString(value reflect.Value) string {
	switch value.Kind() {
	case reflect.Bool:
		return fmt.Sprintf("%v", value.Bool())
	case reflect.Int:
	case reflect.Int8:
	case reflect.Int16:
	case reflect.Int32:
	case reflect.Int64:
		return fmt.Sprintf("%v", value.Int())
	case reflect.Uint:
	case reflect.Uint8:
	case reflect.Uint16:
	case reflect.Uint32:
	case reflect.Uint64:
		return fmt.Sprintf("%v", value.Uint())
	case reflect.Float32:
	case reflect.Float64:
		return fmt.Sprintf("%v", value.Float())
	}
	if res, err := json.Marshal(value.Interface()); err == nil {
		return string(res)
	} else {
		mLogger.ErrorF("ParseValueToString error : %v, %v, %v", value.Kind().String(), value.Interface(), err.Error())
		return ""
	}
}

func ParseToString(param interface{}) (string, error) {
	if param == nil {
		return "", errors.New("param is nil")
	}
	value := reflect.ValueOf(param)
	switch value.Kind() {
	case reflect.Bool:
		return fmt.Sprintf("%v", value.Bool()), nil
	case reflect.Int:
	case reflect.Int8:
	case reflect.Int16:
	case reflect.Int32:
	case reflect.Int64:
		return fmt.Sprintf("%v", value.Int()), nil
	case reflect.Uint:
	case reflect.Uint8:
	case reflect.Uint16:
	case reflect.Uint32:
	case reflect.Uint64:
		return fmt.Sprintf("%v", value.Uint()), nil
	case reflect.Float32:
	case reflect.Float64:
		return fmt.Sprintf("%v", value.Float()), nil
	}
	res, err := json.Marshal(value.Interface())
	return string(res), err
}
