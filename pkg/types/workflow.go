package types

import "context"

type Workflow[T any, R any] func(context.Context, *T) (*R, error)
