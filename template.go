package middleware

import "errors"

func (this *Context) RenderTemplate(name string, model interface{}) error {
	if this.tpl != nil {
		return this.tpl.ExecuteTemplate(this.Response, name, model)
	}
	return errors.New("template 不存在")
}

func (this *Context) RenderTemplateKV(name string, kvs ...interface{}) error {
	if this.tpl == nil {
		return errors.New("template 不存在")
	}
	model := make(map[string]interface{})
	for i := 0; i < len(kvs); i += 2 {
		if v, ok := kvs[i].(string); ok {
			model[v] = kvs[i+1]
		}
	}
	return this.tpl.ExecuteTemplate(this.Response, name, model)
}
