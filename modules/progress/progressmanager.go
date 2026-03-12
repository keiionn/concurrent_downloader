package progress

import (
	"concurrent-downloader/modules/logger"
	"concurrent-downloader/modules/task"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ProgressManager 进度管理器 关注下载进度等数据
type ProgressManager struct {
	Progresses map[string]*task.ProgressTask
	Mu         sync.RWMutex
}

// NewProgressManager 创建一个新的进度管理器
func NewProgressManager(maxTasks int) *ProgressManager {
	return &ProgressManager{
		Progresses: make(map[string]*task.ProgressTask),
	}
}

// AddProgress 添加一个新的下载进度
func (pm *ProgressManager) AddProgress(taskID string, nowTask *task.ProgressTask) {
	pm.Mu.Lock()
	defer pm.Mu.Unlock()
	pm.Progresses[taskID] = nowTask
}

/* // StartTask 开始一个新任务
func (pm *ProgressManager) StartTask(taskID string, nowTask *task.ProgressTask) {
	pm.Mu.Lock()
	defer pm.Mu.Unlock()
	pm.Progresses[taskID].Status = "downloading"
	pm.Progresses[taskID].StartTime = time.Now()
} */

// StartMonitor 启动一个后台协程，定时计算所有任务的速度和进度
func (pm *ProgressManager) StartMonitor() {
	go func() {
		// 每秒触发一次
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			pm.Mu.Lock()
			for _, t := range pm.Progresses {
				// 仅对正在下载的任务计算速度
				if t.Status == "downloading" {
					// 1. 获取当前下载总量（原子读取，防止并发冲突）
					currentSize := atomic.LoadInt64(&t.DownloadedSize)

					// 2. 计算速度：(当前总量 - 上一秒总量) / 1秒
					// 结果单位为 KB/s
					delta := currentSize - t.LastSize
					t.Speed = float64(delta) / 1024.0

					// 3. 更新进度百分比
					if t.TaskInfo.TotalSize > 0 {
						t.Progress = float64(currentSize) / float64(t.TaskInfo.TotalSize) * 100
					}

					// 4. 更新耗时
					t.ElapsedTime = time.Since(t.StartTime)

					// 5. 保存当前快照，供下一秒计算使用
					t.LastSize = currentSize
				}
			}
			pm.Mu.Unlock()
		}
	}()
}

// CompleteProgress 标记任务完成
func (pm *ProgressManager) CompleteProgress(taskID string, err error) {
	pm.Mu.Lock()
	defer pm.Mu.Unlock()
	if progress, ok := pm.Progresses[taskID]; ok {
		if err != nil {
			progress.Status = "failed"
			logger.LogError("任务 %s 下载失败: %v", taskID, err)
		} else {
			progress.Status = "completed"
			progress.Progress = 100
		}
		progress.ElapsedTime = time.Since(progress.StartTime)
	}
}

func (pm *ProgressManager) DeleteProgress(taskID string) {
	pm.Mu.Lock()
	defer pm.Mu.Unlock()
	delete(pm.Progresses, taskID)
}

// GetProgress 获取指定任务的进度
func (pm *ProgressManager) GetProgress(taskID string) (*task.ProgressTask, error) {
	pm.Mu.RLock()
	defer pm.Mu.RUnlock()
	if progress, ok := pm.Progresses[taskID]; ok {
		return &task.ProgressTask{
			TaskInfo:    progress.TaskInfo,
			Speed:       progress.Speed,
			Progress:    progress.Progress,
			StartTime:   progress.StartTime,
			ElapsedTime: progress.ElapsedTime,
		}, nil
	}
	return nil, fmt.Errorf("任务 %s 不存在", taskID)
}

func (pm *ProgressManager) GetProgressManager() *ProgressManager {
	return pm
}

/* // GetAllProgress 获取所有任务的进度
func (pm *ProgressManager) GetAllProgress() map[string]*DownloadProgress {
	pm.Mu.RLock()
	defer pm.Mu.RUnlock()
	return maps.Clone(pm.Progresses)
} */
