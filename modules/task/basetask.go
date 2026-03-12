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

func CheckFilePath(filepath string) bool {
	return filepath != ""
}

func CheckFileName(filename string) bool {
	return filename != ""
}

func CheckConcurrency(concurrency int) bool {
	return concurrency > 0
}

func CheckURL(url string) bool {
	return url != ""
}

func CheckHasRepeatedFileInfo(tasks []BaseTask) bool {
	filePathMap := make(map[string]bool)
	for _, task := range tasks {
		if _, ok := filePathMap[task.FilePath]; ok {
			return true
		}
		filePathMap[task.FilePath] = true
	}
	fileNameMap := make(map[string]bool)
	for _, task := range tasks {
		if _, ok := fileNameMap[task.FileName]; ok {
			return true
		}
		fileNameMap[task.FileName] = true
	}
	return false
}
