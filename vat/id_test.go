package vat

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/domonda/go-types/country"
)

var validVATIDs = map[string]string{
	// "ATU Nr. 1-022 3 006": "ATU10223006", // Real world encounter, but we'll probably never support this mess
	"atu10223006":     "ATU10223006", // Example
	"ATU10223006":     "ATU10223006", // Example
	"ATU67554568":     "ATU67554568", // Real but inactive by now
	"ATU68765099":     "ATU68765099", // Real
	"ATU46983509":     "ATU46983509", // Real
	"ATU65785527":     "ATU65785527", // Real
	"ESW0184081H":     "ESW0184081H", // Real: Amazon EU Sarl Sucursal EN España
	"DE111111125":     "DE111111125", // Example
	"LT347776113":     "LT347776113",
	"DE 167015661":    "DE167015661",
	"ATU 10223006":    "ATU10223006",
	"AT U 10223006":   "ATU10223006",
	"at U 10223006":   "ATU10223006",
	"ATU.10223006":    "ATU10223006",
	"GB123456782":     "GB123456782", // 9-digit VRN, mod-97 valid
	"GB123456782012":  "GB123456782012",
	"GB 123456782012": "GB123456782012",
	"GBGD001":         "GBGD001",
	"GBHA599":         "GBHA599",
	"GB GD001":        "GBGD001",
	"GB HA599":        "GBHA599",
	"IE9S99999L":      "IE9S99999L",
	"IE 9999999LI":    "IE9999999LI",
	// Belgium — first digit 0 or 1, mod-97 control on last 2 digits.
	"BE0776091951": "BE0776091951",
	"BE1234567894": "BE1234567894",
	// Bulgaria — 9-digit EIK and 10-digit EGN both verified.
	"BG175074752":  "BG175074752",
	"BG7523169263": "BG7523169263",
	// Cyprus — 8 digits + control letter.
	"CY10001049T": "CY10001049T",
	// Czech Republic — 8-digit legal entity (mod-11).
	"CZ00177041": "CZ00177041",
	// Denmark — mod-11 weighted.
	"DK25313763": "DK25313763",
	// Estonia — mod-10 weighted.
	"EE100094916": "EE100094916",
	// Greece — mod-11 with powers-of-two weights.
	"EL094019245": "EL094019245",
	// Finland — mod-11 weighted.
	"FI09853608": "FI09853608",
	// France — 2 digit key + 9-digit SIREN, mod-97 derived.
	"FR40303265045": "FR40303265045",
	// Croatia — ISO 7064 MOD 11-10.
	"HR38192148118": "HR38192148118",
	// Hungary — mod-10 weighted.
	"HU12892312": "HU12892312",
	// Italy — Luhn over 11 digits.
	"IT00488410010": "IT00488410010",
	// Lithuania — 9-digit and 12-digit forms.
	"LT119511515":    "LT119511515",
	"LT290061371314": "LT290061371314",
	// Luxembourg — mod-89.
	"LU10000356": "LU10000356",
	// Latvia — legal entity (first digit > 3), mod-11 weighted.
	"LV40003521600": "LV40003521600",
	// Malta — mod-37.
	"MT15121333": "MT15121333",
	// Netherlands — "B" + 2-digit branch, format only (no checksum). The
	// pre-2020 number satisfied an 11-proof, but the post-2020 btw-id for
	// sole proprietors is randomly generated and does not.
	"NL005033019B01": "NL005033019B01", // pre-2020 form (also passes the old 11-proof)
	"NL002483642B72": "NL002483642B72", // post-2020 btw-id, valid in VIES, fails the old 11-proof
	// Poland — mod-11 weighted.
	"PL5260250274": "PL5260250274",
	// Portugal — mod-11 weighted.
	"PT502757191": "PT502757191",
	// Romania — variable length, left-padded mod-11.
	"RO18158683": "RO18158683",
	// Sweden — Luhn on the 10-digit org number; literal "01" branch.
	"SE556677889901": "SE556677889901",
	// Slovenia — mod-11 weighted.
	"SI80267432": "SI80267432",
	// Slovakia — full number divisible by 11.
	"SK2020317068": "SK2020317068",
	// Northern Ireland — same shape and mod-97 checksum as GB.
	"XI123456782":    "XI123456782",
	"XI123456782012": "XI123456782012",
	"XIGD001":        "XIGD001",
	"XIHA599":        "XIHA599",
	// Iceland VSK — 5 or 6 digits, no checksum.
	"IS12345":  "IS12345",
	"IS123456": "IS123456",
	// Liechtenstein — 5 digits, no checksum.
	"LI12345": "LI12345",
	// San Marino — 5 digits, no checksum.
	"SM12345":          "SM12345",
	"DE 1367 25570":    "DE136725570",
	"NO916634773":      "NO916634773",
	"NO 916634773":     "NO916634773",
	"NO 916634773 MVA": "NO916634773MVA",
	"NO977074010MVA":   "NO977074010MVA",
	"NO 977074010":     "NO977074010",
	"NO 977074010 MVA": "NO977074010MVA",
	"CHE-123.456.788":  "CHE123456788",
	"CHE123456788":     "CHE123456788",
	"EU372008134":      "EU372008134", // MOSS scheme VAT
	// Spain — coverage of all four sub-formats.
	"ES12345678Z": "ES12345678Z", // DNI / NIF, mod-23 letter = Z
	"ESY1234567X": "ESY1234567X", // NIE (Y → 1)
	"ESZ0000000M": "ESZ0000000M", // NIE (Z → 2)
	"ESX1234567L": "ESX1234567L", // NIE (X → 0, historic form)
	"ESA82018474": "ESA82018474", // CIF requiring numeric check
	"ESP0000000J": "ESP0000000J", // CIF requiring letter check
}

var invalidVATIDs = []ID{
	"atu12345678",
	"AT/U.12345678",
	" ATU12345678 ",
	"No. 62-1764389",
	"No.821764389",
	// Spain — must reject:
	"EST99600678", // T is not a valid CIF entity prefix
	"ES12345678A", // wrong DNI check letter (should be Z)
	"ESX1234567A", // wrong NIE check letter (should be L)
	"ESA82018470", // wrong CIF numeric check (should be 4)
	"ESP00000000", // P requires a letter check, not a digit
	"ESA0000000J", // A requires a digit check, not a letter
	// Norway — must reject (mod-11 checksum fails):
	"NO916634770", // valid format, last digit broken (3 → 0)
	// GB — regex precedence bug: each alternative must be anchored to ^GB...$
	"GBXGD123",    // 'X' between GB and GD must not slip through
	"GBHA999Y",    // trailing 'Y' after HA\d{3} must not slip through
	"GB12345678",  // 8 digits is not a valid GB length
	"GB123456789", // 9-digit shape OK but mod-97 checksum fails
	// Northern Ireland — shares GB regex + checksum.
	"XI123456789", // checksum fails
	"XIXGD123",    // precedence bug guard
	"XIHA999Y",    // precedence bug guard
	// Iceland — 4 or 7 digits are out of range.
	"IS1234",
	"IS1234567",
	// Liechtenstein — must be exactly 5 digits.
	"LI1234",
	"LI123456",
	// San Marino — must be exactly 5 digits.
	"SM1234",
	"SM123456",
	// IE — regex precedence bug: second alternative must be anchored to ^IE
	"IEX9999999LI", // unexpected 'X' before \d{7}[A-W][A-I] must not slip through
	// Belgium — first digit must be 0 or 1.
	"BE2345678901", // starts with 2
	"BE9999999999", // starts with 9
	"BE123456789",  // only 9 digits
	"BE0776091950", // valid format, mod-97 control broken
	// Bulgaria — 9-digit EIK and 10-digit EGN checksums.
	"BG175074750",  // EIK: last digit broken
	"BG7523169260", // EGN: last digit broken
	// Cyprus — control letter must match.
	"CY10001049X", // wrong control letter (should be T)
	// Czech Republic — 8-digit mod-11.
	"CZ00177040",
	// Denmark — mod-11.
	"DK25313764",
	// Estonia — mod-10.
	"EE100094917",
	// Greece — mod-11.
	"EL094019246",
	// Finland — mod-11.
	"FI09853609",
	// France — mod-97 SIREN key.
	"FR41303265045",
	// Croatia — ISO 7064 MOD 11-10.
	"HR38192148119",
	// Hungary — exactly 8 digits, not 9.
	"HU123456789",
	"HU12892313", // valid format, mod-10 control broken
	// Italy — Luhn.
	"IT00488410011",
	// Lithuania — first-pass and fallback both fail.
	"LT119511516",
	"LT290061371315",
	// Luxembourg — mod-89.
	"LU10000357",
	// Latvia — legal entity mod-11.
	"LV40003521601",
	// Malta — mod-37.
	"MT15121334",
	// Netherlands — format only (no checksum); shape must still hold.
	"NL005033018B0", // branch suffix must be exactly 2 digits
	// Poland — mod-11.
	"PL5260250275",
	// Portugal — mod-11.
	"PT502757192",
	// Romania — mod-11 over left-padded body.
	"RO18158684",
	// Sweden — last two digits must be "01"; Luhn over the org number.
	"SE123456789002", // ends in 02
	"SE123456789010", // ends in 10
	"SE12345678901",  // only 11 digits
	"SE556677889001", // valid format, Luhn fails (org digit moved)
	// Slovenia — mod-11.
	"SI80267433",
	// Slovakia — full number divisible by 11.
	"SK2020317069",
	// Switzerland — no alternative dotted format after normalization.
	"CHE12345678",   // 8 digits
	"CHE1234567890", // 10 digits
}

func Test_NormalizeVATID(t *testing.T) {
	for testID, refID := range validVATIDs {
		result, err := NormalizeVATID(testID)
		if err != nil {
			t.Errorf("NormalizeVATID(%s): %s", string(testID), err)
		} else if string(result) != refID {
			t.Errorf("NormalizeVATID(%s): %s != %s", string(testID), string(result), refID)
		}
	}
}

func Test_VATIDValid(t *testing.T) {
	for _, invalidID := range invalidVATIDs {
		assert.Falsef(t, invalidID.Valid(), "vat.ID should be invalid: %s, Regex uses NormalizedUnchecked: %s", string(invalidID), string(invalidID /*.NormalizedUnchecked()*/))
	}
}

var vatidTestIndices = map[string][][]int{
	"":                         nil,
	"ATU10223006":              {{0, 11}},
	"  ATU10223006":            {{2, 13}},
	"UID: ATU10223006":         {{5, 16}},
	"UID AT U 10223006":        {{4, 17}},
	"UID:AT U 10223006":        {{4, 17}},
	"ATU10223006 ":             {{0, 11}},
	"ATU10223006 ATU 10223006": {{0, 11}, {12, 24}},
	" AT U 10223006 ATU10223006 ATU 10223006 ": {{1, 14}, {15, 26}, {27, 39}},
	"USt-IdNr. DE 136725570":                   {{10, 22}},
}

func Test_VATIDFinder_FindAllIndex(t *testing.T) {
	for str, refIndices := range vatidTestIndices {
		indices := IDFinder.FindAllIndex([]byte(str), -1)
		if len(indices) != len(refIndices) {
			var words []string
			for i := range indices {
				words = append(words, "'"+str[indices[i][0]:indices[i][1]]+"'")
			}
			t.Errorf("VATIDFinder.FindAllIndex('%s') expected %d words but got %d: %s", str, len(refIndices), len(indices), strings.Join(words, " "))
		} else {
			for i := range indices {
				if indices[i][0] != refIndices[i][0] || indices[i][1] != refIndices[i][1] {
					// t.Error(i, indices[i], refIndices[i], len(str))
					t.Errorf("VATIDFinder.FindAllIndex('%s') word %d expected %v '%s' but got %v '%s'", str, i, refIndices[i], str[refIndices[i][0]:refIndices[i][1]], indices[i], str[indices[i][0]:indices[i][1]])
				}
			}
		}
	}
}

// Every rejection must wrap ErrInvalidID so callers can detect invalid input
// with errors.Is, and the message must name the gate that rejected the input
// so users can tell a typo in the country prefix from a wrong check digit.
func TestID_Normalized_rejectionReasons(t *testing.T) {
	tests := []struct {
		name       string
		id         ID
		wantReason string
	}{
		{name: "empty", id: "", wantReason: "too short"},
		{name: "punctuation only counts as empty", id: " - . ", wantReason: "too short"},
		{name: "3 chars", id: "ATU", wantReason: "too short"},
		// The length limit applies after removing spaces and punctuation.
		{name: "17 chars", id: "ATU10223006123456", wantReason: "too long"},
		{name: "unknown country prefix", id: "XX123456789", wantReason: "invalid country code"},
		{name: "country without VAT format", id: "US123456789", wantReason: "unsupported country code"},
		// Greek VAT IDs use the EU prefix EL, not the ISO code GR.
		{name: "Greece with ISO prefix", id: "GR094019245", wantReason: "unsupported country code"},
		{name: "16 chars hit the format check", id: "ATU1022300612345", wantReason: "invalid format"},
		{name: "AT without U", id: "AT10223006", wantReason: "invalid format"},
		{name: "DE starting with 0", id: "DE011111125", wantReason: "invalid format"},
		{name: "AT check digit", id: "ATU10223007", wantReason: "invalid check-sum"},
		{name: "DE check digit", id: "DE111111126", wantReason: "invalid check-sum"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.id.Normalized()
			require.ErrorIs(t, err, ErrInvalidID)
			assert.ErrorContains(t, err, tt.wantReason)

			// Validate and NormalizeVATID share the same rules
			assert.ErrorIs(t, tt.id.Validate(), ErrInvalidID)
			_, err = NormalizeVATID(string(tt.id))
			assert.ErrorIs(t, err, ErrInvalidID)
		})
	}
}

// Length limits are checked after spaces and punctuation are removed,
// so heavily formatted input that is longer than IDMaxLength must still
// be accepted when its normalized form is valid.
func TestID_Normalized_lengthIsCheckedAfterNormalization(t *testing.T) {
	raw := ID("AT - U - 1022 - 3006")
	require.Greater(t, len(raw), IDMaxLength)

	norm, err := raw.Normalized()
	require.NoError(t, err)
	assert.Equal(t, ID("ATU10223006"), norm)
}

func TestStringIsVATID(t *testing.T) {
	// StringIsVATID and BytesAreVATID are the cheap predicates used by
	// finders and must agree with ID.Valid for both valid and invalid input.
	for in := range validVATIDs {
		assert.Truef(t, StringIsVATID(in), "StringIsVATID(%q)", in)
		assert.Truef(t, BytesAreVATID([]byte(in)), "BytesAreVATID(%q)", in)
	}
	for _, in := range slices.Concat(invalidVATIDs, []ID{"", "DE"}) {
		assert.Falsef(t, StringIsVATID(string(in)), "StringIsVATID(%q)", string(in))
		assert.Falsef(t, BytesAreVATID([]byte(in)), "BytesAreVATID(%q)", string(in))
	}
}

func TestID_ValidAndNormalized(t *testing.T) {
	tests := []struct {
		id   ID
		want bool
	}{
		{id: "ATU10223006", want: true},
		{id: "atu10223006", want: false},  // valid, but lower case
		{id: "ATU 10223006", want: false}, // valid, but contains a space
		{id: "ATU12345678", want: false},  // invalid check-sum
		{id: "", want: false},
	}
	for _, tt := range tests {
		t.Run(string(tt.id), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.id.ValidAndNormalized())
		})
	}
}

// ValidateIsNormalized must tell "invalid" (ErrInvalidID) apart from
// "valid but not normalized" (some other error), as documented,
// so callers can decide whether normalizing would fix the value.
func TestID_ValidateIsNormalized(t *testing.T) {
	assert.NoError(t, ID("ATU10223006").ValidateIsNormalized())

	err := ID("atu 10223006").ValidateIsNormalized()
	require.Error(t, err, "valid but not normalized")
	assert.NotErrorIs(t, err, ErrInvalidID, "a valid ID must not be reported as invalid")

	assert.ErrorIs(t, ID("ATU12345678").ValidateIsNormalized(), ErrInvalidID)
}

// A valid but sloppily formatted ID is normalized on the way to NullableID,
// while anything invalid becomes NULL instead of being stored as garbage.
func TestID_NormalizedOrNull(t *testing.T) {
	assert.Equal(t, NullableID("ATU10223006"), ID("at u 1022 3006").NormalizedOrNull())
	assert.Equal(t, Null, ID("ATU12345678").NormalizedOrNull())
	assert.Equal(t, Null, ID("").NormalizedOrNull())
}

// Nullable is a plain conversion: it must neither validate nor normalize,
// so no information is lost when switching between ID and NullableID.
func TestID_Nullable(t *testing.T) {
	assert.Equal(t, NullableID("atu 10223006"), ID("atu 10223006").Nullable())
	assert.Equal(t, NullableID("not a VAT ID"), ID("not a VAT ID").Nullable())
	assert.Equal(t, Null, ID("").Nullable())
}

func TestID_CountryCode(t *testing.T) {
	tests := []struct {
		name string
		id   ID
		want country.Code
	}{
		{name: "Austria", id: "ATU10223006", want: country.AT},
		{name: "normalizes first", id: "atu 10223006", want: country.AT},
		{name: "Germany", id: "DE167015661", want: country.DE},
		// "EU" is not a country, MOSS IDs are mapped to Belgium (Brussels).
		{name: "MOSS", id: "EU372008134", want: country.BE},
		// "XI" is not an ISO 3166-1 code, Northern Ireland belongs to the UK.
		{name: "Northern Ireland", id: "XI123456782", want: country.GB},
		{name: "invalid", id: "ATU12345678", want: country.Invalid},
		{name: "empty", id: "", want: country.Invalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.id.CountryCode()
			assert.Equal(t, tt.want, got)
			if got != country.Invalid {
				assert.True(t, got.Valid(), "CountryCode must return a valid ISO 3166-1 code")
			}
		})
	}
}

func TestID_IsMOSS(t *testing.T) {
	assert.True(t, ID("EU372008134").IsMOSS())
	assert.True(t, ID("eu 372 008 134").IsMOSS(), "IsMOSS normalizes first")
	assert.False(t, ID("ATU10223006").IsMOSS())
	assert.False(t, ID("EU37200813").IsMOSS(), "invalid MOSS ID")
	assert.False(t, ID("").IsMOSS())
}

func TestID_Number(t *testing.T) {
	tests := []struct {
		id   ID
		want string
	}{
		// The number is everything after the 2-letter country prefix.
		// For Austria this includes the "U", which is how VIES expects it.
		{id: "ATU10223006", want: "U10223006"},
		{id: "DE 167 015 661", want: "167015661"},
		{id: "EU372008134", want: "372008134"},
		{id: "NO 916634773 MVA", want: "916634773MVA"},
		// Invalid IDs have no meaningful number part.
		{id: "ATU12345678", want: ""},
		{id: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(string(tt.id), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.id.Number())
		})
	}
}

// String must not lose information: valid IDs are shown normalized,
// invalid input is shown exactly as given so it can be corrected.
func TestID_String(t *testing.T) {
	assert.Equal(t, "ATU10223006", ID("at u 1022 3006").String())
	assert.Equal(t, "atu 1234 5678", ID("atu 1234 5678").String())
	assert.Equal(t, "", ID("").String())
}

func TestID_ScanString(t *testing.T) {
	t.Run("valid input is stored normalized", func(t *testing.T) {
		for _, validate := range []bool{true, false} {
			var id ID
			require.NoError(t, id.ScanString("de 167-015-661", validate))
			assert.Equal(t, ID("DE167015661"), id)
		}
	})
	t.Run("validate rejects invalid input and keeps the old value", func(t *testing.T) {
		id := ID("ATU10223006")
		err := id.ScanString("ATU12345678", true)
		assert.ErrorIs(t, err, ErrInvalidID)
		assert.Equal(t, ID("ATU10223006"), id)
	})
	t.Run("without validate invalid input is stored as is", func(t *testing.T) {
		var id ID
		require.NoError(t, id.ScanString("atu 1234 5678", false))
		assert.Equal(t, ID("atu 1234 5678"), id)
	})
}

func TestID_Scan(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		var id ID
		require.NoError(t, id.Scan("ATU10223006"))
		assert.Equal(t, ID("ATU10223006"), id)
	})
	t.Run("bytes are copied", func(t *testing.T) {
		// database/sql drivers may reuse the []byte buffer after Scan,
		// so the ID must not alias it.
		buf := []byte("ATU10223006")
		var id ID
		require.NoError(t, id.Scan(buf))
		copy(buf, "XXXXXXXXXXX")
		assert.Equal(t, ID("ATU10223006"), id)
	})
	t.Run("NULL is an error", func(t *testing.T) {
		// ID is not nullable, NULL must not silently become an empty ID.
		// Use NullableID for nullable columns.
		id := ID("ATU10223006")
		assert.ErrorIs(t, id.Scan(nil), ErrInvalidID)
	})
	t.Run("unsupported type", func(t *testing.T) {
		var id ID
		assert.ErrorIs(t, id.Scan(int64(10223006)), ErrInvalidID)
	})
}

// Valid IDs are written to the database in normalized form
// so equal VAT IDs compare equal in SQL.
func TestID_Value(t *testing.T) {
	v, err := ID("at u 1022 3006").Value()
	require.NoError(t, err)
	assert.Equal(t, "ATU10223006", v)

	v, err = ID("ATU10223006").Value()
	require.NoError(t, err)
	assert.Equal(t, "ATU10223006", v)
}

// The JSON schema limits must stay in sync with the length limits
// that Normalized enforces, otherwise API validation and Go validation
// would disagree.
func TestID_JSONSchema(t *testing.T) {
	s := ID("").JSONSchema()
	assert.Equal(t, "string", s.Type)
	require.NotNil(t, s.MinLength)
	require.NotNil(t, s.MaxLength)
	assert.Equal(t, uint64(IDMinLength), *s.MinLength)
	assert.Equal(t, uint64(IDMaxLength), *s.MaxLength)
}
