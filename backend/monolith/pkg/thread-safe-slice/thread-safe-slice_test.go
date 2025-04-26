package thread_safe_slice

import (
	"sync"
	"testing"
)

func TestThreadSafeSlice(t *testing.T) {
	type testCase[T any] struct {
		name      string
		setup     func() *ThreadSafeSlice[T]
		predicate func(T) bool
		limit     int
		expectLen int
	}

	tests := []testCase[int]{
		{
			name: "Add elements",
			setup: func() *ThreadSafeSlice[int] {
				s := NewThreadSafeSlice[int]()
				s.Add(1)
				s.Add(2)
				return s
			},
			predicate: nil,
			expectLen: 2,
		},
		{
			name: "Delete element",
			setup: func() *ThreadSafeSlice[int] {
				s := NewThreadSafeSlice[int]()
				s.Add(1)
				s.Add(2)
				s.DeleteIf(func(x int) bool { return x == 1 })
				return s
			},
			predicate: nil,
			expectLen: 1,
		},
		{
			name: "Filter elements",
			setup: func() *ThreadSafeSlice[int] {
				s := NewThreadSafeSlice[int]()
				s.Add(1)
				s.Add(2)
				s.Add(3)
				return s
			},
			predicate: func(x int) bool { return x%2 == 1 },
			limit:     2,
			expectLen: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.setup()
			if tc.predicate != nil {
				filtered := s.Filter(tc.predicate, tc.limit)
				if len(filtered) != tc.expectLen {
					t.Errorf("expected %d elements, got %d", tc.expectLen, len(filtered))
				}
			} else {
				if len(s.items) != tc.expectLen {
					t.Errorf("expected %d elements, got %d", tc.expectLen, len(s.items))
				}
			}
		})
	}
}

func TestThreadSafeSlice_Concurrent(t *testing.T) {
	s := NewThreadSafeSlice[int]()
	var wg sync.WaitGroup
	n := 100

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(val int) {
			s.Add(val)
			wg.Done()
		}(i)
	}
	wg.Wait()

	if len(s.items) != n {
		t.Errorf("expected %d elements, got %d", n, len(s.items))
	}
}
