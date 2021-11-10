package middleware

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

type MemoryInfo struct {

	// 堆对象申请的总内存空间, 申请就会增长, 该值为累计值[bytes]
	TotalHeapAlloc uint64

	// 堆对象申请的总内存空间[bytes]
	HeapAlloc uint64

	// 堆对象申请的总内存空间[bytes]
	Sys uint64

	// 堆对象数量
	NumObjects uint64

	// 释放的对对象数量
	NumFreeObjects uint64

	// cpu数量
	CpuCount int

	// 当前 goroutine 数量
	NumGoroutines int

	// 当前进程cgo calls 数量
	NumCgoCalls int64

	NumGC uint32

	LastGC uint64
}

func MemoryUsage() MemoryInfo {
	res := MemoryInfo{}
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	res.TotalHeapAlloc = memStats.TotalAlloc
	res.Sys = memStats.Sys
	res.HeapAlloc = memStats.HeapAlloc
	res.NumObjects = memStats.HeapObjects
	res.NumFreeObjects = memStats.Frees
	res.NumGC = memStats.NumGC
	res.LastGC = memStats.LastGC
	res.CpuCount = runtime.NumCPU()
	res.NumGoroutines = runtime.NumGoroutine()
	res.NumCgoCalls = runtime.NumCgoCall()
	return res
}

type Checker struct {
	Name        string
	Times       int
	TotalMillis int
}

var checkers = map[string]Checker{}

func Printf(formatter string, items ...interface{}) {
	fmt.Printf(formatter+"\n", items...)
}

func StackTrace() string {
	return string(debug.Stack())
}

func NewChecker(name string) Checker {

	check, has := checkers[name]
	if has {
		check.Times += 1
	}
	return Checker{
		Name:        "",
		Times:       0,
		TotalMillis: 0,
	}
}

func EndChecker(checker Checker) {

}
