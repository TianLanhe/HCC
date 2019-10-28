package main

type Stack struct {
	stack []int
	len   int
	cap   int
}

func (s *Stack) Push(n int) {
	if s.len == s.cap {
		s.stack = append(s.stack, n)
		s.cap++
	} else {
		s.stack[s.len] = n
	}
	s.len++
}

func (s *Stack) Pop() int {
	top := s.Top()
	s.len--
	return top
}

func (s *Stack) Top() int {
	if s.len == 0 {
		panic("no elements")
	}

	return s.stack[s.len-1]
}

func (s *Stack) Empty() bool {
	return s.len == 0
}

func (s *Stack) Len() int {
	return s.len
}

func (s *Stack) Elems() []int {
	return s.stack[:s.len]
}



type SetStack struct {
	stack []IntSet
	len   int
	cap   int
}

func (s *SetStack) Push(n IntSet) {
	if s.len == s.cap {
		s.stack = append(s.stack, n)
		s.cap++
	} else {
		s.stack[s.len] = n
	}
	s.len++
}

func (s *SetStack) Pop() IntSet {
	top := s.Top()
	s.len--
	return top
}

func (s *SetStack) Top() IntSet {
	if s.len == 0 {
		panic("no elements")
	}

	return s.stack[s.len-1]
}

func (s *SetStack) Empty() bool {
	return s.len == 0
}

func (s *SetStack) Len() int {
	return s.len
}

func (s *SetStack) Elems() []IntSet {
	return s.stack[:s.len]
}
