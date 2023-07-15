package bluegenes

import (
	"math"
	"sort"
	"sync"
	"testing"
	"time"
)

type Code[T comparable] struct {
	Gene       Option[*Gene[T]]
	Allele     Option[*Allele[T]]
	Chromosome Option[*Chromosome[T]]
	Genome     Option[*Genome[T]]
}

func (self Code[T]) Recombine(other Code[T], recombinationOpts RecombineOptions) Code[T] {
	child := Code[T]{}
	if self.Gene.ok() && other.Gene.ok() {
		child.Gene.val, _ = self.Gene.val.Recombine(other.Gene.val, []int{}, recombinationOpts)
		child.Gene.isSet = true
	}
	if self.Allele.ok() && other.Allele.ok() {
		child.Allele.val, _ = self.Allele.val.Recombine(other.Allele.val, []int{}, recombinationOpts)
		child.Allele.isSet = true
	}
	if self.Chromosome.ok() && other.Chromosome.ok() {
		child.Chromosome.val, _ = self.Chromosome.val.Recombine(other.Chromosome.val, []int{}, recombinationOpts)
		child.Chromosome.isSet = true
	}
	if self.Genome.ok() && other.Genome.ok() {
		child.Genome.val, _ = self.Genome.val.Recombine(other.Genome.val, []int{}, recombinationOpts)
		child.Genome.isSet = true
	}
	return child
}

func (self Code[T]) copy() Code[T] {
	gm := Code[T]{}
	if self.Gene.ok() {
		gm.Gene = NewOption(self.Gene.val.Copy())
	}
	if self.Allele.ok() {
		gm.Allele = NewOption(self.Allele.val.Copy())
	}
	if self.Chromosome.ok() {
		gm.Chromosome = NewOption(self.Chromosome.val.Copy())
	}
	if self.Genome.ok() {
		gm.Genome = NewOption(self.Genome.val.Copy())
	}
	return gm
}

type OptimizationParams[T comparable] struct {
	MeasureFitness       Option[func(Code[T]) float64]
	Mutate               Option[func(Code[T])]
	InitialPopulation    Option[[]Code[T]]
	MaxIterations        Option[int]
	PopulationSize       Option[int]
	ParentsPerGeneration Option[int]
	FitnessTarget        Option[float64]
	RecombinationOpts    Option[RecombineOptions]
	ParallelCount        Option[int]
	IterationHook        Option[func(int, []ScoredCode[T])]
}

type BenchmarkResult struct {
	CostOfCopy           int
	CostOfMutate         int
	CostOfMeasureFitness int
	CostOfIterationHook  int
}

type ScoredCode[T comparable] struct {
	Code  Code[T]
	Score float64
}

func sortScoredCodes[T comparable](scores []ScoredCode[T]) {
	sort.SliceStable(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
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

func weightedParents[T comparable](scores []ScoredCode[T]) []Code[T] {
	parents := []Code[T]{}
	weight := len(scores)
	for i, l := 0, len(scores); i < l; i++ {
		for j := 0; j < weight; j++ {
			parents = append(parents, scores[i].Code)
		}
		weight--
	}
	return parents
}

func weightedRandomParents[T comparable](parents []Code[T]) (Code[T], Code[T]) {
	dad_and_mom := RandomChoices(parents, 2)
	dad := dad_and_mom[0]
	mom := dad_and_mom[1]
	for mom == dad {
		mom = RandomChoices(parents, 1)[0]
	}
	return dad, mom
}

func Optimize[T comparable](params OptimizationParams[T]) (int, []ScoredCode[T], error) {
	generation_count := 0
	scores := []ScoredCode[T]{}

	if !params.InitialPopulation.ok() {
		return generation_count, scores, missingParameterError{"params.InitialPopulation"}
	}
	if len(params.InitialPopulation.val) < 1 {
		return generation_count, scores, anError{"params.InitialPopulation Must have len > 0"}
	}
	if !params.MeasureFitness.ok() {
		return generation_count, scores, missingParameterError{"params.MeasureFitness"}
	}
	if !params.Mutate.ok() {
		return generation_count, scores, missingParameterError{"params.Mutate"}
	}
	if !params.MaxIterations.ok() {
		params.MaxIterations.val = 1000
	}
	if !params.PopulationSize.ok() {
		params.PopulationSize.val = 100
	}
	if !params.ParentsPerGeneration.ok() {
		params.ParentsPerGeneration.val = 10
	}
	if !params.FitnessTarget.ok() {
		params.FitnessTarget.val = float64(0.99)
	}
	if params.ParallelCount.ok() && params.PopulationSize.val/params.ParallelCount.val < 1 {
		params.PopulationSize.val *= 2
	}
	if params.ParentsPerGeneration.val > params.PopulationSize.val {
		params.ParentsPerGeneration.val = params.PopulationSize.val / 10
	}

	if params.ParallelCount.ok() {
		return optimizeInParallel(params)
	} else {
		return optimizeSequentially(params)
	}
}

func optimizeInParallel[T comparable](params OptimizationParams[T]) (int, []ScoredCode[T], error) {
	generation_count := 0
	scores := []ScoredCode[T]{}
	var wg sync.WaitGroup
	work_done := make(chan ScoredCode[T], params.PopulationSize.val+10)
	done_signal := make(chan bool, params.ParallelCount.val)
	measure_fitness := params.MeasureFitness.val
	Mutate := params.Mutate.val
	for _, code := range params.InitialPopulation.val {
		score := ScoredCode[T]{Code: code, Score: measure_fitness(code)}
		scores = append(scores, score)
	}
	sortScoredCodes(scores)
	best_fitness := scores[0].Score

	for generation_count < params.MaxIterations.val && best_fitness < params.FitnessTarget.val {
		generation_count++
		scores = scores[:params.ParentsPerGeneration.val]
		parents := weightedParents(scores)
		children_to_create := (params.PopulationSize.val - params.ParentsPerGeneration.val) / params.ParallelCount.val

		for i := params.ParallelCount.val; i > 0; i-- {
			diff := 0
			if i == 1 {
				diff = params.PopulationSize.val
				diff -= params.ParallelCount.val * children_to_create
				diff -= params.ParentsPerGeneration.val
			}
			wg.Add(1)
			go func(count int, parents []Code[T], work_done chan<- ScoredCode[T], done_signal chan<- bool) {
				defer wg.Done()
				for c := 0; c < count; c++ {
					mom, dad := weightedRandomParents(parents)
					child := dad.Recombine(mom, params.RecombinationOpts.val)
					Mutate(child)
					s := measure_fitness(child)
					work_done <- ScoredCode[T]{Code: child, Score: s}
				}
				done_signal <- true
			}(children_to_create+diff, parents, work_done, done_signal)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			finished := 0
			for finished < params.ParallelCount.val {
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
		sortScoredCodes(scores)
		best_fitness = scores[0].Score

		if params.IterationHook.ok() {
			params.IterationHook.val(generation_count, scores)
		}
	}

	return generation_count, scores, nil
}

func optimizeSequentially[T comparable](params OptimizationParams[T]) (int, []ScoredCode[T], error) {
	generation_count := 0
	scores := []ScoredCode[T]{}
	measure_fitness := params.MeasureFitness.val
	Mutate := params.Mutate.val
	for _, code := range params.InitialPopulation.val {
		score := ScoredCode[T]{Code: code, Score: measure_fitness(code)}
		scores = append(scores, score)
	}
	sortScoredCodes(scores)
	best_fitness := scores[0].Score

	for generation_count < params.MaxIterations.val && best_fitness < params.FitnessTarget.val {
		generation_count++
		scores = scores[:params.ParentsPerGeneration.val]
		parents := weightedParents(scores)
		for len(scores) < params.PopulationSize.val {
			mom, dad := weightedRandomParents(parents)
			child := dad.Recombine(mom, params.RecombinationOpts.val)
			Mutate(child)
			s := measure_fitness(child)
			scores = append(scores, ScoredCode[T]{Code: child, Score: s})
		}

		sortScoredCodes(scores)
		best_fitness = scores[0].Score

		if params.IterationHook.ok() {
			params.IterationHook.val(generation_count, scores)
		}
	}

	return generation_count, scores, nil
}

func TuneOptimization[T comparable](params OptimizationParams[T], max_threads ...int) (int, error) {
	n_goroutines := 1
	max_goroutines := 4
	if len(max_threads) > 0 {
		max_goroutines = max_threads[0]
	}
	if !params.InitialPopulation.ok() {
		return n_goroutines, missingParameterError{"params.InitialPopulation"}
	}
	if len(params.InitialPopulation.val) < 1 {
		return n_goroutines, anError{"params.InitialPopulation Must have len > 0"}
	}
	if !params.MeasureFitness.ok() {
		return n_goroutines, missingParameterError{"params.MeasureFitness"}
	}
	if !params.Mutate.ok() {
		return n_goroutines, missingParameterError{"params.Mutate"}
	}
	if !params.PopulationSize.ok() {
		params.PopulationSize.val = 100
	}
	if !params.ParentsPerGeneration.ok() {
		params.ParentsPerGeneration.val = 10
	}

	res := BenchmarkOptimization(params)

	n_goroutines = int(math.Log2(float64((res.CostOfMutate + res.CostOfMeasureFitness + res.CostOfIterationHook) / res.CostOfCopy)))

	if n_goroutines > max_goroutines {
		n_goroutines = max_goroutines
	} else if n_goroutines <= 0 {
		n_goroutines = 1
	}

	return n_goroutines, nil
}

func BenchmarkOptimization[T comparable](params OptimizationParams[T]) BenchmarkResult {
	res := testing.Benchmark(func(b *testing.B) {
		gm := params.InitialPopulation.val[0]
		for i := 0; i < b.N; i++ {
			params.Mutate.val(gm)
		}
	})
	CostOfMutate := res.T / time.Duration(res.N)

	res = testing.Benchmark(func(b *testing.B) {
		gm := params.InitialPopulation.val[0]
		for i := 0; i < b.N; i++ {
			params.MeasureFitness.val(gm)
		}
	})
	CostOfMeasureFitness := res.T / time.Duration(res.N)

	res = testing.Benchmark(func(b *testing.B) {
		var bg sync.WaitGroup
		gm := params.InitialPopulation.val[0]
		buffered_chan1 := make(chan Code[T], 1)
		buffered_chan2 := make(chan bool, 1)
		for i := 0; i < b.N; i++ {
			bg.Add(2)
			go func(bc1 <-chan Code[T], bc2 <-chan bool) {
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
			go func(bc1 chan<- Code[T], bc2 chan<- bool) {
				defer bg.Done()
				bc1 <- gm.copy()
				bc2 <- true
			}(buffered_chan1, buffered_chan2)
			bg.Wait()
		}
	})
	CostOfCopy := res.T / time.Duration(res.N)

	CostOfIterationHook := time.Second * 0

	if params.IterationHook.ok() {
		scored := []ScoredCode[T]{}
		for len(scored) < params.PopulationSize.val {
			scored = append(scored, ScoredCode[T]{Code: params.InitialPopulation.val[0], Score: 0.5})
		}
		res := testing.Benchmark(func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				params.IterationHook.val(i, scored)
			}
		})

		CostOfIterationHook = res.T / time.Duration(res.N)
	}

	return BenchmarkResult{
		CostOfMutate:         int(CostOfMutate),
		CostOfMeasureFitness: int(CostOfMeasureFitness),
		CostOfCopy:           int(CostOfCopy),
		CostOfIterationHook:  int(CostOfIterationHook),
	}
}
