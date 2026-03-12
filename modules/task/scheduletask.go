package task

type ScheduleTask struct {
	TaskInfo *BaseTask
}

func NewScheduledTask(baseTask *BaseTask) (*ScheduleTask, error) {
	return &ScheduleTask{
		TaskInfo: baseTask,
	}, nil
}
