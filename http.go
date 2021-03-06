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
	starPathNodes  []starProcessor
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
	root           TrieNode
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
		root: TrieNode{
			Path:    "/",
			Handler: nil,
			Next:    map[string]TrieNode{},
		},
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
		this.root.Handler(ctx)
		atomic.AddInt64(&this.totalExpire, (time.Now().UnixNano()-start)/(1000000))
		return
	}
	handler := this.root.FindPath(r.URL.Path)
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

func (this *Server) Static(path string) {
	this.RegisterHandler(path, StaticProcessor)
}

func (this *Server) RegisterIndex(handler func(Context)) {
	this.Lock()
	defer this.Unlock()
	this.hasIndex = true
	this.root.Handler = handler
}

func RegisterIndex(handler func(Context)) {
	globalServer.RegisterIndex(handler)
}

func RegisterStatic(path string) {
	globalServer.Static(path)
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

func (this *Server) RegisterHandler(path string, handler func(Context)) {
	this.Lock()
	defer this.Unlock()
	if len(path) <= 0 {
		return
	}
	if handler == nil {
		return
	}
	mLogger.InfoF("注册handler: %s", path)
	this.root.AddPath(path, handler)
}

func (this *Server) RegisterRestProcessor(processor func(model interface{}) interface{}) {
	this.Lock()
	this.restProcessors = append(this.restProcessors, processor)
	this.Unlock()
	mLogger.InfoLn("新增restProcessor")
}

type pathProcessor struct {
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
