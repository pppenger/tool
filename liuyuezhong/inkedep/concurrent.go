package main

import (
	"sync"
)

type Concurrent struct {
	*sync.WaitGroup
	*sync.Mutex
	res    []interface{}
	err    error
	worker chan struct{}
}

func NewConcurrent() *Concurrent {
	num := 2
	workers := make(chan struct{}, num)
	for i := 0; i < num; i++ {
		workers <- struct{}{}
	}
	return &Concurrent{
		WaitGroup: &sync.WaitGroup{},
		Mutex:     &sync.Mutex{},
		worker:    workers,
	}
}

func SingleConcurrent() *Concurrent {
	workers := make(chan struct{}, 1)
	for i := 0; i < 1; i++ {
		workers <- struct{}{}
	}
	return &Concurrent{
		WaitGroup: &sync.WaitGroup{},
		Mutex:     &sync.Mutex{},
		worker:    workers,
	}
}

func (c *Concurrent) DoDone() ([]interface{}, error) {
	c.Wait()
	close(c.worker)
	return c.res, c.err
}

func (c *Concurrent) Do(f func() (interface{}, error)) {
	<-c.worker
	c.Add(1)
	go func() {
		ret, err := f()
		c.Lock()
		if c.err == nil && err != nil {
			c.err = err
		}
		if ret != nil {
			c.res = append(c.res, ret)
		}
		c.Unlock()

		c.Done()
		c.worker <- struct{}{}
	}()
}
