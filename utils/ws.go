package utils

import (
	"fmt"
	"strings"
	"time"
)

// ParseDelays 解析延迟时间字符串为 time.Duration 列表
func ParseDelays(delays string) []time.Duration {
	var result []time.Duration
	for d := range strings.SplitSeq(delays, ",") {
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}
		var t time.Duration
		if _, err := fmt.Sscanf(d, "%d", &t); err != nil {
			continue
		}
		result = append(result, t*time.Second)
	}
	if len(result) == 0 {
		result = []time.Duration{1 * time.Second}
	}
	return result
}
