package main

import (
	"context"
	"flag"
	"github.com/tuannh982/simple-workflows-go/examples"
	"github.com/tuannh982/simple-workflows-go/examples/sum2rn"
	"github.com/tuannh982/simple-workflows-go/pkg/api/client"
	"go.uber.org/zap"
)

var (
	workflowID string
	mode       string
	seed       int64
)

func init() {
	flag.StringVar(&workflowID, "workflowID", "default-workflow-id", "workflowID")
	flag.StringVar(&mode, "mode", "schedule", "schedule|await")
	flag.Int64Var(&seed, "seed", 100, "seed")
	flag.Parse()
}

func main() {
	ctx := context.Background()
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	be, err := examples.InitPSQLBackend(logger)
	if err != nil {
		panic(err)
	}
	if mode == "await" {
		workflowResult, workflowErr, err := client.AwaitWorkflowResult(ctx, be, sum2rn.Sum2RandomNumberWorkflow, workflowID)
		if err != nil {
			panic(err)
		}
		examples.PrettyPrint(workflowResult)
		examples.PrettyPrint(workflowErr)
	} else if mode == "schedule" {
		err = client.ScheduleWorkflow(ctx, be, sum2rn.Sum2RandomNumberWorkflow, &sum2rn.Seed{
			Value: seed,
		}, client.WorkflowScheduleOptions{
			WorkflowID: workflowID,
			Version:    "1",
		})
		if err != nil {
			panic(err)
		}
	}
}
