package middleware

import (
	"fmt"
	"log"
	"os"
	"sync"
)

var LogFormatter = "[%s] [%v]"

//日志对象
type Logger interface {
	//打印字符串类型日志
	Log(msg string)

	//打印格式化内容日志
	LogF(formatter string, records ...interface{})

	//打印模板类型日志
	LogTemplate(template string, models ...interface{})

	//打印字符串类型日志
	Info(msg string)

	//打印格式化内容日志
	InfoF(formatter string, records ...interface{})

	//打印模板类型日志
	InfoTemplate(template string, models ...interface{})

	//打印字符串类型日志
	Warn(msg string)

	//打印格式化内容日志
	WarnF(formatter string, records ...interface{})

	//打印模板类型日志
	WarnTemplate(template string, models ...interface{})

	//打印字符串类型日志
	Error(msg string)

	//打印格式化内容日志
	ErrorF(formatter string, records ...interface{})

	//打印模板类型日志
	ErrorTemplate(template string, models ...interface{})
}

type logger struct {
	log.Logger
}

var loggerContainer = map[string]logger{}

var loggerLocker = sync.Mutex{}

//记录一行日志
func (this *logger) Log(msg string) {
	this.Println(msg)
}

//记录一行格式化日志
func (this *logger) LogF(formatter string, records ...interface{}) {
	this.Printf(formatter, records...)
}

//记录模板日志
func (this *logger) LogTemplate(tpl string, models ...interface{}) {
	this.Printf(tpl, models...)
}

//记录一行日志
func (this *logger) Info(msg string) {
	this.Printf(LogFormatter, "info", msg)
}

//记录一行格式化日志
func (this *logger) InfoF(formatter string, records ...interface{}) {
	this.Printf(LogFormatter, "info", fmt.Sprintf(formatter, records...))
}

//记录模板日志
func (this *logger) InfoTemplate(tpl string, models ...interface{}) {
	this.Printf(LogFormatter, "info", fmt.Sprintf(tpl, models...))
}

//记录一行日志
func (this *logger) Error(msg string) {
	this.Printf(LogFormatter, "error", msg)
}

//记录一行格式化日志
func (this *logger) ErrorF(formatter string, records ...interface{}) {
	this.Printf(LogFormatter, "error", fmt.Sprintf(formatter, records...))
}

//记录模板日志
func (this *logger) ErrorTemplate(tpl string, models ...interface{}) {
	this.Printf(LogFormatter, "error", fmt.Sprintf(tpl, models...))
}

//记录一行日志
func (this *logger) Warn(msg string) {
	this.Printf(LogFormatter, "warn", msg)
}

//记录一行格式化日志
func (this *logger) WarnF(formatter string, records ...interface{}) {
	this.Printf(LogFormatter, "warn", fmt.Sprintf(formatter, records...))
}

//记录模板日志
func (this *logger) WarnTemplate(tpl string, models ...interface{}) {
	this.Printf(LogFormatter, "warn", fmt.Sprintf(tpl, models...))
}

//获取日志服务
func GetLogger(name string) Logger {
	res, hasEle := loggerContainer[name]
	if hasEle {
		return &res
	}
	loggerLocker.Lock()
	defer loggerLocker.Unlock()
	err := os.Mkdir("log", os.ModePerm)
	if err != nil {
		//filepath exist
	}
	fs, err := os.OpenFile(fmt.Sprintf("log/%s.log", name), os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	res = logger{
		*log.New(fs, "", log.LstdFlags),
	}
	loggerContainer[name] = res
	return &res
}

//默认日志对象
var DefaultLogger = GetLogger("console")
