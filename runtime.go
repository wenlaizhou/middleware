package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// ServiceEndpoint 服务注册
type ServiceEndpoint struct {

	// 运行时信息
	RuntimeInfo FullRuntimeInfo `json:"runtimeInfo"`

	// 服务名称
	Name string `json:"name"`

	// 服务状态
	Status string `json:"status"`

	// 注册时间
	RegisterTime time.Time
}

// RegisterEndpointService 注册注册服务
func RegisterEndpointService() (map[string]ServiceEndpoint, *sync.RWMutex) {

	services := map[string]ServiceEndpoint{}

	lock := sync.RWMutex{}

	RegisterHandler("/service/endpoint/registry", func(context Context) {
		endpoint := ServiceEndpoint{}
		if err := json.Unmarshal(context.GetBody(), &endpoint); err == nil {
			lock.Lock()
			defer lock.Unlock()
			endpoint.RegisterTime = time.Now()
			services[endpoint.Name] = endpoint
		}
	})

	go func() {
		for {
			time.Sleep(time.Second * 60 * 10)
			lock.Lock()
			defer lock.Unlock()
			for name, ep := range services {
				if time.Now().Sub(ep.RegisterTime) >= time.Second*500 {
					ep.Status = "offline"
					services[name] = ep
				}
			}
		}
	}()

	return services, &lock
}

// RegisterEndpoint 注册到注册中心
func RegisterEndpoint(server string, name string, status string) {
	param := ServiceEndpoint{
		RuntimeInfo: GetFullRuntimeInfo(),
		Name:        name,
		Status:      status,
	}
	if code, _, _, err := PostJsonWithTimeout(
		10, fmt.Sprintf("%v/service/endpoint/registry", server),
		param); err != nil || code != 200 {
		errorMsg := ""
		if err != nil {
			errorMsg = err.Error()
		}
		mLogger.ErrorF("注册服务错误: %v, %v", code, errorMsg)
	}
}

type FullRuntimeInfo struct {
	Memory      MemoryInfo   `json:"memory"`
	OsMemory    OsMemoryInfo `json:"osMemory"`
	Version     VersionInfo  `json:"version"`
	CpuCount    int          `json:"cpuCount"`
	CurrentDisk DiskInfo     `json:"currentDisk"`
	DiskInfos   []DiskInfo   `json:"diskInfos"`
	Load        LoadInfo     `json:"load"`
	Net         NetInfo      `json:"net"`
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

type NetInfo struct {
	Connections []string `json:"connections"` // 连入地址
	Interfaces  []string `json:"interfaces"`  // 网络地址列表
	Listens     []string `json:"listens"`     // 监听端口列表
	TimeWaits   []string `json:"time_waits"`  // time wait
	CloseWaits  []string `json:"close_waits"` // close wait
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
		Net:         GetNetInfo(),
	}
}

func GetNetInfo() NetInfo {

	// netConnectionKindMap :
	// 	"all":   {kindTCP4, kindTCP6, kindUDP4, kindUDP6},
	//	"tcp":   {kindTCP4, kindTCP6},
	//	"tcp4":  {kindTCP4},
	//	"tcp6":  {kindTCP6},
	//	"udp":   {kindUDP4, kindUDP6},
	//	"udp4":  {kindUDP4},
	//	"udp6":  {kindUDP6},
	//	"inet":  {kindTCP4, kindTCP6, kindUDP4, kindUDP6},
	//	"inet4": {kindTCP4, kindUDP4},
	//	"inet6": {kindTCP6, kindUDP6},
	interfaces := []string{}
	if ifs, err := net.Interfaces(); err == nil {
		for _, netAddr := range ifs {
			interfaces = append(interfaces, netAddr.String())
		}
	} else {
		mLogger.ErrorF("获取网络接口错误: %v", err.Error())
	}

	// // http://students.mimuw.edu.pl/lxr/source/include/net/tcp_states.h
	// var tcpStatuses = map[string]string{
	//	"01": "ESTABLISHED",
	//	"02": "SYN_SENT",
	//	"03": "SYN_RECV",
	//	"04": "FIN_WAIT1",
	//	"05": "FIN_WAIT2",
	//	"06": "TIME_WAIT",
	//	"07": "CLOSE",
	//	"08": "CLOSE_WAIT",
	//	"09": "LAST_ACK",
	//	"0A": "LISTEN",
	//	"0B": "CLOSING",
	// }
	connectionList := []string{}
	listens := []string{}
	timewaits := []string{}
	closewaits := []string{}
	if conns, err := net.Connections("tcp"); err == nil {
		for _, conn := range conns {
			switch conn.Status {
			case "LISTEN":
				listens = append(listens, conn.String())
				break
			case "TIME_WAIT":
				timewaits = append(timewaits, conn.String())
				break
			case "CLOSE_WAIT":
				closewaits = append(closewaits, conn.String())
			case "ESTABLISHED":
				connectionList = append(connectionList, conn.String())
				break
			}
		}
	} else {
		mLogger.ErrorF("获取网络接口错误: %v", err.Error())
	}

	return NetInfo{
		Connections: connectionList,
		Interfaces:  interfaces,
		Listens:     listens,
		TimeWaits:   timewaits,
		CloseWaits:  closewaits,
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

func RegisterRuntimeInfoService(path string, enableMetrics bool) []SwaggerPath {
	res := []SwaggerPath{
		SwaggerBuildPath(path, "middleware", "get", "memInfo"),
	}
	RegisterHandler(path, func(context Context) {
		context.ApiResponse(0, "", GetFullRuntimeInfo())
	})
	// if enableMetrics {
	// 	res = append(res, SwaggerBuildPath("/metrics", "middleware", "get", "prometheus endpoint"))
	// 	RegisterHandler("/metrics", func(context Context) {
	// 		runtimeInfo := GetFullRuntimeInfo()
	// 		resp := []string{"# middleware"}
	// 		resp = append(resp, fmt.Sprintf("mem_sys %v", runtimeInfo.OsMemory.Total))
	// 		resp = append(resp, fmt.Sprintf("numObjects %v", mem.NumObjects))
	// 		resp = append(resp, fmt.Sprintf("numFreeObjects %v", mem.NumFreeObjects))
	// 		resp = append(resp, fmt.Sprintf("cpuCount %v", mem.CpuCount))
	// 		resp = append(resp, fmt.Sprintf("numGoroutines %v", mem.NumGoroutines))
	// 		resp = append(resp, fmt.Sprintf("numGc %v", runtimeInfo.Memory.NumGC))
	// 		resp = append(resp, fmt.Sprintf("", runtimeInfo.Load.Load15))
	// 		context.OK(Plain, []byte(strings.Join(resp, "\n")))
	// 	})
	// }
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
