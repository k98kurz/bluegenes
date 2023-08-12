package bluegenes

import (
	"fmt"
	"math/rand"
	"testing"
)

func firstGene() *Gene[int] {
	return &Gene[int]{
		Name:  "test",
		Bases: []int{1, 2, 3},
	}
}

func secondGene() *Gene[int] {
	return &Gene[int]{
		Name:  "tset",
		Bases: []int{4, 5, 6},
	}
}

func thirdGene() *Gene[int] {
	return &Gene[int]{
		Name:  "gen3",
		Bases: []int{7, 8, 9},
	}
}

func fourthGene() *Gene[int] {
	return &Gene[int]{
		Name:  "gen4",
		Bases: []int{10, 11, 12},
	}
}

func firstNucleosome() *Nucleosome[int] {
	return &Nucleosome[int]{
		Name: "al1",
		Genes: []*Gene[int]{
			firstGene(),
			secondGene(),
		},
	}
}

func secondNucleosome() *Nucleosome[int] {
	return &Nucleosome[int]{
		Name: "al2",
		Genes: []*Gene[int]{
			thirdGene(),
			fourthGene(),
		},
	}
}

func firstChromosome() *Chromosome[int] {
	return &Chromosome[int]{
		Name: "c1",
		Nucleosomes: []*Nucleosome[int]{
			firstNucleosome(),
			secondNucleosome(),
		},
	}
}

func secondChromosome() *Chromosome[int] {
	return &Chromosome[int]{
		Name: "c1",
		Nucleosomes: []*Nucleosome[int]{
			secondNucleosome(),
			firstNucleosome(),
		},
	}
}

func firstGenome() *Genome[int] {
	return &Genome[int]{
		Name: "Genome",
		Chromosomes: []*Chromosome[int]{
			firstChromosome(),
			secondChromosome(),
		},
	}
}

func rangeGene(start int, stop int, name ...string) (*Gene[int], error) {
	g := &Gene[int]{Name: "GnR"}
	if start >= stop {
		return g, anError{"start Must be <= stop"}
	}
	for i := start; i <= stop; i++ {
		g.Append(i)
	}
	if len(name) > 0 {
		g.Name = name[0]
	}
	return g, nil
}

func rangeNucleosome(size int, start int, stop int, name ...string) (*Nucleosome[int], error) {
	a := &Nucleosome[int]{Name: "AlR"}

	if size <= 0 {
		return a, anError{"size Must be > 0"}
	}

	for i := 0; i < size; i++ {
		g, err := rangeGene(start, stop, fmt.Sprintf("GnR%d", i))
		if err != nil {
			return a, err
		}
		a.Append(g)
	}

	if len(name) > 0 {
		a.Name = name[0]
	}

	return a, nil
}

func rangeChromosome(size int, nucleosome_size int, start int, stop int,
	name ...string) (*Chromosome[int], error) {
	c := &Chromosome[int]{Name: "ChR"}

	if size <= 0 {
		return c, anError{"size Must be > 0"}
	}

	if nucleosome_size <= 0 {
		return c, anError{"nucleosome_size Must be > 0"}
	}

	for i := 0; i < size; i++ {
		g, err := rangeNucleosome(nucleosome_size, start, stop, fmt.Sprintf("AlR%d", i))
		if err != nil {
			return c, err
		}
		c.Append(g)
	}

	if len(name) > 0 {
		c.Name = name[0]
	}

	return c, nil
}

func rangeGenome(size int, chromosome_size int, nucleosome_size int, start int,
	stop int, name ...string) (*Genome[int], error) {
	g := &Genome[int]{Name: "GenomR"}

	if size <= 0 {
		return g, anError{"size Must be > 0"}
	}

	if nucleosome_size <= 0 {
		return g, anError{"nucleosome_size Must be > 0"}
	}

	if chromosome_size <= 0 {
		return g, anError{"chromosome_size Must be > 0"}
	}

	for i := 0; i < size; i++ {
		c, err := rangeChromosome(chromosome_size, nucleosome_size, start, stop,
			fmt.Sprintf("AlR%d", i))
		if err != nil {
			return g, err
		}
		g.Append(c)
	}

	if len(name) > 0 {
		g.Name = name[0]
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
		} else if c.Name != g.Name {
			t.Errorf("Gene[int].Copy failed to copy Name; got %s, expected %s", c.Name, g.Name)
		} else if len(c.Bases) != len(g.Bases) {
			t.Fatal("Gene[int].Copy failed to copy bases")
		}

		for i, item := range c.Bases {
			if g.Bases[i] != item {
				t.Errorf("Gene[int].Copy failed to copy Bases: got %d, expected %d", item, g.Bases[i])
			}
		}
	})

	t.Run("Insert", func(t *testing.T) {
		t.Parallel()
		g := firstGene()

		g.Insert(1, 15)
		expected := []int{1, 15, 2, 3}
		for i, item := range g.Bases {
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
			observed := g.Bases[len(g.Bases)-1]
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
			observed := g.Bases[:len(seq)]
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
			for k, item := range g.Bases {
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
		for i, item := range g.Bases {
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
				if len(g.Bases) > i-k {
					t.Fatalf("Gene[int].DeleteSequence failed to remove enough items: expected %d len, observed %d", i-k, len(g.Bases))
				}
			}
		}
	})

	t.Run("Substitute", func(t *testing.T) {
		t.Parallel()
		g := firstGene()
		g.Substitute(0, 15)
		expected := []int{15, 2, 3}
		if !equal(expected, g.Bases) {
			t.Fatalf("Gene[int].Substitute failed: expected [%v], observed [%v]", expected, g.Bases)
		}

		g = firstGene()
		g.Substitute(1, 15)
		expected = []int{1, 15, 3}
		if !equal(expected, g.Bases) {
			t.Fatalf("Gene[int].Substitute failed: expected [%v], observed [%v]", expected, g.Bases)
		}

		g = firstGene()
		g.Substitute(2, 15)
		expected = []int{1, 2, 15}
		if !equal(expected, g.Bases) {
			t.Fatalf("Gene[int].Substitute failed: expected [%v], observed [%v]", expected, g.Bases)
		}
	})

	t.Run("Recombine", func(t *testing.T) {
		t.Parallel()
		for i := 0; i < 111; i++ {
			g1, _ := rangeGene(0, 5, "dad")
			g2, _ := rangeGene(6, 11, "mom")
			g3 := &Gene[int]{}
			err := g1.Recombine(g2, []int{}, g3, RecombineOptions{})
			if err != nil {
				t.Fatalf("Gene[int].Recombine failed with error: %s", err.Error())
			}
			parents := newSet[string]()
			for _, item := range g3.Bases {
				if contains(g1.Bases, item) {
					parents.add(g1.Name)
				} else if contains(g2.Bases, item) {
					parents.add(g2.Name)
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

		deserialized := GeneFromMap(observed)

		if g.Name != deserialized.Name {
			t.Errorf("GeneFromMap[int] failed: expected Name='%s', observed %s", g.Name, deserialized.Name)
		}

		if !equal(g.Bases, deserialized.Bases) {
			t.Errorf("GeneFromMap]int] failed: expected Bases=%v, observed %v", g.Bases, deserialized.Bases)
		}
	})

	t.Run("MakeGene", func(t *testing.T) {
		t.Parallel()
		Names := newSet[string]()
		sequences := [][]int{}

		for i := 0; i < 10; i++ {
			g, err := MakeGene[int](MakeOptions[int]{
				NBases:      NewOption(uint(5)),
				BaseFactory: NewOption(factory),
			})
			if err != nil {
				t.Fatalf("MakeGene[int] failed with error: %v", err)
			}
			Names.add(g.Name)

			if !containsSlice(sequences, g.Bases) {
				sequences = append(sequences, g.Bases)
			}
		}

		if Names.len() < 8 {
			t.Fatalf("MakeGene[int] failed to produce enough random Names: expected >= 8, observed %d", Names.len())
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

		if !equal(gene.Bases, unpacked.Bases) {
			t.Errorf("GeneFromSequence[int] failed: expected %v, observed %v", gene.Bases, unpacked.Bases)
		}
	})
}

func TestNucleosome(t *testing.T) {
	t.Run("Copy", func(t *testing.T) {
		t.Parallel()
		a := firstNucleosome()
		c := a.Copy()

		if c == a {
			t.Error("Nucleosome[int].Copy failed; received pointer to same memory")
		} else if c.Name != a.Name {
			t.Errorf("Nucleosome[int].Copy failed to copy Name; got %s, expected %s", c.Name, a.Name)
		} else if len(c.Genes) != len(a.Genes) {
			t.Fatal("Nucleosome[int].Copy failed to copy Genes")
		}

		for i, item := range c.Genes {
			if a.Genes[i] != item {
				t.Errorf("Nucleosome[int].Copy failed to copy Genes: got %v, expected %v", item, a.Genes[i])
			}
		}
	})

	t.Run("Insert", func(t *testing.T) {
		t.Parallel()
		a := firstNucleosome()
		g, _ := rangeGene(0, 5, "range")
		expected_Names := newSet[string]()

		for _, g := range a.Genes {
			expected_Names.add(g.Name)
		}
		expected_Names.add("range")

		a.Insert(1, g)
		observed_Names := newSet[string]()
		for _, g = range a.Genes {
			observed_Names.add(g.Name)
		}

		if !expected_Names.equal(observed_Names) {
			t.Errorf("Nucleosome[int].Insert failed: expected Names %v, observed %v", expected_Names, observed_Names)
		}
	})

	t.Run("Append", func(t *testing.T) {
		t.Parallel()
		a := firstNucleosome()
		for i := 1; i < 11; i++ {
			g, err := rangeGene(0, i)
			if err != nil {
				t.Errorf("Nucleosome[int].Append failed with error: %v", err.Error())
			}
			err = a.Append(g)
			if err != nil {
				t.Errorf("Nucleosome[int].Append failed with error: %v", err.Error())
			}
			observed := a.Genes[len(a.Genes)-1]
			if observed.Name != g.Name {
				t.Errorf("Nucleosome[int].Append did not add to end; expected %v, observed %v", g, observed)
			}
		}
	})

	t.Run("Duplicate", func(t *testing.T) {
		t.Parallel()
		a := firstNucleosome()
		err := a.Duplicate(0)
		if err != nil {
			t.Errorf("Nucleosome[int].Duplicate failed with error: %v", err.Error())
		}
		first := a.Genes[0]
		second := a.Genes[1]
		if first.Name != second.Name {
			t.Fatalf("Nucleosome[int].Duplicate failed to duplicate Gene: expected %v, observed %v", first, second)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		t.Parallel()
		for i := 1; i < 111; i++ {
			a := firstNucleosome()
			expected_size := len(a.Genes) - 1
			err := a.Delete(0)
			if err != nil {
				t.Errorf("Nucleosome[int].Delete failed with error: %v", err.Error())
			}
			if expected_size != len(a.Genes) {
				t.Fatalf("Nucleosome[int].Delete failed to delete Gene: expected %v, observed %v", expected_size, len(a.Genes))
			}
		}
	})

	t.Run("Substitute", func(t *testing.T) {
		t.Parallel()
		a, _ := rangeNucleosome(3, 0, 5)
		g, _ := rangeGene(2, 4)
		a.Substitute(0, g)
		expected := g.Bases
		if !equal(expected, a.Genes[0].Bases) {
			t.Fatalf("Nucleosome[int].Substitute failed: expected [%v], observed [%v]", expected, a.Genes[0].Bases)
		}

		a.Substitute(1, g)
		if !equal(expected, a.Genes[1].Bases) {
			t.Fatalf("Nucleosome[int].Substitute failed: expected [%v], observed [%v]", expected, a.Genes[1].Bases)
		}

		a.Substitute(2, g)
		if !equal(expected, a.Genes[2].Bases) {
			t.Fatalf("Nucleosome[int].Substitute failed: expected [%v], observed [%v]", expected, a.Genes[2].Bases)
		}
	})

	t.Run("Recombine", func(t *testing.T) {
		t.Parallel()
		dad, _ := rangeNucleosome(3, 0, 5, "dad")
		mom, _ := rangeNucleosome(3, 6, 11, "mom")
		child := &Nucleosome[int]{}
		err := dad.Recombine(mom, []int{}, child, RecombineOptions{})
		if err != nil {
			t.Fatalf("Nucleosome[int].Recombine failed with error: %s", err.Error())
		}
		parents := newSet[string]()
		dad_bases := newSet[int]()
		mom_bases := newSet[int]()
		for _, gene := range dad.Genes {
			for _, base := range gene.Bases {
				dad_bases.add(base)
			}
		}
		for _, gene := range mom.Genes {
			for _, base := range gene.Bases {
				mom_bases.add(base)
			}
		}
		for _, gene := range child.Genes {
			for _, base := range gene.Bases {
				if dad_bases.contains(base) {
					parents.add(dad.Name)
				} else if mom_bases.contains(base) {
					parents.add(mom.Name)
				} else {
					t.Fatalf("Nucleosome[int].Recombine failed: encountered base not from parents: %v", base)
				}
			}
		}
		if parents.len() < 2 {
			fmt.Printf("mom: %v\n", mom.ToMap())
			fmt.Printf("dad: %v\n", dad.ToMap())
			fmt.Printf("child: %v\n", child.ToMap())
			t.Fatalf("Nucleosome[int].Recombine failed: expected bases from 2 parents, observed %d", parents.len())
		}
	})

	t.Run("MakeNucleosome", func(t *testing.T) {
		t.Parallel()
		Names := newSet[string]()
		maps := make(map[string]map[string][]map[string][]int)

		for i := 0; i < 10; i++ {
			a, err := MakeNucleosome[int](MakeOptions[int]{
				NGenes:      NewOption(uint(3)),
				NBases:      NewOption(uint(5)),
				BaseFactory: NewOption(factory),
			})
			if err != nil {
				t.Fatalf("MakeNucleosome[int] failed with error: %v", err)
			}
			Names.add(a.Name)

			_, ok := maps[a.Name]
			if !ok {
				maps[a.Name] = a.ToMap()
			}
		}

		if Names.len() < 8 {
			t.Fatalf("MakeNucleosome[int] failed to produce enough random Names: expected >= 8, observed %d", Names.len())
		}

		if len(maps) < 8 {
			t.Fatalf("MakeNucleosome[int] failed to produce enough random sequences: expected >= 8, observed %d", len(maps))
		}
	})

	t.Run("Sequence", func(t *testing.T) {
		t.Parallel()
		nucleosome := firstNucleosome()
		separator := []int{0, 0, 0, 0, 0}
		sequence := nucleosome.Sequence(separator)
		unpacked := NucleosomeFromSequence(sequence, separator)

		if len(unpacked.Genes) != len(nucleosome.Genes) {
			t.Errorf("Nucleosome[int].Sequence -> NucleosomeFromSequence failed: expected %d Genes, observed %d", len(nucleosome.Genes), len(unpacked.Genes))
		}

		repacked := unpacked.Sequence(separator)
		if !equal(sequence, repacked) {
			t.Errorf("Nucleosome[int].Sequence -> NucleosomeFromSequence -> .Sequence failed: expected %v, observed %v", sequence, repacked)
		}
	})

	t.Run("SequenceEmptyGene", func(t *testing.T) {
		t.Parallel()
		nucleosome := firstNucleosome()
		nucleosome.Insert(1, &Gene[int]{Name: "empty", Bases: []int{}})
		separator := []int{0, 0, 0, 0, 0}
		sequence := nucleosome.Sequence(separator)
		unpacked := NucleosomeFromSequence(sequence, separator)

		if len(unpacked.Genes) != len(nucleosome.Genes) {
			t.Errorf("Nucleosome[int].Sequence -> NucleosomeFromSequence failed: expected %d Genes, observed %d", len(nucleosome.Genes), len(unpacked.Genes))
		}

		repacked := unpacked.Sequence(separator)
		if !equal(sequence, repacked) {
			t.Errorf("Nucleosome[int].Sequence -> NucleosomeFromSequence -> .Sequence failed: expected %v, observed %v", sequence, repacked)
		}

		for i, gene := range unpacked.Genes {
			if !equal(gene.Bases, nucleosome.Genes[i].Bases) {
				t.Errorf("Nucleosome[int].Sequence with empty gene -> "+
					"NucleosomeFromSequence failed for Gene[%d]: expected %v, observed %v",
					i, nucleosome.Genes[i].Bases, gene.Bases)
			}
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
		} else if p.Name != c.Name {
			t.Errorf("Chromosome[int].Copy failed to copy Name; got %s, expected %s", p.Name, c.Name)
		} else if len(p.Nucleosomes) != len(c.Nucleosomes) {
			t.Fatal("Chromosome[int].Copy failed to copy nucleosomes")
		}

		for i, nucleosome := range p.Nucleosomes {
			if c.Nucleosomes[i].Name != nucleosome.Name {
				t.Errorf("Chromosome[int].Copy failed to copy nucleosomes: got %v, expected %v", nucleosome.ToMap(), c.Nucleosomes[i].ToMap())
				continue
			}
			for k, gene := range nucleosome.Genes {
				if gene.Name != c.Nucleosomes[i].Genes[k].Name || !equal(gene.Bases, c.Nucleosomes[i].Genes[k].Bases) {
					t.Errorf("Chromosome[int].Copy failed to copy Genes: got %v, expected %v", gene.ToMap(), c.Nucleosomes[i].Genes[k].ToMap())
					break
				}
			}
		}
	})

	t.Run("Insert", func(t *testing.T) {
		t.Parallel()
		c := firstChromosome()
		a, _ := rangeNucleosome(2, 0, 5, "range")
		expected_Names := newSet[string]()

		for _, a := range c.Nucleosomes {
			expected_Names.add(a.Name)
		}
		expected_Names.add("range")

		c.Insert(1, a)
		observed_Names := newSet[string]()
		for _, a = range c.Nucleosomes {
			observed_Names.add(a.Name)
		}

		if !expected_Names.equal(observed_Names) {
			t.Errorf("Chromosome[int].Insert failed: expected Names %v, observed %v", expected_Names, observed_Names)
		}
	})

	t.Run("Append", func(t *testing.T) {
		t.Parallel()
		c := firstChromosome()
		for i := 1; i < 11; i++ {
			a, err := rangeNucleosome(i, 0, 5)
			if err != nil {
				t.Errorf("Chromosome[int].Append failed with error: %v", err.Error())
			}
			err = c.Append(a)
			if err != nil {
				t.Errorf("Chromosome[int].Append failed with error: %v", err.Error())
			}
			observed := c.Nucleosomes[len(c.Nucleosomes)-1]
			if observed.Name != a.Name {
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
		first := c.Nucleosomes[0]
		second := c.Nucleosomes[1]
		if first.Name != second.Name {
			t.Fatalf("Chromosome[int].Duplicate failed to duplicate Gene: expected %v, observed %v", first.ToMap(), second.ToMap())
		}
	})

	t.Run("Delete", func(t *testing.T) {
		t.Parallel()
		for i := 1; i < 111; i++ {
			c := firstChromosome()
			expected_size := len(c.Nucleosomes) - 1
			err := c.Delete(0)
			if err != nil {
				t.Errorf("Chromosome[int].Delete failed with error: %v", err.Error())
			}
			if expected_size != len(c.Nucleosomes) {
				t.Fatalf("Chromosome[int].Delete failed to delete Gene: expected %v, observed %v", expected_size, len(c.Nucleosomes))
			}
		}
	})

	t.Run("Substitute", func(t *testing.T) {
		t.Parallel()
		c, _ := rangeChromosome(2, 3, 0, 5)
		a, _ := rangeNucleosome(2, 2, 4)
		c.Substitute(0, a)
		expected := a.Genes
		if !equal(expected, c.Nucleosomes[0].Genes) {
			t.Fatalf("Chromosome[int].Substitute failed: expected [%v], observed [%v]", expected, c.Nucleosomes[0].Genes)
		}

		c.Substitute(1, a)
		if !equal(expected, c.Nucleosomes[1].Genes) {
			t.Fatalf("Chromosome[int].Substitute failed: expected [%v], observed [%v]", expected, c.Nucleosomes[1].Genes)
		}
	})

	t.Run("Recombine", func(t *testing.T) {
		t.Parallel()
		dad, _ := rangeChromosome(2, 3, 0, 5, "dad")
		mom, _ := rangeChromosome(2, 3, 6, 11, "mom")
		child := &Chromosome[int]{}
		err := dad.Recombine(mom, []int{}, child, RecombineOptions{})
		if err != nil {
			t.Fatalf("Chromosome[int].Recombine failed with error: %s", err.Error())
		}
		parents := newSet[string]()
		dad_bases := newSet[int]()
		mom_bases := newSet[int]()
		for _, nucleosome := range dad.Nucleosomes {
			for _, gene := range nucleosome.Genes {
				for _, base := range gene.Bases {
					dad_bases.add(base)
				}
			}
		}
		for _, nucleosome := range mom.Nucleosomes {
			for _, gene := range nucleosome.Genes {
				for _, base := range gene.Bases {
					mom_bases.add(base)
				}
			}
		}

		for _, nucleosome := range child.Nucleosomes {
			for _, gene := range nucleosome.Genes {
				for _, base := range gene.Bases {
					if dad_bases.contains(base) {
						parents.add(dad.Name)
					} else if mom_bases.contains(base) {
						parents.add(mom.Name)
					} else {
						t.Fatalf("Chromosome[int].Recombine failed: encountered base not from parents: %v", base)
					}
				}
			}
		}
		if parents.len() < 2 {
			fmt.Printf("mom: %v\n", mom.ToMap())
			fmt.Printf("dad: %v\n", dad.ToMap())
			fmt.Printf("child: %v\n", child.ToMap())
			t.Fatalf("Chromosome[int].Recombine failed: expected bases from 2 parents, observed %d", parents.len())
		}
	})

	t.Run("MakeChromosome", func(t *testing.T) {
		t.Parallel()
		Names := newSet[string]()
		maps := make(map[string]map[string][]map[string][]map[string][]int)

		for i := 0; i < 10; i++ {
			a, err := MakeChromosome[int](MakeOptions[int]{
				NNucleosomes: NewOption(uint(3)),
				NGenes:       NewOption(uint(3)),
				NBases:       NewOption(uint(5)),
				BaseFactory:  NewOption(factory),
			})
			if err != nil {
				t.Fatalf("MakeChromosome[int] failed with error: %v", err)
			}
			Names.add(a.Name)

			_, ok := maps[a.Name]
			if !ok {
				maps[a.Name] = a.ToMap()
			}
		}

		if Names.len() < 8 {
			t.Fatalf("MakeChromosome[int] failed to produce enough random Names: expected >= 8, observed %d", Names.len())
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

		if len(unpacked.Nucleosomes) != len(chromosome.Nucleosomes) {
			t.Errorf(
				"Chromosome[int].Sequence -> ChromosomeFromSequence failed: expected %d Nucleosomes, observed %d",
				len(chromosome.Nucleosomes), len(unpacked.Nucleosomes))
		}

		for i := 0; i < len(unpacked.Nucleosomes); i++ {
			ua, ca := unpacked.Nucleosomes[i], chromosome.Nucleosomes[i]
			if len(ua.Genes) != len(ca.Genes) {
				t.Fatalf(
					"Chromosome[int].Sequence -> ChromosomeFromSequence failed:"+
						" expected %d Genes, observed %d", len(ca.Genes),
					len(ua.Genes))
			}

			for j := 0; j < len(ua.Genes); j++ {
				if !equal(ua.Genes[j].Bases, ca.Genes[j].Bases) {
					t.Fatalf(
						"Chromosome[int].Sequence -> ChromosomeFromSequence failed:"+
							" expected %v Bases, observed %v", ca.Genes[j].Bases,
						ua.Genes[j].Bases)
				}
			}
		}

		repacked := unpacked.Sequence(separator)
		if !equal(sequence, repacked) {
			t.Errorf("Chromosome[int].Sequence -> ChromosomeFromSequence -> .Sequence failed: expected %v, observed %v", sequence, repacked)
		}
	})

	t.Run("SequenceEmptyNucleosome", func(t *testing.T) {
		t.Parallel()
		chromosome := firstChromosome()
		chromosome.Insert(1, &Nucleosome[int]{Name: "empty", Genes: []*Gene[int]{}})
		separator := []int{0, 0, 0, 0, 0}
		sequence := chromosome.Sequence(separator)
		unpacked := ChromosomeFromSequence(sequence, separator)

		if len(unpacked.Nucleosomes) != len(chromosome.Nucleosomes) {
			t.Errorf(
				"Chromosome[int].Sequence -> ChromosomeFromSequence failed: expected %d Nucleosomes, observed %d",
				len(chromosome.Nucleosomes), len(unpacked.Nucleosomes))
		}

		for i := 0; i < len(unpacked.Nucleosomes); i++ {
			ua, ca := unpacked.Nucleosomes[i], chromosome.Nucleosomes[i]
			if len(ua.Genes) != len(ca.Genes) {
				t.Fatalf(
					"Chromosome[int].Sequence -> ChromosomeFromSequence failed:"+
						" expected %d Genes, observed %d", len(ca.Genes),
					len(ua.Genes))
			}

			for j := 0; j < len(ua.Genes); j++ {
				if !equal(ua.Genes[j].Bases, ca.Genes[j].Bases) {
					t.Fatalf(
						"Chromosome[int].Sequence -> ChromosomeFromSequence failed:"+
							" expected %v Bases, observed %v", ca.Genes[j].Bases,
						ua.Genes[j].Bases)
				}
			}
		}

		repacked := unpacked.Sequence(separator)
		if !equal(sequence, repacked) {
			t.Errorf("Chromosome[int].Sequence -> ChromosomeFromSequence -> .Sequence failed: expected %v, observed %v", sequence, repacked)
		}
	})

	t.Run("SequenceFloats", func(t *testing.T) {
		t.Parallel()
		opts := MakeOptions[float64]{
			NBases:       NewOption(uint(12)),
			NGenes:       NewOption(uint(12)),
			NNucleosomes: NewOption(uint(2)),
			BaseFactory:  NewOption(rand.Float64),
		}
		chromosome, _ := MakeChromosome[float64](opts)
		separator := []float64{0.0, 0.0, 0.0, 0.0, 0.0}
		sequence := chromosome.Sequence(separator)
		unpacked := ChromosomeFromSequence(sequence, separator)

		if len(unpacked.Nucleosomes) != len(chromosome.Nucleosomes) {
			t.Errorf(
				"Chromosome[float64].Sequence -> ChromosomeFromSequence failed: expected %d Nucleosomes, observed %d",
				len(chromosome.Nucleosomes), len(unpacked.Nucleosomes))
		}

		for i := 0; i < len(unpacked.Nucleosomes); i++ {
			ua, ca := unpacked.Nucleosomes[i], chromosome.Nucleosomes[i]
			if len(ua.Genes) != len(ca.Genes) {
				t.Fatalf(
					"Chromosome[float64].Sequence -> ChromosomeFromSequence failed:"+
						" expected %d Genes, observed %d", len(ca.Genes),
					len(ua.Genes))
			}

			for j := 0; j < len(ua.Genes); j++ {
				if !equal(ua.Genes[j].Bases, ca.Genes[j].Bases) {
					t.Fatalf(
						"Chromosome[float64].Sequence -> ChromosomeFromSequence failed:"+
							" expected %v Bases, observed %v", ca.Genes[j].Bases,
						ua.Genes[j].Bases)
				}
			}
		}

		repacked := unpacked.Sequence(separator)
		if !equal(sequence, repacked) {
			t.Errorf("Chromosome[float64].Sequence -> ChromosomeFromSequence -> .Sequence failed: expected %v, observed %v", sequence, repacked)
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
		} else if p.Name != g.Name {
			t.Errorf("Genome[int].Copy failed to copy Name; got %s, expected %s", p.Name, g.Name)
		} else if len(p.Chromosomes) != len(g.Chromosomes) {
			t.Fatal("Genome[int].Copy failed to copy chromosomes")
		}

		for i, chromosome := range p.Chromosomes {
			if g.Chromosomes[i].Name != chromosome.Name {
				t.Errorf("Genome[int].Copy failed to copy Chromosomes: got %v, expected %v", chromosome.ToMap(), g.Chromosomes[i].ToMap())
				continue
			}
			for k, nucleosome := range chromosome.Nucleosomes {
				for j, gene := range nucleosome.Genes {
					if gene.Name != g.Chromosomes[i].Nucleosomes[k].Genes[j].Name || !equal(gene.Bases, g.Chromosomes[i].Nucleosomes[k].Genes[j].Bases) {
						t.Errorf("Genome[int].Copy failed to copy Genes: got %v, expected %v", gene.ToMap(), g.Chromosomes[i].Nucleosomes[k].Genes[j].ToMap())
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
		expected_Names := newSet[string]()

		for _, c := range g.Chromosomes {
			expected_Names.add(c.Name)
		}
		expected_Names.add("range")

		g.Insert(1, c)
		observed_Names := newSet[string]()
		for _, c = range g.Chromosomes {
			observed_Names.add(c.Name)
		}

		if !expected_Names.equal(observed_Names) {
			t.Errorf("Genome[int].Insert failed: expected Names %v, observed %v", expected_Names, observed_Names)
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
			observed := g.Chromosomes[len(g.Chromosomes)-1]
			if observed.Name != c.Name {
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
		first := g.Chromosomes[0]
		second := g.Chromosomes[1]
		if first.Name != second.Name {
			t.Fatalf("Genome[int].Duplicate failed to duplicate Gene: expected %v, observed %v", first.ToMap(), second.ToMap())
		}
	})

	t.Run("Delete", func(t *testing.T) {
		t.Parallel()
		for i := 1; i < 111; i++ {
			g := firstGenome()
			expected_size := len(g.Chromosomes) - 1
			err := g.Delete(0)
			if err != nil {
				t.Errorf("Genome[int].Delete failed with error: %v", err.Error())
			}
			if expected_size != len(g.Chromosomes) {
				t.Fatalf("Genome[int].Delete failed to delete Gene: expected %v, observed %v", expected_size, len(g.Chromosomes))
			}
		}
	})

	t.Run("Substitute", func(t *testing.T) {
		t.Parallel()
		g, _ := rangeGenome(1, 2, 3, 0, 5)
		c, _ := rangeChromosome(1, 2, 2, 4)
		g.Substitute(0, c)
		expected := c.Nucleosomes
		if !equal(expected, g.Chromosomes[0].Nucleosomes) {
			t.Fatalf("Genome[int].Substitute failed: expected [%v], observed [%v]", expected, g.Chromosomes[0].Nucleosomes)
		}
	})

	t.Run("Recombine", func(t *testing.T) {
		t.Parallel()
		dad, _ := rangeGenome(2, 2, 3, 0, 5, "dad")
		mom, _ := rangeGenome(2, 2, 3, 6, 11, "mom")
		child := &Genome[int]{}
		err := dad.Recombine(mom, []int{}, child, RecombineOptions{})
		if err != nil {
			t.Fatalf("Genome[int].Recombine failed with error: %s", err.Error())
		}
		parents := newSet[string]()
		dad_bases := newSet[int]()
		mom_bases := newSet[int]()
		for _, chromosome := range dad.Chromosomes {
			for _, nucleosome := range chromosome.Nucleosomes {
				for _, gene := range nucleosome.Genes {
					for _, base := range gene.Bases {
						dad_bases.add(base)
					}
				}
			}
		}
		for _, chromosome := range mom.Chromosomes {
			for _, nucleosome := range chromosome.Nucleosomes {
				for _, gene := range nucleosome.Genes {
					for _, base := range gene.Bases {
						mom_bases.add(base)
					}
				}
			}
		}

		for _, chromosome := range child.Chromosomes {
			for _, nucleosome := range chromosome.Nucleosomes {
				for _, gene := range nucleosome.Genes {
					for _, base := range gene.Bases {
						if dad_bases.contains(base) {
							parents.add(dad.Name)
						} else if mom_bases.contains(base) {
							parents.add(mom.Name)
						} else {
							t.Fatalf("Genome[int].Recombine failed: encountered base not from parents: %v", base)
						}
					}
				}
			}
		}
		if parents.len() < 2 {
			fmt.Printf("mom: %v\n", mom.ToMap())
			fmt.Printf("dad: %v\n", dad.ToMap())
			fmt.Printf("child: %v\n", child.ToMap())
			t.Fatalf("Genome[int].Recombine failed: expected bases from 2 parents, observed %d", parents.len())
		}
	})

	t.Run("MakeGenome", func(t *testing.T) {
		t.Parallel()
		Names := newSet[string]()
		maps := make(map[string]map[string][]map[string][]map[string][]map[string][]int)

		for i := 0; i < 10; i++ {
			g, err := MakeGenome[int](MakeOptions[int]{
				NChromosomes: NewOption(uint(2)),
				NNucleosomes: NewOption(uint(3)),
				NGenes:       NewOption(uint(3)),
				NBases:       NewOption(uint(5)),
				BaseFactory:  NewOption(factory),
			})
			if err != nil {
				t.Fatalf("MakeGenome[int] failed with error: %v", err)
			}
			Names.add(g.Name)

			_, ok := maps[g.Name]
			if !ok {
				maps[g.Name] = g.ToMap()
			}
		}

		if Names.len() < 8 {
			t.Fatalf("MakeGenome[int] failed to produce enough random Names: expected >= 8, observed %d", Names.len())
		}

		if len(maps) < 8 {
			t.Fatalf("MakeGenome[int] failed to produce enough random sequences: expected >= 8, observed %d", len(maps))
		}
	})

	t.Run("ToMap", func(t *testing.T) {
		t.Parallel()
		n := firstGenome()
		observed := n.ToMap()

		deserialized := GenomeFromMap(observed)

		if n.Name != deserialized.Name {
			t.Errorf("GenomeFromMap[int] failed: expected Name='%s', observed %s", n.Name, deserialized.Name)
		}

		if len(n.Chromosomes) != len(deserialized.Chromosomes) {
			t.Fatalf("GenomeFromMap]int] failed: expected len(Chromosomes)=%d, observed %d", len(n.Chromosomes), len(deserialized.Chromosomes))
		}

		for i, c1 := range n.Chromosomes {
			c2 := deserialized.Chromosomes[i]
			if c1.Name != c2.Name {
				t.Errorf("GenomeFromMap[int] failed: expected Chromosomes[x].Name='%s', observed %s", c1.Name, c2.Name)
			}
			if len(c1.Nucleosomes) != len(c2.Nucleosomes) {
				t.Fatalf("GenomeFromMap]int] failed: expected len(Nucleosomes)=%d, observed %d", len(c1.Nucleosomes), len(c2.Nucleosomes))
			}

			for k, a1 := range c1.Nucleosomes {
				a2 := c2.Nucleosomes[k]
				if a1.Name != a2.Name {
					t.Errorf("GenomeFromMap[int] failed: expected Chromosomes[x].Nucleosomes[y].Name='%s', observed %s", a1.Name, a2.Name)
				}
				if len(a1.Genes) != len(a2.Genes) {
					t.Fatalf("GenomeFromMap]int] failed: expected len(Genes)=%d, observed %d", len(a1.Genes), len(a2.Genes))
				}

				for l, g1 := range a1.Genes {
					g2 := a2.Genes[l]
					if g1.Name != g2.Name {
						t.Errorf("GenomeFromMap[int] failed: expected Chromosomes[x].Nucleosomes[y].Name='%s', observed %s", g1.Name, g2.Name)
					}
					if !equal(g1.Bases, g2.Bases) {
						t.Fatalf("GenomeFromMap]int] failed: expected Bases=%d, observed %d", len(g1.Bases), len(g2.Bases))
					}
				}
			}
		}
	})

	t.Run("Sequence", func(t *testing.T) {
		t.Parallel()
		genome := firstGenome()
		separator := []int{0, 0, 0, 0, 0}
		sequence := genome.Sequence(separator)
		unpacked := GenomeFromSequence(sequence, separator)

		if len(unpacked.Chromosomes) != len(genome.Chromosomes) {
			t.Errorf("Genome[int].Sequence -> GenomeFromSequence failed: expected %d genes, observed %d", len(genome.Chromosomes), len(unpacked.Chromosomes))
		}

		repacked := unpacked.Sequence(separator)
		if !equal(sequence, repacked) {
			t.Errorf("Genome[int].Sequence -> GenomeFromSequence -> .Sequence failed: expected %v, observed %v", sequence, repacked)
		}
	})

	t.Run("SequenceEmptyChromosome", func(t *testing.T) {
		t.Parallel()
		genome := firstGenome()
		genome.Insert(1, &Chromosome[int]{Name: "empty"})
		separator := []int{0, 0, 0, 0, 0}
		sequence := genome.Sequence(separator)
		unpacked := GenomeFromSequence(sequence, separator)

		if len(unpacked.Chromosomes) != len(genome.Chromosomes) {
			t.Errorf("Genome[int].Sequence -> GenomeFromSequence failed: "+
				"expected %d genes, observed %d\n", len(genome.Chromosomes), len(unpacked.Chromosomes))
		}

		repacked := unpacked.Sequence(separator)
		if !equal(sequence, repacked) {
			t.Errorf("Genome[int].Sequence -> GenomeFromSequence -> .Sequence"+
				" failed: expected %v, observed %v\n", sequence, repacked)
		}

		for i, chromosome := range genome.Chromosomes {
			otherChromosome := unpacked.Chromosomes[i]
			if len(chromosome.Nucleosomes) != len(otherChromosome.Nucleosomes) {
				t.Fatalf("Genome[int].Sequence -> GenomeFromSequence empty Nucleosome"+
					"failed: different number of Nucleosomes: expected %d, observed %d\n",
					len(chromosome.Nucleosomes), len(otherChromosome.Nucleosomes))
			}
			for j, nucleosome := range chromosome.Nucleosomes {
				otherNucleosome := otherChromosome.Nucleosomes[j]
				if len(nucleosome.Genes) != len(otherNucleosome.Genes) {
					t.Fatalf("Genome[int].Sequence -> GenomeFromSequence empty"+
						" Nucleosome failed: different number of Genes: expected "+
						"%d, observed %d\n", len(nucleosome.Genes), len(otherNucleosome.Genes))
				}
			}
		}
	})
}
