package money

// CurrencyNull represents the SQL NULL for Currency and NullableCurrency.
// Currency(CurrencyNull).Valid() == false
// NullableCurrency(CurrencyNull).Valid() == true
const CurrencyNull = ""

const (
	AED = "AED" // United Arab Emirates Dirham
	AFN = "AFN" // Afghanistan Afghani
	ALL = "ALL" // Albania Lek
	AMD = "AMD" // Armenia Dram
	ANG = "ANG" // Netherlands Antilles Guilder
	AOA = "AOA" // Angola Kwanza
	ARS = "ARS" // Argentina Peso
	AUD = "AUD" // Australia Dollar
	AWG = "AWG" // Aruba Guilder
	AZN = "AZN" // Azerbaijan New Manat
	BAM = "BAM" // Bosnia and Herzegovina Convertible Marka
	BBD = "BBD" // Barbados Dollar
	BDT = "BDT" // Bangladesh Taka
	BGN = "BGN" // Bulgaria Lev
	BHD = "BHD" // Bahrain Dinar
	BIF = "BIF" // Burundi Franc
	BMD = "BMD" // Bermuda Dollar
	BND = "BND" // Brunei Darussalam Dollar
	BOB = "BOB" // Bolivia Bolíviano
	BRL = "BRL" // Brazil Real
	BSD = "BSD" // Bahamas Dollar
	BTN = "BTN" // Bhutan Ngultrum
	BWP = "BWP" // Botswana Pula
	BYN = "BYN" // Belarus Ruble
	BZD = "BZD" // Belize Dollar
	CAD = "CAD" // Canada Dollar
	CDF = "CDF" // Congo/Kinshasa Franc
	CHF = "CHF" // Switzerland Franc
	CLP = "CLP" // Chile Peso
	CNY = "CNY" // China Yuan Renminbi
	COP = "COP" // Colombia Peso
	CRC = "CRC" // Costa Rica Colon
	CUC = "CUC" // Cuba Convertible Peso
	CUP = "CUP" // Cuba Peso
	CVE = "CVE" // Cape Verde Escudo
	CZK = "CZK" // Czech Republic Koruna
	DJF = "DJF" // Djibouti Franc
	DKK = "DKK" // Denmark Krone
	DOP = "DOP" // Dominican Republic Peso
	DZD = "DZD" // Algeria Dinar
	EGP = "EGP" // Egypt Pound
	ERN = "ERN" // Eritrea Nakfa
	ETB = "ETB" // Ethiopia Birr
	EUR = "EUR" // Euro Member Countries
	FJD = "FJD" // Fiji Dollar
	FKP = "FKP" // Falkland Islands (Malvinas) Pound
	GBP = "GBP" // United Kingdom Pound
	GEL = "GEL" // Georgia Lari
	GGP = "GGP" // Guernsey Pound
	GHS = "GHS" // Ghana Cedi
	GIP = "GIP" // Gibraltar Pound
	GMD = "GMD" // Gambia Dalasi
	GNF = "GNF" // Guinea Franc
	GTQ = "GTQ" // Guatemala Quetzal
	GYD = "GYD" // Guyana Dollar
	HKD = "HKD" // Hong Kong Dollar
	HNL = "HNL" // Honduras Lempira
	HRK = "HRK" // Croatia Kuna
	HTG = "HTG" // Haiti Gourde
	HUF = "HUF" // Hungary Forint
	IDR = "IDR" // Indonesia Rupiah
	ILS = "ILS" // Israel Shekel
	IMP = "IMP" // Isle of Man Pound
	INR = "INR" // India Rupee
	IQD = "IQD" // Iraq Dinar
	IRR = "IRR" // Iran Rial
	ISK = "ISK" // Iceland Krona
	JEP = "JEP" // Jersey Pound
	JMD = "JMD" // Jamaica Dollar
	JOD = "JOD" // Jordan Dinar
	JPY = "JPY" // Japan Yen
	KES = "KES" // Kenya Shilling
	KGS = "KGS" // Kyrgyzstan Som
	KHR = "KHR" // Cambodia Riel
	KMF = "KMF" // Comoros Franc
	KPW = "KPW" // Korea (North) Won
	KRW = "KRW" // Korea (South) Won
	KWD = "KWD" // Kuwait Dinar
	KYD = "KYD" // Cayman Islands Dollar
	KZT = "KZT" // Kazakhstan Tenge
	LAK = "LAK" // Laos Kip
	LBP = "LBP" // Lebanon Pound
	LKR = "LKR" // Sri Lanka Rupee
	LRD = "LRD" // Liberia Dollar
	LSL = "LSL" // Lesotho Loti
	LYD = "LYD" // Libya Dinar
	MAD = "MAD" // Morocco Dirham
	MDL = "MDL" // Moldova Leu
	MGA = "MGA" // Madagascar Ariary
	MKD = "MKD" // Macedonia Denar
	MMK = "MMK" // Myanmar (Burma) Kyat
	MNT = "MNT" // Mongolia Tughrik
	MOP = "MOP" // Macau Pataca
	MRO = "MRO" // Mauritania Ouguiya
	MUR = "MUR" // Mauritius Rupee
	MVR = "MVR" // Maldives (Maldive Islands) Rufiyaa
	MWK = "MWK" // Malawi Kwacha
	MXN = "MXN" // Mexico Peso
	MYR = "MYR" // Malaysia Ringgit
	MZN = "MZN" // Mozambique Metical
	NAD = "NAD" // Namibia Dollar
	NGN = "NGN" // Nigeria Naira
	NIO = "NIO" // Nicaragua Cordoba
	NOK = "NOK" // Norway Krone
	NPR = "NPR" // Nepal Rupee
	NZD = "NZD" // New Zealand Dollar
	OMR = "OMR" // Oman Rial
	PAB = "PAB" // Panama Balboa
	PEN = "PEN" // Peru Sol
	PGK = "PGK" // Papua New Guinea Kina
	PHP = "PHP" // Philippines Peso
	PKR = "PKR" // Pakistan Rupee
	PLN = "PLN" // Poland Zloty
	PYG = "PYG" // Paraguay Guarani
	QAR = "QAR" // Qatar Riyal
	RON = "RON" // Romania New Leu
	RSD = "RSD" // Serbia Dinar
	RUB = "RUB" // Russia Ruble
	RWF = "RWF" // Rwanda Franc
	SAR = "SAR" // Saudi Arabia Riyal
	SBD = "SBD" // Solomon Islands Dollar
	SCR = "SCR" // Seychelles Rupee
	SDG = "SDG" // Sudan Pound
	SEK = "SEK" // Sweden Krona
	SGD = "SGD" // Singapore Dollar
	SHP = "SHP" // Saint Helena Pound
	SLL = "SLL" // Sierra Leone Leone
	SOS = "SOS" // Somalia Shilling
	SPL = "SPL" // Seborga Luigino
	SRD = "SRD" // Suriname Dollar
	STD = "STD" // São Tomé and Príncipe Dobra
	SVC = "SVC" // El Salvador Colon
	SYP = "SYP" // Syria Pound
	SZL = "SZL" // Swaziland Lilangeni
	THB = "THB" // Thailand Baht
	TJS = "TJS" // Tajikistan Somoni
	TMT = "TMT" // Turkmenistan Manat
	TND = "TND" // Tunisia Dinar
	TOP = "TOP" // Tonga Pa'anga
	TRY = "TRY" // Turkey Lira
	TTD = "TTD" // Trinidad and Tobago Dollar
	TVD = "TVD" // Tuvalu Dollar
	TWD = "TWD" // Taiwan New Dollar
	TZS = "TZS" // Tanzania Shilling
	UAH = "UAH" // Ukraine Hryvnia
	UGX = "UGX" // Uganda Shilling
	USD = "USD" // United States Dollar
	UYU = "UYU" // Uruguay Peso
	UZS = "UZS" // Uzbekistan Som
	VEF = "VEF" // Venezuela Bolivar
	VND = "VND" // Viet Nam Dong
	VUV = "VUV" // Vanuatu Vatu
	WST = "WST" // Samoa Tala
	XAF = "XAF" // Communauté Financière Africaine (BEAC) CFA Franc BEAC
	XCD = "XCD" // East Caribbean Dollar
	XDR = "XDR" // International Monetary Fund (IMF) Special Drawing Rights
	XOF = "XOF" // Communauté Financière Africaine (BCEAO) Franc
	XPF = "XPF" // Comptoirs Français du Pacifique (CFP) Franc
	YER = "YER" // Yemen Rial
	ZAR = "ZAR" // South Africa Rand
	ZMW = "ZMW" // Zambia Kwacha
	ZWD = "ZWD" // Zimbabwe Dollar

	BTC = "BTC" // Bitcoin
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
