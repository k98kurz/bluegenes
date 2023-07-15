package bluegenes

import (
	"math"
	"sort"
	"sync"
	"testing"
	"time"
)

type Code[T comparable] interface {
	Gene[T] | Allele[T] | Chromosome[T] | Genome[T]
}

type GeneticMaterial[T comparable] struct {
	gene       Option[*Gene[T]]
	allele     Option[*Allele[T]]
	chromosome Option[*Chromosome[T]]
	genome     Option[*Genome[T]]
}

func (self GeneticMaterial[T]) Recombine(other GeneticMaterial[T], recombinationOpts RecombineOptions) GeneticMaterial[T] {
	child := GeneticMaterial[T]{}
	if self.gene.ok() && other.gene.ok() {
		child.gene.val, _ = self.gene.val.Recombine(other.gene.val, []int{}, recombinationOpts)
		child.gene.isSet = true
	}
	if self.allele.ok() && other.allele.ok() {
		child.allele.val, _ = self.allele.val.Recombine(other.allele.val, []int{}, recombinationOpts)
		child.allele.isSet = true
	}
	if self.chromosome.ok() && other.chromosome.ok() {
		child.chromosome.val, _ = self.chromosome.val.Recombine(other.chromosome.val, []int{}, recombinationOpts)
		child.chromosome.isSet = true
	}
	if self.genome.ok() && other.genome.ok() {
		child.genome.val, _ = self.genome.val.Recombine(other.genome.val, []int{}, recombinationOpts)
		child.genome.isSet = true
	}
	return child
}

func (self GeneticMaterial[T]) copy() GeneticMaterial[T] {
	gm := GeneticMaterial[T]{}
	if self.gene.ok() {
		gm.gene = NewOption(self.gene.val.Copy())
	}
	if self.allele.ok() {
		gm.allele = NewOption(self.allele.val.Copy())
	}
	if self.chromosome.ok() {
		gm.chromosome = NewOption(self.chromosome.val.Copy())
	}
	if self.genome.ok() {
		gm.genome = NewOption(self.genome.val.Copy())
	}
	return gm
}

type OptimizationParams[T comparable, C Code[T]] struct {
	measure_fitness        Option[func(*C) float64]
	mutate                 Option[func(*C)]
	initial_population     Option[[]*C]
	max_iterations         Option[int]
	population_size        Option[int]
	parents_per_generation Option[int]
	fitness_target         Option[float64]
	recombination_opts     Option[RecombineOptions]
	parallel_count         Option[int]
}

type GMOptimizationParams[T comparable] struct {
	measure_fitness        Option[func(GeneticMaterial[T]) float64]
	mutate                 Option[func(GeneticMaterial[T])]
	initial_population     Option[[]GeneticMaterial[T]]
	max_iterations         Option[int]
	population_size        Option[int]
	parents_per_generation Option[int]
	fitness_target         Option[float64]
	recombination_opts     Option[RecombineOptions]
	parallel_count         Option[int]
}

type OptimizationGenchmarkResults struct {
	cost_of_copy            int
	cost_of_mutate          int
	cost_of_measure_fitness int
}

type ScoredGeneticMaterial[T comparable] struct {
	code  GeneticMaterial[T]
	score float64
}

func SortScoredGeneticMaterials[T comparable](scores []ScoredGeneticMaterial[T]) {
	sort.SliceStable(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
}

func RandomChoices[T any](items []T, k int) []T {
	choices := []T{}

	for len(choices) < k {
		i := RandomInt(0, len(items)-1)
		choices = append(choices, items[i])
	}
	return choices
}

func WeightedParentGeneticMaterials[T comparable](scores []ScoredGeneticMaterial[T]) []GeneticMaterial[T] {
	parents := []GeneticMaterial[T]{}
	weight := len(scores)
	for i, l := 0, len(scores); i < l; i++ {
		for j := 0; j < weight; j++ {
			parents = append(parents, scores[i].code)
		}
		weight--
	}
	return parents
}

func WeightedRandomParentGeneticMaterials[T comparable](parents []GeneticMaterial[T]) (GeneticMaterial[T], GeneticMaterial[T]) {
	dad_and_mom := RandomChoices(parents, 2)
	dad := dad_and_mom[0]
	mom := dad_and_mom[1]
	for mom == dad {
		mom = RandomChoices(parents, 1)[0]
	}
	return dad, mom
}

func OptimizeGeneticMaterial[T comparable](params GMOptimizationParams[T]) (int, []ScoredGeneticMaterial[T], error) {
	generation_count := 0
	scores := []ScoredGeneticMaterial[T]{}

	if !params.initial_population.ok() {
		return generation_count, scores, missingParameterError{"params.initial_population"}
	}
	if len(params.initial_population.val) < 1 {
		return generation_count, scores, anError{"params.initial_population must have len > 0"}
	}
	if !params.measure_fitness.ok() {
		return generation_count, scores, missingParameterError{"params.measure_fitness"}
	}
	if !params.mutate.ok() {
		return generation_count, scores, missingParameterError{"params.mutate"}
	}
	if !params.max_iterations.ok() {
		params.max_iterations.val = 1000
	}
	if !params.population_size.ok() {
		params.population_size.val = 100
	}
	if !params.parents_per_generation.ok() {
		params.parents_per_generation.val = 10
	}
	if !params.fitness_target.ok() {
		params.fitness_target.val = float64(1.0)
	}
	if params.parallel_count.ok() && params.population_size.val/params.parallel_count.val < 1 {
		params.population_size.val *= 2
	}
	if params.parents_per_generation.val > params.population_size.val {
		params.parents_per_generation.val = params.population_size.val / 10
	}

	if params.parallel_count.ok() {
		return optimizeGeneticMaterialInParallel(params)
	} else {
		return optimizeGeneticMaterialSequentially(params)
	}
}

func optimizeGeneticMaterialInParallel[T comparable](params GMOptimizationParams[T]) (int, []ScoredGeneticMaterial[T], error) {
	generation_count := 0
	scores := []ScoredGeneticMaterial[T]{}
	var wg sync.WaitGroup
	work_done := make(chan ScoredGeneticMaterial[T], params.population_size.val+10)
	done_signal := make(chan bool, params.parallel_count.val)
	measure_fitness := params.measure_fitness.val
	mutate := params.mutate.val
	for _, genome := range params.initial_population.val {
		score := ScoredGeneticMaterial[T]{code: genome, score: measure_fitness(genome)}
		scores = append(scores, score)
	}
	SortScoredGeneticMaterials(scores)
	best_fitness := scores[0].score

	for generation_count < params.max_iterations.val && best_fitness < params.fitness_target.val {
		generation_count++
		scores = scores[:params.parents_per_generation.val]
		parents := WeightedParentGeneticMaterials(scores)
		children_to_create := (params.population_size.val - params.parents_per_generation.val) / params.parallel_count.val

		for i := params.parallel_count.val; i > 0; i-- {
			diff := 0
			if i == 1 {
				diff = params.population_size.val
				diff -= params.parallel_count.val * children_to_create
				diff -= params.parents_per_generation.val
			}
			wg.Add(1)
			go func(count int, parents []GeneticMaterial[T], work_done chan<- ScoredGeneticMaterial[T], done_signal chan<- bool) {
				defer wg.Done()
				for c := 0; c < count; c++ {
					mom, dad := WeightedRandomParentGeneticMaterials(parents)
					child := dad.Recombine(mom, params.recombination_opts.val)
					mutate(child)
					s := measure_fitness(child)
					work_done <- ScoredGeneticMaterial[T]{code: child, score: s}
				}
				done_signal <- true
			}(children_to_create+diff, parents, work_done, done_signal)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			finished := 0
			for finished < params.parallel_count.val {
				select {
				case child := <-work_done:
					scores = append(scores, child)
				case <-done_signal:
					finished += 1
				default:
				}
			}
		}()

		wg.Wait()
		SortScoredGeneticMaterials(scores)
		best_fitness = scores[0].score
	}

	return generation_count, scores, nil
}

func optimizeGeneticMaterialSequentially[T comparable](params GMOptimizationParams[T]) (int, []ScoredGeneticMaterial[T], error) {
	generation_count := 0
	scores := []ScoredGeneticMaterial[T]{}
	measure_fitness := params.measure_fitness.val
	mutate := params.mutate.val
	for _, code := range params.initial_population.val {
		score := ScoredGeneticMaterial[T]{code: code, score: measure_fitness(code)}
		scores = append(scores, score)
	}
	SortScoredGeneticMaterials(scores)
	best_fitness := scores[0].score

	for generation_count < params.max_iterations.val && best_fitness < params.fitness_target.val {
		generation_count++
		scores = scores[:params.parents_per_generation.val]
		parents := WeightedParentGeneticMaterials(scores)
		for len(scores) < params.population_size.val {
			mom, dad := WeightedRandomParentGeneticMaterials(parents)
			child := dad.Recombine(mom, params.recombination_opts.val)
			mutate(child)
			s := measure_fitness(child)
			scores = append(scores, ScoredGeneticMaterial[T]{code: child, score: s})
		}

		SortScoredGeneticMaterials(scores)
		best_fitness = scores[0].score
	}

	return generation_count, scores, nil
}

func TuneGeneOptimization[T comparable](params OptimizationParams[T, Gene[T]], max_threads ...int) (int, error) {
	n_goroutines := 1
	max_goroutines := 4
	if len(max_threads) > 0 {
		max_goroutines = max_threads[0]
	}
	if !params.initial_population.ok() {
		return n_goroutines, missingParameterError{"params.initial_population"}
	}
	if len(params.initial_population.val) < 1 {
		return n_goroutines, anError{"params.initial_population must have len > 0"}
	}
	if !params.measure_fitness.ok() {
		return n_goroutines, missingParameterError{"params.measure_fitness"}
	}
	if !params.mutate.ok() {
		return n_goroutines, missingParameterError{"params.mutate"}
	}
	if !params.population_size.ok() {
		params.population_size.val = 100
	}
	if !params.parents_per_generation.ok() {
		params.parents_per_generation.val = 10
	}

	res := benchmarkGeneOptimization(params)

	n_goroutines = int(math.Log2(float64((res.cost_of_mutate + res.cost_of_measure_fitness) / res.cost_of_copy)))

	if n_goroutines > max_goroutines {
		n_goroutines = max_goroutines
	} else if n_goroutines <= 0 {
		n_goroutines = 1
	}

	return n_goroutines, nil
}

func benchmarkGeneOptimization[T comparable](params OptimizationParams[T, Gene[T]]) OptimizationGenchmarkResults {
	res := testing.Benchmark(func(b *testing.B) {
		gene := params.initial_population.val[0]
		for i := 0; i < b.N; i++ {
			params.mutate.val(gene)
		}
	})
	cost_of_mutate := res.T / time.Duration(res.N)

	res = testing.Benchmark(func(b *testing.B) {
		gene := params.initial_population.val[0]
		for i := 0; i < b.N; i++ {
			params.measure_fitness.val(gene)
		}
	})
	cost_of_measure_fitness := res.T / time.Duration(res.N)

	res = testing.Benchmark(func(b *testing.B) {
		var bg sync.WaitGroup
		gene := params.initial_population.val[0]
		buffered_chan1 := make(chan *Gene[T], 1)
		buffered_chan2 := make(chan bool, 1)
		for i := 0; i < b.N; i++ {
			bg.Add(2)
			go func(bc1 <-chan *Gene[T], bc2 <-chan bool) {
				defer bg.Done()
				done := false
				for !done {
					select {
					case <-bc1:
					case <-bc2:
						done = true
					default:
					}
				}
			}(buffered_chan1, buffered_chan2)
			go func(bc1 chan<- *Gene[T], bc2 chan<- bool) {
				defer bg.Done()
				bc1 <- gene.Copy()
				bc2 <- true
			}(buffered_chan1, buffered_chan2)
			bg.Wait()
		}
	})
	cost_of_copy := res.T / time.Duration(res.N)

	return OptimizationGenchmarkResults{
		cost_of_mutate:          int(cost_of_mutate),
		cost_of_measure_fitness: int(cost_of_measure_fitness),
		cost_of_copy:            int(cost_of_copy),
	}
}

func TuneAlleleOptimization[T comparable](params OptimizationParams[T, Allele[T]], max_threads ...int) (int, error) {
	n_goroutines := 1
	max_goroutines := 4
	if len(max_threads) > 0 {
		max_goroutines = max_threads[0]
	}
	if !params.initial_population.ok() {
		return n_goroutines, missingParameterError{"params.initial_population"}
	}
	if len(params.initial_population.val) < 1 {
		return n_goroutines, anError{"params.initial_population must have len > 0"}
	}
	if !params.measure_fitness.ok() {
		return n_goroutines, missingParameterError{"params.measure_fitness"}
	}
	if !params.mutate.ok() {
		return n_goroutines, missingParameterError{"params.mutate"}
	}
	if !params.population_size.ok() {
		params.population_size.val = 100
	}
	if !params.parents_per_generation.ok() {
		params.parents_per_generation.val = 10
	}

	res := benchmarkAlleleOptimization(params)

	n_goroutines = int(math.Log2(float64((res.cost_of_mutate + res.cost_of_measure_fitness) / res.cost_of_copy)))

	if n_goroutines > max_goroutines {
		n_goroutines = max_goroutines
	} else if n_goroutines <= 0 {
		n_goroutines = 1
	}

	return n_goroutines, nil
}

func benchmarkAlleleOptimization[T comparable](params OptimizationParams[T, Allele[T]]) OptimizationGenchmarkResults {
	res := testing.Benchmark(func(b *testing.B) {
		allele := params.initial_population.val[0]
		for i := 0; i < b.N; i++ {
			params.mutate.val(allele)
		}
	})
	cost_of_mutate := res.T / time.Duration(res.N)

	res = testing.Benchmark(func(b *testing.B) {
		allele := params.initial_population.val[0]
		for i := 0; i < b.N; i++ {
			params.measure_fitness.val(allele)
		}
	})
	cost_of_measure_fitness := res.T / time.Duration(res.N)

	res = testing.Benchmark(func(b *testing.B) {
		var bg sync.WaitGroup
		allele := params.initial_population.val[0]
		buffered_chan1 := make(chan *Allele[T], 1)
		buffered_chan2 := make(chan bool, 1)
		for i := 0; i < b.N; i++ {
			bg.Add(2)
			go func(bc1 <-chan *Allele[T], bc2 <-chan bool) {
				defer bg.Done()
				done := false
				for !done {
					select {
					case <-bc1:
					case <-bc2:
						done = true
					default:
					}
				}
			}(buffered_chan1, buffered_chan2)
			go func(bc1 chan<- *Allele[T], bc2 chan<- bool) {
				defer bg.Done()
				bc1 <- allele.Copy()
				bc2 <- true
			}(buffered_chan1, buffered_chan2)
			bg.Wait()
		}
	})
	cost_of_copy := res.T / time.Duration(res.N)

	return OptimizationGenchmarkResults{
		cost_of_mutate:          int(cost_of_mutate),
		cost_of_measure_fitness: int(cost_of_measure_fitness),
		cost_of_copy:            int(cost_of_copy),
	}
}

func TuneChromosomeOptimization[T comparable](params OptimizationParams[T, Chromosome[T]], max_threads ...int) (int, error) {
	n_goroutines := 1
	max_goroutines := 4
	if len(max_threads) > 0 {
		max_goroutines = max_threads[0]
	}
	if !params.initial_population.ok() {
		return n_goroutines, missingParameterError{"params.initial_population"}
	}
	if len(params.initial_population.val) < 1 {
		return n_goroutines, anError{"params.initial_population must have len > 0"}
	}
	if !params.measure_fitness.ok() {
		return n_goroutines, missingParameterError{"params.measure_fitness"}
	}
	if !params.mutate.ok() {
		return n_goroutines, missingParameterError{"params.mutate"}
	}
	if !params.population_size.ok() {
		params.population_size.val = 100
	}
	if !params.parents_per_generation.ok() {
		params.parents_per_generation.val = 10
	}

	res := benchmarkChromosomeOptimization(params)

	n_goroutines = int(math.Log2(float64((res.cost_of_mutate + res.cost_of_measure_fitness) / res.cost_of_copy)))

	if n_goroutines > max_goroutines {
		n_goroutines = max_goroutines
	} else if n_goroutines <= 0 {
		n_goroutines = 1
	}

	return n_goroutines, nil
}

func benchmarkChromosomeOptimization[T comparable](params OptimizationParams[T, Chromosome[T]]) OptimizationGenchmarkResults {
	res := testing.Benchmark(func(b *testing.B) {
		chromosome := params.initial_population.val[0]
		for i := 0; i < b.N; i++ {
			params.mutate.val(chromosome)
		}
	})
	cost_of_mutate := res.T / time.Duration(res.N)

	res = testing.Benchmark(func(b *testing.B) {
		chromosome := params.initial_population.val[0]
		for i := 0; i < b.N; i++ {
			params.measure_fitness.val(chromosome)
		}
	})
	cost_of_measure_fitness := res.T / time.Duration(res.N)

	res = testing.Benchmark(func(b *testing.B) {
		var bg sync.WaitGroup
		chromosome := params.initial_population.val[0]
		buffered_chan1 := make(chan *Chromosome[T], 1)
		buffered_chan2 := make(chan bool, 1)
		for i := 0; i < b.N; i++ {
			bg.Add(2)
			go func(bc1 <-chan *Chromosome[T], bc2 <-chan bool) {
				defer bg.Done()
				done := false
				for !done {
					select {
					case <-bc1:
					case <-bc2:
						done = true
					default:
					}
				}
			}(buffered_chan1, buffered_chan2)
			go func(bc1 chan<- *Chromosome[T], bc2 chan<- bool) {
				defer bg.Done()
				bc1 <- chromosome.Copy()
				bc2 <- true
			}(buffered_chan1, buffered_chan2)
			bg.Wait()
		}
	})
	cost_of_copy := res.T / time.Duration(res.N)

	return OptimizationGenchmarkResults{
		cost_of_mutate:          int(cost_of_mutate),
		cost_of_measure_fitness: int(cost_of_measure_fitness),
		cost_of_copy:            int(cost_of_copy),
	}
}

func TuneGenomeOptimization[T comparable](params OptimizationParams[T, Genome[T]], max_threads ...int) (int, error) {
	n_goroutines := 1
	max_goroutines := 4
	if len(max_threads) > 0 {
		max_goroutines = max_threads[0]
	}
	if !params.initial_population.ok() {
		return n_goroutines, missingParameterError{"params.initial_population"}
	}
	if len(params.initial_population.val) < 1 {
		return n_goroutines, anError{"params.initial_population must have len > 0"}
	}
	if !params.measure_fitness.ok() {
		return n_goroutines, missingParameterError{"params.measure_fitness"}
	}
	if !params.mutate.ok() {
		return n_goroutines, missingParameterError{"params.mutate"}
	}
	if !params.population_size.ok() {
		params.population_size.val = 100
	}
	if !params.parents_per_generation.ok() {
		params.parents_per_generation.val = 10
	}

	res := benchmarkGenomeOptimization(params)

	n_goroutines = int(math.Log2(float64((res.cost_of_mutate + res.cost_of_measure_fitness) / res.cost_of_copy)))

	if n_goroutines > max_goroutines {
		n_goroutines = max_goroutines
	} else if n_goroutines <= 0 {
		n_goroutines = 1
	}

	return n_goroutines, nil
}

func benchmarkGenomeOptimization[T comparable](params OptimizationParams[T, Genome[T]]) OptimizationGenchmarkResults {
	res := testing.Benchmark(func(b *testing.B) {
		genome := params.initial_population.val[0]
		for i := 0; i < b.N; i++ {
			params.mutate.val(genome)
		}
	})
	cost_of_mutate := res.T / time.Duration(res.N)

	res = testing.Benchmark(func(b *testing.B) {
		genome := params.initial_population.val[0]
		for i := 0; i < b.N; i++ {
			params.measure_fitness.val(genome)
		}
	})
	cost_of_measure_fitness := res.T / time.Duration(res.N)

	res = testing.Benchmark(func(b *testing.B) {
		var bg sync.WaitGroup
		genome := params.initial_population.val[0]
		buffered_chan1 := make(chan *Genome[T], 1)
		buffered_chan2 := make(chan bool, 1)
		for i := 0; i < b.N; i++ {
			bg.Add(2)
			go func(bc1 <-chan *Genome[T], bc2 <-chan bool) {
				defer bg.Done()
				done := false
				for !done {
					select {
					case <-bc1:
					case <-bc2:
						done = true
					default:
					}
				}
			}(buffered_chan1, buffered_chan2)
			go func(bc1 chan<- *Genome[T], bc2 chan<- bool) {
				defer bg.Done()
				bc1 <- genome.Copy()
				bc2 <- true
			}(buffered_chan1, buffered_chan2)
			bg.Wait()
		}
	})
	cost_of_copy := res.T / time.Duration(res.N)

	return OptimizationGenchmarkResults{
		cost_of_mutate:          int(cost_of_mutate),
		cost_of_measure_fitness: int(cost_of_measure_fitness),
		cost_of_copy:            int(cost_of_copy),
	}
}

func TuneGeneticMaterialOptimization[T comparable](params GMOptimizationParams[T], max_threads ...int) (int, error) {
	n_goroutines := 1
	max_goroutines := 4
	if len(max_threads) > 0 {
		max_goroutines = max_threads[0]
	}
	if !params.initial_population.ok() {
		return n_goroutines, missingParameterError{"params.initial_population"}
	}
	if len(params.initial_population.val) < 1 {
		return n_goroutines, anError{"params.initial_population must have len > 0"}
	}
	if !params.measure_fitness.ok() {
		return n_goroutines, missingParameterError{"params.measure_fitness"}
	}
	if !params.mutate.ok() {
		return n_goroutines, missingParameterError{"params.mutate"}
	}
	if !params.population_size.ok() {
		params.population_size.val = 100
	}
	if !params.parents_per_generation.ok() {
		params.parents_per_generation.val = 10
	}

	res := benchmarkGeneticMaterialOptimization(params)

	n_goroutines = int(math.Log2(float64((res.cost_of_mutate + res.cost_of_measure_fitness) / res.cost_of_copy)))

	if n_goroutines > max_goroutines {
		n_goroutines = max_goroutines
	} else if n_goroutines <= 0 {
		n_goroutines = 1
	}

	return n_goroutines, nil
}

func benchmarkGeneticMaterialOptimization[T comparable](params GMOptimizationParams[T]) OptimizationGenchmarkResults {
	res := testing.Benchmark(func(b *testing.B) {
		gm := params.initial_population.val[0]
		for i := 0; i < b.N; i++ {
			params.mutate.val(gm)
		}
	})
	cost_of_mutate := res.T / time.Duration(res.N)

	res = testing.Benchmark(func(b *testing.B) {
		gm := params.initial_population.val[0]
		for i := 0; i < b.N; i++ {
			params.measure_fitness.val(gm)
		}
	})
	cost_of_measure_fitness := res.T / time.Duration(res.N)

	res = testing.Benchmark(func(b *testing.B) {
		var bg sync.WaitGroup
		gm := params.initial_population.val[0]
		buffered_chan1 := make(chan GeneticMaterial[T], 1)
		buffered_chan2 := make(chan bool, 1)
		for i := 0; i < b.N; i++ {
			bg.Add(2)
			go func(bc1 <-chan GeneticMaterial[T], bc2 <-chan bool) {
				defer bg.Done()
				done := false
				for !done {
					select {
					case <-bc1:
					case <-bc2:
						done = true
					default:
					}
				}
			}(buffered_chan1, buffered_chan2)
			go func(bc1 chan<- GeneticMaterial[T], bc2 chan<- bool) {
				defer bg.Done()
				bc1 <- gm.copy()
				bc2 <- true
			}(buffered_chan1, buffered_chan2)
			bg.Wait()
		}
	})
	cost_of_copy := res.T / time.Duration(res.N)

	return OptimizationGenchmarkResults{
		cost_of_mutate:          int(cost_of_mutate),
		cost_of_measure_fitness: int(cost_of_measure_fitness),
		cost_of_copy:            int(cost_of_copy),
	}
}
