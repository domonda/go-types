package date

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/domonda/go-types"
	"github.com/domonda/go-types/language"
)

var _ types.Finder = (*Finder)(nil)

func TestFinder_FindAllIndex_Limit(t *testing.T) {
	// types.Finder contract: n < 0 returns all matches,
	// n == 0 returns none, n > 0 returns at most n matches.
	// Callers pass n to stop scanning early, e.g. n = 1 to find the first date.
	str := []byte("2006-01-02, 2017/12/03 and 24.12.2018")
	finder := NewFinder()

	all := finder.FindAllIndex(str, -1)
	require.Equal(t, [][]int{{0, 10}, {12, 22}, {27, 37}}, all)

	assert.Nil(t, finder.FindAllIndex(str, 0), "n == 0")
	assert.Equal(t, all[:1], finder.FindAllIndex(str, 1), "n == 1")
	assert.Equal(t, all[:2], finder.FindAllIndex(str, 2), "n == 2")
	assert.Equal(t, all, finder.FindAllIndex(str, 3), "n == number of matches")
	assert.Equal(t, all, finder.FindAllIndex(str, 10), "n > number of matches")
}

func TestFinder_LanguageHintKeepsUnambiguousDates(t *testing.T) {
	// The language hint only decides ambiguous dd/mm vs mm/dd orderings,
	// so an EN finder must still find a date that can only be day-first.
	str := "Rechnungsdatum: 15.06.2024"
	got := NewFinder(language.EN).FindAllIndex([]byte(str), -1)
	require.Equal(t, [][]int{{16, 26}}, got)
	assert.Equal(t, Date("2024-06-15"), Date(str[16:26]).NormalizedOrUnchanged(language.EN))
}
