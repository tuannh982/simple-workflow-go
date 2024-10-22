package types

import "context"

type Activity[T any, R any] func(context.Context, *T) (*R, error)
