package task

import (
	"concurrent-downloader/utils/tools"
	"concurrent-downloader/utils/web"
	"fmt"
)

type BaseTask struct {
	TaskId      string
	Concurrency int
	TotalSize   int64
	URL         string
	FilePath    string
	FileName    string
}

func NewBaseTask(concurrency int, url string, filePath string, fileName string) (*BaseTask, error) {
	taskId := tools.GenerateTaskID(url)
	totalSize, err := web.GetTotalSize(url)
	if err != nil {
		return nil, fmt.Errorf("获取文件大小失败: %w", err)
	}
	return &BaseTask{
		TaskId:      taskId,
		Concurrency: concurrency,
		TotalSize:   totalSize,
		URL:         url,
		FilePath:    filePath,
		FileName:    fileName,
	}, nil
}
