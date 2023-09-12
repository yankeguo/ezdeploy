package ezsync

import (
	"errors"
	"sync"
)

type ParaGroup struct {
	ch  chan struct{}
	oc  *sync.Once
	_wg *sync.WaitGroup
}

func NewParaGroup(concurrency int) *ParaGroup {
	if concurrency < 1 {
		panic(errors.New("ParaGroup: invalid argument 'concurrency', must > 0"))
	}
	ch := make(chan struct{}, concurrency)
	for i := 0; i < concurrency; i++ {
		ch <- struct{}{}
	}
	return &ParaGroup{
		ch: ch,
		oc: &sync.Once{},
	}
}

func (pg *ParaGroup) wg() *sync.WaitGroup {
	pg.oc.Do(func() {
		pg._wg = &sync.WaitGroup{}
	})
	return pg._wg
}

func (pg *ParaGroup) Mark() {
	pg.wg().Add(1)
}

func (pg *ParaGroup) Add(n int) {
	pg.wg().Add(n)
}

func (pg *ParaGroup) Take() {
	<-pg.ch
}

func (pg *ParaGroup) Done() {
	pg.ch <- struct{}{}
	// if Mark/Add never invoked, just ignore waitGroup
	if pg._wg == nil {
		return
	}
	pg._wg.Done()
}

func (pg *ParaGroup) Wait() {
	pg.wg().Wait()
}
