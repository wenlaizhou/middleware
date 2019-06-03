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
		Host:        host,
		Port:        port,
		CrossDomain: true,
		hasIndex:    false,
		baseTpl:     template.New("middleware.Base"),
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
	if this.CrossDomain {
		ctx.SetHeader(AccessControlAllowOrigin, "*")
		ctx.SetHeader(AccessControlAllowMethods, METHODS)
		ctx.SetHeader(AccessControlAllowHeaders, "*")

		if strings.ToUpper(ctx.GetMethod()) == OPTIONS {
			_ = ctx.Code(202)
			return
		}
	}

	for _, filterNode := range this.filter {
		if filterNode.pathReg.MatchString(r.URL.Path) {
			if !filterNode.handler(ctx) {
				return
			}
		}
	}

	if this.hasIndex && r.URL.Path == "/" {
		this.index.handler(ctx)
		return
	}

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
			pathNode.handler(ctx)
			return
		}
	}
	_ = ctx.Error(StatusNotFound, StatusNotFoundView)
	return
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

func (this *Server) TemplateFunc(name string, function interface{}) {
	this.Lock()
	defer this.Unlock()
	this.baseTpl.Funcs(template.FuncMap{
		name: function})
}

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

var pathParamReg, _ = regexp.Compile("\\{(.*?)\\}")

func RegisterHandler(path string, handler func(Context)) {
	globalServer.RegisterHandler(path, handler)
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

type triNode struct {
	path    string
	pathReg regexp.Regexp
	childs  []triNode
	data    *interface{}
}

type pathProcessor struct {
	pathReg *regexp.Regexp
	params  []string
	handler func(Context)
}

func StaticProcessor(ctx Context) {
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
