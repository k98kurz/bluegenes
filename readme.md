# Bluegenes

This library is meant to be an easy-to-use library for optimization using
genetic algorithms.

## Status

- [x] Models + tests
- [x] Optimization functions + tests
- [x] Optimization tuner (rough first pass/experimental)
- [x] Optional optimization hook called per-generation
- [x] Neural network structs
- [ ] Neural network training and evolution algorithms
- [ ] Regression models

## Overview

The general concept with a genetic algorithm is to evaluate a population for
fitness (an arbitrary metric) and use fitness, recombination, and Mutation to
drive evolution toward a more optimal result. In this simple library, the
genetic material is a sequence of `T Ordered`, and it is organized into the
following hierarchy:

- `type Gene[T Ordered] struct` contains bases (`T`)
- `type Nucleosome[T Ordered] struct` contains `Gene[T]`s
- `type Chromosome[T Ordered] struct` contains `Nucleosome[T]`s
- `type Genome[T Ordered] struct` contains `Chromosome[T]`s
- `type Code[T Ordered] struct` is a wrapper that contains an `Option` for each above type

Each of these classes except `Code[T]` has a `Name string` attribute to identify
the genetic material and a `Mu sync.RWMutex` for safe concurrent operations. The
names can be generated as random alphanumeric strings if not supplied in the
relevant instantiation statements.

There are functions for creating randomized instances of each:

- `func MakeGeneMakeGene[T Ordered](options MakeOptions[T]) (*Gene[T], error)`
- `func MakeNucleosome[T Ordered](options MakeOptions[T]) (*Nucleosome[T], error)`
- `func MakeChromosome[T Ordered](options MakeOptions[T]) (*Chromosome[T], error)`
- `func MakeGenome[T Ordered](options MakeOptions[T]) (*Genome[T], error)`

And there is a type that combines a `Code[T]` with a fitness Score `float64`:

- `type ScoredCode[T Ordered] struct`

There is one optimization function available currently:

- `func Optimize[T Ordered](params OptimizationParams[T]]) (int, []ScoredCode[T], error)`

And there are two functions available for tuning optimizations given an
`OptimizationParams` instance:

- `func TuneOptimization[T Ordered](params OptimizationParams[T, Gene[T]], max_threads ...int) (int, error)`
- `func BenchmarkOptimization[T Ordered](params OptimizationParams[T]) BenchmarkResult`

The first uses the second to estimate how many goroutines should be used for
optimization by benchmarking the three types of operations and calculating a
ratio. (In some cases, the overhead from spinning up goroutines, heap allocation,
copying data into callstacks, and using synchronization features is more costly
than the `Mutate` and `MeasureFitness` functions, in which case a sequential
optimization will be faster than running the optimization in parallel.)

To handle parameters, the following classes and functions are available:

- `type Option[T any] struct`
- `func NewOption[T any](val ...T) Option[T]`
- `type MakeOptions[T Ordered] struct`
- `type RecombineOptions struct`
- `type OptimizationParams[T Ordered] struct`

See the [Usage](#Usage) section below for more details.

Additionally, a simple neural network system is included:

- `type Neuron struct` contains `Weights []float64`, `Bias float64`, and
`ActivationFunction func(float64) float64`
  - `func (n *Neuron) Activate(inputs []float64) float64`
  - `func NewNeuron(weights []float64, bias float64, activationFunc ...func(float64) float64) Neuron`
- `type Layer struct` contains `Neurons []Neuron`
  - `func (l *Layer) FeedForward(inputs []float64) []float64`
  - `func NewLayer(weights [][]float64, biases []float64, activationFunc ...func(float64) float64) Layer`
- `type Network struct` contains `Layers []Layer`
  - `func (n *Network) FeedForward(inputs []float64) []float64`
  - `func NewNetwork(weights [][][]float64, biases [][]float64, activationFunc ...func(float64) float64) Network`

The Usage section will be updated when more useful/advanced features are added.

## Installation

Installation is with `go get`.

```bash
go get github.com/k98kurz/bluegenes@latest
```

There are no external dependencies.

## Functions and types

Below is a summary of the module contents and how to use them.

### Miscellaneous

- `func RandomName(size int) (string, error)`
- `func RandomInt(min, max int) int`
- `func RandomChoices[T any](items []T, k int) []T`

These randomization functions are used internally to generate randomized names
and provide randomized breeeding and recombination. They can also be used in
functions provided as parameters where relevant, e.g. `MakeOptions.BaseFactory`.

- `type Option[T any] struct`
    - `IsSet bool`
    - `val   T`
    - `func (o Option[T]) ok() bool`
- `func NewOption[T any](val ...T) Option[T]`

This is a type that wraps any value used as a parameter. This is used internally
because `IsSet` initializes as `false` and `NewOption(val)` sets it to `true`,
making it easy to supply required params and omit optional params. For example,
`MakeOptions[int]{BaseFactory: NewOption(someFunc), NBases: NewOption(10)}.`

- `type MakeOptions[T Ordered] struct`
    - `BaseFactory  Option[func() T]`
    - `NBases       Option[uint]`
    - `NGenes       Option[uint]`
    - `NNucleosomes     Option[uint]`
    - `NChromosomes Option[uint]`
    - `Name         Option[string]`

This is a type that is used for calls to `Make{X}`. `BaseFactory` and `NBases`
are required. `NGenes` is required for `MakeNucleosome`, `MakeChromosome`, and
`MakeGenome`. `NNucleosomes` is required for `MakeChromosome` and `MakeGenome`.
`NChromosomes` is required for `MakeGenome`. `Name` is always optional and only
applies to the top level; i.e. when calling `MakeNucleosome` with `Name` specified,
only the `Nucleosome` will have the name, while the `Gene`s will have random names.

- `type RecombineOptions struct`
    - `RecombineGenes       Option[bool]`
    - `MatchGenes           Option[bool]`
    - `RecombineNucleosomes     Option[bool]`
    - `MatchNucleosomes         Option[bool]`
    - `RecombineChromosomes Option[bool]`
    - `MatchChromosomes     Option[bool]`

This controls recombination behavior. All are opt-out; default behavior is to
treat each unspecified value as `true`. When evolving an `Nucleosome`, the underlying
`Gene`s will be recombined unless `RecombineGenes` == `NewOption(false)`, and
they will recombine if they have matching names, if `MatchGenes.IsSet` is
`false`, or if `MatchGenes` == `NewOption(false)`. When evolving a `Chromosome`,
the same logic applies regarding `RecombineNucleosomes` and `MatchNucleosomes`, and the
params are also passed into the calls to `Nucleosome.Recombine`, so the options about
gene recombination also apply. Pattern extends to recombining `Genome`s: the
class doing the recombination checks the params before deciding whether or not
to recombine the underlying subunits of genetic code.

### Gene

- `type Gene[T Ordered] struct`
    - `Name  string`
    - `Bases []T`
    - `Mu    sync.RWMutex`
    - `func (g *Gene[T]) Copy() *Gene[T]`
    - `func (g *Gene[T]) Insert(index int, base T) error`
    - `func (g *Gene[T]) Append(base T) error`
    - `func (g *Gene[T]) InsertSequence(index int, sequence []T) error`
    - `func (g *Gene[T]) Delete(index int) error`
    - `func (g *Gene[T]) DeleteSequence(index int, size int) error`
    - `func (g *Gene[T]) Substitute(index int, base T) error`
    - `func (g *Gene[T]) Recombine(other *Gene[T], indices []int, options RecombineOptions) (*Gene[T], error)`
    - `func (g *Gene[T]) ToMap() map[string][]T`
    - `func (g *Gene[T]) Sequence() []T`
- `func MakeGene[T Ordered](options MakeOptions[T]) (*Gene[T], error)`
- `func GeneFromMap[T Ordered](serialized map[string][]T) *Gene[T]`
- `func GeneFromSequence[T Ordered](sequence []T) *Gene[T]`

The `Gene` represents the smallest meaningful unit of genetic code. Because it
is designed for concurrent operations, it must be allocated on the heap. The
`sync.RWMutex` provides thread safety. `Copy` is used to allocate a copy of the
object with a new `sync.RWMutex` but copied values for `Name` and `Bases`.
`Insert`, `Append`, `InsertSequence`, `Delete`, `DeleteSequence`, and
`Substitute` are mutations that occur in biological systems. `Recombine` mixes
the genetic material with another parent `Gene` and returns a child and an
error if the input was bad. `ToMap` and `GeneFromMap` serialize and deserialize
from a map; the idea was to enable easy JSON compatibility. `Sequence` and
`GeneFromSequence` serialize and deserialize from the underlying slice of `T`,
discarding the `Name` in the process.

### Nucleosome

- `type Nucleosome[T Ordered] struct`
    - `Name  string`
    - `Genes []Gene[T]`
    - `Mu    sync.RWMutex`
    - `func (n *Nucleosome[T]) Copy() *Nucleosome[T]`
    - `func (n *Nucleosome[T]) Insert(index int, base T) error`
    - `func (n *Nucleosome[T]) Append(base T) error`
    - `func (n *Nucleosome[T]) InsertSequence(index int, sequence []T) error`
    - `func (n *Nucleosome[T]) Delete(index int) error`
    - `func (n *Nucleosome[T]) DeleteSequence(index int, size int) error`
    - `func (n *Nucleosome[T]) Substitute(index int, base T) error`
    - `func (n *Nucleosome[T]) Recombine(other *Nucleosome[T], indices []int, options RecombineOptions) (*Nucleosome[T], error)`
    - `func (n *Nucleosome[T]) ToMap() map[string][]T`
    - `func (n *Nucleosome[T]) Sequence(separator []T) []T`
- `func MakeNucleosome[T Ordered](options MakeOptions[T]) (*Nucleosome[T], error)`
- `func NucleosomeFromMap[T Ordered](serialized map[string][]map[string][]T) *Nucleosome[T]`
- `func NucleosomeFromSequence[T Ordered](sequence []T, separator []T) *Nucleosome[T]`

The `Nucleosome` is a collection of related `Gene`s. It has similar features to the
`Gene`, with the notable difference that `Gene`s will be separated by the
supplied `separator []T`.

### Chromosome

- `type Chromosome[T Ordered] struct`
    - `Name  string`
    - `Nucleosomes []Nucleosome[T]`
    - `Mu    sync.RWMutex`
    - `func (c *Chromosome[T]) Copy() *Chromosome[T]`
    - `func (c *Chromosome[T]) Insert(index int, base T) error`
    - `func (c *Chromosome[T]) Append(base T) error`
    - `func (c *Chromosome[T]) InsertSequence(index int, sequence []T) error`
    - `func (c *Chromosome[T]) Delete(index int) error`
    - `func (c *Chromosome[T]) DeleteSequence(index int, size int) error`
    - `func (c *Chromosome[T]) Substitute(index int, base T) error`
    - `func (c *Chromosome[T]) Recombine(other *Chromosome[T], indices []int, options RecombineOptions) (*Chromosome[T], error)`
    - `func (c *Chromosome[T]) ToMap() map[string][]T`
    - `func (c *Chromosome[T]) Sequence(separator []T) []T`
- `func MakeChromosome[T Ordered](options MakeOptions[T]) (*Chromosome[T], error)`
- `func ChromosomeFromMap[T Ordered](serialized map[string][]map[string][]map[string][]T) *Chromosome[T]`
- `func ChromosomeFromSequence[T Ordered](sequence []T, separator []T) *Chromosome[T]`

The `Chromosome` is a collection of `Nucleosome`s, which are separated by double
`separator []T`s when converted to a sequence.

### Genome

- `type Genome[T Ordered] struct`
    - `Name  string`
    - `Chromosomes []Chromosome[T]`
    - `Mu    sync.RWMutex`
    - `func (g *Genome[T]) Copy() *Genome[T]`
    - `func (g *Genome[T]) Insert(index int, base T) error`
    - `func (g *Genome[T]) Append(base T) error`
    - `func (g *Genome[T]) InsertSequence(index int, sequence []T) error`
    - `func (g *Genome[T]) Delete(index int) error`
    - `func (g *Genome[T]) DeleteSequence(index int, size int) error`
    - `func (g *Genome[T]) Substitute(index int, base T) error`
    - `func (g *Genome[T]) Recombine(other *Genome[T], indices []int, options RecombineOptions) (*Genome[T], error)`
    - `func (g *Genome[T]) ToMap() map[string][]T`
    - `func (g *Genome[T]) Sequence(separator []T) []T`
- `func MakeGenome[T Ordered](options MakeOptions[T]) (*Genome[T], error)`
- `func GenomeFromMap[T Ordered](serialized map[string][]map[string][]map[string][]map[string][]T) *Genome[T]`
- `func GenomeFromSequence[T Ordered](sequence []T, separator []T) *Genome[T]`

The `Genome` is a collection of `Chromosome`s, which are separated by triple
`separator []T`s when converted to a sequence.

### Optimization

- `func Optimize[T Ordered](params OptimizationParams[T]) (int, []ScoredCode[T], error)`
- `type OptimizationParams[T Ordered] struct`
    - `MeasureFitness       Option[func(Code[T]) float64]`
    - `Mutate               Option[func(Code[T])]`
    - `InitialPopulation    Option[[]Code[T]]`
    - `MaxIterations        Option[int]`
    - `PopulationSize       Option[int]`
    - `ParentsPerGeneration Option[int]`
    - `FitnessTarget        Option[float64]`
    - `RecombinationOpts    Option[RecombineOptions]`
    - `ParallelCount        Option[int]`
    - `IterationHook        Option[func(int, []ScoredCode[T])]`

This function runs the evolutionary algorithm by scoring each member of the
`params.InitialPopulation` using the `params.MeasureFitness` function, then
orders them by descending `Score`, then breeds them at random using a weighted
distribution (probability for choosing the breeder is
`(len(breeders)-index)/sum(len(breders)-index for all indexes i in breeders)`)
until the `params.PopulationSize` is reached, then mutates and scores every
child, reorders them all by descending score, culls the bottom
`params.PopulationSize-params.ParentPerGeneration`, then repeats steps 3-6 until
either `params.MaxIterations` or `FitnessTarget` is reached.
`params.InitialPopulation`, `.MeasureFitness`, and `.Mutate` are required;
sensible defaults are set for `.MaxIterations`, `.PopulationSize`,
`.ParentsPerGeneration`, and `.FitnessTarget` if they are missing.

I recommend normalizing all fitness scores to the interval [0.0, 1.0] in
`.MeasureFitness`, e.g. `1.0 / (1.0 + total_absolute_error)`, for estimators and
classifiers. For competitions between agents, normalization based upon the total
score of all agents is not possible, so normalization should occur based upon
a theoretical maximum possible score (e.g. `points / max_points`) or with a
difficult goal/threshold (e.g. `points / point_threshold`); alternately, provide
the point threshold in `params.FitnessTarget`. (`.FitnessTarget` defaults to
0.99).

If you want to run the optimization with parallelization, supply the
`params.ParallelCount` as a positive int. Note that if the number of threads
exceeds the population size, the number of threads will be set to half the
population size (i.e. each goroutine will handle breeding, mutating, and
evaluating 2 individuals).

- `type ScoredCode[T Ordered] struct`
    - `Code  Code[T]`
    - `Score float64`
- `type Code[T Ordered] struct`
    - `Gene       Option[*Gene[T]]`
    - `Nucleosome     Option[*Nucleosome[T]]`
    - `Chromosome Option[*Chromosome[T]]`
    - `Genome     Option[*Genome[T]]`
    - `func (c Code[T]) Recombine(other Code[T], recombinationOpts RecombineOptions) Code[T]`
    - `func (c Code[T]) copy() Code[T]`

These are used in the optimization logic and are exported for experimentation
with custom optimization loops, e.g. having agents interact in an environment
for a set amount of time before scoring, culling, breeding, and mutating.

- `func TuneOptimization[T Ordered](params OptimizationParams[T], max_threads ...int) (int, error)`
- `func BenchmarkOptimization[T Ordered](params OptimizationParams[T]) BenchmarkResult`
- `type BenchmarkResult struct`
    - `CostOfCopy           int`
    - `CostOfMutate         int`
    - `CostOfMeasureFitness int`
    - `CostOfIterationHook  int`

`TuneOptimization` uses `BenchmarkOptimization` to estimate the benefit from
running an optimization problem with parallelism. There are cases where the
overhead from synchronization between goroutines (and some additional heap
allocation) may outweigh the costs of running the optimization sequentially.
This will be useful if, for example, you want to make a forecasting model that
uses the last 30 days of weather data as a training set and is reset every day;
since the structure of the data and the functions for mutation and measuring
fitness will be the same, it makes sense to tune the optimization at the outset
and run with an optimal level of parallelization during the daily reset.

Note that this works in the broad sense that it selects parallelism for
workloads that I manually determined through my own benchmarks would benefit
from parallelism, but the exact number of goroutines it suggests should be taken
with a grain of salt. Hence, there is an optional `max_threads` parameter that
will clamp the upper end of its estimate. The lower bound is 1, which means no
parallelism.

## Usage

There are are least three ways to use this library: using an included
optimization function, using a custom optimization function, and using the
genetic classes as the basis for an artificial life siMulation. Below is a
trivial example of how to do the first of these three.

```go
package main

import (
    "fmt"
    "math"
    "math/rand"
    "github.com/k98kurz/bluegenes"
)

var target int = 123456

// Produces a fitness score. Passed as parameter to OptimizeGene.
func measureFitness(code bluegenes.Code[int]) float64 {
    sum := 0
    if !code.Gene.Ok() {
        return 0.0
    }
    for _, b := range code.Gene.Val.Bases {
        sum += b
    }
    return 1.0 / (1.0 + math.Abs(float64(sum - target)))
}

// Mutates a gene at random. Passed as parameter to OptimizeGene.
func mutateCode(code *bluegenes.Code[int]) {
    if !code.Gene.Ok() {
        return
    }
	code.Gene.Val.Mu.Lock()
	defer code.Gene.Val.Mu.Unlock()
	for i := 0; i < len(code.Gene.Val.Bases); i++ {
		val := rand.Float64()
		if val <= 0.1 {
			code.Gene.Val.Bases[i] /= bluegenes.RandomInt(1, 3)
		} else if val <= 0.2 {
			code.Gene.Val.Bases[i] *= bluegenes.RandomInt(1, 3)
		} else {
			code.Gene.Val.Bases[i] += bluegenes.RandomInt(-11, 11)
		}
	}
}

func main() {
	// Gene initialization options
	base_factory := func() int { return bluegenes.RandomInt(-10, 10) }
	opts := bluegenes.MakeOptions[int]{
		NBases:      bluegenes.NewOption(uint(5)),
		BaseFactory: bluegenes.NewOption(base_factory),
	}

	// create initial population
	initial_population := []bluegenes.Code[int]{}
	for i := 0; i < 10; i++ {
		gene, _ := bluegenes.MakeGene(opts)
		initial_population = append(initial_population, bluegenes.Code[int]{Gene: bluegenes.NewOption(gene)})
	}

	// optional: log each iteration
	log_iteration := func(gc int, pop []*bluegenes.ScoredCode[int]) {
		fmt.Printf("generation %d, best score %f\n", gc, pop[0].Score)
	}

	// set up parameters
	params := bluegenes.OptimizationParams[int]{
		InitialPopulation: bluegenes.NewOption(initial_population),
		MeasureFitness:    bluegenes.NewOption(measureFitness),
		Mutate:            bluegenes.NewOption(mutateCode),
		MaxIterations:     bluegenes.NewOption(1000),
	}

	// optional: tune the optimization; not necessary for this trivial example
	parallel_size, err := bluegenes.TuneOptimization(params)

	if err != nil {
		fmt.Println("error encountered during tuning:", err)
	} else if parallel_size > 1 {
		params.ParallelCount = bluegenes.NewOption(parallel_size)
	}

	params.IterationHook = bluegenes.NewOption(log_iteration)

	// run optimization
	n_iterations, final_population, err := bluegenes.Optimize(params)

	best_fitness := final_population[0]
	sum := 0
	for _, b := range best_fitness.Code.Gene.Val.Bases{
		sum += b
	}

	fmt.Printf("%d generations passed\n", n_iterations)
	fmt.Printf("the best result had sum=%d compared to target=%d\n", sum, target)
	fmt.Println(best_fitness.Code.Gene.Val.ToMap())
}
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
- TestNucleosome
    - MakeNucleosome
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
    - Nucleosome
        - parallel
        - sequential
    - Chromosome
        - parallel
        - sequential
    - Genome
        - parallel
        - sequential
    - IterationHook
        - parallel
        - sequential
- TestTuneOptimize
    - Gene
        - cheap
        - expensive
    - Nucleosome
        - cheap
        - expensive
    - Chromosome
        - cheap
        - expensive
    - Genome
        - cheap
        - expensive

The `TestTuneOptimize/*` tests take the most time as the `TuneOptimization`
function runs three benchmarks for each of 8 tests. To run an individual test,
use the following:

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
