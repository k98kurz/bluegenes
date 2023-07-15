package bluegenes

import (
	"math/rand"
)

var alphanumerics = []rune{
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o',
	'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D',
	'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S',
	'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '0', '1', '2', '3', '4', '5', '6', '7',
	'8', '9',
}

func RandomName(size int) (string, error) {
	if size < 0 {
		return "", anError{"size must be > 0"}
	}

	s := ""
	l := len(alphanumerics) - 1
	for i := 0; i < size; i++ {
		s = s + string(alphanumerics[RandomInt(0, l)])
	}

	return s, nil
}

func RandomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

type MakeOptions[T comparable] struct {
	NBases       Option[uint]
	NGenes       Option[uint]
	NAlleles     Option[uint]
	NChromosomes Option[uint]
	Name         Option[string]
	BaseFactory  Option[func() T]
}

type RecombineOptions struct {
	RecombineGenes       Option[bool]
	MatchGenes           Option[bool]
	RecombineAlleles     Option[bool]
	MatchAlleles         Option[bool]
	RecombineChromosomes Option[bool]
	MatchChromosomes     Option[bool]
}

func MakeGene[T comparable](options MakeOptions[T]) (*Gene[T], error) {
	g := &Gene[T]{}
	if !options.NBases.ok() {
		return g, missingParameterError{"options.NBases"}
	}
	if !options.BaseFactory.ok() {
		return g, missingParameterError{"options.BaseFactory"}
	}
	for i := 0; i < int(options.NBases.val); i++ {
		b := options.BaseFactory.val()
		g.Append(b)
	}
	if options.Name.ok() {
		g.Name = options.Name.val
	} else {
		g.Name, _ = RandomName(4)
	}
	return g, nil
}

func MakeAllele[T comparable](options MakeOptions[T]) (*Allele[T], error) {
	a := &Allele[T]{}
	if !options.NGenes.ok() {
		return a, missingParameterError{"options.NGenes"}
	}
	if !options.NBases.ok() {
		return a, missingParameterError{"options.NBases"}
	}
	if !options.BaseFactory.ok() {
		return a, missingParameterError{"options.BaseFactory"}
	}
	for i := 0; i < int(options.NGenes.val); i++ {
		g, err := MakeGene(options)
		if err != nil {
			return a, err
		}
		a.Append(g)
	}
	if options.Name.ok() {
		a.Name = options.Name.val
	} else {
		a.Name, _ = RandomName(3)
	}
	return a, nil
}

func MakeChromosome[T comparable](options MakeOptions[T]) (*Chromosome[T], error) {
	c := &Chromosome[T]{}
	if !options.NAlleles.ok() {
		return c, missingParameterError{"options.NAlleles"}
	}
	if !options.NGenes.ok() {
		return c, missingParameterError{"options.NGenes"}
	}
	if !options.NBases.ok() {
		return c, missingParameterError{"options.NBases"}
	}
	if !options.BaseFactory.ok() {
		return c, missingParameterError{"options.BaseFactory"}
	}
	for i := 0; i < int(options.NAlleles.val); i++ {
		a, err := MakeAllele(options)
		if err != nil {
			return c, err
		}
		c.Append(a)
	}
	if options.Name.ok() {
		c.Name = options.Name.val
	} else {
		c.Name, _ = RandomName(3)
	}
	return c, nil
}

func MakeGenome[T comparable](options MakeOptions[T]) (*Genome[T], error) {
	g := &Genome[T]{}
	if !options.NChromosomes.ok() {
		return g, missingParameterError{"options.NAlleles"}
	}
	if !options.NAlleles.ok() {
		return g, missingParameterError{"options.NAlleles"}
	}
	if !options.NGenes.ok() {
		return g, missingParameterError{"options.NGenes"}
	}
	if !options.NBases.ok() {
		return g, missingParameterError{"options.NBases"}
	}
	if !options.BaseFactory.ok() {
		return g, missingParameterError{"options.BaseFactory"}
	}
	for i := 0; i < int(options.NAlleles.val); i++ {
		c, err := MakeChromosome(options)
		if err != nil {
			return g, err
		}
		g.Append(c)
	}
	if options.Name.ok() {
		g.Name = options.Name.val
	} else {
		g.Name, _ = RandomName(3)
	}
	return g, nil
}

func breakSequence[T comparable](sequence []T, separator []T) ([]T, []T) {
	var part []T
	for i, l1, l2 := 0, len(sequence), len(separator); i < l1-l2; i++ {
		part = sequence[i : i+l2]
		if equal(part, separator) {
			return sequence[:i], sequence[i+l2:]
		}
	}
	return []T{}, sequence
}

func GeneFromSequence[T comparable](sequence []T) *Gene[T] {
	g := &Gene[T]{}
	g.Name, _ = RandomName(4)
	g.bases = sequence
	return g
}

func AlleleFromSequence[T comparable](sequence []T, separator []T) *Allele[T] {
	var part []T
	parts := make([][]T, 0)

	part, sequence = breakSequence(sequence, separator)
	for len(part) > 0 {
		parts = append(parts, part)
		part, sequence = breakSequence(sequence, separator)
	}
	parts = append(parts, sequence)

	genes := make([]*Gene[T], 0)
	for _, seq := range parts {
		gene := GeneFromSequence(seq)
		genes = append(genes, gene)
	}

	allele := &Allele[T]{Genes: genes}
	allele.Name, _ = RandomName(3)
	return allele
}

func ChromosomeFromSequence[T comparable](sequence []T, separator []T) *Chromosome[T] {
	var part []T
	parts := make([][]T, 0)
	double_sep := append(separator, separator...)

	part, sequence = breakSequence(sequence, double_sep)
	for len(part) > 0 {
		parts = append(parts, part)
		part, sequence = breakSequence(sequence, double_sep)
	}
	parts = append(parts, sequence)

	alleles := make([]*Allele[T], 0)
	for _, seq := range parts {
		allele := AlleleFromSequence(seq, separator)
		alleles = append(alleles, allele)
	}

	allele := &Chromosome[T]{alleles: alleles}
	allele.Name, _ = RandomName(3)
	return allele
}

func GenomeFromSequence[T comparable](sequence []T, separator []T) *Genome[T] {
	var part []T
	parts := make([][]T, 0)
	triple_sep := append(separator, separator...)
	triple_sep = append(triple_sep, separator...)

	part, sequence = breakSequence(sequence, triple_sep)
	for len(part) > 0 {
		parts = append(parts, part)
		part, sequence = breakSequence(sequence, triple_sep)
	}
	parts = append(parts, sequence)

	chromosomes := make([]*Chromosome[T], 0)
	for _, seq := range parts {
		chromosome := ChromosomeFromSequence(seq, separator)
		chromosomes = append(chromosomes, chromosome)
	}

	genome := &Genome[T]{chromosomes: chromosomes}
	genome.Name, _ = RandomName(3)
	return genome
}
