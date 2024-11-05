package worker

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/tuannh982/simple-workflows-go/pkg/backend"
	"github.com/tuannh982/simple-workflows-go/pkg/registry"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/commons"
	"github.com/tuannh982/simple-workflows-go/pkg/worker/activity_worker"
	"go.uber.org/zap"
	"os"
)

type ActivityWorkersBuilder struct {
	logger             *zap.Logger
	name               string
	backend            backend.Backend
	activityWorkerOpts []func(options *activity_worker.ActivityWorkerOptions)
	activities         []any
}

func NewActivityWorkersBuilder() *ActivityWorkersBuilder {
	return &ActivityWorkersBuilder{
		activityWorkerOpts: make([]func(options *activity_worker.ActivityWorkerOptions), 0),
		activities:         make([]any, 0),
	}
}

func (b *ActivityWorkersBuilder) WithLogger(logger *zap.Logger) *ActivityWorkersBuilder {
	b.logger = logger
	return b
}

func (b *ActivityWorkersBuilder) WithName(name string) *ActivityWorkersBuilder {
	b.name = name
	return b
}

func (b *ActivityWorkersBuilder) WithBackend(backend backend.Backend) *ActivityWorkersBuilder {
	b.backend = backend
	return b
}

func (b *ActivityWorkersBuilder) WithActivityWorkerOpts(opts ...func(options *activity_worker.ActivityWorkerOptions)) *ActivityWorkersBuilder {
	b.activityWorkerOpts = append(b.activityWorkerOpts, opts...)
	return b
}

func (b *ActivityWorkersBuilder) RegisterActivities(activities ...any) *ActivityWorkersBuilder {
	b.activities = append(b.activities, activities...)
	return b
}

func (b *ActivityWorkersBuilder) Build() (*activity_worker.ActivityWorker, error) {
	ar := registry.NewActivityRegistry()
	err := ar.RegisterActivities(b.activities...)
	if err != nil {
		return nil, err
	}
	name := b.name
	if name == "" {
		host := commons.GetOrElse(func() (string, error) { return os.Hostname() }, "unknown host")
		id := uuid.New()
		name = fmt.Sprintf("[%s] ActivityWorker %s", host, id.String())
	}
	aw := activity_worker.NewActivityWorker(name, b.backend, ar, b.backend.DataConverter(), b.logger, b.activityWorkerOpts...)
	return aw, nil
}
