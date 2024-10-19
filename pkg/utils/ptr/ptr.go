package ptr

func Ptr[T any](v T) *T {
	return &v
}

func PtrValue[T any](v *T, ifNil T) T {
	if v == nil {
		return ifNil
	}
	return *v
}
