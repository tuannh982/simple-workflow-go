package workflow

import (
	"context"
	"errors"
	"github.com/tuannh982/simple-workflows-go/pkg/dataconverter"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/history"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/task"
	"github.com/tuannh982/simple-workflows-go/pkg/registry"
	"go.uber.org/zap"
)

type WorkflowTaskExecutor interface {
	Execute(ctx context.Context, task *task.WorkflowTask) (*task.WorkflowTaskResult, error)
}

type workflowTaskExecutor struct {
	workflowRegistry *registry.WorkflowRegistry
	dataConverter    dataconverter.DataConverter
	logger           *zap.Logger
}

func NewWorkflowTaskExecutor(
	workflowRegistry *registry.WorkflowRegistry,
	dataConverter dataconverter.DataConverter,
	logger *zap.Logger,
) WorkflowTaskExecutor {
	return &workflowTaskExecutor{
		workflowRegistry: workflowRegistry,
		dataConverter:    dataConverter,
		logger:           logger,
	}
}

func (w *workflowTaskExecutor) augmentWorkflowTaskEvents(t *task.WorkflowTask) []*history.HistoryEvent {
	taskStartedEvent := &history.HistoryEvent{
		Timestamp:           t.FetchTimestamp,
		WorkflowTaskStarted: &history.WorkflowTaskStarted{},
	}
	l := len(t.NewEvents)
	augmentedNewEvents := make([]*history.HistoryEvent, 0, l+1)
	if len(t.NewEvents) == 0 {
		return []*history.HistoryEvent{taskStartedEvent}
	} else {
		firstEvent := t.NewEvents[0]
		if firstEvent.WorkflowExecutionStarted == nil {
			augmentedNewEvents = append(augmentedNewEvents, taskStartedEvent)
		}
		for i := 0; i < l; i++ {
			augmentedNewEvents = append(augmentedNewEvents, t.NewEvents[i])
		}
	}
	return augmentedNewEvents
}

func (w *workflowTaskExecutor) Execute(_ context.Context, t *task.WorkflowTask) (*task.WorkflowTaskResult, error) {
	if len(t.NewEvents) == 0 {
		return nil, errors.New("no new events, nothing to do")
	}
	t.NewEvents = w.augmentWorkflowTaskEvents(t)
	//
	runtime := NewWorkflowRuntime(w.workflowRegistry, w.dataConverter, t)
	err := runtime.RunSimulation()
	if err != nil {
		return nil, err
	} else {
		return runtime.GetWorkflowTaskResult(), nil
	}
}
