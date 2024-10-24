package activity

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/tuannh982/simple-workflows-go/internal/fn"
	"github.com/tuannh982/simple-workflows-go/pkg/dataconverter"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/history"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/task"
	"github.com/tuannh982/simple-workflows-go/pkg/registry"
	"go.uber.org/zap"
	"testing"
)

var dataConverter = dataconverter.NewJsonDataConverter()

type mockInput struct {
	Msg string
	Err string
}
type mockResult struct{ Echo string }

func mockActivity1(_ context.Context, input *mockInput) (*mockResult, error) {
	return &mockResult{Echo: input.Msg}, errors.New(input.Err)
}

func mockActivity2(_ context.Context, input *mockInput) (*mockResult, error) {
	panic(input.Err)
}

func mockActivity3(_ context.Context, input *mockInput) (*mockResult, error) {
	return &mockResult{Echo: input.Msg}, nil
}

func mockActivity4(_ context.Context, input *mockInput) (*mockResult, error) {
	return nil, errors.New(input.Err)
}

func callActivity(
	executor ActivityTaskExecutor,
	activity any,
	inputBytes []byte,
) (*task.ActivityTaskResult, error) {
	mockTask := &task.ActivityTask{
		WorkflowID: "mock workflow ID",
		TaskScheduleEvent: &history.ActivityScheduled{
			TaskScheduledID: 1,
			Name:            fn.GetFunctionName(activity),
			Input:           inputBytes,
		},
	}
	// execute
	taskResult, err := executor.Execute(context.TODO(), mockTask)
	return taskResult, err
}

func initExecutor(t *testing.T) ActivityTaskExecutor {
	var err error
	r := registry.NewActivityRegistry()
	err = r.RegisterActivities(
		mockActivity1,
		mockActivity2,
		mockActivity3,
		mockActivity4,
	)
	assert.Nil(t, err)
	executor := NewActivityTaskExecutor(r, dataConverter, zap.NewNop())
	return executor
}

func TestActivityTaskExecutor1(t *testing.T) {
	executor := initExecutor(t)
	//
	inputMsg := "hello world"
	inputErr := "bye world"
	input := &mockInput{
		Msg: inputMsg,
		Err: inputErr,
	}
	inputBytes, err := dataConverter.Marshal(input)
	assert.NoError(t, err)
	// test activity 1
	activityTaskResult, err := callActivity(executor, mockActivity1, inputBytes)
	assert.NoError(t, err)
	executionResult := activityTaskResult.ExecutionResult
	resultBytes := executionResult.Result
	assert.NotNil(t, resultBytes)
	wrappedError := executionResult.Error
	assert.NotNil(t, wrappedError)
	// unmarshal
	result := &mockResult{}
	err = dataConverter.Unmarshal(*resultBytes, result)
	assert.NoError(t, err)
	assert.Equal(t, inputMsg, result.Echo)
	assert.Equal(t, inputErr, wrappedError.Message)
}

func TestActivityTaskExecutor2(t *testing.T) {
	executor := initExecutor(t)
	//
	inputMsg := "hello world"
	inputErr := "bye world"
	input := &mockInput{
		Msg: inputMsg,
		Err: inputErr,
	}
	inputBytes, err := dataConverter.Marshal(input)
	assert.NoError(t, err)
	// test activity 2
	_, err = callActivity(executor, mockActivity2, inputBytes)
	assert.Error(t, err)
}

func TestActivityTaskExecutor3(t *testing.T) {
	executor := initExecutor(t)
	//
	inputMsg := "hello world"
	inputErr := "bye world"
	input := &mockInput{
		Msg: inputMsg,
		Err: inputErr,
	}
	inputBytes, err := dataConverter.Marshal(input)
	assert.NoError(t, err)
	// test activity 1
	activityTaskResult, err := callActivity(executor, mockActivity3, inputBytes)
	assert.NoError(t, err)
	executionResult := activityTaskResult.ExecutionResult
	resultBytes := executionResult.Result
	assert.NotNil(t, resultBytes)
	wrappedError := executionResult.Error
	assert.Nil(t, wrappedError)
	// unmarshal
	result := &mockResult{}
	err = dataConverter.Unmarshal(*resultBytes, result)
	assert.NoError(t, err)
	assert.Equal(t, inputMsg, result.Echo)
}

func TestActivityTaskExecutor4(t *testing.T) {
	executor := initExecutor(t)
	//
	inputMsg := "hello world"
	inputErr := "bye world"
	input := &mockInput{
		Msg: inputMsg,
		Err: inputErr,
	}
	inputBytes, err := dataConverter.Marshal(input)
	assert.NoError(t, err)
	// test activity 1
	activityTaskResult, err := callActivity(executor, mockActivity4, inputBytes)
	assert.NoError(t, err)
	executionResult := activityTaskResult.ExecutionResult
	resultBytes := executionResult.Result
	assert.Nil(t, resultBytes)
	wrappedError := executionResult.Error
	assert.NotNil(t, wrappedError)
	assert.Equal(t, inputErr, wrappedError.Message)
}
