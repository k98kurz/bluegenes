package bluegenes

import (
	"math"
	"math/rand"
	"testing"
)

var target = 12345

func mutateGene(gene *Gene[int]) {
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

func mutateAllele(allele *Allele[int]) {
	allele.mu.Lock()
	defer allele.mu.Unlock()
	for _, gene := range allele.genes {
		mutateGene(gene)
	}
}

func mutateChromosome(chromosome *Chromosome[int]) {
	chromosome.mu.Lock()
	defer chromosome.mu.Unlock()
	for _, allele := range chromosome.alleles {
		mutateAllele(allele)
	}
}

func mutateGenome(genome *Genome[int]) {
	genome.mu.Lock()
	defer genome.mu.Unlock()
	for _, chromosome := range genome.chromosomes {
		mutateChromosome(chromosome)
	}
}

func mutateGeneticMaterial(code GeneticMaterial[int]) {
	if code.gene.ok() {
		mutateGene(code.gene.val)
	}
	if code.allele.ok() {
		mutateAllele(code.allele.val)
	}
	if code.chromosome.ok() {
		mutateChromosome(code.chromosome.val)
	}
	if code.genome.ok() {
		mutateGenome(code.genome.val)
	}
}

func measureGeneFitness(gene *Gene[int]) float64 {
	gene.mu.RLock()
	defer gene.mu.RUnlock()
	total := reduce(gene.bases, func(a int, b int) int { return a + b })
	return 1.0 / (math.Abs(float64(total-target)) + 1.0)
}

func measureAlleleFitness(allele *Allele[int]) float64 {
	allele.mu.RLock()
	defer allele.mu.RUnlock()
	total := 0
	for _, gene := range allele.genes {
		total += reduce(gene.bases, func(a int, b int) int { return a + b })
	}
	return 1.0 / (math.Abs(float64(total-target)) + 1.0)
}

func measureChromosomeFitness(chromosome *Chromosome[int]) float64 {
	chromosome.mu.RLock()
	defer chromosome.mu.RUnlock()
	total := 0
	for _, allele := range chromosome.alleles {
		for _, gene := range allele.genes {
			total += reduce(gene.bases, func(a int, b int) int { return a + b })
		}
	}
	return 1.0 / (math.Abs(float64(total-target)) + 1.0)
}

func measureGenomeFitness(genome *Genome[int]) float64 {
	genome.mu.RLock()
	defer genome.mu.RUnlock()
	total := 0
	for _, chromosome := range genome.chromosomes {
		for _, allele := range chromosome.alleles {
			for _, gene := range allele.genes {
				total += reduce(gene.bases, func(a int, b int) int { return a + b })
			}
		}
	}
	return 1.0 / (math.Abs(float64(total-target)) + 1.0)
}

func measureGeneticMaterialFitness(code GeneticMaterial[int]) float64 {
	fitness := 0.0
	fitness_count := 0
	if code.gene.ok() {
		fitness += measureGeneFitness(code.gene.val)
		fitness_count++
	}
	if code.allele.ok() {
		fitness += measureAlleleFitness(code.allele.val)
		fitness_count++
	}
	if code.chromosome.ok() {
		fitness += measureChromosomeFitness(code.chromosome.val)
		fitness_count++
	}
	if code.genome.ok() {
		fitness += measureGenomeFitness(code.genome.val)
		fitness_count++
	}

	return fitness / float64(fitness_count)
}

func mutateGeneExpensive(gene *Gene[int]) {
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

func mutateAlleleExpensive(allele *Allele[int]) {
	allele.mu.Lock()
	defer allele.mu.Unlock()
	for _, gene := range allele.genes {
		mutateGeneExpensive(gene)
	}
}

func mutateChromosomeExpensive(chromosome *Chromosome[int]) {
	chromosome.mu.Lock()
	defer chromosome.mu.Unlock()
	for _, allele := range chromosome.alleles {
		mutateAlleleExpensive(allele)
	}
}

func mutateGenomeExpensive(genome *Genome[int]) {
	genome.mu.Lock()
	defer genome.mu.Unlock()
	for _, chromosome := range genome.chromosomes {
		mutateChromosomeExpensive(chromosome)
	}
}

func measureGeneFitnessExpensive(gene *Gene[int]) float64 {
	gene.mu.RLock()
	defer gene.mu.RUnlock()
	val := 42.0
	for i := 0; i < 1000; i++ {
		val /= 6.9
	}
	total := reduce(gene.bases, func(a int, b int) int { return a + b })
	return 1.0 / (math.Abs(float64(total-target)) + 1.0)
}

func measureAlleleFitnessExpensive(allele *Allele[int]) float64 {
	allele.mu.RLock()
	defer allele.mu.RUnlock()
	val := 42.0
	for i := 0; i < 1000; i++ {
		val /= 6.9
	}
	total := 0
	for _, gene := range allele.genes {
		total += reduce(gene.bases, func(a int, b int) int { return a + b })
	}
	return 1.0 / (math.Abs(float64(total-target)) + 1.0)
}

func measureChromosomeFitnessExpensive(chromosome *Chromosome[int]) float64 {
	chromosome.mu.RLock()
	defer chromosome.mu.RUnlock()
	val := 42.0
	for i := 0; i < 1000; i++ {
		val /= 6.9
	}
	total := 0
	for _, allele := range chromosome.alleles {
		for _, gene := range allele.genes {
			total += reduce(gene.bases, func(a int, b int) int { return a + b })
		}
	}
	return 1.0 / (math.Abs(float64(total-target)) + 1.0)
}

func measureGenomeFitnessExpensive(genome *Genome[int]) float64 {
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
				total += reduce(gene.bases, func(a int, b int) int { return a + b })
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
				n_bases:       NewOption(uint(5)),
				n_genes:       NewOption(uint(4)),
				n_alleles:     NewOption(uint(3)),
				n_chromosomes: NewOption(uint(2)),
				base_factory:  NewOption(base_factory),
			}
			initial_population := []GeneticMaterial[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeGene(opts)
				initial_population = append(initial_population, GeneticMaterial[int]{gene: NewOption(gene)})
			}

			n_iterations, final_population, err := OptimizeGeneticMaterial(GMOptimizationParams[int]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(measureGeneticMaterialFitness),
				mutate:             NewOption(mutateGeneticMaterial),
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
				t.Fatalf("OptimizeGeneticMaterial for Gene[int] failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("OptimizeGeneticMaterial for Gene[int] exceeded max_iterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.score < 0.9 {
				t.Errorf("OptimizeGeneticMaterial for Gene[int] failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.score)
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
			initial_population := []GeneticMaterial[int]{}
			for i := 0; i < 10; i++ {
				gene, _ := MakeGene(opts)
				initial_population = append(initial_population, GeneticMaterial[int]{gene: NewOption(gene)})
			}

			n_iterations, final_population, err := OptimizeGeneticMaterial(GMOptimizationParams[int]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(measureGeneticMaterialFitness),
				mutate:             NewOption(mutateGeneticMaterial),
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
				t.Fatalf("OptimizeGeneticMaterial for Gene[int] failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("OptimizeGeneticMaterial for Gene[int] exceeded max_iterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.score < 0.9 {
				t.Errorf("OptimizeGeneticMaterial for Gene[int] failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.score)
			}
		})
	})

	t.Run("Allele", func(t *testing.T) {
		t.Run("parallel", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				n_bases:       NewOption(uint(5)),
				n_genes:       NewOption(uint(4)),
				n_alleles:     NewOption(uint(3)),
				n_chromosomes: NewOption(uint(2)),
				base_factory:  NewOption(base_factory),
			}
			initial_population := []GeneticMaterial[int]{}
			for i := 0; i < 10; i++ {
				allele, _ := MakeAllele(opts)
				initial_population = append(initial_population, GeneticMaterial[int]{allele: NewOption(allele)})
			}

			n_iterations, final_population, err := OptimizeGeneticMaterial(GMOptimizationParams[int]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(measureGeneticMaterialFitness),
				mutate:             NewOption(mutateGeneticMaterial),
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
				t.Fatalf("OptimizeGeneticMaterial for Allele[int] failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("OptimizeGeneticMaterial for Allele[int] exceeded max_iterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.score < 0.9 {
				t.Errorf("OptimizeGeneticMaterial for Allele[int] failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.score)
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
			initial_population := []GeneticMaterial[int]{}
			for i := 0; i < 10; i++ {
				allele, _ := MakeAllele(opts)
				initial_population = append(initial_population, GeneticMaterial[int]{allele: NewOption(allele)})
			}

			n_iterations, final_population, err := OptimizeGeneticMaterial(GMOptimizationParams[int]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(measureGeneticMaterialFitness),
				mutate:             NewOption(mutateGeneticMaterial),
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
				t.Fatalf("OptimizeGeneticMaterial for Allele[int] failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("OptimizeGeneticMaterial for Allele[int] exceeded max_iterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.score < 0.9 {
				t.Errorf("OptimizeGeneticMaterial for Allele[int] failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.score)
			}
		})
	})

	t.Run("Chromosome", func(t *testing.T) {
		t.Run("parallel", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				n_bases:       NewOption(uint(5)),
				n_genes:       NewOption(uint(4)),
				n_alleles:     NewOption(uint(3)),
				n_chromosomes: NewOption(uint(2)),
				base_factory:  NewOption(base_factory),
			}
			initial_population := []GeneticMaterial[int]{}
			for i := 0; i < 10; i++ {
				chromosome, _ := MakeChromosome(opts)
				initial_population = append(initial_population, GeneticMaterial[int]{chromosome: NewOption(chromosome)})
			}

			n_iterations, final_population, err := OptimizeGeneticMaterial(GMOptimizationParams[int]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(measureGeneticMaterialFitness),
				mutate:             NewOption(mutateGeneticMaterial),
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
				t.Fatalf("OptimizeGeneticMaterial for Chromosome[int] failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("OptimizeGeneticMaterial for Chromosome[int] exceeded max_iterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.score < 0.9 {
				t.Errorf("OptimizeGeneticMaterial for Chromosome[int] failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.score)
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
			initial_population := []GeneticMaterial[int]{}
			for i := 0; i < 10; i++ {
				chromosome, _ := MakeChromosome(opts)
				initial_population = append(initial_population, GeneticMaterial[int]{chromosome: NewOption(chromosome)})
			}

			n_iterations, final_population, err := OptimizeGeneticMaterial(GMOptimizationParams[int]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(measureGeneticMaterialFitness),
				mutate:             NewOption(mutateGeneticMaterial),
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
				t.Fatalf("OptimizeGeneticMaterial for Chromosome[int] failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("OptimizeGeneticMaterial for Chromosome[int] exceeded max_iterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.score < 0.9 {
				t.Errorf("OptimizeGeneticMaterial for Chromosome[int] failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.score)
			}
		})
	})

	t.Run("Allele", func(t *testing.T) {
		t.Run("parallel", func(t *testing.T) {
			base_factory := func() int { return RandomInt(-10, 10) }
			opts := MakeOptions[int]{
				n_bases:       NewOption(uint(5)),
				n_genes:       NewOption(uint(4)),
				n_alleles:     NewOption(uint(3)),
				n_chromosomes: NewOption(uint(2)),
				base_factory:  NewOption(base_factory),
			}
			initial_population := []GeneticMaterial[int]{}
			for i := 0; i < 10; i++ {
				genome, _ := MakeGenome(opts)
				initial_population = append(initial_population, GeneticMaterial[int]{genome: NewOption(genome)})
			}

			n_iterations, final_population, err := OptimizeGeneticMaterial(GMOptimizationParams[int]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(measureGeneticMaterialFitness),
				mutate:             NewOption(mutateGeneticMaterial),
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
				t.Fatalf("OptimizeGeneticMaterial for Genome[int] failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("OptimizeGeneticMaterial for Genome[int] exceeded max_iterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.score < 0.9 {
				t.Errorf("OptimizeGeneticMaterial for Genome[int] failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.score)
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
			initial_population := []GeneticMaterial[int]{}
			for i := 0; i < 10; i++ {
				genome, _ := MakeGenome(opts)
				initial_population = append(initial_population, GeneticMaterial[int]{genome: NewOption(genome)})
			}

			n_iterations, final_population, err := OptimizeGeneticMaterial(GMOptimizationParams[int]{
				initial_population: NewOption(initial_population),
				measure_fitness:    NewOption(measureGeneticMaterialFitness),
				mutate:             NewOption(mutateGeneticMaterial),
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
				t.Fatalf("OptimizeGeneticMaterial for Genome[int] failed with error: %v", err)
			}

			if n_iterations > 1000 {
				t.Errorf("OptimizeGeneticMaterial for Genome[int] exceeded max_iterations with %d iterations", n_iterations)
			}

			best_fitness := final_population[0]

			if n_iterations < 1000 && best_fitness.score < 0.9 {
				t.Errorf("OptimizeGeneticMaterial for Genome[int] failed to meet fitness threshold of 0.9: %f reached instead", best_fitness.score)
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
				measure_fitness:    NewOption(measureGeneFitness),
				mutate:             NewOption(mutateGene),
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
				measure_fitness:    NewOption(measureGeneFitnessExpensive),
				mutate:             NewOption(mutateGeneExpensive),
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
				measure_fitness:    NewOption(measureAlleleFitness),
				mutate:             NewOption(mutateAllele),
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
				measure_fitness:    NewOption(measureAlleleFitnessExpensive),
				mutate:             NewOption(mutateAlleleExpensive),
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
				measure_fitness:    NewOption(measureChromosomeFitness),
				mutate:             NewOption(mutateChromosome),
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
				measure_fitness:    NewOption(measureChromosomeFitnessExpensive),
				mutate:             NewOption(mutateChromosomeExpensive),
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
				measure_fitness:    NewOption(measureGenomeFitness),
				mutate:             NewOption(mutateGenome),
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
				measure_fitness:    NewOption(measureGenomeFitnessExpensive),
				mutate:             NewOption(mutateGenomeExpensive),
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
