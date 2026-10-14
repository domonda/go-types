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

func TestFinder_FindsISO8601DateTimes(t *testing.T) {
	// Normalize accepts an ISO 8601 date-time and uses its date part, so
	// the finder must find it too, with a range that covers the whole
	// date-time: the caller passes the found text to Normalize. The finder
	// used to lower case the candidate first, which hid the upper case T
	// from normalizeDate, so timestamps in text were never found.
	tests := []struct {
		text  string
		want  [][]int
		dates []Date
	}{
		{text: "2023-12-25T10:00:00Z", want: [][]int{{0, 20}}, dates: []Date{"2023-12-25"}},
		{text: "2023-12-25T10:00:00", want: [][]int{{0, 19}}, dates: []Date{"2023-12-25"}},
		{text: "sent at 2023-12-25T10:00:00+02:00.", want: [][]int{{8, 33}}, dates: []Date{"2023-12-25"}},
		{text: "(2024-02-29T23:59:59.123Z) and 01.03.2024", want: [][]int{{1, 25}, {31, 41}}, dates: []Date{"2024-02-29", "2024-03-01"}},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := NewFinder().FindAllIndex([]byte(tt.text), -1)
			require.Equal(t, tt.want, got)
			for i, idx := range got {
				d, err := Normalize(tt.text[idx[0]:idx[1]])
				require.NoError(t, err, "Normalize of the found text %q", tt.text[idx[0]:idx[1]])
				assert.Equal(t, tt.dates[i], d)
			}
		})
	}

	// The finder finds exactly what Normalize accepts, so a date-time with
	// a lower case t, which Normalize rejects, is not found either.
	_, err := Normalize("2023-12-25t10:00:00z")
	require.Error(t, err)
	assert.Nil(t, NewFinder().FindAllIndex([]byte("2023-12-25t10:00:00z"), -1))
}

// TestFinder_NoDateInWordsAndSingleDigits protects against false dates in
// free text: "set" is a month name and "6 4" no day and year.
func TestFinder_NoDateInWordsAndSingleDigits(t *testing.T) {
	for _, text := range []string{"won the set 6 4 easily", "2023-12-25TOTAL"} {
		assert.Empty(t, NewFinder().FindAllIndex([]byte(text), -1), text)
	}
}

// TestFinder_CommonWordMonthNames protects text extraction from false dates:
// the Italian and Spanish month abbreviations "set", "ago", "mag", "lug",
// "ott" and "dic" are common words in other languages, so the finder only
// reads them as month with Italian or Spanish as language hint.
// Normalize still reads them, because its input is meant to be a date.
func TestFinder_CommonWordMonthNames(t *testing.T) {
	for _, text := range []string{"Game set 10 2020", "the mag 12 2021", "lug 10 22", "2023 12 ago", "ott 3 2020", "dic 24 2020"} {
		assert.Nil(t, NewFinder().FindAllIndex([]byte(text), -1), "no hint: %q", text)
		assert.Nil(t, NewFinder(language.EN).FindAllIndex([]byte(text), -1), "EN: %q", text)
	}

	assert.Equal(t, [][]int{{0, 11}}, NewFinder(language.IT).FindAllIndex([]byte("10 set 2020"), -1))
	assert.Equal(t, [][]int{{0, 11}}, NewFinder("IT").FindAllIndex([]byte("10 set 2020"), -1), "hint not normalized")
	assert.Equal(t, [][]int{{3, 14}}, NewFinder(language.ES).FindAllIndex([]byte("el 10 ago 2020"), -1))
	assert.Nil(t, NewFinder(language.ES).FindAllIndex([]byte("10 set 2020"), -1), "set is no Spanish month")
	// The full names and the other abbreviations need no hint
	assert.Equal(t, [][]int{{0, 17}}, NewFinder().FindAllIndex([]byte("10 settembre 2020"), -1))
	assert.Equal(t, [][]int{{0, 11}}, NewFinder().FindAllIndex([]byte("10 sep 2020"), -1))

	norm, err := Normalize("10 set 2020")
	require.NoError(t, err)
	assert.Equal(t, Date("2020-09-10"), norm)
}
