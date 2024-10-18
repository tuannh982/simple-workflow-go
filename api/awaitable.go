package api

type Awaitable[T any] interface {
	Await() (T, error)
}
