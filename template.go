package middleware

import "errors"

func (this *Context) RenderTemplate(name string, model interface{}) error {
	userAgent := this.GetHeader(UserAgent)
	if len(userAgent) <= 0 {
		// 	无User-Agent的判断为api调用
		return this.ApiResponse(0, "", model)
	}
	if this.tpl != nil {
		this.code = 200
		err := this.tpl.ExecuteTemplate(this.Response, name, model)
		this.code = 500
		return err
	}
	this.code = 500
	return errors.New("template 不存在")
}

// 直接转换成接口
func (this *Context) RenderTemplateKV(name string, kvs ...interface{}) error {
	userAgent := this.GetHeader(UserAgent)
	model := make(map[string]interface{})
	kvsLen := len(kvs)
	for i := 0; i < kvsLen; i += 2 {
		if v, ok := kvs[i].(string); ok {
			var value interface{}
			if kvsLen <= i+1 {
				value = nil
			} else {
				value = kvs[i+1]
			}
			model[v] = value
		}
	}
	if len(userAgent) <= 0 {
		// 	无User-Agent的判断为api调用
		return this.ApiResponse(0, "", model)
	}
	if this.tpl == nil {
		return errors.New("template 不存在")
	}
	this.code = 200
	return this.tpl.ExecuteTemplate(this.Response, name, model)
}
