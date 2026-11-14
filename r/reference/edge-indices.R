# edge-indices.R: R reference draws for sizes the first reference file did not cover.
#
# Usage: Rscript r/reference/edge-indices.R <output file>
#
# R's sample() takes random bits 16 at a time. Sizes needing exactly 16 bits (32,769 to 65,536),
# 17 bits (65,537), and the degenerate sizes 1 and 2 exercise the boundaries of that loop.
# After the draws, one runif() value is recorded: it pins down how many uniform numbers the
# draws consumed, which the indices alone cannot show for n = 1.

outputFile <- commandArgs(trailingOnly = TRUE)[1]
if (is.na(outputFile)) stop("usage: Rscript r/reference/edge-indices.R <output file>")

lines <- character(0)
for (n in c(1L, 2L, 40000L, 65536L, 65537L)) {
    set.seed(431)
    draws <- sample.int(n, 500L, replace = TRUE)
    nextUniform <- runif(1)
    lines <- c(lines, paste(n, paste(draws, collapse = " "), sprintf("%.17g", nextUniform), sep = "\t"))
}
writeLines(c(paste0("# ", R.version.string, "; RNGkind: ", paste(RNGkind(), collapse = " / "),
                    "; set.seed(431); sample.int(n, 500, replace = TRUE); then runif(1)"),
             "n\tdraws\tnext_uniform", lines),
           outputFile)
cat("edge-indices: wrote", outputFile, "\n")
