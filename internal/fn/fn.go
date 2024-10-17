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
