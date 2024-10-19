package awaitable

type Awaitable[T any] interface {
	Await() (T, error)
}
