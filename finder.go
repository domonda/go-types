package types

// Finder defines an interface for finding patterns in byte slices.
// It provides a method to find all occurrences of a pattern within a string,
// returning the start and end indices of each match.
type Finder interface {
	// FindAllIndex finds all occurrences of the pattern in str and returns
	// a slice of index pairs [start, end] for each match.
	// The parameter n limits the number of matches returned:
	//   - n < 0: return all matches
	//   - n == 0: return no matches
	//   - n > 0: return at most n matches
	FindAllIndex(str []byte, n int) [][]int
}
