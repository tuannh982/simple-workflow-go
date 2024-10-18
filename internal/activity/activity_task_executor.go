package activity

import (
	"context"
	"fmt"
	"github.com/tuannh982/simple-workflows-go/internal/dataconverter"
	"github.com/tuannh982/simple-workflows-go/internal/dto"
	"github.com/tuannh982/simple-workflows-go/internal/dto/task"
	"github.com/tuannh982/simple-workflows-go/internal/fn"
)

type ActivityTaskExecutor interface {
	Execute(task *task.ActivityTask) (*task.ActivityTaskResult, error)
}

type activityTaskExecutor struct {
	ActivityRegistry *ActivityRegistry
	DataConverter    dataconverter.DataConverter
}

func NewActivityTaskExecutor(
	activityRegistry *ActivityRegistry,
	dataConverter dataconverter.DataConverter,
) ActivityTaskExecutor {
	return &activityTaskExecutor{
		ActivityRegistry: activityRegistry,
		DataConverter:    dataConverter,
	}
}

func (a *activityTaskExecutor) executeActivity(activity Activity, ctx context.Context, input any) (*dto.ExecutionResult, error) {
	var err error
	activityResult, activityErr := activity(ctx, input)
	var marshaledActivityResult *[]byte
	var wrappedActivityError *dto.Error
	if activityErr != nil {
		wrappedActivityError = &dto.Error{
			Message: activityErr.Error(),
		}
	}
	if activityResult != nil {
		var r []byte
		r, err = a.DataConverter.Marshal(activityResult)
		marshaledActivityResult = &r
		if err != nil {
			return nil, err
		}
	}
	return &dto.ExecutionResult{
		Result: marshaledActivityResult,
		Error:  wrappedActivityError,
	}, nil
}

func (a *activityTaskExecutor) Execute(t *task.ActivityTask) (*task.ActivityTaskResult, error) {
	name := t.TaskScheduleEvent.Name
	inputBytes := t.TaskScheduleEvent.Input
	if activity, ok := a.ActivityRegistry.activities[name]; ok {
		ctx := InjectActivityExecutionContext(context.Background(), NewActivityExecutionContext())
		input := fn.InitArgument(activity)
		err := a.DataConverter.Unmarshal(inputBytes, input)
		if err != nil {
			return nil, err
		}
		executionResult, err := a.executeActivity(activity, ctx, input)
		if err != nil {
			return nil, err
		}
		return &task.ActivityTaskResult{
			Task:            t,
			ExecutionResult: executionResult,
		}, nil
	} else {
		return nil, fmt.Errorf("activity %s not found", name)
	}
}
