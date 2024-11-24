package registry

import (
	"fmt"
	"github.com/tuannh982/simple-workflow-go/internal/fn"
)

type WorkflowRegistry struct {
	Workflows map[string]any
}

func NewWorkflowRegistry() *WorkflowRegistry {
	return &WorkflowRegistry{
		Workflows: make(map[string]any),
	}
}

func (a *WorkflowRegistry) RegisterWorkflow(workflow any) error {
	if err := fn.ValidateFn(workflow); err != nil {
		panic(err)
	}
	name := fn.GetFunctionName(workflow)
	if _, ok := a.Workflows[name]; ok {
		return fmt.Errorf("workflow '%s' already registered", name)
	}
	a.Workflows[name] = workflow
	return nil
}

func (a *WorkflowRegistry) RegisterWorkflows(workflows ...any) error {
	for _, workflow := range workflows {
		if err := a.RegisterWorkflow(workflow); err != nil {
			return err
		}
	}
	return nil
}
