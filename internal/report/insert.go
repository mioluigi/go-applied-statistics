package report

import (
	"fmt"
	"strings"
)

// InsertBlocks replaces, in a document, the text between each block's start and end marker
// with the block's body. Every block must have exactly one marker pair in the document, and
// the document must hold no marker for a block that does not exist: a missing or misspelt
// marker would otherwise leave a stale number in place without anyone noticing.
func InsertBlocks(document string, blocks []Block) (string, error) {
	known := map[string]bool{}
	for _, block := range blocks {
		known[block.Name] = true
		start := fmt.Sprintf("<!-- generated:%s:start -->", block.Name)
		end := fmt.Sprintf("<!-- generated:%s:end -->", block.Name)
		if strings.Count(document, start) != 1 || strings.Count(document, end) != 1 {
			return "", fmt.Errorf("the document needs exactly one %q and one %q", start, end)
		}
		from := strings.Index(document, start) + len(start)
		to := strings.Index(document, end)
		if to < from {
			return "", fmt.Errorf("marker %q comes before %q", end, start)
		}
		document = document[:from] + "\n" + block.Body + document[to:]
	}
	for _, piece := range strings.Split(document, "<!-- generated:")[1:] {
		name, _, _ := strings.Cut(piece, ":")
		if !known[name] {
			return "", fmt.Errorf("the document has a marker for %q, which is not a generated block", name)
		}
	}
	return document, nil
}
