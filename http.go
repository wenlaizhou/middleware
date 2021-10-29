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
	"sync/atomic"
	"time"
)

var mLogger = GetLogger("middleware")

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
	successAccess  int64
	successExpire  int64
	totalAccess    int64
	totalExpire    int64
	i18n           I18n
	enableI18n     bool
	sync.RWMutex
}

var globalServer = NewServer("", 0)

func StartServer(host string, port int) {
	globalServer.Lock()
	globalServer.Host = host
	globalServer.Port = port
	globalServer.Unlock()
	globalServer.Start()
}

func GetGlobalServer() Server {
	return globalServer
}

func NewServer(host string, port int) Server {
	srv := Server{
		Host:          host,
		Port:          port,
		CrossDomain:   true,
		hasIndex:      false,
		enableI18n:    false,
		baseTpl:       template.New("middleware.Base"),
		successAccess: 0,
		successExpire: 0,
		totalAccess:   0,
		totalExpire:   0,
	}

	srv.pathNodes = make(map[string]pathProcessor)
	return srv
}

func (this *Server) GetStatus() int {
	this.RLock()
	defer this.RUnlock()
	return this.status
}

func (this *Server) Start() {
	defer func() {
		if err := recover(); err != nil {
			mLogger.ErrorF("%v", err)
		}
	}()
	this.Lock()
	if this.status != 0 {
		this.Unlock()
		return
	}
	this.status = 1
	this.Unlock()
	http.HandleFunc("/", this.ServeHTTP)
	hostStr := fmt.Sprintf("%s:%d", this.Host, this.Port)
	mLogger.Info("server start " + hostStr)
	log.Fatal(http.ListenAndServe(hostStr, nil))
}

func (this *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(w, r)
	ctx.tpl = this.baseTpl
	ctx.restProcessors = this.restProcessors
	ctx.code = 200 // 是否合适
	if this.enableI18n {
		ctx.EnableI18n = true
		ctx.Message = this.i18n
	}
	if this.CrossDomain {
		ctx.SetHeader(AccessControlAllowOrigin, "*")
		ctx.SetHeader(AccessControlAllowMethods, METHODS)
		ctx.SetHeader(AccessControlAllowHeaders, "*")

		if strings.ToUpper(ctx.GetMethod()) == OPTIONS {
			_ = ctx.Code(202)
			return
		}
	}
	atomic.AddInt64(&this.totalAccess, 1)
	start := time.Now().UnixNano()
	for _, filterNode := range this.filter {
		if filterNode.pathReg.MatchString(r.URL.Path) {
			if !filterNode.handler(ctx) {
				atomic.AddInt64(&this.totalExpire, (time.Now().UnixNano()-start)/1000000)
				return
			}
		}
	}
	if this.hasIndex && r.URL.Path == "/" {
		this.index.handler(ctx)
		atomic.AddInt64(&this.totalExpire, (time.Now().UnixNano()-start)/(1000000))
		return
	}
	var handler func(Context)
	for _, pathNode := range this.pathNodes {
		if pathNode.pathReg.MatchString(r.URL.Path) {
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
		atomic.AddInt64(&this.totalExpire, (time.Now().UnixNano()-start)/(1000000))
		return
	}
	handler(ctx)
	if ctx.code == 200 {
		atomic.AddInt64(&this.successAccess, 1)
		atomic.AddInt64(&this.successExpire, (time.Now().UnixNano()-start)/(1000000))
	}
	atomic.AddInt64(&this.totalExpire, (time.Now().UnixNano()-start)/(1000000))
	return
}

func (this *Server) RegisterDefaultIndex() {
	this.RegisterIndex(func(context Context) {
		context.OK(Html, []byte(DefaultIndex))
	})
	this.RegisterHandler("/static/css/bootstrap.v5.min.css", func(context Context) {
		context.OK(Css, []byte(BootstrapCss))
	})
	this.RegisterHandler("/static/images/default_background.jpg", func(context Context) {
		context.OK(Jpeg, defaultBackground)
	})
}

func (this *Server) Static(path string) {
	if !strings.HasSuffix(path, "/") {
		path = fmt.Sprintf("%s/", path)
	}
	this.RegisterHandler(path, StaticProcessor)
}

func (this *Server) RegisterIndex(handler func(Context)) {
	this.Lock()
	defer this.Unlock()
	this.hasIndex = true
	this.index = pathProcessor{
		handler: handler,
	}
}

func (this *Server) RegisterFrontendDist(distPath string) {
	exp := regexp.MustCompile("\\.html$|\\.js$|\\.css$|\\.svg$|\\.icon$|\\.ico$|\\.png$|\\.jpg$|\\.jpeg$|\\.gif$")
	this.RegisterFilter("/.*", func(context Context) bool {
		if exp.MatchString(context.Request.URL.Path) {
			http.ServeFile(context.Response, context.Request, fmt.Sprintf("%s/%s", distPath, context.Request.URL.Path[1:]))
			return false
		}
		return true
	})
}

func RegisterDefaultIndex() {
	globalServer.RegisterDefaultIndex()
}

func RegisterIndex(handler func(Context)) {
	globalServer.RegisterIndex(handler)
}

func RegisterStatic(path string) {
	globalServer.Static(path)
}

func RegisterFrontendDist(distPath string) {
	globalServer.RegisterFrontendDist(distPath)
}

func (this *Server) RegisterTemplate(filePath string) {
	this.Lock()
	var err error
	this.baseTpl, err = includeTemplate(this.baseTpl, ".html", []string{filePath}...)
	if err != nil {
		mLogger.Error(err.Error())
	}
	this.Unlock()
	mLogger.InfoF("render template %v done!", filePath)
}

func RegisterTemplate(filePath string) {
	globalServer.RegisterTemplate(filePath)
}

// warning: 请在设置模板目录前使用
func (this *Server) TemplateFunc(name string, function interface{}) {
	this.Lock()
	defer this.Unlock()
	this.baseTpl.Funcs(template.FuncMap{
		name: function})
}

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

func RegisterHandler(path string, handler func(Context)) {
	globalServer.RegisterHandler(path, handler)
}

func (this *Server) EnableMetrics() {
	this.RegisterHandler("/metrics", func(context Context) {
		context.OK(Plain, []byte(GetMetricsData([]MetricsData{
			{
				Key:   "request_count",
				Value: int64(this.totalAccess),
			},
			{
				Key:   "request_time",
				Value: int64(this.totalExpire),
			},
			{
				Key:   "success_count",
				Value: int64(this.successAccess),
			},
			{
				Key:   "success_time",
				Value: int64(this.successExpire),
			},
		})))
	})
}

func (this *Server) SetI18n(name string) {
	if len(name) <= 0 {
		name = "message"
	}
	cn := LoadConfig(fmt.Sprintf("%s_cn.properties", name))
	en := LoadConfig(fmt.Sprintf("%s_en.properties", name))
	this.i18n = I18n{
		Cn: cn,
		En: en,
	}
	this.enableI18n = true
}

func EnableMetrics() {
	globalServer.EnableMetrics()
}

func SetI18n(name string) {
	globalServer.SetI18n(name)
}

var pathParamReg, _ = regexp.Compile("\\{(.*?)\\}")

func (this *Server) RegisterHandler(path string, handler func(Context)) {
	this.Lock()
	defer this.Unlock()
	if len(path) <= 0 {
		return
	}
	if handler == nil {
		return
	}
	if strings.HasSuffix(path, "/") {
		path = fmt.Sprintf("%s.*", path)
	} else {
		path = fmt.Sprintf("%s$", path)
	}

	if !strings.HasPrefix(path, "/") {
		path = fmt.Sprintf("/%s", path)
	}

	paramMather := pathParamReg.FindAllStringSubmatch(path, -1)

	var params []string

	for _, param := range paramMather {
		params = append(params, param[1])
		path = strings.Replace(path,
			param[0], "(.*)", -1)
	}

	pathReg, err := regexp.Compile(path)
	mLogger.InfoF("注册handler: %s", path)
	if !ProcessError(err) {
		this.pathNodes[path] = pathProcessor{
			pathReg: pathReg,
			handler: handler,
			params:  params,
		}
	}
}

func (this *Server) RegisterRestProcessor(processor func(model interface{}) interface{}) {
	this.Lock()
	this.restProcessors = append(this.restProcessors, processor)
	this.Unlock()
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
