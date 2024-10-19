package activity

import (
	"fmt"
	"github.com/tuannh982/simple-workflows-go/internal/fn"
)

type ActivityRegistry struct {
	Activities map[string]any
}

func NewActivityRegistry() *ActivityRegistry {
	return &ActivityRegistry{
		Activities: make(map[string]any),
	}
}

func (a *ActivityRegistry) RegisterActivity(activity any) error {
	if err := fn.ValidateFn(activity); err != nil {
		panic(err)
	}
	name := fn.GetFunctionName(activity)
	if _, ok := a.Activities[name]; ok {
		return fmt.Errorf("activity '%s' already registered", name)
	}
	a.Activities[name] = activity
	return nil
}
