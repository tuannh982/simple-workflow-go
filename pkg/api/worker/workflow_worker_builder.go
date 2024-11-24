package worker

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/tuannh982/simple-workflow-go/pkg/backend"
	"github.com/tuannh982/simple-workflow-go/pkg/registry"
	"github.com/tuannh982/simple-workflow-go/pkg/utils/commons"
	"github.com/tuannh982/simple-workflow-go/pkg/worker/workflow_worker"
	"go.uber.org/zap"
	"os"
)

type WorkflowWorkersBuilder struct {
	logger             *zap.Logger
	name               string
	backend            backend.Backend
	workflowWorkerOpts []func(options *workflow_worker.WorkflowWorkerOptions)
	workflows          []any
}

func NewWorkflowWorkersBuilder() *WorkflowWorkersBuilder {
	return &WorkflowWorkersBuilder{
		workflowWorkerOpts: make([]func(options *workflow_worker.WorkflowWorkerOptions), 0),
		workflows:          make([]any, 0),
	}
}

func (b *WorkflowWorkersBuilder) WithLogger(logger *zap.Logger) *WorkflowWorkersBuilder {
	b.logger = logger
	return b
}

func (b *WorkflowWorkersBuilder) WithName(name string) *WorkflowWorkersBuilder {
	b.name = name
	return b
}

func (b *WorkflowWorkersBuilder) WithBackend(backend backend.Backend) *WorkflowWorkersBuilder {
	b.backend = backend
	return b
}

func (b *WorkflowWorkersBuilder) WithWorkflowWorkerOpts(opts ...func(options *workflow_worker.WorkflowWorkerOptions)) *WorkflowWorkersBuilder {
	b.workflowWorkerOpts = append(b.workflowWorkerOpts, opts...)
	return b
}

func (b *WorkflowWorkersBuilder) RegisterWorkflows(workflows ...any) *WorkflowWorkersBuilder {
	b.workflows = append(b.workflows, workflows...)
	return b
}

func (b *WorkflowWorkersBuilder) Build() (*workflow_worker.WorkflowWorker, error) {
	wr := registry.NewWorkflowRegistry()
	err := wr.RegisterWorkflows(b.workflows...)
	if err != nil {
		return nil, err
	}
	name := b.name
	if name == "" {
		host := commons.GetOrElse(func() (string, error) { return os.Hostname() }, "unknown host")
		id := uuid.New()
		name = fmt.Sprintf("[%s] WorkflowWorker %s", host, id.String())
	}
	ww := workflow_worker.NewWorkflowWorker(name, b.backend, wr, b.backend.DataConverter(), b.logger, b.workflowWorkerOpts...)
	return ww, nil
}
