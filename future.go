package group

// Future is a value that may not yet be ready.
type Future[T any] interface {
	// Get blocks until the result function is complete.
	Get() (T, error)
}

type future[T any] struct {
	done chan struct{}
	fn   func() (T, error)

	value T
	err   error
}

// NewFuture constructs a new future using the provided result function. The
// result function is called immediately in a new goroutine.
func NewFuture[T any](fn func() (T, error)) Future[T] {
	f := &future[T]{
		done: make(chan struct{}),
		fn:   fn,
	}
	go func() {
		defer close(f.done)
		f.value, f.err = f.fn()
	}()
	return f
}

// Get blocks until result function is complete.
func (f *future[T]) Get() (T, error) {
	<-f.done
	return f.value, f.err
}
