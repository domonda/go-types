package date

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Invoices and bank statements from French, Italian and Spanish
// counterparties write the month as a word, so every month name and
// its usual abbreviation must parse, not only January.
func TestNormalize_RomanceMonthNames(t *testing.T) {
	months := []struct {
		month                    string
		french, italian, spanish []string
	}{
		{"01", []string{"janv.", "janvier"}, []string{"gen", "gennaio"}, []string{"ene", "enero"}},
		{"02", []string{"févr.", "février", "fevr"}, []string{"feb", "febbraio"}, []string{"feb", "febrero"}},
		{"03", []string{"mars"}, []string{"mar", "marzo"}, []string{"mar", "marzo"}},
		{"04", []string{"avr.", "avril"}, []string{"apr", "aprile"}, []string{"abr", "abril"}},
		{"05", []string{"mai"}, []string{"mag", "maggio"}, []string{"may", "mayo"}},
		{"06", []string{"juin"}, []string{"giu", "giugno"}, []string{"jun", "junio"}},
		{"07", []string{"juil.", "juillet"}, []string{"lug", "luglio"}, []string{"jul", "julio"}},
		{"08", []string{"août", "aout"}, []string{"ago", "agosto"}, []string{"ago", "agosto"}},
		{"09", []string{"sept.", "septembre"}, []string{"set", "settembre"}, []string{"sep", "septiembre", "setiembre"}},
		{"10", []string{"oct.", "octobre"}, []string{"ott", "ottobre"}, []string{"oct", "octubre"}},
		{"11", []string{"nov.", "novembre"}, []string{"nov", "novembre"}, []string{"nov", "noviembre"}},
		{"12", []string{"déc.", "décembre"}, []string{"dic", "dicembre"}, []string{"dic", "diciembre"}},
	}
	for _, m := range months {
		want := Date("2019-" + m.month + "-02")
		for _, names := range [][]string{m.french, m.italian, m.spanish} {
			for _, name := range names {
				input := "2 " + name + " 2019"
				got, err := Normalize(input)
				require.NoError(t, err, "Normalize(%q)", input)
				require.Equal(t, want, got, "Normalize(%q)", input)
			}
		}
	}
}
