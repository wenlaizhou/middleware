package middleware

import (
	"errors"
	"strings"
)

//根据请求判断接口
func (c *Context) RenderTemplate(name string, model interface{}) error {
	userAccept := c.GetHeader("Accept")
	if len(userAccept) <= 0 || !strings.Contains(userAccept, "text/html") {
		c.ApiResponse(0, "", model)
		return nil
	}
	if c.tpl != nil {
		c.code = 200
		if c.EnableI18n {
			locale := c.GetCookie("locale")
			if len(locale) <= 0 {
				locale = "cn"
			}
			var message map[string]string
			if locale == "cn" {
				message = c.Message.Cn
			} else {
				message = c.Message.En
			}
			err := c.tpl.ExecuteTemplate(c.Response, name, map[string]interface{}{
				"data":    model,
				"message": message,
			})
			if err != nil {
				c.code = 500
			}
			return err
		} else {
			err := c.tpl.ExecuteTemplate(c.Response, name, model)
			if err != nil {
				c.code = 500
			}
			return err
		}
	}
	c.code = 500
	return errors.New("template 不存在")
}

// 直接转换成接口
func (c *Context) RenderTemplateKV(name string, kvs ...interface{}) error {
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
	userAccept := c.GetHeader("Accept")
	if len(userAccept) <= 0 || !strings.Contains(userAccept, "text/html") {
		c.ApiResponse(0, "", model)
		return nil
	}
	if c.tpl == nil {
		return errors.New("template 不存在")
	}
	c.code = 200
	if c.EnableI18n {
		locale := c.GetCookie("locale")
		if len(locale) <= 0 {
			locale = "cn"
		}
		var message map[string]string
		if locale == "cn" {
			message = c.Message.Cn
		} else {
			message = c.Message.En
		}
		model["message"] = message
	}
	return c.tpl.ExecuteTemplate(c.Response, name, model)
}
