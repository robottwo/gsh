package debounce

import (
	"sync"
	"testing"
	"time"
)

// TestDebounce ensures that multiple rapid calls to the debounced function
// only result in a single invocation of the provided function after the debounce period.
func TestDebounce(t *testing.T) {
	var callCount int
	var mu sync.Mutex

	fn := func() {
		mu.Lock()
		defer mu.Unlock()
		callCount++
	}

	debouncedFn := Debounce(100*time.Millisecond, fn)

	// Call the debounced function multiple times in quick succession
	for i := 0; i < 5; i++ {
		debouncedFn()
		time.Sleep(10 * time.Millisecond) // simulate rapid calls
	}

	// At this point, fn should not have been called yet, since the debounce period hasn't elapsed.
	mu.Lock()
	if callCount != 0 {
		t.Errorf("Expected callCount to be 0 before debounce period, got %d", callCount)
	}
	mu.Unlock()

	// Wait for the debounce period to pass
	time.Sleep(150 * time.Millisecond)

	// Now fn should have been called exactly once.
	mu.Lock()
	defer mu.Unlock()
	if callCount != 1 {
		t.Errorf("Expected callCount to be 1 after debounce period, got %d", callCount)
	}
}

// TestConsecutiveDebounce ensures that if calls resume before the previous debounce completes,
// the timer resets and only one call is made after the final series of calls.
func TestConsecutiveDebounce(t *testing.T) {
	var callCount int
	var mu sync.Mutex

	fn := func() {
		mu.Lock()
		callCount++
		mu.Unlock()
	}

	debouncedFn := Debounce(50*time.Millisecond, fn)

	// Call once
	debouncedFn()

	// Wait less than the debounce period, call again
	time.Sleep(30 * time.Millisecond)
	debouncedFn()

	// Wait again less than the debounce period, call again
	time.Sleep(30 * time.Millisecond)
	debouncedFn()

	// Now wait long enough for the debounce to trigger
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if callCount != 1 {
		t.Errorf("Expected callCount to be 1, got %d", callCount)
	}
}
