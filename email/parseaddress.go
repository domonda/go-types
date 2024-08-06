package email

import (
	"errors"
	"fmt"
	"mime"
	"net/mail"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unsafe"

	"github.com/domonda/go-types/strutil"
)

// Recommended for debugging: https://regex101.com/
const (
	// RFC2821 and RFC2822 state clearly that only 7bit ASCII characters are allowed
	// but addresses encountered in the wild may contain other characters like umlauts.
	// See also https://www.jochentopf.com/email/chars.html
	// Support depends on the server implementation, we have to be conservative
	// and try to parse everything that's encountered in the wild.
	umlautChars       = `àáâãäåæāăąçćĉċčďđèéêëēĕėęěĝğġģĥħìíîïĩīĭįıðĵķĸĺļľłñńņňŋòóôõöøōŏőœŕŗřśŝşšţťŧùúûüũūŭůűųŵýŷÿźżžþßÀÁÂÃÄÅÆĀĂĄÇĆĈĊČĎĐÈÉÊËĒĔĖĘĚĜĞĠĢĤĦÌÍÎÏĨĪĬĮIÐĴĶĸĹĻĽŁÑŃŅŇŊÒÓÔÕÖØŌŎŐŒŔŖŘŚŜŞŠŢŤŦÙÚÛÜŨŪŬŮŰŲŴÝŶŸŹŻŽÞSS`
	atextSpecialChars = `!#$%&'*+\-/=?^_{|}~` + "`" + umlautChars
	atext             = `[a-zA-Z0-9` + atextSpecialChars + `][\.a-zA-Z0-9` + atextSpecialChars + `]*`
	localPart         = `'?(?:[ \t]?(` + `[a-zA-Z0-9\.]` + `|` + atext + `|` + `"[^"]+"` + `))`
	domainChars       = `a-zA-Z0-9` + umlautChars
	domainPart        = `([` + domainChars + `][\-\.` + domainChars + `]*\.[a-zA-Z]{2,})`
	addressRegex      = atext + `@` + domainPart

	quotedNamePart         = `"([^"]*)"[ \t]*<?`
	unquotedNamePart       = `([^<@]*[^<@\s]|[^<,]*[^<,\s])[ \t]*<`
	rfc2047EncodedNamePart = `=\?[[:ascii:]]+\?[[:ascii:]]+\?[[:ascii:]]+\?=[ \t]*<` // Example: =?utf-8?b?wqFIb2xhLCBzZcOxb3Ih?= <
	emptyNamePart          = `<?`
	namePart               = `(?:` + quotedNamePart + `|` + unquotedNamePart + `|` + rfc2047EncodedNamePart + `|` + emptyNamePart + `)`
	nameAddressRegex       = `^` + namePart + localPart + `@` + domainPart + `\s?'?>?`
)

var (
	// AddressRegexp is a compiled regular expression for an email address without name part
	AddressRegexp = regexp.MustCompile(addressRegex)

	// nameAddressRegexp is a regular expression for an email address with name part
	nameAddressRegexp = regexp.MustCompile(nameAddressRegex)
)

func sanitizeAddr(s string) string {
	return strutil.TrimSpace(
		strings.Map(
			func(r rune) rune {
				if strutil.IsSpace(r) {
					return ' '
				}
				if !unicode.IsGraphic(r) || r == unicode.ReplacementChar {
					return -1
				}
				return r
			},
			s,
		),
	)
}

// FindAllAddresses uses the AddressRegexp to find all
// email addresses without name part in the passed text.
// The addresses are not normalized and returned
// in the order they were found in the text.
func FindAllAddresses(text string) []Address {
	text = sanitizeAddr(text)
	found := AddressRegexp.FindAllString(text, -1)
	return *(*[]Address)(unsafe.Pointer(&found)) //#nosec G103 -- unsafe OK
}

// UniqueNormalizedAddressSlice returns the passed
// Address slice modified to only contain the sorted unique
// normalized address parts (address without name part)
// of the passed addresses.
func UniqueNormalizedAddressSlice(addrs []Address) []Address {
	switch len(addrs) {
	case 0:
		return nil
	case 1:
		norm, err := addrs[0].AddressPart()
		if err != nil {
			return nil
		}
		addrs[0] = norm
		return addrs
	}
	m := make(map[Address]struct{}, len(addrs))
	for _, addr := range addrs {
		if norm, err := addr.AddressPart(); err == nil {
			m[norm] = struct{}{}
		}
	}
	addrs = addrs[:len(m)]
	i := 0
	for addr := range m {
		addrs[i] = addr
		i++
	}
	sort.Slice(addrs, func(i, j int) bool { return addrs[i] < addrs[j] })
	return addrs
}

// ParseAddress parses an email address less strict
// than the standard net/mail.ParseAddress function
// fixing malformed addresses and lower cases the address part.
// If the name part is identical with the address part
// then it will not be returned as name.
func ParseAddress(addr string) (mailAddress *mail.Address, err error) {
	addr = sanitizeAddr(addr)

	if addr == "" {
		return nil, errors.New("empty email address")
	}

	mailAddress, unparsed, err := parseAddress(addr)
	if err != nil {
		return nil, err
	}

	if strutil.TrimSpace(unparsed) != "" {
		return nil, fmt.Errorf("parsed email address %s as %s with unexpected remaining characters: %s", addr, mailAddress, unparsed)
	}

	return mailAddress, nil
}

func parseAddress(addr string) (mailAddress *mail.Address, unparsed string, err error) {
	i := nameAddressRegexp.FindStringSubmatchIndex(addr)
	if len(i) != 10 {
		// fmt.Println("REGEX:", nameAddressRegex)
		return nil, "", fmt.Errorf("could not parse email address: %s", addr)
	}

	var name string
	switch {
	case i[3] != -1:
		name = addr[i[2]:i[3]]
	case i[5] != -1:
		name = addr[i[4]:i[5]]
	}
	if name != "" {
		var dec mime.WordDecoder
		name, err = dec.DecodeHeader(name)
		if err != nil {
			return nil, "", err
		}
		name = strings.ReplaceAll(name, `"`, ``)
		name = strings.ReplaceAll(name, `\`, ``)
		name = strings.ReplaceAll(name, "\t", " ")
		name = strutil.TrimSpace(name)
	}

	local := strings.ToLower(addr[i[6]:i[7]])
	local = strings.ReplaceAll(local, `"`, ``)
	local = strings.ReplaceAll(local, " ", ".")
	local = strings.ReplaceAll(local, ",", ".")

	domain := strings.ToLower(addr[i[8]:i[9]])

	unparsed = addr[i[1]:]
	unparsed = strings.TrimLeft(unparsed, " ")

	mailAddress = &mail.Address{
		Name:    name,
		Address: local + "@" + domain,
	}

	// Special case where the address is duplicated in the name part.
	// Example:
	//   "\"Example\" <ar1@example.com>" <ar@example.com>
	if unparsed != "" && !strings.HasPrefix(strings.TrimLeft(unparsed, " "), ",") {
		right, unp, err := parseAddress(unparsed)
		if err == nil && right.Name == "" {
			mailAddress.Address = right.Address
			unparsed = unp
		}
	}

	if mailAddress.Name == mailAddress.Address {
		mailAddress.Name = ""
	}

	return mailAddress, unparsed, nil
}

// ParseAddressList parses an email address list less strict
// than the standard net/mail.ParseAddressList function
// fixing malformed addresses and lower cases the address part.
// ParseAddressList returns an error if list does not contain
// at least one address.
func ParseAddressList(list string) (addrs []*mail.Address, err error) {
	list = strings.TrimRight(sanitizeAddr(list), ", ")

	switch ll := strings.ToLower(list); {
	case ll == "",
		strings.HasPrefix(ll, "undisclosed-recipients"),
		strings.HasPrefix(ll, "undisclosed recipients"):
		return nil, nil
	}

	mailAddress, unparsed, err := parseAddress(list)
	if err != nil {
		return nil, fmt.Errorf("could not parse email address list '%s', because of: %w", list, err)
	}
	addrs = append(addrs, mailAddress)

	for unparsed != "" {
		if unparsed[0] != ',' {
			return nil, fmt.Errorf("expected ',' after parsing email address in unparsed part: '%s' | full list: '%s'", unparsed, list)
		}
		unparsed = strings.TrimLeft(unparsed[1:], " ")
		mailAddress, unparsed, err = parseAddress(unparsed)
		if err != nil {
			return nil, fmt.Errorf("could not parse email address list '%s', because of: %w", list, err)
		}
		addrs = append(addrs, mailAddress)
	}

	return addrs, nil
}
