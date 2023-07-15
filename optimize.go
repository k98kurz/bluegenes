package genetics

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

type OptimizationGenchmarkResults struct {
	cost_of_copy            int
	cost_of_mutate          int
	cost_of_measure_fitness int
}

type ScoredGene[T comparable] struct {
	gene  *Gene[T]
	score float64
}

type ScoredAllele[T comparable] struct {
	allele *Allele[T]
	score  float64
}

type ScoredChromosome[T comparable] struct {
	chromosome *Chromosome[T]
	score      float64
}

type ScoredGenome[T comparable] struct {
	genome *Genome[T]
	score  float64
}

func SortScoredGenes[T comparable](scores []ScoredGene[T]) {
	sort.SliceStable(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
}

func SortScoredAlleles[T comparable](scores []ScoredAllele[T]) {
	sort.SliceStable(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
}

func SortScoredChromosomes[T comparable](scores []ScoredChromosome[T]) {
	sort.SliceStable(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
}

func SortScoredGenomes[T comparable](scores []ScoredGenome[T]) {
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

func WeightedParentGenes[T comparable](scores []ScoredGene[T]) []*Gene[T] {
	parents := []*Gene[T]{}
	weight := len(scores)
	for i, l := 0, len(scores); i < l; i++ {
		for j := 0; j < weight; j++ {
			parents = append(parents, scores[i].gene)
		}
		weight--
	}
	return parents
}

func WeightedRandomParentGenes[T comparable](parents []*Gene[T]) (*Gene[T], *Gene[T]) {
	dad_and_mom := RandomChoices(parents, 2)
	dad := dad_and_mom[0]
	mom := dad_and_mom[1]
	for mom == dad {
		mom = RandomChoices(parents, 1)[0]
	}
	return dad, mom
}

func WeightedParentAlleles[T comparable](scores []ScoredAllele[T]) []*Allele[T] {
	parents := []*Allele[T]{}
	weight := len(scores)
	for i, l := 0, len(scores); i < l; i++ {
		for j := 0; j < weight; j++ {
			parents = append(parents, scores[i].allele)
		}
		weight--
	}
	return parents
}

func WeightedRandomParentAlleles[T comparable](parents []*Allele[T]) (*Allele[T], *Allele[T]) {
	dad_and_mom := RandomChoices(parents, 2)
	dad := dad_and_mom[0]
	mom := dad_and_mom[1]
	for mom == dad {
		mom = RandomChoices(parents, 1)[0]
	}
	return dad, mom
}

func WeightedParentChromosomes[T comparable](scores []ScoredChromosome[T]) []*Chromosome[T] {
	parents := []*Chromosome[T]{}
	weight := len(scores)
	for i, l := 0, len(scores); i < l; i++ {
		for j := 0; j < weight; j++ {
			parents = append(parents, scores[i].chromosome)
		}
		weight--
	}
	return parents
}

func WeightedRandomParentChromosomes[T comparable](parents []*Chromosome[T]) (*Chromosome[T], *Chromosome[T]) {
	dad_and_mom := RandomChoices(parents, 2)
	dad := dad_and_mom[0]
	mom := dad_and_mom[1]
	for mom == dad {
		mom = RandomChoices(parents, 1)[0]
	}
	return dad, mom
}

func WeightedParentGenomes[T comparable](scores []ScoredGenome[T]) []*Genome[T] {
	parents := []*Genome[T]{}
	weight := len(scores)
	for i, l := 0, len(scores); i < l; i++ {
		for j := 0; j < weight; j++ {
			parents = append(parents, scores[i].genome)
		}
		weight--
	}
	return parents
}

func WeightedRandomParentGenomes[T comparable](parents []*Genome[T]) (*Genome[T], *Genome[T]) {
	dad_and_mom := RandomChoices(parents, 2)
	dad := dad_and_mom[0]
	mom := dad_and_mom[1]
	for mom == dad {
		mom = RandomChoices(parents, 1)[0]
	}
	return dad, mom
}

func OptimizeGene[T comparable](params OptimizationParams[T, Gene[T]]) (int, []ScoredGene[T], error) {
	generation_count := 0
	scores := []ScoredGene[T]{}

	if !params.initial_population.ok() {
		return generation_count, scores, MissingParameterError{"params.initial_population"}
	}
	if len(params.initial_population.val) < 1 {
		return generation_count, scores, Error{"params.initial_population must have len > 0"}
	}
	if !params.measure_fitness.ok() {
		return generation_count, scores, MissingParameterError{"params.measure_fitness"}
	}
	if !params.mutate.ok() {
		return generation_count, scores, MissingParameterError{"params.mutate"}
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
		return optimizeGeneInParallel(params)
	} else {
		return optimizeGeneSequentially(params)
	}
}

func optimizeGeneInParallel[T comparable](params OptimizationParams[T, Gene[T]]) (int, []ScoredGene[T], error) {
	generation_count := 0
	scores := []ScoredGene[T]{}

	var wg sync.WaitGroup
	work_done := make(chan ScoredGene[T], params.population_size.val+10)
	done_signal := make(chan bool, params.parallel_count.val)
	measure_fitness := params.measure_fitness.val
	mutate := params.mutate.val
	for _, gene := range params.initial_population.val {
		score := ScoredGene[T]{gene: gene, score: measure_fitness(gene)}
		scores = append(scores, score)
	}
	SortScoredGenes(scores)
	best_fitness := scores[0].score

	for generation_count < params.max_iterations.val && best_fitness < params.fitness_target.val {
		generation_count++
		scores = scores[:params.parents_per_generation.val]
		parents := WeightedParentGenes(scores)
		children_to_create := (params.population_size.val - params.parents_per_generation.val) / params.parallel_count.val

		for i := params.parallel_count.val; i > 0; i-- {
			diff := 0
			if i == 1 {
				diff = params.population_size.val
				diff -= params.parallel_count.val * children_to_create
				diff -= params.parents_per_generation.val
			}
			wg.Add(1)
			go func(count int, parents []*Gene[T], work_done chan<- ScoredGene[T], done_signal chan<- bool) {
				defer wg.Done()
				for c := 0; c < count; c++ {
					mom, dad := WeightedRandomParentGenes(parents)
					child, _ := dad.Recombine(mom, []int{}, params.recombination_opts.val)
					mutate(child)
					s := measure_fitness(child)
					work_done <- ScoredGene[T]{gene: child, score: s}
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
		SortScoredGenes(scores)
		best_fitness = scores[0].score
	}
	return generation_count, scores, nil
}

func optimizeGeneSequentially[T comparable](params OptimizationParams[T, Gene[T]]) (int, []ScoredGene[T], error) {
	generation_count := 0
	scores := []ScoredGene[T]{}

	measure_fitness := params.measure_fitness.val
	mutate := params.mutate.val
	for _, gene := range params.initial_population.val {
		score := ScoredGene[T]{gene: gene, score: measure_fitness(gene)}
		scores = append(scores, score)
	}
	SortScoredGenes(scores)
	best_fitness := scores[0].score

	for generation_count < params.max_iterations.val && best_fitness < params.fitness_target.val {
		generation_count++
		scores = scores[:params.parents_per_generation.val]
		parents := WeightedParentGenes(scores)

		for len(scores) < params.population_size.val {
			mom, dad := WeightedRandomParentGenes(parents)
			child, _ := dad.Recombine(mom, []int{}, params.recombination_opts.val)
			mutate(child)
			s := measure_fitness(child)
			scores = append(scores, ScoredGene[T]{gene: child, score: s})
		}

		SortScoredGenes(scores)
		best_fitness = scores[0].score
	}
	return generation_count, scores, nil
}

func OptimizeAllele[T comparable](params OptimizationParams[T, Allele[T]]) (int, []ScoredAllele[T], error) {
	generation_count := 0
	scores := []ScoredAllele[T]{}

	if !params.initial_population.ok() {
		return generation_count, scores, MissingParameterError{"params.initial_population"}
	}
	if len(params.initial_population.val) < 1 {
		return generation_count, scores, Error{"params.initial_population must have len > 0"}
	}
	if !params.measure_fitness.ok() {
		return generation_count, scores, MissingParameterError{"params.measure_fitness"}
	}
	if !params.mutate.ok() {
		return generation_count, scores, MissingParameterError{"params.mutate"}
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
		return optimizeAlleleInParallel(params)
	} else {
		return optimizeAlleleSequentially(params)
	}
}

func optimizeAlleleInParallel[T comparable](params OptimizationParams[T, Allele[T]]) (int, []ScoredAllele[T], error) {
	var wg sync.WaitGroup
	generation_count := 0
	scores := []ScoredAllele[T]{}
	work_done := make(chan ScoredAllele[T], params.population_size.val+10)
	done_signal := make(chan bool, params.parallel_count.val)
	measure_fitness := params.measure_fitness.val
	mutate := params.mutate.val
	for _, allele := range params.initial_population.val {
		score := ScoredAllele[T]{allele: allele, score: measure_fitness(allele)}
		scores = append(scores, score)
	}
	SortScoredAlleles(scores)
	best_fitness := scores[0].score

	for generation_count < params.max_iterations.val && best_fitness < params.fitness_target.val {
		generation_count++
		scores = scores[:params.parents_per_generation.val]
		parents := WeightedParentAlleles(scores)
		children_to_create := (params.population_size.val - params.parents_per_generation.val) / params.parallel_count.val

		for i := params.parallel_count.val; i > 0; i-- {
			diff := 0
			if i == 1 {
				diff = params.population_size.val
				diff -= params.parallel_count.val * children_to_create
				diff -= params.parents_per_generation.val
			}
			wg.Add(1)
			go func(count int, parents []*Allele[T], work_done chan<- ScoredAllele[T], done_signal chan<- bool) {
				defer wg.Done()
				for c := 0; c < count; c++ {
					mom, dad := WeightedRandomParentAlleles(parents)
					child, _ := dad.Recombine(mom, []int{}, params.recombination_opts.val)
					mutate(child)
					s := measure_fitness(child)
					work_done <- ScoredAllele[T]{allele: child, score: s}
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

		SortScoredAlleles(scores)
		best_fitness = scores[0].score
	}

	return generation_count, scores, nil
}

func optimizeAlleleSequentially[T comparable](params OptimizationParams[T, Allele[T]]) (int, []ScoredAllele[T], error) {
	generation_count := 0
	scores := []ScoredAllele[T]{}
	measure_fitness := params.measure_fitness.val
	mutate := params.mutate.val
	for _, allele := range params.initial_population.val {
		score := ScoredAllele[T]{allele: allele, score: measure_fitness(allele)}
		scores = append(scores, score)
	}
	SortScoredAlleles(scores)
	best_fitness := scores[0].score

	for generation_count < params.max_iterations.val && best_fitness < params.fitness_target.val {
		generation_count++
		scores = scores[:params.parents_per_generation.val]
		parents := WeightedParentAlleles(scores)
		for len(scores) < params.population_size.val {
			mom, dad := WeightedRandomParentAlleles(parents)
			child, _ := dad.Recombine(mom, []int{}, params.recombination_opts.val)
			mutate(child)
			s := measure_fitness(child)
			scores = append(scores, ScoredAllele[T]{allele: child, score: s})
		}

		SortScoredAlleles(scores)
		best_fitness = scores[0].score
	}

	return generation_count, scores, nil
}

func OptimizeChromosome[T comparable](params OptimizationParams[T, Chromosome[T]]) (int, []ScoredChromosome[T], error) {
	generation_count := 0
	scores := []ScoredChromosome[T]{}

	if !params.initial_population.ok() {
		return generation_count, scores, MissingParameterError{"params.initial_population"}
	}
	if len(params.initial_population.val) < 1 {
		return generation_count, scores, Error{"params.initial_population must have len > 0"}
	}
	if !params.measure_fitness.ok() {
		return generation_count, scores, MissingParameterError{"params.measure_fitness"}
	}
	if !params.mutate.ok() {
		return generation_count, scores, MissingParameterError{"params.mutate"}
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
		return optimizeChromosomeInParallel(params)
	} else {
		return optimizeChromosomeSequentially(params)
	}
}

func optimizeChromosomeInParallel[T comparable](params OptimizationParams[T, Chromosome[T]]) (int, []ScoredChromosome[T], error) {
	generation_count := 0
	scores := []ScoredChromosome[T]{}
	var wg sync.WaitGroup
	work_done := make(chan ScoredChromosome[T], params.population_size.val+10)
	done_signal := make(chan bool, params.parallel_count.val)
	measure_fitness := params.measure_fitness.val
	mutate := params.mutate.val
	for _, chromosome := range params.initial_population.val {
		score := ScoredChromosome[T]{chromosome: chromosome, score: measure_fitness(chromosome)}
		scores = append(scores, score)
	}
	SortScoredChromosomes(scores)
	best_fitness := scores[0].score

	for generation_count < params.max_iterations.val && best_fitness < params.fitness_target.val {
		generation_count++
		scores = scores[:params.parents_per_generation.val]
		parents := WeightedParentChromosomes(scores)
		children_to_create := (params.population_size.val - params.parents_per_generation.val) / params.parallel_count.val

		for i := params.parallel_count.val; i > 0; i-- {
			diff := 0
			if i == 1 {
				diff = params.population_size.val
				diff -= params.parallel_count.val * children_to_create
				diff -= params.parents_per_generation.val
			}
			wg.Add(1)
			go func(count int, parents []*Chromosome[T], work_done chan<- ScoredChromosome[T], done_signal chan<- bool) {
				defer wg.Done()
				for c := 0; c < count; c++ {
					mom, dad := WeightedRandomParentChromosomes(parents)
					child, _ := dad.Recombine(mom, []int{}, params.recombination_opts.val)
					mutate(child)
					s := measure_fitness(child)
					work_done <- ScoredChromosome[T]{chromosome: child, score: s}
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

		SortScoredChromosomes(scores)
		best_fitness = scores[0].score
	}

	return generation_count, scores, nil
}

func optimizeChromosomeSequentially[T comparable](params OptimizationParams[T, Chromosome[T]]) (int, []ScoredChromosome[T], error) {
	generation_count := 0
	scores := []ScoredChromosome[T]{}
	measure_fitness := params.measure_fitness.val
	mutate := params.mutate.val
	for _, chromosome := range params.initial_population.val {
		score := ScoredChromosome[T]{chromosome: chromosome, score: measure_fitness(chromosome)}
		scores = append(scores, score)
	}
	SortScoredChromosomes(scores)
	best_fitness := scores[0].score

	for generation_count < params.max_iterations.val && best_fitness < params.fitness_target.val {
		generation_count++
		scores = scores[:params.parents_per_generation.val]
		parents := WeightedParentChromosomes(scores)
		for len(scores) < params.population_size.val {
			mom, dad := WeightedRandomParentChromosomes(parents)
			child, _ := dad.Recombine(mom, []int{}, params.recombination_opts.val)
			mutate(child)
			s := measure_fitness(child)
			scores = append(scores, ScoredChromosome[T]{chromosome: child, score: s})
		}

		SortScoredChromosomes(scores)
		best_fitness = scores[0].score
	}

	return generation_count, scores, nil
}

func OptimizeGenome[T comparable](params OptimizationParams[T, Genome[T]]) (int, []ScoredGenome[T], error) {
	generation_count := 0
	scores := []ScoredGenome[T]{}

	if !params.initial_population.ok() {
		return generation_count, scores, MissingParameterError{"params.initial_population"}
	}
	if len(params.initial_population.val) < 1 {
		return generation_count, scores, Error{"params.initial_population must have len > 0"}
	}
	if !params.measure_fitness.ok() {
		return generation_count, scores, MissingParameterError{"params.measure_fitness"}
	}
	if !params.mutate.ok() {
		return generation_count, scores, MissingParameterError{"params.mutate"}
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
		return optimizeGenomeInParallel(params)
	} else {
		return optimizeGenomeSequentially(params)
	}
}

func optimizeGenomeInParallel[T comparable](params OptimizationParams[T, Genome[T]]) (int, []ScoredGenome[T], error) {
	generation_count := 0
	scores := []ScoredGenome[T]{}
	var wg sync.WaitGroup
	work_done := make(chan ScoredGenome[T], params.population_size.val+10)
	done_signal := make(chan bool, params.parallel_count.val)
	measure_fitness := params.measure_fitness.val
	mutate := params.mutate.val
	for _, genome := range params.initial_population.val {
		score := ScoredGenome[T]{genome: genome, score: measure_fitness(genome)}
		scores = append(scores, score)
	}
	SortScoredGenomes(scores)
	best_fitness := scores[0].score

	for generation_count < params.max_iterations.val && best_fitness < params.fitness_target.val {
		generation_count++
		scores = scores[:params.parents_per_generation.val]
		parents := WeightedParentGenomes(scores)
		children_to_create := (params.population_size.val - params.parents_per_generation.val) / params.parallel_count.val

		for i := params.parallel_count.val; i > 0; i-- {
			diff := 0
			if i == 1 {
				diff = params.population_size.val
				diff -= params.parallel_count.val * children_to_create
				diff -= params.parents_per_generation.val
			}
			wg.Add(1)
			go func(count int, parents []*Genome[T], work_done chan<- ScoredGenome[T], done_signal chan<- bool) {
				defer wg.Done()
				for c := 0; c < count; c++ {
					mom, dad := WeightedRandomParentGenomes(parents)
					child, _ := dad.Recombine(mom, []int{}, params.recombination_opts.val)
					mutate(child)
					s := measure_fitness(child)
					work_done <- ScoredGenome[T]{genome: child, score: s}
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
		SortScoredGenomes(scores)
		best_fitness = scores[0].score
	}

	return generation_count, scores, nil
}

func optimizeGenomeSequentially[T comparable](params OptimizationParams[T, Genome[T]]) (int, []ScoredGenome[T], error) {
	generation_count := 0
	scores := []ScoredGenome[T]{}
	measure_fitness := params.measure_fitness.val
	mutate := params.mutate.val
	for _, genome := range params.initial_population.val {
		score := ScoredGenome[T]{genome: genome, score: measure_fitness(genome)}
		scores = append(scores, score)
	}
	SortScoredGenomes(scores)
	best_fitness := scores[0].score

	for generation_count < params.max_iterations.val && best_fitness < params.fitness_target.val {
		generation_count++
		scores = scores[:params.parents_per_generation.val]
		parents := WeightedParentGenomes(scores)
		for len(scores) < params.population_size.val {
			mom, dad := WeightedRandomParentGenomes(parents)
			child, _ := dad.Recombine(mom, []int{}, params.recombination_opts.val)
			mutate(child)
			s := measure_fitness(child)
			scores = append(scores, ScoredGenome[T]{genome: child, score: s})
		}

		SortScoredGenomes(scores)
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
		return n_goroutines, MissingParameterError{"params.initial_population"}
	}
	if len(params.initial_population.val) < 1 {
		return n_goroutines, Error{"params.initial_population must have len > 0"}
	}
	if !params.measure_fitness.ok() {
		return n_goroutines, MissingParameterError{"params.measure_fitness"}
	}
	if !params.mutate.ok() {
		return n_goroutines, MissingParameterError{"params.mutate"}
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
		return n_goroutines, MissingParameterError{"params.initial_population"}
	}
	if len(params.initial_population.val) < 1 {
		return n_goroutines, Error{"params.initial_population must have len > 0"}
	}
	if !params.measure_fitness.ok() {
		return n_goroutines, MissingParameterError{"params.measure_fitness"}
	}
	if !params.mutate.ok() {
		return n_goroutines, MissingParameterError{"params.mutate"}
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
		return n_goroutines, MissingParameterError{"params.initial_population"}
	}
	if len(params.initial_population.val) < 1 {
		return n_goroutines, Error{"params.initial_population must have len > 0"}
	}
	if !params.measure_fitness.ok() {
		return n_goroutines, MissingParameterError{"params.measure_fitness"}
	}
	if !params.mutate.ok() {
		return n_goroutines, MissingParameterError{"params.mutate"}
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
		return n_goroutines, MissingParameterError{"params.initial_population"}
	}
	if len(params.initial_population.val) < 1 {
		return n_goroutines, Error{"params.initial_population must have len > 0"}
	}
	if !params.measure_fitness.ok() {
		return n_goroutines, MissingParameterError{"params.measure_fitness"}
	}
	if !params.mutate.ok() {
		return n_goroutines, MissingParameterError{"params.mutate"}
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
