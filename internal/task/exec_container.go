package task

import (
	"context"

	"github.com/sourcegraph/conc/pool"
)

type Executable interface {
	Execute(ctx context.Context) error
}

const (
	MaxTaskThreads int = 10
)

type ConcExecutableContainer struct {
	taskList []Executable
	pool     *pool.ErrorPool
}

type Option func(combo *ConcExecutableContainer)

func CreateConcExecutableContainer(taskList []Executable, opts ...Option) Executable {
	e := &ConcExecutableContainer{
		taskList: taskList,
		pool:     pool.New().WithMaxGoroutines(MaxTaskThreads).WithErrors(),
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func WthPoolSize(poolSize int) Option {
	return func(combo *ConcExecutableContainer) {
		combo.pool = pool.New().WithMaxGoroutines(poolSize).WithErrors()
	}
}

func (e *ConcExecutableContainer) Execute(ctx context.Context) error {
	for _, task := range e.taskList {
		e.pool.Go(func() error {
			return task.Execute(ctx)
		})
	}
	return e.pool.Wait()
}
