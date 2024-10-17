package activity

import (
	"fmt"
	"github.com/tuannh982/simple-workflows-go/internal/fn"
)

type WorkflowRegistry struct {
	workflows map[string]Workflow
}

func NewWorkflowRegistry() *WorkflowRegistry {
	return &WorkflowRegistry{
		workflows: make(map[string]Workflow),
	}
}

func (a *WorkflowRegistry) RegisterActivity(workflow Workflow) error {
	if err := fn.ValidateFn(workflow); err != nil {
		panic(err)
	}
	name := fn.GetFunctionName(workflow)
	if _, ok := a.workflows[name]; ok {
		return fmt.Errorf("workflow '%s' already registered", name)
	}
	a.workflows[name] = workflow
	return nil
}
