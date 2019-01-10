package country

import (
	"strings"

	"github.com/domonda/errors"
)

// Code according ISO 3166-1 alpha 2
type Code string

func (cc Code) Valid() bool {
	_, ok := countryMap[cc]
	return ok
}

func (cc Code) CountryName() string {
	return countryMap[cc]
}

// AssignString tries to parse and assign the passed
// source string as value of the implementing object.
// It returns an error if source could not be parsed.
// If the source string could be parsed, but was not
// in the expected normalized format, then false is
// returned for normalized and nil for err.
// AssignString implements strfmt.StringAssignable
func (cc *Code) AssignString(source string) (normalized bool, err error) {
	newCC := Code(strings.ToUpper(source))
	if !newCC.Valid() {
		return false, errors.Errorf("Invalid country code: %#v", source)
	}
	*cc = newCC
	return newCC == Code(source), nil
}

const (
	AF Code = "AF"
	AX Code = "AX"
	AL Code = "AL"
	DZ Code = "DZ"
	AS Code = "AS"
	AD Code = "AD"
	AO Code = "AO"
	AI Code = "AI"
	AQ Code = "AQ"
	AG Code = "AG"
	AR Code = "AR"
	AM Code = "AM"
	AW Code = "AW"
	AU Code = "AU"
	AT Code = "AT"
	AZ Code = "AZ"
	BS Code = "BS"
	BH Code = "BH"
	BD Code = "BD"
	BB Code = "BB"
	BY Code = "BY"
	BE Code = "BE"
	BZ Code = "BZ"
	BJ Code = "BJ"
	BM Code = "BM"
	BT Code = "BT"
	BO Code = "BO"
	BQ Code = "BQ"
	BA Code = "BA"
	BW Code = "BW"
	BV Code = "BV"
	BR Code = "BR"
	IO Code = "IO"
	BN Code = "BN"
	BG Code = "BG"
	BF Code = "BF"
	BI Code = "BI"
	KH Code = "KH"
	CM Code = "CM"
	CA Code = "CA"
	CV Code = "CV"
	KY Code = "KY"
	CF Code = "CF"
	TD Code = "TD"
	CL Code = "CL"
	CN Code = "CN"
	CX Code = "CX"
	CC Code = "CC"
	CO Code = "CO"
	KM Code = "KM"
	CG Code = "CG"
	CD Code = "CD"
	CK Code = "CK"
	CR Code = "CR"
	CI Code = "CI"
	HR Code = "HR"
	CU Code = "CU"
	CW Code = "CW"
	CY Code = "CY"
	CZ Code = "CZ"
	DK Code = "DK"
	DJ Code = "DJ"
	DM Code = "DM"
	DO Code = "DO"
	EC Code = "EC"
	EG Code = "EG"
	SV Code = "SV"
	GQ Code = "GQ"
	ER Code = "ER"
	EE Code = "EE"
	ET Code = "ET"
	FK Code = "FK"
	FO Code = "FO"
	FJ Code = "FJ"
	FI Code = "FI"
	FR Code = "FR"
	GF Code = "GF"
	PF Code = "PF"
	TF Code = "TF"
	GA Code = "GA"
	GM Code = "GM"
	GE Code = "GE"
	DE Code = "DE"
	GH Code = "GH"
	GI Code = "GI"
	GR Code = "GR"
	GL Code = "GL"
	GD Code = "GD"
	GP Code = "GP"
	GU Code = "GU"
	GT Code = "GT"
	GG Code = "GG"
	GN Code = "GN"
	GW Code = "GW"
	GY Code = "GY"
	HT Code = "HT"
	HM Code = "HM"
	VA Code = "VA"
	HN Code = "HN"
	HK Code = "HK"
	HU Code = "HU"
	IS Code = "IS"
	IN Code = "IN"
	ID Code = "ID"
	IR Code = "IR"
	IQ Code = "IQ"
	IE Code = "IE"
	IM Code = "IM"
	IL Code = "IL"
	IT Code = "IT"
	JM Code = "JM"
	JP Code = "JP"
	JE Code = "JE"
	JO Code = "JO"
	KZ Code = "KZ"
	KE Code = "KE"
	KI Code = "KI"
	KP Code = "KP"
	KR Code = "KR"
	KW Code = "KW"
	KG Code = "KG"
	LA Code = "LA"
	LV Code = "LV"
	LB Code = "LB"
	LS Code = "LS"
	LR Code = "LR"
	LY Code = "LY"
	LI Code = "LI"
	LT Code = "LT"
	LU Code = "LU"
	MO Code = "MO"
	MK Code = "MK"
	MG Code = "MG"
	MW Code = "MW"
	MY Code = "MY"
	MV Code = "MV"
	ML Code = "ML"
	MT Code = "MT"
	MH Code = "MH"
	MQ Code = "MQ"
	MR Code = "MR"
	MU Code = "MU"
	YT Code = "YT"
	MX Code = "MX"
	FM Code = "FM"
	MD Code = "MD"
	MC Code = "MC"
	MN Code = "MN"
	ME Code = "ME"
	MS Code = "MS"
	MA Code = "MA"
	MZ Code = "MZ"
	MM Code = "MM"
	NA Code = "NA"
	NR Code = "NR"
	NP Code = "NP"
	NL Code = "NL"
	NC Code = "NC"
	NZ Code = "NZ"
	NI Code = "NI"
	NE Code = "NE"
	NG Code = "NG"
	NU Code = "NU"
	NF Code = "NF"
	MP Code = "MP"
	NO Code = "NO"
	OM Code = "OM"
	PK Code = "PK"
	PW Code = "PW"
	PS Code = "PS"
	PA Code = "PA"
	PG Code = "PG"
	PY Code = "PY"
	PE Code = "PE"
	PH Code = "PH"
	PN Code = "PN"
	PL Code = "PL"
	PT Code = "PT"
	PR Code = "PR"
	QA Code = "QA"
	RE Code = "RE"
	RO Code = "RO"
	RU Code = "RU"
	RW Code = "RW"
	BL Code = "BL"
	SH Code = "SH"
	KN Code = "KN"
	LC Code = "LC"
	MF Code = "MF"
	PM Code = "PM"
	VC Code = "VC"
	WS Code = "WS"
	SM Code = "SM"
	ST Code = "ST"
	SA Code = "SA"
	SN Code = "SN"
	RS Code = "RS"
	SC Code = "SC"
	SL Code = "SL"
	SG Code = "SG"
	SX Code = "SX"
	SK Code = "SK"
	SI Code = "SI"
	SB Code = "SB"
	SO Code = "SO"
	ZA Code = "ZA"
	GS Code = "GS"
	SS Code = "SS"
	ES Code = "ES"
	LK Code = "LK"
	SD Code = "SD"
	SR Code = "SR"
	SJ Code = "SJ"
	SZ Code = "SZ"
	SE Code = "SE"
	CH Code = "CH"
	SY Code = "SY"
	TW Code = "TW"
	TJ Code = "TJ"
	TZ Code = "TZ"
	TH Code = "TH"
	TL Code = "TL"
	TG Code = "TG"
	TK Code = "TK"
	TO Code = "TO"
	TT Code = "TT"
	TN Code = "TN"
	TR Code = "TR"
	TM Code = "TM"
	TC Code = "TC"
	TV Code = "TV"
	UG Code = "UG"
	UA Code = "UA"
	AE Code = "AE"
	GB Code = "GB"
	US Code = "US"
	UM Code = "UM"
	UY Code = "UY"
	UZ Code = "UZ"
	VU Code = "VU"
	VE Code = "VE"
	VN Code = "VN"
	VG Code = "VG"
	VI Code = "VI"
	WF Code = "WF"
	EH Code = "EH"
	YE Code = "YE"
	ZM Code = "ZM"
	ZW Code = "ZW"
)

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
}
