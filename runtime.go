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

	// 服务属性
	Properties map[string]string `json:"properties"`

	// 注册时间
	RegisterTime time.Time
}

// RegisterEndpointService 注册注册服务
//
// enableQuery 是否启动注册中心查询服务
//
// peers 伙伴, 使用, 分隔, 例如: http://192.168.0.11,https://10.21.0.1
// key 是否有认证的key
func RegisterEndpointService(enableQuery bool, peers string, key string) ([]SwaggerPath, map[string]ServiceEndpoint, *sync.RWMutex) {

	swaggerRes := []SwaggerPath{}

	services := map[string]ServiceEndpoint{}

	peerList := []string{}

	peers = strings.TrimSpace(peers)

	if len(peers) > 0 {
		for _, peer := range strings.Split(peers, ",") {
			peer := strings.TrimSpace(peer)
			if len(peer) <= 0 {
				continue
			}
			if !strings.HasPrefix(peer, "http://") && !strings.HasPrefix(peer, "https://") {
				mLogger.ErrorF("注册中心服务peer错误: %v", peer)
				continue
			}
			peerList = append(peerList, peer)
		}
	}

	lock := sync.RWMutex{}

	RegisterHandler("/_service/endpoint/registry", func(context Context) {
		if len(key) > 0 {
			if context.GetHeader("registry-key") != key {
				context.ApiResponse(-1, "key error", nil)
				return
			}
		}
		endpoint := ServiceEndpoint{}
		if err := json.Unmarshal(context.GetBody(), &endpoint); err == nil {
			lock.Lock()
			endpoint.RegisterTime = time.Now()
			services[endpoint.Name] = endpoint
			lock.Unlock()
			if len(peerList) > 0 && len(context.GetQueryParam("noSpread")) <= 0 {
				for _, peer := range peerList {
					sendEndpoint(fmt.Sprintf("%v/_service/endpoint/registry", peer),
						endpoint.Name, endpoint.Status, endpoint.Properties, key, false)
				}
			}
		}
	})

	swaggerRes = append(swaggerRes, SwaggerBuildPath("/_service/endpoint/registry",
		"registry", "POST", "注册中心注册接口"))

	if enableQuery {
		RegisterHandler("/_service/endpoints", func(context Context) {
			lock.RLock()
			defer lock.RUnlock()
			context.ApiResponse(0, "", services)
		})
		swaggerRes = append(swaggerRes, SwaggerBuildPath("/_service/endpoints",
			"registry", "GET", "服务查询接口"))
	}

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

	return swaggerRes, services, &lock
}

// SendEndpoint 注册到注册中心
func SendEndpoint(server string, name string, status string, properties map[string]string, key string) {
	sendEndpoint(server, name, status, properties, key, true)
}

func sendEndpoint(server string, name string, status string, properties map[string]string, key string, spread bool) {
	param := ServiceEndpoint{
		RuntimeInfo: GetFullRuntimeInfo(),
		Name:        name,
		Status:      status,
		Properties:  properties,
	}
	headers := map[string]string{}
	if len(key) > 0 {
		headers["registry-key"] = key
	}
	url := fmt.Sprintf("%v/_service/endpoint/registry", server)
	if !spread {
		url = fmt.Sprintf("%v?noSpread=true", url)
	}
	if code, _, _, err := PostJsonFull(10,
		fmt.Sprintf("%v/_service/endpoint/registry", server),
		headers, param); err != nil || code != 200 {
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

func RegisterRuntimeInfoService(path string, labels map[string]string, enableMetrics bool) []SwaggerPath {
	res := []SwaggerPath{
		SwaggerBuildPath(path, "middleware", "get", "runtime info"),
	}
	if !enableMetrics {
		RegisterHandler(path, func(context Context) {
			context.ApiResponse(0, "", GetFullRuntimeInfo())
		})
		return res
	} else {
		RegisterHandler(path, func(context Context) {
			runtimeInfo := GetFullRuntimeInfo()
			metrics := []MetricsData{}
			var tag map[string]string
			if labels != nil && len(labels) > 0 {
				tag = labels
			} else {
				tag = map[string]string{
					"framework": "middleware",
				}
			}
			metrics = append(metrics, MetricsData{
				Key:   "connections",
				Value: int64(len(runtimeInfo.Net.Connections)),
				Tags:  tag,
			})
			metrics = append(metrics, MetricsData{
				Key:   "listens",
				Value: int64(len(runtimeInfo.Net.Listens)),
				Tags:  tag,
			})
			metrics = append(metrics, MetricsData{
				Key:   "timewaits",
				Value: int64(len(runtimeInfo.Net.TimeWaits)),
				Tags:  tag,
			})
			metrics = append(metrics, MetricsData{
				Key:   "closewaits",
				Value: int64(len(runtimeInfo.Net.CloseWaits)),
				Tags:  tag,
			})
			metrics = append(metrics, MetricsData{
				Key:   "node_memory_total",
				Value: int64(runtimeInfo.OsMemory.Total),
				Tags:  tag,
			})
			metrics = append(metrics, MetricsData{
				Key:   "node_memory_used",
				Value: int64(runtimeInfo.OsMemory.Used),
				Tags:  tag,
			})
			metrics = append(metrics, MetricsData{
				Key:   "node_cpu_total",
				Value: int64(runtimeInfo.CpuCount),
				Tags:  tag,
			})
			metrics = append(metrics, MetricsData{
				Key:   "current_disk_total",
				Value: int64(runtimeInfo.CurrentDisk.Total),
				Tags: map[string]string{
					"path": runtimeInfo.CurrentDisk.Mountpoint,
				},
			})
			metrics = append(metrics, MetricsData{
				Key:   "current_disk_used",
				Value: int64(runtimeInfo.CurrentDisk.Used),
				Tags: map[string]string{
					"path": runtimeInfo.CurrentDisk.Mountpoint,
				},
			})
			metrics = append(metrics, MetricsData{
				Key:   "node_load_1",
				Value: int64(runtimeInfo.Load.Load1),
				Tags:  tag,
			})
			metrics = append(metrics, MetricsData{
				Key:   "node_load_5",
				Value: int64(runtimeInfo.Load.Load5),
				Tags:  tag,
			})
			metrics = append(metrics, MetricsData{
				Key:   "node_load_15",
				Value: int64(runtimeInfo.Load.Load15),
				Tags:  tag,
			})

			PrintMetricsData(metrics, context)
			return
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
