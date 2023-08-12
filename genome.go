package bluegenes

import (
	"math"
	"sort"
	"sync"
)

type Genome[T Ordered] struct {
	Name        string
	Chromosomes []*Chromosome[T]
	Mu          sync.RWMutex
}

func (g *Genome[T]) Copy() *Genome[T] {
	g.Mu.RLock()
	defer g.Mu.RUnlock()
	var another Genome[T]
	another.Name = g.Name
	another.Chromosomes = make([]*Chromosome[T], len(g.Chromosomes))
	copy(another.Chromosomes, g.Chromosomes)
	return &another
}

func (g *Genome[T]) Insert(index int, chromosome *Chromosome[T]) error {
	g.Mu.Lock()
	defer g.Mu.Unlock()
	if 0 > index || index > len(g.Chromosomes) {
		return indexError{}
	}
	if len(g.Chromosomes) == 0 {
		g.Chromosomes = append(g.Chromosomes[:], chromosome)
	} else {
		g.Chromosomes = append(g.Chromosomes[:index+1], g.Chromosomes[index:]...)
	}
	g.Chromosomes[index] = chromosome
	return nil
}

func (g *Genome[T]) Append(chromosome *Chromosome[T]) error {
	g.Mu.Lock()
	defer g.Mu.Unlock()
	g.Chromosomes = append(g.Chromosomes[:], chromosome)
	return nil
}

func (g *Genome[T]) Duplicate(index int) error {
	g.Mu.Lock()
	defer g.Mu.Unlock()
	if 0 > index || index >= len(g.Chromosomes) {
		return indexError{}
	}
	chromosome := g.Chromosomes[index].Copy()
	chromosomes := append(g.Chromosomes[:index], chromosome)
	g.Chromosomes = append(chromosomes, g.Chromosomes[index:]...)
	return nil
}

func (g *Genome[T]) Delete(index int) error {
	g.Mu.Lock()
	defer g.Mu.Unlock()
	if 0 > index || index >= len(g.Chromosomes) {
		return indexError{}
	}
	g.Chromosomes = append(g.Chromosomes[:index], g.Chromosomes[index+1:]...)
	return nil
}

func (g *Genome[T]) Substitute(index int, chromosome *Chromosome[T]) error {
	g.Mu.Lock()
	defer g.Mu.Unlock()
	if 0 > index || index >= len(g.Chromosomes) {
		return indexError{}
	}
	chromosomes := append(g.Chromosomes[:index], chromosome)
	g.Chromosomes = append(chromosomes, g.Chromosomes[index+1:]...)
	return nil
}

func (g *Genome[T]) Recombine(other *Genome[T], indices []int,
	child *Genome[T], options RecombineOptions) error {
	g.Mu.RLock()
	defer g.Mu.RUnlock()
	other.Mu.RLock()
	defer other.Mu.RUnlock()
	min_size, _ := min(len(g.Chromosomes), len(other.Chromosomes))
	max_size, _ := max(len(g.Chromosomes), len(other.Chromosomes))

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
	name := g.Name
	if name != other.Name {
		name_size, err := min(len(name), len(other.Name))
		if err != nil {
			return err
		}
		if name_size > 2 {
			name_swap := RandomInt(1, name_size-1)
			name = g.Name[:name_swap] + other.Name[name_swap:]
		}
	}
	child.Name = name

	for len(child.Chromosomes) < max_size {
		child.Chromosomes = append(child.Chromosomes, &Chromosome[T]{})
	}

	chromosomes := make([]*Chromosome[T], max_size)
	other_chromosomes := make([]*Chromosome[T], max_size)
	copy(chromosomes, g.Chromosomes)
	copy(other_chromosomes, other.Chromosomes)
	swapped := false
	for _, i := range indices {
		if swapped {
			chromosomes = append(chromosomes[:i], g.Chromosomes[i:]...)
			other_chromosomes = append(other_chromosomes[:i], other.Chromosomes[i:]...)
		} else {
			chromosomes = append(chromosomes[:i], other.Chromosomes[i:]...)
			other_chromosomes = append(other_chromosomes[:i], g.Chromosomes[i:]...)
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

func (g *Genome[T]) ToMap() map[string][]map[string][]map[string][]map[string][]T {
	g.Mu.RLock()
	defer g.Mu.RUnlock()
	serialized := make(map[string][]map[string][]map[string][]map[string][]T)
	serialized[g.Name] = []map[string][]map[string][]map[string][]T{}
	for _, chromosome := range g.Chromosomes {
		serialized[g.Name] = append(serialized[g.Name], chromosome.ToMap())
	}
	return serialized
}

func (g *Genome[T]) Sequence(separator []T, placeholder ...[]T) []T {
	g.Mu.RLock()
	defer g.Mu.RUnlock()
	var realPlaceholder []T
	if len(placeholder) > 0 {
		realPlaceholder = placeholder[0]
	} else {
		realPlaceholder = inverseSequence(separator)
	}
	realPlaceholder = append(realPlaceholder, realPlaceholder...)
	realPlaceholder = append(realPlaceholder, realPlaceholder...)
	sequence := make([]T, 0)
	parts := make([][]T, 0)

	for _, chromosome := range g.Chromosomes {
		parts = append(parts, chromosome.Sequence(separator, placeholder...))
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
			sequence = append(sequence, separator...)
			sequence = append(sequence, part...)
		}
	}

	return sequence
}
