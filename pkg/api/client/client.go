package client

import (
	"context"
	"errors"
	"github.com/tuannh982/simple-workflows-go/internal/fn"
	"github.com/tuannh982/simple-workflows-go/pkg/backend"
	"github.com/tuannh982/simple-workflows-go/pkg/dto"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/history"
	"github.com/tuannh982/simple-workflows-go/pkg/types"
	"time"
)

type WorkflowScheduleOptions struct {
	WorkflowID string
	Version    string
}

func ScheduleWorkflow[T any, R any](
	ctx context.Context,
	backend backend.Backend,
	workflow types.Workflow[T, R],
	input *T,
	options WorkflowScheduleOptions,
) error {
	name := fn.GetFunctionName(workflow)
	inputBytes, err := backend.DataConverter().Marshal(input)
	if err != nil {
		panic(err)
	}
	executionStarted := &history.WorkflowExecutionStarted{
		Name:       name,
		Version:    options.Version,
		Input:      inputBytes,
		WorkflowID: options.WorkflowID,
	}
	return backend.CreateWorkflow(ctx, executionStarted)
}

func GetWorkflowResult[T any, R any](
	ctx context.Context,
	backend backend.Backend,
	workflow types.Workflow[T, R],
	workflowID string,
) (*dto.WorkflowExecutionResult, error) {
	name := fn.GetFunctionName(workflow)
	return backend.GetWorkflowResult(ctx, name, workflowID)
}

func AwaitWorkflowResult[T any, R any](
	ctx context.Context,
	backend backend.Backend,
	workflow types.Workflow[T, R],
	workflowID string,
) (*R, error, error) {
	name := fn.GetFunctionName(workflow)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-ticker.C:
			result, err := backend.GetWorkflowResult(ctx, name, workflowID)
			if err != nil {
				return nil, nil, err
			}
			if result.RuntimeStatus == string(dto.WorkflowRuntimeStatusCompleted) {
				var wResult *R
				var wError error
				if result.Result != nil {
					ptr := fn.InitResult(workflow)
					err = backend.DataConverter().Unmarshal(*result.Result, ptr)
					if err != nil {
						return nil, nil, err
					}
					wResult = ptr.(*R)
				}
				if result.Error != nil {
					wError = errors.New(result.Error.Message)
				}
				return wResult, wError, nil
			}
		}
	}
}
