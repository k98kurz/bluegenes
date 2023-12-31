package bluegenes

import (
	"math"
	"sort"
	"sync"
	"testing"
	"time"
)

type Code[T Ordered] struct {
	Gene       Option[*Gene[T]]
	Nucleosome Option[*Nucleosome[T]]
	Chromosome Option[*Chromosome[T]]
	Genome     Option[*Genome[T]]
}

func (c Code[T]) Recombine(other Code[T], child *Code[T],
	recombinationOpts RecombineOptions) {
	if c.Gene.Ok() && other.Gene.Ok() &&
		(!recombinationOpts.RecombineGenes.Ok() ||
			recombinationOpts.RecombineGenes.Val) {
		if !child.Gene.Ok() {
			child.Gene = NewOption(&Gene[T]{})
		}
		_ = c.Gene.Val.Recombine(
			other.Gene.Val, []int{}, child.Gene.Val, recombinationOpts,
		)
		child.Gene.IsSet = true
	}
	if c.Nucleosome.Ok() && other.Nucleosome.Ok() &&
		(!recombinationOpts.RecombineNucleosomes.Ok() ||
			recombinationOpts.RecombineNucleosomes.Val) {
		if !child.Nucleosome.Ok() {
			child.Nucleosome = NewOption(&Nucleosome[T]{})
		}
		_ = c.Nucleosome.Val.Recombine(
			other.Nucleosome.Val, []int{}, child.Nucleosome.Val, recombinationOpts,
		)
		child.Nucleosome.IsSet = true
	}
	if c.Chromosome.Ok() && other.Chromosome.Ok() &&
		(!recombinationOpts.RecombineChromosomes.Ok() ||
			recombinationOpts.RecombineChromosomes.Val) {
		if !child.Chromosome.Ok() {
			child.Chromosome = NewOption(&Chromosome[T]{})
		}
		_ = c.Chromosome.Val.Recombine(
			other.Chromosome.Val, []int{}, child.Chromosome.Val, recombinationOpts,
		)
		child.Chromosome.IsSet = true
	}
	if c.Genome.Ok() && other.Genome.Ok() &&
		(!recombinationOpts.RecombineGenomes.Ok() ||
			recombinationOpts.RecombineGenomes.Val) {
		if !child.Genome.Ok() {
			child.Genome = NewOption(&Genome[T]{})
		}
		_ = c.Genome.Val.Recombine(
			other.Genome.Val, []int{}, child.Genome.Val, recombinationOpts,
		)
		child.Genome.IsSet = true
	}
}

func (c Code[T]) Copy() Code[T] {
	gm := Code[T]{}
	if c.Gene.Ok() {
		gm.Gene = NewOption(c.Gene.Val.Copy())
	}
	if c.Nucleosome.Ok() {
		gm.Nucleosome = NewOption(c.Nucleosome.Val.Copy())
	}
	if c.Chromosome.Ok() {
		gm.Chromosome = NewOption(c.Chromosome.Val.Copy())
	}
	if c.Genome.Ok() {
		gm.Genome = NewOption(c.Genome.Val.Copy())
	}
	return gm
}

type OptimizationParams[T Ordered] struct {
	MeasureFitness       Option[func(Code[T]) float64]
	Mutate               Option[func(*Code[T])]
	InitialPopulation    Option[[]Code[T]]
	MaxIterations        Option[int]
	PopulationSize       Option[int]
	ParentsPerGeneration Option[int]
	FitnessTarget        Option[float64]
	RecombinationOpts    Option[RecombineOptions]
	ParallelCount        Option[int]
	IterationHook        Option[func(int, []*ScoredCode[T])]
}

type BenchmarkResult struct {
	CostOfCopy           int
	CostOfMutate         int
	CostOfMeasureFitness int
	CostOfIterationHook  int
}

type ScoredCode[T Ordered] struct {
	Code  Code[T]
	Score float64
}

func sortScoredCodes[T Ordered](scores []*ScoredCode[T]) {
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

func weightedParents[T Ordered](scores []*ScoredCode[T]) []Code[T] {
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

func weightedRandomParents[T Ordered](parents []Code[T]) (Code[T], Code[T]) {
	dad_and_mom := RandomChoices(parents, 2)
	dad := dad_and_mom[0]
	mom := dad_and_mom[1]
	for mom == dad {
		mom = RandomChoices(parents, 1)[0]
	}
	return dad, mom
}

func Optimize[T Ordered](params OptimizationParams[T]) (int, []*ScoredCode[T], error) {
	generation_count := 0
	scores := []*ScoredCode[T]{}

	if !params.InitialPopulation.Ok() {
		return generation_count, scores, missingParameterError{"params.InitialPopulation"}
	}
	if len(params.InitialPopulation.Val) < 1 {
		return generation_count, scores, anError{"params.InitialPopulation Must have len > 0"}
	}
	if !params.MeasureFitness.Ok() {
		return generation_count, scores, missingParameterError{"params.MeasureFitness"}
	}
	if !params.Mutate.Ok() {
		return generation_count, scores, missingParameterError{"params.Mutate"}
	}
	if !params.MaxIterations.Ok() {
		params.MaxIterations.Val = 1000
	}
	if !params.PopulationSize.Ok() {
		params.PopulationSize.Val = 100
	}
	if params.PopulationSize.Val < 3 {
		return generation_count, scores, anError{"params.PopulationSize must be at least 3"}
	}
	if !params.ParentsPerGeneration.Ok() {
		params.ParentsPerGeneration.Val = 10
	}
	if !params.FitnessTarget.Ok() {
		params.FitnessTarget.Val = float64(0.99)
	}
	if params.ParentsPerGeneration.Val > params.PopulationSize.Val {
		params.ParentsPerGeneration.Val = params.PopulationSize.Val / 10
	}
	if params.ParentsPerGeneration.Val < 2 {
		params.ParentsPerGeneration.Val = 2
	}
	if params.ParallelCount.Ok() && params.PopulationSize.Val/params.ParallelCount.Val < 1 {
		params.ParallelCount.Val = params.PopulationSize.Val / 2
	}

	if params.ParallelCount.Ok() && params.ParallelCount.Val > 1 {
		return optimizeInParallel(params)
	} else {
		return optimizeSequentially(params)
	}
}

func optimizeInParallel[T Ordered](params OptimizationParams[T]) (int, []*ScoredCode[T], error) {
	generation_count := 0
	scores_pool_size, _ := max(params.PopulationSize.Val, len(params.InitialPopulation.Val))
	scores_pool := make(chan *ScoredCode[T], scores_pool_size+10)
	for i := 0; i < scores_pool_size; i++ {
		scores_pool <- &ScoredCode[T]{}
	}
	scores := []*ScoredCode[T]{}
	var wg sync.WaitGroup
	work_done := make(chan *ScoredCode[T], params.PopulationSize.Val+10)
	done_signal := make(chan bool, params.ParallelCount.Val)
	measure_fitness := params.MeasureFitness.Val
	Mutate := params.Mutate.Val
	for _, code := range params.InitialPopulation.Val {
		score := <-scores_pool
		score.Code = code
		score.Score = measure_fitness(code)
		scores = append(scores, score)
	}
	sortScoredCodes(scores)
	best_fitness := scores[0].Score

	for generation_count < params.MaxIterations.Val && best_fitness < params.FitnessTarget.Val {
		generation_count++
		for _, score := range scores[params.ParentsPerGeneration.Val:] {
			scores_pool <- score
		}
		scores = scores[:params.ParentsPerGeneration.Val]
		parents := weightedParents(scores)
		children_to_create := (params.PopulationSize.Val -
			params.ParentsPerGeneration.Val) / params.ParallelCount.Val

		for i := params.ParallelCount.Val; i > 0; i-- {
			diff := 0
			if i == 1 {
				diff = params.PopulationSize.Val - params.ParentsPerGeneration.Val
				diff -= params.ParallelCount.Val * children_to_create
			}
			wg.Add(1)
			go func(count int, parents []Code[T], work_done chan<- *ScoredCode[T], done_signal chan<- bool, scores_pool <-chan *ScoredCode[T]) {
				defer wg.Done()
				for c := 0; c < count; c++ {
					child := <-scores_pool
					mom, dad := weightedRandomParents(parents)
					dad.Recombine(mom, &child.Code, params.RecombinationOpts.Val)
					Mutate(&child.Code)
					child.Score = measure_fitness(child.Code)
					work_done <- child
				}
				done_signal <- true
			}(children_to_create+diff, parents, work_done, done_signal, scores_pool)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			finished := 0
			for finished < params.ParallelCount.Val {
				select {
				case child := <-work_done:
					scores = append(scores, child)
				case <-done_signal:
					finished += 1
				default:
				}
			}
			finished = 0
			for finished == 0 {
				select {
				case child := <-work_done:
					scores = append(scores, child)
				default:
					finished = 1
				}
			}
		}()

		wg.Wait()
		sortScoredCodes(scores)
		best_fitness = scores[0].Score

		if params.IterationHook.Ok() {
			params.IterationHook.Val(generation_count, scores)
		}
	}

	return generation_count, scores, nil
}

func optimizeSequentially[T Ordered](params OptimizationParams[T]) (int,
	[]*ScoredCode[T], error) {
	generation_count := 0
	scores_pool_size, _ := max(params.PopulationSize.Val, len(params.InitialPopulation.Val))
	scores_pool := make(chan *ScoredCode[T], scores_pool_size+10)
	for i := 0; i < scores_pool_size; i++ {
		scores_pool <- &ScoredCode[T]{}
	}
	scores := []*ScoredCode[T]{}
	measure_fitness := params.MeasureFitness.Val
	Mutate := params.Mutate.Val
	for _, code := range params.InitialPopulation.Val {
		score := <-scores_pool
		score.Code = code
		score.Score = measure_fitness(code)
		scores = append(scores, score)
	}
	sortScoredCodes(scores)
	best_fitness := scores[0].Score

	for generation_count < params.MaxIterations.Val && best_fitness < params.FitnessTarget.Val {
		generation_count++
		for _, score := range scores[params.ParentsPerGeneration.Val:] {
			scores_pool <- score
		}
		scores = scores[:params.ParentsPerGeneration.Val]
		parents := weightedParents(scores)
		for len(scores) < params.PopulationSize.Val {
			child := <-scores_pool
			mom, dad := weightedRandomParents(parents)
			dad.Recombine(mom, &child.Code, params.RecombinationOpts.Val)
			Mutate(&child.Code)
			child.Score = measure_fitness(child.Code)
			scores = append(scores, child)
		}

		sortScoredCodes(scores)
		best_fitness = scores[0].Score

		if params.IterationHook.Ok() {
			params.IterationHook.Val(generation_count, scores)
		}
	}

	return generation_count, scores, nil
}

func TuneOptimization[T Ordered](params OptimizationParams[T], max_threads ...int) (int, error) {
	n_goroutines := 1
	max_goroutines := 4
	if len(max_threads) > 0 {
		max_goroutines = max_threads[0]
	}
	if !params.InitialPopulation.Ok() {
		return n_goroutines, missingParameterError{"params.InitialPopulation"}
	}
	if len(params.InitialPopulation.Val) < 1 {
		return n_goroutines, anError{"params.InitialPopulation Must have len > 0"}
	}
	if !params.MeasureFitness.Ok() {
		return n_goroutines, missingParameterError{"params.MeasureFitness"}
	}
	if !params.Mutate.Ok() {
		return n_goroutines, missingParameterError{"params.Mutate"}
	}
	if !params.PopulationSize.Ok() {
		params.PopulationSize.Val = 100
	}
	if !params.ParentsPerGeneration.Ok() {
		params.ParentsPerGeneration.Val = 10
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

func BenchmarkOptimization[T Ordered](params OptimizationParams[T]) BenchmarkResult {
	res := testing.Benchmark(func(b *testing.B) {
		gm := params.InitialPopulation.Val[0]
		for i := 0; i < b.N; i++ {
			params.Mutate.Val(&gm)
		}
	})
	CostOfMutate := res.T / time.Duration(res.N)

	res = testing.Benchmark(func(b *testing.B) {
		gm := params.InitialPopulation.Val[0]
		for i := 0; i < b.N; i++ {
			params.MeasureFitness.Val(gm)
		}
	})
	CostOfMeasureFitness := res.T / time.Duration(res.N)

	res = testing.Benchmark(func(b *testing.B) {
		var bg sync.WaitGroup
		gm := params.InitialPopulation.Val[0]
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
				bc1 <- gm.Copy()
				bc2 <- true
			}(buffered_chan1, buffered_chan2)
			bg.Wait()
		}
	})
	CostOfCopy := res.T / time.Duration(res.N)

	CostOfIterationHook := time.Second * 0

	if params.IterationHook.Ok() {
		scored := []*ScoredCode[T]{}
		for len(scored) < params.PopulationSize.Val {
			scored = append(scored, &ScoredCode[T]{Code: params.InitialPopulation.Val[0], Score: 0.5})
		}
		res := testing.Benchmark(func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				params.IterationHook.Val(i, scored)
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
