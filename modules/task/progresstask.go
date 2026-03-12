package task

import (
	"time"
)

type ProgressTask struct {
	TaskInfo       *BaseTask
	Speed          float64
	StartTime      time.Time
	DownloadedSize int64
	LastSize       int64
	Progress       float64
	ElapsedTime    time.Duration
	Status         string
}

func NewProgressTask(baseTask *BaseTask) (*ProgressTask, error) {
	return &ProgressTask{
		TaskInfo:       baseTask,
		Speed:          0,
		StartTime:      time.Now(),
		DownloadedSize: 0,
		LastSize:       0,
		Progress:       0,
		ElapsedTime:    0,
		Status:         "waiting",
	}, nil
}
