package middleware

import (
	"fmt"
	"strings"
)

// create_time{cluster="TC", name="idc02-sre-kubernetes-00", value="2019-03-26 10:30:25 &#43;0800 CST"} 0
type MetricsData struct {
	Key   string
	Value string
	Tags  map[string]string
}

const metricsTpl = "%v%v %v"

func FormatMetricsData(data MetricsData) string {
	if len(data.Key) <= 0 {
		return ""
	}
	tagsStr := ""
	if len(data.Tags) > 0 {
		var tags []string
		for k, v := range data.Tags {
			tags = append(tags, fmt.Sprintf(`%v="%v"`, k, v))
		}
		tagsStr = fmt.Sprintf("{%v}", strings.Join(tags, ","))
	}
	return fmt.Sprintf(metricsTpl, data.Key, tagsStr, data.Value)
}
