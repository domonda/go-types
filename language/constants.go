package language

// https://iso639-3.sil.org/code_tables/download_tables

const (
	// Null is an empty string and will be treatet as SQL NULL.
	// language.Null.Valid() == false
	Null Code = ""

	AA Code = "aa"
	AB Code = "ab"
	AF Code = "af"
	AK Code = "ak"
	SQ Code = "sq"
	AM Code = "am"
	AR Code = "ar"
	AN Code = "an"
	HY Code = "hy"
	AS Code = "as"
	AV Code = "av"
	AE Code = "ae"
	AY Code = "ay"
	AZ Code = "az"
	BA Code = "ba"
	BM Code = "bm"
	EU Code = "eu"
	BE Code = "be"
	BN Code = "bn"
	BH Code = "bh"
	BI Code = "bi"
	BS Code = "bs"
	BR Code = "br"
	BG Code = "bg"
	MY Code = "my"
	CA Code = "ca"
	CH Code = "ch"
	CE Code = "ce"
	ZH Code = "zh"
	CU Code = "cu"
	CV Code = "cv"
	KW Code = "kw"
	CO Code = "co"
	CR Code = "cr"
	CS Code = "cs"
	DA Code = "da"
	DV Code = "dv"
	NL Code = "nl"
	DZ Code = "dz"
	EN Code = "en"
	EO Code = "eo"
	ET Code = "et"
	EE Code = "ee"
	FO Code = "fo"
	FJ Code = "fj"
	FI Code = "fi"
	FR Code = "fr"
	FY Code = "fy"
	FF Code = "ff"
	KA Code = "ka"
	DE Code = "de"
	GD Code = "gd"
	GA Code = "ga"
	GL Code = "gl"
	GV Code = "gv"
	EL Code = "el"
	GN Code = "gn"
	GU Code = "gu"
	HT Code = "ht"
	HA Code = "ha"
	HE Code = "he"
	HZ Code = "hz"
	HI Code = "hi"
	HO Code = "ho"
	HR Code = "hr"
	HU Code = "hu"
	IG Code = "ig"
	IS Code = "is"
	IO Code = "io"
	II Code = "ii"
	IU Code = "iu"
	IE Code = "ie"
	IA Code = "ia"
	ID Code = "id"
	IK Code = "ik"
	IT Code = "it"
	JV Code = "jv"
	JA Code = "ja"
	KL Code = "kl"
	KN Code = "kn"
	KS Code = "ks"
	KR Code = "kr"
	KK Code = "kk"
	KM Code = "km"
	KI Code = "ki"
	RW Code = "rw"
	KY Code = "ky"
	KV Code = "kv"
	KG Code = "kg"
	KO Code = "ko"
	KJ Code = "kj"
	KU Code = "ku"
	LO Code = "lo"
	LA Code = "la"
	LV Code = "lv"
	LI Code = "li"
	LN Code = "ln"
	LT Code = "lt"
	LB Code = "lb"
	LU Code = "lu"
	LG Code = "lg"
	MK Code = "mk"
	MH Code = "mh"
	ML Code = "ml"
	MI Code = "mi"
	MR Code = "mr"
	MS Code = "ms"
	MG Code = "mg"
	MT Code = "mt"
	MN Code = "mn"
	NA Code = "na"
	NV Code = "nv"
	NR Code = "nr"
	ND Code = "nd"
	NG Code = "ng"
	NE Code = "ne"
	NN Code = "nn"
	NB Code = "nb"
	NO Code = "no"
	NY Code = "ny"
	OC Code = "oc"
	OJ Code = "oj"
	OR Code = "or"
	OM Code = "om"
	OS Code = "os"
	PA Code = "pa"
	FA Code = "fa"
	PI Code = "pi"
	PL Code = "pl"
	PT Code = "pt"
	PS Code = "ps"
	QU Code = "qu"
	RM Code = "rm"
	RO Code = "ro"
	RN Code = "rn"
	RU Code = "ru"
	SG Code = "sg"
	SA Code = "sa"
	SI Code = "si"
	SK Code = "sk"
	SL Code = "sl"
	SE Code = "se"
	SM Code = "sm"
	SN Code = "sn"
	SD Code = "sd"
	SO Code = "so"
	ST Code = "st"
	ES Code = "es"
	SC Code = "sc"
	SR Code = "sr"
	SS Code = "ss"
	SU Code = "su"
	SW Code = "sw"
	SV Code = "sv"
	TY Code = "ty"
	TA Code = "ta"
	TT Code = "tt"
	TE Code = "te"
	TG Code = "tg"
	TL Code = "tl"
	TH Code = "th"
	BO Code = "bo"
	TI Code = "ti"
	TO Code = "to"
	TN Code = "tn"
	TS Code = "ts"
	TK Code = "tk"
	TR Code = "tr"
	TW Code = "tw"
	UG Code = "ug"
	UK Code = "uk"
	UR Code = "ur"
	UZ Code = "uz"
	VE Code = "ve"
	VI Code = "vi"
	VO Code = "vo"
	CY Code = "cy"
	WA Code = "wa"
	WO Code = "wo"
	XH Code = "xh"
	YI Code = "yi"
	YO Code = "yo"
	ZA Code = "za"
	ZU Code = "zu"
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
