package debounce

import (
	"sync"
	"time"
)

// Debounce takes a duration `d` and a function `fn` and returns a new function.
// When the returned function is repeatedly called, `fn` will only be executed
// after there have been no calls for at least `d` duration.
func Debounce(d time.Duration, fn func()) func() {
	var mu sync.Mutex
	var timer *time.Timer

	return func() {
		mu.Lock()
		defer mu.Unlock()

		// If a timer is already running, stop it so we can reset
		if timer != nil {
			timer.Stop()
		}

		// Create a new timer that calls fn after d has passed without new calls
		timer = time.AfterFunc(d, fn)
	}
}

func DebounceWithParam[T any](d time.Duration, fn func(T)) func(T) {
	var mu sync.Mutex
	var timer *time.Timer

	return func(t T) {
		mu.Lock()
		defer mu.Unlock()

		// If a timer is already running, stop it so we can reset
		if timer != nil {
			timer.Stop()
		}

		// Create a new timer that calls fn after d has passed without new calls
		timer = time.AfterFunc(d, func() {
			fn(t)
		})
	}
}
