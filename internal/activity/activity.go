package activity

import "context"

type Activity func(context.Context, any) (any, error)
