package money

import (
	"fmt"
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var currencyTestTable = []string{
	"€", "EUR",
	"EUR", "EUR",
	"eur", "EUR",
	"Euro", "EUR",
	"euro", "EUR",
	"EURO", "EUR",
	" EURO ", "EUR",

	"$", "USD",
	"USD", "USD",
	"usd", "USD",
	"US Dollar", "USD",
	"US$", "USD",
	"dollar", "USD",

	"£", "GBP",
	"GB£", "GBP",
	"Pound Sterling", "GBP",

	"₣", "CHF",
	"Swiss Franc", "CHF",
}

func Test_NormalizeCurrency(t *testing.T) {
	for i := 0; i < len(currencyTestTable); i += 2 {
		testVal, refVal := currencyTestTable[i], currencyTestTable[i+1]
		result, err := NormalizeCurrency(testVal)
		if err != nil {
			t.Errorf("NormalizeCurrency(%s): %s", testVal, err)
		}
		if string(result) != refVal {
			t.Errorf("NormalizeCurrency(%s): %s != %s", testVal, string(result), refVal)
		}
	}
}

func Test_NullCurrency(t *testing.T) {
	assert.False(t, Currency("").Valid())
	assert.True(t, NullableCurrency("").Valid())
}

// TestCurrency_EnumsMatchCurrencyCodeToName ties the two lists of currency
// codes together: Enums() is generated from the constants in currency.go, but
// Valid() resolves through currencyCodeToName in constants.go. A code added to
// only one of them would make Enums() list a code that Valid() rejects, or
// miss one that it accepts.
func TestCurrency_EnumsMatchCurrencyCodeToName(t *testing.T) {
	assert.ElementsMatch(t, slices.Collect(maps.Keys(currencyCodeToName)), Currency("").Enums())
}

// TestCurrency_CustomMethodsAcceptNonNormalizedInput pins why Valid, Validate
// and String carry a #custom marker: they resolve through Normalized and
// therefore accept currency symbols, English names and lower case, while the
// exact-match switch over the enum constants that go-enum generates instead
// would only accept the 3-letter ISO 4217 codes and String would stop
// normalizing. Test_NormalizeCurrency exercises the same inputs through
// NormalizeCurrency only, so a regenerated Valid/Validate/String would not
// fail it.
//
// Value: protects=money.Currency's #custom Valid/Validate/String keep
// accepting symbols, English names and lower case through Normalized;
// fails_when=a go-enum run replaces them with the generated exact-match switch
// and plain string conversion, so Currency("eur") and Currency("€") turn
// invalid;
// why_new=Test_NormalizeCurrency only calls NormalizeCurrency, and the only
// Valid assertion in this package is for the empty Currency;
// seam=none
func TestCurrency_CustomMethodsAcceptNonNormalizedInput(t *testing.T) {
	t.Run("accepted non-normalized input", func(t *testing.T) {
		tests := []struct {
			in   Currency
			want string
		}{
			{in: "EUR", want: "EUR"}, // already normalized
			{in: "eur", want: "EUR"}, // lower case
			{in: "€", want: "EUR"},   // symbol
			{in: "Euro", want: "EUR"},
			{in: " EURO ", want: "EUR"},
			{in: "US$", want: "USD"},
			{in: "dollar", want: "USD"}, // English name
			{in: "Pound Sterling", want: "GBP"},
			{in: "₣", want: "CHF"},
		}
		for _, tt := range tests {
			t.Run(string(tt.in), func(t *testing.T) {
				assert.True(t, tt.in.Valid(), "Currency(%q).Valid()", string(tt.in))
				assert.NoError(t, tt.in.Validate(), "Currency(%q).Validate()", string(tt.in))
				assert.Equal(t, tt.want, tt.in.String(), "Currency(%q).String() must normalize", string(tt.in))
			})
		}
	})

	t.Run("rejected input is returned unchanged", func(t *testing.T) {
		// String has no error result, so an unresolvable currency has to come
		// back unchanged instead of being replaced by the empty string.
		for _, in := range []Currency{"", "xxx", "Atlantean Shell"} {
			t.Run(string(in), func(t *testing.T) {
				assert.False(t, in.Valid(), "Currency(%q).Valid()", string(in))
				assert.Error(t, in.Validate(), "Currency(%q).Validate()", string(in))
				assert.Equal(t, string(in), in.String(), "Currency(%q).String() must return the input unchanged", string(in))
			})
		}
	})
}

// TestCurrency_SymbolTableAndEnumStrings covers the invariant that the removal
// of money.BTC and the ₿ symbol restored: a symbol must only ever normalize to
// a currency that is itself an enumerated, valid ISO 4217 code. ₿ mapped to
// BTC, which was never in currencyCodeToName, so Currency("₿").Valid() was
// true while the Currency it normalized to was invalid.
//
// It also pins the generated EnumStrings, which no other test looks at, and
// the exclusion of the untyped CurrencyNull from both enum value lists.
//
// Value: protects=every currencySymbolToCode / currencyCodeToSymbol entry
// resolves to an enumerated valid Currency, and EnumStrings mirrors Enums
// without CurrencyNull;
// fails_when=a non-ISO-4217 symbol mapping such as ₿ -> BTC is added back, or
// the two generated enum lists drift apart;
// why_new=TestCurrency_EnumsMatchCurrencyCodeToName compares Enums to
// currencyCodeToName only and never follows the symbol tables or EnumStrings;
// seam=none
func TestCurrency_SymbolTableAndEnumStrings(t *testing.T) {
	enums := Currency("").Enums()

	t.Run("symbols normalize to enumerated currencies", func(t *testing.T) {
		for symbol, code := range currencySymbolToCode {
			norm, err := Currency(symbol).Normalized()
			if assert.NoError(t, err, "Currency(%q).Normalized()", symbol) {
				assert.Equal(t, code, norm, "Currency(%q) must normalize to its table entry", symbol)
			}
			assert.Contains(t, enums, code, "the currency %q of symbol %q is not an enum value", string(code), symbol)
			assert.True(t, code.ValidAndNormalized(), "the currency %q of symbol %q is not valid", string(code), symbol)
		}
	})

	t.Run("symbol lookup only knows enumerated currencies", func(t *testing.T) {
		for code := range currencyCodeToSymbol {
			assert.Contains(t, enums, code, "Symbol() knows %q, which is not an enum value", string(code))
		}
	})

	t.Run("Bitcoin is gone", func(t *testing.T) {
		assert.False(t, Currency("₿").Valid(), "₿ is not an ISO 4217 currency")
		assert.False(t, Currency("BTC").Valid(), "BTC is not an ISO 4217 currency")
		assert.NotContains(t, Currency("").EnumStrings(), "BTC")
	})

	t.Run("CurrencyNull sentinel", func(t *testing.T) {
		assert.NotContains(t, enums, Currency(CurrencyNull), "the SQL NULL is not a valid Currency")
		assert.NotContains(t, Currency("").EnumStrings(), CurrencyNull)
	})

	t.Run("EnumStrings matches Enums", func(t *testing.T) {
		// Both lists are generated independently from the same constants, so
		// nothing but this assertion keeps them from drifting apart.
		want := make([]string, len(enums))
		for i, c := range enums {
			want[i] = string(c)
		}
		assert.Equal(t, want, Currency("").EnumStrings())
	})
}

// TestStringIsCurrency pins that the currency detection used on free text
// accepts the same symbols, codes and names as NormalizeCurrency
// and rejects everything else.
func TestStringIsCurrency(t *testing.T) {
	for _, str := range []string{"EUR", "eur", " EUR ", "€", "$", "US$", "£", "Ft", "Euro", "SEK"} {
		assert.True(t, StringIsCurrency(str), "StringIsCurrency(%q)", str)
	}
	for _, str := range []string{"", " ", "XYZ", "EU", "12.34", "BTC"} {
		assert.False(t, StringIsCurrency(str), "StringIsCurrency(%q)", str)
	}
}

// TestCurrency_ScanString pins that ScanString normalizes valid input,
// rejects invalid input only when validating, and keeps the raw string
// otherwise so that non-validating imports don't lose data.
func TestCurrency_ScanString(t *testing.T) {
	tests := []struct {
		source   string
		validate bool
		want     Currency
		wantErr  bool
	}{
		{source: "eur", validate: true, want: EUR},
		{source: " € ", validate: true, want: EUR},
		{source: "US Dollar", validate: false, want: USD},
		{source: "XYZ", validate: true, wantErr: true},
		{source: "", validate: true, wantErr: true},
		{source: "XYZ", validate: false, want: "XYZ"},
		{source: "", validate: false, want: ""},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q validate=%t", tt.source, tt.validate), func(t *testing.T) {
			c := Currency("CHF")
			err := c.ScanString(tt.source, tt.validate)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, CHF, c, "a failed scan must not modify the currency")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, c)
		})
	}
}

// TestCurrency_GetOrDefault pins the nil-pointer safety of GetOrDefault
// and that invalid currencies are replaced by the default.
func TestCurrency_GetOrDefault(t *testing.T) {
	var nilCurrency *Currency
	assert.Equal(t, EUR, nilCurrency.GetOrDefault(EUR))
	assert.Equal(t, EUR, new(Currency).GetOrDefault(EUR), "empty currency is invalid")
	invalid := Currency("XYZ")
	assert.Equal(t, EUR, invalid.GetOrDefault(EUR))
	usd := USD
	assert.Equal(t, USD, usd.GetOrDefault(EUR))
}

// TestCurrency_ValidPtr pins that, unlike NullableCurrency.ValidPtr,
// a nil *Currency is not valid because Currency has no NULL value.
func TestCurrency_ValidPtr(t *testing.T) {
	var nilCurrency *Currency
	assert.False(t, nilCurrency.ValidPtr())
	assert.False(t, new(Currency).ValidPtr())
	eur := Currency("eur")
	assert.True(t, eur.ValidPtr())
}

// TestCurrency_NullableCurrency pins the conversion to NullableCurrency
// where the empty Currency becomes the valid NULL value.
func TestCurrency_NullableCurrency(t *testing.T) {
	assert.Equal(t, NullableCurrency("EUR"), EUR.NullableCurrency())
	empty := Currency("").NullableCurrency()
	assert.True(t, empty.IsNull())
	assert.True(t, empty.Valid())
}

// TestCurrency_NormalizedOrNull pins that unknown currencies degrade to NULL
// instead of being stored in their invalid form.
func TestCurrency_NormalizedOrNull(t *testing.T) {
	assert.Equal(t, NullableCurrency("EUR"), Currency("€").NormalizedOrNull())
	assert.Equal(t, NullableCurrency("USD"), Currency(" usd ").NormalizedOrNull())
	assert.Equal(t, NullableCurrency(CurrencyNull), Currency("XYZ").NormalizedOrNull())
	assert.Equal(t, NullableCurrency(CurrencyNull), Currency("").NormalizedOrNull())
}

// TestCurrency_IsEUR pins that IsEUR recognizes every way the Euro is
// written on invoices, not only the normalized code.
func TestCurrency_IsEUR(t *testing.T) {
	for _, c := range []Currency{"EUR", "eur", "€", "Euro", " EURO "} {
		assert.True(t, c.IsEUR(), "Currency(%q).IsEUR()", string(c))
	}
	for _, c := range []Currency{"USD", "$", "", "XYZ"} {
		assert.False(t, c.IsEUR(), "Currency(%q).IsEUR()", string(c))
	}
}

// TestCurrency_SQL pins that the empty Currency is written to SQL as NULL
// instead of an empty string, that SQL NULL scans back to the empty
// Currency, and that both text representations of drivers are accepted.
func TestCurrency_SQL(t *testing.T) {
	v, err := Currency(CurrencyNull).Value()
	require.NoError(t, err)
	assert.Nil(t, v)

	v, err = EUR.Value()
	require.NoError(t, err)
	assert.Equal(t, "EUR", v)

	var c Currency
	require.NoError(t, c.Scan("USD"))
	assert.Equal(t, USD, c)

	src := []byte("CHF")
	require.NoError(t, c.Scan(src))
	src[0] = 'X' // drivers may reuse the buffer after Scan returns
	assert.Equal(t, CHF, c)

	require.NoError(t, c.Scan(nil))
	assert.Equal(t, Currency(CurrencyNull), c)

	c = EUR
	assert.Error(t, c.Scan(int64(1)))
	assert.Equal(t, EUR, c, "a failed scan must not modify the currency")
}

// TestCurrency_SymbolAndEnglishName pins the display helpers: well known
// currencies get their symbol, all others fall back to their code,
// and unknown codes have no English name.
func TestCurrency_SymbolAndEnglishName(t *testing.T) {
	tests := []struct {
		currency Currency
		symbol   string
		name     string
	}{
		{currency: EUR, symbol: "€", name: "Euro Member Countries"},
		{currency: USD, symbol: "$", name: "United States Dollar"},
		{currency: GBP, symbol: "£", name: "United Kingdom Pound"},
		{currency: CHF, symbol: "₣", name: "Switzerland Franc"},
		{currency: JPY, symbol: "¥", name: "Japan Yen"},
		{currency: SEK, symbol: "SEK", name: "Sweden Krona"},
		{currency: "XYZ", symbol: "XYZ", name: ""},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.symbol, tt.currency.Symbol(), "Currency(%q).Symbol()", string(tt.currency))
		assert.Equal(t, tt.name, tt.currency.EnglishName(), "Currency(%q).EnglishName()", string(tt.currency))
	}
}

// TestCurrency_NormalizedNames pins how written currency names normalize.
// A name has to resolve to the currency it names: "Canadian Dollar" is not
// a US Dollar and "Burundi Franc" is not a Swiss Franc, so a qualified name
// of another currency must never fall through to the USD, CHF or GBP
// guesses. Those guesses are only for the unqualified unit (as in a
// "Dollar" column of an US invoice) or for a qualifier that names the
// country. A qualified name of an unknown currency is an error instead of
// a wrong currency, because a wrong currency silently corrupts the amount.
func TestCurrency_NormalizedNames(t *testing.T) {
	tests := []struct {
		in   string
		want Currency
	}{
		// ISO 4217 names
		{in: "Canadian Dollar", want: CAD},
		{in: "Burundi Franc", want: BIF},
		{in: "Egyptian Pound", want: EGP},
		{in: "Australian Dollar", want: AUD},
		{in: "Hong Kong Dollar", want: HKD},
		{in: "Danish Krone", want: DKK},
		{in: "CFA Franc BEAC", want: XAF},
		{in: "South Sudanese Pound", want: SSP},
		{in: "  canadian DOLLAR ", want: CAD},
		// English names as returned by EnglishName
		{in: "Canada Dollar", want: CAD},
		{in: "Egypt Pound", want: EGP},
		{in: "United States Dollar", want: USD},
		{in: "Switzerland Franc", want: CHF},
		{in: "United Kingdom Pound", want: GBP},
		// US Dollar guesses
		{in: "US Dollar", want: USD},
		{in: "U.S. Dollar", want: USD},
		{in: "US-Dollar", want: USD},
		{in: "Dollar US", want: USD},
		{in: "USA Dollars", want: USD},
		{in: "Dollar", want: USD},
		{in: "dollars", want: USD},
		// Swiss Franc guesses
		{in: "Swiss Franc", want: CHF},
		{in: "Swiss Francs", want: CHF},
		{in: "Schweizer Franken", want: CHF},
		{in: "Franc suisse", want: CHF},
		{in: "Franken", want: CHF},
		{in: "Franc", want: CHF},
		{in: "Francs", want: CHF},
		// British Pound guesses
		{in: "Pound Sterling", want: GBP},
		{in: "British Pound", want: GBP},
		{in: "Sterling", want: GBP},
		{in: "Pound", want: GBP},
		{in: "Pounds", want: GBP},
		// Euro guesses
		{in: "Euro", want: EUR},
		{in: "Euros", want: EUR},
	}
	for _, tt := range tests {
		got, err := Currency(tt.in).Normalized()
		if assert.NoError(t, err, "Currency(%q).Normalized()", tt.in) {
			assert.Equal(t, tt.want, got, "Currency(%q).Normalized()", tt.in)
		}
	}

	for _, in := range []string{
		"Atlantean Dollar",
		"Martian Franc",
		"Lunar Pound",
		"Burundi-Franken",
		"Dollar Franc",
	} {
		got, err := Currency(in).Normalized()
		assert.Error(t, err, "Currency(%q).Normalized() = %q must fail instead of guessing", in, string(got))
	}
}

// TestCurrency_NamesResolveToTheirCurrency pins that every English name
// and every ISO 4217 name in the tables normalizes back to its own
// currency. Two currencies sharing a name, or a name mapped to a code
// that is no enum value, would break this.
func TestCurrency_NamesResolveToTheirCurrency(t *testing.T) {
	enums := Currency("").Enums()
	for _, code := range enums {
		name := code.EnglishName()
		got, err := Currency(name).Normalized()
		if assert.NoError(t, err, "EnglishName %q of %s", name, string(code)) {
			assert.Equal(t, code, got, "EnglishName %q of %s", name, string(code))
		}
	}
	for code, name := range currencyCodeToISOName {
		assert.Contains(t, enums, code, "ISO name %q of %s", name, string(code))
		got, err := Currency(name).Normalized()
		if assert.NoError(t, err, "ISO name %q of %s", name, string(code)) {
			assert.Equal(t, code, got, "ISO name %q of %s", name, string(code))
		}
	}
}

// TestCurrency_CurrentISO4217Codes pins that the codes ISO 4217 introduced
// for redenominations and new currencies are valid enum values with an
// English name, because invoices in those currencies must not be rejected
// as invalid. The withdrawn codes they replaced stay valid, because old
// documents and stored data still carry them.
func TestCurrency_CurrentISO4217Codes(t *testing.T) {
	enums := Currency("").Enums()
	current := []Currency{
		"MRU", // Mauritania, replaced MRO in 2018
		"SLE", // Sierra Leone, replaced SLL in 2022
		"SSP", // South Sudan, since 2011
		"STN", // São Tomé and Príncipe, replaced STD in 2018
		"VED", // Venezuela digital bolívar, since 2021
		"VES", // Venezuela, replaced VEF in 2018
		"XCG", // Curaçao and Sint Maarten, replaced ANG in 2025
		"ZWG", // Zimbabwe Gold, replaced ZWL in 2024
	}
	for _, code := range current {
		assert.True(t, code.Valid(), "Currency(%q).Valid()", string(code))
		assert.True(t, code.ValidAndNormalized(), "Currency(%q).ValidAndNormalized()", string(code))
		assert.Contains(t, enums, code, "Currency(%q) in Enums()", string(code))
		assert.NotEmpty(t, code.EnglishName(), "Currency(%q).EnglishName()", string(code))
	}
	withdrawn := []Currency{"ANG", "HRK", "MRO", "SLL", "STD", "VEF", "ZWD", "ZWL"}
	for _, code := range withdrawn {
		assert.True(t, code.Valid(), "withdrawn Currency(%q) must stay valid for old data", string(code))
		assert.Contains(t, enums, code, "withdrawn Currency(%q) in Enums()", string(code))
	}
}

// TestCurrency_ZWL protects that the Zimbabwe Dollar used from 2009 until
// ZWG replaced it in 2024 is a valid currency, because invoices and bank
// statements of those years carry ZWL. Its name has to differ from the
// earlier Zimbabwe Dollar ZWD and from ZWG, so that each of the three
// names normalizes back to its own code.
func TestCurrency_ZWL(t *testing.T) {
	assert.Equal(t, Currency("ZWL"), ZWL)
	assert.True(t, ZWL.ValidAndNormalized())
	assert.Contains(t, ZWL.Enums(), ZWL)
	assert.Contains(t, ZWL.EnumStrings(), "ZWL")

	got, err := Currency("zwl").Normalized()
	assert.NoError(t, err)
	assert.Equal(t, ZWL, got)

	names := map[Currency]string{ZWD: ZWD.EnglishName(), ZWG: ZWG.EnglishName(), ZWL: ZWL.EnglishName()}
	for code, name := range names {
		assert.NotEmpty(t, name, "EnglishName of %s", string(code))
		normalized, err := Currency(name).Normalized()
		if assert.NoError(t, err, "Currency(%q).Normalized()", name) {
			assert.Equal(t, code, normalized, "Currency(%q).Normalized()", name)
		}
	}
	assert.NotEqual(t, ZWD.EnglishName(), ZWL.EnglishName())
}
