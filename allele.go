package genetics

import (
	"math"
	"sort"
	"sync"
)

type Allele[T comparable] struct {
	name  string
	genes []*Gene[T]
	mu    sync.RWMutex
}

func (a *Allele[T]) Copy() *Allele[T] {
	a.mu.RLock()
	defer a.mu.RUnlock()
	var another Allele[T]
	another.name = a.name
	another.genes = make([]*Gene[T], len(a.genes))
	copy(another.genes, a.genes)
	return &another
}

func (self *Allele[T]) Insert(index int, gene *Gene[T]) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if 0 > index || index > len(self.genes) {
		return IndexError{}
	}
	self.genes = append(self.genes[:index+1], self.genes[index:]...)
	self.genes[index] = gene
	return nil
}

func (self *Allele[T]) Append(gene *Gene[T]) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.genes = append(self.genes[:], gene)
	return nil
}

func (self *Allele[T]) Duplicate(index int) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if 0 > index || index > len(self.genes) {
		return IndexError{}
	}
	gene := self.genes[index].Copy()
	genes := append(self.genes[:index], gene)
	self.genes = append(genes, self.genes[index:]...)
	return nil
}

func (self *Allele[T]) Delete(index int) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if 0 > index || index >= len(self.genes) {
		return IndexError{}
	}
	self.genes = append(self.genes[:index], self.genes[index+1:]...)
	return nil
}

func (self *Allele[T]) Substitute(index int, gene *Gene[T]) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if 0 > index || index >= len(self.genes) {
		return IndexError{}
	}
	genes := append(self.genes[:index], gene)
	self.genes = append(genes, self.genes[index+1:]...)
	return nil
}

func (self *Allele[T]) Recombine(other *Allele[T], indices []int, options RecombineOptions) (*Allele[T], error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	other.mu.RLock()
	defer other.mu.RUnlock()
	another := &Allele[T]{}
	min_size, _ := Min(len(self.genes), len(other.genes))
	max_size, _ := Max(len(self.genes), len(other.genes))

	if len(indices) == 0 && min_size > 1 {
		max_swaps := math.Ceil(math.Log(float64(min_size)))
		swaps, _ := Max(RandomInt(0, int(max_swaps)), 1)
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
		name_size, err := Min(len(name), len(other.name))
		if err != nil {
			return another, err
		}
		name_swap := RandomInt(1, name_size-1)
		name = self.name[:name_swap] + other.name[name_swap:]
	}
	another.name = name

	genes := make([]*Gene[T], max_size)
	other_genes := make([]*Gene[T], max_size)
	copy(genes, self.genes)
	copy(other_genes, other.genes)
	swapped := false
	for _, i := range indices {
		if swapped {
			genes = append(genes[:i], self.genes[i:]...)
			other_genes = append(other_genes[:i], other.genes[i:]...)
		} else {
			genes = append(genes[:i], other.genes[i:]...)
			other_genes = append(other_genes[:i], self.genes[i:]...)
		}
		swapped = !swapped
	}

	if !options.recombine_genes.ok() || options.recombine_genes.val {
		for i := 0; i < min_size; i++ {
			if (options.match_genes.ok() &&
				!options.match_genes.val) ||
				genes[i].name == other_genes[i].name {
				gene, err := genes[i].Recombine(other_genes[i], []int{}, options)
				if err != nil {
					return another, err
				}
				genes[i] = gene
			}
		}
	}

	another.genes = genes
	return another, nil
}

func (self *Allele[T]) ToMap() map[string][]map[string][]T {
	self.mu.RLock()
	defer self.mu.RUnlock()
	serialized := make(map[string][]map[string][]T)
	serialized[self.name] = []map[string][]T{}
	for _, gene := range self.genes {
		serialized[self.name] = append(serialized[self.name], gene.ToMap())
	}
	return serialized
}

func (self *Allele[T]) Sequence(separator []T) []T {
	self.mu.RLock()
	defer self.mu.RUnlock()
	sequence := make([]T, 0)
	parts := make([][]T, 0)

	for _, gene := range self.genes {
		parts = append(parts, gene.Sequence())
	}

	for i, part := range parts {
		if i == 0 {
			sequence = append(sequence, part...)
		} else {
			sequence = append(sequence, separator...)
			sequence = append(sequence, part...)
		}
	}

	return sequence
}
