package main

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/examples/subscription_with_debug"
	"github.com/tuannh982/simple-workflows-go/pkg/api/worker"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

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
	aw, err := worker.NewActivityWorkersBuilder().
		WithBackend(be).
		WithLogger(logger).
		RegisterActivities(
			subscription_with_debug.PaymentActivity,
		).Build()
	if err != nil {
		panic(err)
	}
	ww, err := worker.NewWorkflowWorkersBuilder().
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
