# generate-data.R: the 12 data sets every R and Go run reads (spec FR-1).
#
# Usage: Rscript r/generate-data.R <output folder>
#
# Three shapes at four sizes, each from its own seed so any file can be regenerated alone:
#   normal              Normal(mean 100, sd 10)                      symmetric
#   lognormal           Lognormal(meanlog log(100) - 0.125, sdlog 0.5)  positively skewed, mean 100
#   mirrored-lognormal  200 minus a lognormal as above               negatively skewed, mean 100
#
# Values are rounded to 4 decimal places and written one per line under the header "value".
# The script also writes MANIFEST-generation.tsv; scripts/make-data.sh adds sizes and hashes.

outputFolder <- commandArgs(trailingOnly = TRUE)[1]
if (is.na(outputFolder)) stop("usage: Rscript r/generate-data.R <output folder>")

sizes <- c(100L, 1000L, 10000L, 100000L)
meanlog <- log(100) - 0.125
sdlog <- 0.5

shapes <- list(
    list(name = "normal",
         call = "rnorm(n, mean = 100, sd = 10)",
         draw = function(n) rnorm(n, mean = 100, sd = 10)),
    list(name = "lognormal",
         call = "rlnorm(n, meanlog = log(100) - 0.125, sdlog = 0.5)",
         draw = function(n) rlnorm(n, meanlog = meanlog, sdlog = sdlog)),
    list(name = "mirrored-lognormal",
         call = "200 - rlnorm(n, meanlog = log(100) - 0.125, sdlog = 0.5)",
         draw = function(n) 200 - rlnorm(n, meanlog = meanlog, sdlog = sdlog))
)

manifestLines <- "file\tshape\tn\tgenerator\tseed"
seed <- 8000L
for (shape in shapes) {
    for (n in sizes) {
        seed <- seed + 1L
        set.seed(seed)
        values <- round(shape$draw(n), 4)
        fileName <- paste0(shape$name, "-", format(n, scientific = FALSE), ".csv")
        writeLines(c("value", formatC(values, format = "f", digits = 4)),
                   file.path(outputFolder, fileName))
        manifestLines <- c(manifestLines,
                           paste(fileName, shape$name, format(n, scientific = FALSE),
                                 shape$call, seed, sep = "\t"))
    }
}
writeLines(c(paste0("# ", R.version.string, "; values rounded to 4 decimals"), manifestLines),
           file.path(outputFolder, "MANIFEST-generation.tsv"))
cat("generate-data: wrote", length(manifestLines) - 1, "data sets to", outputFolder, "\n")
