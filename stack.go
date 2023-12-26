package preditor

import "errors"

type Stack[T any] struct {
	data []T
	size int
}

func NewStack[T any](size int) *Stack[T] {
	return &Stack[T]{data: make([]T, size), size: size}
}

var (
	EmptyStack = errors.New("empty stack")
)

func (s *Stack[T]) Pop() (T, error) {
	if len(s.data) == 0 {
		return *new(T), EmptyStack
	}
	last := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return last, nil
}

func (s *Stack[T]) Top() (T, error) {
	if len(s.data) == 0 {
		return *new(T), EmptyStack
	}
	last := s.data[len(s.data)-1]
	return last, nil
}

func (s *Stack[T]) Push(e T) {
	s.data = append(s.data, e)
	if len(s.data) > s.size {
		s.data = []T{e}
	}
}
