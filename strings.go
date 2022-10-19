package middleware

import (
	"bytes"
	"fmt"
	"html/template"
	"reflect"
	"strings"
)

// RenderTemplateKV 字符串模板渲染
func RenderTemplateKV(tpl string, kvs ...interface{}) (string, error) {
	if len(tpl) <= 0 {
		return tpl, nil
	}
	t, err := template.New("renderTemplate").Parse(tpl)
	if err != nil {
		return tpl, err
	}
	buff := bytes.NewBufferString("")
	model := make(map[string]interface{})
	for i := 0; i < len(kvs); i += 2 {
		if v, ok := kvs[i].(string); ok {
			model[v] = kvs[i+1]
		}
	}
	err = t.Execute(buff, model)
	if err != nil {
		return tpl, err
	}
	return buff.String(), nil
}

// RenderTable 类似table类型字符串
//
// 转换为mysql table类型数据
func RenderTable(data string, width int) []map[string]string {
	data = strings.TrimSpace(data)
	if len(data) <= 0 {
		return nil
	}
	lines := strings.Split(data, "\n")
	if len(lines) <= 1 {
		return nil
	}
	headers := strings.Fields(strings.TrimSpace(lines[0]))
	if len(headers) <= 0 {
		return nil
	}
	if width > 0 {
		headers = headers[:width]
	}
	var res []map[string]string
	for _, line := range lines[1:] {
		row := make(map[string]string)
		fields := strings.Fields(strings.TrimSpace(line))
		for i, header := range headers {
			if len(fields)-1 >= i {
				row[header] = fields[i]
			}
		}
		res = append(res, row)
	}
	return res
}

// IsEmpty 判断数据是否为空
func IsEmpty(param interface{}) bool {
	switch reflect.TypeOf(param) {
	case reflect.TypeOf(""):
		return param == nil || fmt.Sprintf("%s", param) == ""
	default:
		return param == nil
	}
}

// StringFormat 格式化字符串
//
// ${name}
func StringFormat(formatter string, params map[string]interface{}) string {
	res := formatter
	if len(params) <= 0 {
		return res
	}
	for k, v := range params {
		if len(k) <= 0 {
			continue
		}
		res = strings.ReplaceAll(res, fmt.Sprintf("${%v}", k), fmt.Sprintf("%v", v))
	}
	return res
}

// StringFormatMap 格式化字符串
//
// ${name}
func StringFormatMap(formatter string, params map[string]string) string {
	res := formatter
	if len(params) <= 0 {
		return res
	}
	for k, v := range params {
		if len(k) <= 0 {
			continue
		}
		res = strings.ReplaceAll(res, fmt.Sprintf("${%v}", k), strings.TrimSpace(v))
	}
	return res
}

// StringFormatConf 格式化字符串
//
// ${name}
func StringFormatConf(formatter string, conf Config) string {
	res := formatter
	if len(conf) <= 0 {
		return res
	}
	for k, v := range conf {
		if len(k) <= 0 {
			continue
		}
		res = strings.ReplaceAll(res, fmt.Sprintf("${%v}", k), fmt.Sprintf("%v", v))
	}
	return res
}

func StringFormatStructs(formatter string, obj interface{}) string {
	res := formatter
	val := reflect.ValueOf(obj)
	typ := reflect.TypeOf(obj)
	if val.NumField() <= 0 {
		return res
	}
	for i := 0; i < val.NumField(); i++ {
		fieldName := typ.Field(i).Name
		fieldName = StringFirstLower(fieldName)
		fieldStr := convertVal(val.Field(i).Elem())
		res = strings.ReplaceAll(res, fmt.Sprintf("${%v}", fieldName), fmt.Sprintf("%v", fieldStr))
	}
	return res
}

// StringFirstLower 第一个字母小写
func StringFirstLower(param string) string {
	if len(param) <= 0 {
		return ""
	}
	return fmt.Sprintf("%v%v", strings.ToLower(param[0:1]), param[1:])
}

func convertVal(dataVal reflect.Value) string {
	switch dataVal.Kind() {
	case reflect.Ptr:
		// indirect pointers
		if dataVal.IsNil() {
			return ""
		} else {
			return convertInterface(dataVal.Elem().Interface())
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%v", dataVal.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%v", dataVal.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.4f", dataVal.Float())
	case reflect.Bool:
		return fmt.Sprintf("%v", dataVal.Bool())
	case reflect.Slice:
		switch t := dataVal.Type(); {
		case t == jsonType:
			return fmt.Sprintf("%v", t)
		case t.Elem().Kind() == reflect.Uint8:
			return fmt.Sprintf("%v", string(dataVal.Bytes()))
		default:
			return fmt.Sprintf("unsupported type %T, a slice of %s", dataVal, t.Elem().Kind())
		}
	case reflect.String:
		return dataVal.String()
	}
	return fmt.Sprintf("unsupported type %T, a %s", dataVal, dataVal.Kind())
}

// StringFirstNotEmpty 选择参数中首个非空字符串
//
//
func StringFirstNotEmpty(params ...string) string {
	if len(params) <= 0 {
		return ""
	}
	for i := 0; i < len(params); i++ {
		p := strings.TrimSpace(params[i])
		if len(p) > 0 {
			return p
		}
	}
	return ""
}
