package workflow

import (
	"context"
	"errors"
	"github.com/tuannh982/simple-workflows-go/pkg/dataconverter"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/history"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/task"
	"github.com/tuannh982/simple-workflows-go/pkg/registry"
)

type WorkflowTaskExecutor interface {
	Execute(ctx context.Context, task *task.WorkflowTask) (*task.WorkflowTaskResult, error)
}

type workflowTaskExecutor struct {
	WorkflowRegistry *registry.WorkflowRegistry
	DataConverter    dataconverter.DataConverter
}

func NewWorkflowTaskExecutor(
	workflowRegistry *registry.WorkflowRegistry,
	dataConverter dataconverter.DataConverter,
) WorkflowTaskExecutor {
	return &workflowTaskExecutor{
		WorkflowRegistry: workflowRegistry,
		DataConverter:    dataConverter,
	}
}

func (w *workflowTaskExecutor) augmentWorkflowTaskEvents(t *task.WorkflowTask) []*history.HistoryEvent {
	l := len(t.NewEvents)
	augmentedNewEvents := make([]*history.HistoryEvent, 0, l+1)
	if t.NewEvents[0].WorkflowTaskStarted == nil {
		augmentedNewEvents = append(augmentedNewEvents, &history.HistoryEvent{
			Timestamp:           t.FetchTimestamp,
			WorkflowTaskStarted: &history.WorkflowTaskStarted{},
		})
	}
	for i := 0; i < l; i++ {
		augmentedNewEvents = append(augmentedNewEvents, t.NewEvents[i])
	}
	return augmentedNewEvents
}

func (w *workflowTaskExecutor) Execute(_ context.Context, t *task.WorkflowTask) (*task.WorkflowTaskResult, error) {
	if len(t.NewEvents) == 0 {
		return nil, errors.New("no new events, nothing to do")
	}
	t.NewEvents = w.augmentWorkflowTaskEvents(t)
	//
	runtime := NewWorkflowRuntime(w.WorkflowRegistry, w.DataConverter, t)
	err := runtime.RunSimulation()
	if err != nil {
		return nil, err
	} else {
		return runtime.GetWorkflowTaskResult(), nil
	}
}
