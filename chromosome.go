package bluegenes

import (
	"math"
	"sort"
	"sync"
)

type Chromosome[T comparable] struct {
	Name    string
	alleles []*Allele[T]
	Mu      sync.RWMutex
}

func (c *Chromosome[T]) Copy() *Chromosome[T] {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	var another Chromosome[T]
	another.Name = c.Name
	another.alleles = make([]*Allele[T], len(c.alleles))
	copy(another.alleles, c.alleles)
	return &another
}

func (self *Chromosome[T]) Insert(index int, allele *Allele[T]) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	if 0 > index || index > len(self.alleles) {
		return indexError{}
	}
	self.alleles = append(self.alleles[:index+1], self.alleles[index:]...)
	self.alleles[index] = allele
	return nil
}

func (self *Chromosome[T]) Append(allele *Allele[T]) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	self.alleles = append(self.alleles[:], allele)
	return nil
}

func (self *Chromosome[T]) Duplicate(index int) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	if 0 > index || index > len(self.alleles) {
		return indexError{}
	}
	allele := self.alleles[index].Copy()
	alleles := append(self.alleles[:index], allele)
	self.alleles = append(alleles, self.alleles[index:]...)
	return nil
}

func (self *Chromosome[T]) Delete(index int) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	if 0 > index || index >= len(self.alleles) {
		return indexError{}
	}
	self.alleles = append(self.alleles[:index], self.alleles[index+1:]...)
	return nil
}

func (self *Chromosome[T]) Substitute(index int, allele *Allele[T]) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	if 0 > index || index >= len(self.alleles) {
		return indexError{}
	}
	alleles := append(self.alleles[:index], allele)
	self.alleles = append(alleles, self.alleles[index+1:]...)
	return nil
}

func (self *Chromosome[T]) Recombine(other *Chromosome[T], indices []int, options RecombineOptions) (*Chromosome[T], error) {
	self.Mu.RLock()
	defer self.Mu.RUnlock()
	other.Mu.RLock()
	defer other.Mu.RUnlock()
	another := &Chromosome[T]{}
	min_size, _ := min(len(self.alleles), len(other.alleles))
	max_size, _ := max(len(self.alleles), len(other.alleles))

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
			return another, indexError{}
		}
	}

	Name := self.Name
	if Name != other.Name {
		Name_size, err := min(len(Name), len(other.Name))
		if err != nil {
			return another, err
		}
		Name_swap := RandomInt(1, Name_size-1)
		Name = self.Name[:Name_swap] + other.Name[Name_swap:]
	}
	another.Name = Name

	alleles := make([]*Allele[T], max_size)
	other_alleles := make([]*Allele[T], max_size)
	copy(alleles, self.alleles)
	copy(other_alleles, other.alleles)
	swapped := false
	for _, i := range indices {
		if swapped {
			alleles = append(alleles[:i], self.alleles[i:]...)
			other_alleles = append(other_alleles[:i], other.alleles[i:]...)
		} else {
			alleles = append(alleles[:i], other.alleles[i:]...)
			other_alleles = append(other_alleles[:i], self.alleles[i:]...)
		}
		swapped = !swapped
	}

	if !options.RecombineAlleles.ok() || options.RecombineAlleles.val {
		for i := 0; i < min_size; i++ {
			if (options.MatchAlleles.ok() &&
				!options.MatchAlleles.val) ||
				alleles[i].Name == other_alleles[i].Name {
				allele, err := alleles[i].Recombine(other_alleles[i], []int{}, options)
				if err != nil {
					return another, err
				}
				alleles[i] = allele
			}
		}
	}

	another.alleles = alleles
	return another, nil
}

func (self *Chromosome[T]) ToMap() map[string][]map[string][]map[string][]T {
	self.Mu.RLock()
	defer self.Mu.RUnlock()
	serialized := make(map[string][]map[string][]map[string][]T)
	serialized[self.Name] = []map[string][]map[string][]T{}
	for _, allele := range self.alleles {
		serialized[self.Name] = append(serialized[self.Name], allele.ToMap())
	}
	return serialized
}

func (self *Chromosome[T]) Sequence(separator []T) []T {
	self.Mu.RLock()
	defer self.Mu.RUnlock()
	sequence := make([]T, 0)
	parts := make([][]T, 0)

	for _, allele := range self.alleles {
		parts = append(parts, allele.Sequence(separator))
	}

	for i, part := range parts {
		if i == 0 {
			sequence = append(sequence, part...)
		} else {
			sequence = append(sequence, separator...)
			sequence = append(sequence, separator...)
			sequence = append(sequence, part...)
		}
	}

	return sequence
}
