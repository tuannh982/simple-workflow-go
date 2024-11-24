package main

import (
	"context"
	"flag"
	"github.com/tuannh982/simple-workflow-go/examples"
	"github.com/tuannh982/simple-workflow-go/examples/subscription_with_debug"
	"github.com/tuannh982/simple-workflow-go/pkg/api/client"
	"github.com/tuannh982/simple-workflow-go/pkg/api/debug"
	"go.uber.org/zap"
	"time"
)

var (
	workflowID    string
	mode          string
	totalAmount   int64
	cycles        int
	cycleDuration time.Duration
)

func init() {
	flag.StringVar(&workflowID, "workflowID", "default-workflow-id", "workflowID")
	flag.StringVar(&mode, "mode", "schedule", "schedule|await|debug")
	flag.Int64Var(&totalAmount, "totalAmount", 1000, "totalAmount")
	flag.IntVar(&cycles, "cycles", 10, "cycles")
	flag.DurationVar(&cycleDuration, "cycleDuration", time.Second*10, "cycleDuration")
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
		workflowResult, workflowErr, err := client.AwaitWorkflowResult(ctx, be, subscription_with_debug.SubscriptionWorkflow, workflowID)
		if err != nil {
			panic(err)
		}
		examples.PrettyPrint(workflowResult)
		examples.PrettyPrint(workflowErr)
	} else if mode == "schedule" {
		err = client.ScheduleWorkflow(ctx, be, subscription_with_debug.SubscriptionWorkflow, &subscription_with_debug.SubscriptionWorkflowInput{
			TotalAmount:   totalAmount,
			Cycles:        cycles,
			CycleDuration: cycleDuration,
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
