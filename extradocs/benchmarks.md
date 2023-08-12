# Command

```bash
go test -bench=. -run=^# -v
```

Some amount of random jitter expected in benchmarks, so they are run multiple
times to ensure accuracy. Representative samples are logged below.

# master (before experimental optimizations)

- BenchmarkOptimize/GeneSequential
- BenchmarkOptimize/GeneSequential-8                   555           2145373 ns/op
- BenchmarkOptimize/GeneParallel
- BenchmarkOptimize/GeneParallel-8                     194           5579576 ns/op
- BenchmarkOptimize/NucleosomeSequential
- BenchmarkOptimize/NucleosomeSequential-8                 160           7218187 ns/op
- BenchmarkOptimize/NucleosomeParallel
- BenchmarkOptimize/NucleosomeParallel-8                    87          13199322 ns/op
- BenchmarkOptimize/ChromosomeSequential
- BenchmarkOptimize/ChromosomeSequential-8              50          22642428 ns/op
- BenchmarkOptimize/ChromosomeParallel
- BenchmarkOptimize/ChromosomeParallel-8                32          42698964 ns/op
- BenchmarkOptimize/GenomeSequential
- BenchmarkOptimize/GenomeSequential-8                  18          80740914 ns/op
- BenchmarkOptimize/GenomeParallel
- BenchmarkOptimize/GenomeParallel-8                     9         193374575 ns/op

Average: 45949917.375 ns/op

# 2023-07-17-experimental-refactor

## part 1

- BenchmarkOptimize/GeneSequential
- BenchmarkOptimize/GeneSequential-8                   535           2133393 ns/op
- BenchmarkOptimize/GeneParallel
- BenchmarkOptimize/GeneParallel-8                     217           5187530 ns/op
- BenchmarkOptimize/NucleosomeSequential
- BenchmarkOptimize/NucleosomeSequential-8                 153           6650617 ns/op
- BenchmarkOptimize/NucleosomeParallel
- BenchmarkOptimize/NucleosomeParallel-8                    85          14451632 ns/op
- BenchmarkOptimize/ChromosomeSequential
- BenchmarkOptimize/ChromosomeSequential-8              54          23091213 ns/op
- BenchmarkOptimize/ChromosomeParallel
- BenchmarkOptimize/ChromosomeParallel-8                27          53027202 ns/op
- BenchmarkOptimize/GenomeSequential
- BenchmarkOptimize/GenomeSequential-8                  13          79100384 ns/op
- BenchmarkOptimize/GenomeParallel
- BenchmarkOptimize/GenomeParallel-8                     6         193227775 ns/op

Average: 47108718.25 ns/op

## part 2

- BenchmarkOptimize/GeneSequential
- BenchmarkOptimize/GeneSequential-8                 52992             20140 ns/op
- BenchmarkOptimize/GeneParallel
- BenchmarkOptimize/GeneParallel-8                   58455             19621 ns/op
- BenchmarkOptimize/NucleosomeSequential
- BenchmarkOptimize/NucleosomeSequential-8               53139             18965 ns/op
- BenchmarkOptimize/NucleosomeParallel
- BenchmarkOptimize/NucleosomeParallel-8                 55908             19905 ns/op
- BenchmarkOptimize/ChromosomeSequential
- BenchmarkOptimize/ChromosomeSequential-8           41264             24824 ns/op
- BenchmarkOptimize/ChromosomeParallel
- BenchmarkOptimize/ChromosomeParallel-8             47565             23419 ns/op
- BenchmarkOptimize/GenomeSequential
- BenchmarkOptimize/GenomeSequential-8               26118             43253 ns/op
- BenchmarkOptimize/GenomeParallel
- BenchmarkOptimize/GenomeParallel-8                 26278             43637 ns/op

Average: 26720.5 ns/op

## made MeasureFitness into func(*Code[T]) float64

- BenchmarkOptimize/GeneSequential
- BenchmarkOptimize/GeneSequential-8                 52788             19049 ns/op
- BenchmarkOptimize/GeneParallel
- BenchmarkOptimize/GeneParallel-8                   57282             19306 ns/op
- BenchmarkOptimize/NucleosomeSequential
- BenchmarkOptimize/NucleosomeSequential-8               59226             19700 ns/op
- BenchmarkOptimize/NucleosomeParallel
- BenchmarkOptimize/NucleosomeParallel-8                 48254             21410 ns/op
- BenchmarkOptimize/ChromosomeSequential
- BenchmarkOptimize/ChromosomeSequential-8           50594             25481 ns/op
- BenchmarkOptimize/ChromosomeParallel
- BenchmarkOptimize/ChromosomeParallel-8             42196             26197 ns/op
- BenchmarkOptimize/GenomeSequential
- BenchmarkOptimize/GenomeSequential-8               28096             42931 ns/op
- BenchmarkOptimize/GenomeParallel
- BenchmarkOptimize/GenomeParallel-8                 25474             42915 ns/op

Average: 27123.625 ns/op

Slightly worse. Undone.

## made Mutate into func(Code[T])

- BenchmarkOptimize/GeneSequential
- BenchmarkOptimize/GeneSequential-8                 64552             18633 ns/op
- BenchmarkOptimize/GeneParallel
- BenchmarkOptimize/GeneParallel-8                   52251             19616 ns/op
- BenchmarkOptimize/NucleosomeSequential
- BenchmarkOptimize/NucleosomeSequential-8               49417             20947 ns/op
- BenchmarkOptimize/NucleosomeParallel
- BenchmarkOptimize/NucleosomeParallel-8                 49714             21095 ns/op
- BenchmarkOptimize/ChromosomeSequential
- BenchmarkOptimize/ChromosomeSequential-8           49522             24352 ns/op
- BenchmarkOptimize/ChromosomeParallel
- BenchmarkOptimize/ChromosomeParallel-8             43490             26826 ns/op
- BenchmarkOptimize/GenomeSequential
- BenchmarkOptimize/GenomeSequential-8               26547             43257 ns/op
- BenchmarkOptimize/GenomeParallel
- BenchmarkOptimize/GenomeParallel-8                 25240             45040 ns/op

Average: 27470.75 ns/op

Also slightly worse. Undone.
