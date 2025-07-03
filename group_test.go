package group

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestGroup(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g := New(ctx)
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
}
