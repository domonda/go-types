package bank

import (
	"encoding"
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/domonda/go-types/country"
	"github.com/domonda/go-types/internal/normtest"
	"github.com/domonda/go-types/strutil"
)

var validBICs = []BIC{
	"BKAUATWW",
	"GIBAATWWXXX",
	"BELADEBEXXX",
	"RBOSGGSX",
	"CHASGB2LXXX",
	"RZTIAT22263",
	"BCEELULL",
	"MARKDEFF",
	"MARKDEFFXXX",
	"GENODEF1JEV",
	"UBSWCHZH80A",
	"CEDELULLXXX",
	"HELADEF1RRS",
	"HELADEFF",
	"GENODEF1S04",
	"GENODE61MA2",
	"DEUTDEDBMAN",
	"SOLADES1HDB",
}

var validBICsWithSpaces = []BIC{
	"BKAU AT WW",
	"GIBAATWW XXX",
	" BELADEBEXXX ",
	"RBOSGGSX   ",
}

var invalidBICs = []BIC{
	"BELADEBEXX",
	"bKAUATWW",
	"GIBAATWWX01",
	"GENODEFOJEV",
	"AMTSGERICHT", // valid syntax, but not a BIC
	"AUTOBANK",    // valid syntax, but not a BIC
	"DEUTSCHLAND", // valid syntax, but not a BIC
	"DIENSTGEBER", // valid syntax, but not a BIC
	"DOCUMENT",    // valid syntax, but not a BIC
	"DOKUMENT",    // valid syntax, but not a BIC
	"FACILITY",    // valid syntax, but not a BIC
	"GELISTET",    // valid syntax, but not a BIC
	"GESAMTNETTO", // valid syntax, but not a BIC
}

func Test_BICValid(t *testing.T) {
	for _, bic := range validBICs {
		if !bic.Valid() {
			t.Errorf("Valid BIC not recognized: %s", string(bic))
		}
	}
	for _, bic := range invalidBICs {
		if bic.Valid() {
			t.Errorf("Invalid BIC not recognized: %s", string(bic))
		}
	}
}

// Valid and Validate accept a BIC with whitespace because they normalize
// first like Normalized, the same way the other types of this module do,
// and Value and MarshalJSON write the normalized BIC.
func TestBIC_ValidateNormalizes(t *testing.T) {
	for _, bic := range validBICsWithSpaces {
		assert.NoError(t, bic.Validate(), "BIC(%q).Validate()", string(bic))
		assert.True(t, bic.Valid(), "BIC(%q).Valid()", string(bic))
	}
	for _, bic := range invalidBICs {
		assert.Error(t, bic.Validate(), "BIC(%q).Validate()", string(bic))
	}
}

// Value and MarshalJSON write a BIC normalized to 11 characters, because
// Validate accepts the 8 character form and whitespace: without
// Value and MarshalJSON trim and uppercase. They do not append "XXX"
// and do not remove spaces inside the BIC. An invalid BIC is written
// after that prettification, not refused.
func TestBIC_ValueAndMarshalJSONNormalize(t *testing.T) {
	tests := []struct {
		bic  BIC
		want string
	}{
		{bic: "BKAUATWWXXX", want: "BKAUATWWXXX"},
		{bic: "BKAUATWW", want: "BKAUATWW"},
		{bic: "BKAU AT WW", want: "BKAU AT WW"},
		{bic: "bkauatww", want: "BKAUATWW"},
		{bic: " GENODEF1JEV ", want: "GENODEF1JEV"},
	}
	for _, tt := range tests {
		v, err := tt.bic.Value()
		require.NoError(t, err)
		assert.Equal(t, tt.want, v, "BIC(%q).Value()", string(tt.bic))

		j, err := json.Marshal(tt.bic)
		require.NoError(t, err)
		assert.Equal(t, `"`+tt.want+`"`, string(j), "json.Marshal(BIC(%q))", string(tt.bic))
	}

	v, err := BIC("deutzzff").Value()
	require.NoError(t, err)
	assert.Equal(t, "DEUTZZFF", v, "Value() of an invalid BIC is trimmed and uppercased")
	j, err := json.Marshal(BIC("deutzzff"))
	require.NoError(t, err)
	assert.Equal(t, `"DEUTZZFF"`, string(j))

	v, err = BIC("").Value()
	require.NoError(t, err)
	assert.Nil(t, v, "an empty BIC is written as SQL NULL")
}

func TestBIC_Normalized(t *testing.T) {
	for _, bic := range append(validBICs, validBICsWithSpaces...) {
		normalized, err := bic.Normalized()
		if err != nil {
			t.Errorf("Error normalizing BIC: %s", err)
		}
		expected := BIC(strings.ReplaceAll(string(bic), " ", ""))
		if len(expected) == 8 {
			expected += "XXX"
		}
		require.Equalf(t, expected, normalized, "Normalized BIC %q", bic)
	}
	for _, bic := range invalidBICs {
		norm, err := bic.Normalized()
		require.Error(t, err, "Normalized invalid BIC error expected")
		require.Equal(t, bic, norm, "Normalized invalid BIC returned unchanged")
	}
}

func TestBIC_Parse(t *testing.T) {
	tests := []struct {
		name            string
		input           BIC
		wantIsValid     bool
		wantBankCode    string
		wantCountryCode country.Code
		wantBranchCode  string
	}{
		{
			name:        "Invalid length",
			input:       "ABCDE",
			wantIsValid: false,
		},
		{
			name:        "Regexp no match",
			input:       "12345678",
			wantIsValid: false,
		},
		{
			name:        "Country code invalid",
			input:       "DEUTZZFF",
			wantIsValid: false,
		},
		{
			name:        "In falseBICs",
			input:       "FAKEBIC12",
			wantIsValid: false,
		},
		{
			name:            "Valid BIC 11 chars",
			input:           "DEUTDEFF500",
			wantIsValid:     true,
			wantBankCode:    "DEUT",
			wantCountryCode: "DE",
			wantBranchCode:  "500",
		},
		{
			name:            "Valid BIC 8 chars",
			input:           "DEUTDEFF",
			wantIsValid:     true,
			wantBankCode:    "DEUT",
			wantCountryCode: "DE",
			wantBranchCode:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bankCode, countryCode, branchCode, isValid := tc.input.Parse()
			require.Equal(t, tc.wantIsValid, isValid)

			assert.Equal(t, tc.wantBankCode, bankCode)
			assert.Equal(t, tc.wantCountryCode, countryCode)
			assert.Equal(t, tc.wantBranchCode, branchCode)
		})
	}
}

func TestBIC_Parse_SWIFTRules(t *testing.T) {
	// Every position of a BIC has its own allowed character set
	// according to ISO 9362, each must be checked individually.
	tests := []struct {
		name string
		bic  BIC
	}{
		{name: "digit in bank code", bic: "DEU1DEFF"},
		{name: "digit in country code", bic: "DEUTD1FF"},
		{name: "lowercase country code", bic: "DEUTdeFF"},
		{name: "location 0 not allowed", bic: "DEUTDE0F"},
		{name: "location 1 not allowed", bic: "DEUTDE1F"},
		{name: "location second char O not allowed", bic: "DEUTDEFO"},
		{name: "location second char lowercase", bic: "DEUTDEFf"},
		{name: "branch starting with X other than XXX", bic: "DEUTDEFFXAB"},
		{name: "branch starting lowercase", bic: "DEUTDEFFa00"},
		{name: "branch second char invalid", bic: "DEUTDEFF5a0"},
		{name: "branch third char invalid", bic: "DEUTDEFF50-"},
		{name: "length 9", bic: "DEUTDEFF5"},
		{name: "length 10", bic: "DEUTDEFF50"},
		{name: "length 12", bic: "DEUTDEFF5000"},
		{name: "empty", bic: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bankCode, countryCode, branchCode, isValid := tt.bic.Parse()
			assert.False(t, isValid)
			assert.Empty(t, bankCode, "no partial results for invalid BIC")
			assert.Empty(t, countryCode, "no partial results for invalid BIC")
			assert.Empty(t, branchCode, "no partial results for invalid BIC")
			assert.Error(t, tt.bic.Validate())
		})
	}
}

// The JSON schema published for BIC must accept the same BICs as Go
func TestBIC_JSONSchema(t *testing.T) {
	s := BIC("").JSONSchema()
	assert.Equal(t, "string", s.Type)
	require.Equal(t, BICRegex, s.Pattern)
	re := regexp.MustCompile(s.Pattern)
	for _, bic := range validBICs {
		assert.Regexp(t, re, string(bic))
	}
}

func TestBIC_Normalize(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		tests := []struct {
			bic       BIC
			wantLong  BIC
			wantShort BIC
		}{
			{"BELADEBE", "BELADEBEXXX", "BELADEBE"},
			{"BELADEBEXXX", "BELADEBEXXX", "BELADEBE"},
			{"GIBAATWW XXX", "GIBAATWWXXX", "GIBAATWW"},
			{"BKAU AT WW", "BKAUATWWXXX", "BKAUATWW"},
			// All whitespace is removed like by IBAN.Normalized, not only
			// ASCII spaces, because a BIC next to an IBAN in a document or
			// form must not fail on a tab, newline or no-break space
			{"BKAU\tAT\tWW", "BKAUATWWXXX", "BKAUATWW"},
			{"BELADEBEXXX\r\n", "BELADEBEXXX", "BELADEBE"},
			{" GIBA ATWW XXX", "GIBAATWWXXX", "GIBAATWW"},
			{"BELA​DEBE", "BELADEBEXXX", "BELADEBE"},
			// A real branch code must never be trimmed
			{"GENODEF1JEV", "GENODEF1JEV", "GENODEF1JEV"},
		}
		for _, tt := range tests {
			t.Run(string(tt.bic), func(t *testing.T) {
				long, err := NormalizeBIC(string(tt.bic))
				require.NoError(t, err)
				assert.Equal(t, tt.wantLong, long)
				assert.Equal(t, strings.ToUpper(strutil.TrimSpace(string(tt.bic))), marshalText(t, tt.bic))
				assert.Equal(t, NullableBIC(tt.wantLong), tt.bic.NormalizedOrNull())
				assert.Equal(t, tt.bic == tt.wantLong, tt.bic.ValidAndNormalized())

				short, err := tt.bic.NormalizedShort()
				require.NoError(t, err)
				assert.Equal(t, tt.wantShort, short)
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		for _, bic := range invalidBICs {
			_, err := NormalizeBIC(string(bic))
			assert.Error(t, err)
			short, err := bic.NormalizedShort()
			assert.Error(t, err)
			assert.Equal(t, bic, short, "invalid BIC must be returned unchanged")
			assert.Equal(t, BICNull, bic.NormalizedOrNull())
			assert.Equal(t, strings.ToUpper(strutil.TrimSpace(string(bic))), marshalText(t, bic), "MarshalText trims and uppercases an invalid BIC")
			assert.False(t, bic.ValidAndNormalized())
			assert.False(t, StringIsBIC(string(bic)))
		}
	})
}

func TestStringIsBIC(t *testing.T) {
	for _, bic := range validBICs {
		assert.True(t, StringIsBIC(string(bic)), bic)
	}
}

func TestBIC_CountryCode(t *testing.T) {
	assert.Equal(t, country.DE, BIC("BELADEBEXXX").CountryCode())
	assert.Equal(t, country.GG, BIC("RBOSGGSX").CountryCode())
	assert.Equal(t, country.Code(""), BIC("DEUTZZFF").CountryCode(), "invalid BIC has no country")
	assert.Equal(t, country.AT, BIC(" BKAU AT WW ").CountryCode(), "a BIC that Valid accepts has a country")
}

func TestBIC_TrimBranchCode(t *testing.T) {
	assert.Equal(t, BIC("GENODEF1"), BIC("GENODEF1JEV").TrimBranchCode())
	assert.Equal(t, BIC("BELADEBE"), BIC("BELADEBE").TrimBranchCode())
	assert.Equal(t, BIC("SHORT"), BIC("SHORT").TrimBranchCode(), "must not panic for short strings")
	// Valid accepts whitespace, so the branch code must be trimmed
	// from the normalized BIC instead of slicing the raw string
	assert.Equal(t, BIC("GENODEF1"), BIC(" GENO DEF1 JEV ").TrimBranchCode())
}

func TestBIC_LocationFlags(t *testing.T) {
	// The second character of the location code encodes
	// test BICs ('0'), passive participants ('1')
	// and receiver pays fees ('2').
	tests := []struct {
		bic                  BIC
		wantTest             bool
		wantPassive          bool
		wantReceiverPaysFees bool
	}{
		{bic: "DEUTDEF0", wantTest: true},
		{bic: "DEUTDEF0500", wantTest: true},
		{bic: "DEUTDEF1", wantPassive: true},
		{bic: "DEUTDEF2", wantReceiverPaysFees: true},
		{bic: "DEUTDEFF"},
		{bic: "DEUTDEFFXXX"},
		// Read from the normalized BIC, because Valid accepts whitespace
		{bic: " DEUT DE F0 ", wantTest: true},
		{bic: "DEUT DE F1", wantPassive: true},
		// Invalid BICs never have flags, even if the 8th character matches
		{bic: "deutDEF0"},
		{bic: "DEUTZZF1"},
		// Too short strings must not panic
		{bic: ""},
		{bic: "DEUT"},
	}
	for _, tt := range tests {
		t.Run(string(tt.bic), func(t *testing.T) {
			assert.Equal(t, tt.wantTest, tt.bic.IsTestBIC(), "IsTestBIC")
			assert.Equal(t, tt.wantPassive, tt.bic.IsPassiveSWIFT(), "IsPassiveSWIFT")
			assert.Equal(t, tt.wantReceiverPaysFees, tt.bic.ReceiverPaysFees(), "ReceiverPaysFees")
		})
	}
}

func TestBIC_ScanString(t *testing.T) {
	t.Run("validate accepts valid BIC", func(t *testing.T) {
		var bic BIC
		require.NoError(t, bic.ScanString("BELADEBEXXX", true))
		assert.Equal(t, BIC("BELADEBEXXX"), bic)
	})
	t.Run("validate rejects invalid BIC", func(t *testing.T) {
		bic := BIC("unchanged")
		require.Error(t, bic.ScanString("DEUTZZFF", true))
		assert.Equal(t, BIC("unchanged"), bic)
	})
	t.Run("without validate assigns invalid BIC as is", func(t *testing.T) {
		var bic BIC
		require.NoError(t, bic.ScanString("DEUTZZFF", false))
		assert.Equal(t, BIC("DEUTZZFF"), bic)
	})
}

func TestBIC_SQL(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    BIC
		wantErr bool
	}{
		{name: "string", value: "BELADEBEXXX", want: "BELADEBEXXX"},
		{name: "bytes", value: []byte("BELADEBEXXX"), want: "BELADEBEXXX"},
		{name: "NULL", value: nil, want: ""},
		// Scan does not validate, so existing invalid data can still be read
		{name: "invalid string", value: "invalid", want: "INVALID"},
		{name: "unsupported type", value: int64(1), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bic := BIC("unchanged")
			err := bic.Scan(tt.value)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, BIC("unchanged"), bic)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, bic)
		})
	}

	t.Run("Scan copies bytes", func(t *testing.T) {
		// database/sql may reuse the []byte buffer after Scan returns
		buf := []byte("BELADEBEXXX")
		var bic BIC
		require.NoError(t, bic.Scan(buf))
		copy(buf, "XXXXXXXXXXX")
		assert.Equal(t, BIC("BELADEBEXXX"), bic)
	})

	t.Run("Value round trip", func(t *testing.T) {
		v, err := BIC("BELADEBEXXX").Value()
		require.NoError(t, err)
		assert.Equal(t, "BELADEBEXXX", v)
		var bic BIC
		require.NoError(t, bic.Scan(v))
		assert.Equal(t, BIC("BELADEBEXXX"), bic)
	})
}

// TestBIC_FalseBICWithBranchCode protects the list of words that look like
// BICs: Normalized appends "XXX" to an 8 character BIC, so the false BIC
// "AUTOBANK" was valid as "AUTOBANKXXX" and with whitespace.
func TestBIC_FalseBICWithBranchCode(t *testing.T) {
	for _, bic := range []BIC{"AUTOBANK", "AUTOBANKXXX", "AUTOBANK XXX", "autobank"} {
		assert.False(t, bic.Valid(), "BIC(%q).Valid()", bic)
		_, _, _, isValid := bic.Parse()
		assert.False(t, isValid, "BIC(%q).Parse()", bic)
	}
	assert.Empty(t, BICFinder.FindAllIndex([]byte("Kontakt: AUTOBANKXXX"), -1))
	assert.True(t, BIC("DEUTDEFFXXX").Valid())
}

// marshalText returns the text that MarshalText writes for v,
// which is what encoding/xml and strfmt write.
func marshalText(t *testing.T, v encoding.TextMarshaler) string {
	t.Helper()
	text, err := v.MarshalText()
	require.NoError(t, err)
	return string(text)
}

// TestBIC_ReadingAndWriting checks the reading and writing contract shared
// by the normalizing types: every reading path stores a valid BIC
// trimmed and uppercased, an 8-character BIC kept, an invalid BIC written
// that way, and fmt prints the stored BIC.
func TestBIC_ReadingAndWriting(t *testing.T) {
	normtest.Run(t, normtest.Spec[BIC]{
		Canonical: "BKAUATWWXXX",
		Variants:  []BIC{"BKAUATWW", "BKAU AT WW", " BKAUATWWXXX "},
		Invalid:   []BIC{"DEUTZZFF", "ABC"},
		Prettify:  normtest.TrimUpper[BIC],
	})
}
