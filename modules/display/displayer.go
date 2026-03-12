package display

import (
	"fmt"
	"strings"

	"concurrent-downloader/modules/logger"
	"concurrent-downloader/modules/progress"
	"concurrent-downloader/utils/tools"
)

// DisplayProgress 显示下载进度
func DisplayProgress(pm *progress.ProgressManager) {
	pm.Mu.RLock()
	defer pm.Mu.RUnlock()

	fmt.Printf("\n\033[H\033[2J") // 清屏
	fmt.Println("=== 下载进度 ===")
	fmt.Printf("%-20s %-30s %-15s %-15s %-10s %-10s\n", "任务ID", "文件名", "进度", "速度", "状态", "耗时")
	fmt.Println(strings.Repeat("-", 120))

	for _, progress := range pm.Progresses {
		statusIcon := "📥"
		switch progress.Status {
		case "completed":
			statusIcon = "✅"
		case "failed":
			statusIcon = "❌"
		}

		fmt.Printf("%-20s %-30s %-14.1f%% %-14.1fKB/s %-10s %s\n",
			progress.TaskInfo.TaskId,
			tools.TruncateString(progress.TaskInfo.FileName, 30),
			progress.Progress,
			progress.Speed,
			statusIcon+" "+progress.Status,
			tools.FormatDuration(progress.ElapsedTime))

		// 记录进度日志
		logger.LogInfo("下载进度: 任务ID=%s, 文件=%s, 进度=%.1f%%, 速度=%.1fKB/s, 状态=%s, 耗时=%s",
			progress.TaskInfo.TaskId,
			progress.TaskInfo.FileName,
			progress.Progress,
			progress.Speed,
			progress.Status,
			tools.FormatDuration(progress.ElapsedTime))
	}
}
