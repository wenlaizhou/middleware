package middleware

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

var LogFormatter = "%s %v"

// 日志对象
type Logger interface {
	// 打印字符串类型日志
	Log(msg string)

	// 打印格式化内容日志
	LogF(formatter string, records ...interface{})

	// 打印模板类型日志
	LogTemplate(template string, models ...interface{})

	// 打印字符串类型日志
	Info(msg string)

	// 逐行打印info日志信息
	InfoLn(v ...interface{})

	// 打印格式化内容日志
	InfoF(formatter string, records ...interface{})

	// 打印模板类型日志
	InfoTemplate(template string, models ...interface{})

	// 打印字符串类型日志
	Warn(msg string)

	// 打印格式化内容日志
	WarnF(formatter string, records ...interface{})

	// 打印模板类型日志
	WarnTemplate(template string, models ...interface{})

	// 打印字符串类型日志
	Error(msg string)

	// 打印格式化内容日志
	ErrorF(formatter string, records ...interface{})

	// 打印模板类型日志
	ErrorTemplate(template string, models ...interface{})
}

type logger struct {
	*log.Logger
	fs *os.File
}

// 获取日志服务
func GetLogger(name string) Logger {
	res, hasEle := loggerContainer[name]
	if hasEle {
		return &res
	}
	loggerLocker.Lock()
	err := os.Mkdir("logs", os.ModePerm)
	if err != nil {
		// filepath exist
	}
	fs, err := os.OpenFile(fmt.Sprintf("logs/%s.log", name), os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	res = logger{
		Logger: log.New(fs, "", log.LstdFlags),
		fs:     fs,
	}
	loggerContainer[name] = res
	loggerLocker.Unlock()
	return &res
}

// 获取日志服务
func GetLoggerClear(name string) Logger {
	res, hasEle := loggerContainer[name]
	if hasEle {
		return &res
	}
	loggerLocker.Lock()
	err := os.Mkdir("logs", os.ModePerm)
	if err != nil {
		// filepath exist
	}
	fs, err := os.OpenFile(fmt.Sprintf("logs/%s.log", name), os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	res = logger{
		Logger: log.New(fs, "", log.Lmsgprefix),
		fs:     fs,
	}
	loggerContainer[name] = res
	loggerLocker.Unlock()
	return &res
}

// 注册日志滚动服务
//
// seconds: 设置日志滚动时间 单位: 秒
func RegisterLogRotate(seconds int) {

	// logger rotate
	Schedule("logger-rotate", seconds, func() {
		RotateLog()
	})

}

// 日志滚动
func RotateLog() {
	loggerLocker.Lock()
	for loggerName, loggerInstance := range loggerContainer {
		loggerFilename := fmt.Sprintf("logs/%s.log", loggerName)
		backupName := fmt.Sprintf("logs/%s.%s.log", loggerName, time.Now().Format("2006-1-2_15-04-05"))
		if loggerInstance.fs == nil {
			fs, err := os.OpenFile(loggerFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
			if err != nil {
				mLogger.ErrorF("%s, 该logger没有对应文件, 并无法创建该logger, 忽略该logger %s", loggerFilename, err.Error())
				continue
			}
			loggerInstance.fs = fs
			loggerInstance.Logger.SetOutput(fs)
			continue
		}
		stat, err := loggerInstance.fs.Stat()
		if err != nil {
			mLogger.ErrorF("%s, 获取状态错误, 忽略该logger %s", loggerFilename, err.Error())
			continue
		}
		if stat.Size() <= 0 {
			continue
		}
		err = loggerInstance.fs.Close()
		if err != nil {
			mLogger.ErrorF("%s, close, %s", loggerFilename, err.Error())
		}
		err = os.Rename(loggerFilename, backupName)
		if err != nil {
			mLogger.ErrorF("%s, 重名: %s, 错误, 日志滚动失败: %s", loggerFilename, backupName, err.Error())
			continue
		}
		fs, err := os.OpenFile(loggerFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
		if err != nil {
			mLogger.ErrorF("%s, 文件打开错误: %s", loggerFilename, err.Error())
		}
		loggerInstance.fs = fs
		loggerInstance.Logger.SetOutput(fs)
	}
	loggerLocker.Unlock()
}

var loggerContainer = map[string]logger{}

var loggerLocker = sync.Mutex{}

// 记录一行日志
func (this *logger) Log(msg string) {
	this.Println(msg)
}

// 记录一行格式化日志
func (this *logger) LogF(formatter string, records ...interface{}) {
	this.Printf(formatter, records...)
}

// 记录模板日志
func (this *logger) LogTemplate(tpl string, models ...interface{}) {
	this.Printf(tpl, models...)
}

// 记录一行日志
func (this *logger) Info(msg string) {
	this.Printf(LogFormatter, "INFO", msg)
}

func (this *logger) InfoLn(v ...interface{}) {
	this.Println(v...)
}

// 记录一行格式化日志
func (this *logger) InfoF(formatter string, records ...interface{}) {
	this.Printf(LogFormatter, "INFO", fmt.Sprintf(formatter, records...))
}

// 记录模板日志
func (this *logger) InfoTemplate(tpl string, models ...interface{}) {
	this.Printf(LogFormatter, "INFO", fmt.Sprintf(tpl, models...))
}

// 记录一行日志
func (this *logger) Error(msg string) {
	this.Printf(LogFormatter, "ERROR", msg)
}

// 记录一行格式化日志
func (this *logger) ErrorF(formatter string, records ...interface{}) {
	this.Printf(LogFormatter, "ERROR", fmt.Sprintf(formatter, records...))
}

// 记录模板日志
func (this *logger) ErrorTemplate(tpl string, models ...interface{}) {
	this.Printf(LogFormatter, "ERROR", fmt.Sprintf(tpl, models...))
}

// 记录一行日志
func (this *logger) Warn(msg string) {
	this.Printf(LogFormatter, "WARN", msg)
}

// 记录一行格式化日志
func (this *logger) WarnF(formatter string, records ...interface{}) {
	this.Printf(LogFormatter, "WARN", fmt.Sprintf(formatter, records...))
}

// 记录模板日志
func (this *logger) WarnTemplate(tpl string, models ...interface{}) {
	this.Printf(LogFormatter, "WARN", fmt.Sprintf(tpl, models...))
}
