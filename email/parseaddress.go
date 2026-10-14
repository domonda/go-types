package email

import (
	"errors"
	"fmt"
	"maps"
	"net/mail"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/domonda/go-errs"

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
	localPart         = `'?(?:[ \t]?(` + `[a-zA-Z0-9\.]` + `|` + atext + `|` + `"(?:[^"\\]|\\.)+"` + `))`
	domainChars       = `a-zA-Z0-9` + umlautChars
	domainPart        = `([` + domainChars + `][\-\.` + domainChars + `]*\.[a-zA-Z]{2,})`
	addressRegex      = atext + `@` + domainPart

	quotedNamePart         = `"([^"]*)"[ \t]*<?`
	unquotedNamePart       = `([^<@]*[^<@\s]|[^<,]*[^<,\s])[ \t]*<`
	rfc2047EncodedNamePart = `((?:` + encodedWord + `[ \t]*)+)<` // Example: =?utf-8?b?wqFIb2xhLCBzZcOxb3Ih?= <
	encodedWord            = `=\?[^?\s]+\?[bBqQ]\?[^?\s]*\?=`
	emptyNamePart          = `<?`
	namePart               = `(?:` + quotedNamePart + `|` + unquotedNamePart + `|` + rfc2047EncodedNamePart + `|` + emptyNamePart + `)`
	nameAddressRegex       = `^` + namePart + localPart + `@` + domainPart + `\s?'?>?`
)

var (
	// AddressRegexp is a compiled regular expression for an email address without name part
	AddressRegexp = regexp.MustCompile(addressRegex)

	// nameAddressRegexp is a regular expression for an email address with name part
	nameAddressRegexp = regexp.MustCompile(nameAddressRegex)

	// unquotedLocalPartRegexp matches a local part that the parser
	// reads back without quotes, see localPart
	unquotedLocalPartRegexp = regexp.MustCompile(`^(?:[a-zA-Z0-9\.]|` + atext + `)$`)

	// encodedWordsRegexp matches a header value that is only
	// RFC 2047 encoded-words separated by whitespace
	encodedWordsRegexp = regexp.MustCompile(`^(?:\s*` + encodedWord + `)+\s*$`)

	// placeholderListRegexp matches a whole lower case list that is only
	// a placeholder for hidden recipients, like "undisclosed-recipients:;",
	// but not an address or a list that only starts with such a word
	placeholderListRegexp = regexp.MustCompile(`^(?:undisclosed|withheld)[- ]recipients(?: *: *;?)?$`)
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

// UniqueNormalizedAddressSlice returns a new slice
// containing the sorted unique normalized address parts
// (address without name part) of the passed addresses.
// Addresses that can't be parsed are skipped.
// Returns nil if there is no valid address.
// The passed slice is not modified.
func UniqueNormalizedAddressSlice(addrs []Address) []Address {
	unique := make(map[Address]struct{}, len(addrs))
	for _, addr := range addrs {
		if norm, err := addr.AddressPart(); err == nil {
			unique[norm] = struct{}{}
		}
	}
	return slices.Sorted(maps.Keys(unique))
}

// ParseAddress parses an email address less strict
// than the standard net/mail.ParseAddress function
// fixing malformed addresses and lower cases the address part.
// If the name part is identical with the address part
// then it will not be returned as name.
// The Address of a quoted local part holds its content without quotes,
// like mail.ParseAddress, so use AddressFrom or Address.AddressPart
// to get the quoted form. Quotes, backslashes and runes that are not
// graphic are removed from the name.
//
// An address without name can be followed by an RFC 5322 comment,
// which becomes the name, like in "mike@doe.com (Mike Doe)".
// The comment can contain nested comments, which stay in the name
// with their parentheses: "mike@doe.com (Mike (Sales))" has the name
// "Mike (Sales)". A backslash escapes a parenthesis, so `\)` doesn't
// end the comment. A comment after the address that isn't closed
// is an error. Parentheses in an unquoted name stay in the name.
func ParseAddress(addr string) (mailAddress *mail.Address, err error) {
	addr = sanitizeAddr(addr)

	if addr == "" {
		return nil, errors.New("empty email address")
	}

	mailAddress, unparsed, ok := parseAddress(addr)
	if !ok {
		return nil, errs.Errorf("could not parse email address: %s", truncateForError(addr))
	}

	if strutil.TrimSpace(unparsed) != "" {
		return nil, fmt.Errorf("parsed email address %s as %s with unexpected remaining characters: %s", addr, mailAddress, unparsed)
	}

	return mailAddress, nil
}

// errInvalidAddress is the error of normalizeAddressForRead.
// It has no call stack like the error of ParseAddress,
// because the callers of normalizeAddressForRead drop it.
var errInvalidAddress = errors.New("invalid email address")

// normalizeAddressForRead is the Normalized method of Address and
// NullableAddress for their methods that keep an address that can't be
// normalized unchanged and drop the error: Scan, UnmarshalJSON,
// UnmarshalXML, UnmarshalXMLAttr, MarshalJSON and MarshalText.
// It returns errInvalidAddress instead of the error of ParseAddress,
// whose call stack costs more than the parsing. Unlike Normalized of
// NullableAddress it returns errInvalidAddress for a blank address,
// which the callers handle before.
func normalizeAddressForRead[S Address | NullableAddress](s S) (S, error) {
	mailAddress, unparsed, ok := parseAddress(sanitizeAddr(string(s)))
	if !ok || strutil.TrimSpace(unparsed) != "" {
		return s, errInvalidAddress
	}
	return S(AddressFrom(mailAddress)), nil
}

// parseAddress parses the address at the start of the sanitized addr
// and returns it with the rest of addr after it, or ok false if addr
// doesn't start with an address. It builds no error, so that a caller
// that drops the error doesn't pay for its call stack.
func parseAddress(addr string) (mailAddress *mail.Address, unparsed string, ok bool) {
	i := nameAddressRegexp.FindStringSubmatchIndex(addr)
	if len(i) != 12 {
		// fmt.Println("REGEX:", nameAddressRegex)
		return nil, "", false
	}

	var name string
	switch {
	case i[3] != -1:
		name = decodeName(addr[i[2]:i[3]])
	case i[5] != -1:
		name = decodeName(addr[i[4]:i[5]])
	case i[7] != -1:
		name = decodeName(addr[i[6]:i[7]])
	}

	local := strings.ToLower(addr[i[8]:i[9]])
	if len(local) >= 2 && local[0] == '"' && local[len(local)-1] == '"' {
		// A quoted local part keeps its content, so the address
		// stays the same mailbox. Like for net/mail.ParseAddress
		// the quotes are not part of mail.Address.Address,
		// addrSpec adds them again where needed.
		local = unquoteLocalPart(local[1 : len(local)-1])
	}

	domain := strings.ToLower(addr[i[10]:i[11]])

	unparsed = addr[i[1]:]
	unparsed = strings.TrimLeft(unparsed, " ")

	// Handle RFC 2822 comment syntax: address (Name)
	// Example: mike@doe.com (Mike Doe)
	// A comment that isn't closed stays unparsed.
	if name == "" {
		if comment, rest, ok := readComment(unparsed); ok {
			name = decodeName(comment)
			unparsed = strings.TrimLeft(rest, " ")
		}
	}

	mailAddress = &mail.Address{
		Name:    name,
		Address: local + "@" + domain,
	}

	// Special case where the address is duplicated in the name part.
	// Example:
	//   "\"Example\" <ar1@example.com>" <ar@example.com>
	// The second address must be within angle brackets, which mark
	// the address of a name, so that "attacker@evil.com ceo@corp.com"
	// is not read as ceo@corp.com, which net/mail rejects.
	if strings.HasPrefix(strings.TrimLeft(unparsed, ` "`), "<") {
		right, unp, ok := parseAddress(unparsed)
		if ok && right.Name == "" {
			mailAddress.Address = right.Address
			unparsed = unp
		}
	}

	if mailAddress.Name == mailAddress.Address {
		mailAddress.Name = ""
	}

	return mailAddress, unparsed, true
}

// readComment returns the content of the RFC 5322 comment at the start
// of s without its outer parentheses, and the rest of s after it.
// ok is false if s doesn't start with a comment or the comment isn't
// closed. Like for SplitAddressList, a comment can contain nested
// comments, and a backslash escapes the next character (quoted-pair),
// so `\)` doesn't close the comment. The content is returned as written,
// cleanName removes the backslashes from a name.
func readComment(s string) (comment, rest string, ok bool) {
	if !strings.HasPrefix(s, "(") {
		return "", s, false
	}
	depth := 0
	for i := 0; i < len(s); i++ {
		// All delimiters are ASCII, so iterating bytes is safe for UTF-8
		switch s[i] {
		case '\\':
			i++
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return s[1:i], s[i+1:], true
			}
		}
	}
	return "", s, false
}

// endsInComment reports if a comment opened in s is still open at the
// end of s, reading s like SplitAddressList: a parenthesis within double
// quotes opens no comment, a double quote within a comment opens no
// quoted string, and within both a backslash escapes the next character.
func endsInComment(s string) bool {
	quoted, depth := false, 0
	for i := 0; i < len(s); i++ {
		// All delimiters are ASCII, so iterating bytes is safe for UTF-8
		c := s[i]
		switch {
		case (quoted || depth > 0) && c == '\\':
			i++
		case quoted:
			quoted = c != '"'
		case c == '(':
			depth++
		case c == ')' && depth > 0:
			depth--
		case c == '"' && depth == 0:
			quoted = true
		}
	}
	return depth > 0
}

// decodeName returns the display name with its RFC 2047 encoded-words
// decoded and cleaned by cleanName. Cleaned after decoding, because an
// encoded-word can hold runes that sanitizeAddr removed from the input.
// A name that can't be decoded, like one in an unknown charset,
// is dropped instead of failing its address, so that no sender
// or recipient is lost because of its name.
func decodeName(name string) string {
	decoded, err := mimeHeaderDecoder.DecodeHeader(name)
	if err != nil {
		return ""
	}
	return cleanName(decoded)
}

// cleanName returns a display name without the quotes and backslashes
// that the parser can't read escaped (quotedNamePart has no escapes),
// without runes that are not graphic, like bidi overrides and control
// characters, and without the replacement character, which sanitizeAddr
// removes when the name is parsed again, and with whitespace mapped
// to spaces.
func cleanName(name string) string {
	return strutil.TrimSpace(strings.Map(
		func(r rune) rune {
			switch {
			case r == '"' || r == '\\' || r == unicode.ReplacementChar:
				return -1
			case unicode.IsSpace(r):
				return ' '
			case !unicode.IsGraphic(r):
				return -1
			}
			return r
		},
		name,
	))
}

// isUnquotedLocalPart reports if the parser reads local back
// without quotes. localPart strips a leading apostrophe.
func isUnquotedLocalPart(local string) bool {
	return !strings.HasPrefix(local, "'") && unquotedLocalPartRegexp.MatchString(local)
}

// quoteLocalPart returns local as quoted string
// with quotes and backslashes escaped as quoted-pairs.
func quoteLocalPart(local string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range local {
		if r == '"' || r == '\\' {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	b.WriteByte('"')
	return b.String()
}

// unquoteLocalPart returns the content of a quoted local part
// with its quoted-pairs (a backslash followed by a rune) unescaped.
// The localPart regex reads a quoted-pair as one unit, so a backslash
// can't escape the closing quote and leave the local part unterminated.
func unquoteLocalPart(quoted string) string {
	if !strings.Contains(quoted, `\`) {
		return quoted
	}
	var b strings.Builder
	escaped := false
	for _, r := range quoted {
		if r == '\\' && !escaped {
			escaped = true
			continue
		}
		escaped = false
		b.WriteRune(r)
	}
	return b.String()
}

// maxAddressListErrors is the number of errors after which
// ParseAddressList stops parsing a list, because every entry that
// can't be parsed is matched against the rest of the list,
// so a long list of garbage would take quadratic time.
const maxAddressListErrors = 100

// truncateForError returns s cut to at most 100 bytes for an error message,
// so that the errors of a long address list don't grow with its length.
func truncateForError(s string) string {
	const maxLen = 100
	if len(s) <= maxLen {
		return s
	}
	cut := maxLen
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut] + "…"
}

// ParseAddressList parses an email address list less strict
// than the standard net/mail.ParseAddressList function
// fixing malformed addresses and lower cases the address part.
//
// All successfully parsed addresses are returned even when other
// addresses in the list can't be parsed. Parsing errors are collected
// and returned joined via errors.Join, so a non-nil error can accompany
// a non-empty result slice. Callers that need a strictly valid list
// should treat any non-nil error as a failure. Parsing stops after
// 100 errors, so addresses after them are not returned.
//
// The empty list and the common "undisclosed-recipients" /
// "withheld-recipients" placeholders, alone or as empty group like
// "undisclosed-recipients:;", are parsed as an empty list without an
// error. Text after a placeholder is parsed like any other list.
// Group syntax like "Team: a@example.com;" is not supported.
//
// Comments are read with nesting like by ParseAddress and
// SplitAddressList: a comma in a comment doesn't separate addresses,
// so a comment that the unquoted name of an address opens and doesn't
// close before the next comma is an error. A comment that isn't closed
// at all holds the rest of the list, like for SplitAddressList, so
// parsing stops at it with an error instead of continuing after the
// next comma like after other errors.
func ParseAddressList(list string) (addrs []*mail.Address, err error) {
	sanitized := sanitizeAddr(list)
	list = strings.TrimRight(sanitized, ", ")
	if list == "" || placeholderListRegexp.MatchString(strings.ToLower(list)) {
		return nil, nil
	}
	// A comma trimmed from the end of the list still follows the last
	// entry, which SplitAddressList reads as part of an unclosed comment
	trailingComma := strings.Contains(sanitized[len(list):], ",")

	var errList []error
	for remaining := list; remaining != ""; {
		if len(errList) >= maxAddressListErrors {
			errList = append(errList, errs.Errorf("stopped parsing email address list after %d errors", len(errList)))
			break
		}
		mailAddress, unparsed, ok := parseAddress(remaining)
		if !ok {
			errList = append(errList, errs.Errorf("could not parse email address list, because of: could not parse email address: %s", truncateForError(remaining)))
			unparsed = remaining
		} else {
			addrs = append(addrs, mailAddress)
			unparsed = strings.TrimLeft(unparsed, " ")
			if unparsed != "" && unparsed[0] != ',' {
				errList = append(errList, errs.Errorf("expected ',' after email address %s in email address list: %s", mailAddress.Address, truncateForError(unparsed)))
			} else if (unparsed != "" || trailingComma) && endsInComment(remaining[:len(remaining)-len(unparsed)]) {
				// Like `Doe (Sales <doe@example.com>, b@example.com`
				errList = append(errList, errs.Errorf("comment not closed before ',' after email address %s in email address list: %s", mailAddress.Address, truncateForError(remaining)))
			} else if unparsed == "" {
				break
			} else {
				remaining = strings.TrimLeft(unparsed[1:], " ")
				continue
			}
		}
		// Recover by skipping past the next separator so the
		// remaining addresses can still be parsed and returned,
		// unless the rest of the list is in a comment that isn't closed.
		if endsInComment(remaining) {
			break
		}
		_, after, found := strings.Cut(unparsed, ",")
		if !found {
			break
		}
		remaining = strings.TrimLeft(after, " ")
	}

	return addrs, errors.Join(errList...)
}
