simple-workflow-go
===

A simple workflow framework written in Go

## Quickstarts

### Backend
Backend is responsible for persisting workflow state, including tasks, events, workflow runtime metadata.

In this example, we will use PSQL as backend.

First, start PSQL server locally

```shell
docker compose -f docker/docker-compose-psql.yaml up -d
```

Then, init a backend instance that connect to PSQL server

```go
const (
    DbHost     = "localhost"
    DbPort     = 5432
    DbName     = "postgres"
    DbUser     = "user"
    DbPassword = "123456"
)

func InitBackend(logger *zap.Logger) (backend.Backend, error) {
    hostname, err := os.Hostname()
    if err != nil {
        return nil, err
    }
    db, err := psql.Connect(DbHost, DbPort, DbUser, DbPassword, DbName, nil)
    if err != nil {
        return nil, err
    }
    err = psql.PrepareDB(db) // auto-create table if not exists
    if err != nil {
        return nil, err
    }
    err = psql.TruncateDB(db) // truncate DB data
    if err != nil {
        return nil, err
    }
    dataConverter := dataconverter.NewJsonDataConverter()
    be := psql.NewPSQLBackend(hostname, dataConverter, db, logger)
    return be, nil
}
```

```go
be, err := psql.InitBackend(logger)
```

### Activities
Activity is a function that used to implement service calls, I/O operation, 
long running operations, or costly actions which are not prefer to be re-executed

```go
func GenerateNumber(seed int64, round int) int64 {
    for _ = range round {
        seed ^= seed << 13
        seed ^= seed << 17
        seed ^= seed << 5
    }
    return seed
}

func GenerateRandomNumberActivity1(ctx context.Context, input *Seed) (*Int64, error) {
    return &Int64{Value: GenerateNumber(input.Value, 19)}, nil
}

func GenerateRandomNumberActivity2(ctx context.Context, input *Seed) (*Int64, error) {
    return &Int64{Value: GenerateNumber(input.Value, 23)}, nil
}
```

**note #1**: code inside activity must be non-deterministic, which is when running the activity twice, 
it always yields the same result 

**note #2**: if you experience an unexpected error while executing activity, 
just call `panic(...)`, the activity will be retried later

### Workflow

Workflow is the orchestration of activities

```go
func Sum2RandomNumberWorkflow(ctx context.Context, input *Seed) (*Int64, error) {
    activity1Result, err := workflow.CallActivity(ctx, GenerateRandomNumberActivity1, input).Await()
    if err != nil {
        return nil, err
    }
    activity2Result, err := workflow.CallActivity(ctx, GenerateRandomNumberActivity2, input).Await()
    if err != nil {
        return nil, err
    }
    sum := activity1Result.Value + activity2Result.Value
    return &Int64{Value: sum}, nil
}
```

**note #1**: DO NOT put any costly operations (IO operations, external service calls, etc.) on workflow code, 
put them in activity code instead

### Workers

Workers, including `ActivityWorker` and `WorkflowWorker` are responsible for executing activity and workflow codes

#### ActivityWorker
```go
aw, err := worker2.NewActivityWorkersBuilder().
    WithName("[e2e test] ActivityWorker").
    WithBackend(be).
    WithLogger(logger).
    RegisterActivities(
        GenerateRandomNumberActivity1,
        GenerateRandomNumberActivity2,
    ).
    Build()
```

#### WorkflowWorker

```go
ww, err := worker2.NewWorkflowWorkersBuilder().
    WithName("[e2e test] WorkflowWorker").
    WithBackend(be).
    WithLogger(logger).
    RegisterWorkflows(
        Sum2RandomNumberWorkflow,
    ).
    Build()
```

### Putting all together

Putting all pieces together, we can implement our worker program

```go
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
```

### Starting a workflow 

After having our worker instance running, we can write codes to start workflows and wait for their result

#### Start workflow

```go
err := client.ScheduleWorkflow(ctx, be, Sum2RandomNumberWorkflow, &Seed{
    Value: seed,
}, client.WorkflowScheduleOptions{
    WorkflowID: workflowID,
    Version:    "1",
})
```

#### Await workflow result

```go
workflowResult, workflowErr, err := client.AwaitWorkflowResult(
    ctx, 
    be, 
    Sum2RandomNumberWorkflow, 
    workflowID, 
)
```

## Build & Test

Run `build.sh` to perform build, unit test, integration test and e2e test