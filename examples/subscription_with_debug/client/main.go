package main

import (
	"context"
	"flag"
	"github.com/tuannh982/simple-workflows-go/examples"
	"github.com/tuannh982/simple-workflows-go/examples/subscription_with_debug"
	"github.com/tuannh982/simple-workflows-go/pkg/api/client"
	"github.com/tuannh982/simple-workflows-go/pkg/api/debug"
	"go.uber.org/zap"
)

var (
	workflowID  string
	mode        string
	totalAmount int64
	cycles      int
)

func init() {
	flag.StringVar(&workflowID, "workflowID", "default-workflow-id", "workflowID")
	flag.StringVar(&mode, "mode", "schedule", "schedule|await|debug")
	flag.Int64Var(&totalAmount, "totalAmount", 1000, "totalAmount")
	flag.IntVar(&cycles, "cycles", 10, "cycles")
	flag.Parse()
}

func main() {
	ctx := context.Background()
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	be, err := subscription_with_debug.InitBackend(logger)
	if err != nil {
		panic(err)
	}
	if mode == "await" {
		workflowResult, workflowErr, err := client.AwaitWorkflowResult(ctx, be, subscription_with_debug.SubscriptionWorkflow, workflowID)
		if err != nil {
			panic(err)
		}
		examples.PrettyPrint(workflowResult)
		examples.PrettyPrint(workflowErr)
	} else if mode == "schedule" {
		err = client.ScheduleWorkflow(ctx, be, subscription_with_debug.SubscriptionWorkflow, &subscription_with_debug.SubscriptionWorkflowInput{
			TotalAmount: totalAmount,
			Cycles:      cycles,
		}, client.WorkflowScheduleOptions{
			WorkflowID: workflowID,
			Version:    "1",
		})
		if err != nil {
			panic(err)
		}
	} else if mode == "debug" {
		dbg := debug.NewWorkflowDebugger(be)
		vars, err := dbg.QueryUserDefinedVars(subscription_with_debug.SubscriptionWorkflow, workflowID)
		if err != nil {
			panic(err)
		}
		examples.PrettyPrint(vars)
	}
}
