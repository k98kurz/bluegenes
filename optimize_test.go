package bluegenes

import (
	"math"
	"math/rand"
	"testing"
)

var target = 12345

func MutateGene(gene *Gene[int]) {
	gene.Mu.Lock()
	defer gene.Mu.Unlock()
	for i := 0; i < len(gene.Bases); i++ {
		val := rand.Float64()
		if val <= 0.1 {
			gene.Bases[i] /= RandomInt(1, 3)
		} else if val <= 0.2 {
			gene.Bases[i] *= RandomInt(1, 3)
		} else if val <= 0.6 {
			gene.Bases[i] += RandomInt(0, 11)
		} else {
			gene.Bases[i] -= RandomInt(0, 11)
		}
	}
}

func MutateNucleosome(nucleosome *Nucleosome[int]) {
	nucleosome.Mu.Lock()
	defer nucleosome.Mu.Unlock()
	for _, gene := range nucleosome.Genes {
		MutateGene(gene)
	}
}

func MutateChromosome(chromosome *Chromosome[int]) {
	chromosome.Mu.Lock()
	defer chromosome.Mu.Unlock()
	for _, nucleosome := range chromosome.Nucleosomes {
		MutateNucleosome(nucleosome)
	}
}

func MutateGenome(genome *Genome[int]) {
	genome.Mu.Lock()
	defer genome.Mu.Unlock()
	for _, chromosome := range genome.Chromosomes {
		MutateChromosome(chromosome)
	}
}

func MutateCode(code *Code[int]) {
	if code.Gene.Ok() {
		MutateGene(code.Gene.Val)
	}
	if code.Nucleosome.Ok() {
		MutateNucleosome(code.Nucleosome.Val)
	}
	if code.Chromosome.Ok() {
		MutateChromosome(code.Chromosome.Val)
	}
	if code.Genome.Ok() {
		MutateGenome(code.Genome.Val)
	}
}

func measureGeneFitness(gene *Gene[int]) float64 {
	gene.Mu.RLock()
	defer gene.Mu.RUnlock()
	total := reduce(gene.Bases, func(a int, b int) int { return a + b })
	return 1.0 / (math.Abs(float64(total-target)) + 1.0)
}

func measureNucleosomeFitness(nucleosome *Nucleosome[int]) float64 {
	nucleosome.Mu.RLock()
	defer nucleosome.Mu.RUnlock()
	total := 0
	for _, gene := range nucleosome.Genes {
		total += reduce(gene.Bases, func(a int, b int) int { return a + b })
	}
	return 1.0 / (math.Abs(float64(total-target)) + 1.0)
}

func measureChromosomeFitness(chromosome *Chromosome[int]) float64 {
	chromosome.Mu.RLock()
	defer chromosome.Mu.RUnlock()
	total := 0
	for _, nucleosome := range chromosome.Nucleosomes {
		for _, gene := range nucleosome.Genes {
			total += reduce(gene.Bases, func(a int, b int) int { return a + b })
		}
	}
	return 1.0 / (math.Abs(float64(total-target)) + 1.0)
}

func measureGenomeFitness(genome *Genome[int]) float64 {
	genome.Mu.RLock()
	defer genome.Mu.RUnlock()
	total := 0
	for _, chromosome := range genome.Chromosomes {
		for _, nucleosome := range chromosome.Nucleosomes {
			for _, gene := range nucleosome.Genes {
				total += reduce(gene.Bases, func(a int, b int) int { return a + b })
			}
		}
	}
	return 1.0 / (math.Abs(float64(total-target)) + 1.0)
}

func measureCodeFitness(code Code[int]) float64 {
	fitness := 0.0
	fitness_count := 0
	if code.Gene.Ok() {
		fitness += measureGeneFitness(code.Gene.Val)
		fitness_count++
	}
	if code.Nucleosome.Ok() {
		fitness += measureNucleosomeFitness(code.Nucleosome.Val)
		fitness_count++
	}
	if code.Chromosome.Ok() {
		fitness += measureChromosomeFitness(code.Chromosome.Val)
		fitness_count++
	}
	if code.Genome.Ok() {
		fitness += measureGenomeFitness(code.Genome.Val)
		fitness_count++
	}

	return fitness / float64(fitness_count)
}

func MutateCodeExpensive(code *Code[int]) {
	val := 42.0
	for i := 0; i < 1000; i++ {
		val /= 6.9
	}
	if code.Gene.Ok() {
		MutateGene(code.Gene.Val)
	}
	if code.Nucleosome.Ok() {
		MutateNucleosome(code.Nucleosome.Val)
	}
	if code.Chromosome.Ok() {
		MutateChromosome(code.Chromosome.Val)
	}
	if code.Genome.Ok() {
		MutateGenome(code.Genome.Val)
	}
}

func measureCodeFitnessExpensive(code Code[int]) float64 {
	val := 42.0
	for i := 0; i < 1000; i++ {
		val /= 6.9
	}
	fitness := 0.0
	fitness_count := 0
	if code.Gene.Ok() {
		fitness += measureGeneFitness(code.Gene.Val)
		fitness_count++
	}
	if code.Nucleosome.Ok() {
		fitness += measureNucleosomeFitness(code.Nucleosome.Val)
		fitness_count++
	}
	if code.Chromosome.Ok() {
		fitness += measureChromosomeFitness(code.Chromosome.Val)
		fitness_count++
	}
	if code.Genome.Ok() {
		fitness += measureGenomeFitness(code.Genome.Val)
		fitness_count++
	}

	return fitness / float64(fitness_count)
}

func TestOptimize(t *testing.T) {
	t.Run("Gene", func(t *testing.T) {
		t.Run("parallel", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NNucleosomes: NewOption(uint(3)),
				NChromosomes: NewOption(uint(2)),
				BaseFactory:  NewOption(base_factory),
			}
			initial_population := []Code[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeGene(opts)
				initial_population = append(initial_population, Code[int]{Gene: NewOption(gene)})
			}

			n_iterations, final_population, err := Optimize(OptimizationParams[int]{
				InitialPopulation: NewOption(initial_population),
				MeasureFitness:    NewOption(measureCodeFitness),
				Mutate:            NewOption(MutateCode),
				MaxIterations:     NewOption(1000),
				ParallelCount:     NewOption(10),
				RecombinationOpts: NewOption(RecombineOptions{
					RecombineGenes:       NewOption(true),
					MatchGenes:           NewOption(false),
					RecombineNucleosomes: NewOption(true),
					MatchNucleosomes:     NewOption(false),
					RecombineChromosomes: NewOption(true),
					MatchChromosomes:     NewOption(false),
				}),
			})

			if err != nil {
				t.Fatalf("Optimize for Gene[int] failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("Optimize for Gene[int] exceeded MaxIterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.Score < 0.9 {
				t.Errorf("Optimize for Gene[int] failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.Score)
			}
		})
		t.Run("sequential", func(t *testing.T) {
			t.Parallel()
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NNucleosomes: NewOption(uint(3)),
				NChromosomes: NewOption(uint(2)),
				BaseFactory:  NewOption(base_factory),
			}
			initial_population := []Code[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeGene(opts)
				initial_population = append(initial_population, Code[int]{Gene: NewOption(gene)})
			}

			n_iterations, final_population, err := Optimize(OptimizationParams[int]{
				InitialPopulation: NewOption(initial_population),
				MeasureFitness:    NewOption(measureCodeFitness),
				Mutate:            NewOption(MutateCode),
				MaxIterations:     NewOption(1000),
				RecombinationOpts: NewOption(RecombineOptions{
					RecombineGenes:       NewOption(true),
					MatchGenes:           NewOption(false),
					RecombineNucleosomes: NewOption(true),
					MatchNucleosomes:     NewOption(false),
					RecombineChromosomes: NewOption(true),
					MatchChromosomes:     NewOption(false),
				}),
			})

			if err != nil {
				t.Fatalf("Optimize for Gene[int] failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("Optimize for Gene[int] exceeded MaxIterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.Score < 0.9 {
				t.Errorf("Optimize for Gene[int] failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.Score)
			}
		})
	})

	t.Run("Nucleosome", func(t *testing.T) {
		t.Run("parallel", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NNucleosomes: NewOption(uint(3)),
				NChromosomes: NewOption(uint(2)),
				BaseFactory:  NewOption(base_factory),
			}
			initial_population := []Code[int]{}
			for i := 0; i < 10; i++ {
				nucleosome, _ := MakeNucleosome(opts)
				initial_population = append(initial_population, Code[int]{Nucleosome: NewOption(nucleosome)})
			}

			n_iterations, final_population, err := Optimize(OptimizationParams[int]{
				InitialPopulation: NewOption(initial_population),
				MeasureFitness:    NewOption(measureCodeFitness),
				Mutate:            NewOption(MutateCode),
				MaxIterations:     NewOption(1000),
				ParallelCount:     NewOption(10),
				RecombinationOpts: NewOption(RecombineOptions{
					RecombineGenes:       NewOption(true),
					MatchGenes:           NewOption(false),
					RecombineNucleosomes: NewOption(true),
					MatchNucleosomes:     NewOption(false),
					RecombineChromosomes: NewOption(true),
					MatchChromosomes:     NewOption(false),
				}),
			})

			if err != nil {
				t.Fatalf("Optimize for Nucleosome[int] failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("Optimize for Nucleosome[int] exceeded MaxIterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.Score < 0.9 {
				t.Errorf("Optimize for Nucleosome[int] failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.Score)
			}
		})
		t.Run("sequential", func(t *testing.T) {
			t.Parallel()
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NNucleosomes: NewOption(uint(3)),
				NChromosomes: NewOption(uint(2)),
				BaseFactory:  NewOption(base_factory),
			}
			initial_population := []Code[int]{}
			for i := 0; i < 10; i++ {
				nucleosome, _ := MakeNucleosome(opts)
				initial_population = append(initial_population, Code[int]{Nucleosome: NewOption(nucleosome)})
			}

			n_iterations, final_population, err := Optimize(OptimizationParams[int]{
				InitialPopulation: NewOption(initial_population),
				MeasureFitness:    NewOption(measureCodeFitness),
				Mutate:            NewOption(MutateCode),
				MaxIterations:     NewOption(1000),
				RecombinationOpts: NewOption(RecombineOptions{
					RecombineGenes:       NewOption(true),
					MatchGenes:           NewOption(false),
					RecombineNucleosomes: NewOption(true),
					MatchNucleosomes:     NewOption(false),
					RecombineChromosomes: NewOption(true),
					MatchChromosomes:     NewOption(false),
				}),
			})

			if err != nil {
				t.Fatalf("Optimize for Nucleosome[int] failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("Optimize for Nucleosome[int] exceeded MaxIterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.Score < 0.9 {
				t.Errorf("Optimize for Nucleosome[int] failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.Score)
			}
		})
	})

	t.Run("Chromosome", func(t *testing.T) {
		t.Run("parallel", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NNucleosomes: NewOption(uint(3)),
				NChromosomes: NewOption(uint(2)),
				BaseFactory:  NewOption(base_factory),
			}
			initial_population := []Code[int]{}
			for i := 0; i < 10; i++ {
				chromosome, _ := MakeChromosome(opts)
				initial_population = append(initial_population, Code[int]{Chromosome: NewOption(chromosome)})
			}

			n_iterations, final_population, err := Optimize(OptimizationParams[int]{
				InitialPopulation: NewOption(initial_population),
				MeasureFitness:    NewOption(measureCodeFitness),
				Mutate:            NewOption(MutateCode),
				MaxIterations:     NewOption(1000),
				ParallelCount:     NewOption(10),
				RecombinationOpts: NewOption(RecombineOptions{
					RecombineGenes:       NewOption(true),
					MatchGenes:           NewOption(false),
					RecombineNucleosomes: NewOption(true),
					MatchNucleosomes:     NewOption(false),
					RecombineChromosomes: NewOption(true),
					MatchChromosomes:     NewOption(false),
				}),
			})

			if err != nil {
				t.Fatalf("Optimize for Chromosome[int] failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("Optimize for Chromosome[int] exceeded MaxIterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.Score < 0.9 {
				t.Errorf("Optimize for Chromosome[int] failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.Score)
			}
		})
		t.Run("sequential", func(t *testing.T) {
			t.Parallel()
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NNucleosomes: NewOption(uint(3)),
				NChromosomes: NewOption(uint(2)),
				BaseFactory:  NewOption(base_factory),
			}
			initial_population := []Code[int]{}
			for i := 0; i < 10; i++ {
				chromosome, _ := MakeChromosome(opts)
				initial_population = append(initial_population, Code[int]{Chromosome: NewOption(chromosome)})
			}

			n_iterations, final_population, err := Optimize(OptimizationParams[int]{
				InitialPopulation: NewOption(initial_population),
				MeasureFitness:    NewOption(measureCodeFitness),
				Mutate:            NewOption(MutateCode),
				MaxIterations:     NewOption(1000),
				RecombinationOpts: NewOption(RecombineOptions{
					RecombineGenes:       NewOption(true),
					MatchGenes:           NewOption(false),
					RecombineNucleosomes: NewOption(true),
					MatchNucleosomes:     NewOption(false),
					RecombineChromosomes: NewOption(true),
					MatchChromosomes:     NewOption(false),
				}),
			})

			if err != nil {
				t.Fatalf("Optimize for Chromosome[int] failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("Optimize for Chromosome[int] exceeded MaxIterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.Score < 0.9 {
				t.Errorf("Optimize for Chromosome[int] failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.Score)
			}
		})
	})

	t.Run("Genome", func(t *testing.T) {
		t.Run("parallel", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NNucleosomes: NewOption(uint(3)),
				NChromosomes: NewOption(uint(2)),
				BaseFactory:  NewOption(base_factory),
			}
			initial_population := []Code[int]{}
			for i := 0; i < 10; i++ {
				genome, _ := MakeGenome(opts)
				initial_population = append(initial_population, Code[int]{Genome: NewOption(genome)})
			}

			n_iterations, final_population, err := Optimize(OptimizationParams[int]{
				InitialPopulation: NewOption(initial_population),
				MeasureFitness:    NewOption(measureCodeFitness),
				Mutate:            NewOption(MutateCode),
				MaxIterations:     NewOption(1000),
				ParallelCount:     NewOption(10),
				RecombinationOpts: NewOption(RecombineOptions{
					RecombineGenes:       NewOption(true),
					MatchGenes:           NewOption(false),
					RecombineNucleosomes: NewOption(true),
					MatchNucleosomes:     NewOption(false),
					RecombineChromosomes: NewOption(true),
					MatchChromosomes:     NewOption(false),
				}),
			})

			if err != nil {
				t.Fatalf("Optimize for Genome[int] failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("Optimize for Genome[int] exceeded MaxIterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.Score < 0.9 {
				t.Errorf("Optimize for Genome[int] failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.Score)
			}
		})
		t.Run("sequential", func(t *testing.T) {
			t.Parallel()
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NNucleosomes: NewOption(uint(3)),
				NChromosomes: NewOption(uint(2)),
				BaseFactory:  NewOption(base_factory),
			}
			initial_population := []Code[int]{}
			for i := 0; i < 10; i++ {
				genome, _ := MakeGenome(opts)
				initial_population = append(initial_population, Code[int]{Genome: NewOption(genome)})
			}

			n_iterations, final_population, err := Optimize(OptimizationParams[int]{
				InitialPopulation: NewOption(initial_population),
				MeasureFitness:    NewOption(measureCodeFitness),
				Mutate:            NewOption(MutateCode),
				MaxIterations:     NewOption(1000),
				RecombinationOpts: NewOption(RecombineOptions{
					RecombineGenes:       NewOption(true),
					MatchGenes:           NewOption(false),
					RecombineNucleosomes: NewOption(true),
					MatchNucleosomes:     NewOption(false),
					RecombineChromosomes: NewOption(true),
					MatchChromosomes:     NewOption(false),
				}),
			})

			if err != nil {
				t.Fatalf("Optimize for Genome[int] failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("Optimize for Genome[int] exceeded MaxIterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.Score < 0.9 {
				t.Errorf("Optimize for Genome[int] failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.Score)
			}
		})
	})

	t.Run("IterationHook", func(t *testing.T) {
		type Log struct {
			count int
			best  *ScoredCode[int]
		}
		t.Run("parallel", func(t *testing.T) {
			logs := []Log{}
			log_iteration := func(gc int, pop []*ScoredCode[int]) {
				logs = append(logs, Log{count: gc, best: pop[0]})
			}
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NNucleosomes: NewOption(uint(3)),
				NChromosomes: NewOption(uint(2)),
				BaseFactory:  NewOption(base_factory),
			}
			initial_population := []Code[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeGene(opts)
				initial_population = append(initial_population, Code[int]{Gene: NewOption(gene)})
			}

			n_iterations, _, err := Optimize(OptimizationParams[int]{
				IterationHook:     NewOption(log_iteration),
				InitialPopulation: NewOption(initial_population),
				MeasureFitness:    NewOption(measureCodeFitness),
				Mutate:            NewOption(MutateCode),
				MaxIterations:     NewOption(1000),
				ParallelCount:     NewOption(10),
				RecombinationOpts: NewOption(RecombineOptions{
					RecombineGenes:       NewOption(true),
					MatchGenes:           NewOption(false),
					RecombineNucleosomes: NewOption(true),
					MatchNucleosomes:     NewOption(false),
					RecombineChromosomes: NewOption(true),
					MatchChromosomes:     NewOption(false),
				}),
			})

			if err != nil {
				t.Fatalf("Optimize for Gene[int] with IterationHook failed with error: %v", err)
			}

			if len(logs) != n_iterations {
				t.Errorf("Optimize for Gene[int] with IterationHook failed: expected %d log, observed %d", n_iterations, len(logs))
			}

			first_log_score := logs[0].best.Score
			final_log_score := logs[len(logs)-1].best.Score
			if first_log_score > final_log_score {
				t.Errorf("Optimize for Gene[int] with IterationHook failed: first score (%f) > final score (%f)",
					first_log_score, final_log_score)
			}
		})
		t.Run("sequential", func(t *testing.T) {
			t.Parallel()
			logs := []Log{}
			log_iteration := func(gc int, pop []*ScoredCode[int]) {
				logs = append(logs, Log{count: gc, best: pop[0]})
			}
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NNucleosomes: NewOption(uint(3)),
				NChromosomes: NewOption(uint(2)),
				BaseFactory:  NewOption(base_factory),
			}
			initial_population := []Code[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeGene(opts)
				initial_population = append(initial_population, Code[int]{Gene: NewOption(gene)})
			}

			n_iterations, _, err := Optimize(OptimizationParams[int]{
				IterationHook:     NewOption(log_iteration),
				InitialPopulation: NewOption(initial_population),
				MeasureFitness:    NewOption(measureCodeFitness),
				Mutate:            NewOption(MutateCode),
				MaxIterations:     NewOption(1000),
				RecombinationOpts: NewOption(RecombineOptions{
					RecombineGenes:       NewOption(true),
					MatchGenes:           NewOption(false),
					RecombineNucleosomes: NewOption(true),
					MatchNucleosomes:     NewOption(false),
					RecombineChromosomes: NewOption(true),
					MatchChromosomes:     NewOption(false),
				}),
			})

			if err != nil {
				t.Fatalf("Optimize for Gene[int] failed with error: %v", err)
			}

			if len(logs) != n_iterations {
				t.Errorf("Optimize for Gene[int] with IterationHook failed: expected %d log, observed %d", n_iterations, len(logs))
			}

			first_log_score := logs[0].best.Score
			final_log_score := logs[len(logs)-1].best.Score
			if first_log_score > final_log_score {
				t.Errorf("Optimize for Gene[int] with IterationHook failed: first score (%f) > final score (%f)",
					first_log_score, final_log_score)
			}
		})
	})
}

func TestTuneOptimize(t *testing.T) {
	t.Run("Gene", func(t *testing.T) {
		t.Run("cheap", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NNucleosomes: NewOption(uint(3)),
				NChromosomes: NewOption(uint(2)),
				BaseFactory:  NewOption(base_factory),
			}
			initial_population := []Code[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeGene(opts)
				initial_population = append(initial_population, Code[int]{Gene: NewOption(gene)})
			}
			params := OptimizationParams[int]{
				InitialPopulation: NewOption(initial_population),
				MeasureFitness:    NewOption(measureCodeFitness),
				Mutate:            NewOption(MutateCode),
				MaxIterations:     NewOption(1000),
				RecombinationOpts: NewOption(RecombineOptions{
					RecombineGenes:       NewOption(true),
					MatchGenes:           NewOption(false),
					RecombineNucleosomes: NewOption(true),
					MatchNucleosomes:     NewOption(false),
					RecombineChromosomes: NewOption(true),
					MatchChromosomes:     NewOption(false),
				}),
			}

			n_goroutines, err := TuneOptimization(params)
			if err != nil {
				t.Fatalf("TuneGeneOptimization failed with error: %v", err)
			}

			if n_goroutines > 1 {
				t.Errorf("TuneGeneOptimization failed: expected 1, observed %d", n_goroutines)
				res := BenchmarkOptimization(params)
				t.Errorf("Mutate: %d, fitness: %d, copy: %d\n", res.CostOfMutate, res.CostOfMeasureFitness, res.CostOfCopy)
				t.Errorf("(Mutate+fitness)/copy: %d\n", (res.CostOfMutate+res.CostOfMeasureFitness)/res.CostOfCopy)
			}
		})
		t.Run("expensive", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NNucleosomes: NewOption(uint(3)),
				NChromosomes: NewOption(uint(2)),
				BaseFactory:  NewOption(base_factory),
			}
			initial_population := []Code[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeGene(opts)
				initial_population = append(initial_population, Code[int]{Gene: NewOption(gene)})
			}
			params := OptimizationParams[int]{
				InitialPopulation: NewOption(initial_population),
				MeasureFitness:    NewOption(measureCodeFitnessExpensive),
				Mutate:            NewOption(MutateCodeExpensive),
				MaxIterations:     NewOption(1000),
				RecombinationOpts: NewOption(RecombineOptions{
					RecombineGenes:       NewOption(true),
					MatchGenes:           NewOption(false),
					RecombineNucleosomes: NewOption(true),
					MatchNucleosomes:     NewOption(false),
					RecombineChromosomes: NewOption(true),
					MatchChromosomes:     NewOption(false),
				}),
			}

			n_goroutines, err := TuneOptimization(params)
			if err != nil {
				t.Fatalf("TuneGeneOptimization failed with error: %v", err)
			}

			if n_goroutines < 1 {
				t.Errorf("TuneGeneOptimization failed: expected 1, observed %d", n_goroutines)
				res := BenchmarkOptimization(params)
				t.Errorf("Mutate: %d, fitness: %d, copy: %d\n", res.CostOfMutate, res.CostOfMeasureFitness, res.CostOfCopy)
				t.Errorf("(Mutate+fitness)/copy: %d\n", (res.CostOfMutate+res.CostOfMeasureFitness)/res.CostOfCopy)
			}
		})
	})

	t.Run("Nucleosome", func(t *testing.T) {
		t.Run("cheap", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NNucleosomes: NewOption(uint(3)),
				NChromosomes: NewOption(uint(2)),
				BaseFactory:  NewOption(base_factory),
			}
			initial_population := []Code[int]{}
			for i := 0; i < 10; i++ {
				nucleosome, _ := MakeNucleosome(opts)
				initial_population = append(initial_population, Code[int]{Nucleosome: NewOption(nucleosome)})
			}
			params := OptimizationParams[int]{
				InitialPopulation: NewOption(initial_population),
				MeasureFitness:    NewOption(measureCodeFitness),
				Mutate:            NewOption(MutateCode),
				MaxIterations:     NewOption(1000),
				RecombinationOpts: NewOption(RecombineOptions{
					RecombineGenes:       NewOption(true),
					MatchGenes:           NewOption(false),
					RecombineNucleosomes: NewOption(true),
					MatchNucleosomes:     NewOption(false),
					RecombineChromosomes: NewOption(true),
					MatchChromosomes:     NewOption(false),
				}),
			}

			n_goroutines, err := TuneOptimization(params)
			if err != nil {
				t.Fatalf("TuneOptimization for Nucleosome[int] failed with error: %v", err)
			}

			if n_goroutines > 1 {
				t.Errorf("TuneOptimization for Nucleosome[int] failed: expected 1, observed %d", n_goroutines)
				res := BenchmarkOptimization(params)
				t.Errorf("Mutate: %d, fitness: %d, copy: %d\n", res.CostOfMutate, res.CostOfMeasureFitness, res.CostOfCopy)
				t.Errorf("(Mutate+fitness)/copy: %d\n", (res.CostOfMutate+res.CostOfMeasureFitness)/res.CostOfCopy)
			}
		})
		t.Run("expensive", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NNucleosomes: NewOption(uint(3)),
				NChromosomes: NewOption(uint(2)),
				BaseFactory:  NewOption(base_factory),
			}
			initial_population := []Code[int]{}
			for i := 0; i < 10; i++ {
				nucleosome, _ := MakeNucleosome(opts)
				initial_population = append(initial_population, Code[int]{Nucleosome: NewOption(nucleosome)})
			}
			params := OptimizationParams[int]{
				InitialPopulation: NewOption(initial_population),
				MeasureFitness:    NewOption(measureCodeFitnessExpensive),
				Mutate:            NewOption(MutateCodeExpensive),
				MaxIterations:     NewOption(1000),
				RecombinationOpts: NewOption(RecombineOptions{
					RecombineGenes:       NewOption(true),
					MatchGenes:           NewOption(false),
					RecombineNucleosomes: NewOption(true),
					MatchNucleosomes:     NewOption(false),
					RecombineChromosomes: NewOption(true),
					MatchChromosomes:     NewOption(false),
				}),
			}

			n_goroutines, err := TuneOptimization(params)
			if err != nil {
				t.Fatalf("TuneOptimization for Nucleosome[int] failed with error: %v", err)
			}

			if n_goroutines < 1 {
				t.Errorf("TuneOptimization for Nucleosome[int] failed: expected 1, observed %d", n_goroutines)
				res := BenchmarkOptimization(params)
				t.Errorf("Mutate: %d, fitness: %d, copy: %d\n", res.CostOfMutate, res.CostOfMeasureFitness, res.CostOfCopy)
				t.Errorf("(Mutate+fitness)/copy: %d\n", (res.CostOfMutate+res.CostOfMeasureFitness)/res.CostOfCopy)
			}
		})
	})

	t.Run("Chromosome", func(t *testing.T) {
		t.Run("cheap", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NNucleosomes: NewOption(uint(3)),
				NChromosomes: NewOption(uint(2)),
				BaseFactory:  NewOption(base_factory),
			}
			initial_population := []Code[int]{}
			for i := 0; i < 10; i++ {
				chromosome, _ := MakeChromosome(opts)
				initial_population = append(initial_population, Code[int]{Chromosome: NewOption(chromosome)})
			}
			params := OptimizationParams[int]{
				InitialPopulation: NewOption(initial_population),
				MeasureFitness:    NewOption(measureCodeFitness),
				Mutate:            NewOption(MutateCode),
				MaxIterations:     NewOption(1000),
				RecombinationOpts: NewOption(RecombineOptions{
					RecombineGenes:       NewOption(true),
					MatchGenes:           NewOption(false),
					RecombineNucleosomes: NewOption(true),
					MatchNucleosomes:     NewOption(false),
					RecombineChromosomes: NewOption(true),
					MatchChromosomes:     NewOption(false),
				}),
			}

			n_goroutines, err := TuneOptimization(params)
			if err != nil {
				t.Fatalf("TuneOptimization failed with error: %v", err)
			}

			if n_goroutines > 1 {
				t.Errorf("TuneOptimization failed: expected 1, observed %d", n_goroutines)
				res := BenchmarkOptimization(params)
				t.Errorf("Mutate: %d, fitness: %d, copy: %d\n", res.CostOfMutate, res.CostOfMeasureFitness, res.CostOfCopy)
				t.Errorf("(Mutate+fitness)/copy: %d\n", (res.CostOfMutate+res.CostOfMeasureFitness)/res.CostOfCopy)
			}
		})
		t.Run("expensive", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NNucleosomes: NewOption(uint(3)),
				NChromosomes: NewOption(uint(2)),
				BaseFactory:  NewOption(base_factory),
			}
			initial_population := []Code[int]{}
			for i := 0; i < 10; i++ {
				chromosome, _ := MakeChromosome(opts)
				initial_population = append(initial_population, Code[int]{Chromosome: NewOption(chromosome)})
			}
			params := OptimizationParams[int]{
				InitialPopulation: NewOption(initial_population),
				MeasureFitness:    NewOption(measureCodeFitnessExpensive),
				Mutate:            NewOption(MutateCodeExpensive),
				MaxIterations:     NewOption(1000),
				RecombinationOpts: NewOption(RecombineOptions{
					RecombineGenes:       NewOption(true),
					MatchGenes:           NewOption(false),
					RecombineNucleosomes: NewOption(true),
					MatchNucleosomes:     NewOption(false),
					RecombineChromosomes: NewOption(true),
					MatchChromosomes:     NewOption(false),
				}),
			}

			n_goroutines, err := TuneOptimization(params)
			if err != nil {
				t.Fatalf("TuneOptimization failed with error: %v", err)
			}

			if n_goroutines < 1 {
				t.Errorf("TuneOptimization failed: expected 1, observed %d", n_goroutines)
				res := BenchmarkOptimization(params)
				t.Errorf("Mutate: %d, fitness: %d, copy: %d\n", res.CostOfMutate, res.CostOfMeasureFitness, res.CostOfCopy)
				t.Errorf("(Mutate+fitness)/copy: %d\n", (res.CostOfMutate+res.CostOfMeasureFitness)/res.CostOfCopy)
			}
		})
	})

	t.Run("Genome", func(t *testing.T) {
		t.Run("cheap", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NNucleosomes: NewOption(uint(3)),
				NChromosomes: NewOption(uint(2)),
				BaseFactory:  NewOption(base_factory),
			}
			initial_population := []Code[int]{}
			for i := 0; i < 10; i++ {
				genome, _ := MakeGenome(opts)
				initial_population = append(initial_population, Code[int]{Genome: NewOption(genome)})
			}
			params := OptimizationParams[int]{
				InitialPopulation: NewOption(initial_population),
				MeasureFitness:    NewOption(measureCodeFitness),
				Mutate:            NewOption(MutateCode),
				MaxIterations:     NewOption(1000),
				RecombinationOpts: NewOption(RecombineOptions{
					RecombineGenes:       NewOption(true),
					MatchGenes:           NewOption(false),
					RecombineNucleosomes: NewOption(true),
					MatchNucleosomes:     NewOption(false),
					RecombineChromosomes: NewOption(true),
					MatchChromosomes:     NewOption(false),
				}),
			}

			n_goroutines, err := TuneOptimization(params)
			if err != nil {
				t.Fatalf("TuneOptimization failed with error: %v", err)
			}

			if n_goroutines > 1 {
				t.Errorf("TuneOptimization failed: expected 1, observed %d", n_goroutines)
				res := BenchmarkOptimization(params)
				t.Errorf("Mutate: %d, fitness: %d, copy: %d\n", res.CostOfMutate, res.CostOfMeasureFitness, res.CostOfCopy)
				t.Errorf("(Mutate+fitness)/copy: %d\n", (res.CostOfMutate+res.CostOfMeasureFitness)/res.CostOfCopy)
			}
		})
		t.Run("expensive", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NNucleosomes: NewOption(uint(3)),
				NChromosomes: NewOption(uint(2)),
				BaseFactory:  NewOption(base_factory),
			}
			initial_population := []Code[int]{}
			for i := 0; i < 10; i++ {
				genome, _ := MakeGenome(opts)
				initial_population = append(initial_population, Code[int]{Genome: NewOption(genome)})
			}
			params := OptimizationParams[int]{
				InitialPopulation: NewOption(initial_population),
				MeasureFitness:    NewOption(measureCodeFitnessExpensive),
				Mutate:            NewOption(MutateCodeExpensive),
				MaxIterations:     NewOption(1000),
				RecombinationOpts: NewOption(RecombineOptions{
					RecombineGenes:       NewOption(true),
					MatchGenes:           NewOption(false),
					RecombineNucleosomes: NewOption(true),
					MatchNucleosomes:     NewOption(false),
					RecombineChromosomes: NewOption(true),
					MatchChromosomes:     NewOption(false),
				}),
			}

			n_goroutines, err := TuneOptimization(params)
			if err != nil {
				t.Fatalf("TuneOptimization failed with error: %v", err)
			}

			if n_goroutines < 1 {
				t.Errorf("TuneOptimization failed: expected 1, observed %d", n_goroutines)
				res := BenchmarkOptimization(params)
				t.Errorf("Mutate: %d, fitness: %d, copy: %d\n", res.CostOfMutate, res.CostOfMeasureFitness, res.CostOfCopy)
				t.Errorf("(Mutate+fitness)/copy: %d\n", (res.CostOfMutate+res.CostOfMeasureFitness)/res.CostOfCopy)
			}
		})
	})
}

func BenchmarkOptimize(b *testing.B) {
	base_factory := func() int { return RandomInt(-10, 10) }
	make_opts := MakeOptions[int]{
		NBases:       NewOption(uint(5)),
		NGenes:       NewOption(uint(4)),
		NNucleosomes: NewOption(uint(3)),
		NChromosomes: NewOption(uint(2)),
		BaseFactory:  NewOption(base_factory),
	}
	population := []Code[int]{}
	for i := 0; i < 100; i++ {
		gene, err := MakeGene(make_opts)
		if err != nil {
			b.Fatal(err)
		}
		population = append(population, Code[int]{
			Gene: NewOption(gene),
		})
	}
	params := OptimizationParams[int]{
		MeasureFitness:       NewOption(measureCodeFitness),
		Mutate:               NewOption(MutateCode),
		InitialPopulation:    NewOption(population),
		MaxIterations:        NewOption(100),
		PopulationSize:       NewOption(100),
		ParentsPerGeneration: NewOption(10),
	}
	b.Run("GeneSequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := Optimize(params)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("GeneParallel", func(b *testing.B) {
		params.ParallelCount = NewOption(10)
		for i := 0; i < b.N; i++ {
			_, _, err := Optimize(params)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	population = []Code[int]{}
	for i := 0; i < 100; i++ {
		nucleosome, err := MakeNucleosome(make_opts)
		if err != nil {
			b.Fatal(err)
		}
		population = append(population, Code[int]{
			Nucleosome: NewOption(nucleosome),
		})
	}
	params = OptimizationParams[int]{
		MeasureFitness:       NewOption(measureCodeFitness),
		Mutate:               NewOption(MutateCode),
		InitialPopulation:    NewOption(population),
		MaxIterations:        NewOption(100),
		PopulationSize:       NewOption(100),
		ParentsPerGeneration: NewOption(10),
	}
	b.Run("NucleosomeSequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := Optimize(params)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("NucleosomeParallel", func(b *testing.B) {
		params.ParallelCount = NewOption(10)
		for i := 0; i < b.N; i++ {
			_, _, err := Optimize(params)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	population = []Code[int]{}
	for i := 0; i < 100; i++ {
		chromosome, err := MakeChromosome(make_opts)
		if err != nil {
			b.Fatal(err)
		}
		population = append(population, Code[int]{
			Chromosome: NewOption(chromosome),
		})
	}
	params = OptimizationParams[int]{
		MeasureFitness:       NewOption(measureCodeFitness),
		Mutate:               NewOption(MutateCode),
		InitialPopulation:    NewOption(population),
		MaxIterations:        NewOption(100),
		PopulationSize:       NewOption(100),
		ParentsPerGeneration: NewOption(10),
	}
	b.Run("ChromosomeSequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := Optimize(params)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("ChromosomeParallel", func(b *testing.B) {
		params.ParallelCount = NewOption(10)
		for i := 0; i < b.N; i++ {
			_, _, err := Optimize(params)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	population = []Code[int]{}
	for i := 0; i < 100; i++ {
		genome, err := MakeGenome(make_opts)
		if err != nil {
			b.Fatal(err)
		}
		population = append(population, Code[int]{
			Genome: NewOption(genome),
		})
	}
	params = OptimizationParams[int]{
		MeasureFitness:       NewOption(measureCodeFitness),
		Mutate:               NewOption(MutateCode),
		InitialPopulation:    NewOption(population),
		MaxIterations:        NewOption(100),
		PopulationSize:       NewOption(100),
		ParentsPerGeneration: NewOption(10),
	}
	b.Run("GenomeSequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := Optimize(params)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("GenomeParallel", func(b *testing.B) {
		params.ParallelCount = NewOption(10)
		for i := 0; i < b.N; i++ {
			_, _, err := Optimize(params)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
