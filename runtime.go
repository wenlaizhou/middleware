package middleware

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	"runtime"
	"runtime/debug"
	"strings"
)

type FullRuntimeInfo struct {
	Memory      MemoryInfo   `json:"memory"`
	OsMemory    OsMemoryInfo `json:"osMemory"`
	Version     VersionInfo  `json:"version"`
	CpuCount    int          `json:"cpuCount"`
	CurrentDisk DiskInfo     `json:"currentDisk"`
	DiskInfos   []DiskInfo   `json:"diskInfos"`
	Load        LoadInfo     `json:"load"`
}

type DiskInfo struct {
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	Fstype      string  `json:"fstype"`
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
}

type LoadInfo struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

type OsMemoryInfo struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
}

type VersionInfo struct {
	Dependencies   []map[string]string `json:"dependencies"`
	RuntimeVersion string              `json:"runtimeVersion"`
	Os             string              `json:"os"`
	OsArch         string              `json:"osArch"`
}

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
}

func GetFullRuntimeInfo() FullRuntimeInfo {
	return FullRuntimeInfo{
		Memory:      GetMemoryInfo(),
		OsMemory:    GetOsMemoryInfo(),
		Version:     GetVersionInfo(),
		CpuCount:    GetCpuCount(),
		CurrentDisk: GetDiskInfo(SelfDir()),
		DiskInfos:   GetDiskInfos(),
		Load:        GetLoadInfo(),
	}
}

func GetLoadInfo() LoadInfo {
	result := LoadInfo{}
	if info, err := load.Avg(); err == nil {
		result.Load1 = info.Load1
		result.Load5 = info.Load5
		result.Load15 = info.Load15
	}
	return result
}

func GetDiskInfos() []DiskInfo {
	result := []DiskInfo{}
	if partitions, err := disk.Partitions(true); err == nil {
		for _, p := range partitions {
			d := GetDiskInfo(p.Mountpoint)
			d.Mountpoint = p.Mountpoint
			d.Device = p.Device
			d.Fstype = p.Fstype
			result = append(result, d)
		}
	}
	return result
}

func GetDiskInfo(dir string) DiskInfo {
	res := DiskInfo{}
	if usage, err := disk.Usage(dir); err == nil {
		res.Total = usage.Total
		res.UsedPercent = usage.UsedPercent
		res.Used = usage.Used
		res.Free = usage.Free
		res.Mountpoint = dir
	}
	return res
}

func GetCpuCount() int {
	if info, err := cpu.Info(); err == nil {
		return len(info)
	}
	return -1
}

func GetOsMemoryInfo() OsMemoryInfo {
	res := OsMemoryInfo{}
	if memInfo, err := mem.VirtualMemory(); err == nil {
		res.Total = memInfo.Total
		res.UsedPercent = memInfo.UsedPercent
		res.Used = memInfo.Used
		res.Available = memInfo.Available
	}
	return res
}

func GetProcesses() ([]*process.Process, error) {
	return process.Processes()
}

func GetVersionInfo() VersionInfo {
	version := VersionInfo{}
	version.RuntimeVersion = runtime.Version()
	buildInfo, buildRes := debug.ReadBuildInfo()
	if buildRes {
		for _, dep := range buildInfo.Deps {
			version.Dependencies = append(version.Dependencies, map[string]string{
				"name":    dep.Path,
				"version": dep.Version,
			})
		}
	}
	version.Os = runtime.GOOS
	version.OsArch = runtime.GOARCH
	return version
}

func GetMemoryInfo() MemoryInfo {
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

	hiddenList := []string{ConfDir}
	if len(hidden) >= 0 {
		hidden = strings.TrimSpace(hidden)
		hiddens := strings.Split(hidden, ",")
		for _, h := range hiddens {
			hiddenList = append(hiddenList, strings.TrimSpace(h))
		}
	}

	RegisterHandler(path, func(context Context) {
		resp := map[string]string{}
		for k, v := range conf {
			if checkHidden(k, hiddenList) {
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

func RegisterMemInfoService(path string, enableMetrics bool) []SwaggerPath {
	res := []SwaggerPath{
		SwaggerBuildPath(path, "middleware", "get", "memInfo"),
	}
	RegisterHandler(path, func(context Context) {
		context.ApiResponse(0, "", GetFullRuntimeInfo())
	})
	if enableMetrics {
		res = append(res, SwaggerBuildPath("/metrics", "middleware", "get", "prometheus endpoint"))
		RegisterHandler("/metrics", func(context Context) {
			// runtimeInfo := GetFullRuntimeInfo()
			resp := []string{"# middleware"}

			// resp = append(resp, fmt.Sprintf("mem_sys %v", mem.Sys))
			// resp = append(resp, fmt.Sprintf("numObjects %v", mem.NumObjects))
			// resp = append(resp, fmt.Sprintf("numFreeObjects %v", mem.NumFreeObjects))
			// resp = append(resp, fmt.Sprintf("cpuCount %v", mem.CpuCount))
			// resp = append(resp, fmt.Sprintf("numGoroutines %v", mem.NumGoroutines))
			// resp = append(resp, fmt.Sprintf("numGc %v", mem.NumGC))
			context.OK(Plain, []byte(strings.Join(resp, "\n")))
		})
	}
	return res
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
