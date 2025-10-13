// Package country provides comprehensive country code handling and validation
// based on ISO 3166-1 alpha-2 standards for Go applications.
//
// The package includes:
// - ISO 3166-1 alpha-2 country code validation and normalization
// - Alternative country code mappings (ITU codes, German names, etc.)
// - European Union membership checking
// - Database integration (Scanner/Valuer interfaces)
// - JSON marshalling/unmarshalling
// - Nullable country code support
package country

// Country code constants for ISO 3166-1 alpha-2 standard.
// These represent all officially recognized country codes.
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

// countryMap maps country codes to their English names.
var countryMap = map[Code]string{
	AF: "Afghanistan",
	AX: "Åland Islands",
	AL: "Albania",
	DZ: "Algeria",
	AS: "American Samoa",
	AD: "Andorra",
	AO: "Angola",
	AI: "Anguilla",
	AQ: "Antarctica",
	AG: "Antigua and Barbuda",
	AR: "Argentina",
	AM: "Armenia",
	AW: "Aruba",
	AU: "Australia",
	AT: "Austria",
	AZ: "Azerbaijan",
	BS: "Bahamas",
	BH: "Bahrain",
	BD: "Bangladesh",
	BB: "Barbados",
	BY: "Belarus",
	BE: "Belgium",
	BZ: "Belize",
	BJ: "Benin",
	BM: "Bermuda",
	BT: "Bhutan",
	BO: "Bolivia, Plurinational State of",
	BQ: "Bonaire, Sint Eustatius and Saba",
	BA: "Bosnia and Herzegovina",
	BW: "Botswana",
	BV: "Bouvet Island",
	BR: "Brazil",
	IO: "British Indian Ocean Territory",
	BN: "Brunei Darussalam",
	BG: "Bulgaria",
	BF: "Burkina Faso",
	BI: "Burundi",
	KH: "Cambodia",
	CM: "Cameroon",
	CA: "Canada",
	CV: "Cape Verde",
	KY: "Cayman Islands",
	CF: "Central African Republic",
	TD: "Chad",
	CL: "Chile",
	CN: "China",
	CX: "Christmas Island",
	CC: "Cocos (Keeling) Islands",
	CO: "Colombia",
	KM: "Comoros",
	CG: "Congo",
	CD: "Congo, the Democratic Republic of the",
	CK: "Cook Islands",
	CR: "Costa Rica",
	CI: "Côte d'Ivoire",
	HR: "Croatia",
	CU: "Cuba",
	CW: "Curaçao",
	CY: "Cyprus",
	CZ: "Czech Republic",
	DK: "Denmark",
	DJ: "Djibouti",
	DM: "Dominica",
	DO: "Dominican Republic",
	EC: "Ecuador",
	EG: "Egypt",
	SV: "El Salvador",
	GQ: "Equatorial Guinea",
	ER: "Eritrea",
	EE: "Estonia",
	ET: "Ethiopia",
	FK: "Falkland Islands (Malvinas)",
	FO: "Faroe Islands",
	FJ: "Fiji",
	FI: "Finland",
	FR: "France",
	GF: "French Guiana",
	PF: "French Polynesia",
	TF: "French Southern Territories",
	GA: "Gabon",
	GM: "Gambia",
	GE: "Georgia",
	DE: "Germany",
	GH: "Ghana",
	GI: "Gibraltar",
	GR: "Greece",
	EL: "Greece",
	GL: "Greenland",
	GD: "Grenada",
	GP: "Guadeloupe",
	GU: "Guam",
	GT: "Guatemala",
	GG: "Guernsey",
	GN: "Guinea",
	GW: "Guinea-Bissau",
	GY: "Guyana",
	HT: "Haiti",
	HM: "Heard Island and McDonald Islands",
	VA: "Holy See (Vatican City State)",
	HN: "Honduras",
	HK: "Hong Kong",
	HU: "Hungary",
	IS: "Iceland",
	IN: "India",
	ID: "Indonesia",
	IR: "Iran, Islamic Republic of",
	IQ: "Iraq",
	IE: "Ireland",
	IM: "Isle of Man",
	IL: "Israel",
	IT: "Italy",
	JM: "Jamaica",
	JP: "Japan",
	JE: "Jersey",
	JO: "Jordan",
	KZ: "Kazakhstan",
	KE: "Kenya",
	KI: "Kiribati",
	KP: "Korea, Democratic People's Republic of",
	KR: "Korea, Republic of",
	KW: "Kuwait",
	KG: "Kyrgyzstan",
	LA: "Lao People's Democratic Republic",
	LV: "Latvia",
	LB: "Lebanon",
	LS: "Lesotho",
	LR: "Liberia",
	LY: "Libya",
	LI: "Liechtenstein",
	LT: "Lithuania",
	LU: "Luxembourg",
	MO: "Macao",
	MK: "Macedonia, the Former Yugoslav Republic of",
	MG: "Madagascar",
	MW: "Malawi",
	MY: "Malaysia",
	MV: "Maldives",
	ML: "Mali",
	MT: "Malta",
	MH: "Marshall Islands",
	MQ: "Martinique",
	MR: "Mauritania",
	MU: "Mauritius",
	YT: "Mayotte",
	MX: "Mexico",
	FM: "Micronesia, Federated States of",
	MD: "Moldova, Republic of",
	MC: "Monaco",
	MN: "Mongolia",
	ME: "Montenegro",
	MS: "Montserrat",
	MA: "Morocco",
	MZ: "Mozambique",
	MM: "Myanmar",
	NA: "Namibia",
	NR: "Nauru",
	NP: "Nepal",
	NL: "Netherlands",
	NC: "New Caledonia",
	NZ: "New Zealand",
	NI: "Nicaragua",
	NE: "Niger",
	NG: "Nigeria",
	NU: "Niue",
	NF: "Norfolk Island",
	MP: "Northern Mariana Islands",
	NO: "Norway",
	OM: "Oman",
	PK: "Pakistan",
	PW: "Palau",
	PS: "Palestine, State of",
	PA: "Panama",
	PG: "Papua New Guinea",
	PY: "Paraguay",
	PE: "Peru",
	PH: "Philippines",
	PN: "Pitcairn",
	PL: "Poland",
	PT: "Portugal",
	PR: "Puerto Rico",
	QA: "Qatar",
	RE: "Réunion",
	RO: "Romania",
	RU: "Russian Federation",
	RW: "Rwanda",
	BL: "Saint Barthélemy",
	SH: "Saint Helena, Ascension and Tristan da Cunha",
	KN: "Saint Kitts and Nevis",
	LC: "Saint Lucia",
	MF: "Saint Martin (French part)",
	PM: "Saint Pierre and Miquelon",
	VC: "Saint Vincent and the Grenadines",
	WS: "Samoa",
	SM: "San Marino",
	ST: "Sao Tome and Principe",
	SA: "Saudi Arabia",
	SN: "Senegal",
	RS: "Serbia",
	SC: "Seychelles",
	SL: "Sierra Leone",
	SG: "Singapore",
	SX: "Sint Maarten (Dutch part)",
	SK: "Slovakia",
	SI: "Slovenia",
	SB: "Solomon Islands",
	SO: "Somalia",
	ZA: "South Africa",
	GS: "South Georgia and the South Sandwich Islands",
	SS: "South Sudan",
	ES: "Spain",
	LK: "Sri Lanka",
	SD: "Sudan",
	SR: "Suriname",
	SJ: "Svalbard and Jan Mayen",
	SZ: "Swaziland",
	SE: "Sweden",
	CH: "Switzerland",
	SY: "Syrian Arab Republic",
	TW: "Taiwan, Province of China",
	TJ: "Tajikistan",
	TZ: "Tanzania, United Republic of",
	TH: "Thailand",
	TL: "Timor-Leste",
	TG: "Togo",
	TK: "Tokelau",
	TO: "Tonga",
	TT: "Trinidad and Tobago",
	TN: "Tunisia",
	TR: "Turkey",
	TM: "Turkmenistan",
	TC: "Turks and Caicos Islands",
	TV: "Tuvalu",
	UG: "Uganda",
	UA: "Ukraine",
	AE: "United Arab Emirates",
	GB: "United Kingdom",
	US: "United States",
	UM: "United States Minor Outlying Islands",
	UY: "Uruguay",
	UZ: "Uzbekistan",
	VU: "Vanuatu",
	VE: "Venezuela, Bolivarian Republic of",
	VN: "Viet Nam",
	VG: "Virgin Islands, British",
	VI: "Virgin Islands, U.S.",
	WF: "Wallis and Futuna",
	EH: "Western Sahara",
	YE: "Yemen",
	ZM: "Zambia",
	ZW: "Zimbabwe",
	XK: "Republic of Kosovo", // unofficial, but still can be used
}

// AltCodes provides alternative country code mappings.
// Includes common codes from other standards, ITU letter codes, and German country names.
var AltCodes = map[string]Code{
	// Common codes from other standards
	"A":   AT,
	"DEU": DE,

	// ITU letter codes
	"AUT": AT,
	"B":   BR,
	"D":   DE,
	"E":   ES,
	"F":   FR,
	"G":   GB,
	"I":   IT,
	"J":   JP,
	"S":   SE,
	"SUI": CH,

	// German country names
	"ALBANIEN":               AL,
	"ANDORRA":                AD,
	"ARMENIEN":               AM,
	"ASERBAIDSCHAN":          AZ,
	"BELGIEN":                BE,
	"BOSNIEN-HERZEGOWINA":    BA,
	"BULGARIEN":              BG,
	"DÄNEMARK":               DK,
	"DEUTSCHLAND":            DE,
	"ESTLAND":                EE,
	"FINNLAND":               FI,
	"FRANKREICH":             FR,
	"GEORGIEN":               GE,
	"GRIECHENLAND":           GR,
	"IRLAND":                 IE,
	"ISLAND":                 IS,
	"ITALIEN":                IT,
	"KASACHSTAN":             KZ,
	"KOSOVO":                 XK,
	"KROATIEN":               HR,
	"LETTLAND":               LV,
	"LIECHTENSTEIN":          LI,
	"LITAUEN":                LT,
	"LUXEMBURG":              LU,
	"MALTA":                  MT,
	"MOLDAWIEN":              MD,
	"MONACO":                 MC,
	"MONTENEGRO":             ME,
	"NIEDERLANDE":            NL,
	"NORDMAZEDONIEN":         MK,
	"NORWEGEN":               NO,
	"ÖSTERREICH":             AT,
	"OESTERREICH":            AT,
	"POLEN":                  PL,
	"PORTUGAL":               PT,
	"RUMÄNIEN":               RO,
	"RUSSLAND":               RU,
	"SAN MARINO":             SM,
	"SCHWEDEN":               SE,
	"SCHWEIZ":                CH,
	"SERBIEN":                RS,
	"SLOWAKEI":               SK,
	"SLOWENIEN":              SI,
	"SPANIEN":                ES,
	"TSCHECHISCHE REPUBLIK":  CZ,
	"TÜRKEI":                 TR,
	"UKRAINE":                UA,
	"UNGARN":                 HU,
	"VATIKANSTADT":           VA,
	"VEREINIGTES KÖNIGREICH": GB,
	"WEISSRUSSLAND":          BY,
	"ZYPERN":                 CY,
}

// euCountries contains the set of European Union member countries.
var euCountries = map[Code]struct{}{
	AT: {},
	BE: {},
	BG: {},
	HR: {},
	CY: {},
	CZ: {},
	DK: {},
	EE: {},
	FI: {},
	FR: {},
	DE: {},
	GR: {},
	HU: {},
	IE: {},
	IT: {},
	LV: {},
	LT: {},
	LU: {},
	MT: {},
	NL: {},
	PL: {},
	PT: {},
	RO: {},
	SK: {},
	SI: {},
	ES: {},
	SE: {},
}
