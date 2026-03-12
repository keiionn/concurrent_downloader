package tools

import (
	"math/rand"
	"strings"

	"concurrent-downloader/modules/logger"
)

// GenerateTaskID 根据URL生成任务ID
func GenerateTaskID(url string) string {
	hash := 0
	for _, c := range url {
		hash = hash*31 + (int(c)+rand.Intn(100))%100
	}
	hash &= 0x7FFFFFFF
	var res strings.Builder
	for hash > 0 {
		res.WriteString(string(rune(hash%10 + '0')))
		hash /= 10
	}
	logger.LogDebug("生成任务ID: %s, URL: %s", res.String(), url)
	return res.String()
}
