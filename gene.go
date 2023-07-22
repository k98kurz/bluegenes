package bluegenes

import (
	"math"
	"sort"
	"sync"
)

type Gene[T Ordered] struct {
	Name  string
	Bases []T
	Mu    sync.RWMutex
}

func (self *Gene[T]) Copy() *Gene[T] {
	self.Mu.RLock()
	defer self.Mu.RUnlock()
	var another Gene[T]
	another.Name = self.Name
	another.Bases = make([]T, len(self.Bases))
	copy(another.Bases, self.Bases)
	return &another
}

func (self *Gene[T]) Insert(index int, base T) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	if 0 > index || index > len(self.Bases) {
		return indexError{}
	}
	if len(self.Bases) == 0 {
		self.Bases = append(self.Bases[:], base)
	} else {
		self.Bases = append(self.Bases[:index+1], self.Bases[index:]...)
	}
	self.Bases[index] = base
	return nil
}

func (self *Gene[T]) Append(base T) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	self.Bases = append(self.Bases[:], base)
	return nil
}

func (self *Gene[T]) InsertSequence(index int, sequence []T) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	if 0 > index || index > len(self.Bases) {
		return indexError{}
	}
	bases := append(self.Bases[:index], sequence...)
	self.Bases = append(bases, self.Bases[index:]...)
	return nil
}

func (self *Gene[T]) Duplicate(index int) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	if 0 > index || index >= len(self.Bases) {
		return indexError{}
	}
	Base := self.Bases[index]
	Bases := append(self.Bases[:index], Base)
	self.Bases = append(Bases, self.Bases[index:]...)
	return nil
}

func (self *Gene[T]) Delete(index int) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	if 0 > index || index >= len(self.Bases) {
		return indexError{}
	}
	self.Bases = append(self.Bases[:index], self.Bases[index+1:]...)
	return nil
}

func (self *Gene[T]) DeleteSequence(index int, size int) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	if 0 > index || index >= len(self.Bases) {
		return indexError{}
	}
	if size == 0 {
		return anError{"size Must be >0"}
	}
	self.Bases = append(self.Bases[:index], self.Bases[index+size:]...)
	return nil
}

func (self *Gene[T]) Substitute(index int, base T) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	if 0 > index || index >= len(self.Bases) {
		return indexError{}
	}
	bases := append(self.Bases[:index], base)
	self.Bases = append(bases, self.Bases[index+1:]...)
	return nil
}

func (self *Gene[T]) Recombine(other *Gene[T], indices []int, child *Gene[T],
	options RecombineOptions) error {
	self.Mu.RLock()
	defer self.Mu.RUnlock()
	other.Mu.RLock()
	defer other.Mu.RUnlock()
	min_size, _ := min(len(self.Bases), len(other.Bases))
	max_size, _ := max(len(self.Bases), len(other.Bases))

	if len(indices) == 0 && min_size > 1 {
		max_swaps := math.Ceil(math.Log(float64(min_size)))
		swaps, _ := max(RandomInt(0, int(max_swaps)), 1)
		idxSet := newSet[int]()
		for i := 0; i < swaps; i++ {
			idxSet.add(RandomInt(0, min_size))
		}
		indices = idxSet.toSlice()
		sort.Ints(indices)
	}
	for _, i := range indices {
		if 0 > i || i >= min_size {
			return indexError{}
		}
	}

	name := self.Name
	if name != other.Name {
		name_size, err := min(len(name), len(other.Name))
		if err != nil {
			return err
		}
		if name_size > 2 {
			name_swap := RandomInt(1, name_size-1)
			name = self.Name[:name_swap] + other.Name[name_swap:]
		}
	}
	child.Name = name

	bases := make([]T, max_size)
	copy(bases, self.Bases)
	swapped := false
	for _, i := range indices {
		if swapped {
			bases = append(bases[:i], self.Bases[i:]...)
		} else {
			bases = append(bases[:i], other.Bases[i:]...)
		}
		swapped = !swapped
	}
	child.Bases = bases

	return nil
}

func (self *Gene[T]) ToMap() map[string][]T {
	self.Mu.RLock()
	defer self.Mu.RUnlock()
	serialized := make(map[string][]T)
	serialized[self.Name] = self.Bases
	return serialized
}

func (self *Gene[T]) Sequence(placeholder ...[]T) []T {
	self.Mu.RLock()
	defer self.Mu.RUnlock()
	if len(placeholder) > 0 && len(self.Bases) == 0 {
		return placeholder[0]
	}
	return self.Bases
}
