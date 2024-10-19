package fn

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()
var contextInterface = reflect.TypeOf((*context.Context)(nil)).Elem()

func GetFunctionName(f any) string {
	name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	startIndex := strings.LastIndexByte(name, '.')
	if startIndex > 0 {
		name = name[startIndex+1:]
	}
	// Compiler adds -fm suffix to a function name which has a receiver
	return strings.TrimSuffix(name, "-fm")
}

func ValidateStructPtr(tpe reflect.Type) error {
	if tpe.Kind() != reflect.Ptr {
		return fmt.Errorf("%v is not pointer", tpe)
	}
	unwrap := tpe.Elem()
	if unwrap.Kind() != reflect.Struct {
		return fmt.Errorf("%v is not pointer of struct", tpe)
	}
	return nil
}

// ValidateFn validation function, will only success if fn has signature of func(context.Context, any) (any, error)
func ValidateFn(fn any) error {
	var err error
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		return fmt.Errorf("%v is not a function", fn)
	}
	if fnType.NumIn() != 2 || fnType.NumOut() != 2 {
		return fmt.Errorf("expected 2 arguments and 2 results, got %d arguments and %d results", fnType.NumIn(), fnType.NumOut())
	}
	contextField := fnType.In(0)
	argumentField := fnType.In(1)
	resultField := fnType.Out(0)
	errorField := fnType.Out(1)
	if !contextField.Implements(contextInterface) {
		return fmt.Errorf("%v is not context.Context", contextField)
	}
	if err = ValidateStructPtr(argumentField); err != nil {
		return err
	}
	if err = ValidateStructPtr(resultField); err != nil {
		return err
	}
	if !errorField.Implements(errorInterface) {
		return fmt.Errorf("%v is not error", errorField)
	}
	return nil
}

// InitArgument init argument struct of fn, fn must have signature of func(context.Context, any) (any, error)
func InitArgument(fn any) any {
	fnType := reflect.TypeOf(fn)
	argumentField := fnType.In(1)
	argumentFieldStructType := argumentField.Elem()
	ptr := reflect.New(argumentFieldStructType)
	return ptr.Interface()
}

// InitResult init result struct of fn, fn must have signature of func(context.Context, any) (any, error)
func InitResult(fn any) any {
	// fn must be pre-validated by ValidateFn first
	fnType := reflect.TypeOf(fn)
	argumentField := fnType.Out(0)
	argumentFieldStructType := argumentField.Elem()
	ptr := reflect.New(argumentFieldStructType)
	return ptr.Interface()
}

// CallFn execute function fn without validations, fn must have signature of func(context.Context, any) (any, error)
func CallFn(fn any, ctx context.Context, input any) (any, error) {
	fnValue := reflect.ValueOf(fn)
	results := fnValue.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(input),
	})
	r := results[0]
	e := results[1]
	if r.IsNil() {
		if e.IsNil() {
			return nil, nil
		} else {
			return nil, e.Interface().(error)
		}
	} else {
		if e.IsNil() {
			return r.Interface(), nil
		} else {
			return r.Interface(), e.Interface().(error)
		}
	}
}
