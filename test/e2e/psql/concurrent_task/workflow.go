package concurrent_task

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/pkg/api/workflow"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/awaitable"
)

type SumWorkflowInput struct {
	NumberOfTasks int
}

type SumWorkflowOutput struct {
	Result int64
}

type GenerateNumberActivityInput struct {
	TaskIndex int
}

type GenerateNumberActivityResult struct {
	Result int64
}

func GenerateNumber(seed int64) int64 {
	var base int64 = 53
	var mod int64 = 1e9 + 7
	return ((seed % mod) * base) % mod
}

func GenerateNumberActivity(ctx context.Context, input *GenerateNumberActivityInput) (*GenerateNumberActivityResult, error) {
	return &GenerateNumberActivityResult{Result: GenerateNumber(int64(input.TaskIndex))}, nil
}

func SumWorkflow(ctx context.Context, input *SumWorkflowInput) (*SumWorkflowOutput, error) {
	tasks := make([]awaitable.Awaitable[*GenerateNumberActivityResult], input.NumberOfTasks)
	for i := 0; i < input.NumberOfTasks; i++ {
		tasks[i] = workflow.CallActivity(ctx, GenerateNumberActivity, &GenerateNumberActivityInput{TaskIndex: i})
	}
	var sum int64 = 0
	for i := 0; i < input.NumberOfTasks; i++ {
		r, err := tasks[i].Await()
		if err != nil {
			return nil, err
		}
		sum += r.Result
	}
	return &SumWorkflowOutput{Result: sum}, nil
}
