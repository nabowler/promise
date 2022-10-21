package promise

import (
	"context"
	"sync"
)

type (
	// Promise is a blocking function that will return the same (T, error) on every call
	Promise[T any] func() (T, error)

	// PromiseNoError is a blocking function that will return the same T on every call
	PromiseNoError[T any] func() T

	// Complete is a non-blocking function that will fulfill a Promise created by You
	Complete[T any] func(T, error)

	// Complete is a non-blocking function that will fulfill a Promise created by YouNoError
	CompleteNoError[T any] func(T)

	tuple[T any] struct {
		val T
		err error
	}
)

// Me returns a Promise that will provide the result of complete.
// If the Context is done before complete, the default value for T
// and ctx.Err() will be returned.
func Me[T any](ctx context.Context, complete func() (T, error)) Promise[T] {
	p, c := You[T](ctx)

	go func() {
		t, err := complete()
		c(t, err)
	}()

	return p
}

// Me returns a Promise that will provide the result of complete.
// If the Context is done before complete, the default value for T
// is returned and ctx.Err() will be ignored.
func MeNoError[T any](ctx context.Context, complete func() T) PromiseNoError[T] {
	p, c := YouNoError[T](ctx)

	go func() {
		c(complete())
	}()

	return p
}

// You returns a Promise and a Completion.
// The Promise will block until Complete is called.
// The first call to Complete will set the return values for the Promise.
// Subsequent calls to Complete will no-op.
func You[T any](ctx context.Context) (Promise[T], Complete[T]) {
	// a buffered channel of 1 is used to ensure that if ctx.Done()
	// is the selected case in the promise readOnce, the first call to Complete
	// will still be able to write to the channel without blocking
	// and everything _should_ be able to be garbage collected.
	ch := make(chan tuple[T], 1)

	// Create a promise that will block until the first call to Complete.
	// Subsequent calls will return the same results.
	readOnce := sync.Once{}
	var tup tuple[T]
	p := func() (T, error) {
		readOnce.Do(func() {
			select {
			case tup = <-ch:
			case <-ctx.Done():
				tup.err = ctx.Err()
			}
		})

		return tup.val, tup.err
	}

	// create a Complete that will only allow a single call to set the value
	setOnce := sync.Once{}
	complete := func(t T, err error) {
		setOnce.Do(func() {
			ch <- tuple[T]{t, err}
			close(ch)
		})
	}

	return p, complete
}

// YouNoError returns a Promise and a Completion.
// The Promise will block until Complete is called.
// The first call to Complete will set the return value for the Promise.
// Subsequent calls to Complete will no-op.
func YouNoError[T any](ctx context.Context) (PromiseNoError[T], CompleteNoError[T]) {
	// as above, so below (just without an error value to consider)
	ch := make(chan T, 1)

	readOnce := sync.Once{}
	var t T
	p := func() T {
		readOnce.Do(func() {
			select {
			case t = <-ch:
			case <-ctx.Done():
			}
		})
		return t
	}

	setOnce := sync.Once{}
	complete := func(t T) {
		setOnce.Do(func() {
			ch <- t
			close(ch)
		})
	}

	return p, complete
}
