package vat

import (
	"bytes"
	"database/sql/driver"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"github.com/domonda/errors"
	"github.com/guregu/null"
	command "github.com/ungerik/go-command"

	"github.com/domonda/go-types/country"
	"github.com/domonda/go-types/strutil"
)

const (
	// IDMinLength is the minium length of a VAT ID
	IDMinLength = 4

	// IDMaxLength is the maximum length of a VAT ID
	IDMaxLength = 14 + 2 // allow 2 spaces

	// IDEmpty an empty/invalid VAT ID that will be represented as NULL in SQL
	IDEmpty ID = ""
)

// ID is a european VAT ID
type ID string

// NormalizeVATID returns str as normalized VAT ID or an error.
func NormalizeVATID(str string) (ID, error) {
	return ID(str).Normalized()
}

// StringIsVATID returns if a string can be parsed as VATID.
func StringIsVATID(str string) bool {
	return ID(str).Valid()
}

// BytesAreVATID returns if a byte string is a valid VAT ID
func BytesAreVATID(str []byte) bool {
	l := len(str)
	if l < IDMinLength || l > IDMaxLength {
		return false
	}
	countryCode := country.Code(str[:2])
	regex, found := vatidRegex[countryCode]
	// return found && regex.Match(str)
	if !found || !regex.Match(str) {
		return false
	}
	check, found := vatidCheckSum[countryCode]
	return !found || check(string(str))
}

func isVATIDSplitRune(r rune) bool {
	return unicode.IsSpace(r) || r == ':'
}

func isVATIDTrimRune(r rune) bool {
	return unicode.IsPunct(r)
}

// AssignString implements strfmt.StringAssignable
func (id *ID) AssignString(str string) error {
	normalized, err := ID(str).Normalized()
	if err != nil {
		return err
	}
	*id = normalized
	return nil
}

// Scan implements the database/sql.Scanner interface.
func (id *ID) Scan(value interface{}) error {
	var ns null.String
	err := ns.Scan(value)
	if err != nil {
		return err
	}
	if ns.Valid {
		*id = ID(ns.String)
	} else {
		*id = ""
	}
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface.
func (id ID) Value() (driver.Value, error) {
	if id == "" {
		return nil, nil
	}
	return string(id), nil
}

func (id ID) Valid() bool {
	_, err := id.Normalized()
	return err == nil
}

func (id ID) ValidAndNormalized() bool {
	norm, err := id.Normalized()
	return err == nil && id == norm
}

func (id ID) CountryCode() country.Code {
	if !id.Valid() {
		return ""
	}
	code := country.Code(id[:2])
	if code == "EL" {
		return "GR"
	}
	return code
}

func (id ID) Number() string {
	if !id.Valid() {
		return ""
	}
	return string(id[2:])
}

func (id ID) Normalized() (ID, error) {
	if len(id) < IDMinLength {
		return "", errors.New("VAT ID is too short")
	}
	countryCode := country.Code(id[:2])
	regex, found := vatidRegex[countryCode]
	if !found {
		return "", errors.New("invalid VAT ID country code: " + string(countryCode))
	}
	normalized := strutil.RemoveRunesString(string(id), unicode.IsSpace)
	if !regex.MatchString(normalized) {
		return "", errors.New("invalid VAT ID format")
	}
	check, found := vatidCheckSum[countryCode]
	if found && !check(normalized) {
		return "", errors.New("invalid VAT ID check-sum")
	}
	return ID(normalized), nil
}

var IDMethodWithoutArgs struct {
	command.ArgsDef

	VATID ID `arg:"vatID"`
}

type IDOnlineCheckResult struct {
	ID           ID       `json:"id"`
	Valid        bool     `json:"valid"`
	CountryCode  string   `json:"countryCode,omitempty"`
	Name         string   `json:"name,omitempty"`
	AddressLines []string `json:"address_lines,omitempty"`
}

// Using http://ec.europa.eu/taxation_customs/vies/vatResponse.html
// Mehrwertsteuer-Informations-Austausch-System (MIAS)
func (id ID) OnlineCheck() (*IDOnlineCheckResult, error) {
	vatid, err := id.Normalized()
	if err != nil {
		return &IDOnlineCheckResult{ID: id, Valid: false}, nil
	}
	values := make(url.Values)
	values.Add("memberStateCode", string(vatid.CountryCode()))
	values.Add("number", vatid.Number())
	values.Add("action", "check")
	values.Add("check", "Verify")
	request, err := http.NewRequest("POST", "http://ec.europa.eu/taxation_customs/vies/vatResponse.html", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	// fs.WriteFile("uid.html", data)

	findNext := func(token string) bool {
		pos := bytes.Index(data, []byte(token))
		if pos == -1 {
			return false
		}
		// fmt.Println("XXXX", pos)
		data = data[pos+len(token):]
		return true
	}
	getUntilNext := func(token string) (string, bool) {
		pos := bytes.Index(data, []byte(token))
		if pos == -1 {
			return "", false
		}
		result := string(data[:pos])
		data = data[pos+len(token):]
		return result, true
	}

	if !findNext("vatResponseFormTable") {
		return nil, errors.New("invalid response")
	}

	if bytes.LastIndex(data, []byte("invalid VAT number")) != -1 || !findNext("valid VAT number") {
		return &IDOnlineCheckResult{ID: vatid, Valid: false}, nil
	}

	if !findNext("Member State") {
		return nil, errors.New("invalid response")
	}
	if !findNext("<td>") {
		return nil, errors.New("invalid response")
	}
	country, ok := getUntilNext("</td>")
	if !ok {
		return nil, errors.New("invalid response")
	}

	if !findNext("Name</td>") {
		return nil, errors.New("invalid response")
	}
	if !findNext("<td>") {
		return nil, errors.New("invalid response")
	}
	name, ok := getUntilNext("</td>")
	if !ok {
		return nil, errors.New("invalid response")
	}
	name = strings.TrimSpace(name)

	if !findNext("Address</td>") {
		return nil, errors.New("invalid response")
	}
	if !findNext("<td>") {
		return nil, errors.New("invalid response")
	}
	address, ok := getUntilNext("</td>")
	if !ok {
		return nil, errors.New("invalid response")
	}
	address = strings.TrimSpace(address)
	// address = strings.Replace(address, "<br />", "\n", -1)

	result := &IDOnlineCheckResult{
		ID:           vatid,
		Valid:        true,
		CountryCode:  country,
		Name:         name,
		AddressLines: strings.Split(address, "<br />"),
	}

	return result, nil
}

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
	"HU": regexp.MustCompile(`^HU\s??\d{9}$`),
	"IE": regexp.MustCompile(`^IE\s??(?:\d[0-9A-Z]\d{5}[A-Z])|(?:\d{7}[A-W][A-I])$`),
	"IT": regexp.MustCompile(`^IT\s??\d{11}$`),
	"LT": regexp.MustCompile(`^LT\s??\d{9}\d{3}?$`),
	"LU": regexp.MustCompile(`^LU\s??\d{8}$`),
	"LV": regexp.MustCompile(`^LV\s??\d{11}$`),
	"MT": regexp.MustCompile(`^MT\s??\d{8}$`),
	"NL": regexp.MustCompile(`^NL\s??\d{9}B\d{2}$`),
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

func vatidCheckSumAT(id string) bool {
	nonSpaceCount := 0
	sum := 0
	for _, r := range id {
		if unicode.IsSpace(r) {
			continue
		}
		nonSpaceCount++
		if nonSpaceCount > 3 {
			intVal := int(r - '0')
			if nonSpaceCount == 11 {
				sum := (10 - (sum+4)%10) % 10
				return intVal == sum
			}
			if nonSpaceCount&1 == 0 {
				// C2, C4, C6, C8
				sum += intVal
			} else {
				// C3, C5, C7
				sum += intVal/5 + intVal*2%10
			}
		}
	}
	return false
}

func vatidCheckSumDE(id string) bool {
	nonSpaceCount := 0
	P := 10
	for _, r := range id {
		if unicode.IsSpace(r) {
			continue
		}
		nonSpaceCount++
		if nonSpaceCount > 2 {
			intVal := int(r - '0')
			if nonSpaceCount == 11 {
				// fmt.Println("final C:", string(r), "P:", P)
				return intVal == (11-P)%10
			}
			M := (intVal + P) % 10
			if M == 0 {
				M = 10
			}
			P = (2 * M) % 11
			// fmt.Println("C:", string(r), "P:", P)
		}
	}
	return false
}

var IDFinder idFinder

type idFinder struct{}

func (idFinder) FindAllIndex(str []byte, n int) (indices [][]int) {
	l := len(str)
	if l < IDMinLength {
		return nil
	}

	wordIndices := strutil.SplitAndTrimIndex(str, isVATIDSplitRune, isVATIDTrimRune)
	// fmt.Println("STRING", string(str), wordIndices)

	for begSpace := 0; begSpace < len(wordIndices); begSpace++ {
		for endSpace := begSpace; endSpace < begSpace+3 && endSpace < len(wordIndices); endSpace++ {
			beg := wordIndices[begSpace][0]
			end := wordIndices[endSpace][1]
			// fmt.Println("TEST", str[beg:end])
			if BytesAreVATID(str[beg:end]) {
				indices = append(indices, []int{beg, end})
				break
			}
		}
	}

	return indices
}
