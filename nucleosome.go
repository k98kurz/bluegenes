package bluegenes

import (
	"math"
	"sort"
	"sync"
)

type Nucleosome[T Ordered] struct {
	Name  string
	Genes []*Gene[T]
	Mu    sync.RWMutex
}

func (n *Nucleosome[T]) Copy() *Nucleosome[T] {
	n.Mu.RLock()
	defer n.Mu.RUnlock()
	var another Nucleosome[T]
	another.Name = n.Name
	another.Genes = make([]*Gene[T], len(n.Genes))
	copy(another.Genes, n.Genes)
	return &another
}

func (n *Nucleosome[T]) Insert(index int, gene *Gene[T]) error {
	n.Mu.Lock()
	defer n.Mu.Unlock()
	if 0 > index || index > len(n.Genes) {
		return indexError{}
	}
	if len(n.Genes) == 0 {
		n.Genes = append(n.Genes[:], gene)
	} else {
		n.Genes = append(n.Genes[:index+1], n.Genes[index:]...)
	}
	n.Genes[index] = gene
	return nil
}

func (n *Nucleosome[T]) Append(gene *Gene[T]) error {
	n.Mu.Lock()
	defer n.Mu.Unlock()
	n.Genes = append(n.Genes[:], gene)
	return nil
}

func (n *Nucleosome[T]) Duplicate(index int) error {
	n.Mu.Lock()
	defer n.Mu.Unlock()
	if 0 > index || index >= len(n.Genes) {
		return indexError{}
	}
	Gene := n.Genes[index].Copy()
	Genes := append(n.Genes[:index], Gene)
	n.Genes = append(Genes, n.Genes[index:]...)
	return nil
}

func (n *Nucleosome[T]) Delete(index int) error {
	n.Mu.Lock()
	defer n.Mu.Unlock()
	if 0 > index || index >= len(n.Genes) {
		return indexError{}
	}
	n.Genes = append(n.Genes[:index], n.Genes[index+1:]...)
	return nil
}

func (n *Nucleosome[T]) Substitute(index int, gene *Gene[T]) error {
	n.Mu.Lock()
	defer n.Mu.Unlock()
	if 0 > index || index >= len(n.Genes) {
		return indexError{}
	}
	genes := append(n.Genes[:index], gene)
	n.Genes = append(genes, n.Genes[index+1:]...)
	return nil
}

func (n *Nucleosome[T]) Recombine(other *Nucleosome[T], indices []int,
	child *Nucleosome[T], options RecombineOptions) error {
	n.Mu.RLock()
	defer n.Mu.RUnlock()
	other.Mu.RLock()
	defer other.Mu.RUnlock()
	min_size, _ := min(len(n.Genes), len(other.Genes))
	max_size, _ := max(len(n.Genes), len(other.Genes))

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

	name := n.Name
	if name != other.Name {
		name_size, err := min(len(name), len(other.Name))
		if err != nil {
			return err
		}
		if name_size > 2 {
			name_swap := RandomInt(1, name_size-1)
			name = n.Name[:name_swap] + other.Name[name_swap:]
		}
	}
	child.Name = name

	for len(child.Genes) < max_size {
		child.Genes = append(child.Genes, &Gene[T]{})
	}

	genes := make([]*Gene[T], max_size)
	other_genes := make([]*Gene[T], max_size)
	copy(genes, n.Genes)
	copy(other_genes, other.Genes)
	swapped := false
	for _, i := range indices {
		if swapped {
			genes = append(genes[:i], n.Genes[i:]...)
			other_genes = append(other_genes[:i], other.Genes[i:]...)
		} else {
			genes = append(genes[:i], other.Genes[i:]...)
			other_genes = append(other_genes[:i], n.Genes[i:]...)
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

func (n *Nucleosome[T]) ToMap() map[string][]map[string][]T {
	n.Mu.RLock()
	defer n.Mu.RUnlock()
	serialized := make(map[string][]map[string][]T)
	serialized[n.Name] = []map[string][]T{}
	for _, gene := range n.Genes {
		serialized[n.Name] = append(serialized[n.Name], gene.ToMap())
	}
	return serialized
}

func (n *Nucleosome[T]) Sequence(separator []T, placeholder ...[]T) []T {
	n.Mu.RLock()
	defer n.Mu.RUnlock()
	var realPlaceholder []T
	if len(placeholder) > 0 {
		realPlaceholder = placeholder[0]
	} else {
		realPlaceholder = inverseSequence(separator)
	}
	sequence := make([]T, 0)
	parts := make([][]T, 0)

	for _, gene := range n.Genes {
		parts = append(parts, gene.Sequence(placeholder...))
	}

	for i, part := range parts {
		if len(part) == 0 {
			part = realPlaceholder
		}
		if i == 0 {
			sequence = append(sequence, part...)
		} else {
			sequence = append(sequence, separator...)
			sequence = append(sequence, part...)
		}
	}

	return sequence
}
