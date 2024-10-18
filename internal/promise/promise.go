package promise

import (
	"context"
	"sync"
)

type Promise[T any] struct {
	value *T
	err   error
	done  bool
	ch    chan struct{}
	once  sync.Once
}

func NewPromise[T any]() *Promise[T] {
	return &Promise[T]{
		value: nil,
		err:   nil,
		done:  false,
		ch:    make(chan struct{}),
	}
}

func (p *Promise[T]) Resolve(v *T) {
	p.once.Do(func() {
		p.value = v
		p.done = true
		close(p.ch)
	})
}

func (p *Promise[T]) Reject(err error) {
	p.once.Do(func() {
		p.err = err
		p.done = true
		close(p.ch)
	})
}

func (p *Promise[T]) Value() *T {
	return p.value
}

func (p *Promise[T]) Error() error {
	return p.err
}

func (p *Promise[T]) IsDone() bool {
	return p.done
}

func (p *Promise[T]) Await(ctx context.Context) (*T, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-p.ch:
		return p.value, p.err
	}
}
