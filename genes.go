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
		return "", anError{"size Must be > 0"}
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

type MakeOptions[T Ordered] struct {
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
	RecombineGenomes     Option[bool]
}

func MakeGene[T Ordered](options MakeOptions[T]) (*Gene[T], error) {
	g := &Gene[T]{}
	if !options.NBases.Ok() {
		return g, missingParameterError{"options.NBases"}
	}
	if !options.BaseFactory.Ok() {
		return g, missingParameterError{"options.BaseFactory"}
	}
	for i := 0; i < int(options.NBases.Val); i++ {
		b := options.BaseFactory.Val()
		g.Append(b)
	}
	if options.Name.Ok() {
		g.Name = options.Name.Val
	} else {
		g.Name, _ = RandomName(4)
	}
	return g, nil
}

func MakeAllele[T Ordered](options MakeOptions[T]) (*Allele[T], error) {
	a := &Allele[T]{}
	if !options.NGenes.Ok() {
		return a, missingParameterError{"options.NGenes"}
	}
	if !options.NBases.Ok() {
		return a, missingParameterError{"options.NBases"}
	}
	if !options.BaseFactory.Ok() {
		return a, missingParameterError{"options.BaseFactory"}
	}
	for i := 0; i < int(options.NGenes.Val); i++ {
		g, err := MakeGene(options)
		if err != nil {
			return a, err
		}
		a.Append(g)
	}
	if options.Name.Ok() {
		a.Name = options.Name.Val
	} else {
		a.Name, _ = RandomName(3)
	}
	return a, nil
}

func MakeChromosome[T Ordered](options MakeOptions[T]) (*Chromosome[T], error) {
	c := &Chromosome[T]{}
	if !options.NAlleles.Ok() {
		return c, missingParameterError{"options.NAlleles"}
	}
	if !options.NGenes.Ok() {
		return c, missingParameterError{"options.NGenes"}
	}
	if !options.NBases.Ok() {
		return c, missingParameterError{"options.NBases"}
	}
	if !options.BaseFactory.Ok() {
		return c, missingParameterError{"options.BaseFactory"}
	}
	for i := 0; i < int(options.NAlleles.Val); i++ {
		a, err := MakeAllele(options)
		if err != nil {
			return c, err
		}
		c.Append(a)
	}
	if options.Name.Ok() {
		c.Name = options.Name.Val
	} else {
		c.Name, _ = RandomName(2)
	}
	return c, nil
}

func MakeGenome[T Ordered](options MakeOptions[T]) (*Genome[T], error) {
	g := &Genome[T]{}
	if !options.NChromosomes.Ok() {
		return g, missingParameterError{"options.NAlleles"}
	}
	if !options.NAlleles.Ok() {
		return g, missingParameterError{"options.NAlleles"}
	}
	if !options.NGenes.Ok() {
		return g, missingParameterError{"options.NGenes"}
	}
	if !options.NBases.Ok() {
		return g, missingParameterError{"options.NBases"}
	}
	if !options.BaseFactory.Ok() {
		return g, missingParameterError{"options.BaseFactory"}
	}
	for i := 0; i < int(options.NAlleles.Val); i++ {
		c, err := MakeChromosome(options)
		if err != nil {
			return g, err
		}
		g.Append(c)
	}
	if options.Name.Ok() {
		g.Name = options.Name.Val
	} else {
		g.Name, _ = RandomName(3)
	}
	return g, nil
}

func breakSequence[T Ordered](sequence []T, separator []T) ([]T, []T) {
	var part []T
	for i, l1, l2 := 0, len(sequence), len(separator); i < l1-l2; i++ {
		part = sequence[i : i+l2]
		if equal(part, separator) {
			return sequence[:i], sequence[i+l2:]
		}
	}
	return []T{}, sequence
}

func inverseSequence[T Ordered](separator []T) []T {
	result := []T{}
	var v interface{}
	if len(separator) == 0 {
		return result
	}
	for i := 0; i < len(separator); i++ {
		v = separator[i]
		switch any(v).(type) {
		case int:
			v = ^v.(int)
		case int8:
			v = ^v.(int8)
		case int16:
			v = ^v.(int16)
		case int32:
			v = ^v.(int32)
		case int64:
			v = ^v.(int64)
		case uint:
			v = ^v.(uint)
		case uint8:
			v = ^v.(uint8)
		case uint16:
			v = ^v.(uint16)
		case uint32:
			v = ^v.(uint32)
		case uint64:
			v = ^v.(uint64)
		case float32:
			v = flipFloat32(v.(float32))
		case float64:
			v = flipFloat64(v.(float64))
		case string:
			v = flipString(v.(string))
		}
		result = append(result, v.(T))
	}
	return result
}

func GeneFromMap[T Ordered](serialized map[string][]T) *Gene[T] {
	g := Gene[T]{}
	for k, v := range serialized {
		g.Name = k
		g.Bases = v
	}

	return &g
}

func GeneFromSequence[T Ordered](sequence []T) *Gene[T] {
	g := &Gene[T]{}
	g.Name, _ = RandomName(4)
	g.Bases = sequence
	return g
}

func AlleleFromMap[T Ordered](serialized map[string][]map[string][]T) *Allele[T] {
	a := Allele[T]{}

	for k1, v1 := range serialized {
		a.Name = k1
		for _, gene := range v1 {
			a.Genes = append(a.Genes, GeneFromMap(gene))
		}
	}

	return &a
}

func AlleleFromSequence[T Ordered](sequence []T, separator []T, placeholder ...[]T) *Allele[T] {
	var part []T
	parts := make([][]T, 0)
	var realPlaceholder []T
	if len(placeholder) > 0 {
		realPlaceholder = placeholder[0]
	} else {
		realPlaceholder = inverseSequence(separator)
	}

	part, sequence = breakSequence(sequence, separator)
	for len(part) > 0 {
		parts = append(parts, part)
		part, sequence = breakSequence(sequence, separator)
	}
	parts = append(parts, sequence)

	genes := make([]*Gene[T], 0)
	var gene *Gene[T]
	for _, seq := range parts {
		if equal(seq, realPlaceholder) {
			name, _ := RandomName(4)
			gene = &Gene[T]{Name: name}
		} else {
			gene = GeneFromSequence(seq)
		}
		genes = append(genes, gene)
	}

	allele := &Allele[T]{Genes: genes}
	allele.Name, _ = RandomName(3)
	return allele
}

func ChromosomeFromMap[T Ordered](serialized map[string][]map[string][]map[string][]T) *Chromosome[T] {
	c := Chromosome[T]{}

	for k1, v1 := range serialized {
		c.Name = k1
		for _, chromosome := range v1 {
			c.Alleles = append(c.Alleles, AlleleFromMap(chromosome))
		}
	}

	return &c
}

func ChromosomeFromSequence[T Ordered](sequence []T, separator []T, placeholder ...[]T) *Chromosome[T] {
	var part []T
	parts := make([][]T, 0)
	double_sep := append(separator, separator...)
	var realPlaceholder []T
	if len(placeholder) > 0 {
		realPlaceholder = placeholder[0]
	} else {
		realPlaceholder = inverseSequence(separator)
	}
	realPlaceholder = append(realPlaceholder, realPlaceholder...)

	part, sequence = breakSequence(sequence, double_sep)
	for len(part) > 0 {
		parts = append(parts, part)
		part, sequence = breakSequence(sequence, double_sep)
	}
	parts = append(parts, sequence)

	alleles := make([]*Allele[T], 0)
	var allele *Allele[T]
	for _, seq := range parts {
		if equal(seq, realPlaceholder) {
			name, _ := RandomName(3)
			allele = &Allele[T]{Name: name}
		} else {
			allele = AlleleFromSequence(seq, separator, placeholder...)
		}
		alleles = append(alleles, allele)
	}

	chromosome := &Chromosome[T]{Alleles: alleles}
	chromosome.Name, _ = RandomName(2)
	return chromosome
}

func GenomeFromMap[T Ordered](serialized map[string][]map[string][]map[string][]map[string][]T) *Genome[T] {
	g := Genome[T]{}

	for k1, v1 := range serialized {
		g.Name = k1
		for _, chromosome := range v1 {
			g.Chromosomes = append(g.Chromosomes, ChromosomeFromMap(chromosome))
		}
	}

	return &g
}

func GenomeFromSequence[T Ordered](sequence []T, separator []T, placeholder ...[]T) *Genome[T] {
	var part []T
	parts := make([][]T, 0)
	triple_sep := append(separator, separator...)
	triple_sep = append(triple_sep, separator...)
	var realPlaceholder []T
	if len(placeholder) > 0 {
		realPlaceholder = placeholder[0]
	} else {
		realPlaceholder = inverseSequence(separator)
	}
	realPlaceholder = append(realPlaceholder, realPlaceholder...)
	realPlaceholder = append(realPlaceholder, realPlaceholder...)

	part, sequence = breakSequence(sequence, triple_sep)
	for len(part) > 0 {
		parts = append(parts, part)
		part, sequence = breakSequence(sequence, triple_sep)
	}
	parts = append(parts, sequence)

	chromosomes := make([]*Chromosome[T], 0)
	var chromosome *Chromosome[T]
	for _, seq := range parts {
		if equal(seq, realPlaceholder) {
			name, _ := RandomName(2)
			chromosome = &Chromosome[T]{Name: name}
		} else {
			chromosome = ChromosomeFromSequence(seq, separator, placeholder...)
		}
		chromosomes = append(chromosomes, chromosome)
	}

	genome := &Genome[T]{Chromosomes: chromosomes}
	genome.Name, _ = RandomName(6)
	return genome
}
