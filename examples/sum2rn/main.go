package main

import (
	"context"
	worker2 "github.com/tuannh982/simple-workflows-go/pkg/api/worker"
	"github.com/tuannh982/simple-workflows-go/test/e2e/psql"
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
	be, err := psql.InitBackend(logger)
	if err != nil {
		panic(err)
	}
	aw, err := worker2.NewActivityWorkersBuilder().WithBackend(be).WithLogger(logger).RegisterActivities(
		GenerateRandomNumberActivity1,
		GenerateRandomNumberActivity2,
	).Build()
	if err != nil {
		panic(err)
	}
	ww, err := worker2.NewWorkflowWorkersBuilder().WithBackend(be).WithLogger(logger).RegisterWorkflows(
		Sum2RandomNumberWorkflow,
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
