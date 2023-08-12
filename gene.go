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

func (g *Gene[T]) Copy() *Gene[T] {
	g.Mu.RLock()
	defer g.Mu.RUnlock()
	var another Gene[T]
	another.Name = g.Name
	another.Bases = make([]T, len(g.Bases))
	copy(another.Bases, g.Bases)
	return &another
}

func (g *Gene[T]) Insert(index int, base T) error {
	g.Mu.Lock()
	defer g.Mu.Unlock()
	if 0 > index || index > len(g.Bases) {
		return indexError{}
	}
	if len(g.Bases) == 0 {
		g.Bases = append(g.Bases[:], base)
	} else {
		g.Bases = append(g.Bases[:index+1], g.Bases[index:]...)
	}
	g.Bases[index] = base
	return nil
}

func (g *Gene[T]) Append(base T) error {
	g.Mu.Lock()
	defer g.Mu.Unlock()
	g.Bases = append(g.Bases[:], base)
	return nil
}

func (g *Gene[T]) InsertSequence(index int, sequence []T) error {
	g.Mu.Lock()
	defer g.Mu.Unlock()
	if 0 > index || index > len(g.Bases) {
		return indexError{}
	}
	bases := append(g.Bases[:index], sequence...)
	g.Bases = append(bases, g.Bases[index:]...)
	return nil
}

func (g *Gene[T]) Duplicate(index int) error {
	g.Mu.Lock()
	defer g.Mu.Unlock()
	if 0 > index || index >= len(g.Bases) {
		return indexError{}
	}
	Base := g.Bases[index]
	Bases := append(g.Bases[:index], Base)
	g.Bases = append(Bases, g.Bases[index:]...)
	return nil
}

func (g *Gene[T]) Delete(index int) error {
	g.Mu.Lock()
	defer g.Mu.Unlock()
	if 0 > index || index >= len(g.Bases) {
		return indexError{}
	}
	g.Bases = append(g.Bases[:index], g.Bases[index+1:]...)
	return nil
}

func (g *Gene[T]) DeleteSequence(index int, size int) error {
	g.Mu.Lock()
	defer g.Mu.Unlock()
	if 0 > index || index >= len(g.Bases) {
		return indexError{}
	}
	if size == 0 {
		return anError{"size Must be >0"}
	}
	g.Bases = append(g.Bases[:index], g.Bases[index+size:]...)
	return nil
}

func (g *Gene[T]) Substitute(index int, base T) error {
	g.Mu.Lock()
	defer g.Mu.Unlock()
	if 0 > index || index >= len(g.Bases) {
		return indexError{}
	}
	bases := append(g.Bases[:index], base)
	g.Bases = append(bases, g.Bases[index+1:]...)
	return nil
}

func (g *Gene[T]) Recombine(other *Gene[T], indices []int, child *Gene[T],
	options RecombineOptions) error {
	g.Mu.RLock()
	defer g.Mu.RUnlock()
	other.Mu.RLock()
	defer other.Mu.RUnlock()
	min_size, _ := min(len(g.Bases), len(other.Bases))
	max_size, _ := max(len(g.Bases), len(other.Bases))

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

	bases := make([]T, max_size)
	copy(bases, g.Bases)
	swapped := false
	for _, i := range indices {
		if swapped {
			bases = append(bases[:i], g.Bases[i:]...)
		} else {
			bases = append(bases[:i], other.Bases[i:]...)
		}
		swapped = !swapped
	}
	child.Bases = bases

	return nil
}

func (g *Gene[T]) ToMap() map[string][]T {
	g.Mu.RLock()
	defer g.Mu.RUnlock()
	serialized := make(map[string][]T)
	serialized[g.Name] = g.Bases
	return serialized
}

func (g *Gene[T]) Sequence(placeholder ...[]T) []T {
	g.Mu.RLock()
	defer g.Mu.RUnlock()
	if len(placeholder) > 0 && len(g.Bases) == 0 {
		return placeholder[0]
	}
	return g.Bases
}
