package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Context 上下文数据结构
type Context struct {
	Request        *http.Request
	Response       http.ResponseWriter
	body           []byte
	tpl            *template.Template
	restProcessors []func(model interface{}) interface{}
	writeable      bool
	code           int
	Message        I18n
	EnableI18n     bool
	pathParams     map[string]string
}

type I18n struct {
	Cn map[string]string
	En map[string]string
}

func (c *Context) GetPathParam(key string) string {
	value, ok := c.pathParams[key]
	if ok {
		return value
	}
	return ""
}

// GetBody 获取请求体
func (c *Context) GetBody() []byte {
	if len(c.body) > 0 {
		return c.body
	}
	data, err := ioutil.ReadAll(c.Request.Body)
	c.body = data
	if err == nil && len(data) > 0 {
		c.body = data
		return c.body
	}
	return nil
}

// GetJSON 获取body中
//
// json类型数据体
func (c *Context) GetJSON() (map[string]interface{}, error) {
	res := make(map[string]interface{})
	if len(c.GetBody()) > 0 {
		err := json.Unmarshal(c.GetBody(), &res)
		return res, err
	}
	return res, nil
}

// GetJsonParamStr 获取json对象中key对应的字符串
//
// 没有该key则返回""
func GetJsonParamStr(key string, jsonObj map[string]interface{}) string {
	val, hasVal := jsonObj[key]
	if !hasVal {
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf("%v", val))
}

// GetQueryParam 获取query参数
func (c *Context) GetQueryParam(key string) string {
	return c.Request.URL.Query().Get(key)
}

// GetUri 获取去除querystring之后的请求路径
//
// 以 / 为开头
func (c *Context) GetUri() string {
	return c.Request.URL.Path
}

// WriteJSON 返回json类型数据
func (c *Context) WriteJSON(data interface{}) {
	res, err := json.Marshal(data)
	if err != nil {
		mLogger.ErrorF("json marshal error : %v", err.Error())
		return
	}
	c.OK(ApplicationJson, res)
}

// 获取content-type值
func (c *Context) GetContentType() string {
	return c.Request.Header.Get(ContentType)
}

// 获取header对应值
func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

// 获取cookie值
func (c *Context) GetCookie(key string) string {
	cook, err := c.Request.Cookie(key)
	if err != nil {
		return ""
	}
	return cook.Value
}

// 设置cookie值
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Response, cookie)
}

// 获取locale设置
func (c *Context) Locale() string {
	return c.GetCookie("locale")
}

// 302跳转
func (c *Context) Redirect(path string) error {
	if !c.writeable {
		return errors.New("禁止重复写入response")
	}
	c.writeable = false
	c.code = http.StatusFound
	http.Redirect(c.Response, c.Request, path, http.StatusFound)
	return nil
}

func (c *Context) ProxyPass(path string, timeoutSeconds int) {
	client := http.Client{
		Timeout: time.Second * time.Duration(timeoutSeconds),
	}
	req, _ := http.NewRequest(strings.ToUpper(c.GetMethod()), path, c.Request.Body)
	req.Header = c.Request.Header.Clone()
	resp, err := client.Do(req)
	if err != nil {
		c.Error(500, err.Error())
		return
	}
	for k, _ := range resp.Header {
		c.SetHeader(k, resp.Header.Get(k))
	}
	bodyStr := ""
	body, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		bodyStr = string(body)
	}
	c.Error(resp.StatusCode, bodyStr)
	return
}

// 返回http: 200
func (c *Context) OK(contentType string, content []byte) {
	if !c.writeable {
		mLogger.Error("禁止重复写入response")
		return
	}
	c.writeable = false
	if len(contentType) > 0 {
		c.SetHeader(ContentType, contentType)
	}
	c.SetHeader("server", "framework")
	c.code = 200
	_, err := c.Response.Write(content)
	if err != nil {
		mLogger.ErrorF("context response Ok error : %v", err.Error())
		return
	}
}

// 返回对应http编码
func (c *Context) Code(static int) {
	if !c.writeable {
		mLogger.Error("禁止重复写入response")
		return
	}
	c.writeable = false
	c.SetHeader("server", "framework")
	c.code = static
	c.Response.WriteHeader(static)
	return
}

// 返回http错误响应
//
// 请自行设定 contentType
func (c *Context) Error(static int, htmlStr string) {
	if !c.writeable {
		mLogger.Error("禁止重复写入response")
		return
	}
	c.writeable = false
	c.SetHeader("server", "framework")
	c.code = static
	c.Response.WriteHeader(static)
	_, _ = c.Response.Write([]byte(htmlStr))
	return
}

// 设置httpheader值
func (c *Context) SetHeader(key string, value string) {
	c.Response.Header().Set(key, value)
}

// DelHeader 删除httpheader对应值
func (c *Context) DelHeader(key string) {
	c.Response.Header().Del(key)
}

func newContext(w http.ResponseWriter, r *http.Request) Context {
	return Context{
		EnableI18n: false,
		writeable:  true,
		Response:   w,
		Request:    r,
		pathParams: map[string]string{},
	}
}

// GetMethod 获取http方法
func (c *Context) GetMethod() string {
	return c.Request.Method
}

// JSON 返回json数据
func (c *Context) JSON(jsonStr string) {
	c.OK(ApplicationJson, []byte(jsonStr))
}

// RemoteAddr 获取http请求address
func (c *Context) RemoteAddr() string {
	xForward := c.Request.Header.Get("x-forwarded-for")
	if len(xForward) > 0 {
		return xForward
	}
	realIp := c.Request.Header.Get("X-Real-IP")
	if len(realIp) > 0 {
		return realIp
	}
	return c.Request.RemoteAddr
}

// http文件服务
func (c *Context) ServeFile(filePath string) {
	if !c.writeable {
		return
	}
	http.ServeFile(c.Response, c.Request, filePath)
	c.writeable = false
	return
}

// 设置最后修改时间
func (c *Context) SetLastModified(modtime time.Time) {
	w := c.Response
	if !isZeroTime(modtime) {
		w.Header().Set("Last-Modified", modtime.UTC().Format(TimeFormat))
	}
}

// 写入未修改
func (c *Context) WriteNotModified() error {
	// RFC 7232 section 4.1:
	// a sender SHOULD NOT generate representation metadata other than the
	// above listed fields unless said metadata exists for the purpose of
	// guiding cache updates (e.g., Last-Modified might be useful if the
	// response does not have an ETag field).
	if !c.writeable {
		return errors.New("禁止重复写入response")
	}
	c.writeable = false
	w := c.Response
	h := w.Header()
	delete(h, "Content-Type")
	delete(h, "Content-Length")
	if h.Get("Etag") != "" {
		delete(h, "Last-Modified")
	}
	w.WriteHeader(StatusNotModified)
	return nil
}

// 下载二进制文件
func (c *Context) DownloadContent(fileName string, data []byte) {
	c.SetHeader("Content-disposition", fmt.Sprintf("attachment;filename=%s", fileName))
	c.code = 200
	_, _ = c.Response.Write(data)
	return
}
