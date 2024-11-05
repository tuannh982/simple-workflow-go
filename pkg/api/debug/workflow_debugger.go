package debug

import (
	"context"
	"errors"
	"fmt"
	"github.com/tuannh982/simple-workflows-go/internal/fn"
	workflow2 "github.com/tuannh982/simple-workflows-go/internal/workflow"
	"github.com/tuannh982/simple-workflows-go/pkg/backend"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/history"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/task"
	"github.com/tuannh982/simple-workflows-go/pkg/registry"
)

type WorkflowDebugger interface {
	QueryUserDefinedVars(workflow any, workflowID string) (map[string]any, error)
}

type workflowDebugger struct {
	backend backend.Backend
}

func NewWorkflowDebugger(backend backend.Backend) WorkflowDebugger {
	return &workflowDebugger{
		backend: backend,
	}
}

func (d *workflowDebugger) QueryUserDefinedVars(workflow any, workflowID string) (map[string]any, error) {
	wr := registry.NewWorkflowRegistry()
	err := wr.RegisterWorkflow(workflow)
	if err != nil {
		return nil, err
	}
	workflowHistory, err := d.backend.GetWorkflowHistory(context.Background(), workflowID)
	if err != nil {
		return nil, err
	}
	if len(workflowHistory) == 0 {
		return nil, errors.New("workflow not exists")
	}
	if e := workflowHistory[0].WorkflowExecutionStarted; e != nil {
		name := fn.GetFunctionName(workflow)
		if name != e.Name {
			return nil, fmt.Errorf("workflow execution started for workflow '%s' has name '%s' instead of '%s'", workflowID, e.Name, name)
		}
		t := &task.WorkflowTask{
			WorkflowID: workflowID,
			OldEvents:  workflowHistory,
			NewEvents:  make([]*history.HistoryEvent, 0),
		}
		runtime := workflow2.NewWorkflowRuntime(wr, d.backend.DataConverter(), t)
		err = runtime.RunSimulation()
		if err != nil {
			return nil, err
		}
		return runtime.WorkflowExecutionContext.UserDefinedVars, nil
	} else {
		return nil, errors.New("workflow first history event is not WorkflowExecutionStarted")
	}
}
