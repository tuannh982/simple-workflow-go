package commons

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func GetOrElse[T any](fn func() (T, error), defaultValue T) T {
	v, err := fn()
	if err != nil {
		return defaultValue
	}
	return v
}
