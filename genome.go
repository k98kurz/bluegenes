package bluegenes

import (
	"math"
	"sort"
	"sync"
)

type Genome[T comparable] struct {
	Name        string
	Chromosomes []*Chromosome[T]
	Mu          sync.RWMutex
}

func (self *Genome[T]) Copy() *Genome[T] {
	self.Mu.RLock()
	defer self.Mu.RUnlock()
	var another Genome[T]
	another.Name = self.Name
	another.Chromosomes = make([]*Chromosome[T], len(self.Chromosomes))
	copy(another.Chromosomes, self.Chromosomes)
	return &another
}

func (self *Genome[T]) Insert(index int, chromosome *Chromosome[T]) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	if 0 > index || index > len(self.Chromosomes) {
		return indexError{}
	}
	if len(self.Chromosomes) == 0 {
		self.Chromosomes = append(self.Chromosomes[:], chromosome)
	} else {
		self.Chromosomes = append(self.Chromosomes[:index+1], self.Chromosomes[index:]...)
	}
	self.Chromosomes[index] = chromosome
	return nil
}

func (self *Genome[T]) Append(chromosome *Chromosome[T]) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	self.Chromosomes = append(self.Chromosomes[:], chromosome)
	return nil
}

func (self *Genome[T]) Duplicate(index int) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	if 0 > index || index > len(self.Chromosomes) {
		return indexError{}
	}
	chromosome := self.Chromosomes[index].Copy()
	chromosomes := append(self.Chromosomes[:index], chromosome)
	self.Chromosomes = append(chromosomes, self.Chromosomes[index:]...)
	return nil
}

func (self *Genome[T]) Delete(index int) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	if 0 > index || index >= len(self.Chromosomes) {
		return indexError{}
	}
	self.Chromosomes = append(self.Chromosomes[:index], self.Chromosomes[index+1:]...)
	return nil
}

func (self *Genome[T]) Substitute(index int, chromosome *Chromosome[T]) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	if 0 > index || index >= len(self.Chromosomes) {
		return indexError{}
	}
	chromosomes := append(self.Chromosomes[:index], chromosome)
	self.Chromosomes = append(chromosomes, self.Chromosomes[index+1:]...)
	return nil
}

func (self *Genome[T]) Recombine(other *Genome[T], indices []int,
	child *Genome[T], options RecombineOptions) error {
	self.Mu.RLock()
	defer self.Mu.RUnlock()
	other.Mu.RLock()
	defer other.Mu.RUnlock()
	min_size, _ := min(len(self.Chromosomes), len(other.Chromosomes))
	max_size, _ := max(len(self.Chromosomes), len(other.Chromosomes))

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

	for len(child.Chromosomes) < max_size {
		child.Chromosomes = append(child.Chromosomes, &Chromosome[T]{})
	}

	chromosomes := make([]*Chromosome[T], max_size)
	other_chromosomes := make([]*Chromosome[T], max_size)
	copy(chromosomes, self.Chromosomes)
	copy(other_chromosomes, other.Chromosomes)
	swapped := false
	for _, i := range indices {
		if swapped {
			chromosomes = append(chromosomes[:i], self.Chromosomes[i:]...)
			other_chromosomes = append(other_chromosomes[:i], other.Chromosomes[i:]...)
		} else {
			chromosomes = append(chromosomes[:i], other.Chromosomes[i:]...)
			other_chromosomes = append(other_chromosomes[:i], self.Chromosomes[i:]...)
		}
		swapped = !swapped
	}

	if !options.RecombineChromosomes.Ok() || options.RecombineChromosomes.Val {
		for i := 0; i < min_size; i++ {
			if (options.MatchChromosomes.Ok() &&
				!options.MatchChromosomes.Val) ||
				!options.MatchChromosomes.Ok() ||
				chromosomes[i].Name == other_chromosomes[i].Name {
				err := chromosomes[i].Recombine(other_chromosomes[i], []int{},
					child.Chromosomes[i], options)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (self *Genome[T]) ToMap() map[string][]map[string][]map[string][]map[string][]T {
	self.Mu.RLock()
	defer self.Mu.RUnlock()
	serialized := make(map[string][]map[string][]map[string][]map[string][]T)
	serialized[self.Name] = []map[string][]map[string][]map[string][]T{}
	for _, chromosome := range self.Chromosomes {
		serialized[self.Name] = append(serialized[self.Name], chromosome.ToMap())
	}
	return serialized
}

func (self *Genome[T]) Sequence(separator []T) []T {
	self.Mu.RLock()
	defer self.Mu.RUnlock()
	sequence := make([]T, 0)
	parts := make([][]T, 0)

	for _, chromosome := range self.Chromosomes {
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
