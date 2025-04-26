package thread_safe_slice

import (
	"errors"
	"sync"
)

var IndexOutOfRangeError = errors.New("index out of range")

type ThreadSafeSlice[T any] struct {
	mu    sync.RWMutex
	items []T
}

// NewThreadSafeSlice creates a new thread-safe slice.
func NewThreadSafeSlice[T any]() *ThreadSafeSlice[T] {
	return &ThreadSafeSlice[T]{items: make([]T, 0)}
}

// NewThreadSafeSliceN creates a new thread-safe slice with a specified capacity.
func NewThreadSafeSliceN[T any](n int) *ThreadSafeSlice[T] {
	if n < 0 {
		return &ThreadSafeSlice[T]{items: make([]T, 0)}
	}
	return &ThreadSafeSlice[T]{items: make([]T, n)}
}

func (s *ThreadSafeSlice[T]) SetByIdx(idx int, item T) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if idx < 0 || idx >= len(s.items) {
		return IndexOutOfRangeError
	}
	s.items[idx] = item

	return nil
}

// Add — добавляет элемент
func (s *ThreadSafeSlice[T]) Add(item T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
}

// DeleteIf deletes an item that satisfies the predicate.
func (s *ThreadSafeSlice[T]) DeleteIf(predicate func(T) bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for idx, item := range s.items {
		if predicate(item) {
			s.items = append(s.items[:idx], s.items[idx+1:]...)
			return nil
		}
	}
	return errors.New("item not found")
}

// Filter returns a slice of items that satisfy the predicate.
func (s *ThreadSafeSlice[T]) Filter(predicate func(T) bool, limit int) []T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []T
	for _, item := range s.items {
		if len(result) == limit {
			break
		}
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

func (s *ThreadSafeSlice[T]) GetSliceCopy() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	copySlice := make([]T, len(s.items))
	for i, v := range s.items {
		copySlice[i] = v
	}
	return copySlice
}
