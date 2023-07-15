package genetics

import (
	"math"
	"sort"
	"sync"
)

type Gene[T comparable] struct {
	name  string
	bases []T
	mu    sync.RWMutex
}

func (self *Gene[T]) Copy() *Gene[T] {
	self.mu.RLock()
	defer self.mu.RUnlock()
	var another Gene[T]
	another.name = self.name
	another.bases = make([]T, len(self.bases))
	copy(another.bases, self.bases)
	return &another
}

func (self *Gene[T]) Insert(index int, base T) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if 0 > index || index > len(self.bases) {
		return IndexError{}
	}
	self.bases = append(self.bases[:index+1], self.bases[index:]...)
	self.bases[index] = base
	return nil
}

func (self *Gene[T]) Append(base T) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.bases = append(self.bases[:], base)
	return nil
}

func (self *Gene[T]) InsertSequence(index int, sequence []T) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if 0 > index || index > len(self.bases) {
		return IndexError{}
	}
	bases := append(self.bases[:index], sequence...)
	self.bases = append(bases, self.bases[index:]...)
	return nil
}

func (self *Gene[T]) Delete(index int) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if 0 > index || index >= len(self.bases) {
		return IndexError{}
	}
	self.bases = append(self.bases[:index], self.bases[index+1:]...)
	return nil
}

func (self *Gene[T]) DeleteSequence(index int, size int) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if 0 > index || index >= len(self.bases) {
		return IndexError{}
	}
	if size == 0 {
		return Error{"size must be >0"}
	}
	self.bases = append(self.bases[:index], self.bases[index+size:]...)
	return nil
}

func (self *Gene[T]) Substitute(index int, base T) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if 0 > index || index >= len(self.bases) {
		return IndexError{}
	}
	bases := append(self.bases[:index], base)
	self.bases = append(bases, self.bases[index+1:]...)
	return nil
}

func (self *Gene[T]) Recombine(other *Gene[T], indices []int, options RecombineOptions) (*Gene[T], error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	other.mu.RLock()
	defer other.mu.RUnlock()
	another := &Gene[T]{}
	min_size, _ := min(len(self.bases), len(other.bases))
	max_size, _ := max(len(self.bases), len(other.bases))

	if len(indices) == 0 && min_size > 1 {
		max_swaps := math.Ceil(math.Log(float64(min_size)))
		swaps, _ := max(RandomInt(0, int(max_swaps)), 1)
		idxSet := NewSet[int]()
		for i := 0; i < swaps; i++ {
			idxSet.Add(RandomInt(0, min_size))
		}
		indices = idxSet.ToSlice()
		sort.Ints(indices)
	}
	for _, i := range indices {
		if 0 > i || i >= min_size {
			return another, IndexError{}
		}
	}

	name := self.name
	if name != other.name {
		name_size, err := min(len(name), len(other.name))
		if err != nil {
			return another, err
		}
		name_swap := RandomInt(1, name_size-1)
		name = self.name[:name_swap] + other.name[name_swap:]
	}
	another.name = name

	bases := make([]T, max_size)
	copy(bases, self.bases)
	swapped := false
	for _, i := range indices {
		if swapped {
			bases = append(bases[:i], self.bases[i:]...)
		} else {
			bases = append(bases[:i], other.bases[i:]...)
		}
		swapped = !swapped
	}
	another.bases = bases

	return another, nil
}

func (self *Gene[T]) ToMap() map[string][]T {
	self.mu.RLock()
	defer self.mu.RUnlock()
	serialized := make(map[string][]T)
	serialized[self.name] = self.bases
	return serialized
}

func (self *Gene[T]) Sequence() []T {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.bases
}
