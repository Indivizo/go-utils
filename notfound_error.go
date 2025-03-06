package go_utils

import (
	"errors"
	"sync"
)

// ErrNotFound is a generic error representing a "not found" case
var ErrNotFound = errors.New("resource not found")

var (
	mu             sync.RWMutex
	notFoundErrors = make(map[error]struct{})
)

// RegisterNotFoundError allows services to register their database-specific "not found" errors
func RegisterNotFoundError(err error) {
	mu.Lock()
	defer mu.Unlock()
	notFoundErrors[err] = struct{}{}
}

// IsNotFoundError checks whether an error is one of the registered "not found" errors
func IsNotFoundError(err error) bool {
	mu.RLock()
	defer mu.RUnlock()

	// Allow checking both registered errors and our generic ErrNotFound
	if errors.Is(err, ErrNotFound) {
		return true
	}

	_, exists := notFoundErrors[err]
	return exists
}
