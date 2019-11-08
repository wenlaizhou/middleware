package middleware

import (
	"encoding/json"
	"fmt"
)

// create_time{cluster="TC", name="idc02-sre-kubernetes-00", value="2019-03-26 10:30:25 &#43;0800 CST"} 0
type MetricsData struct {
	Key   string
	Value int
	Tags  map[string]string
}

const metricsTpl = "%v%v %v"

func FormatMetricsData(data MetricsData) string {
	if len(data.Key) <= 0 {
		return ""
	}
	tagsStr, err := json.Marshal(data.Tags)
	if err != nil {
		return fmt.Sprintf(metricsTpl, data.Key, "", data.Value)
	}
	return fmt.Sprintf(metricsTpl, data.Key, string(tagsStr), data.Value)
}
