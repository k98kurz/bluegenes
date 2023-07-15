package genetics

import (
	"fmt"
	"testing"
)

func firstGene() *Gene[int] {
	return &Gene[int]{
		name:  "test",
		bases: []int{1, 2, 3},
	}
}

func secondGene() *Gene[int] {
	return &Gene[int]{
		name:  "tset",
		bases: []int{4, 5, 6},
	}
}

func thirdGene() *Gene[int] {
	return &Gene[int]{
		name:  "gen3",
		bases: []int{7, 8, 9},
	}
}

func fourthGene() *Gene[int] {
	return &Gene[int]{
		name:  "gen4",
		bases: []int{10, 11, 12},
	}
}

func firstAllele() *Allele[int] {
	return &Allele[int]{
		name: "al1",
		genes: []*Gene[int]{
			firstGene(),
			secondGene(),
		},
	}
}

func secondAllele() *Allele[int] {
	return &Allele[int]{
		name: "al2",
		genes: []*Gene[int]{
			thirdGene(),
			fourthGene(),
		},
	}
}

func firstChromosome() *Chromosome[int] {
	return &Chromosome[int]{
		name: "c1",
		alleles: []*Allele[int]{
			firstAllele(),
			secondAllele(),
		},
	}
}

func secondChromosome() *Chromosome[int] {
	return &Chromosome[int]{
		name: "c1",
		alleles: []*Allele[int]{
			secondAllele(),
			firstAllele(),
		},
	}
}

func firstGenome() *Genome[int] {
	return &Genome[int]{
		name: "Genome",
		chromosomes: []*Chromosome[int]{
			firstChromosome(),
			secondChromosome(),
		},
	}
}

func rangeGene(start int, stop int, name ...string) (*Gene[int], error) {
	g := &Gene[int]{name: "GnR"}
	if start >= stop {
		return g, Error{"start must be <= stop"}
	}
	for i := start; i <= stop; i++ {
		g.Append(i)
	}
	if len(name) > 0 {
		g.name = name[0]
	}
	return g, nil
}

func rangeAllele(size int, start int, stop int, name ...string) (*Allele[int], error) {
	a := &Allele[int]{name: "AlR"}

	if size <= 0 {
		return a, Error{"size must be > 0"}
	}

	for i := 0; i < size; i++ {
		g, err := rangeGene(start, stop, fmt.Sprintf("GnR%d", i))
		if err != nil {
			return a, err
		}
		a.Append(g)
	}

	if len(name) > 0 {
		a.name = name[0]
	}

	return a, nil
}

func rangeChromosome(size int, allele_size int, start int, stop int, name ...string) (*Chromosome[int], error) {
	c := &Chromosome[int]{name: "ChR"}

	if size <= 0 {
		return c, Error{"size must be > 0"}
	}

	if allele_size <= 0 {
		return c, Error{"allele_size must be > 0"}
	}

	for i := 0; i < size; i++ {
		g, err := rangeAllele(allele_size, start, stop, fmt.Sprintf("AlR%d", i))
		if err != nil {
			return c, err
		}
		c.Append(g)
	}

	if len(name) > 0 {
		c.name = name[0]
	}

	return c, nil
}

func rangeGenome(size int, chromosome_size int, allele_size int, start int, stop int, name ...string) (*Genome[int], error) {
	g := &Genome[int]{name: "GenomR"}

	if size <= 0 {
		return g, Error{"size must be > 0"}
	}

	if allele_size <= 0 {
		return g, Error{"allele_size must be > 0"}
	}

	if chromosome_size <= 0 {
		return g, Error{"chromosome_size must be > 0"}
	}

	for i := 0; i < size; i++ {
		c, err := rangeChromosome(chromosome_size, allele_size, start, stop, fmt.Sprintf("AlR%d", i))
		if err != nil {
			return g, err
		}
		g.Append(c)
	}

	if len(name) > 0 {
		g.name = name[0]
	}

	return g, nil
}

func factory() int {
	return RandomInt(0, 10)
}

func TestGene(t *testing.T) {
	t.Run("Copy", func(t *testing.T) {
		t.Parallel()
		g := firstGene()
		c := g.Copy()

		if c == g {
			t.Error("Gene[int].Copy failed; received pointer to same memory")
		} else if c.name != g.name {
			t.Errorf("Gene[int].Copy failed to copy name; got %s, expected %s", c.name, g.name)
		} else if len(c.bases) != len(g.bases) {
			t.Fatal("Gene[int].Copy failed to copy bases")
		}

		for i, item := range c.bases {
			if g.bases[i] != item {
				t.Errorf("Gene[int].Copy failed to copy bases: got %d, expected %d", item, g.bases[i])
			}
		}
	})

	t.Run("Insert", func(t *testing.T) {
		t.Parallel()
		g := firstGene()

		g.Insert(1, 15)
		expected := []int{1, 15, 2, 3}
		for i, item := range g.bases {
			if item != expected[i] {
				t.Errorf("Gene[int].Insert produced invalid result at index %d: expected %d, got %d", i, expected[i], item)
			}
		}
	})

	t.Run("Append", func(t *testing.T) {
		t.Parallel()
		g := firstGene()
		for i := 1; i < 1111; i++ {
			err := g.Append(i)
			if err != nil {
				t.Errorf("Gene[int].Append failed with error: %v", err.Error())
			}
			observed := g.bases[len(g.bases)-1]
			if observed != i {
				t.Errorf("Gene[int].Append did not add to end; expected %d, observed %d", i, observed)
			}
		}
	})

	t.Run("InsertSequence", func(t *testing.T) {
		t.Parallel()
		for i := 3; i < 111; i++ {
			g := firstGene()
			seq := []int{}
			for k := i; k > 0; k-- {
				seq = append(seq, k)
			}
			err := g.InsertSequence(0, seq)
			if err != nil {
				t.Errorf("Gene[int].InsertSequence failed with error: %v", err.Error())
			}
			observed := g.bases[:len(seq)]
			for k, item := range observed {
				if item != seq[k] {
					t.Fatalf("Gene[int].InsertSequence failed at index %d: expected %d, observed %d", k, seq[k], item)
				}
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		t.Parallel()
		for i := 1; i < 111; i++ {
			g := firstGene()
			expected := []int{1, 3}
			err := g.Delete(1)
			if err != nil {
				t.Errorf("Gene[int].Delete failed with error: %v", err.Error())
			}
			for k, item := range g.bases {
				if item != expected[k] {
					t.Fatalf("Gene[int].Delete failed at index %d: expected %d, observed %d", k, expected[k], item)
				}
			}
		}
	})

	t.Run("DeleteSequence", func(t *testing.T) {
		t.Parallel()
		g, _ := rangeGene(0, 5)
		g.DeleteSequence(0, 2)
		expected := []int{2, 3, 4, 5}
		for i, item := range g.bases {
			if item != expected[i] {
				t.Fatalf("Gene[int].DeleteSequence result fail at index %d: expected %d, observed %d", i, expected[i], item)
			}
		}

		for i := 5; i < 111; i++ {
			for k := 1; k < i; k++ {
				g, _ := rangeGene(1, i)
				err := g.DeleteSequence(0, k)
				if err != nil {
					t.Errorf("Gene[int].DeleteSequence failed with error: %s", err.Error())
					continue
				}
				if len(g.bases) > i-k {
					t.Fatalf("Gene[int].DeleteSequence failed to remove enough items: expected %d len, observed %d", i-k, len(g.bases))
				}
			}
		}
	})

	t.Run("Substitute", func(t *testing.T) {
		t.Parallel()
		g := firstGene()
		g.Substitute(0, 15)
		expected := []int{15, 2, 3}
		if !equal(expected, g.bases) {
			t.Fatalf("Gene[int].Substitute failed: expected [%v], observed [%v]", expected, g.bases)
		}

		g = firstGene()
		g.Substitute(1, 15)
		expected = []int{1, 15, 3}
		if !equal(expected, g.bases) {
			t.Fatalf("Gene[int].Substitute failed: expected [%v], observed [%v]", expected, g.bases)
		}

		g = firstGene()
		g.Substitute(2, 15)
		expected = []int{1, 2, 15}
		if !equal(expected, g.bases) {
			t.Fatalf("Gene[int].Substitute failed: expected [%v], observed [%v]", expected, g.bases)
		}
	})

	t.Run("Recombine", func(t *testing.T) {
		t.Parallel()
		for i := 0; i < 111; i++ {
			g1, _ := rangeGene(0, 5, "dad")
			g2, _ := rangeGene(6, 11, "mom")
			g3, err := g1.Recombine(g2, []int{}, RecombineOptions{})
			if err != nil {
				t.Fatalf("Gene[int].Recombine failed with error: %s", err.Error())
			}
			parents := NewSet[string]()
			for _, item := range g3.bases {
				if contains(g1.bases, item) {
					parents.Add(g1.name)
				} else if contains(g2.bases, item) {
					parents.Add(g2.name)
				} else {
					t.Fatalf("encountered item not from parents: %d", item)
				}
			}
		}
	})

	t.Run("ToMap", func(t *testing.T) {
		t.Parallel()
		g := firstGene()
		expected := make(map[string][]int)
		expected["test"] = []int{1, 2, 3}
		observed := g.ToMap()

		for k, v := range expected {
			v2, ok := observed[k]
			if !ok || !equal(v, v2) {
				t.Fatalf("Gene[int].ToMap failed: expected %v, observed %v", expected, observed)
			}
		}
		for k, v := range observed {
			v2, ok := expected[k]
			if !ok || !equal(v, v2) {
				t.Fatalf("Gene[int].ToMap failed: expected %v, observed %v", expected, observed)
			}
		}

		g = secondGene()
		expected = make(map[string][]int)
		expected["tset"] = []int{4, 5, 6}
		observed = g.ToMap()

		for k, v := range expected {
			v2, ok := observed[k]
			if !ok || !equal(v, v2) {
				t.Fatalf("Gene[int].ToMap failed: expected %v, observed %v", expected, observed)
			}
		}
		for k, v := range observed {
			v2, ok := expected[k]
			if !ok || !equal(v, v2) {
				t.Fatalf("Gene[int].ToMap failed: expected %v, observed %v", expected, observed)
			}
		}
	})

	t.Run("MakeGene", func(t *testing.T) {
		t.Parallel()
		names := NewSet[string]()
		sequences := [][]int{}

		for i := 0; i < 10; i++ {
			g, err := MakeGene[int](MakeOptions[int]{
				n_bases:      NewOption(uint(5)),
				base_factory: NewOption(factory),
			})
			if err != nil {
				t.Fatalf("MakeGene[int] failed with error: %v", err)
			}
			names.Add(g.name)

			if !containsSlice(sequences, g.bases) {
				sequences = append(sequences, g.bases)
			}
		}

		if names.Len() < 8 {
			t.Fatalf("MakeGene[int] failed to produce enough random names: expected >= 8, observed %d", names.Len())
		}

		if len(sequences) < 8 {
			t.Fatalf("MakeGene[int] failed to produce enough random sequences: expected >= 8, observed %d", len(sequences))
		}
	})

	t.Run("Sequence", func(t *testing.T) {
		t.Parallel()
		gene := firstGene()
		sequence := gene.Sequence()
		unpacked := GeneFromSequence(sequence)

		if !equal(gene.bases, unpacked.bases) {
			t.Errorf("GeneFromSequence[int] failed: expected %v, observed %v", gene.bases, unpacked.bases)
		}
	})
}

func TestAllele(t *testing.T) {
	t.Run("Copy", func(t *testing.T) {
		t.Parallel()
		a := firstAllele()
		c := a.Copy()

		if c == a {
			t.Error("Allele[int].Copy failed; received pointer to same memory")
		} else if c.name != a.name {
			t.Errorf("Allele[int].Copy failed to copy name; got %s, expected %s", c.name, a.name)
		} else if len(c.genes) != len(a.genes) {
			t.Fatal("Allele[int].Copy failed to copy genes")
		}

		for i, item := range c.genes {
			if a.genes[i] != item {
				t.Errorf("Allele[int].Copy failed to copy genes: got %v, expected %v", item, a.genes[i])
			}
		}
	})

	t.Run("Insert", func(t *testing.T) {
		t.Parallel()
		a := firstAllele()
		g, _ := rangeGene(0, 5, "range")
		expected_names := NewSet[string]()

		for _, g := range a.genes {
			expected_names.Add(g.name)
		}
		expected_names.Add("range")

		a.Insert(1, g)
		observed_names := NewSet[string]()
		for _, g = range a.genes {
			observed_names.Add(g.name)
		}

		if !expected_names.Equal(observed_names) {
			t.Errorf("Allele[int].Insert failed: expected names %v, observed %v", expected_names, observed_names)
		}
	})

	t.Run("Append", func(t *testing.T) {
		t.Parallel()
		a := firstAllele()
		for i := 1; i < 11; i++ {
			g, err := rangeGene(0, i)
			if err != nil {
				t.Errorf("Allele[int].Append failed with error: %v", err.Error())
			}
			err = a.Append(g)
			if err != nil {
				t.Errorf("Allele[int].Append failed with error: %v", err.Error())
			}
			observed := a.genes[len(a.genes)-1]
			if observed.name != g.name {
				t.Errorf("Allele[int].Append did not add to end; expected %v, observed %v", g, observed)
			}
		}
	})

	t.Run("Duplicate", func(t *testing.T) {
		t.Parallel()
		a := firstAllele()
		err := a.Duplicate(0)
		if err != nil {
			t.Errorf("Allele[int].Duplicate failed with error: %v", err.Error())
		}
		first := a.genes[0]
		second := a.genes[1]
		if first.name != second.name {
			t.Fatalf("Allele[int].Duplicate failed to duplicate Gene: expected %v, observed %v", first, second)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		t.Parallel()
		for i := 1; i < 111; i++ {
			a := firstAllele()
			expected_size := len(a.genes) - 1
			err := a.Delete(0)
			if err != nil {
				t.Errorf("Allele[int].Delete failed with error: %v", err.Error())
			}
			if expected_size != len(a.genes) {
				t.Fatalf("Allele[int].Delete failed to delete Gene: expected %v, observed %v", expected_size, len(a.genes))
			}
		}
	})

	t.Run("Substitute", func(t *testing.T) {
		t.Parallel()
		a, _ := rangeAllele(3, 0, 5)
		g, _ := rangeGene(2, 4)
		a.Substitute(0, g)
		expected := g.bases
		if !equal(expected, a.genes[0].bases) {
			t.Fatalf("Allele[int].Substitute failed: expected [%v], observed [%v]", expected, a.genes[0].bases)
		}

		a.Substitute(1, g)
		if !equal(expected, a.genes[1].bases) {
			t.Fatalf("Allele[int].Substitute failed: expected [%v], observed [%v]", expected, a.genes[1].bases)
		}

		a.Substitute(2, g)
		if !equal(expected, a.genes[2].bases) {
			t.Fatalf("Allele[int].Substitute failed: expected [%v], observed [%v]", expected, a.genes[2].bases)
		}
	})

	t.Run("Recombine", func(t *testing.T) {
		t.Parallel()
		dad, _ := rangeAllele(3, 0, 5, "dad")
		mom, _ := rangeAllele(3, 6, 11, "mom")
		child, err := dad.Recombine(mom, []int{}, RecombineOptions{})
		if err != nil {
			t.Fatalf("Allele[int].Recombine failed with error: %s", err.Error())
		}
		parents := NewSet[string]()
		dad_bases := NewSet[int]()
		mom_bases := NewSet[int]()
		for _, gene := range dad.genes {
			for _, base := range gene.bases {
				dad_bases.Add(base)
			}
		}
		for _, gene := range mom.genes {
			for _, base := range gene.bases {
				mom_bases.Add(base)
			}
		}
		for _, gene := range child.genes {
			for _, base := range gene.bases {
				if dad_bases.Contains(base) {
					parents.Add(dad.name)
				} else if mom_bases.Contains(base) {
					parents.Add(mom.name)
				} else {
					t.Fatalf("Allele[int].Recombine failed: encountered base not from parents: %v", base)
				}
			}
		}
		if parents.Len() < 2 {
			fmt.Printf("mom: %v\n", mom.ToMap())
			fmt.Printf("dad: %v\n", dad.ToMap())
			fmt.Printf("child: %v\n", child.ToMap())
			t.Fatalf("Allele[int].Recombine failed: expected bases from 2 parents, observed %d", parents.Len())
		}
	})

	t.Run("MakeAllele", func(t *testing.T) {
		t.Parallel()
		names := NewSet[string]()
		maps := make(map[string]map[string][]map[string][]int)

		for i := 0; i < 10; i++ {
			a, err := MakeAllele[int](MakeOptions[int]{
				n_genes:      NewOption(uint(3)),
				n_bases:      NewOption(uint(5)),
				base_factory: NewOption(factory),
			})
			if err != nil {
				t.Fatalf("MakeAllele[int] failed with error: %v", err)
			}
			names.Add(a.name)

			_, ok := maps[a.name]
			if !ok {
				maps[a.name] = a.ToMap()
			}
		}

		if names.Len() < 8 {
			t.Fatalf("MakeAllele[int] failed to produce enough random names: expected >= 8, observed %d", names.Len())
		}

		if len(maps) < 8 {
			t.Fatalf("MakeAllele[int] failed to produce enough random sequences: expected >= 8, observed %d", len(maps))
		}
	})

	t.Run("Sequence", func(t *testing.T) {
		t.Parallel()
		allele := firstAllele()
		separator := []int{0, 0, 0, 0, 0}
		sequence := allele.Sequence(separator)
		unpacked := AlleleFromSequence(sequence, separator)

		if len(unpacked.genes) != len(allele.genes) {
			t.Errorf("Allele[int].Sequence -> AlleleFromSequence failed: expected %d genes, observed %d", len(allele.genes), len(unpacked.genes))
		}

		repacked := unpacked.Sequence(separator)
		if !equal(sequence, repacked) {
			t.Errorf("Allele[int].Sequence -> AlleleFromSequence -> .Sequence failed: expected %v, observed %v", sequence, repacked)
		}
	})
}

func TestChromosome(t *testing.T) {
	t.Run("Copy", func(t *testing.T) {
		t.Parallel()
		c := firstChromosome()
		p := c.Copy()

		if p == c {
			t.Error("Chromosome[int].Copy failed; received pointer to same memory")
		} else if p.name != c.name {
			t.Errorf("Chromosome[int].Copy failed to copy name; got %s, expected %s", p.name, c.name)
		} else if len(p.alleles) != len(c.alleles) {
			t.Fatal("Chromosome[int].Copy failed to copy alleles")
		}

		for i, allele := range p.alleles {
			if c.alleles[i].name != allele.name {
				t.Errorf("Chromosome[int].Copy failed to copy alleles: got %v, expected %v", allele.ToMap(), c.alleles[i].ToMap())
				continue
			}
			for k, gene := range allele.genes {
				if gene.name != c.alleles[i].genes[k].name || !equal(gene.bases, c.alleles[i].genes[k].bases) {
					t.Errorf("Chromosome[int].Copy failed to copy genes: got %v, expected %v", gene.ToMap(), c.alleles[i].genes[k].ToMap())
					break
				}
			}
		}
	})

	t.Run("Insert", func(t *testing.T) {
		t.Parallel()
		c := firstChromosome()
		a, _ := rangeAllele(2, 0, 5, "range")
		expected_names := NewSet[string]()

		for _, a := range c.alleles {
			expected_names.Add(a.name)
		}
		expected_names.Add("range")

		c.Insert(1, a)
		observed_names := NewSet[string]()
		for _, a = range c.alleles {
			observed_names.Add(a.name)
		}

		if !expected_names.Equal(observed_names) {
			t.Errorf("Chromosome[int].Insert failed: expected names %v, observed %v", expected_names, observed_names)
		}
	})

	t.Run("Append", func(t *testing.T) {
		t.Parallel()
		c := firstChromosome()
		for i := 1; i < 11; i++ {
			a, err := rangeAllele(i, 0, 5)
			if err != nil {
				t.Errorf("Chromosome[int].Append failed with error: %v", err.Error())
			}
			err = c.Append(a)
			if err != nil {
				t.Errorf("Chromosome[int].Append failed with error: %v", err.Error())
			}
			observed := c.alleles[len(c.alleles)-1]
			if observed.name != a.name {
				t.Errorf("Chromosome[int].Append did not add to end; expected %v, observed %v", a, observed)
			}
		}
	})

	t.Run("Duplicate", func(t *testing.T) {
		t.Parallel()
		c := firstChromosome()
		err := c.Duplicate(0)
		if err != nil {
			t.Errorf("Chromosome[int].Duplicate failed with error: %v", err.Error())
		}
		first := c.alleles[0]
		second := c.alleles[1]
		if first.name != second.name {
			t.Fatalf("Chromosome[int].Duplicate failed to duplicate Gene: expected %v, observed %v", first.ToMap(), second.ToMap())
		}
	})

	t.Run("Delete", func(t *testing.T) {
		t.Parallel()
		for i := 1; i < 111; i++ {
			c := firstChromosome()
			expected_size := len(c.alleles) - 1
			err := c.Delete(0)
			if err != nil {
				t.Errorf("Chromosome[int].Delete failed with error: %v", err.Error())
			}
			if expected_size != len(c.alleles) {
				t.Fatalf("Chromosome[int].Delete failed to delete Gene: expected %v, observed %v", expected_size, len(c.alleles))
			}
		}
	})

	t.Run("Substitute", func(t *testing.T) {
		t.Parallel()
		c, _ := rangeChromosome(2, 3, 0, 5)
		a, _ := rangeAllele(2, 2, 4)
		c.Substitute(0, a)
		expected := a.genes
		if !equal(expected, c.alleles[0].genes) {
			t.Fatalf("Chromosome[int].Substitute failed: expected [%v], observed [%v]", expected, c.alleles[0].genes)
		}

		c.Substitute(1, a)
		if !equal(expected, c.alleles[1].genes) {
			t.Fatalf("Chromosome[int].Substitute failed: expected [%v], observed [%v]", expected, c.alleles[1].genes)
		}
	})

	t.Run("Recombine", func(t *testing.T) {
		t.Parallel()
		dad, _ := rangeChromosome(2, 3, 0, 5, "dad")
		mom, _ := rangeChromosome(2, 3, 6, 11, "mom")
		child, err := dad.Recombine(mom, []int{}, RecombineOptions{})
		if err != nil {
			t.Fatalf("Chromosome[int].Recombine failed with error: %s", err.Error())
		}
		parents := NewSet[string]()
		dad_bases := NewSet[int]()
		mom_bases := NewSet[int]()
		for _, allele := range dad.alleles {
			for _, gene := range allele.genes {
				for _, base := range gene.bases {
					dad_bases.Add(base)
				}
			}
		}
		for _, allele := range mom.alleles {
			for _, gene := range allele.genes {
				for _, base := range gene.bases {
					mom_bases.Add(base)
				}
			}
		}

		for _, allele := range child.alleles {
			for _, gene := range allele.genes {
				for _, base := range gene.bases {
					if dad_bases.Contains(base) {
						parents.Add(dad.name)
					} else if mom_bases.Contains(base) {
						parents.Add(mom.name)
					} else {
						t.Fatalf("Chromosome[int].Recombine failed: encountered base not from parents: %v", base)
					}
				}
			}
		}
		if parents.Len() < 2 {
			fmt.Printf("mom: %v\n", mom.ToMap())
			fmt.Printf("dad: %v\n", dad.ToMap())
			fmt.Printf("child: %v\n", child.ToMap())
			t.Fatalf("Chromosome[int].Recombine failed: expected bases from 2 parents, observed %d", parents.Len())
		}
	})

	t.Run("MakeChromosome", func(t *testing.T) {
		t.Parallel()
		names := NewSet[string]()
		maps := make(map[string]map[string][]map[string][]map[string][]int)

		for i := 0; i < 10; i++ {
			a, err := MakeChromosome[int](MakeOptions[int]{
				n_alleles:    NewOption(uint(3)),
				n_genes:      NewOption(uint(3)),
				n_bases:      NewOption(uint(5)),
				base_factory: NewOption(factory),
			})
			if err != nil {
				t.Fatalf("MakeChromosome[int] failed with error: %v", err)
			}
			names.Add(a.name)

			_, ok := maps[a.name]
			if !ok {
				maps[a.name] = a.ToMap()
			}
		}

		if names.Len() < 8 {
			t.Fatalf("MakeChromosome[int] failed to produce enough random names: expected >= 8, observed %d", names.Len())
		}

		if len(maps) < 8 {
			t.Fatalf("MakeChromosome[int] failed to produce enough random sequences: expected >= 8, observed %d", len(maps))
		}
	})

	t.Run("Sequence", func(t *testing.T) {
		t.Parallel()
		chromosome := firstChromosome()
		separator := []int{0, 0, 0, 0, 0}
		sequence := chromosome.Sequence(separator)
		unpacked := ChromosomeFromSequence(sequence, separator)

		if len(unpacked.alleles) != len(chromosome.alleles) {
			t.Errorf("Chromosome[int].Sequence -> ChromosomeFromSequence failed: expected %d genes, observed %d", len(chromosome.alleles), len(unpacked.alleles))
		}

		repacked := unpacked.Sequence(separator)
		if !equal(sequence, repacked) {
			t.Errorf("Chromosome[int].Sequence -> ChromosomeFromSequence -> .Sequence failed: expected %v, observed %v", sequence, repacked)
		}
	})
}

func TestGenome(t *testing.T) {
	t.Run("Copy", func(t *testing.T) {
		t.Parallel()
		g := firstGenome()
		p := g.Copy()

		if p == g {
			t.Error("Genome[int].Copy failed; received pointer to same memory")
		} else if p.name != g.name {
			t.Errorf("Genome[int].Copy failed to copy name; got %s, expected %s", p.name, g.name)
		} else if len(p.chromosomes) != len(g.chromosomes) {
			t.Fatal("Genome[int].Copy failed to copy chromosomes")
		}

		for i, chromosome := range p.chromosomes {
			if g.chromosomes[i].name != chromosome.name {
				t.Errorf("Genome[int].Copy failed to copy chromosomes: got %v, expected %v", chromosome.ToMap(), g.chromosomes[i].ToMap())
				continue
			}
			for k, allele := range chromosome.alleles {
				for j, gene := range allele.genes {
					if gene.name != g.chromosomes[i].alleles[k].genes[j].name || !equal(gene.bases, g.chromosomes[i].alleles[k].genes[j].bases) {
						t.Errorf("Genome[int].Copy failed to copy genes: got %v, expected %v", gene.ToMap(), g.chromosomes[i].alleles[k].genes[j].ToMap())
						break
					}
				}
			}
		}
	})

	t.Run("Insert", func(t *testing.T) {
		t.Parallel()
		g := firstGenome()
		c, _ := rangeChromosome(1, 2, 0, 5, "range")
		expected_names := NewSet[string]()

		for _, c := range g.chromosomes {
			expected_names.Add(c.name)
		}
		expected_names.Add("range")

		g.Insert(1, c)
		observed_names := NewSet[string]()
		for _, c = range g.chromosomes {
			observed_names.Add(c.name)
		}

		if !expected_names.Equal(observed_names) {
			t.Errorf("Genome[int].Insert failed: expected names %v, observed %v", expected_names, observed_names)
		}
	})

	t.Run("Append", func(t *testing.T) {
		t.Parallel()
		g := firstGenome()
		for i := 1; i < 11; i++ {
			c, err := rangeChromosome(1, i, 0, 5)
			if err != nil {
				t.Errorf("Genome[int].Append failed with error: %v", err.Error())
			}
			err = g.Append(c)
			if err != nil {
				t.Errorf("Genome[int].Append failed with error: %v", err.Error())
			}
			observed := g.chromosomes[len(g.chromosomes)-1]
			if observed.name != c.name {
				t.Errorf("Genome[int].Append did not add to end; expected %v, observed %v", c, observed)
			}
		}
	})

	t.Run("Duplicate", func(t *testing.T) {
		t.Parallel()
		g := firstGenome()
		err := g.Duplicate(0)
		if err != nil {
			t.Errorf("Genome[int].Duplicate failed with error: %v", err.Error())
		}
		first := g.chromosomes[0]
		second := g.chromosomes[1]
		if first.name != second.name {
			t.Fatalf("Genome[int].Duplicate failed to duplicate Gene: expected %v, observed %v", first.ToMap(), second.ToMap())
		}
	})

	t.Run("Delete", func(t *testing.T) {
		t.Parallel()
		for i := 1; i < 111; i++ {
			g := firstGenome()
			expected_size := len(g.chromosomes) - 1
			err := g.Delete(0)
			if err != nil {
				t.Errorf("Genome[int].Delete failed with error: %v", err.Error())
			}
			if expected_size != len(g.chromosomes) {
				t.Fatalf("Genome[int].Delete failed to delete Gene: expected %v, observed %v", expected_size, len(g.chromosomes))
			}
		}
	})

	t.Run("Substitute", func(t *testing.T) {
		t.Parallel()
		g, _ := rangeGenome(1, 2, 3, 0, 5)
		c, _ := rangeChromosome(1, 2, 2, 4)
		g.Substitute(0, c)
		expected := c.alleles
		if !equal(expected, g.chromosomes[0].alleles) {
			t.Fatalf("Genome[int].Substitute failed: expected [%v], observed [%v]", expected, g.chromosomes[0].alleles)
		}
	})

	t.Run("Recombine", func(t *testing.T) {
		t.Parallel()
		dad, _ := rangeGenome(2, 2, 3, 0, 5, "dad")
		mom, _ := rangeGenome(2, 2, 3, 6, 11, "mom")
		child, err := dad.Recombine(mom, []int{}, RecombineOptions{})
		if err != nil {
			t.Fatalf("Genome[int].Recombine failed with error: %s", err.Error())
		}
		parents := NewSet[string]()
		dad_bases := NewSet[int]()
		mom_bases := NewSet[int]()
		for _, chromosome := range dad.chromosomes {
			for _, allele := range chromosome.alleles {
				for _, gene := range allele.genes {
					for _, base := range gene.bases {
						dad_bases.Add(base)
					}
				}
			}
		}
		for _, chromosome := range mom.chromosomes {
			for _, allele := range chromosome.alleles {
				for _, gene := range allele.genes {
					for _, base := range gene.bases {
						mom_bases.Add(base)
					}
				}
			}
		}

		for _, chromosome := range child.chromosomes {
			for _, allele := range chromosome.alleles {
				for _, gene := range allele.genes {
					for _, base := range gene.bases {
						if dad_bases.Contains(base) {
							parents.Add(dad.name)
						} else if mom_bases.Contains(base) {
							parents.Add(mom.name)
						} else {
							t.Fatalf("Genome[int].Recombine failed: encountered base not from parents: %v", base)
						}
					}
				}
			}
		}
		if parents.Len() < 2 {
			fmt.Printf("mom: %v\n", mom.ToMap())
			fmt.Printf("dad: %v\n", dad.ToMap())
			fmt.Printf("child: %v\n", child.ToMap())
			t.Fatalf("Genome[int].Recombine failed: expected bases from 2 parents, observed %d", parents.Len())
		}
	})

	t.Run("MakeGenome", func(t *testing.T) {
		t.Parallel()
		names := NewSet[string]()
		maps := make(map[string]map[string][]map[string][]map[string][]map[string][]int)

		for i := 0; i < 10; i++ {
			g, err := MakeGenome[int](MakeOptions[int]{
				n_chromosomes: NewOption(uint(2)),
				n_alleles:     NewOption(uint(3)),
				n_genes:       NewOption(uint(3)),
				n_bases:       NewOption(uint(5)),
				base_factory:  NewOption(factory),
			})
			if err != nil {
				t.Fatalf("MakeGenome[int] failed with error: %v", err)
			}
			names.Add(g.name)

			_, ok := maps[g.name]
			if !ok {
				maps[g.name] = g.ToMap()
			}
		}

		if names.Len() < 8 {
			t.Fatalf("MakeGenome[int] failed to produce enough random names: expected >= 8, observed %d", names.Len())
		}

		if len(maps) < 8 {
			t.Fatalf("MakeGenome[int] failed to produce enough random sequences: expected >= 8, observed %d", len(maps))
		}
	})

	t.Run("Sequence", func(t *testing.T) {
		t.Parallel()
		genome := firstGenome()
		separator := []int{0, 0, 0, 0, 0}
		sequence := genome.Sequence(separator)
		unpacked := GenomeFromSequence(sequence, separator)

		if len(unpacked.chromosomes) != len(genome.chromosomes) {
			t.Errorf("Genome[int].Sequence -> GenomeFromSequence failed: expected %d genes, observed %d", len(genome.chromosomes), len(unpacked.chromosomes))
		}

		repacked := unpacked.Sequence(separator)
		if !equal(sequence, repacked) {
			t.Errorf("Genome[int].Sequence -> GenomeFromSequence -> .Sequence failed: expected %v, observed %v", sequence, repacked)
		}
	})
}
