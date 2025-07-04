package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.chrisrx.dev/group"
)

func TestGroup(t *testing.T) {
	t.Parallel()

	t.Run("basic", func(t *testing.T) {
		g := group.New(t.Context())
		for i := range 10 {
			g.Go(func(ctx context.Context) error {
				time.Sleep(10 * time.Millisecond)
				fmt.Printf("loop %d\n", i)
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("method chaining", func(t *testing.T) {
		if err := group.New(t.Context()).Go(func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		}).Go(func(ctx context.Context) error {
			time.Sleep(500 * time.Millisecond)
			return nil
		}).Wait(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResultGroup(t *testing.T) {
	t.Parallel()

	t.Run("basic", func(t *testing.T) {
		g := group.NewResultGroup[string](t.Context())
		results := make([]group.Future[string], 0)
		for i := range 10 {
			results = append(results, g.Go(func(ctx context.Context) (string, error) {
				time.Sleep(500 * time.Millisecond)
				return fmt.Sprintf("loop %d", i), nil
			}))
		}

		time.Sleep(600 * time.Millisecond)

		for _, result := range results {
			v, err := result.Get()
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println(v)
		}
		if err := g.Wait(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("iterator", func(t *testing.T) {
		g := group.NewResultGroup[string](t.Context())
		for i := range 10 {
			g.Go(func(ctx context.Context) (string, error) {
				time.Sleep(500 * time.Millisecond)
				return fmt.Sprintf("loop %d", i), nil
			})
		}

		time.Sleep(600 * time.Millisecond)

		var i int
		for v, err := range g.Get() {
			if err != nil {
				t.Fatal(err)
			}
			i++
			fmt.Println(v)
		}
		assert.Equal(t, 10, i)
	})
}
