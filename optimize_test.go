package genetics

import (
	"math"
	"math/rand"
	"testing"
)

var target = 12345

func MutateGene(gene *Gene[int]) {
	gene.mu.Lock()
	defer gene.mu.Unlock()
	for i := 0; i < len(gene.bases); i++ {
		val := rand.Float64()
		if val <= 0.1 {
			gene.bases[i] /= RandomInt(1, 3)
		} else if val <= 0.2 {
			gene.bases[i] *= RandomInt(1, 3)
		} else if val <= 0.6 {
			gene.bases[i] += RandomInt(0, 11)
		} else {
			gene.bases[i] -= RandomInt(0, 11)
		}
	}
}

func MutateAllele(allele *Allele[int]) {
	allele.mu.Lock()
	defer allele.mu.Unlock()
	for _, gene := range allele.genes {
		MutateGene(gene)
	}
}

func MutateChromosome(chromosome *Chromosome[int]) {
	chromosome.mu.Lock()
	defer chromosome.mu.Unlock()
	for _, allele := range chromosome.alleles {
		MutateAllele(allele)
	}
}

func MutateGenome(genome *Genome[int]) {
	genome.mu.Lock()
	defer genome.mu.Unlock()
	for _, chromosome := range genome.chromosomes {
		MutateChromosome(chromosome)
	}
}

func MeasureGeneFitness(gene *Gene[int]) float64 {
	gene.mu.RLock()
	defer gene.mu.RUnlock()
	total := Reduce(gene.bases, func(a int, b int) int { return a + b })
	return 1.0 / (math.Abs(float64(total-target)) + 1.0)
}

func MeasureAlleleFitness(allele *Allele[int]) float64 {
	allele.mu.RLock()
	defer allele.mu.RUnlock()
	total := 0
	for _, gene := range allele.genes {
		total += Reduce(gene.bases, func(a int, b int) int { return a + b })
	}
	return 1.0 / (math.Abs(float64(total-target)) + 1.0)
}

func MeasureChromosomeFitness(chromosome *Chromosome[int]) float64 {
	chromosome.mu.RLock()
	defer chromosome.mu.RUnlock()
	total := 0
	for _, allele := range chromosome.alleles {
		for _, gene := range allele.genes {
			total += Reduce(gene.bases, func(a int, b int) int { return a + b })
		}
	}
	return 1.0 / (math.Abs(float64(total-target)) + 1.0)
}

func MeasureGenomeFitness(genome *Genome[int]) float64 {
	genome.mu.RLock()
	defer genome.mu.RUnlock()
	total := 0
	for _, chromosome := range genome.chromosomes {
		for _, allele := range chromosome.alleles {
			for _, gene := range allele.genes {
				total += Reduce(gene.bases, func(a int, b int) int { return a + b })
			}
		}
	}
	return 1.0 / (math.Abs(float64(total-target)) + 1.0)
}

func MutateGeneExpensive(gene *Gene[int]) {
	gene.mu.Lock()
	defer gene.mu.Unlock()
	val := 42.0
	for i := 0; i < 1000; i++ {
		val /= 6.9
	}
	for i := 0; i < len(gene.bases); i++ {
		val = rand.Float64()
		if val <= 0.1 {
			gene.bases[i] /= RandomInt(1, 3)
		} else if val <= 0.2 {
			gene.bases[i] *= RandomInt(1, 3)
		} else if val <= 0.6 {
			gene.bases[i] += RandomInt(0, 11)
		} else {
			gene.bases[i] -= RandomInt(0, 11)
		}
	}
}

func MutateAlleleExpensive(allele *Allele[int]) {
	allele.mu.Lock()
	defer allele.mu.Unlock()
	for _, gene := range allele.genes {
		MutateGeneExpensive(gene)
	}
}

func MutateChromosomeExpensive(chromosome *Chromosome[int]) {
	chromosome.mu.Lock()
	defer chromosome.mu.Unlock()
	for _, allele := range chromosome.alleles {
		MutateAlleleExpensive(allele)
	}
}

func MutateGenomeExpensive(genome *Genome[int]) {
	genome.mu.Lock()
	defer genome.mu.Unlock()
	for _, chromosome := range genome.chromosomes {
		MutateChromosomeExpensive(chromosome)
	}
}

func MeasureGeneFitnessExpensive(gene *Gene[int]) float64 {
	gene.mu.RLock()
	defer gene.mu.RUnlock()
	val := 42.0
	for i := 0; i < 1000; i++ {
		val /= 6.9
	}
	total := Reduce(gene.bases, func(a int, b int) int { return a + b })
	return 1.0 / (math.Abs(float64(total-target)) + 1.0)
}

func MeasureAlleleFitnessExpensive(allele *Allele[int]) float64 {
	allele.mu.RLock()
	defer allele.mu.RUnlock()
	val := 42.0
	for i := 0; i < 1000; i++ {
		val /= 6.9
	}
	total := 0
	for _, gene := range allele.genes {
		total += Reduce(gene.bases, func(a int, b int) int { return a + b })
	}
	return 1.0 / (math.Abs(float64(total-target)) + 1.0)
}

func MeasureChromosomeFitnessExpensive(chromosome *Chromosome[int]) float64 {
	chromosome.mu.RLock()
	defer chromosome.mu.RUnlock()
	val := 42.0
	for i := 0; i < 1000; i++ {
		val /= 6.9
	}
	total := 0
	for _, allele := range chromosome.alleles {
		for _, gene := range allele.genes {
			total += Reduce(gene.bases, func(a int, b int) int { return a + b })
		}
	}
	return 1.0 / (math.Abs(float64(total-target)) + 1.0)
}

func MeasureGenomeFitnessExpensive(genome *Genome[int]) float64 {
	genome.mu.RLock()
	defer genome.mu.RUnlock()
	val := 42.0
	for i := 0; i < 1000; i++ {
		val /= 6.9
	}
	total := 0
	for _, chromosome := range genome.chromosomes {
		for _, allele := range chromosome.alleles {
			for _, gene := range allele.genes {
				total += Reduce(gene.bases, func(a int, b int) int { return a + b })
			}
		}
	}
	return 1.0 / (math.Abs(float64(total-target)) + 1.0)
}

func TestOptimize(t *testing.T) {
	t.Run("Gene", func(t *testing.T) {
		t.Run("parallel", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				n_bases:      NewOption(uint(5)),
				base_factory: NewOption(base_factory),
			}
			initial_population := []*Gene[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeGene(opts)
				initial_population = append(initial_population, gene)
			}

			n_iterations, final_population, err := OptimizeGene(OptimizationParams[int, Gene[int]]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(MeasureGeneFitness),
				mutate:             NewOption(MutateGene),
				max_iterations:     NewOption(1000),
				parallel_count:     NewOption(10),
			})

			if err != nil {
				t.Fatalf("OptimizeGene failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("OptimizeGene exceeded max_iterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.score < 0.9 {
				t.Errorf("OptimizeGene failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.score)
			}
		})
		t.Run("sequential", func(t *testing.T) {
			t.Parallel()
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				n_bases:      NewOption(uint(5)),
				base_factory: NewOption(base_factory),
			}
			initial_population := []*Gene[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeGene(opts)
				initial_population = append(initial_population, gene)
			}

			n_iterations, final_population, err := OptimizeGene(OptimizationParams[int, Gene[int]]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(MeasureGeneFitness),
				mutate:             NewOption(MutateGene),
				max_iterations:     NewOption(1000),
			})

			if err != nil {
				t.Fatalf("OptimizeGene failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("OptimizeGene exceeded max_iterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.score < 0.9 {
				t.Errorf("OptimizeGene failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.score)
			}
		})
	})

	t.Run("Allele", func(t *testing.T) {
		t.Run("parallel", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				n_bases:      NewOption(uint(5)),
				n_genes:      NewOption(uint(4)),
				base_factory: NewOption(base_factory),
			}
			initial_population := []*Allele[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeAllele(opts)
				initial_population = append(initial_population, gene)
			}

			n_iterations, final_population, err := OptimizeAllele(OptimizationParams[int, Allele[int]]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(MeasureAlleleFitness),
				mutate:             NewOption(MutateAllele),
				max_iterations:     NewOption(1000),
				parallel_count:     NewOption(10),
				recombination_opts: NewOption(RecombineOptions{
					recombine_genes:       NewOption(true),
					match_genes:           NewOption(false),
					recombine_alleles:     NewOption(true),
					match_alleles:         NewOption(false),
					recombine_chromosomes: NewOption(true),
					match_chromosomes:     NewOption(false),
				}),
			})

			if err != nil {
				t.Fatalf("OptimizeAllele failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("OptimizeAllele exceeded max_iterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.score < 0.9 {
				t.Errorf("OptimizeAllele failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.score)
			}
		})
		t.Run("sequential", func(t *testing.T) {
			t.Parallel()
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				n_bases:      NewOption(uint(5)),
				n_genes:      NewOption(uint(4)),
				base_factory: NewOption(base_factory),
			}
			initial_population := []*Allele[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeAllele(opts)
				initial_population = append(initial_population, gene)
			}

			n_iterations, final_population, err := OptimizeAllele(OptimizationParams[int, Allele[int]]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(MeasureAlleleFitness),
				mutate:             NewOption(MutateAllele),
				max_iterations:     NewOption(1000),
				recombination_opts: NewOption(RecombineOptions{
					recombine_genes:       NewOption(true),
					match_genes:           NewOption(false),
					recombine_alleles:     NewOption(true),
					match_alleles:         NewOption(false),
					recombine_chromosomes: NewOption(true),
					match_chromosomes:     NewOption(false),
				}),
			})

			if err != nil {
				t.Fatalf("OptimizeAllele failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("OptimizeAllele exceeded max_iterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.score < 0.9 {
				t.Errorf("OptimizeAllele failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.score)
			}
		})
	})

	t.Run("Chromosome", func(t *testing.T) {
		t.Run("parallel", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				n_bases:      NewOption(uint(5)),
				n_genes:      NewOption(uint(4)),
				n_alleles:    NewOption(uint(3)),
				base_factory: NewOption(base_factory),
			}
			initial_population := []*Chromosome[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeChromosome(opts)
				initial_population = append(initial_population, gene)
			}

			n_iterations, final_population, err := OptimizeChromosome(OptimizationParams[int, Chromosome[int]]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(MeasureChromosomeFitness),
				mutate:             NewOption(MutateChromosome),
				max_iterations:     NewOption(1000),
				parallel_count:     NewOption(10),
				recombination_opts: NewOption(RecombineOptions{
					recombine_genes:       NewOption(true),
					match_genes:           NewOption(false),
					recombine_alleles:     NewOption(true),
					match_alleles:         NewOption(false),
					recombine_chromosomes: NewOption(true),
					match_chromosomes:     NewOption(false),
				}),
			})

			if err != nil {
				t.Fatalf("OptimizeChromosome failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("OptimizeChromosome exceeded max_iterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.score < 0.9 {
				t.Errorf("OptimizeChromosome failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.score)
			}
		})
		t.Run("sequential", func(t *testing.T) {
			t.Parallel()
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				n_bases:      NewOption(uint(5)),
				n_genes:      NewOption(uint(4)),
				n_alleles:    NewOption(uint(3)),
				base_factory: NewOption(base_factory),
			}
			initial_population := []*Chromosome[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeChromosome(opts)
				initial_population = append(initial_population, gene)
			}

			n_iterations, final_population, err := OptimizeChromosome(OptimizationParams[int, Chromosome[int]]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(MeasureChromosomeFitness),
				mutate:             NewOption(MutateChromosome),
				max_iterations:     NewOption(1000),
				recombination_opts: NewOption(RecombineOptions{
					recombine_genes:       NewOption(true),
					match_genes:           NewOption(false),
					recombine_alleles:     NewOption(true),
					match_alleles:         NewOption(false),
					recombine_chromosomes: NewOption(true),
					match_chromosomes:     NewOption(false),
				}),
			})

			if err != nil {
				t.Fatalf("OptimizeChromosome failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("OptimizeChromosome exceeded max_iterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.score < 0.9 {
				t.Errorf("OptimizeChromosome failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.score)
			}
		})
	})

	t.Run("Genome", func(t *testing.T) {
		t.Run("parallel", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				n_bases:       NewOption(uint(5)),
				n_genes:       NewOption(uint(4)),
				n_alleles:     NewOption(uint(3)),
				n_chromosomes: NewOption(uint(2)),
				base_factory:  NewOption(base_factory),
			}
			initial_population := []*Genome[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeGenome(opts)
				initial_population = append(initial_population, gene)
			}

			n_iterations, final_population, err := OptimizeGenome(OptimizationParams[int, Genome[int]]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(MeasureGenomeFitness),
				mutate:             NewOption(MutateGenome),
				max_iterations:     NewOption(1000),
				parallel_count:     NewOption(10),
				recombination_opts: NewOption(RecombineOptions{
					recombine_genes:       NewOption(true),
					match_genes:           NewOption(false),
					recombine_alleles:     NewOption(true),
					match_alleles:         NewOption(false),
					recombine_chromosomes: NewOption(true),
					match_chromosomes:     NewOption(false),
				}),
			})

			if err != nil {
				t.Fatalf("OptimizeGenome failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("OptimizeGenome exceeded max_iterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.score < 0.9 {
				t.Errorf("OptimizeGenome failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.score)
			}
		})
		t.Run("sequential", func(t *testing.T) {
			t.Parallel()
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				n_bases:       NewOption(uint(5)),
				n_genes:       NewOption(uint(4)),
				n_alleles:     NewOption(uint(3)),
				n_chromosomes: NewOption(uint(2)),
				base_factory:  NewOption(base_factory),
			}
			initial_population := []*Genome[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeGenome(opts)
				initial_population = append(initial_population, gene)
			}

			n_iterations, final_population, err := OptimizeGenome(OptimizationParams[int, Genome[int]]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(MeasureGenomeFitness),
				mutate:             NewOption(MutateGenome),
				max_iterations:     NewOption(1000),
				recombination_opts: NewOption(RecombineOptions{
					recombine_genes:       NewOption(true),
					match_genes:           NewOption(false),
					recombine_alleles:     NewOption(true),
					match_alleles:         NewOption(false),
					recombine_chromosomes: NewOption(true),
					match_chromosomes:     NewOption(false),
				}),
			})

			if err != nil {
				t.Fatalf("OptimizeGenome failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("OptimizeGenome exceeded max_iterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.score < 0.9 {
				t.Errorf("OptimizeGenome failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.score)
			}
		})
	})
}

func TestTuneOptimize(t *testing.T) {
	t.Run("Gene", func(t *testing.T) {
		t.Run("cheap", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				n_bases:       NewOption(uint(5)),
				n_genes:       NewOption(uint(4)),
				n_alleles:     NewOption(uint(3)),
				n_chromosomes: NewOption(uint(2)),
				base_factory:  NewOption(base_factory),
			}
			initial_population := []*Gene[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeGene(opts)
				initial_population = append(initial_population, gene)
			}
			params := OptimizationParams[int, Gene[int]]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(MeasureGeneFitness),
				mutate:             NewOption(MutateGene),
				max_iterations:     NewOption(1000),
				recombination_opts: NewOption(RecombineOptions{
					recombine_genes:       NewOption(true),
					match_genes:           NewOption(false),
					recombine_alleles:     NewOption(true),
					match_alleles:         NewOption(false),
					recombine_chromosomes: NewOption(true),
					match_chromosomes:     NewOption(false),
				}),
			}

			n_goroutines, err := TuneGeneOptimization(params)
			if err != nil {
				t.Fatalf("TuneGeneOptimization failed with error: %v", err)
			}

			if n_goroutines > 1 {
				t.Errorf("TuneGeneOptimization failed: expected 1, observed %d", n_goroutines)
				res := benchmarkGeneOptimization(params)
				t.Errorf("mutate: %d, fitness: %d, copy: %d\n", res.cost_of_mutate, res.cost_of_measure_fitness, res.cost_of_copy)
				t.Errorf("(mutate+fitness)/copy: %d\n", (res.cost_of_mutate+res.cost_of_measure_fitness)/res.cost_of_copy)
			}
		})
		t.Run("expensive", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				n_bases:       NewOption(uint(5)),
				n_genes:       NewOption(uint(4)),
				n_alleles:     NewOption(uint(3)),
				n_chromosomes: NewOption(uint(2)),
				base_factory:  NewOption(base_factory),
			}
			initial_population := []*Gene[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeGene(opts)
				initial_population = append(initial_population, gene)
			}
			params := OptimizationParams[int, Gene[int]]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(MeasureGeneFitnessExpensive),
				mutate:             NewOption(MutateGeneExpensive),
				max_iterations:     NewOption(1000),
				recombination_opts: NewOption(RecombineOptions{
					recombine_genes:       NewOption(true),
					match_genes:           NewOption(false),
					recombine_alleles:     NewOption(true),
					match_alleles:         NewOption(false),
					recombine_chromosomes: NewOption(true),
					match_chromosomes:     NewOption(false),
				}),
			}

			n_goroutines, err := TuneGeneOptimization(params)
			if err != nil {
				t.Fatalf("TuneGeneOptimization failed with error: %v", err)
			}

			if n_goroutines < 1 {
				t.Errorf("TuneGeneOptimization failed: expected 1, observed %d", n_goroutines)
				res := benchmarkGeneOptimization(params)
				t.Errorf("mutate: %d, fitness: %d, copy: %d\n", res.cost_of_mutate, res.cost_of_measure_fitness, res.cost_of_copy)
				t.Errorf("(mutate+fitness)/copy: %d\n", (res.cost_of_mutate+res.cost_of_measure_fitness)/res.cost_of_copy)
			}
		})
	})
	t.Run("Allele", func(t *testing.T) {
		t.Run("cheap", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				n_bases:       NewOption(uint(5)),
				n_genes:       NewOption(uint(4)),
				n_alleles:     NewOption(uint(3)),
				n_chromosomes: NewOption(uint(2)),
				base_factory:  NewOption(base_factory),
			}
			initial_population := []*Allele[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeAllele(opts)
				initial_population = append(initial_population, gene)
			}
			params := OptimizationParams[int, Allele[int]]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(MeasureAlleleFitness),
				mutate:             NewOption(MutateAllele),
				max_iterations:     NewOption(1000),
				recombination_opts: NewOption(RecombineOptions{
					recombine_genes:       NewOption(true),
					match_genes:           NewOption(false),
					recombine_alleles:     NewOption(true),
					match_alleles:         NewOption(false),
					recombine_chromosomes: NewOption(true),
					match_chromosomes:     NewOption(false),
				}),
			}

			n_goroutines, err := TuneAlleleOptimization(params)
			if err != nil {
				t.Fatalf("TuneAlleleOptimization failed with error: %v", err)
			}

			if n_goroutines > 1 {
				t.Errorf("TuneAlleleOptimization failed: expected 1, observed %d", n_goroutines)
				res := benchmarkAlleleOptimization(params)
				t.Errorf("mutate: %d, fitness: %d, copy: %d\n", res.cost_of_mutate, res.cost_of_measure_fitness, res.cost_of_copy)
				t.Errorf("(mutate+fitness)/copy: %d\n", (res.cost_of_mutate+res.cost_of_measure_fitness)/res.cost_of_copy)
			}
		})
		t.Run("expensive", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				n_bases:       NewOption(uint(5)),
				n_genes:       NewOption(uint(4)),
				n_alleles:     NewOption(uint(3)),
				n_chromosomes: NewOption(uint(2)),
				base_factory:  NewOption(base_factory),
			}
			initial_population := []*Allele[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeAllele(opts)
				initial_population = append(initial_population, gene)
			}
			params := OptimizationParams[int, Allele[int]]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(MeasureAlleleFitnessExpensive),
				mutate:             NewOption(MutateAlleleExpensive),
				max_iterations:     NewOption(1000),
				recombination_opts: NewOption(RecombineOptions{
					recombine_genes:       NewOption(true),
					match_genes:           NewOption(false),
					recombine_alleles:     NewOption(true),
					match_alleles:         NewOption(false),
					recombine_chromosomes: NewOption(true),
					match_chromosomes:     NewOption(false),
				}),
			}

			n_goroutines, err := TuneAlleleOptimization(params)
			if err != nil {
				t.Fatalf("TuneAlleleOptimization failed with error: %v", err)
			}

			if n_goroutines < 1 {
				t.Errorf("TuneAlleleOptimization failed: expected 1, observed %d", n_goroutines)
				res := benchmarkAlleleOptimization(params)
				t.Errorf("mutate: %d, fitness: %d, copy: %d\n", res.cost_of_mutate, res.cost_of_measure_fitness, res.cost_of_copy)
				t.Errorf("(mutate+fitness)/copy: %d\n", (res.cost_of_mutate+res.cost_of_measure_fitness)/res.cost_of_copy)
			}
		})
	})
	t.Run("Chromosome", func(t *testing.T) {
		t.Run("cheap", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				n_bases:       NewOption(uint(5)),
				n_genes:       NewOption(uint(4)),
				n_alleles:     NewOption(uint(3)),
				n_chromosomes: NewOption(uint(2)),
				base_factory:  NewOption(base_factory),
			}
			initial_population := []*Chromosome[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeChromosome(opts)
				initial_population = append(initial_population, gene)
			}
			params := OptimizationParams[int, Chromosome[int]]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(MeasureChromosomeFitness),
				mutate:             NewOption(MutateChromosome),
				max_iterations:     NewOption(1000),
				recombination_opts: NewOption(RecombineOptions{
					recombine_genes:       NewOption(true),
					match_genes:           NewOption(false),
					recombine_alleles:     NewOption(true),
					match_alleles:         NewOption(false),
					recombine_chromosomes: NewOption(true),
					match_chromosomes:     NewOption(false),
				}),
			}

			n_goroutines, err := TuneChromosomeOptimization(params)
			if err != nil {
				t.Fatalf("TuneChromosomeOptimization failed with error: %v", err)
			}

			if n_goroutines > 1 {
				t.Errorf("TuneChromosomeOptimization failed: expected 1, observed %d", n_goroutines)
				res := benchmarkChromosomeOptimization(params)
				t.Errorf("mutate: %d, fitness: %d, copy: %d\n", res.cost_of_mutate, res.cost_of_measure_fitness, res.cost_of_copy)
				t.Errorf("(mutate+fitness)/copy: %d\n", (res.cost_of_mutate+res.cost_of_measure_fitness)/res.cost_of_copy)
			}
		})
		t.Run("expensive", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				n_bases:       NewOption(uint(5)),
				n_genes:       NewOption(uint(4)),
				n_alleles:     NewOption(uint(3)),
				n_chromosomes: NewOption(uint(2)),
				base_factory:  NewOption(base_factory),
			}
			initial_population := []*Chromosome[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeChromosome(opts)
				initial_population = append(initial_population, gene)
			}
			params := OptimizationParams[int, Chromosome[int]]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(MeasureChromosomeFitnessExpensive),
				mutate:             NewOption(MutateChromosomeExpensive),
				max_iterations:     NewOption(1000),
				recombination_opts: NewOption(RecombineOptions{
					recombine_genes:       NewOption(true),
					match_genes:           NewOption(false),
					recombine_alleles:     NewOption(true),
					match_alleles:         NewOption(false),
					recombine_chromosomes: NewOption(true),
					match_chromosomes:     NewOption(false),
				}),
			}

			n_goroutines, err := TuneChromosomeOptimization(params)
			if err != nil {
				t.Fatalf("TuneChromosomeOptimization failed with error: %v", err)
			}

			if n_goroutines < 1 {
				t.Errorf("TuneChromosomeOptimization failed: expected 1, observed %d", n_goroutines)
				res := benchmarkChromosomeOptimization(params)
				t.Errorf("mutate: %d, fitness: %d, copy: %d\n", res.cost_of_mutate, res.cost_of_measure_fitness, res.cost_of_copy)
				t.Errorf("(mutate+fitness)/copy: %d\n", (res.cost_of_mutate+res.cost_of_measure_fitness)/res.cost_of_copy)
			}
		})
	})

	t.Run("Genome", func(t *testing.T) {
		t.Run("cheap", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				n_bases:       NewOption(uint(5)),
				n_genes:       NewOption(uint(4)),
				n_alleles:     NewOption(uint(3)),
				n_chromosomes: NewOption(uint(2)),
				base_factory:  NewOption(base_factory),
			}
			initial_population := []*Genome[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeGenome(opts)
				initial_population = append(initial_population, gene)
			}
			params := OptimizationParams[int, Genome[int]]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(MeasureGenomeFitness),
				mutate:             NewOption(MutateGenome),
				max_iterations:     NewOption(1000),
				recombination_opts: NewOption(RecombineOptions{
					recombine_genes:       NewOption(true),
					match_genes:           NewOption(false),
					recombine_alleles:     NewOption(true),
					match_alleles:         NewOption(false),
					recombine_chromosomes: NewOption(true),
					match_chromosomes:     NewOption(false),
				}),
			}

			n_goroutines, err := TuneGenomeOptimization(params)
			if err != nil {
				t.Fatalf("TuneGenomeOptimization failed with error: %v", err)
			}

			if n_goroutines > 1 {
				t.Errorf("TuneGenomeOptimization failed: expected 1, observed %d", n_goroutines)
				res := benchmarkGenomeOptimization(params)
				t.Errorf("mutate: %d, fitness: %d, copy: %d\n", res.cost_of_mutate, res.cost_of_measure_fitness, res.cost_of_copy)
				t.Errorf("(mutate+fitness)/copy: %d\n", (res.cost_of_mutate+res.cost_of_measure_fitness)/res.cost_of_copy)
			}
		})
		t.Run("expensive", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				n_bases:       NewOption(uint(5)),
				n_genes:       NewOption(uint(4)),
				n_alleles:     NewOption(uint(3)),
				n_chromosomes: NewOption(uint(2)),
				base_factory:  NewOption(base_factory),
			}
			initial_population := []*Genome[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeGenome(opts)
				initial_population = append(initial_population, gene)
			}
			params := OptimizationParams[int, Genome[int]]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(MeasureGenomeFitnessExpensive),
				mutate:             NewOption(MutateGenomeExpensive),
				max_iterations:     NewOption(1000),
				recombination_opts: NewOption(RecombineOptions{
					recombine_genes:       NewOption(true),
					match_genes:           NewOption(false),
					recombine_alleles:     NewOption(true),
					match_alleles:         NewOption(false),
					recombine_chromosomes: NewOption(true),
					match_chromosomes:     NewOption(false),
				}),
			}

			n_goroutines, err := TuneGenomeOptimization(params)
			if err != nil {
				t.Fatalf("TuneGenomeOptimization failed with error: %v", err)
			}

			if n_goroutines < 1 {
				t.Errorf("TuneGenomeOptimization failed: expected 1, observed %d", n_goroutines)
				res := benchmarkGenomeOptimization(params)
				t.Errorf("mutate: %d, fitness: %d, copy: %d\n", res.cost_of_mutate, res.cost_of_measure_fitness, res.cost_of_copy)
				t.Errorf("(mutate+fitness)/copy: %d\n", (res.cost_of_mutate+res.cost_of_measure_fitness)/res.cost_of_copy)
			}
		})
	})
}
