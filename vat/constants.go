package vat

import (
	"regexp"

	"github.com/domonda/go-types/country"
)

const (
	// IDMinLength is the minium length of a VAT ID
	IDMinLength = 4

	// IDMaxLength is the maximum length of a VAT ID
	IDMaxLength = 14 + 2 // allow 2 spaces
)

// https://de.wikipedia.org/wiki/Umsatzsteuer-Identifikationsnummer
// http://www.pruefziffernberechnung.de/U/USt-IdNr.shtml
var vatidRegex = map[country.Code]*regexp.Regexp{
	"AT": regexp.MustCompile(`^AT\s??U\s??\d{8}$`),
	"BE": regexp.MustCompile(`^BE\s??\d{10}$`),
	"BG": regexp.MustCompile(`^BG\s??\d{9,10}$`),
	"CY": regexp.MustCompile(`^CY\s??\d{8}[A-Z]$`),
	"CZ": regexp.MustCompile(`^CZ\s??\d{8,10}$`),
	"DE": regexp.MustCompile(`^DE\s??[1-9]\d{8}$`),
	"DK": regexp.MustCompile(`^DK\s??\d{8}$`),
	"EE": regexp.MustCompile(`^EE\s??\d{9}$`),
	"EL": regexp.MustCompile(`^EL\s??\d{9}$`), // greece GR
	"ES": regexp.MustCompile(`^ES\s??X\s??\d{7}X$`),
	"FI": regexp.MustCompile(`^FI\s??\d{8}$`),
	"FR": regexp.MustCompile(`^FR\s??[0-9A-Z][0-9A-Z]\s??\d{9}$`),
	"GB": regexp.MustCompile(`^GB\s??(?:\d{9})|(?:\d{12})|(?:GD\d{3})|(?:HA\d{3})$`),
	"HR": regexp.MustCompile(`^HR\s??\d{11}$`),
	"HU": regexp.MustCompile(`^HU\s??\d{8,9}$`),
	"IE": regexp.MustCompile(`^IE\s??(?:\d[0-9A-Z]\d{5}[A-Z])|(?:\d{7}[A-W][A-I])$`),
	"IT": regexp.MustCompile(`^IT\s??\d{11}$`),
	"LT": regexp.MustCompile(`^LT\s??\d{9}\d{3}?$`),
	"LU": regexp.MustCompile(`^LU\s??\d{8}$`),
	"LV": regexp.MustCompile(`^LV\s??\d{11}$`),
	"MT": regexp.MustCompile(`^MT\s??\d{8}$`),
	"NL": regexp.MustCompile(`^NL\s??\d{9}B\d{2}$`),
	"NO": regexp.MustCompile(`^NO\s??\d{9}$`),
	"PL": regexp.MustCompile(`^PL\s??\d{10}$`),
	"PT": regexp.MustCompile(`^PT\s??\d{9}$`),
	"RO": regexp.MustCompile(`^RO\s??\d{2,10}$`),
	"SE": regexp.MustCompile(`^SE\s??\d{12}$`),
	"SI": regexp.MustCompile(`^SI\s??\d{8}$`),
	"SK": regexp.MustCompile(`^SK\s??\d{10}$`),
}

// https://www.bmf.gv.at/egovernment/fon/fuer-softwarehersteller/BMF_UID_Konstruktionsregeln.pdf
var vatidCheckSum = map[country.Code]func(string) bool{
	"AT": vatidCheckSumAT,
	"DE": vatidCheckSumDE,
}
