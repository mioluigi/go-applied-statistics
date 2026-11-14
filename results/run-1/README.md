# Official run 1 (kept as evidence)

Measured on the Mac on 2026-11-14 at commit `202d3cc`, before performance step 4.

- Its **CPU profile** (`cpu-profile.txt`) is what chose step 4: the random generator took 59 % of the time
  (`Index`, `refill`, `uint32`, `Float64`), more than the median (30 %).
- Its timings are the "before" of step 4 in the README (`runs.tsv`, Go R stream, `compute_ms`).
- Its machine record shows a 1-minute **load average of 117 at the start**, far above the script's warning
  level, although the timings were stable (standard error of the mean at or below about 1 % for the R
  conditions, no drift between the first and last quarter of the runs, and the course program at 5.08 s
  as in the spike). The final run was made again on a quiet machine; see `results/machine.txt`.

The final results are in the folder above. Do not edit these files.
