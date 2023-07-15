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

func MutateAllele(allele *Allele[int]) {
	allele.Mu.Lock()
	defer allele.Mu.Unlock()
	for _, gene := range allele.Genes {
		MutateGene(gene)
	}
}

func MutateChromosome(chromosome *Chromosome[int]) {
	chromosome.Mu.Lock()
	defer chromosome.Mu.Unlock()
	for _, allele := range chromosome.Alleles {
		MutateAllele(allele)
	}
}

func MutateGenome(genome *Genome[int]) {
	genome.Mu.Lock()
	defer genome.Mu.Unlock()
	for _, chromosome := range genome.Chromosomes {
		MutateChromosome(chromosome)
	}
}

func MutateCode(code Code[int]) {
	if code.Gene.ok() {
		MutateGene(code.Gene.val)
	}
	if code.Allele.ok() {
		MutateAllele(code.Allele.val)
	}
	if code.Chromosome.ok() {
		MutateChromosome(code.Chromosome.val)
	}
	if code.Genome.ok() {
		MutateGenome(code.Genome.val)
	}
}

func measureGeneFitness(gene *Gene[int]) float64 {
	gene.Mu.RLock()
	defer gene.Mu.RUnlock()
	total := reduce(gene.Bases, func(a int, b int) int { return a + b })
	return 1.0 / (math.Abs(float64(total-target)) + 1.0)
}

func measureAlleleFitness(allele *Allele[int]) float64 {
	allele.Mu.RLock()
	defer allele.Mu.RUnlock()
	total := 0
	for _, gene := range allele.Genes {
		total += reduce(gene.Bases, func(a int, b int) int { return a + b })
	}
	return 1.0 / (math.Abs(float64(total-target)) + 1.0)
}

func measureChromosomeFitness(chromosome *Chromosome[int]) float64 {
	chromosome.Mu.RLock()
	defer chromosome.Mu.RUnlock()
	total := 0
	for _, allele := range chromosome.Alleles {
		for _, gene := range allele.Genes {
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
		for _, allele := range chromosome.Alleles {
			for _, gene := range allele.Genes {
				total += reduce(gene.Bases, func(a int, b int) int { return a + b })
			}
		}
	}
	return 1.0 / (math.Abs(float64(total-target)) + 1.0)
}

func measureCodeFitness(code Code[int]) float64 {
	fitness := 0.0
	fitness_count := 0
	if code.Gene.ok() {
		fitness += measureGeneFitness(code.Gene.val)
		fitness_count++
	}
	if code.Allele.ok() {
		fitness += measureAlleleFitness(code.Allele.val)
		fitness_count++
	}
	if code.Chromosome.ok() {
		fitness += measureChromosomeFitness(code.Chromosome.val)
		fitness_count++
	}
	if code.Genome.ok() {
		fitness += measureGenomeFitness(code.Genome.val)
		fitness_count++
	}

	return fitness / float64(fitness_count)
}

func MutateCodeExpensive(code Code[int]) {
	val := 42.0
	for i := 0; i < 1000; i++ {
		val /= 6.9
	}
	if code.Gene.ok() {
		MutateGene(code.Gene.val)
	}
	if code.Allele.ok() {
		MutateAllele(code.Allele.val)
	}
	if code.Chromosome.ok() {
		MutateChromosome(code.Chromosome.val)
	}
	if code.Genome.ok() {
		MutateGenome(code.Genome.val)
	}
}

func measureCodeFitnessExpensive(code Code[int]) float64 {
	val := 42.0
	for i := 0; i < 1000; i++ {
		val /= 6.9
	}
	fitness := 0.0
	fitness_count := 0
	if code.Gene.ok() {
		fitness += measureGeneFitness(code.Gene.val)
		fitness_count++
	}
	if code.Allele.ok() {
		fitness += measureAlleleFitness(code.Allele.val)
		fitness_count++
	}
	if code.Chromosome.ok() {
		fitness += measureChromosomeFitness(code.Chromosome.val)
		fitness_count++
	}
	if code.Genome.ok() {
		fitness += measureGenomeFitness(code.Genome.val)
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
				NAlleles:     NewOption(uint(3)),
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
					RecombineAlleles:     NewOption(true),
					MatchAlleles:         NewOption(false),
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
				NAlleles:     NewOption(uint(3)),
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
					RecombineAlleles:     NewOption(true),
					MatchAlleles:         NewOption(false),
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

	t.Run("Allele", func(t *testing.T) {
		t.Run("parallel", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NAlleles:     NewOption(uint(3)),
				NChromosomes: NewOption(uint(2)),
				BaseFactory:  NewOption(base_factory),
			}
			initial_population := []Code[int]{}
			for i := 0; i < 10; i++ {
				allele, _ := MakeAllele(opts)
				initial_population = append(initial_population, Code[int]{Allele: NewOption(allele)})
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
					RecombineAlleles:     NewOption(true),
					MatchAlleles:         NewOption(false),
					RecombineChromosomes: NewOption(true),
					MatchChromosomes:     NewOption(false),
				}),
			})

			if err != nil {
				t.Fatalf("Optimize for Allele[int] failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("Optimize for Allele[int] exceeded MaxIterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.Score < 0.9 {
				t.Errorf("Optimize for Allele[int] failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.Score)
			}
		})
		t.Run("sequential", func(t *testing.T) {
			t.Parallel()
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NAlleles:     NewOption(uint(3)),
				NChromosomes: NewOption(uint(2)),
				BaseFactory:  NewOption(base_factory),
			}
			initial_population := []Code[int]{}
			for i := 0; i < 10; i++ {
				allele, _ := MakeAllele(opts)
				initial_population = append(initial_population, Code[int]{Allele: NewOption(allele)})
			}

			n_iterations, final_population, err := Optimize(OptimizationParams[int]{
				InitialPopulation: NewOption(initial_population),
				MeasureFitness:    NewOption(measureCodeFitness),
				Mutate:            NewOption(MutateCode),
				MaxIterations:     NewOption(1000),
				RecombinationOpts: NewOption(RecombineOptions{
					RecombineGenes:       NewOption(true),
					MatchGenes:           NewOption(false),
					RecombineAlleles:     NewOption(true),
					MatchAlleles:         NewOption(false),
					RecombineChromosomes: NewOption(true),
					MatchChromosomes:     NewOption(false),
				}),
			})

			if err != nil {
				t.Fatalf("Optimize for Allele[int] failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("Optimize for Allele[int] exceeded MaxIterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.Score < 0.9 {
				t.Errorf("Optimize for Allele[int] failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.Score)
			}
		})
	})

	t.Run("Chromosome", func(t *testing.T) {
		t.Run("parallel", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NAlleles:     NewOption(uint(3)),
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
					RecombineAlleles:     NewOption(true),
					MatchAlleles:         NewOption(false),
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
				NAlleles:     NewOption(uint(3)),
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
					RecombineAlleles:     NewOption(true),
					MatchAlleles:         NewOption(false),
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

	t.Run("Allele", func(t *testing.T) {
		t.Run("parallel", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NAlleles:     NewOption(uint(3)),
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
					RecombineAlleles:     NewOption(true),
					MatchAlleles:         NewOption(false),
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
				NAlleles:     NewOption(uint(3)),
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
					RecombineAlleles:     NewOption(true),
					MatchAlleles:         NewOption(false),
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
}

func TestTuneOptimize(t *testing.T) {
	t.Run("Gene", func(t *testing.T) {
		t.Run("cheap", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NAlleles:     NewOption(uint(3)),
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
					RecombineAlleles:     NewOption(true),
					MatchAlleles:         NewOption(false),
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
				NAlleles:     NewOption(uint(3)),
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
					RecombineAlleles:     NewOption(true),
					MatchAlleles:         NewOption(false),
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
	t.Run("Allele", func(t *testing.T) {
		t.Run("cheap", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				NBases:       NewOption(uint(5)),
				NGenes:       NewOption(uint(4)),
				NAlleles:     NewOption(uint(3)),
				NChromosomes: NewOption(uint(2)),
				BaseFactory:  NewOption(base_factory),
			}
			initial_population := []Code[int]{}
			for i := 0; i < 10; i++ {
				allele, _ := MakeAllele(opts)
				initial_population = append(initial_population, Code[int]{Allele: NewOption(allele)})
			}
			params := OptimizationParams[int]{
				InitialPopulation: NewOption(initial_population),
				MeasureFitness:    NewOption(measureCodeFitness),
				Mutate:            NewOption(MutateCode),
				MaxIterations:     NewOption(1000),
				RecombinationOpts: NewOption(RecombineOptions{
					RecombineGenes:       NewOption(true),
					MatchGenes:           NewOption(false),
					RecombineAlleles:     NewOption(true),
					MatchAlleles:         NewOption(false),
					RecombineChromosomes: NewOption(true),
					MatchChromosomes:     NewOption(false),
				}),
			}

			n_goroutines, err := TuneOptimization(params)
			if err != nil {
				t.Fatalf("TuneOptimization for Allele[int] failed with error: %v", err)
			}

			if n_goroutines > 1 {
				t.Errorf("TuneOptimization for Allele[int] failed: expected 1, observed %d", n_goroutines)
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
				NAlleles:     NewOption(uint(3)),
				NChromosomes: NewOption(uint(2)),
				BaseFactory:  NewOption(base_factory),
			}
			initial_population := []Code[int]{}
			for i := 0; i < 10; i++ {
				allele, _ := MakeAllele(opts)
				initial_population = append(initial_population, Code[int]{Allele: NewOption(allele)})
			}
			params := OptimizationParams[int]{
				InitialPopulation: NewOption(initial_population),
				MeasureFitness:    NewOption(measureCodeFitnessExpensive),
				Mutate:            NewOption(MutateCodeExpensive),
				MaxIterations:     NewOption(1000),
				RecombinationOpts: NewOption(RecombineOptions{
					RecombineGenes:       NewOption(true),
					MatchGenes:           NewOption(false),
					RecombineAlleles:     NewOption(true),
					MatchAlleles:         NewOption(false),
					RecombineChromosomes: NewOption(true),
					MatchChromosomes:     NewOption(false),
				}),
			}

			n_goroutines, err := TuneOptimization(params)
			if err != nil {
				t.Fatalf("TuneOptimization for Allele[int] failed with error: %v", err)
			}

			if n_goroutines < 1 {
				t.Errorf("TuneOptimization for Allele[int] failed: expected 1, observed %d", n_goroutines)
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
				NAlleles:     NewOption(uint(3)),
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
					RecombineAlleles:     NewOption(true),
					MatchAlleles:         NewOption(false),
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
				NAlleles:     NewOption(uint(3)),
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
					RecombineAlleles:     NewOption(true),
					MatchAlleles:         NewOption(false),
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
				NAlleles:     NewOption(uint(3)),
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
					RecombineAlleles:     NewOption(true),
					MatchAlleles:         NewOption(false),
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
				NAlleles:     NewOption(uint(3)),
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
					RecombineAlleles:     NewOption(true),
					MatchAlleles:         NewOption(false),
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
