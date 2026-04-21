package eventbus

import (
	"context"
	"reflect"
	"sync"
)

var (
	mu   sync.RWMutex
	buss = make(map[reflect.Type]any)
)

// Register initializes the global bus for T. If the bus already exists, the
// existing instance is returned and later options are ignored.
func Register[T any](opts ...Option[T]) *Bus[T] {
	key := eventTypeOf[T]()

	mu.Lock()
	defer mu.Unlock()

	if bus, ok := buss[key]; ok {
		return bus.(*Bus[T])
	}

	created := New(opts...)
	buss[key] = created
	return created
}

// Get returns the global bus for T, creating it with default options if needed.
func Get[T any]() *Bus[T] {
	key := eventTypeOf[T]()

	mu.RLock()
	bus, ok := buss[key]
	mu.RUnlock()
	if ok {
		return bus.(*Bus[T])
	}

	mu.Lock()
	defer mu.Unlock()

	if bus, ok = buss[key]; ok {
		return bus.(*Bus[T])
	}

	created := New[T]()
	buss[key] = created
	return created
}

func eventTypeOf[T any]() reflect.Type {
	typ := reflect.TypeFor[T]()
	if typ == nil {
		panic("eventbus: nil event type")
	}
	return typ
}

// Subscribe registers a handler on the global bus for T.
func Subscribe[T any](h Handler[T]) (uint64, func()) {
	return Get[T]().Subscribe(h)
}

// SubscribeOnce registers a handler that runs at most once on the global bus for T.
func SubscribeOnce[T any](h Handler[T]) (uint64, func()) {
	return Get[T]().SubscribeOnce(h)
}

// Publish synchronously publishes event to the global bus for T.
func Publish[T any](ctx context.Context, event T) error {
	return Get[T]().Publish(ctx, event)
}

// PublishAsync asynchronously publishes event to the global bus for T.
func PublishAsync[T any](ctx context.Context, event T) {
	Get[T]().PublishAsync(ctx, event)
}
