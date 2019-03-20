package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"sync"
)

// 上下文数据结构
type Context struct {
	Request    *http.Request
	Response   http.ResponseWriter
	body       []byte
	tpl        *template.Template
	pathParams map[string]string
	writeable  bool
	sync.RWMutex
}

// 获取路径参数, /{参数名称}
func (this *Context) GetPathParam(key string) string {
	value, ok := this.pathParams[key]
	if ok {
		return value
	}
	return ""
}

// 获取请求体
func (this *Context) GetBody() []byte {
	this.Lock()
	defer this.Unlock()
	if len(this.body) > 0 {
		return this.body
	}
	data, err := ioutil.ReadAll(this.Request.Body)
	this.body = data
	if err == nil && len(data) > 0 {
		this.body = data
		return this.body
	}
	return nil
}

// 获取body中
//
// json类型数据体
func (this *Context) GetJSON() (map[string]interface{}, error) {
	res := make(map[string]interface{})
	if len(this.GetBody()) > 0 {
		err := json.Unmarshal(this.GetBody(), &res)
		return res, err
	}
	return res, nil
}

// 获取json对象中key对应的字符串
//
// 没有该key则返回""
func GetJsonParamStr(key string, jsonObj map[string]interface{}) string {
	val, hasVal := jsonObj[key]
	if !hasVal {
		return ""
	}
	return fmt.Sprintf("%v", val)
}

// 获取query参数
func (this *Context) GetQueryParam(key string) string {
	return this.Request.URL.Query().Get(key)
}

func (this *Context) WriteJSON(data interface{}) error {
	res, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = this.OK(ApplicationJson, res)
	return err
}

func (this *Context) GetContentType() string {
	return this.Request.Header.Get(ContentType)
}

func (this *Context) GetHeader(key string) string {
	return this.Request.Header.Get(key)
}

func (this *Context) GetCookie(key string) string {
	cook, err := this.Request.Cookie(key)
	if err != nil {
		return ""
	}
	return cook.Value
}

func (this *Context) SetCookie(c *http.Cookie) {
	http.SetCookie(this.Response, c)
}

// 302跳转
func (this *Context) Redirect(path string) error {
	this.Lock()
	defer this.Unlock()
	if !this.writeable {
		return errors.New("禁止重复写入response")
	}
	this.writeable = false
	http.Redirect(this.Response, this.Request, path, http.StatusFound)
	return nil
}

func (this *Context) OK(contentType string, content []byte) error {
	this.Lock()
	defer this.Unlock()
	if !this.writeable {
		return errors.New("禁止重复写入response")
	}
	this.writeable = false
	if len(contentType) > 0 {
		this.SetHeader(ContentType, contentType)
	}
	this.SetHeader("server", "framework")
	_, err := this.Response.Write(content)
	return err
}

func (this *Context) Code(static int) error {
	this.Lock()
	defer this.Unlock()
	if !this.writeable {
		return errors.New("禁止重复写入response")
	}
	this.writeable = false
	this.SetHeader("server", "framework")
	this.Response.WriteHeader(static)
	return nil
}

func (this *Context) Error(static int, htmlStr string) error {
	this.Lock()
	defer this.Unlock()
	if !this.writeable {
		return errors.New("禁止重复写入response")
	}
	this.writeable = false
	this.SetHeader("server", "framework")
	this.SetHeader(ContentType, Html)
	this.Response.WriteHeader(static)
	_, _ = this.Response.Write([]byte(htmlStr))
	return nil
}

func (this *Context) SetHeader(key string, value string) {
	this.Response.Header().Set(key, value)
}

func (this *Context) DelHeader(key string) {
	this.Response.Header().Del(key)
}

func newContext(w http.ResponseWriter, r *http.Request) Context {
	return Context{
		writeable:  true,
		Response:   w,
		Request:    r,
		pathParams: make(map[string]string),
	}
}

func (this *Context) GetMethod() string {
	return this.Request.Method
}

func (this *Context) JSON(jsonStr string) error {
	err := this.OK(ApplicationJson, []byte(jsonStr))
	return err
}

func (this *Context) RemoteAddr() string {
	return this.Request.RemoteAddr
}

// http文件服务
func (this *Context) ServeFile(filePath string) {
	this.Lock()
	defer this.Unlock()
	if !this.writeable {
		return
	}
	http.ServeFile(this.Response, this.Request, filePath)
	this.writeable = false
	return
}

func (this *Context) DownloadContent(fileName string, data []byte) {
	this.SetHeader("Content-disposition", fmt.Sprintf("attachment;filename=%s", fileName))
	_, _ = this.Response.Write(data)
	return
}
