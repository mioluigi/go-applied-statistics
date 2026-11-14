# bootstrap-se.R: bootstrap standard errors of the median and the mean, in idiomatic base R.
#
# Usage: Rscript r/bootstrap-se.R <data.csv> [replicates = 1000] [seed = 431] [medians file]
#
# Method (Efron and Tibshirani 1993): draw B resamples of the data with replacement, compute
# the statistic on each, and report the standard deviation of the B values.
#
# Output, one tab-separated line on standard output (the same record the Go program writes):
#   data set, n, B, seed, stream, sample median, SE median, sample mean, SE mean
# Timing goes to standard error: read_ms (reading the file) and compute_ms (the B resamples).
# With a fourth argument, the B resample medians are written to that file, one per line.

arguments <- commandArgs(trailingOnly = TRUE)
fail <- function(message) {
    cat("bootstrap-se.R: ", message, "\n", sep = "", file = stderr())
    quit(status = 2)
}
if (length(arguments) < 1) fail("usage: Rscript r/bootstrap-se.R <data.csv> [replicates] [seed] [medians file]")
dataFile <- arguments[1]
replicates <- if (length(arguments) >= 2) suppressWarnings(as.integer(arguments[2])) else 1000L
seed <- if (length(arguments) >= 3) suppressWarnings(as.integer(arguments[3])) else 431L
mediansFile <- if (length(arguments) >= 4) arguments[4] else NA
if (!file.exists(dataFile)) fail(paste("cannot read", dataFile))
if (is.na(replicates) || replicates < 2) fail("replicates must be a whole number of at least 2")
if (is.na(seed)) fail("seed must be a whole number")

readStart <- proc.time()[["elapsed"]]
x <- scan(dataFile, what = numeric(), skip = 1, quiet = TRUE)
n <- length(x)
readMs <- (proc.time()[["elapsed"]] - readStart) * 1000
if (n < 2) fail("the data set needs at least 2 observations")

# One set of B resamples; both statistics are computed on each resample.
computeStart <- proc.time()[["elapsed"]]
set.seed(seed)
medians <- numeric(replicates)
means <- numeric(replicates)
for (b in seq_len(replicates)) {
    resample <- sample(x, n, replace = TRUE)
    means[b] <- mean(resample)
    medians[b] <- median(resample)
}
standardErrorMedian <- sd(medians)
standardErrorMean <- sd(means)
computeMs <- (proc.time()[["elapsed"]] - computeStart) * 1000

exact <- function(value) sprintf("%.17g", value)
cat(paste(basename(dataFile), n, replicates, seed, "r",
          exact(median(x)), exact(standardErrorMedian),
          exact(mean(x)), exact(standardErrorMean), sep = "\t"), "\n", sep = "")
cat(sprintf("read_ms=%.3f compute_ms=%.3f\n", readMs, computeMs), file = stderr())
if (!is.na(mediansFile)) writeLines(exact(medians), mediansFile)
