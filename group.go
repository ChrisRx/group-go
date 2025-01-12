package group // import "go.chrisrx.dev/group"

import (
	"context"
	"sync"
)

type GroupOption func(*Group)

func WithLimit(n int) GroupOption {
	return func(g *Group) {
		g.limit = make(chan struct{}, n)
	}
}

type Group struct {
	wg sync.WaitGroup

	ctx    context.Context
	cancel context.CancelCauseFunc

	limit chan struct{}
	ready chan struct{}

	once sync.Once
	err  error
}

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

func (g *Group) Go(fn func(context.Context) error) *Group {
	if g.ready != nil {
		close(g.ready)
		g.ready = nil
	}

	if g.limit != nil {
		g.limit <- struct{}{}
	}

	g.wg.Add(1)
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

func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel(g.err)
	}
	return g.err
}

func (g *Group) WaitC() <-chan error {
	errch := make(chan error)
	go func() {
		defer close(errch)
		<-g.ready
		if err := g.Wait(); err != nil {
			errch <- err
		}
	}()
	return errch
}
