package activity

import (
	"context"
	"fmt"
	"github.com/tuannh982/simple-workflows-go/internal/fn"
	"github.com/tuannh982/simple-workflows-go/pkg/dataconverter"
	"github.com/tuannh982/simple-workflows-go/pkg/dto"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/task"
	"github.com/tuannh982/simple-workflows-go/pkg/registry"
	"go.uber.org/zap"
	"runtime/debug"
)

type ActivityTaskExecutor interface {
	Execute(ctx context.Context, task *task.ActivityTask) (*task.ActivityTaskResult, error)
}

type activityTaskExecutor struct {
	activityRegistry *registry.ActivityRegistry
	dataConverter    dataconverter.DataConverter
	logger           *zap.Logger
}

func NewActivityTaskExecutor(
	activityRegistry *registry.ActivityRegistry,
	dataConverter dataconverter.DataConverter,
	logger *zap.Logger,
) ActivityTaskExecutor {
	return &activityTaskExecutor{
		activityRegistry: activityRegistry,
		dataConverter:    dataConverter,
		logger:           logger,
	}
}

func (a *activityTaskExecutor) executeActivity(
	activity any,
	ctx context.Context,
	input any,
) (executionResult *dto.ExecutionResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v\nstack: %s", r, string(debug.Stack()))
		}
	}()
	callResult, callErr := fn.CallFn(activity, ctx, input)
	marshaledCallResult, err := dto.ExtractResultFromFnCallResult(callResult, a.dataConverter.Marshal)
	if err != nil {
		return nil, err
	}
	wrappedCallError := dto.ExtractErrorFromFnCallError(callErr)
	return &dto.ExecutionResult{
		Result: marshaledCallResult,
		Error:  wrappedCallError,
	}, nil
}

func (a *activityTaskExecutor) Execute(_ context.Context, t *task.ActivityTask) (*task.ActivityTaskResult, error) {
	name := t.TaskScheduleEvent.Name
	inputBytes := t.TaskScheduleEvent.Input
	if activity, ok := a.activityRegistry.Activities[name]; ok {
		callCtx := InjectActivityExecutionContext(context.Background(), NewActivityExecutionContext())
		input := fn.InitArgument(activity)
		err := a.dataConverter.Unmarshal(inputBytes, input)
		if err != nil {
			return nil, err
		}
		executionResult, err := a.executeActivity(activity, callCtx, input)
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
