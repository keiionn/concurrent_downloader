package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"concurrent-downloader/modules/display"
	"concurrent-downloader/modules/logger"
	"concurrent-downloader/modules/progress"
	"concurrent-downloader/modules/task"
)

func main() {
	fmt.Println("=== 简易多线程下载器 ===")
	fmt.Println("支持多文件同时下载，每个文件分段下载")
	fmt.Println()

	// 从文件加载任务
	file, err := os.Open("tasks.json")
	if err != nil {
		logger.LogError("打开任务文件失败: %v", err)
		return
	}
	defer file.Close()

	var config struct {
		Tasks []task.BaseTask `json:"tasks"`
	}

	if err := json.NewDecoder(file).Decode(&config); err != nil {
		logger.LogError("解析任务文件失败: %v", err)
		return
	}
	logger.LogInfo("成功加载 %d 个任务", len(config.Tasks))
	// 准备任务数据
	tasks := config.Tasks

	scheduleManager := progress.NewScheduleManager(len(config.Tasks))
	progressManager := progress.NewProgressManager(len(config.Tasks))

	// 1. 【新增】启动速度监控器（计算每秒下载量）
	progressManager.StartMonitor()

	for i := 0; i < len(config.Tasks); i++ {
		t, err := task.NewBaseTask(tasks[i].Concurrency, tasks[i].URL, tasks[i].FilePath, tasks[i].FileName)
		if err != nil {
			logger.LogError("创建任务失败: %s, 错误=%v", t.TaskId, err)
			return
		}
		tasks = append(tasks, *t)
	}

	if task.CheckHasRepeatedFileInfo(tasks) {
		logger.LogError("任务配置中存在重复的文件路径或文件名")
		return
	}

	if err := os.MkdirAll("downloads", 0755); err != nil {
		logger.LogError("创建下载目录失败: %v", err)
		return
	}

	// 2. 【新增】启动动态 UI 刷新协程
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				display.DisplayProgress(progressManager)
				time.Sleep(500 * time.Millisecond) // 每 0.5 秒刷新一次
			}
		}
	}()

	// 启动任务
	for _, nowTask := range tasks {
		// 注意顺序：先 AddProgress 确保 display 能找到这个 ID，再 AddSchedule 启动下载
		pTask, _ := task.NewProgressTask(&nowTask)
		progressManager.AddProgress(nowTask.TaskId, pTask)

		scheduleManager.AddSchedule(task.ScheduleTask{
			TaskInfo: &nowTask,
		}, progressManager)

		time.Sleep(200 * time.Millisecond)
	}

	// 3. 等待所有下载任务协程结束
	scheduleManager.Wg.Wait()

	// 4. 收尾：停止 UI 协程并做最后一次打印
	cancel()
	time.Sleep(200 * time.Millisecond) // 给 UI 协程一点退出的时间
	display.DisplayProgress(progressManager)

	fmt.Println("\n所有任务处理完毕")
}
