package middleware

import (
	"fmt"
	"strings"
)

type MetricsData struct {
	Key   string
	Value int64
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

func PrintMetricsData(data []MetricsData, context Context) {
	context.OK(defaultContentType, []byte(GetMetricsData(data)))
}

func GetMetricsData(data []MetricsData) string {
	var lines []string
	for _, metric := range data {
		lines = append(lines, FormatMetricsData(metric))
	}
	return strings.Join(lines, "\n")
}
