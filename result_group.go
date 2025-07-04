package group

import (
	"context"
	"iter"
	"sync"
)

const resultsBuffer = 1000

// ResultGroup manages a pool of goroutines that return a result value.
type ResultGroup[T any] struct {
	g *Group

	mu      sync.Mutex
	results chan Future[T]
}

// NewResultGroup constructs a new result group using the provided options.
func NewResultGroup[T any](ctx context.Context, opts ...GroupOption) *ResultGroup[T] {
	return &ResultGroup[T]{
		g:       New(ctx, opts...),
		results: make(chan Future[T], resultsBuffer),
	}
}

// Go runs the provided function in a goroutine and returns a future containing
// a result value or error.
//
// If an error is encountered, the context for the group is canceled. This
// happens regardless if the error is checked on the future.
//
// If a concurrency limit is set, calls to Go will block once the number of
// running goroutines is reached and will continue blocking until a running
// goroutine returns.
func (r *ResultGroup[T]) Go(fn func(context.Context) (T, error)) Future[T] {
	f := NewFuture(func() (T, error) {
		return fn(r.g.ctx)
	})
	r.g.Go(func(ctx context.Context) error {
		_, err := f.Get()

		// The results queue might not ever be read from so anything past the
		// channel buffer should just be dropped. This will be typical in cases
		// where the returned future is handled manually (i.e. instead of using
		// [ResultGroup.Get]).
		select {
		case r.results <- f:
		default:
		}
		return err
	})
	return f
}

// Get returns an iterator of result/error pairs. It blocks until all results
// are read or the group context is done.
func (r *ResultGroup[T]) Get() iter.Seq2[T, error] {
	go func() {
		select {
		case <-r.g.Done():
		case <-r.g.ctx.Done():
		}
		r.mu.Lock()
		close(r.results)
		r.results = make(chan Future[T], resultsBuffer)
		r.mu.Unlock()
	}()

	return func(yield func(T, error) bool) {
		r.mu.Lock()
		results := r.results
		r.mu.Unlock()
		for result := range results {
			if result == nil {
				return
			}
			if !yield(result.Get()) {
				return
			}
		}
	}
}

// Wait blocks until all the goroutines in this group have returned. If any
// errors occur, the first error encountered will be returned. It will also
// block until at least one goroutine is scheduled.
func (r *ResultGroup[T]) Wait() error {
	return r.g.Wait()
}
