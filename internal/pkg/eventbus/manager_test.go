package eventbus

import (
	"context"
	"reflect"
	"sync"
	"testing"
)

type userCreatedEvent struct {
	UserID int64
}

type orderCreatedEvent struct {
	OrderID int64
}

func TestGetReturnsSameBusForSameType(t *testing.T) {
	resetManagerForTest(t)

	first := Get[userCreatedEvent]()
	second := Get[userCreatedEvent]()

	if first != second {
		t.Fatalf("expected same bus instance for same event type")
	}
}

func TestGetReturnsDifferentBusForDifferentTypes(t *testing.T) {
	resetManagerForTest(t)

	userBus := Get[userCreatedEvent]()
	orderBus := Get[orderCreatedEvent]()

	if userBus == any(orderBus) {
		t.Fatalf("expected different bus instances for different event types")
	}
}

func TestGetCreatesOnlyOneBusUnderConcurrency(t *testing.T) {
	resetManagerForTest(t)

	const workers = 32

	results := make(chan *Bus[userCreatedEvent], workers)
	var wg sync.WaitGroup

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- Get[userCreatedEvent]()
		}()
	}

	wg.Wait()
	close(results)

	var first *Bus[userCreatedEvent]
	for bus := range results {
		if first == nil {
			first = bus
			continue
		}
		if bus != first {
			t.Fatalf("expected same bus instance for concurrent access")
		}
	}
}

func TestRegisterReturnsConfiguredGlobalBus(t *testing.T) {
	resetManagerForTest(t)

	var handled bool
	registered := Register[testEvent](WithErrorHandler[testEvent](func(ctx context.Context, err error) {
		handled = true
	}))
	got := Get[testEvent]()

	if registered != got {
		t.Fatalf("expected registered bus to be returned by Get")
	}

	got.handleError(context.Background(), assertErr{})
	if !handled {
		t.Fatalf("expected registered error handler to be used")
	}
}

type assertErr struct{}

func (assertErr) Error() string {
	return "assert"
}

func resetManagerForTest(t *testing.T) {
	t.Helper()

	mu.Lock()
	buss = make(map[reflect.Type]any)
	mu.Unlock()
}
