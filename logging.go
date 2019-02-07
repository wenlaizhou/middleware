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
