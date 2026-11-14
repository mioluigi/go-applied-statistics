# parity-reference.R: reference values from R for spike 001 (R <-> Go comparability).
#
# Usage: Rscript r/spike/parity-reference.R <output folder>
#
# Writes two small files that a Go program must reproduce if it claims to draw the same
# bootstrap samples as R for the same seed:
#   sample-indices.tsv  2000 draws of sample.int(n, replace = TRUE) for several n, seed 431
#   bootstrap-se.tsv    bootstrap standard errors of the median and mean on exact integer data

outputFolder <- commandArgs(trailingOnly = TRUE)[1]
if (is.na(outputFolder)) stop("usage: Rscript parity-reference.R <output folder>")

provenance <- paste0("# ", R.version.string, "; RNGkind: ", paste(RNGkind(), collapse = " / "))

# 1. Raw index draws: the heart of sample(x, n, replace = TRUE).
sizes <- c(10L, 25L, 100L, 225L, 400L, 1000L, 100000L)
indexLines <- character(0)
for (n in sizes) {
    set.seed(431)
    draws <- sample.int(n, 2000L, replace = TRUE)
    indexLines <- c(indexLines, paste(n, paste(draws, collapse = " "), sep = "\t"))
}
writeLines(c(paste0(provenance, "; set.seed(431); sample.int(n, 2000, replace = TRUE)"),
             "n\tdraws", indexLines),
           file.path(outputFolder, "sample-indices.tsv"))

# 2. A whole bootstrap on data that is exact in floating point: x = (37 * i) mod 101.
#    n = 25 is odd, n = 100 is even (median averages the two middle values).
replicates <- 200L
standardErrorLines <- character(0)
for (n in c(25L, 100L)) {
    x <- as.numeric((37L * seq_len(n)) %% 101L)
    set.seed(431)
    medians <- numeric(replicates)
    means <- numeric(replicates)
    for (b in seq_len(replicates)) {
        resample <- sample(x, n, replace = TRUE)
        medians[b] <- median(resample)
        means[b] <- mean(resample)
    }
    standardErrorLines <- c(standardErrorLines,
        paste(n, replicates, 431L,
              format(sd(medians), digits = 17), format(sd(means), digits = 17),
              format(median(x), digits = 17), sep = "\t"))
}
writeLines(c(paste0(provenance, "; x = (37 * i) mod 101; sample(x, n, replace = TRUE)"),
             "n\treplicates\tseed\tse_median\tse_mean\tsample_median", standardErrorLines),
           file.path(outputFolder, "bootstrap-se.tsv"))

cat("parity-reference: wrote sample-indices.tsv and bootstrap-se.tsv to", outputFolder, "\n")
