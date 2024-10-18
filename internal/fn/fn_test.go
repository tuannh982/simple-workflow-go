package fn

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func fn1(_ any) (any, error) { return nil, nil }

type fn2Struct struct{}

func (t *fn2Struct) fn2(_ any) (any, error) { return nil, nil }

func TestGetFunctionName(t *testing.T) {
	tfn2 := &fn2Struct{}
	assert.Equal(t, GetFunctionName(fn1), "fn1")
	assert.Equal(t, GetFunctionName(tfn2.fn2), "fn2")
}

type input struct{}
type result struct{}

func fn3(_ context.Context, _ *input) (*result, error)  { return nil, nil }
func fn4(_ context.Context, _ input) (*result, error)   { return nil, nil }
func fn5(_ context.Context, _ input) (result, error)    { return result{}, nil }
func fn6(_ context.Context, _ any) (any, error)         { return result{}, nil }
func fn7(_ any) (any, error)                            { return nil, nil }
func fn8(_ context.Context, _ any) error                { return nil }
func fn9(_ context.Context, _ any) any                  { return nil }
func fn10(_ context.Context, _ any, _ any) (any, error) { return nil, nil }

func TestValidateFn(t *testing.T) {
	assert.NoError(t, ValidateFn(fn3))
	assert.Error(t, ValidateFn(fn4))
	assert.Error(t, ValidateFn(fn5))
	assert.Error(t, ValidateFn(fn6))
	assert.Error(t, ValidateFn(fn7))
	assert.Error(t, ValidateFn(fn8))
	assert.Error(t, ValidateFn(fn9))
	assert.Error(t, ValidateFn(fn10))
}

func TestInitArgument(t *testing.T) {
	ptr := InitArgument(fn3)
	_, ok := ptr.(*input)
	assert.True(t, ok)
	fmt.Println(ptr)
}
