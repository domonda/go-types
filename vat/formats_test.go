package vat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCheckSums_edgeBranches exercises the rarely hit branches of the
// country check-sum functions: fallback weight sets, remainders that map
// to a check digit of 0, remainders that make a number invalid by
// construction, and number forms that are deliberately validated by shape
// only. A wrong branch here silently rejects real VAT IDs or accepts
// mistyped ones.
//
// Most IDs below are synthetic: they were searched for to hit one specific
// branch and independently verified with python-stdnum 2.2
// (stdnum.eu.vat / stdnum.no.mva), so they do not just mirror this
// package's own implementation.
func TestCheckSums_edgeBranches(t *testing.T) {
	valid := []struct {
		id     ID
		branch string
	}{
		{"BG010978463", "BG 9-digit EIK: first remainder is 10, fallback weights"},
		{"BG010447890", "BG 9-digit EIK: fallback remainder is also 10, check digit 0"},
		{"BG5001010050", "BG 10-digit EGN: remainder 10 maps to check digit 0"},
		{"CZ12345776", "CZ 8-digit legal entity: regular case, check digit is 11 - remainder"},
		{"CZ12345831", "CZ 8-digit legal entity: remainder 0 maps to check digit 1"},
		{"CZ12345750", "CZ 8-digit legal entity: remainder 1 maps to check digit 0"},
		{"CZ010107373", "CZ 9-digit personal number: shape only"},
		{"CZ600064954", "CZ 9-digit individual starting with 6: shape only"},
		{"CZ7103192745", "CZ 10-digit personal number: shape only"},
		{"EL010685460", "EL remainder 10 maps to check digit 0"},
		{"FI11002220", "FI remainder 0 means check digit 0"},
		{"FRK7399859412", "FR alphanumeric key: shape only"},
		{"HR00010036102", "HR ISO 7064 intermediate value 0 is replaced by 10"},
		{"LT011081410", "LT first remainder is 10, retry with shifted weights"},
		{"LT026919410", "LT shifted remainder is also 10, check digit 0"},
		{"LV32000000091", "LV personal code without birth date: shape only"},
		{"LV10107028484", "LV personal code with birth date: shape only"},
		{"NO916634730MVA", "NO remainder 0 means check digit 0"},
		{"PT501188320", "PT remainder 0 (check 11) maps to check digit 0"},
		{"PT500475610", "PT remainder 1 (check 10) maps to check digit 0"},
		{"RO10051940", "RO check value 10 maps to check digit 0"},
		{"SI10447890", "SI check value 10 maps to check digit 0"},
		{"ESK1234567L", "ES historical NIF with K prefix uses the DNI letter table"},
		{"ESL1234567L", "ES historical NIF with L prefix uses the DNI letter table"},
		{"ESM1234567L", "ES historical NIF with M prefix uses the DNI letter table"},
		{"ESC12345674", "ES CIF entity C accepts a check digit"},
		{"ESC1234567D", "ES CIF entity C also accepts the equivalent check letter"},
	}
	for _, tt := range valid {
		t.Run(string(tt.id), func(t *testing.T) {
			norm, err := tt.id.Normalized()
			assert.NoError(t, err, tt.branch)
			assert.Equal(t, tt.id, norm, tt.branch)
		})
	}

	// For these numbers no last digit can make them valid,
	// so every variant of the last digit must be rejected.
	invalidForAnyLastDigit := []struct {
		id     ID
		branch string
	}{
		{"FI10146968", "FI remainder 1 is invalid by construction"},
		{"LV40003522562", "LV legal entity remainder 4 would need check digit 10"},
		{"NO916634790", "NO remainder 1 would need check digit 10"},
		{"PL0010131130", "PL remainder 10 would need check digit 10"},
	}
	for _, tt := range invalidForAnyLastDigit {
		t.Run(string(tt.id), func(t *testing.T) {
			body := tt.id[:len(tt.id)-1]
			for d := '0'; d <= '9'; d++ {
				id := body + ID(d)
				assert.Falsef(t, id.Valid(), "%s: %s", id, tt.branch)
			}
		})
	}
}
