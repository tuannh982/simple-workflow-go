package main

import (
	"context"
	"github.com/tuannh982/simple-workflow-go/examples"
	"github.com/tuannh982/simple-workflow-go/examples/subscription_with_debug"
	"github.com/tuannh982/simple-workflow-go/pkg/api/worker"
	"github.com/tuannh982/simple-workflow-go/pkg/worker/activity_worker"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx := context.Background()
	logger, err := examples.GetLogger()
	if err != nil {
		panic(err)
	}
	be, err := examples.InitPSQLBackend(logger)
	if err != nil {
		panic(err)
	}
	aw, err := worker.NewActivityWorkersBuilder().
		WithName("demo activity worker").
		WithBackend(be).
		WithLogger(logger).
		RegisterActivities(
			subscription_with_debug.PaymentActivity,
		).
		WithActivityWorkerOpts(
			activity_worker.WithTaskProcessorMaxBackoffInterval(1 * time.Minute),
		).
		Build()
	if err != nil {
		panic(err)
	}
	ww, err := worker.NewWorkflowWorkersBuilder().
		WithName("demo workflow worker").
		WithBackend(be).
		WithLogger(logger).
		RegisterWorkflows(
			subscription_with_debug.SubscriptionWorkflow,
		).Build()
	if err != nil {
		panic(err)
	}
	aw.Start(ctx)
	defer aw.Stop(ctx)
	ww.Start(ctx)
	defer ww.Stop(ctx)
	//
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
