package bluegenes

import (
	"math"
	"sort"
	"sync"
)

type Genome[T comparable] struct {
	Name        string
	chromosomes []*Chromosome[T]
	mu          sync.RWMutex
}

func (self *Genome[T]) Copy() *Genome[T] {
	self.mu.RLock()
	defer self.mu.RUnlock()
	var another Genome[T]
	another.Name = self.Name
	another.chromosomes = make([]*Chromosome[T], len(self.chromosomes))
	copy(another.chromosomes, self.chromosomes)
	return &another
}

func (self *Genome[T]) Insert(index int, chromosome *Chromosome[T]) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if 0 > index || index > len(self.chromosomes) {
		return indexError{}
	}
	self.chromosomes = append(self.chromosomes[:index+1], self.chromosomes[index:]...)
	self.chromosomes[index] = chromosome
	return nil
}

func (self *Genome[T]) Append(allele *Chromosome[T]) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.chromosomes = append(self.chromosomes[:], allele)
	return nil
}

func (self *Genome[T]) Duplicate(index int) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if 0 > index || index > len(self.chromosomes) {
		return indexError{}
	}
	chromosome := self.chromosomes[index].Copy()
	chromosomes := append(self.chromosomes[:index], chromosome)
	self.chromosomes = append(chromosomes, self.chromosomes[index:]...)
	return nil
}

func (self *Genome[T]) Delete(index int) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if 0 > index || index >= len(self.chromosomes) {
		return indexError{}
	}
	self.chromosomes = append(self.chromosomes[:index], self.chromosomes[index+1:]...)
	return nil
}

func (self *Genome[T]) Substitute(index int, chromosome *Chromosome[T]) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if 0 > index || index >= len(self.chromosomes) {
		return indexError{}
	}
	chromosomes := append(self.chromosomes[:index], chromosome)
	self.chromosomes = append(chromosomes, self.chromosomes[index+1:]...)
	return nil
}

func (self *Genome[T]) Recombine(other *Genome[T], indices []int, options RecombineOptions) (*Genome[T], error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	other.mu.RLock()
	defer other.mu.RUnlock()
	another := &Genome[T]{}
	min_size, _ := min(len(self.chromosomes), len(other.chromosomes))
	max_size, _ := max(len(self.chromosomes), len(other.chromosomes))

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

	chromosomes := make([]*Chromosome[T], max_size)
	other_chromosomes := make([]*Chromosome[T], max_size)
	copy(chromosomes, self.chromosomes)
	copy(other_chromosomes, other.chromosomes)
	swapped := false
	for _, i := range indices {
		if swapped {
			chromosomes = append(chromosomes[:i], self.chromosomes[i:]...)
			other_chromosomes = append(other_chromosomes[:i], other.chromosomes[i:]...)
		} else {
			chromosomes = append(chromosomes[:i], other.chromosomes[i:]...)
			other_chromosomes = append(other_chromosomes[:i], self.chromosomes[i:]...)
		}
		swapped = !swapped
	}

	if !options.recombine_chromosomes.ok() || options.recombine_chromosomes.val {
		for i := 0; i < min_size; i++ {
			if (options.match_chromosomes.ok() &&
				!options.match_chromosomes.val) ||
				chromosomes[i].Name == other_chromosomes[i].Name {
				gene, err := chromosomes[i].Recombine(other_chromosomes[i], []int{}, options)
				if err != nil {
					return another, err
				}
				chromosomes[i] = gene
			}
		}
	}

	another.chromosomes = chromosomes
	return another, nil
}

func (self *Genome[T]) ToMap() map[string][]map[string][]map[string][]map[string][]T {
	self.mu.RLock()
	defer self.mu.RUnlock()
	serialized := make(map[string][]map[string][]map[string][]map[string][]T)
	serialized[self.Name] = []map[string][]map[string][]map[string][]T{}
	for _, chromosome := range self.chromosomes {
		serialized[self.Name] = append(serialized[self.Name], chromosome.ToMap())
	}
	return serialized
}

func (self *Genome[T]) Sequence(separator []T) []T {
	self.mu.RLock()
	defer self.mu.RUnlock()
	sequence := make([]T, 0)
	parts := make([][]T, 0)

	for _, chromosome := range self.chromosomes {
		parts = append(parts, chromosome.Sequence(separator))
	}

	for i, part := range parts {
		if i == 0 {
			sequence = append(sequence, part...)
		} else {
			sequence = append(sequence, separator...)
			sequence = append(sequence, separator...)
			sequence = append(sequence, separator...)
			sequence = append(sequence, part...)
		}
	}

	return sequence
}
