package activity

import "context"

type Workflow func(context.Context, any) (any, error)
