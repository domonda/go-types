package vat

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/domonda/go-types"
)

var _ types.Finder = IDFinder

func TestIDFinder_FindAllIndex_Limit(t *testing.T) {
	// types.Finder contract: n < 0 returns all matches,
	// n == 0 returns none, n > 0 returns at most n matches.
	// Callers pass n to stop scanning early, e.g. n = 1 to find the first VAT ID.
	str := []byte("ATU10223006 DE 136725570 EL094019245")

	all := IDFinder.FindAllIndex(str, -1)
	require.Equal(t, [][]int{{0, 11}, {12, 24}, {25, 36}}, all)

	assert.Nil(t, IDFinder.FindAllIndex(str, 0), "n == 0")
	assert.Equal(t, all[:1], IDFinder.FindAllIndex(str, 1), "n == 1")
	assert.Equal(t, all[:2], IDFinder.FindAllIndex(str, 2), "n == 2")
	assert.Equal(t, all, IDFinder.FindAllIndex(str, 3), "n == number of matches")
	assert.Equal(t, all, IDFinder.FindAllIndex(str, 10), "n > number of matches")
}
