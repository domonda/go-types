package language

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"sync"

	"github.com/invopop/jsonschema"

	"github.com/domonda/go-types"
	"github.com/domonda/go-types/strutil"
)

// Code represents a language code in its normalized form as an ISO 639-1 two-character language code.
// Code implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and treats an empty Code string as SQL NULL value.
type Code string //#enum

// Null is an empty string and will be treated as SQL NULL.
// language.Null.Valid() == false
//
// Declared as a conversion instead of `Null Code = ""` so that go-enum
// does not pick it up as one of the Code enum values: Null is not a valid
// language code and must stay out of Code.Enums() and Code.EnumStrings().
const Null = Code("")

// Compile-time check that Code implements types.NormalizableValidator[Code]
var _ types.NormalizableValidator[Code] = Code("")

// Valid returns true if the Code is a valid ISO 639-1 language code.
//
// Hand-written: a map lookup over all ISO 639-1 codes instead of the
// switch over every enum value that go-enum would generate.
//
// #custom
func (c Code) Valid() bool {
	_, ok := codeNames[c]
	return ok
}

// Validate returns an error if c is not a valid ISO 639-1 language code,
// exactly like Valid. Value and encoding/json store c as given, so
// validation only accepts the normalized code. Use Normalized to accept
// other spellings like "EN", "eng" and "en-US".
//
// Hand-written: returns an error that names the offending input
// instead of the generic error that go-enum would generate.
//
// #custom
func (c Code) Validate() error {
	if !c.Valid() {
		return fmt.Errorf("invalid language.Code: %q", string(c))
	}
	return nil
}

// ValidAndNormalized returns true if the Code is valid and already normalized.
func (c Code) ValidAndNormalized() bool {
	norm, err := c.Normalized()
	return err == nil && c == norm
}

// Normalized returns the normalized ISO 639-1 two-letter language code.
//
// Accepted input shapes:
//   - ISO 639-1 two-letter codes, case-insensitive, with surrounding
//     whitespace trimmed ("en", " EN ", "En").
//   - ISO 639-2 three-letter codes, both the bibliographic (B) and
//     terminologic (T) variants where they differ ("eng", "deu", "ger").
//   - ISO 639-3 three-letter codes, where a 639-1 equivalent exists
//     ("eng", "deu", "fra"). 639-3 codes for languages without a 639-1
//     assignment are rejected.
//   - BCP-47 language tags: only the leading language subtag is
//     consulted; script, region, variant, and extension subtags are
//     dropped ("en-US" → "en", "zh-Hant-CN" → "zh", "sr-Latn" → "sr").
//     POSIX-style underscores ("en_US") are accepted as separators.
//
// Returns an error wrapping the original input if no 639-1 code can be
// derived. The error preserves the unnormalized value for diagnostics.
//
// See:
//   - https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes
//   - https://www.rfc-editor.org/info/bcp47
//   - https://www.unicode.org/reports/tr35/#Unicode_locale_identifier
func (c Code) Normalized() (Code, error) {
	s := strings.ToLower(strutil.TrimSpace(string(c)))
	if s == "" {
		return c, fmt.Errorf("invalid language.Code: %q", string(c))
	}
	// Strip BCP-47 region/script/variant subtags by keeping only the
	// primary language subtag. POSIX locale strings ("en_US") use '_'.
	if i := strings.IndexAny(s, "-_"); i >= 0 {
		s = s[:i]
	}
	switch len(s) {
	case 2:
		if _, ok := codeNames[Code(s)]; ok {
			return Code(s), nil
		}
	case 3:
		if code, ok := iso6393To1[s]; ok {
			return code, nil
		}
	}
	return c, fmt.Errorf("invalid language.Code: %q", string(c))
}

// ParseCode parses and normalizes a language identifier given in any of
// the formats this package understands, returning the canonical ISO
// 639-1 two-letter Code.
//
// It tries the cheapest and most common interpretations first and only
// falls back to the more expensive ones:
//
//  1. An input that is already a canonical ISO 639-1 code ("en", "de")
//     is returned unchanged without any further work.
//  2. Otherwise the input is run through Normalized, which accepts case
//     and whitespace variants, ISO 639-2/T and 639-2/B codes, ISO 639-3
//     codes with a 639-1 equivalent, BCP-47 tags, and POSIX locale
//     strings ("EN", " eng ", "ger", "en-US", "zh_Hant").
//  3. As a last resort the input is matched case-insensitively against
//     the English language names of the ISO 639-1 and ISO 639-3 tables
//     ("German", "english", "Catalan").
//
// ParseCode returns an error wrapping the original input if no ISO 639-1
// code can be derived. Like Normalized, the error preserves the
// unnormalized value for diagnostics.
func ParseCode(str string) (Code, error) {
	// 1. Short path: the input is already a canonical ISO 639-1 code.
	if c := Code(str); c.Valid() {
		return c, nil
	}
	// 2. Codes and tags: case and whitespace variants, ISO 639-2/3,
	//    639-2/B bibliographic variants, BCP-47 tags, POSIX locales.
	c, err := Code(str).Normalized()
	if err == nil {
		return c, nil
	}
	// 3. English language name.
	if named, ok := codeForName(str); ok {
		return named, nil
	}
	// Reuse the input-preserving Code and error returned by Normalized.
	return c, err
}

// nameToCode is the lazily-built reverse index from a lower-cased English
// language name to its ISO 639-1 Code. It is consulted only by ParseCode
// when the input is not a recognizable code or tag, so it is built on
// first use rather than at package initialization.
var nameToCode = sync.OnceValue(func() map[string]Code {
	m := make(map[string]Code, len(codeNames)+len(iso6393Names))
	// Index the ISO 639-3 names first so the curated ISO 639-1 names
	// added below take precedence on any name collision.
	for code3, name := range iso6393Names {
		if code1, ok := iso6393To1[code3]; ok {
			m[strings.ToLower(name)] = code1
		}
	}
	for code, names := range codeNames {
		// A codeNames value may list synonyms separated by "; ".
		for name := range strings.SplitSeq(names, ";") {
			if name = strings.ToLower(strings.TrimSpace(name)); name != "" {
				m[name] = code
			}
		}
	}
	return m
})

// codeForName resolves an English language name to its ISO 639-1 Code,
// matching case-insensitively. The bool result reports whether a match
// was found.
func codeForName(name string) (Code, bool) {
	key := strings.ToLower(strutil.TrimSpace(name))
	if key == "" {
		return Null, false
	}
	c, ok := nameToCode()[key]
	return c, ok
}

// LanguageName returns the English name of the language for the code.
// The code is normalized first, see Normalized.
// Returns an empty string if the code is invalid.
func (c Code) LanguageName() string {
	norm, err := c.Normalized()
	if err != nil {
		return ""
	}
	return codeNames[norm]
}

// String returns the normalized code if possible, else it will be returned unchanged as string.
// String implements the fmt.Stringer interface.
//
// Hand-written: normalizes before returning, unlike the plain string
// conversion that go-enum would generate.
//
// #custom
func (c Code) String() string {
	norm, err := c.Normalized()
	if err != nil {
		return string(c)
	}
	return string(norm)
}

// Scan implements the database/sql.Scanner interface.
func (c *Code) Scan(value any) error {
	switch x := value.(type) {
	case string:
		*c = Code(x)
	case []byte:
		*c = Code(x)
	case nil:
		*c = Null
	default:
		return fmt.Errorf("can't scan SQL value of type %T as language.Code", value)
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (c Code) Value() (driver.Value, error) {
	if c == Null {
		return nil, nil
	}
	return string(c), nil
}

// JSONSchema returns the JSON schema definition for the Code type.
func (Code) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Title:   "Language Code",
		Type:    "string",
		Pattern: `^[a-z]{2}$`,
	}
}

///////////////////////////////////////////////////////////////////////////////
// ISO 639-1 language code constants
//
// Source: https://iso639-3.sil.org/code_tables/download_tables
//
// These have to live in the same file as `type Code string //#enum` because
// go-enum collects the enum values of a type from the file that declares it.

const (
	AA Code = "aa" // Afar
	AB Code = "ab" // Abkhazian
	AF Code = "af" // Afrikaans
	AK Code = "ak" // Akan
	SQ Code = "sq" // Albanian
	AM Code = "am" // Amharic
	AR Code = "ar" // Arabic
	AN Code = "an" // Aragonese
	HY Code = "hy" // Armenian
	AS Code = "as" // Assamese
	AV Code = "av" // Avaric
	AE Code = "ae" // Avestan
	AY Code = "ay" // Aymara
	AZ Code = "az" // Azerbaijani
	BA Code = "ba" // Bashkir
	BM Code = "bm" // Bambara
	EU Code = "eu" // Basque
	BE Code = "be" // Belarusian
	BN Code = "bn" // Bengali
	BH Code = "bh" // Bihari languages
	BI Code = "bi" // Bislama
	BS Code = "bs" // Bosnian
	BR Code = "br" // Breton
	BG Code = "bg" // Bulgarian
	MY Code = "my" // Burmese
	CA Code = "ca" // Catalan; Valencian
	CH Code = "ch" // Chamorro
	CE Code = "ce" // Chechen
	ZH Code = "zh" // Chinese
	CU Code = "cu" // Church Slavic; Old Slavonic; Church Slavonic; Old Bulgarian; Old Church Slavonic
	CV Code = "cv" // Chuvash
	KW Code = "kw" // Cornish
	CO Code = "co" // Corsican
	CR Code = "cr" // Cree
	CS Code = "cs" // Czech
	DA Code = "da" // Danish
	DV Code = "dv" // Divehi; Dhivehi; Maldivian
	NL Code = "nl" // Dutch; Flemish
	DZ Code = "dz" // Dzongkha
	EN Code = "en" // English
	EO Code = "eo" // Esperanto
	ET Code = "et" // Estonian
	EE Code = "ee" // Ewe
	FO Code = "fo" // Faroese
	FJ Code = "fj" // Fijian
	FI Code = "fi" // Finnish
	FR Code = "fr" // French
	FY Code = "fy" // Western Frisian
	FF Code = "ff" // Fulah
	KA Code = "ka" // Georgian
	DE Code = "de" // German
	GD Code = "gd" // Gaelic; Scottish Gaelic
	GA Code = "ga" // Irish
	GL Code = "gl" // Galician
	GV Code = "gv" // Manx
	EL Code = "el" // Greek, Modern (1453-)
	GN Code = "gn" // Guarani
	GU Code = "gu" // Gujarati
	HT Code = "ht" // Haitian; Haitian Creole
	HA Code = "ha" // Hausa
	HE Code = "he" // Hebrew
	HZ Code = "hz" // Herero
	HI Code = "hi" // Hindi
	HO Code = "ho" // Hiri Motu
	HR Code = "hr" // Croatian
	HU Code = "hu" // Hungarian
	IG Code = "ig" // Igbo
	IS Code = "is" // Icelandic
	IO Code = "io" // Ido
	II Code = "ii" // Sichuan Yi; Nuosu
	IU Code = "iu" // Inuktitut
	IE Code = "ie" // Interlingue; Occidental
	IA Code = "ia" // Interlingua (International Auxiliary Language Association)
	ID Code = "id" // Indonesian
	IK Code = "ik" // Inupiaq
	IT Code = "it" // Italian
	JV Code = "jv" // Javanese
	JA Code = "ja" // Japanese
	KL Code = "kl" // Kalaallisut; Greenlandic
	KN Code = "kn" // Kannada
	KS Code = "ks" // Kashmiri
	KR Code = "kr" // Kanuri
	KK Code = "kk" // Kazakh
	KM Code = "km" // Central Khmer
	KI Code = "ki" // Kikuyu; Gikuyu
	RW Code = "rw" // Kinyarwanda
	KY Code = "ky" // Kirghiz; Kyrgyz
	KV Code = "kv" // Komi
	KG Code = "kg" // Kongo
	KO Code = "ko" // Korean
	KJ Code = "kj" // Kuanyama; Kwanyama
	KU Code = "ku" // Kurdish
	LO Code = "lo" // Lao
	LA Code = "la" // Latin
	LV Code = "lv" // Latvian
	LI Code = "li" // Limburgan; Limburger; Limburgish
	LN Code = "ln" // Lingala
	LT Code = "lt" // Lithuanian
	LB Code = "lb" // Luxembourgish; Letzeburgesch
	LU Code = "lu" // Luba-Katanga
	LG Code = "lg" // Ganda
	MK Code = "mk" // Macedonian
	MH Code = "mh" // Marshallese
	ML Code = "ml" // Malayalam
	MI Code = "mi" // Maori
	MR Code = "mr" // Marathi
	MS Code = "ms" // Malay
	MG Code = "mg" // Malagasy
	MT Code = "mt" // Maltese
	MN Code = "mn" // Mongolian
	NA Code = "na" // Nauru
	NV Code = "nv" // Navajo; Navaho
	NR Code = "nr" // Ndebele, South; South Ndebele
	ND Code = "nd" // Ndebele, North; North Ndebele
	NG Code = "ng" // Ndonga
	NE Code = "ne" // Nepali
	NN Code = "nn" // Norwegian Nynorsk; Nynorsk, Norwegian
	NB Code = "nb" // Bokmål, Norwegian; Norwegian Bokmål
	NO Code = "no" // Norwegian
	NY Code = "ny" // Chichewa; Chewa; Nyanja
	OC Code = "oc" // Occitan (post 1500); Provençal
	OJ Code = "oj" // Ojibwa
	OR Code = "or" // Oriya
	OM Code = "om" // Oromo
	OS Code = "os" // Ossetian; Ossetic
	PA Code = "pa" // Panjabi; Punjabi
	FA Code = "fa" // Persian
	PI Code = "pi" // Pali
	PL Code = "pl" // Polish
	PT Code = "pt" // Portuguese
	PS Code = "ps" // Pushto; Pashto
	QU Code = "qu" // Quechua
	RM Code = "rm" // Romansh
	RO Code = "ro" // Romanian; Moldavian; Moldovan
	RN Code = "rn" // Rundi
	RU Code = "ru" // Russian
	SG Code = "sg" // Sango
	SA Code = "sa" // Sanskrit
	SI Code = "si" // Sinhala; Sinhalese
	SK Code = "sk" // Slovak
	SL Code = "sl" // Slovenian
	SE Code = "se" // Northern Sami
	SM Code = "sm" // Samoan
	SN Code = "sn" // Shona
	SD Code = "sd" // Sindhi
	SO Code = "so" // Somali
	ST Code = "st" // Sotho, Southern
	ES Code = "es" // Spanish; Castilian
	SC Code = "sc" // Sardinian
	SR Code = "sr" // Serbian
	SS Code = "ss" // Swati
	SU Code = "su" // Sundanese
	SW Code = "sw" // Swahili
	SV Code = "sv" // Swedish
	TY Code = "ty" // Tahitian
	TA Code = "ta" // Tamil
	TT Code = "tt" // Tatar
	TE Code = "te" // Telugu
	TG Code = "tg" // Tajik
	TL Code = "tl" // Tagalog
	TH Code = "th" // Thai
	BO Code = "bo" // Tibetan
	TI Code = "ti" // Tigrinya
	TO Code = "to" // Tonga (Tonga Islands)
	TN Code = "tn" // Tswana
	TS Code = "ts" // Tsonga
	TK Code = "tk" // Turkmen
	TR Code = "tr" // Turkish
	TW Code = "tw" // Twi
	UG Code = "ug" // Uighur; Uyghur
	UK Code = "uk" // Ukrainian
	UR Code = "ur" // Urdu
	UZ Code = "uz" // Uzbek
	VE Code = "ve" // Venda
	VI Code = "vi" // Vietnamese
	VO Code = "vo" // Volapük
	CY Code = "cy" // Welsh
	WA Code = "wa" // Walloon
	WO Code = "wo" // Wolof
	XH Code = "xh" // Xhosa
	YI Code = "yi" // Yiddish
	YO Code = "yo" // Yoruba
	ZA Code = "za" // Zhuang; Chuang
	ZU Code = "zu" // Zulu
)

// Enums returns all valid values for Code
func (Code) Enums() []Code {
	return []Code{
		AA,
		AB,
		AF,
		AK,
		SQ,
		AM,
		AR,
		AN,
		HY,
		AS,
		AV,
		AE,
		AY,
		AZ,
		BA,
		BM,
		EU,
		BE,
		BN,
		BH,
		BI,
		BS,
		BR,
		BG,
		MY,
		CA,
		CH,
		CE,
		ZH,
		CU,
		CV,
		KW,
		CO,
		CR,
		CS,
		DA,
		DV,
		NL,
		DZ,
		EN,
		EO,
		ET,
		EE,
		FO,
		FJ,
		FI,
		FR,
		FY,
		FF,
		KA,
		DE,
		GD,
		GA,
		GL,
		GV,
		EL,
		GN,
		GU,
		HT,
		HA,
		HE,
		HZ,
		HI,
		HO,
		HR,
		HU,
		IG,
		IS,
		IO,
		II,
		IU,
		IE,
		IA,
		ID,
		IK,
		IT,
		JV,
		JA,
		KL,
		KN,
		KS,
		KR,
		KK,
		KM,
		KI,
		RW,
		KY,
		KV,
		KG,
		KO,
		KJ,
		KU,
		LO,
		LA,
		LV,
		LI,
		LN,
		LT,
		LB,
		LU,
		LG,
		MK,
		MH,
		ML,
		MI,
		MR,
		MS,
		MG,
		MT,
		MN,
		NA,
		NV,
		NR,
		ND,
		NG,
		NE,
		NN,
		NB,
		NO,
		NY,
		OC,
		OJ,
		OR,
		OM,
		OS,
		PA,
		FA,
		PI,
		PL,
		PT,
		PS,
		QU,
		RM,
		RO,
		RN,
		RU,
		SG,
		SA,
		SI,
		SK,
		SL,
		SE,
		SM,
		SN,
		SD,
		SO,
		ST,
		ES,
		SC,
		SR,
		SS,
		SU,
		SW,
		SV,
		TY,
		TA,
		TT,
		TE,
		TG,
		TL,
		TH,
		BO,
		TI,
		TO,
		TN,
		TS,
		TK,
		TR,
		TW,
		UG,
		UK,
		UR,
		UZ,
		VE,
		VI,
		VO,
		CY,
		WA,
		WO,
		XH,
		YI,
		YO,
		ZA,
		ZU,
	}
}

// EnumStrings returns all valid values for Code as strings
func (Code) EnumStrings() []string {
	return []string{
		"aa",
		"ab",
		"af",
		"ak",
		"sq",
		"am",
		"ar",
		"an",
		"hy",
		"as",
		"av",
		"ae",
		"ay",
		"az",
		"ba",
		"bm",
		"eu",
		"be",
		"bn",
		"bh",
		"bi",
		"bs",
		"br",
		"bg",
		"my",
		"ca",
		"ch",
		"ce",
		"zh",
		"cu",
		"cv",
		"kw",
		"co",
		"cr",
		"cs",
		"da",
		"dv",
		"nl",
		"dz",
		"en",
		"eo",
		"et",
		"ee",
		"fo",
		"fj",
		"fi",
		"fr",
		"fy",
		"ff",
		"ka",
		"de",
		"gd",
		"ga",
		"gl",
		"gv",
		"el",
		"gn",
		"gu",
		"ht",
		"ha",
		"he",
		"hz",
		"hi",
		"ho",
		"hr",
		"hu",
		"ig",
		"is",
		"io",
		"ii",
		"iu",
		"ie",
		"ia",
		"id",
		"ik",
		"it",
		"jv",
		"ja",
		"kl",
		"kn",
		"ks",
		"kr",
		"kk",
		"km",
		"ki",
		"rw",
		"ky",
		"kv",
		"kg",
		"ko",
		"kj",
		"ku",
		"lo",
		"la",
		"lv",
		"li",
		"ln",
		"lt",
		"lb",
		"lu",
		"lg",
		"mk",
		"mh",
		"ml",
		"mi",
		"mr",
		"ms",
		"mg",
		"mt",
		"mn",
		"na",
		"nv",
		"nr",
		"nd",
		"ng",
		"ne",
		"nn",
		"nb",
		"no",
		"ny",
		"oc",
		"oj",
		"or",
		"om",
		"os",
		"pa",
		"fa",
		"pi",
		"pl",
		"pt",
		"ps",
		"qu",
		"rm",
		"ro",
		"rn",
		"ru",
		"sg",
		"sa",
		"si",
		"sk",
		"sl",
		"se",
		"sm",
		"sn",
		"sd",
		"so",
		"st",
		"es",
		"sc",
		"sr",
		"ss",
		"su",
		"sw",
		"sv",
		"ty",
		"ta",
		"tt",
		"te",
		"tg",
		"tl",
		"th",
		"bo",
		"ti",
		"to",
		"tn",
		"ts",
		"tk",
		"tr",
		"tw",
		"ug",
		"uk",
		"ur",
		"uz",
		"ve",
		"vi",
		"vo",
		"cy",
		"wa",
		"wo",
		"xh",
		"yi",
		"yo",
		"za",
		"zu",
	}
}
