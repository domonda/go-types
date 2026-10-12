// Package country provides comprehensive country code handling and validation
// based on ISO 3166-1 alpha-2 standards for Go applications.
//
// The package includes:
//   - ISO 3166-1 alpha-2 country code validation and normalization
//   - Alternative country code mappings (ITU codes, German names, etc.)
//   - European Union membership checking
//   - Database integration (Scanner/Valuer interfaces)
//   - JSON marshalling/unmarshalling
//   - Nullable country code support
package country

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"iter"
	"maps"
	"strings"
	"sync"

	"github.com/invopop/jsonschema"

	"github.com/domonda/go-types"
	"github.com/domonda/go-types/strutil"
)

// Invalid represents an invalid country code.
//
// Declared as a conversion instead of `Invalid Code = ""` so that go-enum
// does not pick it up as one of the Code enum values: Invalid is not a
// valid country code and must stay out of Code.Enums() and
// Code.EnumStrings().
const Invalid = Code("")

// Compile-time check that Code implements types.NormalizableValidator[Code]
var _ types.NormalizableValidator[Code] = Code("")

// Code represents a country code according to ISO 3166-1 alpha-2 standard.
// Code implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and treats an empty Code string as SQL NULL.
// See NullableCode for a nullable version of this type.
type Code string //#enum

// Valid returns true if the normalized Code is valid.
// See Normalized for the normalization process.
//
// Hand-written: accepts non-normalized input (lower case, AltCodes,
// country names) via Normalized, unlike the exact-match switch over every
// enum value that go-enum would generate.
//
// #custom
func (c Code) Valid() bool {
	_, err := c.Normalized()
	return err == nil
}

// ValidAndNormalized returns true if the Code is valid and already normalized.
func (c Code) ValidAndNormalized() bool {
	norm, err := c.Normalized()
	return err == nil && c == norm
}

// Validate returns an error if the normalized Code is not valid.
// See Normalized for the normalization process.
//
// Hand-written: returns the error from Normalized, which names the
// offending input, instead of the generic error that go-enum would
// generate.
//
// #custom
func (c Code) Validate() error {
	_, err := c.Normalized()
	return err
}

// Normalized uses the whitespace-trimmed uppercase string of the code to look up
// and return the standard ISO 3166-1 alpha-2 code.
// If not found, then AltCodes is used to look up alternative code and name mappings
// using the whitespace-trimmed uppercase code.
// If no mapping exists, then the original Code is returned unchanged together with an error.
func (c Code) Normalized() (Code, error) {
	norm := Code(strings.ToUpper(strutil.TrimSpace(string(c))))
	if _, ok := countryMap[norm]; ok {
		return norm, nil
	}
	if norm, ok := AltCodes[string(norm)]; ok {
		return norm, nil
	}
	return c, fmt.Errorf("invalid country.Code: '%s'", string(c))
}

// ParseCode parses and normalizes a country identifier given in any of
// the formats this package understands, returning the canonical ISO
// 3166-1 alpha-2 Code.
//
// It tries the cheapest and most common interpretations first and only
// falls back to the more expensive ones:
//
//  1. An input that is already a canonical ISO 3166-1 alpha-2 code
//     ("DE", "FR") is returned unchanged without any further work.
//  2. Otherwise the input is run through Normalized, which accepts case
//     and whitespace variants of alpha-2 codes plus the alternative
//     codes and German country names of AltCodes ("d", " AUT ", "SUI",
//     "Deutschland", "Österreich").
//  3. As a last resort the input is matched case-insensitively against
//     the English country names ("Germany", "united kingdom").
//
// ParseCode returns an error wrapping the original input if no ISO
// 3166-1 alpha-2 code can be derived. Like Normalized, the error
// preserves the unnormalized value for diagnostics.
func ParseCode(str string) (Code, error) {
	// 1. Short path: the input is already a canonical alpha-2 code.
	if _, ok := countryMap[Code(str)]; ok {
		return Code(str), nil
	}
	// 2. Codes and names handled by Normalized: case and whitespace
	//    variants, plus the alternative codes and German country names
	//    of AltCodes.
	c, err := Code(str).Normalized()
	if err == nil {
		return c, nil
	}
	// 3. English country name.
	if named, ok := codeForName(str); ok {
		return named, nil
	}
	// Reuse the input-preserving Code and error returned by Normalized.
	return c, err
}

// nameToCode is the lazily-built reverse index from a lower-cased English
// country name to its ISO 3166-1 alpha-2 Code. It is consulted only by
// ParseCode when the input is not a recognizable code, so it is built on
// first use rather than at package initialization.
var nameToCode = sync.OnceValue(func() map[string]Code {
	return newNameToCode(maps.Keys(countryMap))
})

// newNameToCode builds the name index for codes.
// EL, the European Union's code for Greece, is skipped because it
// shares the name "Greece" with GR: the name must resolve to the
// ISO 3166-1 code GR independent of the iteration order of codes.
func newNameToCode(codes iter.Seq[Code]) map[string]Code {
	m := make(map[string]Code, len(countryMap))
	for code := range codes {
		if code == EL {
			continue
		}
		m[strings.ToLower(countryMap[code])] = code
	}
	return m
}

// codeForName resolves an English country name to its ISO 3166-1 alpha-2
// Code, matching case-insensitively. The bool result reports whether a
// match was found.
func codeForName(name string) (Code, bool) {
	key := strings.ToLower(strutil.TrimSpace(name))
	if key == "" {
		return Invalid, false
	}
	c, ok := nameToCode()[key]
	return c, ok
}

// IsEU indicates if a country is a member of the European Union.
// EL, the European Union's own code for Greece, is a member like GR.
func (c Code) IsEU() bool {
	norm, err := c.Normalized()
	if err != nil {
		return false
	}
	_, ok := euCountries[norm]
	return ok
}

// EnglishName returns the English name of the country.
// Returns an empty string if the code is invalid.
func (c Code) EnglishName() string {
	norm, err := c.Normalized()
	if err != nil {
		return ""
	}
	return countryMap[norm]
}

// Scan implements the database/sql.Scanner interface.
func (c *Code) Scan(value any) error {
	switch x := value.(type) {
	case string:
		*c = Code(x)
	case []byte:
		*c = Code(x)
	case nil:
		*c = Invalid
	default:
		return fmt.Errorf("can't scan SQL value of type %T as country.Code", value)
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
// Returns nil for SQL NULL if the Code is Invalid.
func (c Code) Value() (driver.Value, error) {
	if c == Invalid {
		return nil, nil
	}
	norm, _ := c.Normalized()
	return string(norm), nil
}

// MarshalJSON implements encoding/json.Marshaler.
// Returns the normalized code as a JSON string.
func (c Code) MarshalJSON() ([]byte, error) {
	norm, _ := c.Normalized()
	return json.Marshal(string(norm))
}

// JSONSchema returns the JSON schema definition for the Code type.
func (Code) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title:   "ISO 3166-1 alpha 2 Country Code",
		Type:    "string",
		Pattern: "^[A-Z]{2}$",
	}
}

// ScanString tries to parse and assign the passed source string as value of the implementing type.
// If validate is true, the source string is checked for validity before assignment.
// If validate is false and the source string can still be assigned in some non-normalized way,
// it will be assigned without returning an error.
func (c *Code) ScanString(source string, validate bool) error {
	code, err := Code(source).Normalized()
	if err != nil {
		if validate {
			return err
		}
		code = Code(source)
	}
	*c = code
	return nil
}

// String returns the normalized code if possible,
// else it will be returned unchanged as string.
//
// String implements the fmt.Stringer interface.
//
// Hand-written: normalizes before returning, unlike the plain string
// conversion that go-enum would generate.
//
// #custom
func (c Code) String() string {
	norm, _ := c.Normalized()
	return string(norm)
}

// Nullable returns the Code as NullableCode.
// Country code Invalid is returned as Null.
func (c Code) Nullable() NullableCode {
	return NullableCode(c)
}

///////////////////////////////////////////////////////////////////////////////
// ISO 3166-1 alpha-2 country code constants
//
// These represent all officially recognized country codes and have to live
// in the same file as `type Code string //#enum` because go-enum collects
// the enum values of a type from the file that declares it.

const (
	AF Code = "AF" // Afghanistan
	AX Code = "AX" // Åland Islands
	AL Code = "AL" // Albania
	DZ Code = "DZ" // Algeria
	AS Code = "AS" // American Samoa
	AD Code = "AD" // Andorra
	AO Code = "AO" // Angola
	AI Code = "AI" // Anguilla
	AQ Code = "AQ" // Antarctica
	AG Code = "AG" // Antigua and Barbuda
	AR Code = "AR" // Argentina
	AM Code = "AM" // Armenia
	AW Code = "AW" // Aruba
	AU Code = "AU" // Australia
	AT Code = "AT" // Austria
	AZ Code = "AZ" // Azerbaijan
	BS Code = "BS" // Bahamas
	BH Code = "BH" // Bahrain
	BD Code = "BD" // Bangladesh
	BB Code = "BB" // Barbados
	BY Code = "BY" // Belarus
	BE Code = "BE" // Belgium
	BZ Code = "BZ" // Belize
	BJ Code = "BJ" // Benin
	BM Code = "BM" // Bermuda
	BT Code = "BT" // Bhutan
	BO Code = "BO" // Bolivia, Plurinational State of
	BQ Code = "BQ" // Bonaire, Sint Eustatius and Saba
	BA Code = "BA" // Bosnia and Herzegovina
	BW Code = "BW" // Botswana
	BV Code = "BV" // Bouvet Island
	BR Code = "BR" // Brazil
	IO Code = "IO" // British Indian Ocean Territory
	BN Code = "BN" // Brunei Darussalam
	BG Code = "BG" // Bulgaria
	BF Code = "BF" // Burkina Faso
	BI Code = "BI" // Burundi
	KH Code = "KH" // Cambodia
	CM Code = "CM" // Cameroon
	CA Code = "CA" // Canada
	CV Code = "CV" // Cape Verde
	KY Code = "KY" // Cayman Islands
	CF Code = "CF" // Central African Republic
	TD Code = "TD" // Chad
	CL Code = "CL" // Chile
	CN Code = "CN" // China
	CX Code = "CX" // Christmas Island
	CC Code = "CC" // Cocos (Keeling) Islands
	CO Code = "CO" // Colombia
	KM Code = "KM" // Comoros
	CG Code = "CG" // Congo
	CD Code = "CD" // Congo, the Democratic Republic of the
	CK Code = "CK" // Cook Islands
	CR Code = "CR" // Costa Rica
	CI Code = "CI" // Côte d'Ivoire
	HR Code = "HR" // Croatia
	CU Code = "CU" // Cuba
	CW Code = "CW" // Curaçao
	CY Code = "CY" // Cyprus
	CZ Code = "CZ" // Czech Republic
	DK Code = "DK" // Denmark
	DJ Code = "DJ" // Djibouti
	DM Code = "DM" // Dominica
	DO Code = "DO" // Dominican Republic
	EC Code = "EC" // Ecuador
	EG Code = "EG" // Egypt
	SV Code = "SV" // El Salvador
	GQ Code = "GQ" // Equatorial Guinea
	ER Code = "ER" // Eritrea
	EE Code = "EE" // Estonia
	ET Code = "ET" // Ethiopia
	FK Code = "FK" // Falkland Islands (Malvinas)
	FO Code = "FO" // Faroe Islands
	FJ Code = "FJ" // Fiji
	FI Code = "FI" // Finland
	FR Code = "FR" // France
	GF Code = "GF" // French Guiana
	PF Code = "PF" // French Polynesia
	TF Code = "TF" // French Southern Territories
	GA Code = "GA" // Gabon
	GM Code = "GM" // Gambia
	GE Code = "GE" // Georgia
	DE Code = "DE" // Germany
	GH Code = "GH" // Ghana
	GI Code = "GI" // Gibraltar
	GR Code = "GR" // Greece
	EL Code = "EL" // Greece (alternative)
	GL Code = "GL" // Greenland
	GD Code = "GD" // Grenada
	GP Code = "GP" // Guadeloupe
	GU Code = "GU" // Guam
	GT Code = "GT" // Guatemala
	GG Code = "GG" // Guernsey
	GN Code = "GN" // Guinea
	GW Code = "GW" // Guinea-Bissau
	GY Code = "GY" // Guyana
	HT Code = "HT" // Haiti
	HM Code = "HM" // Heard Island and McDonald Islands
	VA Code = "VA" // Holy See (Vatican City State)
	HN Code = "HN" // Honduras
	HK Code = "HK" // Hong Kong
	HU Code = "HU" // Hungary
	IS Code = "IS" // Iceland
	IN Code = "IN" // India
	ID Code = "ID" // Indonesia
	IR Code = "IR" // Iran, Islamic Republic of
	IQ Code = "IQ" // Iraq
	IE Code = "IE" // Ireland
	IM Code = "IM" // Isle of Man
	IL Code = "IL" // Israel
	IT Code = "IT" // Italy
	JM Code = "JM" // Jamaica
	JP Code = "JP" // Japan
	JE Code = "JE" // Jersey
	JO Code = "JO" // Jordan
	KZ Code = "KZ" // Kazakhstan
	KE Code = "KE" // Kenya
	KI Code = "KI" // Kiribati
	KP Code = "KP" // Korea, Democratic People's Republic of
	KR Code = "KR" // Korea, Republic of
	KW Code = "KW" // Kuwait
	KG Code = "KG" // Kyrgyzstan
	LA Code = "LA" // Lao People's Democratic Republic
	LV Code = "LV" // Latvia
	LB Code = "LB" // Lebanon
	LS Code = "LS" // Lesotho
	LR Code = "LR" // Liberia
	LY Code = "LY" // Libya
	LI Code = "LI" // Liechtenstein
	LT Code = "LT" // Lithuania
	LU Code = "LU" // Luxembourg
	MO Code = "MO" // Macao
	MK Code = "MK" // Macedonia, the Former Yugoslav Republic of
	MG Code = "MG" // Madagascar
	MW Code = "MW" // Malawi
	MY Code = "MY" // Malaysia
	MV Code = "MV" // Maldives
	ML Code = "ML" // Mali
	MT Code = "MT" // Malta
	MH Code = "MH" // Marshall Islands
	MQ Code = "MQ" // Martinique
	MR Code = "MR" // Mauritania
	MU Code = "MU" // Mauritius
	YT Code = "YT" // Mayotte
	MX Code = "MX" // Mexico
	FM Code = "FM" // Micronesia, Federated States of
	MD Code = "MD" // Moldova, Republic of
	MC Code = "MC" // Monaco
	MN Code = "MN" // Mongolia
	ME Code = "ME" // Montenegro
	MS Code = "MS" // Montserrat
	MA Code = "MA" // Morocco
	MZ Code = "MZ" // Mozambique
	MM Code = "MM" // Myanmar
	NA Code = "NA" // Namibia
	NR Code = "NR" // Nauru
	NP Code = "NP" // Nepal
	NL Code = "NL" // Netherlands
	NC Code = "NC" // New Caledonia
	NZ Code = "NZ" // New Zealand
	NI Code = "NI" // Nicaragua
	NE Code = "NE" // Niger
	NG Code = "NG" // Nigeria
	NU Code = "NU" // Niue
	NF Code = "NF" // Norfolk Island
	MP Code = "MP" // Northern Mariana Islands
	NO Code = "NO" // Norway
	OM Code = "OM" // Oman
	PK Code = "PK" // Pakistan
	PW Code = "PW" // Palau
	PS Code = "PS" // Palestine, State of
	PA Code = "PA" // Panama
	PG Code = "PG" // Papua New Guinea
	PY Code = "PY" // Paraguay
	PE Code = "PE" // Peru
	PH Code = "PH" // Philippines
	PN Code = "PN" // Pitcairn
	PL Code = "PL" // Poland
	PT Code = "PT" // Portugal
	PR Code = "PR" // Puerto Rico
	QA Code = "QA" // Qatar
	RE Code = "RE" // Réunion
	RO Code = "RO" // Romania
	RU Code = "RU" // Russian Federation
	RW Code = "RW" // Rwanda
	BL Code = "BL" // Saint Barthélemy
	SH Code = "SH" // Saint Helena, Ascension and Tristan da Cunha
	KN Code = "KN" // Saint Kitts and Nevis
	LC Code = "LC" // Saint Lucia
	MF Code = "MF" // Saint Martin (French part)
	PM Code = "PM" // Saint Pierre and Miquelon
	VC Code = "VC" // Saint Vincent and the Grenadines
	WS Code = "WS" // Samoa
	SM Code = "SM" // San Marino
	ST Code = "ST" // Sao Tome and Principe
	SA Code = "SA" // Saudi Arabia
	SN Code = "SN" // Senegal
	RS Code = "RS" // Serbia
	SC Code = "SC" // Seychelles
	SL Code = "SL" // Sierra Leone
	SG Code = "SG" // Singapore
	SX Code = "SX" // Sint Maarten (Dutch part)
	SK Code = "SK" // Slovakia
	SI Code = "SI" // Slovenia
	SB Code = "SB" // Solomon Islands
	SO Code = "SO" // Somalia
	ZA Code = "ZA" // South Africa
	GS Code = "GS" // South Georgia and the South Sandwich Islands
	SS Code = "SS" // South Sudan
	ES Code = "ES" // Spain
	LK Code = "LK" // Sri Lanka
	SD Code = "SD" // Sudan
	SR Code = "SR" // Suriname
	SJ Code = "SJ" // Svalbard and Jan Mayen
	SZ Code = "SZ" // Swaziland
	SE Code = "SE" // Sweden
	CH Code = "CH" // Switzerland
	SY Code = "SY" // Syrian Arab Republic
	TW Code = "TW" // Taiwan, Province of China
	TJ Code = "TJ" // Tajikistan
	TZ Code = "TZ" // Tanzania, United Republic of
	TH Code = "TH" // Thailand
	TL Code = "TL" // Timor-Leste
	TG Code = "TG" // Togo
	TK Code = "TK" // Tokelau
	TO Code = "TO" // Tonga
	TT Code = "TT" // Trinidad and Tobago
	TN Code = "TN" // Tunisia
	TR Code = "TR" // Turkey
	TM Code = "TM" // Turkmenistan
	TC Code = "TC" // Turks and Caicos Islands
	TV Code = "TV" // Tuvalu
	UG Code = "UG" // Uganda
	UA Code = "UA" // Ukraine
	AE Code = "AE" // United Arab Emirates
	GB Code = "GB" // United Kingdom
	US Code = "US" // United States
	UM Code = "UM" // United States Minor Outlying Islands
	UY Code = "UY" // Uruguay
	UZ Code = "UZ" // Uzbekistan
	VU Code = "VU" // Vanuatu
	VE Code = "VE" // Venezuela, Bolivarian Republic of
	VN Code = "VN" // Viet Nam
	VG Code = "VG" // Virgin Islands, British
	VI Code = "VI" // Virgin Islands, U.S.
	WF Code = "WF" // Wallis and Futuna
	EH Code = "EH" // Western Sahara
	YE Code = "YE" // Yemen
	ZM Code = "ZM" // Zambia
	ZW Code = "ZW" // Zimbabwe
	XK Code = "XK" // Republic of Kosovo (unofficial, but still can be used)
)

// Enums returns all valid values for Code
func (Code) Enums() []Code {
	return []Code{
		AF,
		AX,
		AL,
		DZ,
		AS,
		AD,
		AO,
		AI,
		AQ,
		AG,
		AR,
		AM,
		AW,
		AU,
		AT,
		AZ,
		BS,
		BH,
		BD,
		BB,
		BY,
		BE,
		BZ,
		BJ,
		BM,
		BT,
		BO,
		BQ,
		BA,
		BW,
		BV,
		BR,
		IO,
		BN,
		BG,
		BF,
		BI,
		KH,
		CM,
		CA,
		CV,
		KY,
		CF,
		TD,
		CL,
		CN,
		CX,
		CC,
		CO,
		KM,
		CG,
		CD,
		CK,
		CR,
		CI,
		HR,
		CU,
		CW,
		CY,
		CZ,
		DK,
		DJ,
		DM,
		DO,
		EC,
		EG,
		SV,
		GQ,
		ER,
		EE,
		ET,
		FK,
		FO,
		FJ,
		FI,
		FR,
		GF,
		PF,
		TF,
		GA,
		GM,
		GE,
		DE,
		GH,
		GI,
		GR,
		EL,
		GL,
		GD,
		GP,
		GU,
		GT,
		GG,
		GN,
		GW,
		GY,
		HT,
		HM,
		VA,
		HN,
		HK,
		HU,
		IS,
		IN,
		ID,
		IR,
		IQ,
		IE,
		IM,
		IL,
		IT,
		JM,
		JP,
		JE,
		JO,
		KZ,
		KE,
		KI,
		KP,
		KR,
		KW,
		KG,
		LA,
		LV,
		LB,
		LS,
		LR,
		LY,
		LI,
		LT,
		LU,
		MO,
		MK,
		MG,
		MW,
		MY,
		MV,
		ML,
		MT,
		MH,
		MQ,
		MR,
		MU,
		YT,
		MX,
		FM,
		MD,
		MC,
		MN,
		ME,
		MS,
		MA,
		MZ,
		MM,
		NA,
		NR,
		NP,
		NL,
		NC,
		NZ,
		NI,
		NE,
		NG,
		NU,
		NF,
		MP,
		NO,
		OM,
		PK,
		PW,
		PS,
		PA,
		PG,
		PY,
		PE,
		PH,
		PN,
		PL,
		PT,
		PR,
		QA,
		RE,
		RO,
		RU,
		RW,
		BL,
		SH,
		KN,
		LC,
		MF,
		PM,
		VC,
		WS,
		SM,
		ST,
		SA,
		SN,
		RS,
		SC,
		SL,
		SG,
		SX,
		SK,
		SI,
		SB,
		SO,
		ZA,
		GS,
		SS,
		ES,
		LK,
		SD,
		SR,
		SJ,
		SZ,
		SE,
		CH,
		SY,
		TW,
		TJ,
		TZ,
		TH,
		TL,
		TG,
		TK,
		TO,
		TT,
		TN,
		TR,
		TM,
		TC,
		TV,
		UG,
		UA,
		AE,
		GB,
		US,
		UM,
		UY,
		UZ,
		VU,
		VE,
		VN,
		VG,
		VI,
		WF,
		EH,
		YE,
		ZM,
		ZW,
		XK,
	}
}

// EnumStrings returns all valid values for Code as strings
func (Code) EnumStrings() []string {
	return []string{
		"AF",
		"AX",
		"AL",
		"DZ",
		"AS",
		"AD",
		"AO",
		"AI",
		"AQ",
		"AG",
		"AR",
		"AM",
		"AW",
		"AU",
		"AT",
		"AZ",
		"BS",
		"BH",
		"BD",
		"BB",
		"BY",
		"BE",
		"BZ",
		"BJ",
		"BM",
		"BT",
		"BO",
		"BQ",
		"BA",
		"BW",
		"BV",
		"BR",
		"IO",
		"BN",
		"BG",
		"BF",
		"BI",
		"KH",
		"CM",
		"CA",
		"CV",
		"KY",
		"CF",
		"TD",
		"CL",
		"CN",
		"CX",
		"CC",
		"CO",
		"KM",
		"CG",
		"CD",
		"CK",
		"CR",
		"CI",
		"HR",
		"CU",
		"CW",
		"CY",
		"CZ",
		"DK",
		"DJ",
		"DM",
		"DO",
		"EC",
		"EG",
		"SV",
		"GQ",
		"ER",
		"EE",
		"ET",
		"FK",
		"FO",
		"FJ",
		"FI",
		"FR",
		"GF",
		"PF",
		"TF",
		"GA",
		"GM",
		"GE",
		"DE",
		"GH",
		"GI",
		"GR",
		"EL",
		"GL",
		"GD",
		"GP",
		"GU",
		"GT",
		"GG",
		"GN",
		"GW",
		"GY",
		"HT",
		"HM",
		"VA",
		"HN",
		"HK",
		"HU",
		"IS",
		"IN",
		"ID",
		"IR",
		"IQ",
		"IE",
		"IM",
		"IL",
		"IT",
		"JM",
		"JP",
		"JE",
		"JO",
		"KZ",
		"KE",
		"KI",
		"KP",
		"KR",
		"KW",
		"KG",
		"LA",
		"LV",
		"LB",
		"LS",
		"LR",
		"LY",
		"LI",
		"LT",
		"LU",
		"MO",
		"MK",
		"MG",
		"MW",
		"MY",
		"MV",
		"ML",
		"MT",
		"MH",
		"MQ",
		"MR",
		"MU",
		"YT",
		"MX",
		"FM",
		"MD",
		"MC",
		"MN",
		"ME",
		"MS",
		"MA",
		"MZ",
		"MM",
		"NA",
		"NR",
		"NP",
		"NL",
		"NC",
		"NZ",
		"NI",
		"NE",
		"NG",
		"NU",
		"NF",
		"MP",
		"NO",
		"OM",
		"PK",
		"PW",
		"PS",
		"PA",
		"PG",
		"PY",
		"PE",
		"PH",
		"PN",
		"PL",
		"PT",
		"PR",
		"QA",
		"RE",
		"RO",
		"RU",
		"RW",
		"BL",
		"SH",
		"KN",
		"LC",
		"MF",
		"PM",
		"VC",
		"WS",
		"SM",
		"ST",
		"SA",
		"SN",
		"RS",
		"SC",
		"SL",
		"SG",
		"SX",
		"SK",
		"SI",
		"SB",
		"SO",
		"ZA",
		"GS",
		"SS",
		"ES",
		"LK",
		"SD",
		"SR",
		"SJ",
		"SZ",
		"SE",
		"CH",
		"SY",
		"TW",
		"TJ",
		"TZ",
		"TH",
		"TL",
		"TG",
		"TK",
		"TO",
		"TT",
		"TN",
		"TR",
		"TM",
		"TC",
		"TV",
		"UG",
		"UA",
		"AE",
		"GB",
		"US",
		"UM",
		"UY",
		"UZ",
		"VU",
		"VE",
		"VN",
		"VG",
		"VI",
		"WF",
		"EH",
		"YE",
		"ZM",
		"ZW",
		"XK",
	}
}
