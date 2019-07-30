package middleware

import (
	"bytes"
	"html/template"
	"strings"
)

// 字符串模板渲染
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

// 类似table类型字符串
//
// 转换为mysql table类型数据
func RenderTable(data string) []map[string]string {
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
