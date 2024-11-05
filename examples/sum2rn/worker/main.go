package main

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/examples/sum2rn"
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
	be, err := sum2rn.InitBackend(logger)
	if err != nil {
		panic(err)
	}
	aw, err := worker.NewActivityWorkersBuilder().
		WithBackend(be).
		WithLogger(logger).
		RegisterActivities(
			sum2rn.GenerateRandomNumberActivity1,
			sum2rn.GenerateRandomNumberActivity2,
		).Build()
	if err != nil {
		panic(err)
	}
	ww, err := worker.NewWorkflowWorkersBuilder().
		WithBackend(be).
		WithLogger(logger).
		RegisterWorkflows(
			sum2rn.Sum2RandomNumberWorkflow,
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
