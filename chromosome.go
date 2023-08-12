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

func (c *Chromosome[T]) Insert(index int, nucleosome *Nucleosome[T]) error {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	if 0 > index || index > len(c.Nucleosomes) {
		return indexError{}
	}
	if len(c.Nucleosomes) == 0 {
		c.Nucleosomes = append(c.Nucleosomes[:], nucleosome)
	} else {
		c.Nucleosomes = append(c.Nucleosomes[:index+1], c.Nucleosomes[index:]...)
	}
	c.Nucleosomes[index] = nucleosome
	return nil
}

func (c *Chromosome[T]) Append(nucleosome *Nucleosome[T]) error {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	c.Nucleosomes = append(c.Nucleosomes[:], nucleosome)
	return nil
}

func (c *Chromosome[T]) Duplicate(index int) error {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	if 0 > index || index >= len(c.Nucleosomes) {
		return indexError{}
	}
	nucleosome := c.Nucleosomes[index].Copy()
	nucleosomes := append(c.Nucleosomes[:index], nucleosome)
	c.Nucleosomes = append(nucleosomes, c.Nucleosomes[index:]...)
	return nil
}

func (c *Chromosome[T]) Delete(index int) error {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	if 0 > index || index >= len(c.Nucleosomes) {
		return indexError{}
	}
	c.Nucleosomes = append(c.Nucleosomes[:index], c.Nucleosomes[index+1:]...)
	return nil
}

func (c *Chromosome[T]) Substitute(index int, nucleosome *Nucleosome[T]) error {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	if 0 > index || index >= len(c.Nucleosomes) {
		return indexError{}
	}
	nucleosomes := append(c.Nucleosomes[:index], nucleosome)
	c.Nucleosomes = append(nucleosomes, c.Nucleosomes[index+1:]...)
	return nil
}

func (c *Chromosome[T]) Recombine(other *Chromosome[T], indices []int,
	child *Chromosome[T], options RecombineOptions) error {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	other.Mu.RLock()
	defer other.Mu.RUnlock()
	min_size, _ := min(len(c.Nucleosomes), len(other.Nucleosomes))
	max_size, _ := max(len(c.Nucleosomes), len(other.Nucleosomes))

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

	name := c.Name
	if name != other.Name {
		name_size, err := min(len(name), len(other.Name))
		if err != nil {
			return err
		}
		if name_size > 2 {
			name_swap := RandomInt(1, name_size-1)
			name = c.Name[:name_swap] + other.Name[name_swap:]
		}
	}
	child.Name = name

	for len(child.Nucleosomes) < max_size {
		child.Nucleosomes = append(child.Nucleosomes, &Nucleosome[T]{})
	}

	nucleosomes := make([]*Nucleosome[T], max_size)
	other_nucleosomes := make([]*Nucleosome[T], max_size)
	copy(nucleosomes, c.Nucleosomes)
	copy(other_nucleosomes, other.Nucleosomes)
	swapped := false
	for _, i := range indices {
		if swapped {
			nucleosomes = append(nucleosomes[:i], c.Nucleosomes[i:]...)
			other_nucleosomes = append(other_nucleosomes[:i], other.Nucleosomes[i:]...)
		} else {
			nucleosomes = append(nucleosomes[:i], other.Nucleosomes[i:]...)
			other_nucleosomes = append(other_nucleosomes[:i], c.Nucleosomes[i:]...)
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

func (c *Chromosome[T]) ToMap() map[string][]map[string][]map[string][]T {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	serialized := make(map[string][]map[string][]map[string][]T)
	serialized[c.Name] = []map[string][]map[string][]T{}
	for _, nucleosome := range c.Nucleosomes {
		serialized[c.Name] = append(serialized[c.Name], nucleosome.ToMap())
	}
	return serialized
}

func (c *Chromosome[T]) Sequence(separator []T, placeholder ...[]T) []T {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	var realPlaceholder []T
	if len(placeholder) > 0 {
		realPlaceholder = placeholder[0]
	} else {
		realPlaceholder = inverseSequence(separator)
	}
	realPlaceholder = append(realPlaceholder, realPlaceholder...)
	sequence := make([]T, 0)
	parts := make([][]T, 0)

	for _, nucleosome := range c.Nucleosomes {
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
