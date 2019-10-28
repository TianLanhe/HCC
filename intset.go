package main

import (
	"bytes"
	"fmt"
)

type IntSet struct {
	set  []uint
}

const BITS_NUM = 32 << (^uint(0) >> 63)

func (s IntSet) Has(x int) bool {
	word, bit := x/BITS_NUM, uint(x%BITS_NUM)
	return word < len(s.set) && s.set[word]&(1<<bit) != 0
}

func (s *IntSet) Add(x int) {
	word, bit := x/BITS_NUM, uint(x%BITS_NUM)
	if word >= len(s.set) {
		newSet := make([]uint, word+1)
		copy(newSet, s.set)
		s.set = newSet
	}
	s.set[word] |= (1 << bit)
}

func (s *IntSet) UnionWith(t IntSet) {
	for i, word := range t.set {
		if i < len(s.set) {
			s.set[i] |= word
		} else {
			s.set = append(s.set, word)
		}
	}
}

func (s IntSet) String() string {
	var buf bytes.Buffer
	for i, word := range s.set {
		if word == 0 {
			continue
		}
		for j := 0; j < BITS_NUM; j++ {
			if word&(1<<uint(j)) != 0 {
				if buf.Len() > 0 {
					buf.WriteByte(',')
				}
				fmt.Fprintf(&buf, "%02d", i*BITS_NUM+j)
			}
		}
	}
	return buf.String()
}

func (s *IntSet) Len() int {
	count := 0
	for _, word := range s.set {
		for word != 0 {
			word &= word - 1
			count++
		}
	}
	return count
}

func (s *IntSet) AddAll(nums ...int) {
	for _, x := range nums {
		s.Add(x)
	}
}

func (s *IntSet) Remove(x int) {
	word, bit := x/BITS_NUM, uint(x%BITS_NUM)
	if word < len(s.set) {
		s.set[word] &^= (1 << bit)
	}
}

func (s *IntSet) Clear() {
	s.set = make([]uint, 0)
}

func (s *IntSet) Copy() *IntSet {
	newSet := make([]uint, len(s.set))
	copy(newSet, s.set)
	return &IntSet{
		set:  newSet,
	}
}

func (s *IntSet) IntersectWith(t IntSet) {
	for i := range s.set {
		if i == len(t.set) {
			s.set = s.set[:i]
			break
		}
		s.set[i] &= t.set[i]
	}
}

func (s *IntSet) DifferentWith(t IntSet) {
	for i, word := range t.set {
		if i >= len(s.set) {
			break
		}
		s.set[i] &^= word
	}
}

func (s *IntSet) SymmetricDifferentWith(t IntSet) {
	/*tempSet := s.Copy()
	s.DifferentWith(t)
	t.DifferentWith(*tempSet)
	s.UnionWith(t)*/
	for i, word := range t.set {
		if i >= len(s.set) {
			s.set = append(s.set, word)
		} else {
			s.set[i] ^= word
		}
	}
}

func (s *IntSet) Elems() []int {
	var slice []int
	for i, word := range s.set {
		if word == 0 {
			continue
		}
		for j := 0; j < BITS_NUM; j++ {
			if word&(1<<uint(j)) != 0 {
				slice = append(slice, i*BITS_NUM+j)
			}
		}
	}
	return slice
}

func (s *IntSet) Equal(t IntSet) bool {
	var i int

	for i < len(s.set) && i < len(t.set) {
		if s.set[i] != t.set[i] {
			return false
		}
		i++
	}

	for j := i; j < len(s.set); j++ {
		if s.set[j] != 0 {
			return false
		}
	}

	for j := i; j < len(t.set); j++ {
		if t.set[j] != 0 {
			return false
		}
	}

	return true
}

// package main

// import (
// 	"bytes"
// 	"fmt"
// )

// type IntSet struct {
// 	set  []uint
// 	size int
// }

// const BITS_NUM = 32 << (^uint(0) >> 63)

// func (s IntSet) Has(x int) bool {
// 	word, bit := x/BITS_NUM, uint(x%BITS_NUM)
// 	return word < len(s.set) && s.set[word]&(1<<bit) != 0
// }

// func (s *IntSet) Add(x int) {
// 	if s.Has(x) {
// 		return
// 	}

// 	word, bit := x/BITS_NUM, uint(x%BITS_NUM)
// 	if word >= len(s.set) {
// 		newSet := make([]uint, word+1)
// 		copy(newSet, s.set)
// 		s.set = newSet
// 	}
// 	s.set[word] |= (1 << bit)
// 	s.size++
// }

// func (s *IntSet) UnionWith(t IntSet) {
// 	for i, word := range t.set {
// 		if i < len(s.set) {
// 			s.set[i] |= word
// 		} else {
// 			s.set = append(s.set, word)
// 		}
// 	}
// 	s.size = s.len()
// }

// func (s IntSet) String() string {
// 	var buf bytes.Buffer
// 	for i, word := range s.set {
// 		if word == 0 {
// 			continue
// 		}
// 		for j := 0; j < BITS_NUM; j++ {
// 			if word&(1<<uint(j)) != 0 {
// 				if buf.Len() > 0 {
// 					buf.WriteByte(',')
// 				}
// 				fmt.Fprintf(&buf, "%02d", i*BITS_NUM+j)
// 			}
// 		}
// 	}
// 	return buf.String()
// }

// func (s *IntSet) Len() int {
// 	return s.size
// }

// func (s *IntSet) len() int {
// 	count := 0
// 	for _, word := range s.set {
// 		for word != 0 {
// 			word &= word - 1
// 			count++
// 		}
// 	}
// 	return count
// }

// func (s *IntSet) AddAll(nums ...int) {
// 	for _, x := range nums {
// 		s.Add(x)
// 	}
// }

// func (s *IntSet) Remove(x int) {
// 	if s.Has(x) {
// 		word, bit := x/BITS_NUM, uint(x%BITS_NUM)
// 		if word < len(s.set) {
// 			s.set[word] &^= (1 << bit)
// 		}
// 		s.size--
// 	}
// }

// func (s *IntSet) Clear() {
// 	s.set = make([]uint, 0)
// 	s.size = 0
// }

// func (s *IntSet) Copy() *IntSet {
// 	newSet := make([]uint, len(s.set))
// 	copy(newSet, s.set)
// 	return &IntSet{
// 		set:  newSet,
// 		size: s.size,
// 	}
// }

// func (s *IntSet) IntersectWith(t IntSet) {
// 	for i := range s.set {
// 		if i == len(t.set) {
// 			s.set = s.set[:i]
// 			break
// 		}
// 		s.set[i] &= t.set[i]
// 	}
// 	s.size = s.len()
// }

// func (s *IntSet) DifferentWith(t IntSet) {
// 	for i, word := range t.set {
// 		if i >= len(s.set) {
// 			break
// 		}
// 		s.set[i] &^= word
// 	}
// 	s.size = s.len()
// }

// func (s *IntSet) SymmetricDifferentWith(t IntSet) {
// 	/*tempSet := s.Copy()
// 	s.DifferentWith(t)
// 	t.DifferentWith(*tempSet)
// 	s.UnionWith(t)*/
// 	for i, word := range t.set {
// 		if i >= len(s.set) {
// 			s.set = append(s.set, word)
// 		} else {
// 			s.set[i] ^= word
// 		}
// 	}
// 	s.size = s.len()
// }

// func (s *IntSet) Elems() []int {
// 	var slice []int
// 	for i, word := range s.set {
// 		if word == 0 {
// 			continue
// 		}
// 		for j := 0; j < BITS_NUM; j++ {
// 			if word&(1<<uint(j)) != 0 {
// 				slice = append(slice, i*BITS_NUM+j)
// 			}
// 		}
// 	}
// 	return slice
// }

// func (s *IntSet) Equal(t IntSet) bool {
// 	var i int

// 	for i < len(s.set) && i < len(t.set) {
// 		if s.set[i] != t.set[i] {
// 			return false
// 		}
// 		i++
// 	}

// 	for j := i; j < len(s.set); j++ {
// 		if s.set[j] != 0 {
// 			return false
// 		}
// 	}

// 	for j := i; j < len(t.set); j++ {
// 		if t.set[j] != 0 {
// 			return false
// 		}
// 	}

// 	return true
// }
