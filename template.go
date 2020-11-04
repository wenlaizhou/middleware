package middleware

import (
	"errors"
	"strings"
)

//根据请求判断接口
func (this *Context) RenderTemplate(name string, model interface{}) error {
	userAccept := this.GetHeader("Accept")
	if len(userAccept) <= 0 || !strings.Contains(userAccept, "text/html") {
		return this.ApiResponse(0, "", model)
	}
	if this.tpl != nil {
		this.code = 200
		if this.EnableI18n {
			locale := this.GetCookie("locale")
			if len(locale) <= 0 {
				locale = "cn"
			}
			var message map[string]string
			if locale == "cn" {
				message = this.Message.Cn
			} else {
				message = this.Message.En
			}
			err := this.tpl.ExecuteTemplate(this.Response, name, map[string]interface{}{
				"data":    model,
				"message": message,
			})
			if err != nil {
				this.code = 500
			}
			return err
		} else {
			err := this.tpl.ExecuteTemplate(this.Response, name, model)
			if err != nil {
				this.code = 500
			}
			return err
		}
	}
	this.code = 500
	return errors.New("template 不存在")
}

// 直接转换成接口
func (this *Context) RenderTemplateKV(name string, kvs ...interface{}) error {
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
	userAccept := this.GetHeader("Accept")
	if len(userAccept) <= 0 || !strings.Contains(userAccept, "text/html") {
		return this.ApiResponse(0, "", model)
	}
	if this.tpl == nil {
		return errors.New("template 不存在")
	}
	this.code = 200
	if this.EnableI18n {
		locale := this.GetCookie("locale")
		if len(locale) <= 0 {
			locale = "cn"
		}
		var message map[string]string
		if locale == "cn" {
			message = this.Message.Cn
		} else {
			message = this.Message.En
		}
		model["message"] = message
	}
	return this.tpl.ExecuteTemplate(this.Response, name, model)
}
