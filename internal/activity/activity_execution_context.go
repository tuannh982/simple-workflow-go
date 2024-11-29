package activity

import "github.com/tuannh982/simple-workflow-go/pkg/dto/task"

type ActivityExecutionContext struct {
	Task *task.ActivityTask
}

func NewActivityExecutionContext(
	task *task.ActivityTask,
) *ActivityExecutionContext {
	return &ActivityExecutionContext{
		Task: task,
	}
}
