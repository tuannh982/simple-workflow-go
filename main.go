package main

import (
	"context"
	"fmt"
	"github.com/tuannh982/simple-workflows-go/internal/activity"
	"github.com/tuannh982/simple-workflows-go/internal/dataconverter"
	"github.com/tuannh982/simple-workflows-go/internal/workflow"
	"github.com/tuannh982/simple-workflows-go/pkg/api"
)

type HelloActivityInput struct {
	From string
}

type HelloActivityResult struct {
	Response string
}

func HelloActivity(ctx context.Context, input *HelloActivityInput) (*HelloActivityResult, error) {
	fmt.Printf("received message from %s", input.From)
	responseMsg := fmt.Sprintf("Hello, %s!", input.From)
	return &HelloActivityResult{responseMsg}, nil
}

type HelloWorkflowInput struct {
	From string
}

type HelloWorkflowResult struct {
	Response string
}

func HelloWorkflow(ctx context.Context, input *HelloWorkflowInput) (*HelloWorkflowResult, error) {
	result, err := api.CallActivity(ctx, HelloActivity, &HelloActivityInput{From: input.From}).Await()
	if err != nil {
		panic(err)
	}
	return &HelloWorkflowResult{Response: result.Response}, nil
}

func main() {
	dataConverter := dataconverter.NewJsonDataConverter()
	activityRegistry := activity.NewActivityRegistry()
	_ = activityRegistry.RegisterActivity(HelloActivity)
	workflowRegistry := workflow.NewWorkflowRegistry()
	_ = workflowRegistry.RegisterWorkflow(HelloWorkflow)
	_ = dataConverter
	// TODO wip
}
