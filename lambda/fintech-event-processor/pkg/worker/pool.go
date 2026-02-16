package worker

import (
	"sync"
)

// Pool is a simple worker pool for limiting concurrency.
type Pool struct {
	sem chan struct{}
	wg  sync.WaitGroup
}

// NewPool creates a new worker pool with the specified concurrency limit.
func NewPool(workers int) *Pool {
	if workers <= 0 {
		workers = 1
	}

	return &Pool{
		sem: make(chan struct{}, workers),
	}
}

// Submit adds a task to the pool. It blocks if all workers are busy.
func (p *Pool) Submit(task func()) {
	p.sem <- struct{}{}

	p.wg.Add(1)

	go func() {
		defer func() {
			<-p.sem
			p.wg.Done()
		}()

		task()
	}()
}

// Wait blocks until all submitted tasks complete.
func (p *Pool) Wait() {
	p.wg.Wait()
}

// SubmitAndWait submits multiple tasks and waits for completion.
func (p *Pool) SubmitAndWait(tasks []func()) {
	for _, task := range tasks {
		p.Submit(task)
	}

	p.Wait()
}

// Size returns the pool's concurrency limit.
func (p *Pool) Size() int {
	return cap(p.sem)
}

// Running returns the number of currently running tasks.
func (p *Pool) Running() int {
	return len(p.sem)
}
