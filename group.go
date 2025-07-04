// Package group provides [group.Group] for managing pools of goroutines.
package group // import "go.chrisrx.dev/group"

import (
	"context"
	"sync"
)

type GroupOption func(*Group)

// WithLimit sets the bounded concurrency for a pool of goroutines.
func WithLimit(n int) GroupOption {
	return func(g *Group) {
		g.limit = make(chan struct{}, n)
	}
}

// Group manages a pool of goroutines.
type Group struct {
	wg sync.WaitGroup

	ctx    context.Context
	cancel context.CancelCauseFunc

	limit chan struct{}

	mu    sync.Mutex
	ready chan struct{}
	done  chan error

	once sync.Once
	err  error
}

// New constructs a new group using the provided options.
func New(ctx context.Context, opts ...GroupOption) *Group {
	g := &Group{
		ready: make(chan struct{}),
	}
	for _, opt := range opts {
		opt(g)
	}
	g.ctx, g.cancel = context.WithCancelCause(ctx)
	return g
}

// Go runs the provided function in a goroutine. If an error is encountered,
// the context for the group is canceled.
//
// If a concurrency limit is set, calls to Go will block once the number of
// running goroutines is reached and will continue blocking until a running
// goroutine returns.
func (g *Group) Go(fn func(context.Context) error) *Group {
	if g.limit != nil {
		g.limit <- struct{}{}
	}

	g.wg.Add(1)

	g.mu.Lock()
	if g.ready != nil {
		close(g.ready)
		g.ready = nil
	}
	g.mu.Unlock()

	go func() {
		defer func() {
			if g.limit != nil {
				<-g.limit
			}
			g.wg.Done()
		}()

		if err := fn(g.ctx); err != nil {
			g.once.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel(g.err)
				}
			})
		}
	}()
	return g
}

// Wait blocks until all the goroutines in this group have returned. If any
// errors occur, the first error encountered will be returned. It will also
// block until at least one goroutine is scheduled.
func (g *Group) Wait() error {
	g.mu.Lock()
	ready := g.ready
	g.mu.Unlock()
	if ready != nil {
		<-ready
	}

	g.wg.Wait()
	if g.cancel != nil {
		g.cancel(g.err)
	}
	return g.err
}

// Done blocks until all the goroutines in this group have returned. If any
// errors occur, the first error encountered is sent on the returned channel,
// otherwise the channel is closed.
func (g *Group) Done() <-chan error {
	g.mu.Lock()
	if g.done == nil {
		g.done = make(chan error)
		go func() {
			g.done <- g.Wait()
			g.mu.Lock()
			close(g.done)
			g.done = nil
			g.mu.Unlock()
		}()
	}
	g.mu.Unlock()
	return g.done
}
