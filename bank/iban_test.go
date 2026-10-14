package bank

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/domonda/go-types/country"
	"github.com/domonda/go-types/internal/normtest"
)

var toNormalizeIBANs = map[string]IBAN{
	// // 2 parts
	"AT74 2011182151456500": "AT742011182151456500",
	"AT 742011182151456500": "AT742011182151456500",
	// 5 parts
	"AT05 1937 0711 1000 0044": "AT051937071110000044",
	// ":AT05 1937 0711 1000 0044.": "AT051937071110000044",
	"FO97 5432 0388 8999 44": "FO9754320388899944",
	// 6 parts
	"DE36 7015 0000 0081 1342 98": "DE36701500000081134298",
	"DE 3670 1500 0000 8113 4298": "DE36701500000081134298",
	"AT 82 38487 000 000 27185":   "AT823848700000027185",
	"LV80 BANK 0000 4351 9500 1":  "LV80BANK0000435195001",
	// 7 parts
	"DE 36 7015 0000 0081 1342 98": "DE36701500000081134298",
	// 8 parts
	"MU17 BOMM 0101 1010 3030 0200 000M UR":  "MU17BOMM0101101030300200000MUR",
	"MT84 MALT 0110 0001 2345 MTLC AST0 01S": "MT84MALT011000012345MTLCAST001S",
	"QA58 DOHB 0000 1234 5678 90AB CDEF G":   "QA58DOHB00001234567890ABCDEFG",
}

var invalidIBANs = []string{
	"at05 1937 0711 1000 0044",
}

// http://www.rbs.co.uk/corporate/international/g0/guide-to-international-business/regulatory-information/iban/iban-example.ashx
var countryIBANTable = []string{
	"Albania", "AL47 2121 1009 0000 0002 3569 8741",
	"Andorra", "AD12 0001 2030 2003 5910 0100",
	"Austria", "AT61 1904 3002 3457 3201",
	"Azerbaijan", "AZ21 NABZ 0000 0000 1370 1000 1944",
	"Bahrain", "BH67 BMAG 0000 1299 1234 56",
	"Belgium", "BE62 5100 0754 7061",
	"Bosniaand Herzegovina", "BA39 1290 0794 0102 8494",
	"Bulgaria", "BG80 BNBG 9661 1020 3456 78",
	"Croatia", "HR12 1001 0051 8630 0016 0",
	"Cyprus", "CY17 0020 0128 0000 0012 0052 7600",
	"Czech Republic", "CZ65 0800 0000 1920 0014 5399",
	"Denmark", "DK50 0040 0440 1162 43",
	"Estonia", "EE38 2200 2210 2014 5685",
	"Faroe Islands", "FO97 5432 0388 8999 44",
	"Finland", "FI21 1234 5600 0007 85",
	"France", "FR14 2004 1010 0505 0001 3M02 606",
	"Georgia", "GE29 NB00 0000 0101 9049 17",
	"Germany", "DE89 3704 0044 0532 0130 00",
	"Gibraltar", "GI75 NWBK 0000 0000 7099 453",
	"Greece", "GR16 0110 1250 0000 0001 2300 695",
	"Greenland", "GL56 0444 9876 5432 10",
	"Hungary", "HU42 1177 3016 1111 1018 0000 0000",
	"Iceland", "IS14 0159 2600 7654 5510 7303 39",
	"Ireland", "IE29 AIBK 9311 5212 3456 78",
	"Israel", "IL62 0108 0000 0009 9999 999",
	"Italy", "IT40 S054 2811 1010 0000 0123 456",
	"Jordan", "JO94 CBJO 0010 0000 0000 0131 0003 02",
	"Kuwait", "KW81 CBKU 0000 0000 0000 1234 5601 01",
	"Latvia", "LV80 BANK 0000 4351 9500 1",
	"Lebanon", "LB62 0999 0000 0001 0019 0122 9114",
	"Liechtenstein", "LI21 0881 0000 2324 013A A",
	"Lithuania", "LT12 1000 0111 0100 1000",
	"Luxembourg", "LU28 0019 4006 4475 0000",
	"Macedonia", "MK07 2501 2000 0058 984",
	"Malta", "MT84 MALT 0110 0001 2345 MTLC AST0 01S",
	"Mauritius", "MU17 BOMM 0101 1010 3030 0200 000M UR",
	"Moldova", "MD24 AG00 0225 1000 1310 4168",
	"Monaco", "MC93 2005 2222 1001 1223 3M44 555",
	"Montenegro", "ME25 5050 0001 2345 6789 51",
	"Netherlands", "NL39 RABO 0300 0652 64",
	"Norway", "NO93 8601 1117 947",
	"Pakistan", "PK36 SCBL 0000 0011 2345 6702",
	"Poland", "PL60 1020 1026 0000 0422 7020 1111",
	"Portugal", "PT50 0002 0123 1234 5678 9015 4",
	"Qatar", "QA58 DOHB 0000 1234 5678 90AB CDEF G",
	"Romania", "RO49 AAAA 1B31 0075 9384 0000",
	"San Marino", "SM86 U032 2509 8000 0000 0270 100",
	"Saudi Arabia", "SA03 8000 0000 6080 1016 7519",
	"Serbia", "RS35 2600 0560 1001 6113 79",
	"Slovak Republic", "SK31 1200 0000 1987 4263 7541",
	"Slovenia", "SI56 1910 0000 0123 438",
	"Spain", "ES80 2310 0001 1800 0001 2345",
	"Sweden", "SE35 5000 0000 0549 1000 0003",
	"Switzerland", "CH93 0076 2011 6238 5295 7",
	"Tunisia", "TN59 1000 6035 1835 9847 8831",
	"Turkey", "TR33 0006 1005 1978 6457 8413 26",
	"UAE", "AE07 0331 2345 6789 0123 456",
	"Egypt", "EG81 0037 0061 0840 2031 8813 7204 3",
	// "United Kingdom", "GB29 RBOS 6016 1331 9268 19", // invalid checksum
}

func Test_NormalizeIBAN(t *testing.T) {
	for testIBAN, expectedIBAN := range toNormalizeIBANs {
		normalized, err := NormalizeIBAN(testIBAN)
		if err != nil {
			t.Errorf("NormalizeIBAN(%s): %s", string(testIBAN), err.Error())
			continue
		}
		if normalized != expectedIBAN {
			t.Errorf("NormalizeIBAN(%s): %s != %s", string(testIBAN), string(normalized), string(expectedIBAN))
		}
	}

	for _, invalidIBAN := range invalidIBANs {
		normalized, err := NormalizeIBAN(invalidIBAN)
		if err == nil {
			t.Errorf("Should NOT be valid NormalizeIBAN(%s): %s", string(invalidIBAN), string(normalized))
			require.Equal(t, invalidIBAN, normalized, "Normalized invalid IBAN returned unchanged")
		}
	}

	for i := 0; i < len(countryIBANTable); i += 2 {
		country, iban := countryIBANTable[i], countryIBANTable[i+1]
		normalized, err := IBAN(iban).NormalizedWithSpaces()
		if err != nil {
			t.Errorf("NormalizedSpacedOrErr(%s): %s [%s]", string(iban), err.Error(), string(country))
			continue
		}
		if string(normalized) != iban {
			t.Errorf("NormalizedSpacedOrErr(%s): %s != %s [%s]", string(iban), string(normalized), string(iban), string(country))
		}
	}
}

var bankAndAccountNumbersTable = map[IBAN][2]string{
	"AT252011183728861100": {"20111", "83728861100"},
}

func Test_IBAN_BankAndAccountNumbers(t *testing.T) {
	for iban, numbers := range bankAndAccountNumbersTable {
		bankNr, accountNr, err := iban.BankAndAccountNumbers()
		if err != nil {
			t.Errorf("error from IBAN.BankAndAccountNumbers(%s): %s", string(iban), err)
			continue
		}
		if bankNr != numbers[0] || accountNr != numbers[1] {
			t.Errorf("IBAN.BankAndAccountNumbers(%s): not correct", string(iban))
		}
	}
}

func TestIBAN_Normalized_Rejects(t *testing.T) {
	// Each case must be rejected by the rule named in wantErr, not by an
	// earlier check, to make sure every validation step is effective.
	tests := []struct {
		name    string
		iban    IBAN
		wantErr string
	}{
		{name: "empty", iban: "", wantErr: "empty"},
		{name: "only whitespace", iban: " \t\n", wantErr: "empty"},
		{name: "shorter than shortest country (NO)", iban: "NO938601111794", wantErr: "too short"},
		{name: "unknown country", iban: "XX89370400440532013000", wantErr: "country code"},
		{name: "lowercase country", iban: "de89370400440532013000", wantErr: "country code"},
		{name: "one digit missing", iban: "DE89 3704 0044 0532 0130 0", wantErr: "length"},
		{name: "one digit too much", iban: "DE89 3704 0044 0532 0130 000", wantErr: "length"},
		{name: "dashes are not removed", iban: "DE89-3704-0044-0532-0130-00", wantErr: "length"},
		{name: "non numeric check digits", iban: "DEAB370400440532013000", wantErr: "characters"},
		{name: "lowercase BBAN", iban: "GB29nwbk60161331926819", wantErr: "characters"},
		{name: "punctuation", iban: "DE8937040044053201300!", wantErr: "characters"},
		{name: "wrong last digit", iban: "DE89370400440532013001", wantErr: "check sum"},
		{name: "swapped check digits", iban: "DE98370400440532013000", wantErr: "check sum"},
		{name: "swapped BBAN digits", iban: "DE89370400440532010300", wantErr: "check sum"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			norm, err := tt.iban.Normalized()
			require.ErrorContains(t, err, tt.wantErr)
			assert.Equal(t, tt.iban, norm, "invalid IBAN must be returned unchanged")

			assert.False(t, tt.iban.Valid())
			assert.Error(t, tt.iban.Validate())
			assert.False(t, StringIsIBAN(string(tt.iban)))
			assert.False(t, tt.iban.ValidAndNormalized())
			assert.Equal(t, IBANNull, tt.iban.NormalizedOrNull())
			if strings.TrimSpace(string(tt.iban)) == "" {
				assert.Equal(t, "", marshalText(t, tt.iban), "MarshalText of a blank IBAN")
			} else {
				assert.Equal(t, string(tt.iban), marshalText(t, tt.iban), "MarshalText of an invalid IBAN must write it unchanged")
			}

			spaced, err := tt.iban.NormalizedWithSpaces()
			assert.Error(t, err)
			assert.Equal(t, tt.iban, spaced)
		})
	}
}

func TestIBAN_Normalized_Whitespace(t *testing.T) {
	// IBANs extracted from PDFs and e-mails contain all kinds of whitespace
	const want IBAN = "DE89370400440532013000"
	for _, iban := range []IBAN{
		"DE89370400440532013000",
		"DE89 3704 0044 0532 0130 00",
		"DE89 3704 0044 0532 0130 00 ",
		"DE89\t3704\n0044\r\n0532 0130 00",
		"DE89 3704 0044 0532 0130 00", // non-breaking space
		"DE89​3704​0044​0532​0130​00", // zero width space
		// Leading whitespace must be accepted like trailing whitespace,
		// the country code is checked after removing the whitespace.
		" DE89370400440532013000",
		"\tDE89 3704 0044 0532 0130 00\n",
		" DE89 3704 0044 0532 0130 00", // leading non-breaking space
	} {
		t.Run(string(iban), func(t *testing.T) {
			norm, err := iban.Normalized()
			require.NoError(t, err)
			assert.Equal(t, want, norm)
			assert.True(t, iban.Valid())
			assert.True(t, StringIsIBAN(string(iban)))
			assert.Equal(t, string(want), marshalText(t, iban))
			assert.Equal(t, NullableIBAN(want), iban.NormalizedOrNull())
			assert.Equal(t, iban == want, iban.ValidAndNormalized(), "only the form without whitespace is normalized")
		})
	}
}

func TestIBAN_NormalizedWithSpaces(t *testing.T) {
	// Groups of 4 with a shorter last group, without trailing space
	spaced, err := IBAN("NO9386011117947").NormalizedWithSpaces()
	require.NoError(t, err)
	assert.Equal(t, IBAN("NO93 8601 1117 947"), spaced)

	// Length divisible by 4 must not get a trailing space
	spaced, err = IBAN("AT611904300234573201").NormalizedWithSpaces()
	require.NoError(t, err)
	assert.Equal(t, IBAN("AT61 1904 3002 3457 3201"), spaced)

	// Spaced IBAN is re-grouped
	spaced, err = IBAN("AT 61 19043 002 34573201").NormalizedWithSpaces()
	require.NoError(t, err)
	assert.Equal(t, IBAN("AT61 1904 3002 3457 3201"), spaced)
}

func TestIBAN_CountryCode(t *testing.T) {
	assert.Equal(t, country.DE, IBAN("DE89 3704 0044 0532 0130 00").CountryCode())
	assert.Equal(t, country.NO, IBAN("NO9386011117947").CountryCode())
	assert.Equal(t, country.Invalid, IBAN("DE89370400440532013001").CountryCode(), "invalid check sum")
	assert.Equal(t, country.Invalid, IBAN("").CountryCode())
}

// The JSON schema published for IBAN must not reject IBANs
// that are valid in normalized form in Go.
func TestIBAN_JSONSchema(t *testing.T) {
	s := IBAN("").JSONSchema()
	assert.Equal(t, "string", s.Type)
	require.Equal(t, IBANRegex, s.Pattern)
	re := regexp.MustCompile(s.Pattern)
	for i := 1; i < len(countryIBANTable); i += 2 {
		norm, err := NormalizeIBAN(countryIBANTable[i])
		require.NoError(t, err)
		assert.Regexp(t, re, string(norm))
	}
}

func TestIBAN_ScanString(t *testing.T) {
	t.Run("normalizes valid IBAN", func(t *testing.T) {
		for _, validate := range []bool{true, false} {
			var iban IBAN
			require.NoError(t, iban.ScanString("DE89 3704 0044 0532 0130 00", validate))
			assert.Equal(t, IBAN("DE89370400440532013000"), iban)
		}
	})
	t.Run("validate rejects invalid IBAN", func(t *testing.T) {
		iban := IBAN("unchanged")
		require.Error(t, iban.ScanString("DE89370400440532013001", true))
		assert.Equal(t, IBAN("unchanged"), iban)
	})
	t.Run("without validate assigns invalid IBAN as is", func(t *testing.T) {
		var iban IBAN
		require.NoError(t, iban.ScanString("DE89 3704 0044 0532 0130 01", false))
		assert.Equal(t, IBAN("DE89 3704 0044 0532 0130 01"), iban)
	})
}

func TestIBAN_SQL(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    IBAN
		wantErr bool
	}{
		{name: "string", value: "DE89370400440532013000", want: "DE89370400440532013000"},
		{name: "bytes", value: []byte("DE89370400440532013000"), want: "DE89370400440532013000"},
		{name: "NULL", value: nil, want: ""},
		// Scan does not validate, so existing invalid data can still be read
		{name: "invalid string", value: "invalid", want: "invalid"},
		{name: "unsupported type", value: int64(1), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iban := IBAN("unchanged")
			err := iban.Scan(tt.value)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, IBAN("unchanged"), iban)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, iban)
		})
	}

	t.Run("Scan copies bytes", func(t *testing.T) {
		// database/sql may reuse the []byte buffer after Scan returns
		buf := []byte("DE89370400440532013000")
		var iban IBAN
		require.NoError(t, iban.Scan(buf))
		copy(buf, "XXXXXXXXXXXXXXXXXXXXXX")
		assert.Equal(t, IBAN("DE89370400440532013000"), iban)
	})

	t.Run("Value round trip", func(t *testing.T) {
		v, err := IBAN("DE89370400440532013000").Value()
		require.NoError(t, err)
		assert.Equal(t, "DE89370400440532013000", v)
		var iban IBAN
		require.NoError(t, iban.Scan(v))
		assert.Equal(t, IBAN("DE89370400440532013000"), iban)
	})
}

// Value and MarshalJSON write an IBAN normalized, because Validate accepts
// the IBAN grouped with spaces as printed on statements: without
// normalization the grouped and the compact IBAN would reach SQL and JSON
// as different strings. A IBAN that can't be normalized is refused by Value, because the
// database is where an invalid value would stay, and written unchanged
// by MarshalJSON, so a response that holds one still renders.
func TestIBAN_ValueAndMarshalJSONNormalize(t *testing.T) {
	tests := []struct {
		iban IBAN
		want string
	}{
		{iban: "DE89370400440532013000", want: "DE89370400440532013000"},
		{iban: "DE89 3704 0044 0532 0130 00", want: "DE89370400440532013000"},
		{iban: " AT61 1904 3002 3457 3201 ", want: "AT611904300234573201"},
	}
	for _, tt := range tests {
		v, err := tt.iban.Value()
		require.NoError(t, err)
		assert.Equal(t, tt.want, v, "IBAN(%q).Value()", string(tt.iban))

		j, err := json.Marshal(tt.iban)
		require.NoError(t, err)
		assert.Equal(t, `"`+tt.want+`"`, string(j), "json.Marshal(IBAN(%q))", string(tt.iban))
	}

	_, err := IBAN("DE00 1234").Value()
	assert.Error(t, err, "Value() of an invalid IBAN")
	j, err := json.Marshal(IBAN("DE00 1234"))
	require.NoError(t, err)
	assert.Equal(t, `"DE00 1234"`, string(j))

	v, err := IBAN("").Value()
	require.NoError(t, err)
	assert.Nil(t, v, "an empty IBAN is written as SQL NULL")
}

func TestIBAN_BankAndAccountNumbers_Countries(t *testing.T) {
	tests := []struct {
		iban          IBAN
		wantBankNo    string
		wantAccountNo string
	}{
		{"DE89370400440532013000", "37040044", "0532013000"},
		{"CH9300762011623852957", "00762", "011623852957"},
		{"AT611904300234573201", "19043", "00234573201"},
		// IBANs are usually written in groups of 4 and are valid with
		// whitespace, so the numbers must be taken from the normalized
		// IBAN. Slicing the raw string returns wrong numbers without error.
		{"DE89 3704 0044 0532 0130 00", "37040044", "0532013000"},
		{" CH93 0076 2011 6238 5295 7", "00762", "011623852957"},
		{"AT61 1904 3002 3457 3201\n", "19043", "00234573201"},
	}
	for _, tt := range tests {
		t.Run(string(tt.iban), func(t *testing.T) {
			bankNo, accountNo, err := tt.iban.BankAndAccountNumbers()
			require.NoError(t, err)
			assert.Equal(t, tt.wantBankNo, bankNo)
			assert.Equal(t, tt.wantAccountNo, accountNo)
		})
	}

	t.Run("invalid IBAN", func(t *testing.T) {
		iban := IBAN("DE89370400440532013001")
		_, _, err := iban.BankAndAccountNumbers()
		require.ErrorContains(t, err, "invalid IBAN")
	})
	t.Run("unsupported country", func(t *testing.T) {
		iban := IBAN("FR1420041010050500013M02606")
		require.True(t, iban.Valid())
		_, _, err := iban.BankAndAccountNumbers()
		require.ErrorContains(t, err, "can't extract")
	})
}

func TestIBAN_isCheckSumValid(t *testing.T) {
	assert.True(t, IBAN("DE89370400440532013000").isCheckSumValid())
	assert.False(t, IBAN("DE89370400440532013001").isCheckSumValid())
	assert.False(t, IBAN("DE89").isCheckSumValid(), "too short must not panic")
	// Characters outside of A-Z and 0-9 have no checksum value
	// and must not be mapped to some arbitrary number
	assert.False(t, IBAN("de89370400440532013000").isCheckSumValid())
}

// TestIBAN_ReadingAndWriting checks the reading and writing contract shared
// by the normalizing types: every reading path stores a valid IBAN
// normalized and an invalid one unchanged, Value refuses an invalid IBAN,
// and fmt prints the stored IBAN.
func TestIBAN_ReadingAndWriting(t *testing.T) {
	normtest.Run(t, normtest.Spec[IBAN]{
		Canonical: "DE89370400440532013000",
		Variants:  []IBAN{"DE89 3704 0044 0532 0130 00", " DE89370400440532013000\n"},
		Invalid:   []IBAN{"DE00 1234", "de89370400440532013000"},
	})
}
