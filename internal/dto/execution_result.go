package dto

type ExecutionResult struct {
	Result *[]byte
	Error  *Error
}

func ExtractErrorFromFnCallError(err error) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Message: err.Error(),
	}
}

func ExtractResultFromFnCallResult(result any, marshaller func(any) ([]byte, error)) (*[]byte, error) {
	if result == nil {
		return nil, nil
	}
	b, err := marshaller(result)
	if err != nil {
		return nil, err
	}
	return &b, nil
}
