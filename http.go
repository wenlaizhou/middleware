package middleware

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

var mLogger = GetLogger("middleware")

// Server对象
type Server struct {
	Host           string
	Port           int
	baseTpl        *template.Template
	pathNodes      map[string]pathProcessor
	index          pathProcessor
	restProcessors []func(model interface{}) interface{}
	hasIndex       bool
	CrossDomain    bool
	status         int
	filter         []filterProcessor
	i18n           I18n
	enableI18n     bool
	swagger        *SwaggerData
	sync.RWMutex
}

// 默认全局单一http服务
var globalServer = NewServer("", 0)

// 启动服务
func StartServer(host string, port int) {
	globalServer.Lock()
	globalServer.Host = host
	globalServer.Port = port
	globalServer.Unlock()
	globalServer.Start()
}

// 获取全局唯一Server
func GetGlobalServer() Server {
	return globalServer
}

// 创建服务
func NewServer(host string, port int) Server {
	srv := Server{
		Host:        host,
		Port:        port,
		CrossDomain: true,
		hasIndex:    false,
		enableI18n:  false,
		baseTpl:     template.New("middleware.Base"),
		swagger: &SwaggerData{
			Title:       "",
			Version:     "",
			Description: "",
			Host:        "",
		},
	}

	srv.pathNodes = make(map[string]pathProcessor)
	return srv
}

func (t *Server) GetStatus() int {
	t.RLock()
	defer t.RUnlock()
	return t.status
}

func (t *Server) Start() {
	defer func() {
		if err := recover(); err != nil {
			mLogger.ErrorF("%v", err)
		}
	}()
	t.Lock()
	if t.status != 0 {
		t.Unlock()
		return
	}
	t.status = 1
	t.Unlock()
	http.HandleFunc("/", t.ServeHTTP)
	hostStr := fmt.Sprintf("%s:%d", t.Host, t.Port)
	mLogger.Info("server start " + hostStr)
	log.Fatal(http.ListenAndServe(hostStr, nil))
}

var accessLogger = GetLogger("access")

// 核心处理逻辑
func (t *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := TimeEpoch()
	startTime := time.Now().Format(TimeFormat)
	ctx := newContext(w, r)
	ctx.tpl = t.baseTpl
	ctx.restProcessors = t.restProcessors
	ctx.code = 200 // 是否合适
	if t.enableI18n {
		ctx.EnableI18n = true
		ctx.Message = t.i18n
	}
	if t.CrossDomain {
		ctx.SetHeader(AccessControlAllowOrigin, "*")
		ctx.SetHeader(AccessControlAllowMethods, METHODS)
		ctx.SetHeader(AccessControlAllowHeaders, "*")

		if strings.ToUpper(ctx.GetMethod()) == OPTIONS {
			_ = ctx.Code(202)
			return
		}
	}
	for _, filterNode := range t.filter {
		if filterNode.pathReg.MatchString(r.URL.Path) {
			if !filterNode.handler(ctx) {
				return
			}
		}
	}
	if t.hasIndex && r.URL.Path == "/" {
		t.index.handler(ctx)
		return
	}
	var handler func(Context)
	for _, pathNode := range t.pathNodes {
		if pathNode.pathReg.MatchString(r.URL.Path) {
			if len(pathNode.params) <= 0 {
				handler = pathNode.handler
				break
			}
			pathParams := pathNode.pathReg.FindAllStringSubmatch(r.URL.Path, 10) // 最多10个路径参数
			if len(pathParams) > 0 && len(pathParams[0]) > 0 {
				for i, pathParam := range pathParams[0][1:] {
					if len(pathNode.params) < i+1 {
						break
					}
					ctx.pathParams[pathNode.params[i]] = pathParam
				}
			}
			handler = pathNode.handler
			break
		}
	}
	if handler == nil {
		_ = ctx.Error(StatusNotFound, StatusNotFoundView)
		return
	}
	handler(ctx)
	accessLogger.InfoF(`"%v", "%v", "%v", "%v", %v, %v`, startTime, ctx.Request.RequestURI, ctx.Request.RemoteAddr, ctx.GetMethod(), ctx.code, TimeEpoch()-start)
	return
}

func (t *Server) RegisterDefaultIndex(link string) {
	t.RegisterIndex(func(context Context) {
		context.OK(Html, []byte(fmt.Sprintf(DefaultIndex, link)))
	})
	t.RegisterHandler("/static/default/css/bootstrap.v5.min", func(context Context) {
		context.OK(Css, []byte(BootstrapCss))
	})
	t.RegisterHandler("/static/default/images/default_background", func(context Context) {
		context.OK(Jpeg, defaultBackground)
	})
}

//设置静态文件目录
func (t *Server) Static(path string) {
	if !strings.HasSuffix(path, "/") {
		path = fmt.Sprintf("%s/", path)
	}
	t.RegisterHandler(path, StaticProcessor)
}

// 注册首页
func (t *Server) RegisterIndex(handler func(Context)) {
	t.Lock()
	defer t.Unlock()
	t.hasIndex = true
	t.index = pathProcessor{
		handler: handler,
	}
}

// 结合 react 前端, 注册前端dist目录
func (t *Server) RegisterFrontendDist(distPath string) {
	exp := regexp.MustCompile("\\.html$|\\.js$|\\.css$|\\.svg$|\\.icon$|\\.ico$|\\.png$|\\.jpg$|\\.jpeg$|\\.gif$")
	t.RegisterFilter("/.*", func(context Context) bool {
		if exp.MatchString(context.Request.URL.Path) {
			http.ServeFile(context.Response, context.Request, fmt.Sprintf("%s/%s", distPath, context.Request.URL.Path[1:]))
			return false
		}
		return true
	})
	t.RegisterIndex(func(context Context) {
		http.ServeFile(context.Response, context.Request, fmt.Sprintf("%s/index.html", distPath))
	})
}

func RegisterDefaultIndex(link string) {
	globalServer.RegisterDefaultIndex(link)
}

// 注册首页处理器
func RegisterIndex(handler func(Context)) {
	globalServer.RegisterIndex(handler)
}

// 注册静态文件目录
func RegisterStatic(path string) {
	globalServer.Static(path)
}

// 注册前端编译后程序路径
func RegisterFrontendDist(distPath string) {
	globalServer.RegisterFrontendDist(distPath)
}

// 注册模板服务
func (t *Server) RegisterTemplate(filePath string) {
	t.Lock()
	var err error
	t.baseTpl, err = includeTemplate(t.baseTpl, ".html", []string{filePath}...)
	if err != nil {
		mLogger.Error(err.Error())
	}
	t.Unlock()
	mLogger.InfoF("render template %v done!", filePath)
}

// 注册模板服务
func RegisterTemplate(filePath string) {
	globalServer.RegisterTemplate(filePath)
}

// 注册模板函数
// warning: 请在设置模板目录前使用
func (t *Server) TemplateFunc(name string, function interface{}) {
	t.Lock()
	defer t.Unlock()
	t.baseTpl.Funcs(template.FuncMap{
		name: function})
}

// 注册模板函数
// warning: 请在设置模板目录前使用
func TemplateFunc(name string, function interface{}) {
	globalServer.TemplateFunc(name, function)
}

func includeTemplate(tpl *template.Template, suffix string, filePaths ...string) (*template.Template, error) {
	fileList := make([]string, 0)
	for _, filePath := range filePaths {
		info, err := os.Stat(filePath)
		if err != nil {
			mLogger.Error(err.Error())
			continue
		}
		if info.IsDir() {
			_ = filepath.Walk(filePath, func(path string, innerInfo os.FileInfo, err error) error {
				if !innerInfo.IsDir() {
					// 后缀名过滤
					if filepath.Ext(innerInfo.Name()) == suffix {
						fileList = append(fileList, path)
					}
				}
				return nil
			})
		} else {
			if filepath.Ext(filePath) == suffix {
				fileList = append(fileList, filePath)
			}
		}
	}
	mLogger.InfoLn("获取模板文件列表")
	mLogger.InfoLn(strings.Join(fileList, ","))
	if tpl == nil {
		return template.ParseFiles(fileList...)
	}
	return tpl.ParseFiles(fileList...)
}

// 注册http请求处理器
//
// @param path:路径, 可以用 {占位符} 进行路径参数设置
func RegisterHandler(path string, handler func(Context)) {
	globalServer.RegisterHandler(path, handler)
}

func (t *Server) SetI18n(name string) {
	if len(name) <= 0 {
		name = "message"
	}
	cn := LoadConfig(fmt.Sprintf("%s_cn.properties", name))
	en := LoadConfig(fmt.Sprintf("%s_en.properties", name))
	t.i18n = I18n{
		Cn: cn,
		En: en,
	}
	t.enableI18n = true
}

func SetI18n(name string) {
	globalServer.SetI18n(name)
}

var pathParamReg, _ = regexp.Compile("\\{(.+?)\\}")

// 注册服务
func (t *Server) RegisterHandler(path string, handler func(Context)) {
	t.Lock()
	defer t.Unlock()
	if len(path) <= 0 {
		return
	}
	if handler == nil {
		return
	}

	var params = []string{}

	paramMather := pathParamReg.FindAllStringSubmatch(path, -1)

	for _, param := range paramMather {
		params = append(params, param[1])
		path = strings.Replace(path,
			param[0], "(.+?)", -1)
	}

	if strings.HasSuffix(path, "/") {
		path = fmt.Sprintf("%s.*", path)
	} else {
		path = fmt.Sprintf("%s$", path)
	}

	if !strings.HasPrefix(path, "/") {
		path = fmt.Sprintf("/%s", path)
	}

	path = fmt.Sprintf("^%s", path)

	pathReg, err := regexp.Compile(path)
	mLogger.InfoF("注册handler: %s", path)
	if !ProcessError(err) {
		t.pathNodes[path] = pathProcessor{
			pathReg: pathReg,
			handler: handler,
			params:  params,
		}
	}
}

func (t *Server) RegisterRestProcessor(processor func(model interface{}) interface{}) {
	t.Lock()
	t.restProcessors = append(t.restProcessors, processor)
	t.Unlock()
	mLogger.InfoLn("新增restProcessor")
}

type pathProcessor struct {
	pathReg *regexp.Regexp
	params  []string
	handler func(Context)
}

type starProcessor struct {
	pathReg *regexp.Regexp
	handler func(Context)
}

func StaticProcessor(ctx Context) {
	ctx.code = 200
	http.ServeFile(ctx.Response, ctx.Request, ctx.Request.URL.Path[1:])
}

// 错误处理
//
// return true 错误发生
//
// false 无错误
func ProcessError(err error) bool {
	if err != nil {
		mLogger.Error(err.Error())
		return true
	}
	return false
}
