# Generated tables

Written by `go run ./cmd/tables` from `results/` and `prices/`. Do not edit.

## headline

- **Same results.** 264 of 264 comparison checks passed; 12,000 of 12,000 resample medians are identical between R and Go.
- **Computation only, one core** (n = 100,000, B = 1,000): R 4239.6 ms, Go 2376.0 ms: Go is **1.8×** faster.
- **Whole process, small job** (n = 100): R 0.168 s, of which 76.2 % is R starting up; Go about 0.005 s, which is below the timer's 0.01 s step, so the ratio is approximate: **16.0×** from medians, **30.8×** from means.
- **Whole process, large job** (n = 100,000): R 4.419 s, Go 2.393 s (**1.8×**).
- **Parallel Go** (8 workers, n = 100,000): computation **7.4×** faster than one worker.
- **Memory** (n = 100,000, peak footprint): R 110.9 MB, of which 46.6 MB is R just starting; Go 8.5 MB.
- **Cloud cost** (`c7g.medium`, one job at a time): Go saves **45.8 % to 96.8 %** of the compute cost across the four sizes.
- **In dollars**: a million jobs cost $1.69 (n = 100) to $44.56 (n = 100,000) in R; Go saves $1.64 to $20.43 per million, so each dollar spent on the rewrite is recovered after about 48,958 to 610,863 jobs.

## mersenne

| Generator | Algorithm | First three uniform numbers for seed 431 |
|---|---|---|
| R 4.4.2: `set.seed(431); runif(3)` | 32-bit Mersenne Twister, R's seeding | 0.75096705253235996, 0.58476094505749643, 0.80907909828238189 |
| This project: `rrand.New(431)` | the same, written from R's source | 0.75096705253235996, 0.58476094505749643, 0.80907909828238189 |
| `github.com/seehuhn/mt19937`, `Seed(431)` | 64-bit Mersenne Twister, its own seeding | 0.23481082912892687, 0.0021374877061913254, 0.8711851171715842 |

## example

Data set `data/lognormal-1000.csv`, B = 1,000 resamples, seed 431; every digit as the program printed it.

| Implementation | Sample median | Bootstrap SE of the median | Sample mean | Bootstrap SE of the mean |
|---|---|---|---|---|
| R, `sample()` loop (`r/bootstrap-se.R`) | 86.648300000000006 | 1.503050518008813 | 97.8793036 | 1.7145995070818436 |
| Go, R stream (`bootse`) | 86.648300000000006 | 1.503050518008813 | 97.8793036 | 1.7145995070818436 |
| R, `boot` package (`r/bootstrap-se-boot.R`) | 86.648300000000006 | 1.5442582547127548 | 97.8793036 | 1.7239710285052126 |
| Go, block stream (`bootse -stream block`) | 86.648300000000006 | 1.5390110210425931 | 97.8793036 | 1.7159476269489826 |

## correctness

| Comparison (12 data sets each) | Checks passed | Largest relative difference |
|---|---|---|
| Go (R stream) against R (sample loop): same draws | 72 of 72 | 2e-16 |
| Go (block stream) against R (sample loop): different draws | 60 of 60 | 0.065 |
| Go (block stream), 10 workers against 1 worker | 60 of 60 | 0 |
| R (boot package) against R (sample loop): different draws | 60 of 60 | 0.065 |
| Go (R stream) against theory | 12 of 12 | 0.037 |

Resample medians compared between R and Go: **12,000, of which 12,000 identical** (same 64-bit value).

## startup

| Start-up condition | Runs | Mean (s) | Median (s) | Slowest run (s) | SD (s) | Std. error of mean | Peak footprint (MB) | Max resident (MB) |
|---|---|---|---|---|---|---|---|---|
| R, empty script | 200 | 0.128 | 0.130 | 0.220 | 0.012 | 0.6 % | 46.6 | 63.4 |
| R, loading `boot` only | 200 | 0.128 | 0.130 | 0.190 | 0.010 | 0.5 % | 47.0 | 63.9 |
| Go, printing its help | 200 | 0.003 | 0.000 | 0.260 | 0.019 | 40.1 % ⚠ | 3.3 | 5.3 |

## scaling

| n | Implementation | Runs | Whole process: mean (s) | Median (s) | SD (s) | Std. error of mean | Read (ms) | Compute (ms) |
|---|---|---|---|---|---|---|---|---|
| 100 | R, `sample()` loop | 200 | 0.168 | 0.160 | 0.015 | 0.6 % | 0.2 | 30.4 |
| 100 | R, `boot` package | 200 | 0.164 | 0.160 | 0.011 | 0.5 % | 0.1 | 29.1 |
| 100 | Go, R stream (1 core) | 200 | 0.005 | 0.010 | 0.006 | 7.8 % ⚠ | 0.3 | 2.2 |
| 100 | Go, block stream (1 worker) | 200 | 0.005 | 0.010 | 0.005 | 6.8 % ⚠ | 0.4 | 1.8 |
| 100 | Go, block stream (8 workers) | 200 | 0.003 | 0.000 | 0.005 | 11.6 % ⚠ | 0.4 | 0.5 |
| 1,000 | R, `sample()` loop | 200 | 0.193 | 0.190 | 0.014 | 0.5 % | 0.3 | 55.0 |
| 1,000 | R, `boot` package | 200 | 0.193 | 0.190 | 0.013 | 0.5 % | 0.2 | 57.1 |
| 1,000 | Go, R stream (1 core) | 200 | 0.021 | 0.020 | 0.003 | 0.9 % | 0.5 | 16.9 |
| 1,000 | Go, block stream (1 worker) | 200 | 0.020 | 0.020 | 0.001 | 0.5 % | 0.5 | 16.7 |
| 1,000 | Go, block stream (8 workers) | 200 | 0.009 | 0.010 | 0.003 | 2.3 % ⚠ | 0.5 | 2.5 |
| 10,000 | R, `sample()` loop | 200 | 0.704 | 0.700 | 0.018 | 0.2 % | 1.2 | 559.1 |
| 10,000 | R, `boot` package | 200 | 0.724 | 0.720 | 0.013 | 0.1 % | 1.1 | 586.4 |
| 10,000 | Go, R stream (1 core) | 200 | 0.256 | 0.260 | 0.005 | 0.1 % | 1.1 | 251.1 |
| 10,000 | Go, block stream (1 worker) | 200 | 0.161 | 0.160 | 0.003 | 0.1 % | 1.1 | 153.8 |
| 10,000 | Go, block stream (8 workers) | 200 | 0.028 | 0.030 | 0.004 | 1.1 % | 1.0 | 21.1 |
| 100,000 | R, `sample()` loop | 30 | 4.419 | 4.420 | 0.027 | 0.1 % | 9.7 | 4239.6 |
| 100,000 | R, `boot` package | 30 | 5.374 | 5.370 | 0.028 | 0.1 % | 9.7 | 5198.8 |
| 100,000 | Go, R stream (1 core) | 30 | 2.393 | 2.390 | 0.006 | 0.0 % | 6.8 | 2376.0 |
| 100,000 | Go, block stream (1 worker) | 30 | 1.522 | 1.520 | 0.005 | 0.1 % | 6.9 | 1503.0 |
| 100,000 | Go, block stream (8 workers) | 30 | 0.222 | 0.220 | 0.006 | 0.5 % | 6.9 | 202.5 |

## ratios

| n | R ÷ Go (R stream), whole process, means | same, medians | R ÷ Go (R stream), computation only | R ÷ Go (block stream, 1 worker), computation only | R start-up as share of R's time | Go: 8 workers ÷ 1 worker, computation | R ÷ Go with 8 workers, whole process | `boot` ÷ `sample()` loop |
|---|---|---|---|---|---|---|---|---|
| 100 | 30.8× | 16.0× | 13.6× | 16.4× | 76.2 % | 3.5× | 59.9× | 1.0× |
| 1,000 | 9.4× | 9.5× | 3.2× | 3.3× | 66.2 % | 6.6× | 20.6× | 1.0× |
| 10,000 | 2.7× | 2.7× | 2.2× | 3.6× | 18.2 % | 7.3× | 25.3× | 1.0× |
| 100,000 | 1.8× | 1.8× | 1.8× | 2.8× | 2.9 % | 7.4× | 19.9× | 1.2× |

## shapes

| Shape (n = 10,000) | Runs | R whole process (s) | Go whole process (s) | R ÷ Go | R compute (ms) | Go compute (ms) | R ÷ Go, computation |
|---|---|---|---|---|---|---|---|
| normal | 200 | 0.699 | 0.258 | 2.7× | 557.4 | 251.3 | 2.2× |
| lognormal | 200 | 0.704 | 0.256 | 2.7× | 559.1 | 251.1 | 2.2× |
| mirrored-lognormal | 200 | 0.700 | 0.258 | 2.7× | 557.4 | 251.6 | 2.2× |

## course

| Program | Runs | Mean time (s) | Memory (MB: peak footprint; Go: allocated per run) |
|---|---|---|---|
| Course R program, unchanged (whole process) | 30 | 4.725 | 72.3 |
| Go, same resampling volume (benchmark, computation only) | 5 | 0.144 | 1.4 |

Ratio: **32.9×**. Same volume of resampling (40,000 resamples, mean and median), not the same program.

## memory

| n | Implementation | Peak footprint: mean (MB) | SD (MB) | Max resident: mean (MB) | Footprint above the language's start-up (MB) |
|---|---|---|---|---|---|
| 100 | R, `sample()` loop | 64.6 | 3.2 | 82.2 | 18.0 |
| 100 | R, `boot` package | 64.3 | 3.3 | 82.1 | 17.4 |
| 100 | Go, R stream (1 core) | 3.3 | 0.1 | 5.3 | 0.1 |
| 100 | Go, block stream (1 worker) | 3.4 | 0.1 | 5.3 | 0.1 |
| 100 | Go, block stream (8 workers) | 4.1 | 0.1 | 6.0 | 0.8 |
| 1,000 | R, `sample()` loop | 82.8 | 3.6 | 100.0 | 36.2 |
| 1,000 | R, `boot` package | 88.2 | 3.2 | 106.0 | 41.2 |
| 1,000 | Go, R stream (1 core) | 3.4 | 0.1 | 5.4 | 0.2 |
| 1,000 | Go, block stream (1 worker) | 3.4 | 0.1 | 5.4 | 0.2 |
| 1,000 | Go, block stream (8 workers) | 4.3 | 0.1 | 6.2 | 1.0 |
| 10,000 | R, `sample()` loop | 123.3 | 3.9 | 217.4 | 76.7 |
| 10,000 | R, `boot` package | 112.0 | 2.2 | 143.4 | 65.0 |
| 10,000 | Go, R stream (1 core) | 4.1 | 0.1 | 6.1 | 0.8 |
| 10,000 | Go, block stream (1 worker) | 4.1 | 0.1 | 6.1 | 0.8 |
| 10,000 | Go, block stream (8 workers) | 5.4 | 0.1 | 7.4 | 2.1 |
| 100,000 | R, `sample()` loop | 110.9 | 2.6 | 238.6 | 64.4 |
| 100,000 | R, `boot` package | 564.6 | 19.0 | 822.2 | 517.6 |
| 100,000 | Go, R stream (1 core) | 8.5 | 0.5 | 10.5 | 5.2 |
| 100,000 | Go, block stream (1 worker) | 8.9 | 0.5 | 10.9 | 5.6 |
| 100,000 | Go, block stream (8 workers) | 14.0 | 0.5 | 16.0 | 10.7 |

## benchmarks

**Performance steps, before and after** (median of repeated benchmark runs)

| Step | Before | After | Speed-up | Memory allocated per run | Allocations per run |
|---|---|---|---|---|---|
| 1. Median by selection instead of sorting (n = 10,000) | 550.4 µs | 47.1 µs | 11.7× | 81.9 kB → 0 B | 1 → 0 |
| 2. One resample buffer per worker (n = 10,000, B = 100) | 25.53 ms | 25.00 ms | 1.0× | 8.2 MB → 86.6 kB | 104 → 7 |
| 3. Index bit count without a logarithm (n = 100) | 23.9 ns | 13.1 ns | 1.8× | 0 B → 0 B | 0 → 0 |
|    same, n = 100,000 | 30.6 ns | 20.5 ns | 1.5× | 0 B → 0 B | 0 → 0 |
| 4. Generator without remainders or floating point (n = 100) | 13.1 ns | 8.2 ns | 1.6× | 0 B → 0 B | 0 → 0 |
|    same, n = 100,000 | 20.5 ns | 12.8 ns | 1.6× | 0 B → 0 B | 0 → 0 |
| Whole engine, steps 1 and 2 (n = 1,000, B = 100) | 5.15 ms | 1.56 ms | 3.3× | 1.6 MB → 12.9 kB | 204 → 8 |
| Whole engine, steps 1 and 2 (n = 100,000, B = 100) | 880.83 ms | 234.90 ms | 3.7× | 160.6 MB → 807.5 kB | 207 → 8 |
| Parallel workers: 10 instead of 1 (block stream, n = 10,000, B = 1,000) | 152.57 ms | 22.44 ms | 6.8× | 130.5 kB → 869.5 kB | 2007 → 2025 |

**All component benchmarks** (`go test -bench . -benchmem`)

| Benchmark | Repetitions | Time per operation | Bytes per operation | Allocations per operation |
|---|---|---|---|---|
| `Float64` | 5 | 3.1 ns | 0 B | 0 |
| `Step3Index/before=logarithm/n=100` | 5 | 23.9 ns | 0 B | 0 |
| `Step3Index/after=bitlength/n=100` | 5 | 13.1 ns | 0 B | 0 |
| `Step3Index/before=logarithm/n=100000` | 5 | 30.6 ns | 0 B | 0 |
| `Step3Index/after=bitlength/n=100000` | 5 | 20.5 ns | 0 B | 0 |
| `Step4Generator/before=uniforms/n=100` | 5 | 13.1 ns | 0 B | 0 |
| `Step4Generator/after=bits/n=100` | 5 | 8.2 ns | 0 B | 0 |
| `Step4Generator/before=uniforms/n=100000` | 5 | 20.5 ns | 0 B | 0 |
| `Step4Generator/after=bits/n=100000` | 5 | 12.8 ns | 0 B | 0 |
| `Step1Median/before=sort` | 5 | 550.4 µs | 81.9 kB | 1 |
| `Step1Median/after=select` | 5 | 47.1 µs | 0 B | 0 |
| `Step2Resample/before=allocate` | 5 | 25.53 ms | 8.2 MB | 104 |
| `Step2Resample/after=reuse` | 5 | 25.00 ms | 86.6 kB | 7 |
| `Engine/before/n=1000` | 5 | 5.15 ms | 1.6 MB | 204 |
| `Engine/after/n=1000` | 5 | 1.56 ms | 12.9 kB | 8 |
| `Engine/before/n=100000` | 5 | 880.83 ms | 160.6 MB | 207 |
| `Engine/after/n=100000` | 5 | 234.90 ms | 807.5 kB | 8 |
| `Workers/workers=1` | 5 | 152.57 ms | 130.5 kB | 2007 |
| `Workers/workers=10` | 5 | 22.44 ms | 869.5 kB | 2025 |
| `CourseStudyVolume` | 5 | 143.53 ms | 1.4 MB | 2401 |

## step4

Go, R stream, computation time for B = 1,000. Before: official run 1 (commit `202d3cc`). After: the final run (commit `3a19148`).

| n | Runs before | Before step 4 (ms) | Runs after | After step 4 (ms) | Speed-up |
|---|---|---|---|---|---|
| 100 | 200 | 2.9 | 200 | 2.2 | 1.3× |
| 1,000 | 200 | 19.2 | 200 | 16.9 | 1.1× |
| 10,000 | 200 | 351.9 | 200 | 251.1 | 1.4× |
| 100,000 | 30 | 3165.3 | 30 | 2376.0 | 1.3× |

## cost

**Published prices used** (on-demand, Linux)

| Instance | vCPUs | Memory (GiB) | Region | USD per hour | USD per vCPU-hour | Source | Captured |
|---|---|---|---|---|---|---|---|
| AWS EC2 `c7g.medium` | 1 | 2 | us-east-1 (N. Virginia) | $0.0363 | $0.036300 | https://cloudprice.net/aws/ec2 | 2026-11-14 |
| AWS EC2 `c7g.2xlarge` | 8 | 16 | us-east-1 (N. Virginia) | $0.2900 | $0.036250 | https://cloudprice.net/aws/ec2 | 2026-11-14 |

**Scenario A, right-sized: one job at a time on `c7g.medium` (1 vCPU).** Cost of one million jobs (B = 1,000), from mean whole-process times.

| n | R, `sample()` loop | Go, R stream | Saving (from means) | Saving (from medians) | Jobs to save one dollar | Go, block stream, 1 worker | Saving, block stream |
|---|---|---|---|---|---|---|---|
| 100 | $1.69 | $0.0550 | **96.8 %** | 93.8 % | 610,863 | $0.0534 | 96.8 % |
| 1,000 | $1.95 | $0.2077 | **89.3 %** | 89.5 % | 574,586 | $0.2037 | 89.5 % |
| 10,000 | $7.09 | $2.58 | **63.6 %** | 62.9 % | 221,617 | $1.62 | 77.2 % |
| 100,000 | $44.56 | $24.13 | **45.8 %** | 45.9 % | 48,958 | $15.35 | 65.6 % |

**Scenario B, one `c7g.2xlarge` (8 vCPUs) kept busy.** One-core programs run 8 jobs at once; parallel Go runs one job at a time on 8 workers. Savings are against R.

| n | R, 8 jobs at once | Go, R stream, 8 jobs at once | Saving | Go, block stream, 1 worker each, 8 jobs at once | Saving | Go, block stream, 8 workers, 1 job at a time | Saving |
|---|---|---|---|---|---|---|---|
| 100 | $1.69 | $0.0549 | 96.8 % | $0.0534 | 96.8 % | $0.2256 | 86.7 % |
| 1,000 | $1.95 | $0.2074 | 89.3 % | $0.2034 | 89.5 % | $0.7572 | 61.1 % |
| 10,000 | $7.08 | $2.58 | 63.6 % | $1.62 | 77.2 % | $2.24 | 68.4 % |
| 100,000 | $44.50 | $24.10 | 45.8 % | $15.33 | 65.6 % | $17.88 | 59.8 % |

Memory the "8 jobs at once" columns assume, at n = 100,000 (peak footprint × 8): R `sample()` loop 887.5 MB, R `boot` 4516.6 MB, Go one-core 68.0 MB. One Go process with 8 workers: 14.0 MB.

## effort

| Code | Lines of code | Lines of tests | What it supplies |
|---|---|---|---|
| R: `r/bootstrap-se.R` (base R) | 36 | – | Everything else is built into R |
| R: `r/bootstrap-se-boot.R` (`boot` package) | 35 | – | Resampling comes from the package |
| Go: `rrand` | 82 | 269 | R's random generator and `sample()` (built into R) |
| Go: `stats` | 85 | 144 | `mean`, `sd`, `median` (built into R) |
| Go: `bootstrap` | 116 | 400 | The resampling loop, the streams, parallel workers |
| Go: `dataset` | 49 | 84 | Reading and checking the data file (`scan` in R) |
| Go: `record` | 65 | 56 | The shared output line |
| Go: `main.go` (command line) | 233 | 172 | Flags, validation, logging, profiles |
| **Go, total** | **630** | **1125** | **17.5× the base-R script** |

Lines counted: not blank and not only a comment.

## predictions

| # | Prediction (with its conditions) | Measured | Outcome |
|---|---|---|---|
| P4 | Go does the course study's resampling volume in under 0.20 s of computation | 0.144 s | **Held** |
| P5 | Computation only, R ÷ Go (one core) is between 2× and 10× at n = 1,000 and n = 10,000 | 3.2× and 2.2× | **Held** |
| P6 | R's start-up is more than 50 % of its whole-process time at n = 100 and under 20 % at n = 100,000 | 76.2 % and 2.9 % | **Held** |
| P7 | Every different-draws check passes at 15 %, and the largest difference is under 11 % | 48 checks, largest 6.5 % | **Held** |

## machine

- machine: MacBookPro18,2 · Apple M1 Max · 64 GB · macOS 27.0.1
- cpus: 10
- parallel workers: 8
- go: go1.27.2
- r: R version 4.4.2 (2024-10-31)
- r package boot: 1.3.32
- commit: 3a19148
- measured: 2026-11-14 19:54 EST
- runs per condition: 200 (large conditions: 30); B = 1000; seed = 431
- time tool: /usr/bin/time -l
- load average (1 minute) before: 6.25
- load average (1 minute) after: 1.87
