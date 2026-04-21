package eventbus

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

type Handler[T any] func(ctx context.Context, event T) error

type ErrorHandler func(ctx context.Context, err error)

type Bus[T any] struct {
	mu       sync.RWMutex
	handlers map[uint64]Handler[T]
	nextID   uint64

	onError ErrorHandler

	// 用于限制 PublishAsync 的 goroutine 并发数；nil 表示不限制
	asyncSem chan struct{}
}

func New[T any](opts ...Option[T]) *Bus[T] {
	b := &Bus[T]{
		handlers: make(map[uint64]Handler[T]),
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

type Option[T any] func(*Bus[T])

func WithErrorHandler[T any](fn ErrorHandler) Option[T] {
	return func(b *Bus[T]) {
		b.onError = fn
	}
}

func WithAsyncConcurrency[T any](n int) Option[T] {
	return func(b *Bus[T]) {
		if n > 0 {
			b.asyncSem = make(chan struct{}, n)
		}
	}
}

// Subscribe 返回 handlerID 和取消订阅函数
func (b *Bus[T]) Subscribe(h Handler[T]) (uint64, func()) {
	if h == nil {
		panic("eventbus: nil handler")
	}

	id := atomic.AddUint64(&b.nextID, 1)

	b.mu.Lock()
	b.handlers[id] = h
	b.mu.Unlock()

	return id, func() {
		b.Unsubscribe(id)
	}
}

func (b *Bus[T]) SubscribeOnce(h Handler[T]) (uint64, func()) {
	if h == nil {
		panic("eventbus: nil handler")
	}

	var id uint64
	var called atomic.Bool

	wrapper := func(ctx context.Context, event T) error {
		if !called.CompareAndSwap(false, true) {
			return nil
		}
		b.Unsubscribe(id)
		return h(ctx, event)
	}

	id = atomic.AddUint64(&b.nextID, 1)

	b.mu.Lock()
	b.handlers[id] = wrapper
	b.mu.Unlock()

	return id, func() {
		b.Unsubscribe(id)
	}
}

func (b *Bus[T]) Unsubscribe(id uint64) {
	b.mu.Lock()
	delete(b.handlers, id)
	b.mu.Unlock()
}

func (b *Bus[T]) Publish(ctx context.Context, event T) error {
	handlers := b.snapshotHandlers()
	errs := make([]error, 0, len(handlers))
	for _, h := range handlers {
		if err := b.callHandler(ctx, h, event); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (b *Bus[T]) PublishAsync(ctx context.Context, event T) {
	handlers := b.snapshotHandlers()
	for _, h := range handlers {
		go b.callHandlerAsync(ctx, h, event)
	}
}

func (b *Bus[T]) HandlerCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.handlers)
}

func (b *Bus[T]) snapshotHandlers() []Handler[T] {
	b.mu.RLock()
	defer b.mu.RUnlock()

	handlers := make([]Handler[T], 0, len(b.handlers))
	for _, h := range b.handlers {
		handlers = append(handlers, h)
	}
	return handlers
}

func (b *Bus[T]) callHandlerAsync(ctx context.Context, h Handler[T], event T) {
	if b.asyncSem != nil {
		select {
		case b.asyncSem <- struct{}{}:
			defer func() { <-b.asyncSem }()
		case <-ctx.Done():
			b.handleError(ctx, ctx.Err())
			return
		}
	}

	if err := b.callHandler(ctx, h, event); err != nil {
		b.handleError(ctx, err)
	}
}

func (b *Bus[T]) callHandler(ctx context.Context, h Handler[T], event T) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("event handler panic: %v\n%s", r, debug.Stack())
		}
	}()
	if err = h(ctx, event); err != nil {
		return err
	}
	return nil
}

func (b *Bus[T]) handleError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	if b.onError != nil {
		b.onError(ctx, err)
		return
	}
	slog.Error("eventbus handler failed", "error", err)
}
