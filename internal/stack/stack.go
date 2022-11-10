package stack

import "sync"

type (
	Stack[K comparable] struct {
		mu     sync.RWMutex
		values []K
		errs   []error
	}
)

func (s *Stack[K]) Push(k K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = append(s.values, k)
	s.errs = append(s.errs, nil)
	return
}

func (s *Stack[K]) Pop() (k K, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.values) <= 0 {
		return
	}
	i := len(s.values) - 1
	k = s.values[i]
	ok = true
	s.values = s.values[:i]
	s.errs = s.errs[:i]
	return
}

func (s *Stack[K]) Peek() (k K, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.values) <= 0 {
		return
	}
	k = s.values[len(s.values)-1]
	ok = true
	return
}

func (s *Stack[K]) Log(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errs[len(s.values)-1] = err
	return
}

func (s *Stack[K]) Err() (err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.values) <= 0 {
		return
	}
	err = s.errs[len(s.errs)-1]
	return
}

func (s *Stack[K]) Has(k K) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.values {
		if v == k {
			return true
		}
	}
	return false
}

func (s *Stack[K]) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.values)
}

func (s *Stack[K]) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = nil
	s.errs = nil
	return
}
