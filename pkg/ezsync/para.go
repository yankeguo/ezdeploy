package ezsync

import (
	"context"
)

func DoPara[T any](ctx context.Context, vs []T, concurrency int, fn func(ctx context.Context, v T) (err error)) (err error) {
	pg := NewParaGroup(concurrency)
	eg := NewErrorGroup()
	for _, _v := range vs {
		v := _v
		pg.Mark()
		go func() {
			pg.Take()
			defer pg.Done()
			eg.Add(fn(ctx, v))
		}()
	}
	pg.Wait()
	err = eg.Unwrap()
	return
}
