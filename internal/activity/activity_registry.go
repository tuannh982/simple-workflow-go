package activity

import (
	"fmt"
	"github.com/tuannh982/simple-workflows-go/internal/fn"
)

type ActivityRegistry struct {
	activities map[string]any
}

func NewActivityRegistry() *ActivityRegistry {
	return &ActivityRegistry{
		activities: make(map[string]any),
	}
}

func (a *ActivityRegistry) RegisterActivity(activity any) error {
	if err := fn.ValidateFn(activity); err != nil {
		panic(err)
	}
	name := fn.GetFunctionName(activity)
	if _, ok := a.activities[name]; ok {
		return fmt.Errorf("activity '%s' already registered", name)
	}
	a.activities[name] = activity
	return nil
}
