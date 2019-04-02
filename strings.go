package middleware

import (
	"bytes"
	"html/template"
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
	err = t.Execute(buff, kvs)
	if err != nil {
		return tpl, err
	}
	return buff.String(), nil
}
