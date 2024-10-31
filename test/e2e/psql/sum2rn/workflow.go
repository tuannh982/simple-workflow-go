package concurrent_task

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/pkg/api/workflow"
)

type Seed struct {
	Value int64
}

type Int64 struct {
	Value int64
}

func GenerateNumber(seed int64, round int) int64 {
	for _ = range round {
		seed ^= seed << 13
		seed ^= seed << 17
		seed ^= seed << 5
	}
	return seed
}

func GenerateRandomNumberActivity1(ctx context.Context, input *Seed) (*Int64, error) {
	return &Int64{Value: GenerateNumber(input.Value, 19)}, nil
}

func GenerateRandomNumberActivity2(ctx context.Context, input *Seed) (*Int64, error) {
	return &Int64{Value: GenerateNumber(input.Value, 23)}, nil
}

func Sum2RandomNumberWorkflow(ctx context.Context, input *Seed) (*Int64, error) {
	activity1Result, err := workflow.CallActivity(ctx, GenerateRandomNumberActivity1, input).Await()
	if err != nil {
		return nil, err
	}
	activity2Result, err := workflow.CallActivity(ctx, GenerateRandomNumberActivity2, input).Await()
	if err != nil {
		return nil, err
	}
	sum := activity1Result.Value + activity2Result.Value
	return &Int64{Value: sum}, nil
}
