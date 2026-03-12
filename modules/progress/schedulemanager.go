package progress

import (
	"context"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"concurrent-downloader/modules/logger"
	"concurrent-downloader/modules/task"
	"concurrent-downloader/utils/web"
)

// ScheduleManager 调度管理器
type ScheduleManager struct {
	MaxTasks  int           // 同时下载的最大文件数
	Semaphore chan struct{} // 信号量控制任务并发
	Wg        sync.WaitGroup

	Progresses map[string]*task.ScheduleTask
	Mu         sync.RWMutex
}

// NewScheduleManager 创建一个新的调度管理器
func NewScheduleManager(maxTasks int) *ScheduleManager {
	return &ScheduleManager{
		MaxTasks:   maxTasks,
		Semaphore:  make(chan struct{}, maxTasks),
		Wg:         sync.WaitGroup{},
		Progresses: make(map[string]*task.ScheduleTask),
	}
}

// AddSchedule 添加一个下载任务
func (sm *ScheduleManager) AddSchedule(downloadTask task.ScheduleTask, pm *ProgressManager) {

	taskID := downloadTask.TaskInfo.TaskId
	fileName := downloadTask.TaskInfo.FileName

	sm.Wg.Go(func() {

		sm.Semaphore <- struct{}{}
		defer func() { <-sm.Semaphore }()

		sm.Mu.Lock()

		sm.Progresses[taskID] = &downloadTask
		sm.Mu.Unlock()

		logger.LogInfo("开始任务: 任务ID=%s, 文件=%s, URL=%s", taskID, fileName, downloadTask.TaskInfo.URL)
		if err := sm.DownloadFileWithProgress(taskID, pm); err != nil {
			logger.LogError("任务失败: 任务ID=%s, 错误=%v", taskID, err)
		} else {
			logger.LogInfo("任务完成: 任务ID=%s, 文件=%s", taskID, fileName)
		}
	})
}

func (sm *ScheduleManager) DownloadFileWithProgress(taskID string, pm *ProgressManager) error {

	sm.Mu.RLock()
	schedulePtr := sm.Progresses[taskID]
	sm.Mu.RUnlock()

	pm.Mu.RLock()
	progressPtr := pm.Progresses[taskID]
	pm.Mu.RUnlock()

	if schedulePtr == nil {
		logger.LogError("任务 %s 未初始化", taskID)
		return errors.New("任务未初始化")
	}

	// 2. 创建与预分配临时文件
	tempFile := schedulePtr.TaskInfo.FilePath + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		logger.LogError("创建临时文件失败: %s, 错误=%v", taskID, err)
		return err
	}
	defer file.Close()

	if err := file.Truncate(schedulePtr.TaskInfo.TotalSize); err != nil {
		logger.LogError("预分配空间失败: %s, 错误=%v", taskID, err)
		return err
	}

	// 3. 统一分块逻辑
	// 即使是小文件，也可以直接走 web.DownloadRange，没必要专门写个单文件下载函数
	totalSize := schedulePtr.TaskInfo.TotalSize
	concurrency := schedulePtr.TaskInfo.Concurrency

	// 自动降级：如果文件太小，强制并发为 1
	if totalSize < 1024*1024 {
		concurrency = 1
	}
	chunkSize := totalSize / int64(concurrency)

	// 4. 并发下载执行
	var wg sync.WaitGroup
	errChan := make(chan error, concurrency)

	// 建议：此处传入带超时的 Context，而不是 Background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	progressPtr.StartTime = time.Now()
	progressPtr.Status = "downloading"

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			start := int64(index) * chunkSize
			end := start + chunkSize - 1
			if index == concurrency-1 {
				end = totalSize - 1
			}

			err := web.DownloadRange(ctx, schedulePtr.TaskInfo.URL, file, start, end, func(n int64) {
				atomic.AddInt64(&progressPtr.DownloadedSize, n)
			})

			if err != nil {
				errChan <- err
				cancel() // 一个分块失败，取消其他所有分块，节省流量
			}
		}(i)
	}
	logger.LogDebug("等待所有分块完成: %s", taskID)
	wg.Wait()
	logger.LogDebug("所有分块已返回，准备关闭文件: %s", taskID)
	close(errChan)

	// 5. 错误检查
	for e := range errChan {
		if e != nil {
			return e
		}
	}
	file.Close()
	logger.LogDebug("文件已关闭，准备重命名: %s", taskID)
	// 6. 收尾工作：重命名
	if err := os.Rename(tempFile, schedulePtr.TaskInfo.FilePath); err != nil {
		logger.LogError("重命名文件失败: %s, 错误=%v", taskID, err)
		return err
	}

	progressPtr.Status = "completed"
	logger.LogInfo("任务 %s 完成，耗时 %v", taskID, time.Since(progressPtr.StartTime))
	return nil
}
