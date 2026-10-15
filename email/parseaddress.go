package email

import (
	"errors"
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
	domainPart        = `([` + domainChars + `](?:\.?[\-` + domainChars + `])*\.[a-zA-Z]{2,})` // no empty label like in f.net..doe, which net/mail rejects
	addressRegex      = atext + `@` + domainPart

	quotedNamePart = `"((?:[^"\\]|\\.)*)"[ \t]*<?` // with quoted-pairs like \" and \\
	// An RFC 2047 encoded-word like =?utf-8?b?wqFIb2xhLCBzZcOxb3Ih?= is read
	// like any other word of an unquoted name, see ParseAddress
	unquotedNamePart = `([^<@]*[^<@\s]|[^<,]*[^<,\s])[ \t]*<`
	encodedWord      = `=\?[^?\s]+\?[bBqQ]\?[^?\s]*\?=`
	emptyNamePart    = `<?`
	namePart         = `(?:` + quotedNamePart + `|` + unquotedNamePart + `|` + emptyNamePart + `)`
	nameAddressRegex = `^` + namePart + localPart + `@` + domainPart + `\s?'?>?`
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

// sanitizeAddr returns s trimmed, with whitespace mapped to spaces and
// without runes that are not graphic, the replacement character and
// invalid UTF-8. A backslash that escapes a removed rune in a quoted
// string or a comment is removed with it, so that it doesn't escape the
// next rune instead: the sanitized text has its quotes, comments and
// commas where net/mail finds them in s, and so the same recipients.
// s is trimmed first, so that a backslash before the runes that
// strutil.TrimSpace removes from its end is kept like by the Value methods
// of the collections, which write the trimmed text: removing it with them
// accepted a text in Validate that Value refused.
func sanitizeAddr(s string) string {
	s = strutil.TrimSpace(s)
	if strings.IndexByte(s, '\\') < 0 {
		return strutil.TrimSpace(strings.Map(sanitizeRune, s))
	}
	return strutil.TrimSpace(sanitizeRunes(s, nil))
}

// sanitizeRunes returns s sanitized like by sanitizeAddr, but not trimmed.
// If offsets is not nil, it is set to the offset in s of the rune that
// each byte of the result comes from, followed by len(s), so that
// hasHiddenAddrSpecRune finds the text of s that the parser read.
// For s trimmed like by strutil.TrimSpace the result is sanitizeAddr(s),
// because the first and the last rune of s are kept.
func sanitizeRunes(s string, offsets *[]int) string {
	sanitized := make([]byte, 0, len(s))
	if offsets != nil {
		*offsets = make([]int, 0, len(s)+1)
	}
	var state listState
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		escaped := state.escaped
		for j := i; j < i+size; j++ {
			state.read(s[j])
		}
		switch mapped := sanitizeRune(r); {
		case mapped == ' ':
			sanitized = append(sanitized, ' ')
			if offsets != nil {
				*offsets = append(*offsets, i)
			}
		case mapped < 0 && escaped:
			// The escaping backslash is the last byte written
			sanitized = sanitized[:len(sanitized)-1]
			if offsets != nil {
				*offsets = (*offsets)[:len(*offsets)-1]
			}
		case mapped >= 0:
			sanitized = append(sanitized, s[i:i+size]...)
			if offsets != nil {
				for j := i; j < i+size; j++ {
					*offsets = append(*offsets, j)
				}
			}
		}
		i += size
	}
	if offsets != nil {
		*offsets = append(*offsets, len(s))
	}
	return string(sanitized)
}

// sanitizeRune maps whitespace to a space and runes that are not graphic
// and the replacement character to -1 for sanitizeAddr.
func sanitizeRune(r rune) rune {
	if strutil.IsSpace(r) {
		return ' '
	}
	if !unicode.IsGraphic(r) || r == unicode.ReplacementChar {
		return -1
	}
	return r
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
// to get the quoted form. A rune that is not ASCII is lower cased only to
// a rune that is not ASCII either, so the Kelvin sign U+212A, which lower
// cases to k, stays in the other mailbox "<U+212A>ate"@corp.com instead of
// reading it as kate@corp.com.
//
// A quoted name is read with its quoted-pairs unescaped like by
// mail.ParseAddress, so `"Say \"Hi\"" <a@b.com>` has the name `Say "Hi"`,
// and an escaped quote doesn't end the name. Quotes and backslashes in an
// unquoted name or a comment are removed, because there they are no part
// of the name: a stray quote of a malformed name or the backslash of an
// escaped parenthesis. RFC 2047 encoded-words are decoded after that,
// their text keeps its quotes and backslashes, but not in a quoted name
// whose every =? is written with a backslash like `=\?`, which holds no
// encoded-word for Python's email package either: AddressFrom writes the
// name "=?utf-8?q?x?=" as `"=\?utf-8?q?x?="`. Runes that are not graphic
// are removed from the name. An encoded-word is read like any other word
// of an unquoted name, also with characters written unencoded that
// RFC 2047 doesn't allow there: an angle bracket in it opens the address,
// and an @ makes the name invalid like in any unquoted name, so
// `=?utf-8?q?a@b.com?= <c@example.com>` is an error.
//
// An @ in an unquoted name, outside its quoted strings and comments, is an
// error, because other parsers read an address there: Python's
// email.utils.getaddresses reads `x@evil.com <ceo@corp.com>,` as the two
// recipients x@evil.com and ceo@corp.com. Only a name that is the address
// itself, like in `a@example.com <a@example.com>`, is valid, and a quoted
// name can hold an @: `"x@evil.com" <ceo@corp.com>`. For the same reason
// these are errors: a second address in angle brackets after the address
// that is not the same address, like in `<x@evil.com> <ceo@corp.com>`,
// a quoted name that is not followed by an angle bracket, which other
// parsers read as part of the local part, like in
// `"x@evil.com" ceo@corp.com`, and an unquoted name with a quoted @ or an
// encoded-word and one of ;:)>[]\, outside its quoted strings and
// comments, where Python's parsers end the name and read its first word,
// decoded, as an address, like in `"x@evil.com"; <ceo@corp.com>`,
// `Team:"x@evil.com":<ceo@corp.com>` and
// `=?utf-8?q?x=40evil.com?=, <ceo@corp.com>`. An encoded-word in an
// unquoted name whose text is ASCII with one of the specials @"()<>[]:;\,
// is an error too, because Python's email package writes such text
// unencoded and unquoted when it writes the header, so that a message to
// `=?utf-8?q?x=40evil.com?= <ceo@corp.com>` holds
// `x@evil.com <ceo@corp.com>`, which is sent to x@evil.com when it is read
// again. It encodes text with other runes again, so
// `=?utf-8?q?J=C3=BCrgen_=28j=40x.de=29?= <j@x.de>` is valid, but writes
// such an encoded-word in a quoted name unquoted too if the name has other
// runes, like in `"ä =?utf-8?q?x=40y.com?=" <a@b.com>`, so it is an error
// in a quoted name too. So is an =? that doesn't start an encoded-word
// that the parser decodes, like one in an unknown charset or with an
// invalid encoding, which Python can still decode, and a local part with
// an encoded-word like =?utf-8?q?admin?=@corp.com, which Python decodes
// and sends to admin@corp.com, or that starts with =?.
// An address within a quoted string or a comment is an error too, also if
// it isn't closed, because other parsers read it as part of them: for them
// `"x@evil.com" "<ceo@corp.com>` has no address in angle brackets.
//
// An address without name can be followed by an RFC 5322 comment,
// which becomes the name, like in "mike@doe.com (Mike Doe)".
// The comment can contain nested comments, which stay in the name
// with their parentheses: "mike@doe.com (Mike (Sales))" has the name
// "Mike (Sales)". A backslash escapes a parenthesis, so `\)` doesn't
// end the comment. A comment after the address that isn't closed
// is an error. Closed parentheses in an unquoted name stay in the name.
func ParseAddress(addr string) (mailAddress *mail.Address, err error) {
	addr = sanitizeAddr(addr)

	if addr == "" {
		return nil, errors.New("empty email address")
	}

	parsed, ok := parseAddress(addr)
	if problem := singleAddressProblem(addr, parsed, ok); problem != "" {
		return nil, errs.Errorf("%s: %s", problem, truncateForError(addr))
	}
	return parsed.addr, nil
}

// singleAddressProblem returns why the sanitized addr, read by parseAddress
// as parsed with ok, is no single address that ParseAddress accepts,
// or "" if it is one.
func singleAddressProblem(addr string, parsed parsedAddress, ok bool) string {
	switch {
	case !ok:
		return "could not parse email address"
	case strutil.TrimSpace(parsed.unparsed) != "":
		return "unexpected characters after the address of email address"
	case parsed.problem != "":
		return parsed.problem + " of email address"
	case withinQuoteOrComment(addr, parsed.specs):
		return "address within a quoted string or comment of email address"
	}
	return ""
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
	addr := sanitizeAddr(string(s))
	parsed, ok := parseAddress(addr)
	if singleAddressProblem(addr, parsed, ok) != "" {
		return s, errInvalidAddress
	}
	return S(AddressFrom(parsed.addr)), nil
}

// specRange is the byte range of the address part of a parsed address,
// from the start of its local part to the end of its domain, or of other
// text that other parsers read as part of an address, see
// parsedAddress.specs, in the sanitized text that was parsed.
type specRange struct{ start, end int }

// parsedAddress is an address that parseAddress read from a sanitized text.
type parsedAddress struct {
	addr *mail.Address
	// unparsed is the rest of the text after the address
	unparsed string
	// spec is the range of the address part in the text
	spec specRange
	// specs are the ranges in the text, in increasing order, of spec and
	// of the other text that other parsers read as part of an address,
	// where the collections refuse a rune that the parser drops, see
	// hasHiddenAddrSpecRune: an unquoted name that is the address, the
	// empty range after a comment after the address, after which Python's
	// email.utils appends a word to the domain, like the no-break space of
	// "a@b.com (x)<U+00A0>, c@d.com", and a repeated address.
	specs []specRange
	// problem is why the display name or the local part makes the address
	// invalid, because other parsers read another address in it, or "":
	// an unquoted name with an @ outside its quoted strings and comments
	// that is not the address itself, or with an address that Python's
	// email package reads in it, see unquotedNameProblem, a quoted name
	// that is not followed by an angle bracket, or a local part with an
	// encoded-word
	problem string
}

// parseAddress parses the address at the start of the sanitized addr
// and returns it with the rest of addr after it, or ok false if addr
// doesn't start with an address. It builds no error, so that a caller
// that drops the error doesn't pay for its call stack.
func parseAddress(addr string) (parsed parsedAddress, ok bool) {
	parsed, ok = parseNameAddress(addr)
	if !ok {
		return parsedAddress{}, false
	}

	// Special case where the address is repeated in angle brackets
	// after it, like in "<a@example.com> <a@example.com>". The second
	// address must be within angle brackets, which mark the address of
	// a name, so that "attacker@evil.com ceo@corp.com" is not read as
	// ceo@corp.com, which net/mail rejects, and it must be the same
	// address as written, because other parsers read both addresses:
	// "<attacker@evil.com> <ceo@corp.com>" was read as ceo@corp.com.
	// A loop instead of a recursion, so that a long chain of addresses
	// doesn't exhaust the stack. The repeated address has no name: only
	// spaces and quotes before its angle bracket, and no name in a comment.
	first := parsed.spec
	for strings.HasPrefix(strings.TrimLeft(parsed.unparsed, ` "`), "<") {
		right, ok := parseNameAddress(parsed.unparsed)
		if !ok || right.addr.Name != "" {
			break
		}
		offset := len(addr) - len(parsed.unparsed)
		spec := specRange{start: offset + right.spec.start, end: offset + right.spec.end}
		if !sameAddressSpec(addr[first.start:first.end], addr[spec.start:spec.end]) {
			break
		}
		// With the empty range after a comment after the repeated address,
		// like the () of `<a@b.com> <a@b.com> ()<U+00A0>, c@d.com`
		for _, s := range right.specs {
			parsed.specs = append(parsed.specs, specRange{start: offset + s.start, end: offset + s.end})
		}
		parsed.unparsed = right.unparsed
	}

	if parsed.addr.Name == parsed.addr.Address {
		parsed.addr.Name = ""
	}
	return parsed, true
}

// parseNameAddress parses the address with its display name or comment
// at the start of the sanitized addr like parseAddress, but without
// reading a repeated address after it.
func parseNameAddress(addr string) (parsed parsedAddress, ok bool) {
	i := nameAddressRegexp.FindStringSubmatchIndex(addr)
	if len(i) != 10 {
		// fmt.Println("REGEX:", nameAddressRegex)
		return parsedAddress{}, false
	}
	spec := specRange{start: i[6], end: i[9]}
	parsed.spec = spec
	parsed.specs = []specRange{spec}

	var name string
	switch {
	case i[3] != -1:
		quoted := addr[i[2]:i[3]]
		name = unescapeQuotedPairs(quoted)
		if !strings.Contains(quoted, "=?") {
			// Every =? is escaped like `=\?`, as AddressFrom writes it,
			// so the name holds no encoded-word
			name = cleanName(name)
		} else {
			// Python's email package writes the decoded text of a quoted
			// name encoded again if it has other runes than ASCII, but an
			// encoded-word with ASCII text in it unencoded and unquoted,
			// so that it sends `"ä =?utf-8?q?x=40y.com?=" <a@b.com>` to
			// `ä x@y.com` once it is read again. Such an encoded-word is
			// refused also in a name of ASCII, which Python keeps quoted,
			// because the stored text can hold other runes than the parser
			// reads, like a soft hyphen. An =? without ?= after it in the
			// name is no encoded-word that Python decodes.
			undecodable, asciiSpecial := readEncodedWords(name)
			if (undecodable && hasEncodedWordText(name)) || asciiSpecial {
				parsed.problem = `encoded-word that can't be decoded or decodes to ASCII with one of @"()<>[]:;\, in quoted display name`
			}
			name = decodeName(name)
		}
		if !strings.Contains(addr[i[3]:spec.start], "<") {
			// Like `"x@evil.com" ceo@corp.com`, which Python's email.utils
			// reads as the one address `"x@evil.com" ceo@corp.com`
			parsed.problem = "quoted display name without angle brackets"
		}
	case i[5] != -1:
		unquoted := addr[i[4]:i[5]]
		name = decodeName(removeQuotes(unquoted))
		switch {
		case hasAtOutsideQuotes(unquoted):
			if !sameAddressSpec(unquoted, addr[spec.start:spec.end]) {
				parsed.problem = "@ in unquoted display name that is not the address"
				break
			}
			// The name is the address, which other parsers read
			// as address, so its runes are checked like those of one
			parsed.specs = []specRange{{start: i[4], end: i[5]}, spec}
		default:
			parsed.problem = unquotedNameProblem(unquoted)
		}
	}

	if rawLocal := addr[i[6]:i[7]]; parsed.problem == "" &&
		(hasEncodedWordText(rawLocal) || strings.HasPrefix(strings.TrimPrefix(rawLocal, `"`), "=?")) {
		// Like =?utf-8?q?admin?=@corp.com and "=?utf-8?q?admin?="@corp.com,
		// which net/mail reads as written, but Python's email package
		// decodes and sends to admin@corp.com. One that starts with =?
		// Python reads up to the ?= after it, also after the address, so
		// it sends `=?utf-8?q?x@evil.com,?= <b@c.com>` to x@evil.com
		// once it is read again.
		parsed.problem = "encoded-word in local part"
	}

	local := lowerAddressPart(addr[i[6]:i[7]])
	if len(local) >= 2 && local[0] == '"' && local[len(local)-1] == '"' {
		// A quoted local part keeps its content, so the address
		// stays the same mailbox. Like for net/mail.ParseAddress
		// the quotes are not part of mail.Address.Address,
		// addrSpec adds them again where needed.
		local = unescapeQuotedPairs(local[1 : len(local)-1])
	}

	domain := lowerAddressPart(addr[i[8]:i[9]])

	unparsed := addr[i[1]:]
	unparsed = strings.TrimLeft(unparsed, " ")

	// Handle RFC 2822 comment syntax: address (Name)
	// Example: mike@doe.com (Mike Doe)
	// A comment that isn't closed stays unparsed.
	if name == "" {
		if comment, rest, ok := readComment(unparsed); ok {
			name = decodeName(removeQuotes(comment))
			end := len(addr) - len(rest)
			parsed.specs = append(parsed.specs, specRange{start: end, end: end})
			unparsed = strings.TrimLeft(rest, " ")
		}
	}

	parsed.addr = &mail.Address{
		Name:    name,
		Address: local + "@" + domain,
	}
	parsed.unparsed = unparsed
	return parsed, true
}

// hasAtOutsideQuotes reports whether text has an @ outside its quoted
// strings and comments, read like nextListSeparator reads them.
func hasAtOutsideQuotes(text string) bool {
	var state listState
	for i := 0; i < len(text); i++ {
		if text[i] == '@' && !state.quoted && state.comment == 0 {
			return true
		}
		state.read(text[i])
	}
	return false
}

// unquotedNameProblem returns why the unquoted display name text without
// an @ outside its quoted strings and comments makes its address invalid,
// because Python's email package reads an address in it, or "":
//
//   - An =? that doesn't start an encoded-word that the parser decodes,
//     which Python can still decode, or an encoded-word whose text is ASCII
//     with one of the specials @"()<>[]:;\, which Python writes unencoded
//     and unquoted when it writes the header, see readEncodedWords.
//   - A quoted @ or an encoded-word, and one of ;:)>[]\, outside the
//     quoted strings and comments, where Python's parsers end the name
//     before its angle bracket and read its first word, decoded, as the
//     address, like the x@evil.com of `"x@evil.com"; <ceo@corp.com>`,
//     `Team:"x@evil.com":<ceo@corp.com>` and
//     `=?utf-8?q?x=40evil.com?=, <ceo@corp.com>`, for which smtplib sends
//     the message to x@evil.com. The comma splits the name only for the
//     encoded-word, the address pattern reads no unquoted name with
//     a comma and an @.
//
// The encoded-words are read in the name as written, because Python reads
// an encoded-word up to its ?= also across a quote, like the `a"b` of
// `"=?utf-8?q?a"b?= <x@y.com>`, and without its quotes and backslashes,
// as the parser decodes it, because Python decodes one in a quoted string
// within the name too, also with an escaped =? like `=\?`, once it wrote
// the quoted string again without the backslash.
func unquotedNameProblem(text string) string {
	var (
		state     listState
		quotedAt  bool
		phraseEnd bool
		unquoted  = removeQuotes(text)
		encoded   = strings.Contains(unquoted, "=?")
	)
	for i := 0; i < len(text); i++ {
		c := text[i]
		switch {
		case state.quoted:
			quotedAt = quotedAt || c == '@'
		case state.comment == 0:
			phraseEnd = phraseEnd || strings.IndexByte(`;:)>[]\,`, c) >= 0
		}
		state.read(c)
	}
	if encoded {
		undecodable, asciiSpecial := readEncodedWords(unquoted)
		if !undecodable && !asciiSpecial {
			undecodable, asciiSpecial = readEncodedWords(text)
		}
		if undecodable || asciiSpecial {
			return `encoded-word that can't be decoded or decodes to ASCII with one of @"()<>[]:;\, in unquoted display name`
		}
	}
	if (quotedAt || encoded) && phraseEnd {
		return `quoted @ or encoded-word and one of ;:)>[]\, in unquoted display name`
	}
	return ""
}

// encodedWordStartRegexp matches an RFC 2047 encoded-word
// at the start of a text, see readEncodedWords.
var encodedWordStartRegexp = regexp.MustCompile(`^` + encodedWord)

// readEncodedWords reads the RFC 2047 encoded-words of text, a display name
// without the quotes and quoted-pairs that the parser removes before it
// decodes the name. undecodable reports an =? that doesn't start an
// encoded-word that the parser decodes, like one in a charset that the
// parser doesn't know, with an invalid encoding, or with a space in it,
// which Python's email package can still decode. asciiSpecial reports an
// encoded-word whose text is ASCII with one of the specials @"()<>[]:;\,
// of RFC 5322 other than the dot. Python writes such text unencoded and
// unquoted when it writes the header, where it is an address, a quoted
// string, a comment or the end of the name: for
// `=?utf-8?q?x=40evil.com?= <ceo@corp.com>` it writes
// `x@evil.com <ceo@corp.com>`, which is sent to x@evil.com when the
// message is read again. Text with other runes or bytes it encodes again,
// so `=?utf-8?q?J=C3=BCrgen_=28j=40x.de=29?=` is no such word.
func readEncodedWords(text string) (undecodable, asciiSpecial bool) {
	for {
		i := strings.Index(text, "=?")
		if i < 0 {
			return false, asciiSpecial
		}
		word := encodedWordStartRegexp.FindString(text[i:])
		if word == "" {
			return true, asciiSpecial
		}
		decoded, err := mimeHeaderDecoder.Decode(word)
		if err != nil {
			return true, asciiSpecial
		}
		if strings.ContainsAny(decoded, `@"()<>[]:;\,`) && isASCII(decoded) {
			asciiSpecial = true
		}
		text = text[i+len(word):]
	}
}

// hasEncodedWordText reports whether s, a local part or display name
// as written, has an =? followed by a ?=, like an RFC 2047 encoded-word.
func hasEncodedWordText(s string) bool {
	i := strings.Index(s, "=?")
	return i >= 0 && strings.Contains(s[i+2:], "?=")
}

// lowerAddressPart returns the local part or domain s lower cased like by
// strings.ToLower, but with the runes that are not ASCII and lower case to
// ASCII kept: the Kelvin sign U+212A, which lower cases to k, and the
// capital I with dot U+0130, which lower cases to i. net/mail and mail
// servers read "<U+212A>ate"@corp.com as another mailbox than kate@corp.com,
// so a recipient check would see another address than the message is
// sent to.
func lowerAddressPart(s string) string {
	if isASCII(s) {
		return strings.ToLower(s)
	}
	return strings.Map(
		func(r rune) rune {
			if lower := unicode.ToLower(r); r < utf8.RuneSelf || lower >= utf8.RuneSelf {
				return lower
			}
			return r
		},
		s,
	)
}

// sameAddressSpec reports whether the address parts a and b, as written,
// are the same address: with the same local part, and the same domain
// ignoring case like the parser lower cases it, see lowerAddressPart.
// strings.EqualFold also folds the Kelvin sign U+212A to k and the long s
// U+017F to s, so `ceo@<U+212A>orp.com <ceo@korp.com>` was valid as the
// address ceo@korp.com, while other parsers read its name as another one.
func sameAddressSpec(a, b string) bool {
	i, j := strings.LastIndexByte(a, '@'), strings.LastIndexByte(b, '@')
	return i >= 0 && j >= 0 && a[:i] == b[:j] && lowerAddressPart(a[i+1:]) == lowerAddressPart(b[j+1:])
}

// readComment returns the content of the RFC 5322 comment at the start
// of s without its outer parentheses, and the rest of s after it.
// ok is false if s doesn't start with a comment or the comment isn't
// closed. Like for SplitAddressList, a comment can contain nested
// comments, and a backslash escapes the next character (quoted-pair),
// so `\)` doesn't close the comment. The content is returned as written,
// removeQuotes removes the backslashes from a name.
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

// decodeName returns the display name with its RFC 2047 encoded-words
// decoded and cleaned by cleanName. Cleaned after decoding, because an
// encoded-word can hold runes that sanitizeAddr removed from the input.
// A name that can't be decoded, like one in an unknown charset, is
// dropped, so that ParseAddressList still returns the address. Such a name
// of an unquoted or quoted display name makes the address invalid,
// because Python's email package can still decode it, see
// unquotedNameProblem, but not of a comment after the address.
func decodeName(name string) string {
	decoded, err := mimeHeaderDecoder.DecodeHeader(name)
	if err != nil {
		return ""
	}
	return cleanName(decoded)
}

// removeQuotes returns an unquoted display name or the content of a
// comment without its quotes and backslashes, which are no part of the
// name there: a stray quote of a malformed name like `Jo"hn <j@x.com>`,
// the quotes of a quoted word within an unquoted name, which net/mail
// removes too, or the backslash that escapes a parenthesis in a comment.
// Only a quoted name holds quotes and backslashes, as quoted-pairs.
// It is called before the encoded-words are decoded, so that their text
// keeps its quotes and backslashes, like the B encoded name that
// AddressFrom writes for a name containing "=?".
func removeQuotes(name string) string {
	if !strings.ContainsAny(name, `"\`) {
		return name
	}
	return strings.Map(
		func(r rune) rune {
			if r == '"' || r == '\\' {
				return -1
			}
			return r
		},
		name,
	)
}

// cleanName returns a display name without runes that are not graphic,
// like bidi overrides and control characters, and without the replacement
// character, which sanitizeAddr removes when the name is parsed again,
// and with whitespace mapped to spaces. Quotes and backslashes are kept,
// AddressFrom escapes them as quoted-pairs.
func cleanName(name string) string {
	return strutil.TrimSpace(strings.Map(
		func(r rune) rune {
			switch {
			case r == unicode.ReplacementChar:
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

// quoteString returns s as quoted string, like a quoted local part or
// display name, with quotes and backslashes escaped as quoted-pairs
// like mail.Address.String escapes them.
func quoteString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		if r == '"' || r == '\\' {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	b.WriteByte('"')
	return b.String()
}

// unescapeQuotedPairs returns the content of a quoted local part or
// display name with its quoted-pairs (a backslash followed by a rune)
// unescaped. The localPart and quotedNamePart regexes read a quoted-pair
// as one unit, so a backslash can't escape the closing quote and leave
// the quoted string unterminated.
func unescapeQuotedPairs(quoted string) string {
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
// ParseAddressList stops parsing a list,
// so that the errors of a long list of garbage stay short.
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
// Quotes, angle brackets and comments are read like by SplitAddressList,
// so that a list without an error has the entries that SplitAddressList
// returns: a comma in them doesn't separate addresses, so an angle bracket
// that an address opens and doesn't close before the next comma is an
// error, like in `Doe <doe@example.com, b@example.com`. One that isn't
// closed at all holds the rest of the list, like for SplitAddressList, so
// parsing stops at it with an error instead of continuing after the next
// comma like after other errors. An address part within a quoted string
// or a comment is an error, also if it isn't closed, and is not returned,
// because other parsers read it as part of the quoted string or comment,
// like in `"Doe <doe@example.com>, b@example.com`, see ParseAddress.
// An address with an invalid display name, like the unquoted
// `x@evil.com <ceo@corp.com>`, is returned with an error.
// An empty entry is an error, a run of them one error, like the leading
// comma of `,<a@example.com>`, which is no display name of the address.
// After an error parsing continues after the entry that SplitAddressList
// finds there, so every address is returned at most once, also when an
// unquoted display name with commas is followed by an error, like in
// `Doe, John <john@example.com>; b@example.com`. Only the last part of an
// entry with such a name that can't be parsed is parsed once more, so the
// address of `Doe, John (john@corp) <j@x.com>` is still returned, and the
// one in angle brackets after a quoted name that the first part opened,
// like the one of `Support, "x\" <a@b.com>, y" <c@d.com>`.
// Every entry is parsed within its end, so that parsing takes time linear
// in the length of the list.
func ParseAddressList(list string) (addrs []*mail.Address, err error) {
	return parseAddressList(sanitizeAddr(list), nil)
}

// parseAddressList is ParseAddressList for a list sanitized by sanitizeAddr.
// If specs is not nil, the ranges of the address parts of the returned
// addresses in sanitized are appended to it, see parsedAddress.specs.
func parseAddressList(sanitized string, specs *[]specRange) (addrs []*mail.Address, err error) {
	list := strings.TrimRight(sanitized, ", ")
	if list == "" || isPlaceholderList(list) {
		return nil, nil
	}
	// A comma trimmed from the end of the list still follows the last
	// entry, which SplitAddressList reads as part of an unclosed quote,
	// angle bracket or comment
	trailingComma := strings.Contains(sanitized[len(list):], ",")

	var (
		errList []error
		// alone is the offset in list up to which every part is an entry
		// of its own, see listEntryEnd
		alone int
	)
	for remaining := list; remaining != ""; {
		if len(errList) >= maxAddressListErrors {
			errList = append(errList, errs.Errorf("stopped parsing email address list after %d errors", len(errList)))
			break
		}
		if remaining[0] == ',' {
			// A run of empty entries, which SplitAddressList skips, is one
			// error, and parsing continues after it. Moving past one of them
			// at a time joined the others again with the next entry that has
			// angle brackets, see listEntryEnd, so that entry was scanned once
			// per empty entry up to the error limit, like the 1 MB of commas
			// before `<a@b.com>` for 6 seconds.
			errList = append(errList, errs.Errorf("empty entry in email address list: %s", truncateForError(remaining)))
			remaining = strings.TrimLeft(remaining, ", ")
			continue
		}
		offset := len(list) - len(remaining)
		// end is the index of the comma where SplitAddressList ends the
		// entry, -1 if the entry holds the rest of the list, and lastPart
		// the index of the last part of an entry with an unquoted display
		// name with commas. The address is parsed within the entry, so
		// that the address pattern doesn't scan the rest of the list
		// after every entry that it can't parse, which made parsing
		// a long list of garbage take seconds.
		var lastPart, end int
		if offset < alone {
			end = nextListSeparator(remaining)
		} else {
			var aloneParts int
			lastPart, end, aloneParts = listEntryEnd(remaining)
			alone = offset + aloneParts
		}
		entryText := remaining
		if end >= 0 {
			entryText = remaining[:end]
		}
		var (
			parsed parsedAddress
			ok     bool
		)
		// An entry without an @ has no address
		// for the address pattern to find
		if strings.IndexByte(entryText, '@') >= 0 {
			parsed, ok = parseAddress(entryText)
		}
		unparsed := strings.TrimLeft(parsed.unparsed, " ")
		entry := entryText[:len(entryText)-len(unparsed)]
		// An address within a quoted string or a comment that the entry
		// opened before it, like the one of `"x@evil.com" "<ceo@corp.com>`,
		// is no address of the entry: other parsers read the rest of the
		// entry within it, also if it isn't closed
		within := ok && withinQuoteOrComment(entryText, parsed.specs)
		switch {
		case !ok || within:
			if within {
				errList = append(errList, errs.Errorf("email address %s within a quoted string or comment in email address list: %s", parsed.addr.Address, truncateForError(remaining)))
			} else {
				errList = append(errList, errs.Errorf("could not parse email address list, because of: could not parse email address: %s", truncateForError(remaining)))
			}
			// After the entry that can't be parsed, so that the parts of
			// its unquoted display name are not parsed again, each with
			// another error. But the last part of an entry with a display
			// name with commas is parsed once without the parts before it,
			// so that the address of a name with a comma and an @, which
			// the address pattern can't read, is still found, like the one
			// of `Doe, John (john@corp) <j@x.com>`, and the address in
			// angle brackets after a quoted string, which an unquoted name
			// before it opened for the address pattern, like the one of
			// `Support, "x\" <a@b.com>, y" <c@d.com>`.
			if lastPart > 0 {
				remaining = strings.TrimLeft(remaining[lastPart:], " ")
				continue
			}
		default:
			// The entry is one entry of SplitAddressList: without a part of
			// a display name it has no comma that separates entries, and the
			// address pattern reads the commas of such a name only as part of
			// an unquoted name up to the angle bracket of the last part, and
			// the first part is not empty, because a run of empty entries is
			// skipped above. The empty first part of `,<a@example.com>` made
			// "," the display name of an entry that SplitAddressList splits.
			addrs = append(addrs, parsed.addr)
			if specs != nil {
				for _, spec := range parsed.specs {
					*specs = append(*specs, specRange{start: offset + spec.start, end: offset + spec.end})
				}
			}
			if parsed.problem != "" {
				// Like `x@evil.com <ceo@corp.com>`, which
				// other parsers read as two addresses
				errList = append(errList, errs.Errorf("%s of email address %s in email address list: %s", parsed.problem, parsed.addr.Address, truncateForError(remaining)))
			}
			if unparsed != "" && unparsed[0] != ',' {
				errList = append(errList, errs.Errorf("expected ',' after email address %s in email address list: %s", parsed.addr.Address, truncateForError(unparsed)))
			} else if open := openAtEnd(entry); open != "" && (unparsed != "" || trailingComma) {
				// Like `x <a@example.com, b@example.com`. Only an angle
				// bracket can be open here, a quote or comment that isn't
				// closed holds an address part, see within.
				errList = append(errList, errs.Errorf("%s not closed before ',' after email address %s in email address list: %s", open, parsed.addr.Address, truncateForError(remaining)))
			}
		}
		// Without a comma the rest of the list is part of this entry,
		// like after a quote, angle bracket or comment that isn't closed.
		if end < 0 {
			break
		}
		remaining = strings.TrimLeft(remaining[end+1:], " ")
	}

	return addrs, errors.Join(errList...)
}

// withinQuoteOrComment reports whether one of the ranges specs of text,
// in increasing order, starts within a quoted string or a comment,
// read like nextListSeparator reads them.
func withinQuoteOrComment(text string, specs []specRange) bool {
	var state listState
	i := 0
	for _, spec := range specs {
		for ; i < spec.start; i++ {
			state.read(text[i])
		}
		if state.quoted || state.comment > 0 {
			return true
		}
	}
	return false
}
