[![API Reference](https://img.shields.io/badge/api-reference-blue.svg)](https://pkg.go.dev/mod/go.chrisrx.dev/group)

# group

group is a library for managing pools of goroutines. It has been adapted from [errgoup](https://pkg.go.dev/golang.org/x/sync@v0.15.0/errgroup) with small improvements to the API. It has zero dependencies and is intentionally very simple.

## Usage

This will create a new group and start 10 goroutines:

```go
g := group.New(ctx)
for i := range 10 {
	g.Go(func(ctx context.Context) error {
		fmt.Printf("loop %d\n", i)
		return nil
	})
}
if err := g.Wait(); err != nil {
	log.Fatal(err)
}
```

The parent context provided is used to create a child context that group uses internally, which is passed through to each goroutine. If any goroutine produces an error, this child context is canceled, allowing the other goroutines to stop/cleanup:

The option `WithLimit` can be passed to the group constructor to establish a bound on concurrency:


```go
g := group.New(ctx, group.WithLimit(2))
for i := range 10 {
	g.Go(func(ctx context.Context) error {
		fmt.Printf("loop %d\n", i)
		return nil
	})
}
if err := g.Wait(); err != nil {
	log.Fatal(err)
}
```

Here, only 2 goroutines will ever be running at a given time.

## Notes

* https://github.com/golang/go/issues/57534
