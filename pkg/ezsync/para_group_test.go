package ezsync

import (
	"github.com/stretchr/testify/require"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewParaGroup(t *testing.T) {
	pg := NewParaGroup(2)
	var v int64

	pg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			pg.Take()
			defer pg.Done()
			require.LessOrEqual(t, atomic.AddInt64(&v, 1), int64(2))
			time.Sleep(time.Millisecond * 10)
			atomic.AddInt64(&v, -1)
		}()
	}
	pg.Wait()
}
