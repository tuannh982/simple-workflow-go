package activity

import (
	"github.com/tuannh982/simple-workflow-go/pkg/dto/task"
	"time"
)

type ActivityExecutionContext struct {
	Task                         *task.ActivityTask
	UserDefinedNextExecutionTime *time.Time
	UserDefinedBackoffDuration   *time.Duration
}

func NewActivityExecutionContext(
	task *task.ActivityTask,
) *ActivityExecutionContext {
	return &ActivityExecutionContext{
		Task: task,
	}
}

func (ctx *ActivityExecutionContext) NextExecutionTime() *time.Time {
	if ctx.UserDefinedNextExecutionTime != nil {
		return ctx.UserDefinedNextExecutionTime
	} else if ctx.UserDefinedBackoffDuration != nil {
		t := time.Now()
		t = t.Add(*ctx.UserDefinedBackoffDuration)
		return &t
	} else {
		return nil
	}
}
