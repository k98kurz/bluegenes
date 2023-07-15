package bluegenes

import (
	"math"
	"sort"
	"sync"
)

type Allele[T comparable] struct {
	Name  string
	Genes []*Gene[T]
	mu    sync.RWMutex
}

func (a *Allele[T]) Copy() *Allele[T] {
	a.mu.RLock()
	defer a.mu.RUnlock()
	var another Allele[T]
	another.Name = a.Name
	another.Genes = make([]*Gene[T], len(a.Genes))
	copy(another.Genes, a.Genes)
	return &another
}

func (self *Allele[T]) Insert(index int, gene *Gene[T]) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if 0 > index || index > len(self.Genes) {
		return indexError{}
	}
	self.Genes = append(self.Genes[:index+1], self.Genes[index:]...)
	self.Genes[index] = gene
	return nil
}

func (self *Allele[T]) Append(gene *Gene[T]) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.Genes = append(self.Genes[:], gene)
	return nil
}

func (self *Allele[T]) Duplicate(index int) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if 0 > index || index > len(self.Genes) {
		return indexError{}
	}
	Gene := self.Genes[index].Copy()
	Genes := append(self.Genes[:index], Gene)
	self.Genes = append(Genes, self.Genes[index:]...)
	return nil
}

func (self *Allele[T]) Delete(index int) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if 0 > index || index >= len(self.Genes) {
		return indexError{}
	}
	self.Genes = append(self.Genes[:index], self.Genes[index+1:]...)
	return nil
}

func (self *Allele[T]) Substitute(index int, gene *Gene[T]) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if 0 > index || index >= len(self.Genes) {
		return indexError{}
	}
	genes := append(self.Genes[:index], gene)
	self.Genes = append(genes, self.Genes[index+1:]...)
	return nil
}

func (self *Allele[T]) Recombine(other *Allele[T], indices []int, options RecombineOptions) (*Allele[T], error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	other.mu.RLock()
	defer other.mu.RUnlock()
	another := &Allele[T]{}
	min_size, _ := min(len(self.Genes), len(other.Genes))
	max_size, _ := max(len(self.Genes), len(other.Genes))

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

	genes := make([]*Gene[T], max_size)
	other_genes := make([]*Gene[T], max_size)
	copy(genes, self.Genes)
	copy(other_genes, other.Genes)
	swapped := false
	for _, i := range indices {
		if swapped {
			genes = append(genes[:i], self.Genes[i:]...)
			other_genes = append(other_genes[:i], other.Genes[i:]...)
		} else {
			genes = append(genes[:i], other.Genes[i:]...)
			other_genes = append(other_genes[:i], self.Genes[i:]...)
		}
		swapped = !swapped
	}

	if !options.RecombineGenes.ok() || options.RecombineGenes.val {
		for i := 0; i < min_size; i++ {
			if (options.MatchGenes.ok() &&
				!options.MatchGenes.val) ||
				genes[i].Name == other_genes[i].Name {
				gene, err := genes[i].Recombine(other_genes[i], []int{}, options)
				if err != nil {
					return another, err
				}
				genes[i] = gene
			}
		}
	}

	another.Genes = genes
	return another, nil
}

func (self *Allele[T]) ToMap() map[string][]map[string][]T {
	self.mu.RLock()
	defer self.mu.RUnlock()
	serialized := make(map[string][]map[string][]T)
	serialized[self.Name] = []map[string][]T{}
	for _, gene := range self.Genes {
		serialized[self.Name] = append(serialized[self.Name], gene.ToMap())
	}
	return serialized
}

func (self *Allele[T]) Sequence(separator []T) []T {
	self.mu.RLock()
	defer self.mu.RUnlock()
	sequence := make([]T, 0)
	parts := make([][]T, 0)

	for _, gene := range self.Genes {
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
