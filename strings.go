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

// StringFormatInterface 格式化字符串
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
