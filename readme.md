# Bluegenes

This library is meant to be an easy-to-use library for optimization using
genetic algorithms.

## Status

- [x] Models + tests
- [x] Optimization functions + tests
- [x] Optimization tuner (rough first pass/experimental)
- [ ] Optional optimization hook called per-generation

## Overview

The general concept with a genetic algorithm is to evaluate a population for
fitness (an arbitrary metric) and use fitness, recombination, and mutation to
drive evolution toward a more optimal result. In this simple library, the
genetic material is a sequence of `T comparable`, and it is organized into the
following hierarchy:

- `type Gene[T comparable] struct` contains bases (`T`)
- `type Allele[T comparable] struct` contains `Gene[T]`s
- `type Chromosome[T comparable] struct` contains `Allele[T]`s
- `type Genome[T comparable] struct` contains `Chromosome[T]`s
- `type Code[T comparable] interface` is a constraint type for the above

Each of these classes has a `Name string` attribute to identify the genetic
material and a `mu sync.RWMutex` for safe concurrent operations. The names can
be generated as random alphanumeric strings if not supplied in the relevant
instantiation statements.

There are functions for creating randomized instances of each:

- `func MakeGeneMakeGene[T comparable](options MakeOptions[T]) (*Gene[T], error)`
- `func MakeAllele[T comparable](options MakeOptions[T]) (*Allele[T], error)`
- `func MakeChromosome[T comparable](options MakeOptions[T]) (*Chromosome[T], error)`
- `func MakeGenome[T comparable](options MakeOptions[T]) (*Genome[T], error)`

And there are four types that combine a piece of genetic material with a fitness
score:

- `type ScoredGene[T comparable] struct`
- `type ScoredAllele[T comparable] struct`
- `type ScoredChromosome[T comparable] struct`
- `type ScoredGenome[T comparable] struct`

There are four optimization functions available currently:

- `func OptimizeGene[T comparable](params OptimizationParams[T, Gene[T]]) (int, []ScoredGene[T], error)`
- `func OptimizeAllele[T comparable](params OptimizationParams[T, Allele[T]]) (int, []ScoredAllele[T], error)`
- `func OptimizeChromosome[T comparable](params OptimizationParams[T, Chromosome[T]]) (int, []ScoredChromosome[T], error)`
- `func OptimizeGenome[T comparable](params OptimizationParams[T, Genome[T]]) (int, []ScoredGenome[T], error)`

There are also four functions available for tuning optimizations given an
`OptimizationParams` instance:

- `func TuneGeneOptimization[T comparable](params OptimizationParams[T, Gene[T]], max_threads ...int) (int, error)`
- `func TuneAlleleOptimization[T comparable](params OptimizationParams[T, Allele[T]], max_threads ...int) (int, error)`
- `func TuneChromosomeOptimization[T comparable](params OptimizationParams[T, Chromosome[T]], max_threads ...int) (int, error)`
- `func TuneGenomeOptimization[T comparable](params OptimizationParams[T, Genome[T]], max_threads ...int) (int, error)`

These estimate how many goroutines should be used for optimization by
benchmarking the three types of operations and calculating a ratio. (In some
cases, the overhead from spinning up goroutines, heap allocation, copying data
into callstacks, and using `sync` features is more costly than the `mutate` and
`measure_fitness` functions, in which case a sequential optimization will be
faster than running the optimization in parallel.)

To handle parameters, the following classes and functions are available:

- `type Option[T any] struct`
- `func NewOption[T any](val ...T) Option[T]`
- `type MakeOptions[T comparable] struct`
- `type RecombineOptions struct`
- `type OptimizationParams[T comparable, C Code[T]] struct`

See the [Usage](#Usage) section below for more details.

## Installation

Installation is with `go get`.

```bash
go get
```

There are no external dependencies.

## Usage

There are are least three ways to use this library: using an included
optimization function, using a custom optimization function, and using the
genetic classes as the basis for an artificial life simulation. Below is a
trivial example of how to do the first of these three.

```go
package main

import (
    "math"
    "github.com/k98kurz/gobluegenes"
)

target := 123456

// Produces a fitness score. Passed as parameter to OptimizeGene.
func measureFitness(gene *Gene[int]) float64 {
    sum := 0
    for _, b := gene.bases {
        sum += b
    }
    return 1.0 / (1.0 + math.Abs(float64(sum - target)))
}

// Mutates a gene at random. Passed as parameter to OptimizeGene.
func mutateGene(gene *Gene[int]) {
	gene.mu.Lock()
	defer gene.mu.Unlock()
	for i := 0; i < len(gene.bases); i++ {
		val := rand.Float64()
		if val <= 0.1 {
			gene.bases[i] /= gobluegenes.RandomInt(1, 3)
		} else if val <= 0.2 {
			gene.bases[i] *= gobluegenes.RandomInt(1, 3)
		} else if val <= 0.6 {
			gene.bases[i] += gobluegenes.RandomInt(0, 11)
		} else {
			gene.bases[i] -= gobluegenes.RandomInt(0, 11)
		}
	}
}

// Gene initialization options
base_factory := func() int { return RandomInt(-10, 10) }
opts := gobluegenes.MakeOptions[int]{
	NBases:      gobluegenes.NewOption(uint(5)),
	BaseFactory: gobluegenes.NewOption(base_factory),
}

// create initial population
initial_population := []*gobluegenes.Gene[int]{}
for i := 0; i < 10; i++ {
	gene, _ := gobluegenes.MakeGene(opts)
	initial_population = append(initial_population, gene)
}

// set up parameters
params := OptimizationParams[int, gobluegenes.Gene[int]]{
	initial_population: gobluegenes.NewOption(initial_population),
	measure_fitness:    gobluegenes.NewOption(MeasureGeneFitness),
	mutate:             gobluegenes.NewOption(MutateGene),
	max_iterations:     gobluegenes.NewOption(1000),
}

// optional: tune the optimization; not necessary for this trivial example
parallel_size, err := gobluegenes.TuneGeneOptimization(params)

if err != nil {
    fmt.Println("error encountered during tuning:", err)
} else if parallel_size > 1 {
    params.parallel_count = gobluegenes.NewOption(parallel_size)
}

// run optimization
n_iterations, final_population, err := gobluegenes.OptimizeGene(params)

best_fitness := final_population[0]
sum := 0
for _, b := range best_fitness.gene{
	sum += b
}

fmt.Printf("%d generations passed", n_iterations)
fmt.Printf("the best result had sum=%d compared to target=%d", sum, target)
fmt.Println(best_fitness.gene)
```

Creating custom fitness functions or artificial life simulations is left as an
exercise to the reader.

## Testing

To test, clone the repository and run the following:

```bash
go test -v
```

The following tests are included:

- TestGene
    - MakeGene
    - Append
    - Copy
    - Delete
    - DeleteSequence
    - Insert
    - InsertSequence
    - Recombine
    - Substitute
    - ToMap
    - Sequence
- TestAllele
    - MakeAllele
    - Append
    - Copy
    - Delete
    - DeleteSequence
    - Insert
    - InsertSequence
    - Recombine
    - Substitute
    - ToMap
    - Sequence
- TestChromosome
    - MakeChromosome
    - Append
    - Copy
    - Delete
    - DeleteSequence
    - Insert
    - InsertSequence
    - Recombine
    - Substitute
    - ToMap
    - Sequence
- TestGenome
    - MakeGenome
    - Append
    - Copy
    - Delete
    - DeleteSequence
    - Insert
    - InsertSequence
    - Recombine
    - Substitute
    - ToMap
    - Sequence
- TestOptimize
    - Gene
        - parallel
        - sequential
    - Allele
        - parallel
        - sequential
    - Chromosome
        - parallel
        - sequential
    - Genome
        - parallel
        - sequential
- TestTuneOptimize
    - Gene
        - cheap
        - expensive
    - Allele
        - cheap
        - expensive
    - Chromosome
        - cheap
        - expensive
    - Genome
        - cheap
        - expensive

The `TestTuneOptimize/*` tests take the most time as the `Tune[C]Optimization`
functions each run three benchmarks. To run an individual test, use the
following:

```bash
go test -run Test/Name -v
```

For example: `go test -run TestTuneOptimize/Gene/cheap`

## ISC License

ISC License

Copyleft (c) 2023 k98kurz

Permission to use, copy, modify, and/or distribute this software
for any purpose with or without fee is hereby granted, provided
that the above copyleft notice and this permission notice appear in
all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL
WARRANTIES WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE
AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT, INDIRECT, OR
CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS
OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT,
NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF OR IN
CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
