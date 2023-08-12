package bluegenes

import (
	"math"
	"sort"
	"sync"
)

type Chromosome[T Ordered] struct {
	Name        string
	Nucleosomes []*Nucleosome[T]
	Mu          sync.RWMutex
}

func (c *Chromosome[T]) Copy() *Chromosome[T] {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	var another Chromosome[T]
	another.Name = c.Name
	another.Nucleosomes = make([]*Nucleosome[T], len(c.Nucleosomes))
	copy(another.Nucleosomes, c.Nucleosomes)
	return &another
}

func (self *Chromosome[T]) Insert(index int, nucleosome *Nucleosome[T]) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	if 0 > index || index > len(self.Nucleosomes) {
		return indexError{}
	}
	if len(self.Nucleosomes) == 0 {
		self.Nucleosomes = append(self.Nucleosomes[:], nucleosome)
	} else {
		self.Nucleosomes = append(self.Nucleosomes[:index+1], self.Nucleosomes[index:]...)
	}
	self.Nucleosomes[index] = nucleosome
	return nil
}

func (self *Chromosome[T]) Append(nucleosome *Nucleosome[T]) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	self.Nucleosomes = append(self.Nucleosomes[:], nucleosome)
	return nil
}

func (self *Chromosome[T]) Duplicate(index int) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	if 0 > index || index >= len(self.Nucleosomes) {
		return indexError{}
	}
	nucleosome := self.Nucleosomes[index].Copy()
	nucleosomes := append(self.Nucleosomes[:index], nucleosome)
	self.Nucleosomes = append(nucleosomes, self.Nucleosomes[index:]...)
	return nil
}

func (self *Chromosome[T]) Delete(index int) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	if 0 > index || index >= len(self.Nucleosomes) {
		return indexError{}
	}
	self.Nucleosomes = append(self.Nucleosomes[:index], self.Nucleosomes[index+1:]...)
	return nil
}

func (self *Chromosome[T]) Substitute(index int, nucleosome *Nucleosome[T]) error {
	self.Mu.Lock()
	defer self.Mu.Unlock()
	if 0 > index || index >= len(self.Nucleosomes) {
		return indexError{}
	}
	nucleosomes := append(self.Nucleosomes[:index], nucleosome)
	self.Nucleosomes = append(nucleosomes, self.Nucleosomes[index+1:]...)
	return nil
}

func (self *Chromosome[T]) Recombine(other *Chromosome[T], indices []int,
	child *Chromosome[T], options RecombineOptions) error {
	self.Mu.RLock()
	defer self.Mu.RUnlock()
	other.Mu.RLock()
	defer other.Mu.RUnlock()
	min_size, _ := min(len(self.Nucleosomes), len(other.Nucleosomes))
	max_size, _ := max(len(self.Nucleosomes), len(other.Nucleosomes))

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

	for len(child.Nucleosomes) < max_size {
		child.Nucleosomes = append(child.Nucleosomes, &Nucleosome[T]{})
	}

	nucleosomes := make([]*Nucleosome[T], max_size)
	other_nucleosomes := make([]*Nucleosome[T], max_size)
	copy(nucleosomes, self.Nucleosomes)
	copy(other_nucleosomes, other.Nucleosomes)
	swapped := false
	for _, i := range indices {
		if swapped {
			nucleosomes = append(nucleosomes[:i], self.Nucleosomes[i:]...)
			other_nucleosomes = append(other_nucleosomes[:i], other.Nucleosomes[i:]...)
		} else {
			nucleosomes = append(nucleosomes[:i], other.Nucleosomes[i:]...)
			other_nucleosomes = append(other_nucleosomes[:i], self.Nucleosomes[i:]...)
		}
		swapped = !swapped
	}

	if !options.RecombineNucleosomes.Ok() || options.RecombineNucleosomes.Val {
		for i := 0; i < min_size; i++ {
			if (options.MatchNucleosomes.Ok() &&
				!options.MatchNucleosomes.Val) ||
				!options.MatchNucleosomes.Ok() ||
				nucleosomes[i].Name == other_nucleosomes[i].Name {
				err := nucleosomes[i].Recombine(other_nucleosomes[i], []int{},
					child.Nucleosomes[i], options)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (self *Chromosome[T]) ToMap() map[string][]map[string][]map[string][]T {
	self.Mu.RLock()
	defer self.Mu.RUnlock()
	serialized := make(map[string][]map[string][]map[string][]T)
	serialized[self.Name] = []map[string][]map[string][]T{}
	for _, nucleosome := range self.Nucleosomes {
		serialized[self.Name] = append(serialized[self.Name], nucleosome.ToMap())
	}
	return serialized
}

func (self *Chromosome[T]) Sequence(separator []T, placeholder ...[]T) []T {
	self.Mu.RLock()
	defer self.Mu.RUnlock()
	var realPlaceholder []T
	if len(placeholder) > 0 {
		realPlaceholder = placeholder[0]
	} else {
		realPlaceholder = inverseSequence(separator)
	}
	realPlaceholder = append(realPlaceholder, realPlaceholder...)
	sequence := make([]T, 0)
	parts := make([][]T, 0)

	for _, nucleosome := range self.Nucleosomes {
		parts = append(parts, nucleosome.Sequence(separator, placeholder...))
	}

	for i, part := range parts {
		if len(part) == 0 {
			part = realPlaceholder
		}
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
