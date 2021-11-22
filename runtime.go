package middleware

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

type MemoryInfo struct {

	// 堆对象申请的总内存空间, 申请就会增长, 该值为累计值[bytes]
	TotalHeapAlloc uint64 `json:"totalHeapAlloc"`

	// 堆对象申请的总内存空间[bytes]
	HeapAlloc uint64 `json:"heapAlloc"`

	// 堆对象申请的总内存空间[bytes]
	Sys uint64 `json:"sys"`

	// 堆对象数量
	NumObjects uint64 `json:"numObjects"`

	// 释放的对对象数量
	NumFreeObjects uint64 `json:"numFreeObjects"`

	// cpu数量
	CpuCount int `json:"cpuCount"`

	// 当前 goroutine 数量
	NumGoroutines int `json:"numGoroutines"`

	// 当前进程cgo calls 数量
	NumCgoCalls int64 `json:"numCgoCalls"`

	NumGC uint32 `json:"numGc"`

	LastGC uint64 `json:"lastGc"`

	Os string `json:"os"`

	OsArch string `json:"osArch"`

	RuntimeVersion string `json:"runtimeVersion"`

	Dependencies []map[string]string `json:"dependencies"`
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
	res.Os = runtime.GOOS
	res.OsArch = runtime.GOARCH
	res.RuntimeVersion = runtime.Version()
	buildInfo, buildRes := debug.ReadBuildInfo()
	if buildRes {
		for _, dep := range buildInfo.Deps {
			res.Dependencies = append(res.Dependencies, map[string]string{
				"name":    dep.Path,
				"version": dep.Version,
			})
		}
	}
	return res
}

type CheckerResult struct {
	Name        string
	Times       int
	TotalMillis int64
}

type Checker struct {
	Start int64
	Name  string
}

var checkers = map[string]CheckerResult{}

func Printf(formatter string, items ...interface{}) {
	fmt.Printf(formatter+"\n", items...)
}

func StackTrace() string {
	return string(debug.Stack())
}

func checkHidden(key string, hiddens []string) bool {
	if hiddens == nil {
		return false
	}
	for _, h := range hiddens {
		if strings.Contains(key, h) {
			return true
		}
	}
	return false
}

func RegisterConfService(conf Config, path string, hidden string) SwaggerPath {

	hiddensList := []string{ConfDir}
	if len(hidden) >= 0 {
		hidden = strings.TrimSpace(hidden)
		hiddens := strings.Split(hidden, ",")
		for _, h := range hiddens {
			hiddensList = append(hiddensList, strings.TrimSpace(h))
		}
	}

	RegisterHandler(path, func(context Context) {
		resp := map[string]string{}
		for k, v := range conf {
			if checkHidden(k, hiddensList) {
				continue
			}
			resp[k] = v
		}
		if len(resp) <= 0 {
			context.ApiResponse(0, "", "")
			return
		}
		res := ""
		for k, v := range resp {
			if len(res) > 0 {
				res = fmt.Sprintf("%s\n%s = %s", res, k, v)
			} else {
				res = fmt.Sprintf("%s = %s", k, v)
			}
		}
		context.ApiResponse(0, "", res)
	})

	return SwaggerBuildPath(path, "middleware", "get", "config service")
}

func (thisSelf Checker) End() {
	end := TimeEpoch() - thisSelf.Start
	res, has := checkers[thisSelf.Name]
	if has {
		res.Times += 1
		res.TotalMillis += end
		checkers[thisSelf.Name] = res
	} else {
		checkers[thisSelf.Name] = CheckerResult{
			Name:        thisSelf.Name,
			Times:       1,
			TotalMillis: end,
		}
	}
}

func CheckersInfo() []CheckerResult {
	res := []CheckerResult{}
	for _, checker := range checkers {
		res = append(res, checker)
	}
	return res
}

func GetChecker(name string) Checker {
	return Checker{
		Start: TimeEpoch(),
		Name:  name,
	}
}
