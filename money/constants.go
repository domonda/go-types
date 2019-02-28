package money

// CurrencyNull represents the SQL NULL for Currency and NullableCurrency.
// Currency(CurrencyNull).Valid() == false
// NullableCurrency(CurrencyNull).Valid() == true
const CurrencyNull = ""

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
	SLL Currency = "SLL" // Sierra Leone Leone
	SOS Currency = "SOS" // Somalia Shilling
	SPL Currency = "SPL" // Seborga Luigino
	SRD Currency = "SRD" // Suriname Dollar
	STD Currency = "STD" // São Tomé and Príncipe Dobra
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
	VEF Currency = "VEF" // Venezuela Bolivar
	VND Currency = "VND" // Viet Nam Dong
	VUV Currency = "VUV" // Vanuatu Vatu
	WST Currency = "WST" // Samoa Tala
	XAF Currency = "XAF" // Communauté Financière Africaine (BEAC) CFA Franc BEAC
	XCD Currency = "XCD" // East Caribbean Dollar
	XDR Currency = "XDR" // International Monetary Fund (IMF) Special Drawing Rights
	XOF Currency = "XOF" // Communauté Financière Africaine (BCEAO) Franc
	XPF Currency = "XPF" // Comptoirs Français du Pacifique (CFP) Franc
	YER Currency = "YER" // Yemen Rial
	ZAR Currency = "ZAR" // South Africa Rand
	ZMW Currency = "ZMW" // Zambia Kwacha
	ZWD Currency = "ZWD" // Zimbabwe Dollar

	BTC Currency = "BTC" // Bitcoin
)

var currencySymbolToCode = map[string]Currency{
	"€":    EUR,
	"$":    USD,
	"US$":  USD,
	"A$":   AUD,
	"Can$": CAD,
	"C$":   CAD,
	"HK$":  HKD,
	"NZ$":  NZD,
	"S$":   SGD,
	"£":    GBP,
	"GB£":  GBP,
	"₣":    CHF,
	"Ft":   HUF,
	"kn":   HRK,
	"¥":    JPY,
	"₿":    BTC,
}

var currencyCodeToSymbol = map[Currency]string{
	EUR: "€",
	USD: "$",
	GBP: "£",
	CHF: "₣",
	JPY: "¥",
}

var currencyCodeToName = map[Currency]string{
	AED: "United Arab Emirates Dirham",
	AFN: "Afghanistan Afghani",
	ALL: "Albania Lek",
	AMD: "Armenia Dram",
	ANG: "Netherlands Antilles Guilder",
	AOA: "Angola Kwanza",
	ARS: "Argentina Peso",
	AUD: "Australia Dollar",
	AWG: "Aruba Guilder",
	AZN: "Azerbaijan New Manat",
	BAM: "Bosnia and Herzegovina Convertible Marka",
	BBD: "Barbados Dollar",
	BDT: "Bangladesh Taka",
	BGN: "Bulgaria Lev",
	BHD: "Bahrain Dinar",
	BIF: "Burundi Franc",
	BMD: "Bermuda Dollar",
	BND: "Brunei Darussalam Dollar",
	BOB: "Bolivia Bolíviano",
	BRL: "Brazil Real",
	BSD: "Bahamas Dollar",
	BTN: "Bhutan Ngultrum",
	BWP: "Botswana Pula",
	BYN: "Belarus Ruble",
	BZD: "Belize Dollar",
	CAD: "Canada Dollar",
	CDF: "Congo/Kinshasa Franc",
	CHF: "Switzerland Franc",
	CLP: "Chile Peso",
	CNY: "China Yuan Renminbi",
	COP: "Colombia Peso",
	CRC: "Costa Rica Colon",
	CUC: "Cuba Convertible Peso",
	CUP: "Cuba Peso",
	CVE: "Cape Verde Escudo",
	CZK: "Czech Republic Koruna",
	DJF: "Djibouti Franc",
	DKK: "Denmark Krone",
	DOP: "Dominican Republic Peso",
	DZD: "Algeria Dinar",
	EGP: "Egypt Pound",
	ERN: "Eritrea Nakfa",
	ETB: "Ethiopia Birr",
	EUR: "Euro Member Countries",
	FJD: "Fiji Dollar",
	FKP: "Falkland Islands (Malvinas) Pound",
	GBP: "United Kingdom Pound",
	GEL: "Georgia Lari",
	GGP: "Guernsey Pound",
	GHS: "Ghana Cedi",
	GIP: "Gibraltar Pound",
	GMD: "Gambia Dalasi",
	GNF: "Guinea Franc",
	GTQ: "Guatemala Quetzal",
	GYD: "Guyana Dollar",
	HKD: "Hong Kong Dollar",
	HNL: "Honduras Lempira",
	HRK: "Croatia Kuna",
	HTG: "Haiti Gourde",
	HUF: "Hungary Forint",
	IDR: "Indonesia Rupiah",
	ILS: "Israel Shekel",
	IMP: "Isle of Man Pound",
	INR: "India Rupee",
	IQD: "Iraq Dinar",
	IRR: "Iran Rial",
	ISK: "Iceland Krona",
	JEP: "Jersey Pound",
	JMD: "Jamaica Dollar",
	JOD: "Jordan Dinar",
	JPY: "Japan Yen",
	KES: "Kenya Shilling",
	KGS: "Kyrgyzstan Som",
	KHR: "Cambodia Riel",
	KMF: "Comoros Franc",
	KPW: "Korea (North) Won",
	KRW: "Korea (South) Won",
	KWD: "Kuwait Dinar",
	KYD: "Cayman Islands Dollar",
	KZT: "Kazakhstan Tenge",
	LAK: "Laos Kip",
	LBP: "Lebanon Pound",
	LKR: "Sri Lanka Rupee",
	LRD: "Liberia Dollar",
	LSL: "Lesotho Loti",
	LYD: "Libya Dinar",
	MAD: "Morocco Dirham",
	MDL: "Moldova Leu",
	MGA: "Madagascar Ariary",
	MKD: "Macedonia Denar",
	MMK: "Myanmar (Burma) Kyat",
	MNT: "Mongolia Tughrik",
	MOP: "Macau Pataca",
	MRO: "Mauritania Ouguiya",
	MUR: "Mauritius Rupee",
	MVR: "Maldives (Maldive Islands) Rufiyaa",
	MWK: "Malawi Kwacha",
	MXN: "Mexico Peso",
	MYR: "Malaysia Ringgit",
	MZN: "Mozambique Metical",
	NAD: "Namibia Dollar",
	NGN: "Nigeria Naira",
	NIO: "Nicaragua Cordoba",
	NOK: "Norway Krone",
	NPR: "Nepal Rupee",
	NZD: "New Zealand Dollar",
	OMR: "Oman Rial",
	PAB: "Panama Balboa",
	PEN: "Peru Sol",
	PGK: "Papua New Guinea Kina",
	PHP: "Philippines Peso",
	PKR: "Pakistan Rupee",
	PLN: "Poland Zloty",
	PYG: "Paraguay Guarani",
	QAR: "Qatar Riyal",
	RON: "Romania New Leu",
	RSD: "Serbia Dinar",
	RUB: "Russia Ruble",
	RWF: "Rwanda Franc",
	SAR: "Saudi Arabia Riyal",
	SBD: "Solomon Islands Dollar",
	SCR: "Seychelles Rupee",
	SDG: "Sudan Pound",
	SEK: "Sweden Krona",
	SGD: "Singapore Dollar",
	SHP: "Saint Helena Pound",
	SLL: "Sierra Leone Leone",
	SOS: "Somalia Shilling",
	SPL: "Seborga Luigino",
	SRD: "Suriname Dollar",
	STD: "São Tomé and Príncipe Dobra",
	SVC: "El Salvador Colon",
	SYP: "Syria Pound",
	SZL: "Swaziland Lilangeni",
	THB: "Thailand Baht",
	TJS: "Tajikistan Somoni",
	TMT: "Turkmenistan Manat",
	TND: "Tunisia Dinar",
	TOP: "Tonga Pa'anga",
	TRY: "Turkey Lira",
	TTD: "Trinidad and Tobago Dollar",
	TVD: "Tuvalu Dollar",
	TWD: "Taiwan New Dollar",
	TZS: "Tanzania Shilling",
	UAH: "Ukraine Hryvnia",
	UGX: "Uganda Shilling",
	USD: "United States Dollar",
	UYU: "Uruguay Peso",
	UZS: "Uzbekistan Som",
	VEF: "Venezuela Bolivar",
	VND: "Viet Nam Dong",
	VUV: "Vanuatu Vatu",
	WST: "Samoa Tala",
	XAF: "Communauté Financière Africaine (BEAC) CFA Franc BEAC",
	XCD: "East Caribbean Dollar",
	XDR: "International Monetary Fund (IMF) Special Drawing Rights",
	XOF: "Communauté Financière Africaine (BCEAO) Franc",
	XPF: "Comptoirs Français du Pacifique (CFP) Franc",
	YER: "Yemen Rial",
	ZAR: "South Africa Rand",
	ZMW: "Zambia Kwacha",
	ZWD: "Zimbabwe Dollar",
}
