package genetics

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
		return "", Error{"size must be > 0"}
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
	n_bases       Option[uint]
	n_genes       Option[uint]
	n_alleles     Option[uint]
	n_chromosomes Option[uint]
	name          Option[string]
	base_factory  Option[func() T]
}

type RecombineOptions struct {
	recombine_genes       Option[bool]
	match_genes           Option[bool]
	recombine_alleles     Option[bool]
	match_alleles         Option[bool]
	recombine_chromosomes Option[bool]
	match_chromosomes     Option[bool]
}

func MakeGene[T comparable](options MakeOptions[T]) (*Gene[T], error) {
	g := &Gene[T]{}
	if !options.n_bases.ok() {
		return g, MissingParameterError{"options.n_bases"}
	}
	if !options.base_factory.ok() {
		return g, MissingParameterError{"options.base_factory"}
	}
	for i := 0; i < int(options.n_bases.val); i++ {
		b := options.base_factory.val()
		g.Append(b)
	}
	if options.name.ok() {
		g.name = options.name.val
	} else {
		g.name, _ = RandomName(4)
	}
	return g, nil
}

func MakeAllele[T comparable](options MakeOptions[T]) (*Allele[T], error) {
	a := &Allele[T]{}
	if !options.n_genes.ok() {
		return a, MissingParameterError{"options.n_genes"}
	}
	if !options.n_bases.ok() {
		return a, MissingParameterError{"options.n_bases"}
	}
	if !options.base_factory.ok() {
		return a, MissingParameterError{"options.base_factory"}
	}
	for i := 0; i < int(options.n_genes.val); i++ {
		g, err := MakeGene(options)
		if err != nil {
			return a, err
		}
		a.Append(g)
	}
	if options.name.ok() {
		a.name = options.name.val
	} else {
		a.name, _ = RandomName(3)
	}
	return a, nil
}

func MakeChromosome[T comparable](options MakeOptions[T]) (*Chromosome[T], error) {
	c := &Chromosome[T]{}
	if !options.n_alleles.ok() {
		return c, MissingParameterError{"options.n_alleles"}
	}
	if !options.n_genes.ok() {
		return c, MissingParameterError{"options.n_genes"}
	}
	if !options.n_bases.ok() {
		return c, MissingParameterError{"options.n_bases"}
	}
	if !options.base_factory.ok() {
		return c, MissingParameterError{"options.base_factory"}
	}
	for i := 0; i < int(options.n_alleles.val); i++ {
		a, err := MakeAllele(options)
		if err != nil {
			return c, err
		}
		c.Append(a)
	}
	if options.name.ok() {
		c.name = options.name.val
	} else {
		c.name, _ = RandomName(3)
	}
	return c, nil
}

func MakeGenome[T comparable](options MakeOptions[T]) (*Genome[T], error) {
	g := &Genome[T]{}
	if !options.n_chromosomes.ok() {
		return g, MissingParameterError{"options.n_alleles"}
	}
	if !options.n_alleles.ok() {
		return g, MissingParameterError{"options.n_alleles"}
	}
	if !options.n_genes.ok() {
		return g, MissingParameterError{"options.n_genes"}
	}
	if !options.n_bases.ok() {
		return g, MissingParameterError{"options.n_bases"}
	}
	if !options.base_factory.ok() {
		return g, MissingParameterError{"options.base_factory"}
	}
	for i := 0; i < int(options.n_alleles.val); i++ {
		c, err := MakeChromosome(options)
		if err != nil {
			return g, err
		}
		g.Append(c)
	}
	if options.name.ok() {
		g.name = options.name.val
	} else {
		g.name, _ = RandomName(3)
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
	g.name, _ = RandomName(4)
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

	allele := &Allele[T]{genes: genes}
	allele.name, _ = RandomName(3)
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
	allele.name, _ = RandomName(3)
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
	genome.name, _ = RandomName(3)
	return genome
}
