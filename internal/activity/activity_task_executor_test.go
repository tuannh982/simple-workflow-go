package activity

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/tuannh982/simple-workflows-go/internal/dataconverter"
	"github.com/tuannh982/simple-workflows-go/internal/dto/history"
	"github.com/tuannh982/simple-workflows-go/internal/dto/task"
	"github.com/tuannh982/simple-workflows-go/internal/fn"
	"testing"
)

var dataConverter = dataconverter.NewJsonDataConverter()

type mockInput struct {
	Msg string
	Err string
}
type mockResult struct{ Echo string }

func mockActivity(_ context.Context, input *mockInput) (*mockResult, error) {
	return &mockResult{Echo: input.Msg}, errors.New(input.Err)
}

func TestActivityTaskExecutor(t *testing.T) {
	var err error
	registry := NewActivityRegistry()
	err = registry.RegisterActivity(mockActivity)
	assert.Nil(t, err)
	executor := NewActivityTaskExecutor(registry, dataConverter)
	//
	inputMsg := "hello world"
	inputErr := "bye world"
	input := &mockInput{
		Msg: inputMsg,
		Err: inputErr,
	}
	inputBytes, err := dataConverter.Marshal(input)
	assert.NoError(t, err)
	mockTask := &task.ActivityTask{
		WorkflowID: "mock workflow ID",
		TaskScheduleEvent: &history.ActivityScheduled{
			TaskScheduledID: 1,
			Name:            fn.GetFunctionName(mockActivity),
			Input:           inputBytes,
		},
	}
	// execute
	taskExecutionResult, err := executor.Execute(mockTask)
	assert.NoError(t, err)
	// extract result and error
	executionResult := taskExecutionResult.ExecutionResult
	resultBytes := executionResult.Result
	assert.NotNil(t, resultBytes)
	wrappedError := executionResult.Error
	assert.NotNil(t, wrappedError)
	// unmarshal
	result := &mockResult{}
	err = dataConverter.Unmarshal(*resultBytes, result)
	assert.NoError(t, err)
	assert.Equal(t, result.Echo, inputMsg)
	assert.Equal(t, wrappedError.Message, inputErr)
}
