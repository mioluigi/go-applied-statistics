# bootstrap-se-boot.R: the same analysis as bootstrap-se.R with the CRAN package "boot"
# (Davison and Hinkley 1997; https://cran.r-project.org/package=boot).
#
# Usage: Rscript r/bootstrap-se-boot.R <data.csv> [replicates = 1000] [seed = 431]
#
# boot() draws all resample indices in one call and deals them to replicates in its own
# order, so for the same seed its resamples differ from a sample() loop's. Its standard
# errors agree with bootstrap-se.R statistically, not digit for digit; the output's "stream"
# field says "boot" to mark that.

arguments <- commandArgs(trailingOnly = TRUE)
fail <- function(message) {
    cat("bootstrap-se-boot.R: ", message, "\n", sep = "", file = stderr())
    quit(status = 2)
}
if (length(arguments) < 1) fail("usage: Rscript r/bootstrap-se-boot.R <data.csv> [replicates] [seed]")
dataFile <- arguments[1]
replicates <- if (length(arguments) >= 2) suppressWarnings(as.integer(arguments[2])) else 1000L
seed <- if (length(arguments) >= 3) suppressWarnings(as.integer(arguments[3])) else 431L
if (!file.exists(dataFile)) fail(paste("cannot read", dataFile))
if (is.na(replicates) || replicates < 2) fail("replicates must be a whole number of at least 2")
if (is.na(seed)) fail("seed must be a whole number")

loadStart <- proc.time()[["elapsed"]]
suppressPackageStartupMessages(library(boot))
loadMs <- (proc.time()[["elapsed"]] - loadStart) * 1000

readStart <- proc.time()[["elapsed"]]
x <- scan(dataFile, what = numeric(), skip = 1, quiet = TRUE)
n <- length(x)
readMs <- (proc.time()[["elapsed"]] - readStart) * 1000
if (n < 2) fail("the data set needs at least 2 observations")

# The statistic receives the data and the indices of one resample; it returns both values.
medianAndMean <- function(data, indices) {
    resample <- data[indices]
    c(median(resample), mean(resample))
}

computeStart <- proc.time()[["elapsed"]]
set.seed(seed)
result <- boot(x, medianAndMean, R = replicates)
standardErrorMedian <- sd(result$t[, 1])
standardErrorMean <- sd(result$t[, 2])
computeMs <- (proc.time()[["elapsed"]] - computeStart) * 1000

exact <- function(value) sprintf("%.17g", value)
cat(paste(basename(dataFile), n, replicates, seed, "boot",
          exact(median(x)), exact(standardErrorMedian),
          exact(mean(x)), exact(standardErrorMean), sep = "\t"), "\n", sep = "")
cat(sprintf("load_ms=%.3f read_ms=%.3f compute_ms=%.3f\n", loadMs, readMs, computeMs), file = stderr())
