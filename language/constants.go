package language

// https://iso639-3.sil.org/code_tables/download_tables

const (
	// Null is an empty string and will be treatet as SQL NULL.
	// language.Null.Valid() == false
	Null Code = ""

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
