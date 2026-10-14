package money

import (
	"database/sql/driver"
	"encoding/xml"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/domonda/go-types"
	"github.com/domonda/go-types/internal"
	"github.com/domonda/go-types/strutil"
)

// Compile-time check that Currency implements types.NormalizableValidator[Currency]
var _ types.NormalizableValidator[Currency] = Currency("")

// StringIsCurrency returns if a string can be parsed as Currency.
func StringIsCurrency(str string) bool {
	_, err := NormalizeCurrency(str)
	return err == nil
}

// Currency is holds a 3 character ISO 4217 alphabetic code.
// The main difference between Currency and NullableCurrency is:
// Currency("").Valid() == false
// NullableCurrency("").Valid() == true
//
// Scan, UnmarshalJSON, UnmarshalXML and ScanString store a valid currency
// normalized, a blank one as CurrencyNull and an invalid one unchanged.
// Value writes the normalized currency, SQL NULL for a blank one, and
// returns an error for an invalid one. MarshalJSON and MarshalText write
// the normalized currency or an invalid one unchanged, and fmt prints the
// currency as stored. Currency has no UnmarshalText, so encoding/json reads
// and writes the keys of a map[Currency]V as given. See "Reading and
// writing values" in the README of the module.
type Currency string //#enum

// NormalizeCurrency returns str as normalized Currency or an error.
func NormalizeCurrency(str string) (Currency, error) {
	return Currency(str).Normalized()
}

// ScanString implements the strfmt.Scannable interface.
// Without validate it stores source like Scan, with validate it stores
// the normalized currency or returns the error of Normalized,
// also for a blank source.
func (c *Currency) ScanString(source string, validate bool) error {
	return internal.ScanStringNormalized(c, source, validate, false, Currency.Normalized, nil)
}

// GetOrDefault returns the value c references if it is valid and c is not nil.
// Safe to call on a nil pointer.
func (c *Currency) GetOrDefault(defaultVal Currency) Currency {
	if !c.ValidPtr() {
		return defaultVal
	}
	return *c
}

// NullableCurrency returns c as NullableCurrency where
// NullableCurrency("").Valid() == true
func (c Currency) NullableCurrency() NullableCurrency {
	return NullableCurrency(c)
}

// Valid indicates if c can be normalized to a valid currency.
//
// Hand-written: accepts non-normalized input (lower case, currency
// symbols, English names) via Normalized, unlike the exact-match switch
// over every enum value that go-enum would generate.
//
// #custom
func (c Currency) Valid() bool {
	return c.Validate() == nil
}

// ValidAndNormalized returns true if the Currency is valid and already normalized.
func (c Currency) ValidAndNormalized() bool {
	norm, err := c.Normalized()
	return err == nil && c == norm
}

// Validate returns an error if c can not be normalized to a valid currency.
//
// Hand-written: returns the error from Normalized, which names the
// offending input, instead of the generic error that go-enum would
// generate.
//
// #custom
func (c Currency) Validate() error {
	_, err := c.Normalized()
	return err
}

// ValidPtr returns if c is not nil and can be normalized to a valid currency.
// Safe to call on a nil pointer.
func (c *Currency) ValidPtr() bool {
	return c != nil && c.Valid()
}

// Normalized normalizes a currency string.
//
// Besides the ISO 4217 code in any letter case it accepts the currency
// symbols like € or $, the English names returned by EnglishName,
// and the ISO 4217 currency names of two or more words like
// "Canadian Dollar", all compared case-insensitively.
//
// Other strings are guessed by the unit they name: "EUR", "Euro(s)"
// as EUR, "Dollar(s)" as USD, "Franc(s)" or "Franken" as CHF, and
// "Pound(s)", "Pfund" or "Sterling" as GBP. The unit may be qualified with
// the country (US, U.S., USA or United States for USD, Swiss, Schweizer,
// Schweizerische(r) or suisse(s) for CHF, British, Great British or livre
// for GBP) or with the ISO code like in
// "Dollars (USD)". A unit qualified with anything else is an error
// instead of a guess, because a dollar, franc or pound of another
// country is a different currency, and so is a country without a unit,
// because "Made in USA" is free text and not a currency.
//
// Returns the Currency unchanged in case of an error.
func (c Currency) Normalized() (Currency, error) {
	str := strutil.TrimSpace(string(c))

	if str == "" {
		return c, errors.New("empty currency code")
	}

	result, found := currencySymbolToCode[str]
	if found {
		return result, nil
	}

	str = strings.ToUpper(str)

	if len(str) > 3 {
		if result, found = currencyNameToCode[str]; found {
			return result, nil
		}
	}
	// Only letters, spaces and the punctuation of names like
	// "U.S. Dollar", "US-Dollar" or "Dollars (USD)" are guessed,
	// so that "USD 1.5" or "€ USD" are no currencies
	if len(str) > 3 && !strings.ContainsFunc(str, isNoCurrencyNameRune) {
		// Words of letters only, so "U.S." becomes "US"
		// and "US-Dollar" the two words "US" and "DOLLAR"
		words := strings.FieldsFunc(strings.ReplaceAll(str, ".", ""), func(r rune) bool { return !unicode.IsLetter(r) })
		switch {
		case unitWithQualifiers(words, []string{"EUR", "EURO", "EUROS"}):
			return "EUR", nil
		case unitWithQualifiers(words, []string{"FRANC", "FRANCS", "FRANKEN", "CHF"}, "SWISS", "SCHWEIZER", "SCHWEITZER", "SCHWEIZERISCHE", "SCHWEIZERISCHER", "SCHWEIZERISCHEN", "SUISSE", "SUISSES"):
			return "CHF", nil
		case unitWithQualifiers(words, []string{"POUND", "POUNDS", "PFUND", "STERLING", "GBP"}, "GREAT", "BRITISH", "LIVRE", "LIVRES"):
			return "GBP", nil
		case unitWithQualifiers(words, []string{"DOLLAR", "DOLLARS", "USD"}, "US", "USA", "UNITED", "STATES"):
			return "USD", nil
		}
	}

	if _, ok := currencyCodeToName[Currency(str)]; !ok {
		return c, fmt.Errorf("invalid currency code %q", str)
	}

	return Currency(str), nil
}

// isNoCurrencyNameRune returns if r can't be part of a currency name.
func isNoCurrencyNameRune(r rune) bool {
	return !unicode.IsLetter(r) && !unicode.IsSpace(r) && !strings.ContainsRune(".-()", r)
}

// unitWithQualifiers returns if the words contain at least one of the
// units and all other words are qualifiers, so that a unit qualified
// with an unknown word or a qualifier without a unit is not a currency.
func unitWithQualifiers(words, units []string, qualifiers ...string) bool {
	hasUnit := false
	for _, word := range words {
		switch {
		case slices.Contains(units, word):
			hasUnit = true
		case !slices.Contains(qualifiers, word):
			return false
		}
	}
	return hasUnit
}

// NormalizedOrNull returns the normalized currency or CurrencyNull
// if it was not a valid currency.
func (c Currency) NormalizedOrNull() NullableCurrency {
	norm, err := c.Normalized()
	if err != nil {
		return CurrencyNull
	}
	return NullableCurrency(norm)
}

// IsEUR returns if the currency can be normalized as Euro.
func (c Currency) IsEUR() bool {
	norm, _ := c.Normalized()
	return norm == EUR
}

// Scan implements the database/sql.Scanner interface.
// A valid currency is stored normalized, SQL NULL and a blank string as
// CurrencyNull, and a currency that can't be normalized unchanged without
// an error, so that one invalid value doesn't fail a whole query;
// Validate reports it.
func (c *Currency) Scan(value any) error {
	return internal.ScanNormalized(c, value, Currency.Normalized, nil)
}

// Value implements the driver database/sql/driver.Valuer interface.
// It returns the normalized currency, SQL NULL for a blank Currency,
// or the error of Normalized for a currency that can't be normalized.
func (c Currency) Value() (driver.Value, error) {
	return internal.ValueNormalized(c, Currency.Normalized, nil)
}

// normalizedForValue returns the currency that the Value methods of
// CurrencyAmount and CurrencyDecimalAmount write: the currency of Value,
// "" for a blank one, for which Value returns SQL NULL, or the error of
// Normalized for a currency that can't be normalized.
func (c Currency) normalizedForValue() (Currency, error) {
	currency, err := c.Value()
	if err != nil {
		return "", err
	}
	str, _ := currency.(string) // nil for a blank currency
	return Currency(str), nil
}

// MarshalJSON implements encoding/json.Marshaler by writing the
// normalized currency, "" for a blank Currency, or the currency
// unchanged if it can't be normalized.
func (c Currency) MarshalJSON() ([]byte, error) {
	return internal.MarshalJSONNormalized(c, false, Currency.Normalized, nil)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// A JSON string is stored like Scan stores it, JSON null leaves the
// Currency unchanged like encoding/json does for a string type,
// and any other JSON value is an error.
func (c *Currency) UnmarshalJSON(data []byte) error {
	return internal.UnmarshalJSONNormalized(c, data, false, Currency.Normalized, nil)
}

// MarshalText implements encoding.TextMarshaler like MarshalJSON,
// so that encoding/xml and strfmt write the normalized currency.
// encoding/json writes map keys of string types as they are
// and doesn't call MarshalText for them.
func (c Currency) MarshalText() ([]byte, error) {
	return internal.MarshalTextNormalized(c, Currency.Normalized, nil)
}

// MarshalXMLAttr implements encoding/xml.MarshalerAttr by writing the
// text of MarshalText, and no attribute for a nil pointer, because
// encoding/xml calls MarshalText through a nil pointer attribute field
// and panics.
func (c *Currency) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return internal.MarshalXMLAttrText(c, name)
}

// UnmarshalXML implements encoding/xml.Unmarshaler
// by storing the text of the element like Scan stores it.
func (c *Currency) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return internal.UnmarshalXMLNormalized(c, d, start, Currency.Normalized, nil)
}

// UnmarshalXMLAttr implements encoding/xml.UnmarshalerAttr
// by storing the attribute value like Scan stores it.
func (c *Currency) UnmarshalXMLAttr(attr xml.Attr) error {
	return internal.UnmarshalXMLAttrNormalized(c, attr, Currency.Normalized, nil)
}

// Symbol returns the currency symbol like € for EUR if available,
// or currency code if no widely recognized symbol is available.
func (c Currency) Symbol() string {
	if s, ok := currencyCodeToSymbol[c]; ok {
		return s
	}
	return string(c)
}

// EnglishName returns the english name of the currency
func (c Currency) EnglishName() string {
	return currencyCodeToName[c]
}

// MinorUnits returns the number of decimal places of the minor unit
// of the currency defined by ISO 4217, like 2 for the cents of EUR,
// 0 for JPY and 3 for the fils of KWD.
// Codes without a minor unit defined by ISO 4217, like XDR,
// codes that are not part of ISO 4217, like GGP, and invalid codes
// return 2. The currency is normalized first, so "kwd" returns 3.
func (c Currency) MinorUnits() int {
	norm, err := c.Normalized()
	if err == nil {
		c = norm
	}
	return c.minorUnitsOfNormalized()
}

// minorUnitsOfNormalized returns MinorUnits of the normalized or blank c
// without normalizing it again.
func (c Currency) minorUnitsOfNormalized() int {
	if minorUnits, ok := currencyCodeToMinorUnits[c]; ok {
		return minorUnits
	}
	return 2
}

///////////////////////////////////////////////////////////////////////////////
// ISO 4217 currency code constants
//
// CurrencyNull stays untyped in constants.go because it is shared with
// NullableCurrency. It is not a valid Currency, so do not move it here.

const (
	AED Currency = "AED" // United Arab Emirates Dirham
	AFN Currency = "AFN" // Afghanistan Afghani
	ALL Currency = "ALL" // Albania Lek
	AMD Currency = "AMD" // Armenia Dram
	ANG Currency = "ANG" // Netherlands Antilles Guilder
	AOA Currency = "AOA" // Angola Kwanza
	ARS Currency = "ARS" // Argentina Peso
	AUD Currency = "AUD" // Australia Dollar
	AWG Currency = "AWG" // Aruba Guilder
	AZN Currency = "AZN" // Azerbaijan New Manat
	BAM Currency = "BAM" // Bosnia and Herzegovina Convertible Marka
	BBD Currency = "BBD" // Barbados Dollar
	BDT Currency = "BDT" // Bangladesh Taka
	BGN Currency = "BGN" // Bulgaria Lev
	BHD Currency = "BHD" // Bahrain Dinar
	BIF Currency = "BIF" // Burundi Franc
	BMD Currency = "BMD" // Bermuda Dollar
	BND Currency = "BND" // Brunei Darussalam Dollar
	BOB Currency = "BOB" // Bolivia Bolíviano
	BRL Currency = "BRL" // Brazil Real
	BSD Currency = "BSD" // Bahamas Dollar
	BTN Currency = "BTN" // Bhutan Ngultrum
	BWP Currency = "BWP" // Botswana Pula
	BYN Currency = "BYN" // Belarus Ruble
	BZD Currency = "BZD" // Belize Dollar
	CAD Currency = "CAD" // Canada Dollar
	CDF Currency = "CDF" // Congo/Kinshasa Franc
	CHF Currency = "CHF" // Switzerland Franc
	CLP Currency = "CLP" // Chile Peso
	CNY Currency = "CNY" // China Yuan Renminbi
	COP Currency = "COP" // Colombia Peso
	CRC Currency = "CRC" // Costa Rica Colon
	CUC Currency = "CUC" // Cuba Convertible Peso
	CUP Currency = "CUP" // Cuba Peso
	CVE Currency = "CVE" // Cape Verde Escudo
	CZK Currency = "CZK" // Czech Republic Koruna
	DJF Currency = "DJF" // Djibouti Franc
	DKK Currency = "DKK" // Denmark Krone
	DOP Currency = "DOP" // Dominican Republic Peso
	DZD Currency = "DZD" // Algeria Dinar
	EGP Currency = "EGP" // Egypt Pound
	ERN Currency = "ERN" // Eritrea Nakfa
	ETB Currency = "ETB" // Ethiopia Birr
	EUR Currency = "EUR" // Euro Member Countries
	FJD Currency = "FJD" // Fiji Dollar
	FKP Currency = "FKP" // Falkland Islands (Malvinas) Pound
	GBP Currency = "GBP" // United Kingdom Pound
	GEL Currency = "GEL" // Georgia Lari
	GGP Currency = "GGP" // Guernsey Pound
	GHS Currency = "GHS" // Ghana Cedi
	GIP Currency = "GIP" // Gibraltar Pound
	GMD Currency = "GMD" // Gambia Dalasi
	GNF Currency = "GNF" // Guinea Franc
	GTQ Currency = "GTQ" // Guatemala Quetzal
	GYD Currency = "GYD" // Guyana Dollar
	HKD Currency = "HKD" // Hong Kong Dollar
	HNL Currency = "HNL" // Honduras Lempira
	HRK Currency = "HRK" // Croatia Kuna
	HTG Currency = "HTG" // Haiti Gourde
	HUF Currency = "HUF" // Hungary Forint
	IDR Currency = "IDR" // Indonesia Rupiah
	ILS Currency = "ILS" // Israel Shekel
	IMP Currency = "IMP" // Isle of Man Pound
	INR Currency = "INR" // India Rupee
	IQD Currency = "IQD" // Iraq Dinar
	IRR Currency = "IRR" // Iran Rial
	ISK Currency = "ISK" // Iceland Krona
	JEP Currency = "JEP" // Jersey Pound
	JMD Currency = "JMD" // Jamaica Dollar
	JOD Currency = "JOD" // Jordan Dinar
	JPY Currency = "JPY" // Japan Yen
	KES Currency = "KES" // Kenya Shilling
	KGS Currency = "KGS" // Kyrgyzstan Som
	KHR Currency = "KHR" // Cambodia Riel
	KMF Currency = "KMF" // Comoros Franc
	KPW Currency = "KPW" // Korea (North) Won
	KRW Currency = "KRW" // Korea (South) Won
	KWD Currency = "KWD" // Kuwait Dinar
	KYD Currency = "KYD" // Cayman Islands Dollar
	KZT Currency = "KZT" // Kazakhstan Tenge
	LAK Currency = "LAK" // Laos Kip
	LBP Currency = "LBP" // Lebanon Pound
	LKR Currency = "LKR" // Sri Lanka Rupee
	LRD Currency = "LRD" // Liberia Dollar
	LSL Currency = "LSL" // Lesotho Loti
	LYD Currency = "LYD" // Libya Dinar
	MAD Currency = "MAD" // Morocco Dirham
	MDL Currency = "MDL" // Moldova Leu
	MGA Currency = "MGA" // Madagascar Ariary
	MKD Currency = "MKD" // Macedonia Denar
	MMK Currency = "MMK" // Myanmar (Burma) Kyat
	MNT Currency = "MNT" // Mongolia Tughrik
	MOP Currency = "MOP" // Macau Pataca
	MRO Currency = "MRO" // Mauritania Ouguiya
	MRU Currency = "MRU" // Mauritania New Ouguiya
	MUR Currency = "MUR" // Mauritius Rupee
	MVR Currency = "MVR" // Maldives (Maldive Islands) Rufiyaa
	MWK Currency = "MWK" // Malawi Kwacha
	MXN Currency = "MXN" // Mexico Peso
	MYR Currency = "MYR" // Malaysia Ringgit
	MZN Currency = "MZN" // Mozambique Metical
	NAD Currency = "NAD" // Namibia Dollar
	NGN Currency = "NGN" // Nigeria Naira
	NIO Currency = "NIO" // Nicaragua Cordoba
	NOK Currency = "NOK" // Norway Krone
	NPR Currency = "NPR" // Nepal Rupee
	NZD Currency = "NZD" // New Zealand Dollar
	OMR Currency = "OMR" // Oman Rial
	PAB Currency = "PAB" // Panama Balboa
	PEN Currency = "PEN" // Peru Sol
	PGK Currency = "PGK" // Papua New Guinea Kina
	PHP Currency = "PHP" // Philippines Peso
	PKR Currency = "PKR" // Pakistan Rupee
	PLN Currency = "PLN" // Poland Zloty
	PYG Currency = "PYG" // Paraguay Guarani
	QAR Currency = "QAR" // Qatar Riyal
	RON Currency = "RON" // Romania New Leu
	RSD Currency = "RSD" // Serbia Dinar
	RUB Currency = "RUB" // Russia Ruble
	RWF Currency = "RWF" // Rwanda Franc
	SAR Currency = "SAR" // Saudi Arabia Riyal
	SBD Currency = "SBD" // Solomon Islands Dollar
	SCR Currency = "SCR" // Seychelles Rupee
	SDG Currency = "SDG" // Sudan Pound
	SEK Currency = "SEK" // Sweden Krona
	SGD Currency = "SGD" // Singapore Dollar
	SHP Currency = "SHP" // Saint Helena Pound
	SLE Currency = "SLE" // Sierra Leone New Leone
	SLL Currency = "SLL" // Sierra Leone Leone
	SOS Currency = "SOS" // Somalia Shilling
	SPL Currency = "SPL" // Seborga Luigino
	SRD Currency = "SRD" // Suriname Dollar
	SSP Currency = "SSP" // South Sudan Pound
	STD Currency = "STD" // São Tomé and Príncipe Dobra
	STN Currency = "STN" // São Tomé and Príncipe New Dobra
	SVC Currency = "SVC" // El Salvador Colon
	SYP Currency = "SYP" // Syria Pound
	SZL Currency = "SZL" // Swaziland Lilangeni
	THB Currency = "THB" // Thailand Baht
	TJS Currency = "TJS" // Tajikistan Somoni
	TMT Currency = "TMT" // Turkmenistan Manat
	TND Currency = "TND" // Tunisia Dinar
	TOP Currency = "TOP" // Tonga Pa'anga
	TRY Currency = "TRY" // Turkey Lira
	TTD Currency = "TTD" // Trinidad and Tobago Dollar
	TVD Currency = "TVD" // Tuvalu Dollar
	TWD Currency = "TWD" // Taiwan New Dollar
	TZS Currency = "TZS" // Tanzania Shilling
	UAH Currency = "UAH" // Ukraine Hryvnia
	UGX Currency = "UGX" // Uganda Shilling
	USD Currency = "USD" // United States Dollar
	UYU Currency = "UYU" // Uruguay Peso
	UZS Currency = "UZS" // Uzbekistan Som
	VED Currency = "VED" // Venezuela Bolívar Digital
	VEF Currency = "VEF" // Venezuela Bolivar
	VES Currency = "VES" // Venezuela Bolívar Soberano
	VND Currency = "VND" // Viet Nam Dong
	VUV Currency = "VUV" // Vanuatu Vatu
	WST Currency = "WST" // Samoa Tala
	XAF Currency = "XAF" // Communauté Financière Africaine (BEAC) CFA Franc BEAC
	XCD Currency = "XCD" // East Caribbean Dollar
	XCG Currency = "XCG" // Caribbean Guilder
	XDR Currency = "XDR" // International Monetary Fund (IMF) Special Drawing Rights
	XOF Currency = "XOF" // Communauté Financière Africaine (BCEAO) Franc
	XPF Currency = "XPF" // Comptoirs Français du Pacifique (CFP) Franc
	YER Currency = "YER" // Yemen Rial
	ZAR Currency = "ZAR" // South Africa Rand
	ZMW Currency = "ZMW" // Zambia Kwacha
	ZWD Currency = "ZWD" // Zimbabwe Dollar
	ZWG Currency = "ZWG" // Zimbabwe Gold
	ZWL Currency = "ZWL" // Zimbabwe Dollar (2009-2024)
)

// Enums returns all valid values for Currency
func (Currency) Enums() []Currency {
	return []Currency{
		AED,
		AFN,
		ALL,
		AMD,
		ANG,
		AOA,
		ARS,
		AUD,
		AWG,
		AZN,
		BAM,
		BBD,
		BDT,
		BGN,
		BHD,
		BIF,
		BMD,
		BND,
		BOB,
		BRL,
		BSD,
		BTN,
		BWP,
		BYN,
		BZD,
		CAD,
		CDF,
		CHF,
		CLP,
		CNY,
		COP,
		CRC,
		CUC,
		CUP,
		CVE,
		CZK,
		DJF,
		DKK,
		DOP,
		DZD,
		EGP,
		ERN,
		ETB,
		EUR,
		FJD,
		FKP,
		GBP,
		GEL,
		GGP,
		GHS,
		GIP,
		GMD,
		GNF,
		GTQ,
		GYD,
		HKD,
		HNL,
		HRK,
		HTG,
		HUF,
		IDR,
		ILS,
		IMP,
		INR,
		IQD,
		IRR,
		ISK,
		JEP,
		JMD,
		JOD,
		JPY,
		KES,
		KGS,
		KHR,
		KMF,
		KPW,
		KRW,
		KWD,
		KYD,
		KZT,
		LAK,
		LBP,
		LKR,
		LRD,
		LSL,
		LYD,
		MAD,
		MDL,
		MGA,
		MKD,
		MMK,
		MNT,
		MOP,
		MRO,
		MRU,
		MUR,
		MVR,
		MWK,
		MXN,
		MYR,
		MZN,
		NAD,
		NGN,
		NIO,
		NOK,
		NPR,
		NZD,
		OMR,
		PAB,
		PEN,
		PGK,
		PHP,
		PKR,
		PLN,
		PYG,
		QAR,
		RON,
		RSD,
		RUB,
		RWF,
		SAR,
		SBD,
		SCR,
		SDG,
		SEK,
		SGD,
		SHP,
		SLE,
		SLL,
		SOS,
		SPL,
		SRD,
		SSP,
		STD,
		STN,
		SVC,
		SYP,
		SZL,
		THB,
		TJS,
		TMT,
		TND,
		TOP,
		TRY,
		TTD,
		TVD,
		TWD,
		TZS,
		UAH,
		UGX,
		USD,
		UYU,
		UZS,
		VED,
		VEF,
		VES,
		VND,
		VUV,
		WST,
		XAF,
		XCD,
		XCG,
		XDR,
		XOF,
		XPF,
		YER,
		ZAR,
		ZMW,
		ZWD,
		ZWG,
		ZWL,
	}
}

// EnumStrings returns all valid values for Currency as strings
func (Currency) EnumStrings() []string {
	return []string{
		"AED",
		"AFN",
		"ALL",
		"AMD",
		"ANG",
		"AOA",
		"ARS",
		"AUD",
		"AWG",
		"AZN",
		"BAM",
		"BBD",
		"BDT",
		"BGN",
		"BHD",
		"BIF",
		"BMD",
		"BND",
		"BOB",
		"BRL",
		"BSD",
		"BTN",
		"BWP",
		"BYN",
		"BZD",
		"CAD",
		"CDF",
		"CHF",
		"CLP",
		"CNY",
		"COP",
		"CRC",
		"CUC",
		"CUP",
		"CVE",
		"CZK",
		"DJF",
		"DKK",
		"DOP",
		"DZD",
		"EGP",
		"ERN",
		"ETB",
		"EUR",
		"FJD",
		"FKP",
		"GBP",
		"GEL",
		"GGP",
		"GHS",
		"GIP",
		"GMD",
		"GNF",
		"GTQ",
		"GYD",
		"HKD",
		"HNL",
		"HRK",
		"HTG",
		"HUF",
		"IDR",
		"ILS",
		"IMP",
		"INR",
		"IQD",
		"IRR",
		"ISK",
		"JEP",
		"JMD",
		"JOD",
		"JPY",
		"KES",
		"KGS",
		"KHR",
		"KMF",
		"KPW",
		"KRW",
		"KWD",
		"KYD",
		"KZT",
		"LAK",
		"LBP",
		"LKR",
		"LRD",
		"LSL",
		"LYD",
		"MAD",
		"MDL",
		"MGA",
		"MKD",
		"MMK",
		"MNT",
		"MOP",
		"MRO",
		"MRU",
		"MUR",
		"MVR",
		"MWK",
		"MXN",
		"MYR",
		"MZN",
		"NAD",
		"NGN",
		"NIO",
		"NOK",
		"NPR",
		"NZD",
		"OMR",
		"PAB",
		"PEN",
		"PGK",
		"PHP",
		"PKR",
		"PLN",
		"PYG",
		"QAR",
		"RON",
		"RSD",
		"RUB",
		"RWF",
		"SAR",
		"SBD",
		"SCR",
		"SDG",
		"SEK",
		"SGD",
		"SHP",
		"SLE",
		"SLL",
		"SOS",
		"SPL",
		"SRD",
		"SSP",
		"STD",
		"STN",
		"SVC",
		"SYP",
		"SZL",
		"THB",
		"TJS",
		"TMT",
		"TND",
		"TOP",
		"TRY",
		"TTD",
		"TVD",
		"TWD",
		"TZS",
		"UAH",
		"UGX",
		"USD",
		"UYU",
		"UZS",
		"VED",
		"VEF",
		"VES",
		"VND",
		"VUV",
		"WST",
		"XAF",
		"XCD",
		"XCG",
		"XDR",
		"XOF",
		"XPF",
		"YER",
		"ZAR",
		"ZMW",
		"ZWD",
		"ZWG",
		"ZWL",
	}
}

// String implements the fmt.Stringer interface for Currency
func (c Currency) String() string {
	return string(c)
}
