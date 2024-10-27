package task

import (
	"context"
)

type DataProvider[T any] interface {
	// Provide need to close the channel internally after the task
	Provide(ctx context.Context) (<-chan T, <-chan error)
}

type DataHandler[T any] interface {
	Execute(ctx context.Context, data T) error
}

type Task[T any] struct {
	provider      DataProvider[T]
	filterManager FilterChain[T]
}

func NewTask[T any](provider DataProvider[T], handler DataHandler[T], filters ...Filter[T]) Executable {
	return &Task[T]{
		provider:      provider,
		filterManager: NewFilterChainManager(handler, filters...),
	}
}

func (this *Task[T]) Execute(ctx context.Context) error {
	ch, errCh := this.provider.Provide(ctx)
	var err error
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			err = context.Cause(ctx)
			if err != nil {
				break FOR_LOOP
			}
		case data, open := <-ch:
			if open {
				err = this.filterManager.Proceed(ctx, &data)
				break FOR_LOOP
			} else {
				continue
			}
		case err1, open := <-errCh:
			if open {
				err = err1
				break FOR_LOOP
			} else {
				continue
			}
		}
	}
	return err
}

type Filter[T any] interface {
	DoFilter(ctx context.Context, data *T, chain FilterChain[T]) error
}

type FilterChain[T any] interface {
	Proceed(ctx context.Context, data *T) error
}

// reference: https://www.baeldung.com/intercepting-filter-pattern-in-java
type FilterChainManager[T any] struct {
	filters []Filter[T]
	target  DataHandler[T]
	index   int
}

func NewFilterChainManager[T any](target DataHandler[T], filters ...Filter[T]) *FilterChainManager[T] {
	return &FilterChainManager[T]{
		filters: filters,
		target:  target,
	}
}

func (this *FilterChainManager[T]) Proceed(ctx context.Context, data *T) error {
	if this.hasNext() {
		next := this.next()
		return next.DoFilter(ctx, data, this)
	}
	return this.target.Execute(ctx, *data)
}

func (this *FilterChainManager[T]) next() Filter[T] {
	f := this.filters[this.index]
	this.index++
	return f
}

func (this *FilterChainManager[T]) hasNext() bool {
	b := this.index < len(this.filters)
	return b
}
