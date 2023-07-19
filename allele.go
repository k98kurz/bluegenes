package bluegenes

import (
	"math"
	"sort"
	"sync"
)

type Allele[T comparable] struct {
	Name  string
	Genes []*Gene[T]
	Mu    sync.RWMutex
}

func (a *Allele[T]) Copy() *Allele[T] {
	a.Mu.RLock()
	defer a.Mu.RUnlock()
	var another Allele[T]
	another.Name = a.Name
	another.Genes = make([]*Gene[T], len(a.Genes))
	copy(another.Genes, a.Genes)
	return &another
}

func (self *Allele[T]) Insert(index int, gene *Gene[T]) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	if 0 > index || index > len(self.Genes) {
		return indexError{}
	}
	if len(self.Genes) == 0 {
		self.Genes = append(self.Genes[:], gene)
	} else {
		self.Genes = append(self.Genes[:index+1], self.Genes[index:]...)
	}
	self.Genes[index] = gene
	return nil
}

func (self *Allele[T]) Append(gene *Gene[T]) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	self.Genes = append(self.Genes[:], gene)
	return nil
}

func (self *Allele[T]) Duplicate(index int) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	if 0 > index || index > len(self.Genes) {
		return indexError{}
	}
	Gene := self.Genes[index].Copy()
	Genes := append(self.Genes[:index], Gene)
	self.Genes = append(Genes, self.Genes[index:]...)
	return nil
}

func (self *Allele[T]) Delete(index int) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	if 0 > index || index >= len(self.Genes) {
		return indexError{}
	}
	self.Genes = append(self.Genes[:index], self.Genes[index+1:]...)
	return nil
}

func (self *Allele[T]) Substitute(index int, gene *Gene[T]) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	if 0 > index || index >= len(self.Genes) {
		return indexError{}
	}
	genes := append(self.Genes[:index], gene)
	self.Genes = append(genes, self.Genes[index+1:]...)
	return nil
}

func (self *Allele[T]) Recombine(other *Allele[T], indices []int,
	child *Allele[T], options RecombineOptions) error {
	self.Mu.RLock()
	defer self.Mu.RUnlock()
	other.Mu.RLock()
	defer other.Mu.RUnlock()
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
			return indexError{}
		}
	}

	name := self.Name
	if name != other.Name {
		name_size, err := min(len(name), len(other.Name))
		if err != nil {
			return err
		}
		name_swap := RandomInt(1, name_size-1)
		name = self.Name[:name_swap] + other.Name[name_swap:]
	}
	child.Name = name

	for len(child.Genes) < max_size {
		child.Genes = append(child.Genes, &Gene[T]{})
	}

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

	if !options.RecombineGenes.Ok() || options.RecombineGenes.Val {
		for i := 0; i < min_size; i++ {
			if (options.MatchGenes.Ok() &&
				!options.MatchGenes.Val) ||
				!options.MatchGenes.Ok() ||
				genes[i].Name == other_genes[i].Name {
				err := genes[i].Recombine(other_genes[i], []int{},
					child.Genes[i], options)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (self *Allele[T]) ToMap() map[string][]map[string][]T {
	self.Mu.RLock()
	defer self.Mu.RUnlock()
	serialized := make(map[string][]map[string][]T)
	serialized[self.Name] = []map[string][]T{}
	for _, gene := range self.Genes {
		serialized[self.Name] = append(serialized[self.Name], gene.ToMap())
	}
	return serialized
}

func (self *Allele[T]) Sequence(separator []T) []T {
	self.Mu.RLock()
	defer self.Mu.RUnlock()
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
