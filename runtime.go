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
	res.CpuCount = runtime.NumCPU()
	res.NumGoroutines = runtime.NumGoroutine()
	res.NumCgoCalls = runtime.NumCgoCall()
	return res
}

func Printf(formatter string, items ...interface{}) {
	fmt.Printf(formatter+"\n", items...)
}

func StackTrace() string {
	return string(debug.Stack())
}
