package language

import (
	"database/sql/driver"
	"strings"

	"github.com/guregu/null"
)

// Code according to ISO 639-1 Code.
// Code implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and will treat an empty string Code as SQL NULL value.
type Code string

func (lc Code) Valid() bool {
	_, ok := codeNames[lc]
	return ok
}

func (lc Code) Normalized() Code {
	// TODO normalize 3 letter codes https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes
	// TODO normalize BCP-47 language codes, such as "en-US" or "sr-Latn"
	// http://www.unicode.org/reports/tr35/#Unicode_locale_identifier.
	normalized := Code(strings.ToLower(string(lc)))
	if _, ok := codeNames[normalized]; !ok {
		return ""
	}
	return normalized
}

func (lc Code) LanguageName() string {
	return codeNames[lc]
}

// Scan implements the database/sql.Scanner interface.
func (lc *Code) Scan(value interface{}) error {
	var ns null.String
	err := ns.Scan(value)
	if err != nil {
		return err
	}
	if ns.Valid {
		*lc = Code(ns.String)
	} else {
		*lc = ""
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (lc Code) Value() (driver.Value, error) {
	if lc == "" {
		return nil, nil
	}
	return string(lc), nil
}

var codeNames = map[Code]string{
	"aa": "Afar",
	"ab": "Abkhazian",
	"af": "Afrikaans",
	"ak": "Akan",
	"sq": "Albanian",
	"am": "Amharic",
	"ar": "Arabic",
	"an": "Aragonese",
	"hy": "Armenian",
	"as": "Assamese",
	"av": "Avaric",
	"ae": "Avestan",
	"ay": "Aymara",
	"az": "Azerbaijani",
	"ba": "Bashkir",
	"bm": "Bambara",
	"eu": "Basque",
	"be": "Belarusian",
	"bn": "Bengali",
	"bh": "Bihari languages",
	"bi": "Bislama",
	"bs": "Bosnian",
	"br": "Breton",
	"bg": "Bulgarian",
	"my": "Burmese",
	"ca": "Catalan; Valencian",
	"ch": "Chamorro",
	"ce": "Chechen",
	"zh": "Chinese",
	"cu": "Church Slavic; Old Slavonic; Church Slavonic; Old Bulgarian; Old Church Slavonic",
	"cv": "Chuvash",
	"kw": "Cornish",
	"co": "Corsican",
	"cr": "Cree",
	"cs": "Czech",
	"da": "Danish",
	"dv": "Divehi; Dhivehi; Maldivian",
	"nl": "Dutch; Flemish",
	"dz": "Dzongkha",
	"en": "English",
	"eo": "Esperanto",
	"et": "Estonian",
	"ee": "Ewe",
	"fo": "Faroese",
	"fj": "Fijian",
	"fi": "Finnish",
	"fr": "French",
	"fy": "Western Frisian",
	"ff": "Fulah",
	"ka": "Georgian",
	"de": "German",
	"gd": "Gaelic; Scottish Gaelic",
	"ga": "Irish",
	"gl": "Galician",
	"gv": "Manx",
	"el": "Greek, Modern (1453-)",
	"gn": "Guarani",
	"gu": "Gujarati",
	"ht": "Haitian; Haitian Creole",
	"ha": "Hausa",
	"he": "Hebrew",
	"hz": "Herero",
	"hi": "Hindi",
	"ho": "Hiri Motu",
	"hr": "Croatian",
	"hu": "Hungarian",
	"ig": "Igbo",
	"is": "Icelandic",
	"io": "Ido",
	"ii": "Sichuan Yi; Nuosu",
	"iu": "Inuktitut",
	"ie": "Interlingue; Occidental",
	"ia": "Interlingua (International Auxiliary Language Association)",
	"id": "Indonesian",
	"ik": "Inupiaq",
	"it": "Italian",
	"jv": "Javanese",
	"ja": "Japanese",
	"kl": "Kalaallisut; Greenlandic",
	"kn": "Kannada",
	"ks": "Kashmiri",
	"kr": "Kanuri",
	"kk": "Kazakh",
	"km": "Central Khmer",
	"ki": "Kikuyu; Gikuyu",
	"rw": "Kinyarwanda",
	"ky": "Kirghiz; Kyrgyz",
	"kv": "Komi",
	"kg": "Kongo",
	"ko": "Korean",
	"kj": "Kuanyama; Kwanyama",
	"ku": "Kurdish",
	"lo": "Lao",
	"la": "Latin",
	"lv": "Latvian",
	"li": "Limburgan; Limburger; Limburgish",
	"ln": "Lingala",
	"lt": "Lithuanian",
	"lb": "Luxembourgish; Letzeburgesch",
	"lu": "Luba-Katanga",
	"lg": "Ganda",
	"mk": "Macedonian",
	"mh": "Marshallese",
	"ml": "Malayalam",
	"mi": "Maori",
	"mr": "Marathi",
	"ms": "Malay",
	"mg": "Malagasy",
	"mt": "Maltese",
	"mn": "Mongolian",
	"na": "Nauru",
	"nv": "Navajo; Navaho",
	"nr": "Ndebele, South; South Ndebele",
	"nd": "Ndebele, North; North Ndebele",
	"ng": "Ndonga",
	"ne": "Nepali",
	"nn": "Norwegian Nynorsk; Nynorsk, Norwegian",
	"nb": "Bokmål, Norwegian; Norwegian Bokmål",
	"no": "Norwegian",
	"ny": "Chichewa; Chewa; Nyanja",
	"oc": "Occitan (post 1500); Provençal",
	"oj": "Ojibwa",
	"or": "Oriya",
	"om": "Oromo",
	"os": "Ossetian; Ossetic",
	"pa": "Panjabi; Punjabi",
	"fa": "Persian",
	"pi": "Pali",
	"pl": "Polish",
	"pt": "Portuguese",
	"ps": "Pushto; Pashto",
	"qu": "Quechua",
	"rm": "Romansh",
	"ro": "Romanian; Moldavian; Moldovan",
	"rn": "Rundi",
	"ru": "Russian",
	"sg": "Sango",
	"sa": "Sanskrit",
	"si": "Sinhala; Sinhalese",
	"sk": "Slovak",
	"sl": "Slovenian",
	"se": "Northern Sami",
	"sm": "Samoan",
	"sn": "Shona",
	"sd": "Sindhi",
	"so": "Somali",
	"st": "Sotho, Southern",
	"es": "Spanish; Castilian",
	"sc": "Sardinian",
	"sr": "Serbian",
	"ss": "Swati",
	"su": "Sundanese",
	"sw": "Swahili",
	"sv": "Swedish",
	"ty": "Tahitian",
	"ta": "Tamil",
	"tt": "Tatar",
	"te": "Telugu",
	"tg": "Tajik",
	"tl": "Tagalog",
	"th": "Thai",
	"bo": "Tibetan",
	"ti": "Tigrinya",
	"to": "Tonga (Tonga Islands)",
	"tn": "Tswana",
	"ts": "Tsonga",
	"tk": "Turkmen",
	"tr": "Turkish",
	"tw": "Twi",
	"ug": "Uighur; Uyghur",
	"uk": "Ukrainian",
	"ur": "Urdu",
	"uz": "Uzbek",
	"ve": "Venda",
	"vi": "Vietnamese",
	"vo": "Volapük",
	"cy": "Welsh",
	"wa": "Walloon",
	"wo": "Wolof",
	"xh": "Xhosa",
	"yi": "Yiddish",
	"yo": "Yoruba",
	"za": "Zhuang; Chuang",
	"zu": "Zulu",
}
