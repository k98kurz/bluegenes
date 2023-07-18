# Command

```bash
go test -bench=. -run=^# -v
```

# master (before experimental optimizations)

- BenchmarkOptimize/GeneSequential
- BenchmarkOptimize/GeneSequential-8                   555           2145373 ns/op
- BenchmarkOptimize/GeneParallel
- BenchmarkOptimize/GeneParallel-8                     194           5579576 ns/op
- BenchmarkOptimize/AlleleSequential
- BenchmarkOptimize/AlleleSequential-8                 160           7218187 ns/op
- BenchmarkOptimize/AlleleParallel
- BenchmarkOptimize/AlleleParallel-8                    87          13199322 ns/op
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
- BenchmarkOptimize/AlleleSequential
- BenchmarkOptimize/AlleleSequential-8                 153           6650617 ns/op
- BenchmarkOptimize/AlleleParallel
- BenchmarkOptimize/AlleleParallel-8                    85          14451632 ns/op
- BenchmarkOptimize/ChromosomeSequential
- BenchmarkOptimize/ChromosomeSequential-8              54          23091213 ns/op
- BenchmarkOptimize/ChromosomeParallel
- BenchmarkOptimize/ChromosomeParallel-8                27          53027202 ns/op
- BenchmarkOptimize/GenomeSequential
- BenchmarkOptimize/GenomeSequential-8                  13          79100384 ns/op
- BenchmarkOptimize/GenomeParallel
- BenchmarkOptimize/GenomeParallel-8                     6         193227775 ns/op

## part 2

- BenchmarkOptimize/GeneSequential
- BenchmarkOptimize/GeneSequential-8                 52992             20140 ns/op
- BenchmarkOptimize/GeneParallel
- BenchmarkOptimize/GeneParallel-8                   58455             19621 ns/op
- BenchmarkOptimize/AlleleSequential
- BenchmarkOptimize/AlleleSequential-8               53139             18965 ns/op
- BenchmarkOptimize/AlleleParallel
- BenchmarkOptimize/AlleleParallel-8                 55908             19905 ns/op
- BenchmarkOptimize/ChromosomeSequential
- BenchmarkOptimize/ChromosomeSequential-8           41264             24824 ns/op
- BenchmarkOptimize/ChromosomeParallel
- BenchmarkOptimize/ChromosomeParallel-8             47565             23419 ns/op
- BenchmarkOptimize/GenomeSequential
- BenchmarkOptimize/GenomeSequential-8               26118             43253 ns/op
- BenchmarkOptimize/GenomeParallel
- BenchmarkOptimize/GenomeParallel-8                 26278             43637 ns/op

Average: 26720.5 ns/op
