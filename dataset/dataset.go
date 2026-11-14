// Package dataset reads the one-column CSV files the analysis runs on: a header line, then
// one number per line. It validates as it reads, so a bad file is reported with its line.
package dataset

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

// MinObservations is the fewest observations a data set may have: a standard error needs two.
const MinObservations = 2

// ErrInvalid marks every error caused by the content of a data set (as opposed to a file
// that cannot be opened). Test for it with errors.Is.
var ErrInvalid = errors.New("invalid data set")

// ReadFile reads the data set stored at path.
func ReadFile(path string) ([]float64, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read data set: %w", err)
	}
	defer file.Close()
	return Read(file, path)
}

// Read parses a data set from reader; name is used in error messages.
//
// The first line must be a header (any text that is not a number), matching what the R
// scripts skip. Blank lines are ignored, and Windows line endings are accepted. Every other
// line must be one finite number.
func Read(reader io.Reader, name string) ([]float64, error) {
	var values []float64
	scanner := bufio.NewScanner(reader)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		text := strings.TrimSpace(scanner.Text())
		value, parseError := strconv.ParseFloat(text, 64)
		switch {
		case lineNumber == 1 && parseError == nil:
			return nil, fmt.Errorf("%w: %s, line 1: found the number %q where the header line should be; "+
				"add a header such as \"value\"", ErrInvalid, name, text)
		case lineNumber == 1 || text == "":
			continue
		case parseError != nil:
			return nil, fmt.Errorf("%w: %s, line %d: %q is not a number", ErrInvalid, name, lineNumber, text)
		case math.IsNaN(value) || math.IsInf(value, 0):
			return nil, fmt.Errorf("%w: %s, line %d: %q is not a finite number", ErrInvalid, name, lineNumber, text)
		}
		values = append(values, value)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("cannot read data set %s: %w", name, err)
	}
	if len(values) < MinObservations {
		return nil, fmt.Errorf("%w: %s has %d observation(s); at least %d are needed",
			ErrInvalid, name, len(values), MinObservations)
	}
	return values, nil
}
