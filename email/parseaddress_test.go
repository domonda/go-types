package email

import (
	"fmt"
	"net/mail"
	"reflect"
	"runtime/debug"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/domonda/go-errs"

	"github.com/domonda/go-types/strutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	validEmailAddresses = map[string]*mail.Address{
		`<erik@domonda.com>`:                                            {Name: "", Address: "erik@domonda.com"},
		`"Unger, Erik" <u.erik@domonda.com>`:                            {Name: "Unger, Erik", Address: "u.erik@domonda.com"},
		`Erik Unger <erik@domonda.com>`:                                 {Name: "Erik Unger", Address: "erik@domonda.com"},
		`Erik Unger    <erik@domonda.com>`:                              {Name: "Erik Unger", Address: "erik@domonda.com"},
		`Erik Unger <Erik.Unger@domonda.com>`:                           {Name: "Erik Unger", Address: "erik.unger@domonda.com"},
		`"Erik Unger" <erik@domonda.com>`:                               {Name: "Erik Unger", Address: "erik@domonda.com"},
		`"Erik Unger" <"Erik.Unger"@domonda.com>`:                       {Name: "Erik Unger", Address: "erik.unger@domonda.com"},
		`" Erik Unger " <erik@domonda.com>`:                             {Name: "Erik Unger", Address: "erik@domonda.com"},
		`"@Erik" <erik@domonda.com>`:                                    {Name: "@Erik", Address: "erik@domonda.com"},
		`"erik.unger@domonda.com" <erik@domonda.com>`:                   {Name: "erik.unger@domonda.com", Address: "erik@domonda.com"}, // Use "erik.unger" in name part vs "erik" in address part to test picking up the right part
		`"Erik Unger-Phd </Domonda-IT>" <Erik.Unger-Phd@domonda-it.at>`: {Name: "Erik Unger-Phd </Domonda-IT>", Address: "erik.unger-phd@domonda-it.at"},
		`refill@example24.de`:                                           {Name: "", Address: "refill@example24.de"},
		`x@mail.example.com`:                                            {Name: "", Address: "x@mail.example.com"},
		`x@mail-example.com`:                                            {Name: "", Address: "x@mail-example.com"},
		`er+bill@mail-billwerk.co.uk`:                                   {Name: "", Address: "er+bill@mail-billwerk.co.uk"},
		`some_underscore@msn.com`:                                       {Name: "", Address: "some_underscore@msn.com"},
		`nasa@7examples.com`:                                            {Name: "", Address: "nasa@7examples.com"},
		`a@we-work.com`:                                                 {Name: "", Address: "a@we-work.com"},
		`customerinfo@email.spammers.com`:                               {Name: "", Address: "customerinfo@email.spammers.com"},
		`Domonda < er+vk+baurauslagen+wirklich@domonda.com>`:            {Name: "Domonda", Address: "er+vk+baurauslagen+wirklich@domonda.com"},
		`Domonda < er+vk+baurauslagen+wirklich@domonda.com >`:           {Name: "Domonda", Address: "er+vk+baurauslagen+wirklich@domonda.com"},
		`_underscore@example.com`:                                       {Name: "", Address: "_underscore@example.com"},

		// Special characters:
		`"Unger, Erik"` + "\t<u.erik@domonda.com>":          {Name: "Unger, Erik", Address: "u.erik@domonda.com"},
		"Erik\tUnger <erik@domonda.com>":                    {Name: "Erik Unger", Address: "erik@domonda.com"}, // Replace tabs in name with spaces
		"Erik�Unger <erik@domonda.com>":                     {Name: "ErikUnger", Address: "erik@domonda.com"},
		"Erik\nUnger <�erik@domonda.com�>":                  {Name: "Erik Unger", Address: "erik@domonda.com"},
		`erik.unger@domonda.com <erik.unger@domonda.com>`: {Name: "", Address: "erik.unger@domonda.com"},

		// Quoted local parts keep their content, only lower cased, so the
		// address stays the same mailbox. Like for net/mail.ParseAddress,
		// mail.Address.Address holds the content without the quotes.
		`"scanner@" <"example.at scanner"@example.at>`:     {Name: "scanner@", Address: "example.at scanner@example.at"},
		`"Unger, Erik" <"Unger, Erik"@domonda.com>`:        {Name: "Unger, Erik", Address: "unger, erik@domonda.com"},
		`"\"Example\" <ar1@example.com>" <ar@example.com>`: {Name: `"Example" <ar1@example.com>`, Address: "ar@example.com"}, // Use "ar1" in name part vs "ar" in address part to test picking up the right part, the escaped quotes are part of the name like for net/mail
		`<xy=erik@example.com>`:                            {Name: "", Address: "xy=erik@example.com"},
		// `A!#$%&'*+-/=?^_` + "`" + `{|}~@example.com`:      {Name: "", Address: "a!#$%&'*+-/=?^_`{|}~@example.com"},
		`A!#$%&'*+-/=?^_{|}~@example.com`: {Name: "", Address: "a!#$%&'*+-/=?^_{|}~@example.com"},

		`"Some.Name1@xbüro-yy-zzz.de" <Some.Name@xbüro-yy-zzz.de>`: {Name: "Some.Name1@xbüro-yy-zzz.de", Address: "some.name@xbüro-yy-zzz.de"}, // Use "Some.Name1" in name part vs "Some.Name" in address part to test picking up the right part

		// RFC 2047 encoding
		`=?utf-8?b?wqFIb2xhLCBzZcOxb3Ih?= <senor@hola.com>`: {Name: `¡Hola, señor!`, Address: "senor@hola.com"}, // mime.BEncoding.Encode("utf-8", `¡Hola, señor!`)
		// Non-stdlib charsets decoded via the CharsetReader (Central-European senders)
		`=?iso-8859-2?q?=A3ukasz?= <lukasz@example.pl>`:     {Name: `Łukasz`, Address: "lukasz@example.pl"}, // 0xA3 = Ł in ISO 8859-2
		`=?windows-1250?Q?Wa=B3=EAsa?= <walesa@example.pl>`: {Name: `Wałęsa`, Address: "walesa@example.pl"}, // 0xB3 = ł, 0xEA = ę in Windows-1250

		`'stupid@quoting.me'`:                         {Name: ``, Address: "stupid@quoting.me"},
		`<'stupid@quoting.me'>`:                       {Name: ``, Address: "stupid@quoting.me"},
		`"'stupid@quoting.me'" <'stupid@quoting.me'>`: {Name: `'stupid@quoting.me'`, Address: "stupid@quoting.me"},

		`wow@xx.consulting`: {Name: ``, Address: "wow@xx.consulting"},

		`witha+plus@gmail.com`:        {Name: ``, Address: "witha+plus@gmail.com"},
		`endingwithnums777@gmail.com`: {Name: ``, Address: "endingwithnums777@gmail.com"},

		`YouWon't@belivethisßällm.bHt`:                      {Name: ``, Address: "youwon't@belivethisßällm.bht"},
		`"alte.mücke@united-b.de" <alte.mücke@united-b.de>`: {Name: ``, Address: "alte.mücke@united-b.de"},

		`<some.name@t.pl>`: {Name: ``, Address: "some.name@t.pl"}, // Allow single character domain
		`"Dipl.-Ing. Extra (Wichtig" <some.name@t.pl>`: {Name: `Dipl.-Ing. Extra (Wichtig`, Address: "some.name@t.pl"},

		// `Non standard comma, in name <comma@example.com>`: {Name: `Non standard comma, in name`, Address: "comma@example.com"},

		`"Hello, World | example.com" <hw@example.com>`: {Name: `Hello, World | example.com`, Address: "hw@example.com"},
		`Hello, World | example.com <hw@example.com>`:   {Name: `Hello, World | example.com`, Address: "hw@example.com"},

		// old-style RFC 2822 comment syntax
		`u.erik@domonda.com (Erik Unger)`: {Name: "Erik Unger", Address: "u.erik@domonda.com"},
	}

	invalidEmailAddresses = map[string]struct{}{
		``:             {},
		` `:            {},
		"\t":           {},
		`,`:            {},
		`, `:           {},
		` , `:          {},
		`@`:            {},
		`@domonda.com`: {},
		// `.@domonda.com`: {},
		// `+@domonda.com`:       {}, // allowed?
		// `_@domonda.com`:       {}, // allowed?
		`Hello World!`:           {},
		`erik@`:                  {},
		`erik@domonda.com,`:      {},
		`erik@domonda.com, `:     {},
		`erik@domonda.com , `:    {},
		`.erik@domonda.com`:      {},
		`,erik@domonda.com`:      {},
		`, erik@domonda.com`:     {},
		` , erik@domonda.com`:    {},
		`unger erik@domonda.com`: {},
		// An @ in an unquoted name, see TestParseAddress_AtInUnquotedName
		`@Erik <erik@domonda.com>`:                  {},
		`erik.unger@domonda.com <erik@domonda.com>`: {},
		// `If need of a ''Declaration of Compliance'' please contact us@example.com`: {},
	}

	invalidEmailAddressLists = map[string]struct{}{
		`@`:                      {},
		`@domonda.com`:           {},
		`Hello World!`:           {},
		`erik@`:                  {},
		`.erik@domonda.com`:      {},
		`,erik@domonda.com`:      {},
		`, erik@domonda.com`:     {},
		` , erik@domonda.com`:    {},
		`unger erik@domonda.com`: {},
		// An @ in an unquoted name, see TestParseAddress_AtInUnquotedName
		`@Erik <erik@domonda.com>`:                  {},
		`erik.unger@domonda.com <erik@domonda.com>`: {},
	}
)

func TestParseAddress(t *testing.T) {
	// Print the very long and complex regex strings for debugging with https://regex101.com/
	// Note that ` has to be removed from the regex for https://regex101.com/ to be able to parse it
	// fmt.Fprintln(os.Stderr, "addressRegex:", strings.ReplaceAll(addressRegex, "`", ""))
	// fmt.Fprintln(os.Stderr, "nameAddressRegex:", strings.ReplaceAll(nameAddressRegex, "`", ""))
	// debugRegex := `^` + namePart // + localPart + `@` + domainPart
	// fmt.Fprintln(os.Stderr, "debugRegex:", strings.ReplaceAll(debugRegex, "`", ""))

	var (
		result *mail.Address
		err    error
	)
	// Quick first debug test before trying complete list:
	// result, err = ParseAddress(`"alte.mücke@united-b.de" <alte.mücke@united-b.de>`)
	// require.NoError(t, err)

	// Test very special case: the escaped quotes are part of the name,
	// like for net/mail, so the address in it is no address of the name
	result, err = ParseAddress(`"\"Example\" <ar1@example.com>" <ar@example.com>`)
	if !assert.NoError(t, err, "valid email address") {
		fmt.Println("DEBUG nameAddressRegex:", nameAddressRegex)
	}
	if result.Name != `"Example" <ar1@example.com>` || result.Address != `ar@example.com` {
		t.Fatalf("wrong result: %v", result)
	}

	for addr, expected := range validEmailAddresses {
		t.Run(addr, func(t *testing.T) {
			result, err := ParseAddress(addr)
			if !assert.NoError(t, err, "valid email address") {
				fmt.Println("DEBUG nameAddressRegex:", nameAddressRegex)
			}
			assert.Equal(t, expected, result, "address: %s", addr)
		})
	}

	for addr := range invalidEmailAddresses {
		t.Run(addr, func(t *testing.T) {
			result, err := ParseAddress(addr)
			if err == nil {
				t.Errorf("should not be able to be parsed as email address %s, but got: %s\nRegex: %s", addr, result, nameAddressRegex)
			}
		})
	}
}

// TestParseAddress_CommentNameCleaned protects that a name from the
// RFC 2822 comment syntax "addr (name)" can't turn into another address:
// its quotes and backslashes are removed like from an unquoted name, a
// backslash only escapes the next rune, and the normalized address reads
// back as the same single address with the same name.
func TestParseAddress_CommentNameCleaned(t *testing.T) {
	for addr, want := range map[string]string{
		`a@trusted.com (x" attacker@evil.com, z)`: "x attacker@evil.com, z",
		`a@trusted.com (back\slash)`:              "backslash",
	} {
		parsed, err := ParseAddress(addr)
		require.NoError(t, err, addr)
		assert.Equal(t, want, parsed.Name, addr)
		assert.Equal(t, "a@trusted.com", parsed.Address, addr)

		again, err := ParseAddressList(string(AddressFrom(parsed)))
		require.NoError(t, err, addr)
		assert.Equal(t, []*mail.Address{parsed}, again, addr)
	}
}

// TestParseAddress_QuotedNameEscapes protects that a quoted name is read
// with its quoted-pairs, like net/mail reads it. The parser read no
// escapes, so an escaped quote ended the name, the name lost its quotes
// and backslashes, and the rest of an address list was read as other
// recipients than net/mail reads, see
// TestParseAddressList_EscapedQuoteInName. The normalized address escapes
// the quotes and backslashes again, so it parses back to the same name
// with this parser and with net/mail, and normalizing it again doesn't
// change it.
func TestParseAddress_QuotedNameEscapes(t *testing.T) {
	tests := []struct {
		addr     string
		wantName string
		wantAddr string
		wantNorm Address
	}{
		{addr: `"Say \"Hi\"" <a@b.com>`, wantName: `Say "Hi"`, wantAddr: "a@b.com", wantNorm: `"Say \"Hi\"" <a@b.com>`},
		{addr: `"back\\slash" <a@b.com>`, wantName: `back\slash`, wantAddr: "a@b.com", wantNorm: `"back\\slash" <a@b.com>`},
		{addr: `"\"\"" <a@b.com>`, wantName: `""`, wantAddr: "a@b.com", wantNorm: `"\"\"" <a@b.com>`},
		// A quoted-pair of another rune is that rune
		{addr: `"J\o\h\n" <a@b.com>`, wantName: `John`, wantAddr: "a@b.com", wantNorm: `"John" <a@b.com>`},
		// An address in the name is no address of the name
		{addr: `"\"Example\" <ar1@example.com>" <ar@example.com>`, wantName: `"Example" <ar1@example.com>`, wantAddr: "ar@example.com", wantNorm: `"\"Example\" <ar1@example.com>" <ar@example.com>`},
		{addr: `"x\" <ok@example.com>" <other@example.net>`, wantName: `x" <ok@example.com>`, wantAddr: "other@example.net", wantNorm: `"x\" <ok@example.com>" <other@example.net>`},
		// The text of an encoded-word keeps its quotes and backslashes.
		// Not ASCII text with them, which Python's email package writes
		// unencoded, see TestParseAddress_EncodedWordThatPythonReadsAsAddress
		{addr: `=?utf-8?q?Say_=22H=C3=AF=22?= <a@b.com>`, wantName: `Say "Hï"`, wantAddr: "a@b.com", wantNorm: `"Say \"Hï\"" <a@b.com>`},
		{addr: `=?utf-8?b?w6RcYg==?= <a@b.com>`, wantName: `ä\b`, wantAddr: "a@b.com", wantNorm: `"ä\\b" <a@b.com>`},
		// A non-ASCII name with a backslash, which ASCII must B encode
		{addr: "\"Jürgen \\\\ Müller\" <j@example.com>", wantName: "Jürgen \\ Müller", wantAddr: "j@example.com", wantNorm: "\"Jürgen \\\\ Müller\" <j@example.com>"},
	}
	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			parsed, err := ParseAddress(tt.addr)
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, parsed.Name)
			assert.Equal(t, tt.wantAddr, parsed.Address)
			stdlib, err := mail.ParseAddress(tt.addr)
			require.NoError(t, err)
			assert.Equal(t, stdlib, parsed, "like net/mail")

			norm, err := Address(tt.addr).Normalized()
			require.NoError(t, err)
			assert.Equal(t, tt.wantNorm, norm)
			again, err := norm.Normalized()
			require.NoError(t, err)
			assert.Equal(t, norm, again, "idempotent")
			stdlib, err = mail.ParseAddress(string(norm))
			require.NoError(t, err)
			assert.Equal(t, parsed, stdlib, "net/mail reads the normalized address")

			// The ASCII form for a message header reads back too
			fromASCII, err := ParseAddress(Address(tt.addr).ASCII())
			require.NoError(t, err)
			assert.Equal(t, parsed, fromASCII, "ASCII %s", Address(tt.addr).ASCII())
		})
	}

	// Outside a quoted name quotes and backslashes are no part of the name
	parsed, err := ParseAddress(`John "Johnny" Doe <a@example.com>`)
	require.NoError(t, err)
	assert.Equal(t, "John Johnny Doe", parsed.Name)

	// A quote that isn't closed holds the rest of the text for other
	// parsers, which read no address in angle brackets in it, see
	// TestParseAddress_UnclosedQuoteOrComment
	for _, addr := range []string{
		`Jo"hn <a@example.com>`,
		`"Doe <a@example.com>`,
	} {
		_, err := ParseAddress(addr)
		assert.Error(t, err, addr)
	}
}

// TestParseAddress_UnclosedQuoteOrComment protects the recipient of a
// stored address: a quote or comment that isn't closed holds the rest of
// the text for other parsers, while the parser read an angle bracket in it
// as the start of the address. `"x@evil.com" "<ceo@corp.com>` validated
// as ceo@corp.com, while Python's email.utils.getaddresses reads
// x@evil.com, and `Doe (Sales <doe@example.com>` as doe@example.com, while
// Python reads the address Doe. Address.Validate and the Validate methods
// of the collections refuse them, also as the last entry of a list.
func TestParseAddress_UnclosedQuoteOrComment(t *testing.T) {
	for _, addr := range []string{
		`"x@evil.com" "<ceo@corp.com>`,
		`"Doe <doe@example.com>`,
		`Jo"hn <j@example.com>`,
		`Doe (Sales <doe@example.com>`,
		`Doe (Sales \) <doe@example.com>`,
		`=?utf-8?q?(Sales?= <a@example.com>`,
		// The repeated address after it is within the quote
		`<doe@example.com> "<doe@example.com>`,
	} {
		t.Run(addr, func(t *testing.T) {
			_, err := ParseAddress(addr)
			assert.Error(t, err, "ParseAddress")
			assert.Error(t, Address(addr).Validate(), "Address.Validate")
			assert.Error(t, NullableAddress(addr).Validate(), "NullableAddress.Validate")
			_, err = Address(addr).Normalized()
			assert.Error(t, err, "Normalized")
			for method, err := range validateCollections(Address(addr)) {
				assert.Error(t, err, method)
			}
			// Also as last entry of a list,
			// which no comma after it ends
			list := AddressList("b@example.com, " + addr)
			assert.Error(t, list.Validate(), "last entry of %q", list)
			// The address within the quote or comment is not returned
			parsed, err := list.Parse()
			assert.Error(t, err)
			assert.Equal(t, []string{"b@example.com"}, addressParts(parsed))
		})
	}

	// The quote of the local part closes the quote of the name, so the
	// quotes are balanced, but the address starts within the quoted
	// string `"hn <"`. Python reads one address of the whole text.
	// Found by the differential test against Python.
	for _, addr := range []string{
		`Jo"hn <"a\"b"@example.com>`,
		`Jo"hn <"a\"b"@example.com`,
	} {
		_, err := ParseAddress(addr)
		assert.Error(t, err, addr)
		assert.Error(t, AddressList(addr).Validate(), addr)
		assert.Error(t, MakeAddressSet(Address(addr)).Validate(), addr)
	}

	// Closed, the quote is a name and the comment part of the name.
	// An unclosed angle bracket at the end holds no other address,
	// like for Python, which reads the same address.
	for addr, want := range map[string]*mail.Address{
		`"x@evil.com" <ceo@corp.com>`:   {Name: "x@evil.com", Address: "ceo@corp.com"},
		`Doe (Sales) <doe@example.com>`: {Name: "Doe (Sales)", Address: "doe@example.com"},
		`x <a@example.com`:              {Name: "x", Address: "a@example.com"},
	} {
		parsed, err := ParseAddress(addr)
		require.NoError(t, err, addr)
		assert.Equal(t, want, parsed, addr)
		assert.NoError(t, AddressList("b@example.com, "+addr).Validate(), addr)
	}
}

// TestParseAddress_QuotedNameWithoutAngleBrackets protects the recipient
// of a stored address: a quoted display name was read as name also
// without an angle bracket after it, so `"x@evil.com" ceo@corp.com`
// validated as ceo@corp.com, and `"<a@b.co""m>"@corp.com` as m>@corp.com,
// while Python's email.utils reads a quoted string followed by a word as
// one local part: the one address `"x@evil.com" ceo@corp.com`. They are
// errors now, lenient parsing still returns the address after the name.
// Found by the differential test against Python.
func TestParseAddress_QuotedNameWithoutAngleBrackets(t *testing.T) {
	for _, addr := range []string{
		`"x@evil.com" ceo@corp.com`,
		`"Doe, John" o'brien@example.com`,
		`""a@example.com`,
		`"<a@b.co""m>"@corp.com`,
		`"x@evil.com <a@x.de>, <a"b@evil.com`,
	} {
		t.Run(addr, func(t *testing.T) {
			_, err := ParseAddress(addr)
			assert.Error(t, err, "ParseAddress")
			for method, err := range validateCollections(Address(addr)) {
				assert.Error(t, err, method)
			}
			parsed, err := ParseAddressList(addr)
			assert.Error(t, err)
			assert.Len(t, parsed, 1)
		})
	}

	// A quoted local part is no name, and a quoted name before
	// an angle bracket is valid, also without a space
	for addr, want := range map[string]*mail.Address{
		`"a b"@example.com`:              {Address: "a b@example.com"},
		`"x@evil.com"<ceo@corp.com>`:     {Name: "x@evil.com", Address: "ceo@corp.com"},
		"\"x@evil.com\"\t<ceo@corp.com>": {Name: "x@evil.com", Address: "ceo@corp.com"},
	} {
		parsed, err := ParseAddress(addr)
		require.NoError(t, err, addr)
		assert.Equal(t, want, parsed, addr)
	}
}

// TestParseAddress_QuotedAtInNameThatLenientParsersEnd protects the
// recipients of a stored address against lenient parsers: Python's
// email.utils.getaddresses without strict, the default before the fix of
// CVE-2023-27043, ends a display name at ;)>[ and ] and reads its first
// word as an address, so for `"x@evil.com"; <ceo@corp.com>`, which validated
// as ceo@corp.com, it also sends to x@evil.com. Python's strict parser
// refuses these names. Found by the differential test against Python.
// The default header parser of Python's email package, whose header
// smtplib.send_message reads the recipients from, also ends the name at
// a colon, as the name of a group, and reads a backslash before a quote as
// part of the address: it sends `Team:"x@evil.com":<ceo@corp.com>` only to
// x@evil.com and `\"x@evil.com" <ceo@corp.com>` only to \x@evil.com.
func TestParseAddress_QuotedAtInNameThatLenientParsersEnd(t *testing.T) {
	for _, addr := range []string{
		`"x@evil.com"; <ceo@corp.com>`,
		`"x@evil.com") <ceo@corp.com>`,
		`"x@evil.com" > <ceo@corp.com>`,
		`"x@evil.com" [EXTERNAL] <ceo@corp.com>`,
		`(c) "x@evil.com" John] <ceo@corp.com>`,
		`Team:"x@evil.com":<ceo@corp.com>`,
		`\"x@evil.com" <ceo@corp.com>`,
		`"x@evil.com" \ <ceo@corp.com>`,
	} {
		t.Run(addr, func(t *testing.T) {
			_, err := ParseAddress(addr)
			assert.Error(t, err, "ParseAddress")
			for method, err := range validateCollections(Address(addr)) {
				assert.Error(t, err, method)
			}
			parsed, err := ParseAddressList(addr)
			assert.Error(t, err)
			assert.Equal(t, []string{"ceo@corp.com"}, addressParts(parsed))
		})
	}

	// Without a quoted @ the first word is no address with a domain,
	// and in a comment or quoted string ;)>[] don't end the name
	for _, addr := range []string{
		`John Doe [External] <ceo@corp.com>`,
		`"Doe, John" [External] <ceo@corp.com>`,
		`"x@evil.com" (Sales; EU) <ceo@corp.com>`,
		`"x@evil.com [EXTERNAL]" <ceo@corp.com>`,
		`John "x@evil.com" Doe <ceo@corp.com>`,
		`"x@evil.com" (Team:) <ceo@corp.com>`,
		`"Team: x@evil.com" <ceo@corp.com>`,
	} {
		parsed, err := ParseAddress(addr)
		require.NoError(t, err, addr)
		assert.Equal(t, "ceo@corp.com", parsed.Address, addr)
		for method, err := range validateCollections(Address(addr)) {
			assert.NoError(t, err, "%s of %q", method, addr)
		}
	}
}

// TestParseAddress_EncodedWordThatPythonReadsAsAddress protects the
// recipients of a stored address against Python's email package, whose
// default header parser decodes the encoded-words of a display name: these
// validated as ceo@corp.com, while smtplib.send_message of a message with
// one of them as To header sends to x@evil.com instead, either at once,
// because the parser ends the name at one of ;:)>[]\, and reads its first
// word, decoded, as the address, or once the message is written and read
// again, because Python writes the decoded ASCII text of a name unencoded
// and unquoted, so that `x@evil.com <ceo@corp.com>` is written. Python also
// decodes words that the parser can't, like one in a charset it doesn't
// know or with an invalid encoding, so these are errors too.
func TestParseAddress_EncodedWordThatPythonReadsAsAddress(t *testing.T) {
	for _, addr := range []string{
		// The name ends at one of ;:)>[]\
		`"=?utf-8?q?x=40evil.com?="; <ceo@corp.com>`,
		`=?utf-8?q?x=40evil.com?=; <ceo@corp.com>`,
		`=?utf-8?q?x=40evil.com?=) <ceo@corp.com>`,
		`=?utf-8?q?x=40evil.com?= > <ceo@corp.com>`,
		`=?utf-8?q?x=40evil.com?= [x] <ceo@corp.com>`,
		`=?utf-8?b?eEBldmlsLmNvbQ==?=; <ceo@corp.com>`,
		`Team:=?utf-8?q?x=40evil.com?=:<ceo@corp.com>`,
		// Also in text with other runes, which Python reads as address
		// before it would encode it again, also in an encoded-word right
		// after another one, which Python decodes in a quoted string: it
		// sends the second to `aä x@evil.com`
		`=?utf-8?q?=C3=A4=2C_x=40evil.com?=; <ceo@corp.com>`,
		`"=?utf-8?q?a?==?utf-8?q?=C3=A4_x=40evil.com?="; <ceo@corp.com>`,
		// The decoded ASCII text is written unencoded and unquoted, where
		// its @, quotes, parentheses, angle brackets and :;[]\, are read
		`=?utf-8?q?x=40evil.com?= <ceo@corp.com>`,
		`=?utf-8?b?eEBldmlsLmNvbQ==?= <ceo@corp.com>`,
		`John =?utf-8?q?x=40evil.com?= <ceo@corp.com>`,
		`=?utf-8?q?=C3=A4?= =?utf-8?q?x=40evil.com?= <ceo@corp.com>`,
		// Python decodes an encoded-word right after another one too and
		// writes `ax@evil.com <ceo@corp.com>`
		`=?utf-8?q?a?==?utf-8?q?x=40evil.com?= <ceo@corp.com>`,
		`=?utf-8?q?=3Cx=40evil.com=3E?= <ceo@corp.com>`,
		`"x@evil.com" =?utf-8?q?=3B?= <ceo@corp.com>`,
		`"x@evil.com" =?utf-8?q?=29?= <ceo@corp.com>`,
		`"x@evil.com" =?utf-8?q?=28?= <ceo@corp.com>`,
		`"x@evil.com" =?utf-8?q?=3C?= <ceo@corp.com>`,
		`"x@evil.com" =?utf-8?q?=3E?= <ceo@corp.com>`,
		`"x@evil.com" =?utf-8?q?=5C?= <ceo@corp.com>`,
		`"x@evil.com" =?utf-8?q?=5B?= <ceo@corp.com>`,
		`"x@evil.com" =?utf-8?q?=5D?= <ceo@corp.com>`,
		`"x@evil.com" =?utf-8?q?=2C?= <ceo@corp.com>`,
		`=?utf-8?q?a=22?= <ceo@corp.com>`,
		// Also within quotes next to other runes, which Python encodes
		// again, but not the encoded-word with ASCII text. Python keeps
		// a quoted name of ASCII quoted, but the stored text can hold
		// runes that the parser drops, and Python reads
		// `"=?utf-8?q?x=40evil.com?=<U+00AD>" <ceo@corp.com>` as the
		// address x@evil.com with the soft hyphen encoded after it
		`"ä =?utf-8?q?x=40evil.com?=" <ceo@corp.com>`,
		`"=?utf-8?q?x=40evil.com?= ä" <ceo@corp.com>`,
		`"=?utf-8?q?=C3=A4?= =?utf-8?q?x=40evil.com?=" <ceo@corp.com>`,
		// The colon starts a group, Python sends this one to
		// `b " <x@evil.cöm>, a` instead of ceo@corp.com
		`"=?utf-8?b?YTpi?= \" <x@evil.cöm>, =?utf-8?b?YTpi?= " <ceo@corp.com>`,
		`"=?utf-8?q?x=40evil.com?=" <ceo@corp.com>`,
		`"John =?utf-8?q?x=40evil.com?=" <ceo@corp.com>`,
		`"=?x-unknown?q?x=40evil.com?= ä" <ceo@corp.com>`,
		`"ä =?utf-8?b?eEBldmlsLmNvbQ?=" <ceo@corp.com>`,
		// Python reads an encoded-word up to its ?= across a quote, also
		// one with an escaped =?, once it wrote the quoted string again
		// without the backslash, so it sends the last two to
		// `"x@evil.com <ceo@corp.com>` and `"b! <ceo@corp.com>`
		`"=?utf-8?q?a"b?= <ceo@corp.com>`,
		`"=?utf-8?q?J"=C3=BCrgen?= <ceo@corp.com>`,
		`"=\?utf-8?q?"x=40evil.com?= <ceo@corp.com>`,
		`"=\?utf-8?q?"b?=! <ceo@corp.com>`,
		// Python writes the ASCII text ; unencoded after the quoted @ that
		// it encodes again, where it ends the name: it sends this to the
		// address `x@evil.com ä` once it is read again
		`"x@evil.com ä =?utf-8?q?=3B?=" <ceo@corp.com>`,
		// Python reads the encoded-words of a quoted name with their
		// quoted-pairs, here the backslash of a\b, and sends to `Jürgena\b`
		`"=?utf-8?q?J=C3=BCrgen?= =?utf-8?q?a\b?=" <ceo@corp.com>`,
		// An escaped =? in a quoted string of an unquoted name, which
		// Python writes without the backslash and decodes when it reads
		// the message again
		`"=\?utf-8?q?x=40evil.com?="; <ceo@corp.com>`,
		// Encoded-words that Python decodes, but the parser doesn't
		`=?x-unknown?q?x=40evil.com?= <ceo@corp.com>`,
		`=?utf-7?q?x+AEA-evil.com?= <ceo@corp.com>`,
		`=?utf-8?b?eEBldmlsLmNvbQ?= <ceo@corp.com>`,
		`=?utf-8?q?x=40evil.com ?= <ceo@corp.com>`,
		`=?utf-8?q?x=40evil.com=ZZ?= <ceo@corp.com>`,
		`'=?\=?utf-8?b?eEBldmlsLmNvbQ==?= <ceo@corp.com>`,
	} {
		t.Run(addr, func(t *testing.T) {
			_, err := ParseAddress(addr)
			assert.Error(t, err, "ParseAddress")
			assert.Error(t, NullableAddress(addr).Validate(), "NullableAddress.Validate")
			for method, err := range validateCollections(Address(addr)) {
				assert.Error(t, err, method)
			}
			// Lenient parsing still returns the address in angle brackets
			parsed, err := ParseAddressList(addr)
			assert.Error(t, err)
			assert.Equal(t, []string{"ceo@corp.com"}, addressParts(parsed))
		})
	}
	// A comma after the name, which SplitAddressList joins with the address.
	// Python ends the name at it and sends to its first word, decoded, too,
	// which is an address for these lists.
	for _, list := range []string{
		`=?utf-8?q?x=40evil.com?=, <ceo@corp.com>`,
		`"=\?utf-8?q?x=40evil.com?=", <ceo@corp.com>`,
	} {
		assert.Error(t, AddressList(list).Validate(), list)
		parsed, err := ParseAddressList(list)
		assert.Error(t, err, list)
		assert.Equal(t, []string{"ceo@corp.com"}, addressParts(parsed), list)
	}
	// For this list the word is x without a domain, like the Müller of the
	// valid `Müller, =?utf-8?q?J=C3=BCrgen?= <j@x.de>`, see
	// TestParseAddress_OrdinaryEncodedNames, so it is valid like before the
	// encoded-word checks: a comma alone only ends the name
	assert.NoError(t, AddressList(`=?utf-8?q?x?=, <ceo@corp.com>`).Validate())
	// The decoded quote holds the next entry for Python: it sends this list
	// to x@evil.com once it is read again, not to b@b.cc and ceo@corp.com
	assert.Error(t, AddressList(`=?utf-8?q?a=22?= <ceo@corp.com>, "x@evil.com" <b@b.cc>`).Validate())
	// Python reads an =? in a quoted name up to a ?= after the name, here
	// the one of the local part, and sends to `x@evil.com" <a@corp.com>`
	_, err := ParseAddress(`"=?utf-8?q?x@evil.com" <a?=@corp.com>`)
	assert.Error(t, err)

	// Python writes an encoded-word with other runes than ASCII encoded
	// again, and the decoded text of a comment in the comment, so it reads
	// no address in them. An escaped =? in a quoted name starts no
	// encoded-word for Python until it writes the name again, here one
	// that decodes to x, no address.
	for addr, want := range map[string]*mail.Address{
		`=?utf-8?Q?J=C3=BCrgen_=28juergen=40example.com=29?= <ceo@corp.com>`: {Name: "Jürgen (juergen@example.com)", Address: "ceo@corp.com"},
		`"=?utf-8?q?J=C3=BCrgen_x=40evil.com?=" <ceo@corp.com>`:              {Name: "Jürgen x@evil.com", Address: "ceo@corp.com"},
		`"=\?utf-8?q?x?=" <ceo@corp.com>`:                                    {Name: "=?utf-8?q?x?=", Address: "ceo@corp.com"},
		// A byte that can't be decoded Python writes encoded again too
		`=?utf-8?q?x=40evil.com=FF?= <ceo@corp.com>`:            {Name: "x@evil.com", Address: "ceo@corp.com"},
		`ceo@corp.com (=?utf-8?q?x=40evil.com?=)`:               {Name: "x@evil.com", Address: "ceo@corp.com"},
		`=?utf-8?q?J=C3=BCrgen?= =?utf-8?q?_M?= <ceo@corp.com>`: {Name: "Jürgen M", Address: "ceo@corp.com"},
	} {
		t.Run(addr, func(t *testing.T) {
			parsed, err := ParseAddress(addr)
			require.NoError(t, err)
			assert.Equal(t, want, parsed)
			for method, err := range validateCollections(Address(addr)) {
				assert.NoError(t, err, method)
			}
		})
	}
}

// TestParseAddress_OrdinaryEncodedNames protects ordinary names that mail
// clients write against the checks of encoded-words that Python's email
// package reads as another recipient: these were refused, because an
// encoded-word whose text was ASCII with one of ()<>[]:;\, or any
// encoded-word with one of ;:)>[]\, anywhere in an unquoted name was an
// error, so that ParseMIMEMessage dropped such a From, and stored values
// that were valid before failed Validate, also those that Normalized of
// the last release stored: the first, which is how mail.Address.String
// splits the name into encoded-words of at most 75 characters, and the
// B encoded "Really=?". Python reads no other recipient with a domain in
// them: a ) or [ only ends the name, also in ASCII text that it writes
// unencoded, a ;, [ or , before the encoded-word ends it before any
// address, and an =? within a word is no encoded-word for it.
func TestParseAddress_OrdinaryEncodedNames(t *testing.T) {
	for addr, want := range map[string]Address{
		`=?utf-8?b?SsO8cmdlbiBNw7xsbGVyLCBCdWNoaGFsdHVuZyB1bmQgRmluYW56ZW4gKEFi?= =?utf-8?b?dGVpbHVuZyBOb3JkKQ==?= <j.mueller@example.com>`: `"Jürgen Müller, Buchhaltung und Finanzen (Abteilung Nord)" <j.mueller@example.com>`,
		// The brackets written unencoded in the encoded-word, which
		// RFC 2047 doesn't allow, are part of the word for Python too
		`=?utf-8?q?J=C3=BCrgen_M=C3=BCller_[Extern]?= <j@example.de>`: `"Jürgen Müller [Extern]" <j@example.de>`,
		// Written with an escaped =? like `=\?` that both Python and the
		// parser read as the name Really=?
		`"Really=?" <a@example.com>`:                  `"Really=\?" <a@example.com>`,
		`=?utf-8?b?UmVhbGx5PT8=?= <a@example.com>`:    `"Really=\?" <a@example.com>`,
		`Müller, =?utf-8?q?J=C3=BCrgen?= <j@x.de>`:    `"Müller, Jürgen" <j@x.de>`,
		`Doe; =?utf-8?q?J=C3=BCrgen?= <ceo@corp.com>`: `"Doe; Jürgen" <ceo@corp.com>`,
		`[Ext] =?utf-8?q?J=C3=BCrgen?= <j@x.de>`:      `"[Ext] Jürgen" <j@x.de>`,
		// The first name as Python's email package encodes it
		`=?utf-8?q?J=C3=BCrgen_M=C3=BCller=2C_Buchhaltung_und_Finanzen_=28Abteilung_Nord=29?= <j.mueller@example.com>`: `"Jürgen Müller, Buchhaltung und Finanzen (Abteilung Nord)" <j.mueller@example.com>`,
	} {
		t.Run(addr, func(t *testing.T) {
			require.NoError(t, Address(addr).Validate(), "Address.Validate")
			require.NoError(t, AddressList(addr).Validate(), "AddressList.Validate")
			norm, err := Address(addr).Normalized()
			require.NoError(t, err)
			assert.Equal(t, want, norm)
			// What Value writes reads back as itself
			assert.True(t, norm.ValidAndNormalized(), "ValidAndNormalized of %q", norm)
			assert.NoError(t, AddressList(norm).Validate(), "AddressList.Validate of %q", norm)

			msg, err := ParseMIMEMessage(strings.NewReader(strings.Join([]string{
				"From: " + addr,
				"To: b@example.com",
				"Date: Thu, 22 Jan 2026 16:21:11 +0100",
				"Content-Type: text/plain; charset=utf-8",
				"",
				"Body",
			}, "\r\n")))
			require.NoError(t, err)
			assert.Equal(t, want, msg.From)
		})
	}
}

// TestParseAddress_EncodedWordThatPythonDecodesAfterWriting protects the
// recipients of a stored list against Python's email package, which writes
// a quoted name again without the backslash of an escaped =? like `=\?`,
// and the decoded ASCII text of an encoded-word unencoded, so that an =?
// at the start of a word of the written name starts an encoded-word when
// the message is read again, which Python reads up to the next ?=, also
// one after the name. These validated, and Normalized wrote the first list
// unchanged, but Python wrote it as
// `"=?utf-8?q?x" <y?=@corp.com>, "<evil@evil.com>" <c@corp.com>` and sent
// it to evil@evil.com once it was read again. So a name is refused if, as
// Python writes it, it has an =? that the parser can't decode, or an
// encoded-word that decodes to ASCII with an @ or another encoded-word.
func TestParseAddress_EncodedWordThatPythonDecodesAfterWriting(t *testing.T) {
	for _, list := range []string{
		`"=\?utf-8?q?x" <y?=@corp.com>, "<evil@evil.com>" <c@corp.com>`,
		// A lone =? reads up to a ?= in another entry too: Python sends
		// this list to evil@evil.com after writing it
		`"=\? x" <a@corp.com>, <z?q?z?=@corp.com>, "<evil@evil.com>" <c@corp.com>`,
		`"=\?" <a@corp.com>, <z?q?z?=@corp.com>, "<evil@evil.com>" <c@corp.com>`,
		// The decoded =? is written unencoded, Python reads
		// `=?utf-8?q?x <y?=` as encoded-word after writing the list
		`=?utf-8?b?PT8=?=utf-8?q?x <y?=@corp.com>, "<evil@evil.com>" <c@corp.com>`,
		// Also right after another encoded-word, where it starts a word:
		// Python writes `=?utf-8?q?=C3=A4?==?x <a@corp.com>` and reads the
		// local part ä=?x instead of a@corp.com
		`=?utf-8?q?=C3=A4?= =?utf-8?q?=3D=3Fx?= <a@corp.com>, <z?q?z?=@corp.com>, "<evil@evil.com>" <c@corp.com>`,
		// Python writes the escaped text of the long name as encoded-word
		// `=?utf-8?q?x=40evil=2Ecom=2C?=` and decodes it to x@evil.com,
		// which it writes unencoded once the read header is written again
		`"` + strings.Repeat("ä", 40) + ` =\?utf-8?q?x=40evil=2Ecom=2C?=" <ceo@corp.com>`,
		// The same once Python wrote it, and a name that decodes to an
		// encoded-word, which Python writes as x@evil.com <ceo@corp.com>
		// once the read header is written again
		`"=\?utf-8?q?x=40evil.com?=" <ceo@corp.com>`,
		`"ä =\?utf-8?q?x=40evil.com?=" <ceo@corp.com>`,
		`=?utf-8?q?=3D?= =?utf-8?q?=3Futf-8=3Fq=3Fx=3D40evil.com=3F=3D?= <ceo@corp.com>`,
		`"=?utf-8?q?=3D=3Futf-8=3Fq=3Fx=3D40evil.com=3F=3D?=" <ceo@corp.com>`,
		`"=?utf-8?q?=3D=3F_J=C3=BCrgen?=" <j@x.de>`,
		`"=\? x@evil.com" <ceo@corp.com>`,
	} {
		t.Run(list, func(t *testing.T) {
			assert.Error(t, AddressList(list).Validate(), "AddressList.Validate")
			_, err := AddressList(list).Normalized()
			assert.Error(t, err, "AddressList.Normalized")
			first := AddressList(list).SplitRaw()[0]
			assert.Error(t, first.Validate(), "Address.Validate of %q", first)
		})
	}

	// AddressFrom writes such a name, but in an Address that Validate refuses
	assert.Error(t, AddressFrom(&mail.Address{Name: "=? x", Address: "a@corp.com"}).Validate())

	// An =? within a word Python reads as written, also after writing it
	for _, list := range []string{
		`"Really=?" <a@corp.com>, <z?q?z?=@corp.com>, "<evil@evil.com>" <c@corp.com>`,
		`"Really=\?" <a@corp.com>, <z?q?z?=@corp.com>, "<evil@evil.com>" <c@corp.com>`,
	} {
		parsed, err := ParseAddressList(list)
		require.NoError(t, err, list)
		assert.Equal(t, []string{"a@corp.com", "z?q?z?=@corp.com", "c@corp.com"}, addressParts(parsed), list)
	}
}

// TestParseAddress_EncodedWordInLocalPart protects the recipient of a stored
// address: net/mail and the parser read an encoded-word in a local part as
// written, but Python's email package decodes it, so these validated as
// another mailbox than smtplib.send_message sends to, admin@corp.com.
func TestParseAddress_EncodedWordInLocalPart(t *testing.T) {
	for _, addr := range []string{
		`=?utf-8?q?admin?=@corp.com`,
		`"=?utf-8?q?admin?="@corp.com`,
		`John <=?utf-8?q?admin?=@corp.com>`,
		`=?utf-8?q?admin?=x@corp.com`,
		// Within quotes Python also decodes one after a space,
		// it sends to "a admin"@corp.com
		`"a =?utf-8?q?admin?="@corp.com`,
		// With an escaped =?, which Normalized wrote without quotes
		// as =?utf-8?q?admin?=@corp.com
		`John <"=\?utf-8?q?admin?="@corp.com>`,
		// After a dot Python decodes one too where it reads the local part
		// as words, like after two dots or before a trailing dot: it writes
		// the second as `x., x@evil.com,.@corp.com`, which it sends to
		// x@evil.com. The parser refuses an =? after any dot, also where
		// Python reads the local part as written, like in the third.
		strings.Repeat("a", 64) + `..=?utf-8?q?=2C_x=40evil.com=2C?=@corp.com`,
		`x.=?utf-8?q?=2C_x=40evil.com=2C?=.@corp.com`,
		`x.=?utf-8?q?admin?=@corp.com`,
	} {
		t.Run(addr, func(t *testing.T) {
			_, err := ParseAddress(addr)
			assert.Error(t, err, "ParseAddress")
			for method, err := range validateCollections(Address(addr)) {
				assert.Error(t, err, method)
			}
			// Lenient parsing returns the address as written
			parsed, err := ParseAddressList(addr)
			assert.Error(t, err)
			assert.Len(t, parsed, 1)
		})
	}
	// Python reads an encoded-word at the start of a local part up to the
	// ?= after it, also after the address: it sends the first list to
	// x@evil.com once it is read again, the second one to the one address
	// `x"@evil.com,  <ceo@corp.com>`. Within quotes it also decodes one
	// after a space.
	assert.Error(t, AddressList(`=?utf-8?q?x@evil.com,?= <b@c.com>`).Validate())
	assert.Error(t, AddressList(`"=?utf-8?q?x"@evil.com, ?= <ceo@corp.com>`).Validate())
	assert.Error(t, AddressList(`"a =?utf-8?q?x"@evil.com, ?= <ceo@corp.com>`).Validate())

	// An =? within a word is no encoded-word for Python
	parsed, err := ParseAddress(`a=?b@corp.com`)
	require.NoError(t, err)
	assert.Equal(t, "a=?b@corp.com", parsed.Address)
	parsed, err = ParseAddress(`a=?utf-8?q?b?=@corp.com`)
	require.NoError(t, err)
	assert.Equal(t, "a=?utf-8?q?b?=@corp.com", parsed.Address)
}

// TestLowerAddressPart protects the mailbox of an address: the address part
// was lower cased with strings.ToLower, which maps the Kelvin sign U+212A
// to k and the capital I with dot U+0130 to i, so that the other mailbox
// "<U+212A>ate"@corp.com was read as kate@corp.com, and a recipient check saw
// another address than net/mail and mail servers send to. No rune that is
// not ASCII is lower cased to ASCII, every other rune like by
// strings.ToLower, so the rest of the address part is normalized like
// before.
func TestLowerAddressPart(t *testing.T) {
	for r := rune(utf8.RuneSelf); r <= unicode.MaxRune; r++ {
		if lower := []rune(lowerAddressPart(string(r))); len(lower) != 1 || lower[0] < utf8.RuneSelf {
			t.Fatalf("lowerAddressPart(%U) = %q", r, string(lower))
		}
	}
	assert.Equal(t, "kate", lowerAddressPart("Kate"))
	assert.Equal(t, "\u212Aate", lowerAddressPart("\u212AATE"))
	assert.Equal(t, "\u0130stanbul", lowerAddressPart("\u0130STANBUL"))
	assert.Equal(t, "jürgen.müller", lowerAddressPart("Jürgen.MÜLLER"))

	parsed, err := ParseAddress("\"\u212Aate\"@Corp.com")
	require.NoError(t, err)
	assert.Equal(t, "\u212Aate@corp.com", parsed.Address)
	norm, err := Address("\"\u212Aate\"@Corp.com").Normalized()
	require.NoError(t, err)
	assert.Equal(t, Address("\"\u212Aate\"@corp.com"), norm)
}

// TestSameAddressSpec protects the recipients of a stored address: an
// unquoted name with an @ is valid only if it is the address itself, which
// compared the domains with strings.EqualFold, which also folds the Kelvin
// sign U+212A to k and the long s U+017F to s, so that the name
// ceo@<U+212A>orp.com, another address for other parsers, passed as
// ceo@korp.com. The domains are compared as the parser lower cases them.
func TestSameAddressSpec(t *testing.T) {
	assert.True(t, sameAddressSpec("ceo@Corp.COM", "ceo@corp.com"))
	assert.True(t, sameAddressSpec("ceo@MÜLLER.de", "ceo@müller.de"))
	assert.False(t, sameAddressSpec("Ceo@corp.com", "ceo@corp.com"), "the local part as written")
	assert.False(t, sameAddressSpec("ceo@\u212Aorp.com", "ceo@korp.com"))
	assert.False(t, sameAddressSpec("ceo@corp.\u017Fe", "ceo@corp.se"))

	for _, addr := range []string{
		"ceo@\u212Aorp.com <ceo@korp.com>",
		"ceo@corp.\u017Fe <ceo@corp.se>",
	} {
		_, err := ParseAddress(addr)
		assert.Error(t, err, addr)
		for method, err := range validateCollections(Address(addr)) {
			assert.Error(t, err, "%s of %q", method, addr)
		}
	}
}

// TestParseAddress_RepeatedAddressWithComment protects the recipients of a
// stored list: the runes that the parser drops after a comment after the
// address are refused by the collections, because Python's email.utils
// appends the word after the comment to the domain, see parsedAddress.specs,
// but after a repeated address the comment was not checked, so that the
// no-break space of this list passed AddressList.Validate.
func TestParseAddress_RepeatedAddressWithComment(t *testing.T) {
	const list = "<a@b.com> <a@b.com> ()\u00a0, c@d.com"
	parsed, err := ParseAddressList(list)
	require.NoError(t, err, "the parser reads the no-break space as space")
	assert.Equal(t, []string{"a@b.com", "c@d.com"}, addressParts(parsed))
	assert.Error(t, AddressList(list).Validate())
	assert.NoError(t, AddressList("<a@b.com> <a@b.com> () , c@d.com").Validate())
}

// TestParseAddress_AtInUnquotedName protects the recipients of a stored
// address: an unquoted display name with an @ was read as name, and a
// second address in angle brackets replaced the first one, so that these
// addresses validated as ceo@corp.com, and Normalized of
// `<x@evil.com> <ceo@corp.com>` returned ceo@corp.com, while Python's
// email.utils.getaddresses, which smtplib.send_message uses, reads them
// with a comma after them as the two recipients x@evil.com and
// ceo@corp.com. They are errors now. A quoted name can hold an @, and an
// unquoted name or a second address that is the same address as written,
// with the domain in any case, is the address.
func TestParseAddress_AtInUnquotedName(t *testing.T) {
	for _, addr := range []string{
		`x@evil.com <ceo@corp.com>`,
		`Bob <x@evil.com> <ceo@corp.com>`,
		`<x@evil.com> <ceo@corp.com>`,
		// A repeated address has no name, like before
		`<ceo@corp.com> <ceo@corp.com> (Bob)`,
		`x@evil.com (y) <ceo@corp.com>`,
		`"a,b"@evil.com <ceo@corp.com>`,
		`@Bob <ceo@corp.com>`,
		// An encoded-word is a word of the unquoted name
		`=?utf-8?q?x@evil.com?= <ceo@corp.com>`,
		// The encoded @ too, see TestParseAddress_EncodedWordThatPythonReadsAsAddress
		`=?utf-8?q?x=40evil.com?= <ceo@corp.com>`,
		// The local part of the name is another mailbox
		`Ceo@corp.com <ceo@corp.com>`,
		`<Ceo@corp.com> <ceo@corp.com>`,
	} {
		t.Run(addr, func(t *testing.T) {
			_, err := ParseAddress(addr)
			assert.Error(t, err, "ParseAddress")
			_, err = Address(addr).Normalized()
			assert.Error(t, err, "Normalized")
			assert.Error(t, NullableAddress(addr).Validate(), "NullableAddress.Validate")
			for method, err := range validateCollections(Address(addr)) {
				assert.Error(t, err, method)
			}
			assert.Error(t, AddressList(addr+",").Validate(), "with a comma like for Python")
			// Lenient parsing still returns one of the addresses
			parsed, err := ParseAddressList(addr)
			assert.Error(t, err)
			assert.Len(t, parsed, 1)
		})
	}

	for addr, want := range map[string]*mail.Address{
		`"x@evil.com" <ceo@corp.com>`:          {Name: "x@evil.com", Address: "ceo@corp.com"},
		`John "x@evil.com" Doe <ceo@corp.com>`: {Name: "John x@evil.com Doe", Address: "ceo@corp.com"},
		`John (x@evil.com) <ceo@corp.com>`:     {Name: "John (x@evil.com)", Address: "ceo@corp.com"},
		`ceo@corp.com <ceo@corp.com>`:          {Address: "ceo@corp.com"},
		`ceo@Corp.COM <ceo@corp.com>`:          {Name: "ceo@Corp.COM", Address: "ceo@corp.com"},
		`<ceo@corp.com> <ceo@Corp.com>`:        {Address: "ceo@corp.com"},
		`Boss <ceo@corp.com> <ceo@corp.com>`:   {Name: "Boss", Address: "ceo@corp.com"},
		`"a,b"@corp.com <"a,b"@corp.com>`:      {Address: "a,b@corp.com"},
		// An encoded @ in text with other runes, which Python encodes again
		`=?utf-8?q?x=40evil.com_=C3=A4?= <ceo@corp.com>`: {Name: "x@evil.com ä", Address: "ceo@corp.com"},
	} {
		t.Run(addr, func(t *testing.T) {
			parsed, err := ParseAddress(addr)
			require.NoError(t, err)
			assert.Equal(t, want, parsed)
			for method, err := range validateCollections(Address(addr)) {
				assert.NoError(t, err, method)
			}
		})
	}
}

// TestParseAddressList_EscapedQuoteInName protects the recipients of a
// message: the parser read no escapes in quoted names, so for this list
// AddressList.Validate passed and ParseAddressList returned ok@example.com
// and ok3@example.com, while net/mail, like a mail server, reads
// other@example.net and ok3@example.com. A check of the recipients saw
// other addresses than the message was sent to.
func TestParseAddressList_EscapedQuoteInName(t *testing.T) {
	const list = `"x\" <ok@example.com>, ok2@example.com (," <other@example.net>, "z)" <ok3@example.com>`
	parsed, err := ParseAddressList(list)
	require.NoError(t, err)
	stdlib, err := mail.ParseAddressList(list)
	require.NoError(t, err)
	assert.Equal(t, stdlib, parsed, "like net/mail")
	require.Len(t, parsed, 2)
	assert.Equal(t, "other@example.net", parsed[0].Address)
	assert.Equal(t, "ok3@example.com", parsed[1].Address)

	// SplitAddressList finds the same entries
	raw := SplitAddressList(list)
	require.Len(t, raw, len(parsed))
	for i, entry := range raw {
		norm, err := entry.Normalized()
		require.NoError(t, err)
		assert.Equal(t, AddressFrom(parsed[i]), norm)
	}
}

// TestNormalizeAddressForRead protects that Scan, UnmarshalJSON and the
// other methods that use normalizeAddressForRead to skip building an
// error they drop store and write exactly what Normalized returns:
// the normalized address, or an address that can't be normalized
// unchanged. Only the error differs, it has no call stack.
func TestNormalizeAddressForRead(t *testing.T) {
	addrs := []string{
		"not an address",
		"a@example.com junk",
		`a@example.com (Mike (Sales))`,
		`a@example.com (x (y)`,
		`Doe (Sales <doe@example.com>`,
		"\x00a@example.com\x00",
	}
	for addr := range validEmailAddresses {
		addrs = append(addrs, addr)
	}
	for addr := range invalidEmailAddresses {
		addrs = append(addrs, addr)
	}
	for _, addr := range addrs {
		want, wantErr := Address(addr).Normalized()
		got, err := normalizeAddressForRead(Address(addr))
		assert.Equal(t, want, got, "Address %q", addr)
		if wantErr != nil {
			assert.ErrorIs(t, err, errInvalidAddress, "Address %q", addr)
		} else {
			assert.NoError(t, err, "Address %q", addr)
		}

		if NullableAddress(addr).IsNull() {
			// Normalized returns null without error for a blank address,
			// which the read methods store before normalizing
			continue
		}
		wantNullable, wantErr := NullableAddress(addr).Normalized()
		gotNullable, err := normalizeAddressForRead(NullableAddress(addr))
		assert.Equal(t, wantNullable, gotNullable, "NullableAddress %q", addr)
		assert.Equal(t, wantErr != nil, err != nil, "NullableAddress %q", addr)
	}
}

// TestParseAddress_NestedComment protects that the comment after an
// address is read with RFC 5322 nesting and quoted-pairs, like
// SplitAddressList reads it: a comment that ended at its first ')'
// left the rest as unparsed text or, in a list, as another address.
// A nested comment stays in the name with its parentheses, so that the
// name reads like the comment; the normalized address quotes the name.
func TestParseAddress_NestedComment(t *testing.T) {
	tests := []struct {
		addr     string
		wantName string
		wantNorm Address
	}{
		{addr: `mike@doe.com (Mike Doe)`, wantName: "Mike Doe", wantNorm: `"Mike Doe" <mike@doe.com>`},
		{addr: `mike@doe.com (Mike (Sales))`, wantName: "Mike (Sales)", wantNorm: `"Mike (Sales)" <mike@doe.com>`},
		{addr: `a@example.com (outer (inner, x))`, wantName: "outer (inner, x)", wantNorm: `"outer (inner, x)" <a@example.com>`},
		{addr: `<a@example.com> ((a) b (c (d)))`, wantName: "(a) b (c (d))", wantNorm: `"(a) b (c (d))" <a@example.com>`},
		// An escaped parenthesis doesn't open or close a comment,
		// cleanName removes the backslash like from any name
		{addr: `a@example.com (x \) y)`, wantName: "x ) y", wantNorm: `"x ) y" <a@example.com>`},
		{addr: `a@example.com (x \( y)`, wantName: "x ( y", wantNorm: `"x ( y" <a@example.com>`},
	}
	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			parsed, err := ParseAddress(tt.addr)
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, parsed.Name)
			norm, err := Address(tt.addr).Normalized()
			require.NoError(t, err)
			assert.Equal(t, tt.wantNorm, norm)
			// The quoted name reads back unchanged
			again, err := norm.Normalized()
			require.NoError(t, err)
			assert.Equal(t, norm, again)
		})
	}

	// A comment that isn't closed is not read as name. It was read up
	// to its first ')', so `a@example.com (x (y)` was valid.
	for _, addr := range []string{
		`a@example.com (unclosed`,
		`a@example.com (x (y)`,
		`a@example.com (x \)`,
	} {
		_, err := ParseAddress(addr)
		assert.Error(t, err, addr)
	}
}

// TestParseAddressList_Comments protects that ParseAddressList reads the
// commas in a comment like SplitAddressList: they don't separate
// addresses, and a comment that isn't closed holds the rest of the list.
// Otherwise a list was valid for ParseAddressList with other entries
// than SplitAddressList finds, and AddressList.SplitNormalized and
// SplitRaw disagreed about the same stored list without an error.
func TestParseAddressList_Comments(t *testing.T) {
	tests := []struct {
		list    string
		want    []string // the addresses parts of the parsed addresses
		wantErr bool
	}{
		{list: `a@example.com (outer (inner, x)), b@example.com`, want: []string{"a@example.com", "b@example.com"}},
		{list: `a@example.com (x \), y), b@example.com`, want: []string{"a@example.com", "b@example.com"}},
		{list: `John (Sales, EU) <a@example.com>, b@example.com`, want: []string{"a@example.com", "b@example.com"}},
		// An address within a comment of its name is no address, also if no
		// comma follows: Python's email.utils reads the address Doe with
		// the comment "Sales <doe@example.com>", see
		// TestParseAddress_UnclosedQuoteOrComment
		{list: `b@example.com, Doe (Sales <doe@example.com>`, want: []string{"b@example.com"}, wantErr: true},
		// The comment holds the rest of the list, so b@example.com is no
		// address: it was returned after the comment ended at "(y)"
		{list: `a@example.com (x, (y), b@example.com`, want: []string{"a@example.com"}, wantErr: true},
		{list: `a@example.com, b@example.com (x, c@example.com`, want: []string{"a@example.com", "b@example.com"}, wantErr: true},
		// The comma after doe@example.com is in the comment of its name
		{list: `Doe (Sales <doe@example.com>, b@example.com`, want: nil, wantErr: true},
		{list: `Doe (Sales <doe@example.com>,`, want: nil, wantErr: true},
		// A comment that is closed later doesn't hold the rest of the list,
		// parsing continues after the next comma like after other errors
		{list: `Doe (Sales <doe@example.com>, b), c@example.com`, want: []string{"c@example.com"}, wantErr: true},
		// An entry that can't be parsed is skipped up to the next comma,
		// but not into a comment that it opened and didn't close
		{list: `bad (x, b@example.com`, want: nil, wantErr: true},
		{list: `bad, b@example.com`, want: []string{"b@example.com"}, wantErr: true},
		{list: `"bad (x", b@example.com`, want: []string{"b@example.com"}, wantErr: true},
		// An unclosed comment holds only the rest of the list after it:
		// the valid entries between an earlier error and that comment
		// are separate entries for SplitAddressList and are returned,
		// because Message.Recipients ignores the error and lost them
		{list: `garbage, c@d.com, e@f.com (unclosed`, want: []string{"c@d.com", "e@f.com"}, wantErr: true},
		{list: `a@b.com junk, c@d.com (unclosed`, want: []string{"a@b.com", "c@d.com"}, wantErr: true},
		{list: `bad@@, a@b.cc, c@d.ee, z@y.zz (x`, want: []string{"a@b.cc", "c@d.ee", "z@y.zz"}, wantErr: true},
		// An escaped ')' doesn't close the comment that an unquoted
		// name opened, so the comma is inside it and b is not separate
		{list: `Doe (Sales \) <doe@example.com>, b@example.com`, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.list, func(t *testing.T) {
			addrs, err := ParseAddressList(tt.list)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			var got []string
			for _, addr := range addrs {
				got = append(got, addr.Address)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestParseAddressList_UnclosedQuoteOrAngle protects that ParseAddressList
// reads quotes and angle brackets like SplitAddressList, like comments in
// TestParseAddressList_Comments: a comma after an unclosed quote or angle
// bracket is in it and separates no addresses. `"Doe <a@b.cc>, z@y.zz`
// parsed as two addresses without an error, while SplitAddressList and
// so AddressSet.Scan read one entry and net/mail rejects the list.
// A leading or doubled comma is an empty entry, which became the display
// name "," of the next address. An address within an unclosed quote is
// no address, other parsers read it as part of the quoted string, see
// TestParseAddress_UnclosedQuoteOrComment.
func TestParseAddressList_UnclosedQuoteOrAngle(t *testing.T) {
	tests := []struct {
		list    string
		want    []string // the addresses parts of the parsed addresses
		wantErr bool
	}{
		{list: `"Doe <a@b.cc>, z@y.zz`, want: nil, wantErr: true},
		{list: `x <a@b.cc, z@y.zz`, want: []string{"a@b.cc"}, wantErr: true},
		{list: `Jo"hn <a@example.com>, b@example.com`, want: nil, wantErr: true},
		// The comma trimmed from the end of the list is in the bracket
		{list: `<a@b.cc,`, want: []string{"a@b.cc"}, wantErr: true},
		// Without a comma after it the angle bracket holds no other entry,
		// so the address is valid like for ParseAddress
		{list: `x <a@b.cc`, want: []string{"a@b.cc"}},
		// A quote holds the address also at the end of the list
		{list: `"Doe <a@b.cc>`, want: nil, wantErr: true},
		{list: `a@b.cc, Jo"hn <c@d.ee>`, want: []string{"a@b.cc"}, wantErr: true},
		// An empty entry is no display name
		{list: `,<a@example.com>`, want: []string{"a@example.com"}, wantErr: true},
		{list: `a@example.com,, <b@example.com>`, want: []string{"a@example.com", "b@example.com"}, wantErr: true},
		// After an earlier error, parsing recovers at every comma that
		// separates entries, up to the entry with the unclosed quote or
		// angle bracket, which holds the rest of the list
		{list: `bad, a@b.cc, x <c@d.ee, e@f.gg`, want: []string{"a@b.cc", "c@d.ee"}, wantErr: true},
		{list: `bad, a@b.cc, "e <e@f.com>, g@h.com`, want: []string{"a@b.cc"}, wantErr: true},
		{list: `bad (x, y), a@b.cc, Jo"hn <c@d.ee>, e@f.gg`, want: []string{"a@b.cc"}, wantErr: true},
		// A comma that separates entries for SplitAddressList between the
		// quotes of the address in angle brackets, so the first entry is
		// `x@y.com "<"a`, whose address is returned with an error like one
		// with other text after it. The pattern read the address in angle
		// brackets across the comma before each entry was parsed within
		// its end, and returned no address.
		{list: `x@y.com "<"a,b"@x.com>`, want: []string{"x@y.com"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.list, func(t *testing.T) {
			addrs, err := ParseAddressList(tt.list)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			var got []string
			for _, addr := range addrs {
				got = append(got, addr.Address)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestParseAddressList_NoAddressFromQuoteOrComment protects the recipients
// that lenient parsing returns, like Message.Recipients of inbound mail:
// an address part within a quoted string or comment, closed or not, is
// never returned, because other parsers read it as part of the quoted
// string or comment. For the first list ParseAddressList returned
// ceo@corp.com, which the unquoted name Support opened a quote before,
// and dropped attacker@evil.com, the address in angle brackets after the
// quoted name that Python's email.utils reads. The last part of such an
// entry is parsed once more, so that address is returned now.
func TestParseAddressList_NoAddressFromQuoteOrComment(t *testing.T) {
	tests := []struct {
		list string
		want []string // the address parts of the parsed addresses
	}{
		{list: `Support, "x\" <ceo@corp.com>, y (," <attacker@evil.com>, z@z.com`, want: []string{"attacker@evil.com", "z@z.com"}},
		{list: `Support, "x\" <a@b.com>, y" <c@d.com>`, want: []string{"c@d.com"}},
		{list: `"x@evil.com" "<a@corp.com>`, want: []string{}},
		{list: `b@corp.com, "x@evil.com" "<a@corp.com>`, want: []string{"b@corp.com"}},
		{list: `"x <a@b.com> y", z@z.com`, want: []string{"z@z.com"}},
		{list: `Doe (Sales <doe@example.com>, b), c@example.com`, want: []string{"c@example.com"}},
		{list: `(x <a@b.com>) <c@d.com>, e@f.com`, want: []string{"e@f.com"}},
		{list: `Jo"hn <"a\"b"@example.com>, e@f.com`, want: []string{"e@f.com"}},
	}
	for _, tt := range tests {
		t.Run(tt.list, func(t *testing.T) {
			addrs, err := ParseAddressList(tt.list)
			assert.Error(t, err)
			assert.Equal(t, tt.want, addressParts(addrs))
		})
	}
}

// TestParseAddressList_ErrorAfterDisplayNameWithCommas protects that every
// address is returned at most once. After an error parsing continued at the
// first comma after the start of the entry, which for an unquoted display
// name with commas is inside the address just appended, so the address was
// appended again once per comma of its name, each time with another error.
// With 100 commas the error limit stopped parsing, and the recipients after
// the name were lost. Parsing continues where SplitAddressList ends the
// entry, also after an entry that can't be parsed.
func TestParseAddressList_ErrorAfterDisplayNameWithCommas(t *testing.T) {
	tests := []struct {
		name       string
		list       string
		want       []string // the address parts of the parsed addresses
		wantErrors int
	}{
		{
			name:       "semicolon as separator",
			list:       `Doe, John <john@example.com>; Smith, Anna <anna@example.com>`,
			want:       []string{"john@example.com", "anna@example.com"},
			wantErrors: 1,
		},
		{
			name:       "stray parenthesis",
			list:       `Doe, Jr., John <john@example.com>) , b@example.com`,
			want:       []string{"john@example.com", "b@example.com"},
			wantErrors: 1,
		},
		{
			name:       "more name commas than the error limit",
			list:       strings.Repeat("Doe, ", 5000) + "John <a@b.cc> junk, b@c.dd",
			want:       []string{"a@b.cc", "b@c.dd"},
			wantErrors: 1,
		},
		{
			// No comma after the address, so the rest of the list
			// belongs to its entry
			name:       "text after the address at the end",
			list:       `Doe, Jr., John <john@example.com> junk`,
			want:       []string{"john@example.com"},
			wantErrors: 1,
		},
		{
			name:       "text after the address up to the next name",
			list:       `a@b.com junk, Doe, x <y@z.com> j, k@l.mm`,
			want:       []string{"a@b.com", "y@z.com", "k@l.mm"},
			wantErrors: 2,
		},
		{
			// The entry that can't be parsed is skipped with an error for
			// the entry and one for its last part, which is parsed once
			// more, not with one error per part of its display name
			name:       "invalid address with display name commas",
			list:       strings.Repeat("Doe, ", 500) + "John <bad>, b@example.com",
			want:       []string{"b@example.com"},
			wantErrors: 2,
		},
		{
			// The address pattern can't read a name with a comma and an @,
			// so the address of the last part is returned with an error
			// for the entry, like before parsing continued after it
			name:       "display name with comma and @",
			list:       `Doe, John (john@corp) <j@x.com>, b@example.com`,
			want:       []string{"j@x.com", "b@example.com"},
			wantErrors: 1,
		},
		{
			// An empty entry is an invalid entry of its own,
			// the address after it is parsed
			name:       "doubled comma",
			list:       `x@y.com,, a@b.com`,
			want:       []string{"x@y.com", "a@b.com"},
			wantErrors: 1,
		},
		{
			// The parts without an @ before an address are entries of
			// their own, the parts of a display name after it are not
			name:       "invalid entry before address and display name with commas",
			list:       `bad, worse, a@b.com, Doe, John <j@x.com>`,
			want:       []string{"a@b.com", "j@x.com"},
			wantErrors: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addrs, err := ParseAddressList(tt.list)
			require.Error(t, err)
			var got []string
			for _, addr := range addrs {
				got = append(got, addr.Address)
			}
			assert.Equal(t, tt.want, got)
			joined, ok := err.(interface{ Unwrap() []error })
			require.True(t, ok, "errors.Join error")
			assert.Len(t, joined.Unwrap(), tt.wantErrors)
			// Every address belongs to its own entry of SplitAddressList
			assert.LessOrEqual(t, len(addrs), len(SplitAddressList(tt.list)))
		})
	}
}

// TestParseAddressList_EncodedWordNameHoldsNoAddress protects the recipients
// of a message: an RFC 2047 encoded-word was read up to its closing "?=",
// also with an unencoded comma or @ in it, so that the list below validated
// with the one recipient ceo@corp.example, and SplitAddressList kept it as
// one entry, while net/mail and mail servers read the comma as separator and
// =?utf-8?q?spy@evil.example as another recipient. An encoded-word is read
// like any other word of an unquoted display name, so the parser reads the
// recipients that net/mail reads, and an unencoded @ in it is an @ in an
// unquoted name, see TestParseAddress_AtInUnquotedName.
func TestParseAddressList_EncodedWordNameHoldsNoAddress(t *testing.T) {
	// Python's email package reads these encoded-words up to their ?=,
	// which makes the first recipient x@evil.example and the second one
	// the name of c@x.com, so the local part that starts with =? and the
	// =? of the name `=?utf-8?q?`, which starts no encoded-word for the
	// parser, are errors now, see TestParseAddress_EncodedWordInLocalPart
	// and TestParseAddress_EncodedWordThatPythonReadsAsAddress, but the
	// recipients of net/mail are still returned
	for _, list := range []string{
		`=?utf-8?q?spy@evil.example,?= <ceo@corp.example>`,
		`=?utf-8?q?<a@b.com>,?= <c@x.com>`,
	} {
		t.Run(list, func(t *testing.T) {
			parsed, err := ParseAddressList(list)
			require.Error(t, err)
			stdlib, err := mail.ParseAddressList(list)
			require.NoError(t, err)
			require.Len(t, parsed, len(stdlib), "the recipients of net/mail")
			for i := range stdlib {
				assert.Equal(t, strings.ToLower(stdlib[i].Address), parsed[i].Address)
			}
			assert.Len(t, SplitAddressList(list), len(stdlib), "SplitAddressList finds every recipient")
		})
	}

	// The angle bracket in an encoded-word opens the address, like for
	// net/mail, so the encoded-word is no name of the address after it
	for _, list := range []string{
		`=?utf-8?q?<x?= <c@x.com>`,
		`=?utf-8?q?<x?= <c@x.com>, d@x.com`,
	} {
		_, err := mail.ParseAddressList(list)
		require.Error(t, err, "net/mail")
		_, err = ParseAddressList(list)
		assert.Error(t, err, list)
	}

	// Without a comma the unencoded @ is in the unquoted name, which other
	// parsers read as address, see TestParseAddress_AtInUnquotedName
	_, err := ParseAddressList(`=?utf-8?q?spy@evil.example?= <ceo@corp.example>`)
	assert.Error(t, err)

	// Encoded-words with commas and @ encoded are one name. Not ASCII
	// with an @, see TestParseAddress_EncodedWordThatPythonReadsAsAddress
	for addr, wantName := range map[string]string{
		`=?utf-8?q?J=C3=BCrgen_M=C3=BCller?= <j@x.de>`:    "Jürgen Müller",
		`=?utf-8?q?=C3=A4=40b=2C?= <j@x.de>`:              "ä@b,",
		`=?utf-8?b?wqFIb2xhLCBzZcOxb3Ih?= <j@x.de>`:       "¡Hola, señor!",
		`=?utf-8?q?J=C3=BCrgen?= =?utf-8?q?_M?= <j@x.de>`: "Jürgen M",
	} {
		parsed, err := ParseAddressList(addr)
		require.NoError(t, err, addr)
		require.Len(t, parsed, 1, addr)
		assert.Equal(t, wantName, parsed[0].Name, addr)
		assert.Equal(t, "j@x.de", parsed[0].Address, addr)
	}
}

// TestParseAddress_DomainWithoutEmptyLabel protects that the parser reads
// no domain that net/mail and mail servers reject: `e@f.net..Doe` was valid
// and normalized to e@f.net..doe, which net/mail rejects, so a validated
// address could not be sent to. A local part with consecutive dots is still
// valid, and is quoted when normalized, so that net/mail reads it.
func TestParseAddress_DomainWithoutEmptyLabel(t *testing.T) {
	for _, addr := range []string{
		`e@f.net..Doe`,
		`a@b..com`,
		`Doe <a@b..com>`,
		`a@b.com..`,
	} {
		_, err := ParseAddress(addr)
		assert.Error(t, err, addr)
		assert.Error(t, AddressList(addr).Validate(), addr)
	}

	norm, err := Address(`Doe..John@Example.com`).Normalized()
	require.NoError(t, err)
	assert.Equal(t, Address(`"doe..john"@example.com`), norm)
	stdlib, err := mail.ParseAddress(string(norm))
	require.NoError(t, err, "net/mail reads the normalized address")
	assert.Equal(t, "doe..john@example.com", stdlib.Address)
}

// TestParseAddressList_EscapedDroppedRune protects that a backslash in a
// quoted name escapes the rune after it also when the parser drops that
// rune, like a zero width joiner: it escaped the next rune instead, so
// that the parser read the quote after it as part of the name, while
// net/mail reads it as the end of the name and other recipients after it.
func TestParseAddressList_EscapedDroppedRune(t *testing.T) {
	const list = "\"x\\\u200d\" <ok@example.com>, b@example.com"
	parsed, err := ParseAddressList(list)
	require.NoError(t, err)
	stdlib, err := mail.ParseAddressList(list)
	require.NoError(t, err)
	require.Len(t, parsed, len(stdlib))
	for i := range parsed {
		assert.Equal(t, stdlib[i].Address, parsed[i].Address)
	}
	assert.Equal(t, "x", parsed[0].Name)
}

func TestFindAllAddresses(t *testing.T) {
	tests := []struct {
		text string
		want []Address
	}{
		{
			text: "",
			want: nil,
		},
		{
			text: `Hello@world.com`,
			want: []Address{`Hello@world.com`},
		},
		{
			text: `Some.Name@xbüro-yy-zzz.de`,
			want: []Address{`Some.Name@xbüro-yy-zzz.de`}, // FindAllAddresses does not normalize to lower case
		},
		{
			text: `Hello@world.com for@example.com`,
			want: []Address{`Hello@world.com`, `for@example.com`},
		},
		{
			text: `Me Hello@world.com for nothing! for@example.com some text with@at symbol`,
			want: []Address{`Hello@world.com`, `for@example.com`},
		},
		{
			text: `Me <Hello@world.com> for nothing! <for@example.com> some text with@at symbol`,
			want: []Address{`Hello@world.com`, `for@example.com`},
		},
		{
			text: "Leading spaces Fax +49 (0)66 66 666 666 what@a-test.com\n",
			want: []Address{`what@a-test.com`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			if got := FindAllAddresses(tt.text); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindAllAddresses() = %v, want %v\nUsed regex: %s", got, tt.want, addressRegex)
			}
		})
	}
}

func TestParseAddressList(t *testing.T) {
	emptyLists := []string{
		"",
		"undisclosed-recipients",
		"undisclosed recipients",
		"Undisclosed Recipients",
		"undisclosed-recipients:;",
		"Undisclosed-recipients:;",
		"withheld recipients:;",
		"Withheld recipients:;",
		"withheld-recipients:;",
		"Withheld-recipients:;",
		"undisclosed-recipients: ;",
		"undisclosed-recipients:",
		"undisclosed-recipients:;, ",
	}
	for _, tt := range emptyLists {
		parsed, err := ParseAddressList(tt)
		if err != nil {
			t.Errorf("%q parsing error: %s", tt, err)
		}
		if parsed != nil {
			t.Errorf("%q not parsed as empty list", tt)
		}
	}

	// Test specifics list we had problems with before
	problemLists := map[string]int{
		`<hello@example.com>,`: 1,
		`"\"Example\" <ar1@example.com>" <ar@example.com>, test@example.com`: 2,

		// The quoted name was read without escapes, up to its first escaped quote,
		// so that a quote of the next address closed it and the comma between
		// them was read as part of a name
		`"\"Example\" <ar1@example.com>" <ar@example.com>, "alte.mücke@united-b.de" <alte.mücke@united-b.de>`: 2,
	}
	for problemList, numAddrs := range problemLists {
		parsed, err := ParseAddressList(problemList)
		if err != nil || len(parsed) != numAddrs {
			fmt.Println("Regex:", nameAddressRegex)
			fmt.Println("AddressList:", problemList)
			for _, addr := range parsed {
				fmt.Println(addr)
			}
			t.Fatalf("Error: %v\nlen(parsed): %d, numAddrs: %d\nResult: %v\n\nproblemList: '%s'", err, len(parsed), numAddrs, parsed, problemList)
		}
	}

	// The test lists are created by joining all validEmailAddresses.
	// First they are joined sorted and reverse-sorted by name to
	// create reproducable combinations.
	// Then additional random combinations are also created.

	const numRandomCombinations = 0 // TODO use 100

	// For every list create variations with different separators
	separators := []string{", ", ",", " ,", " , "}

	// Map from joined address-list to source addresses
	// which are also keys of validEmailAddresses
	tests := make(map[string][]string)

	{
		var sortedAddrs []string
		for addr := range validEmailAddresses {
			sortedAddrs = append(sortedAddrs, addr)
		}
		sort.Strings(sortedAddrs)
		for _, separator := range separators {
			list := strings.Join(sortedAddrs, separator)
			tests[list] = sortedAddrs
		}
	}

	{
		var reverseSortedAddrs []string
		for addr := range validEmailAddresses {
			reverseSortedAddrs = append(reverseSortedAddrs, addr)
		}
		sort.Sort(sort.Reverse(sort.StringSlice(reverseSortedAddrs)))
		for _, separator := range separators {
			list := strings.Join(reverseSortedAddrs, separator)
			tests[list] = reverseSortedAddrs
		}
	}

	for range numRandomCombinations {
		// Every range over the validEmailAddresses map
		// will produce a new random order
		var randomAddrs []string
		for addr := range validEmailAddresses {
			randomAddrs = append(randomAddrs, addr)
		}
		for _, separator := range separators {
			list := strings.Join(randomAddrs, separator)
			tests[list] = randomAddrs
		}
	}

	// Run all prepared tests
	for addressList, sourceAddrs := range tests {
		t.Run(addressList, func(t *testing.T) {
			// fmt.Println("\nLIST:", addressList)
			parsed, err := ParseAddressList(addressList)
			if err != nil {
				t.Fatalf("Error: %v\n\nAddressList: %s\n\nRegex: %s\n", err, addressList, nameAddressRegex)
			}
			if len(parsed) != len(sourceAddrs) {
				for _, addr := range parsed {
					fmt.Println(addr)
				}
				t.Fatalf("len(parsed):%d != len(sourceAddrs):%d", len(parsed), len(sourceAddrs))
			}
			for i := range parsed {
				parsedAddr := parsed[i]
				sourceAddr := sourceAddrs[i]
				expected := validEmailAddresses[sourceAddr]
				if expected.Address != parsedAddr.Address || expected.Name != parsedAddr.Name {
					t.Errorf("expected %#v but got %#v from list '%s'", expected, parsedAddr, addressList)
				}
			}
		})
	}

	// ParseAddressList should be able to parse single addresses like ParseEmailAddress
	for addr, expected := range validEmailAddresses {
		t.Run(addr, func(t *testing.T) {
			results, err := ParseAddressList(addr)
			assert.NoError(t, err, "valid email address")
			if assert.Len(t, results, 1, "parsed list length 1 for 1 address") {
				assert.Equal(t, expected, results[0], "expected: '%s'", expected)
			}
		})
	}

	// Invalid addresses are also invalid lists
	// except for empty trimmed strings wich are an empty list
	for l := range invalidEmailAddressLists {
		t.Run(l, func(t *testing.T) {
			result, err := ParseAddressList(l)
			if err == nil {
				t.Errorf("should not be able to be parsed as email address list: %s, but got: %s\nRegex: %s", l, result, nameAddressRegex)
			}
		})
	}
}

// TestParseAddressList_PartialResult verifies that ParseAddressList collects
// parsing errors, joins them, and still returns all addresses that could be
// parsed even when other entries in the list are invalid.
func TestParseAddressList_PartialResult(t *testing.T) {
	t.Run("invalid entry between valid ones", func(t *testing.T) {
		addrs, err := ParseAddressList("alice@example.com, not a valid address, bob@example.com")
		require.Error(t, err, "the invalid middle entry must produce an error")

		got := make([]string, len(addrs))
		for i, a := range addrs {
			got[i] = a.Address
		}
		require.Equal(t, []string{"alice@example.com", "bob@example.com"}, got,
			"the valid addresses must still be returned")
	})

	t.Run("leading invalid entry", func(t *testing.T) {
		addrs, err := ParseAddressList("@broken, valid@example.com")
		require.Error(t, err)
		require.Len(t, addrs, 1)
		require.Equal(t, "valid@example.com", addrs[0].Address)
	})

	t.Run("all valid has no error", func(t *testing.T) {
		addrs, err := ParseAddressList("alice@example.com, bob@example.com")
		require.NoError(t, err)
		require.Len(t, addrs, 2)
	})

	t.Run("all invalid returns nil and error", func(t *testing.T) {
		addrs, err := ParseAddressList("@broken, also broken")
		require.Error(t, err)
		require.Nil(t, addrs)
	})
}

// TestUniqueNormalizedAddressSlice protects that the result is a sorted set
// of bare lower case mailboxes: duplicates that only differ in case or
// display name collapse, and unparsable addresses are dropped instead of
// being passed on as recipients.
func TestUniqueNormalizedAddressSlice(t *testing.T) {
	tests := []struct {
		name  string
		addrs []Address
		want  []Address
	}{
		{name: "nil", addrs: nil, want: nil},
		{name: "empty", addrs: []Address{}, want: nil},
		{name: "single", addrs: []Address{`John <John@Example.com>`}, want: []Address{`john@example.com`}},
		{name: "single invalid", addrs: []Address{`not an address`}, want: nil},
		{
			name:  "duplicates, sorting and invalid",
			addrs: []Address{`b@example.com`, `not an address`, `A@Example.com`, `"Someone" <a@example.com>`, `B <B@example.com>`},
			want:  []Address{`a@example.com`, `b@example.com`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := slices.Clone(tt.addrs)
			assert.Equal(t, tt.want, UniqueNormalizedAddressSlice(tt.addrs))
			// The caller's slice must not be modified, it typically still
			// holds the original recipients with their display names.
			assert.Equal(t, input, tt.addrs, "input slice modified")
		})
	}
}

// TestParseAddressList_GarbageAfterAddress protects that unparsable text
// following an address (instead of the ',' separator) is reported as error
// while the addresses before and after it are still returned.
func TestParseAddressList_GarbageAfterAddress(t *testing.T) {
	addrs, err := ParseAddressList("a@example.com junk, c@example.com")
	require.Error(t, err)
	require.Len(t, addrs, 2)
	require.Equal(t, "a@example.com", addrs[0].Address)
	require.Equal(t, "c@example.com", addrs[1].Address)

	addrs, err = ParseAddressList("a@example.com junk")
	require.Error(t, err, "trailing garbage without another separator")
	require.Len(t, addrs, 1)
	require.Equal(t, "a@example.com", addrs[0].Address)
}

// TestParseAddressList_LongGarbageList protects the parsing of inbound mail
// from a header with many unparsable entries: every entry was matched
// against the rest of the list and its error held the whole list, so
// 10,000 entries took seconds and hundreds of megabytes of error text.
// Parsing stops after maxAddressListErrors and the errors stay short.
func TestParseAddressList_LongGarbageList(t *testing.T) {
	list := "a@example.com, " + strings.Repeat("x, ", 10000) + "b@example.com"
	addrs, err := ParseAddressList(list)
	require.Error(t, err)
	require.Len(t, addrs, 1, "addresses after too many errors are not returned")
	assert.Equal(t, "a@example.com", addrs[0].Address)
	assert.Contains(t, err.Error(), "stopped parsing email address list after 100 errors")

	// The error text must not grow with the list: a bounded number of
	// errors, each holding at most 100 bytes of the list. The messages
	// are measured without their call stacks, because err.Error() holds
	// the file paths of the call stacks, which are longer the deeper the
	// module is checked out, so that a bound on err.Error() failed
	// in a long checkout path.
	joined, ok := err.(interface{ Unwrap() []error })
	require.True(t, ok, "errors.Join error")
	errList := joined.Unwrap()
	assert.Len(t, errList, maxAddressListErrors+1, "the errors and the one saying that parsing stopped")
	for _, e := range errList {
		msg := errs.UnwrapCallStack(e).Error()
		assert.LessOrEqual(t, len(msg), 200, "error message: %s", msg)
	}
}

// TestParseAddressList_LongInvalidListTakesLinearTime protects the parsing
// of inbound mail, where ParseMIMEMessage parses every address header:
// after every entry that it couldn't parse, the address pattern scanned the
// rest of the list for an angle bracket or @, so lists like these of 1 MB
// took 6 seconds each up to the error limit, and finding the end of each
// entry scanned the rest of the list too. Every entry is parsed within its
// end now, an entry without an @ not at all, and the parts that are entries
// of their own are scanned once. Parsing these lists of 3 MB takes about
// five times a single scan of the list with openAtEnd now, and more than
// 40 times without any of these three. The time is compared to the scan,
// so that the bound holds with the race detector and on slow machines,
// and the fastest of three runs is taken against noise.
// The empty entries before an address with angle brackets were skipped one
// at a time, each time joining the others with that address again, so the
// rest of the list was parsed once per empty entry: 1 MB of commas before
// `<a@b.com>` took 6 seconds. A run of empty entries is skipped at once now.
func TestParseAddressList_LongInvalidListTakesLinearTime(t *testing.T) {
	fastest := func(f func()) time.Duration {
		var best time.Duration
		for i := range 3 {
			start := time.Now()
			f()
			if d := time.Since(start); i == 0 || d < best {
				best = d
			}
		}
		return best
	}
	for name, list := range map[string]string{
		"entries without @":            strings.Repeat("x, ", 1_000_000),
		"commas between two addresses": "a@b.com" + strings.Repeat(",", 3_000_000) + "c@d.com",
		"one entry without @":          strings.Repeat("x", 3_000_000),
		"commas before angle brackets": strings.Repeat(",", 3_000_000) + "<a@b.com>",
		"commas before a display name": "a@b.com" + strings.Repeat(",", 3_000_000) + "x <c@d.com>",
		"empty entries with spaces":    strings.Repeat(", ", 1_500_000) + "x <c@d.com>",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := ParseAddressList(list)
			require.Error(t, err)
			scan := fastest(func() { openAtEnd(list) })
			parse := fastest(func() { _, _ = ParseAddressList(list) })
			assert.Less(t, parse, 30*scan, "ParseAddressList takes %v, a scan of the list %v", parse, scan)
		})
	}
}

// BenchmarkParseAddressList_Long measures ParseAddressList on lists of
// 1 MB, invalid ones that the address pattern scanned to their end after
// every entry, see TestParseAddressList_LongInvalidListTakesLinearTime,
// and a valid one and one with an unquoted display name of 1 MB, whose
// entries are scanned once by the pattern and once to find their end.
func BenchmarkParseAddressList_Long(b *testing.B) {
	for _, bc := range []struct{ name, list string }{
		{name: "entries without @", list: strings.Repeat("x, ", 333_000)},
		{name: "commas between two addresses", list: "a@b.com" + strings.Repeat(",", 1_000_000) + "c@d.com"},
		{name: "no comma", list: strings.Repeat("x", 1_000_000)},
		{name: "valid", list: strings.Repeat("a@b.com, ", 110_000)},
		{name: "name with commas", list: strings.Repeat("Doe, ", 200_000) + "John <a@b.cc>"},
	} {
		b.Run(bc.name, func(b *testing.B) {
			for b.Loop() {
				_, _ = ParseAddressList(bc.list)
			}
		})
	}
}

// TestParseAddress_LongChainOfRepeatedAddresses protects the process: the
// parser read an address repeated in angle brackets after an address with
// a recursion per address, so millions of them like `<a@x.com> <a@x.com> ...`
// exhausted the goroutine stack, a fatal error that no recover catches.
// A loop reads them now. The stack limit is lowered, so that the recursion
// failed with these 20,000 addresses already.
func TestParseAddress_LongChainOfRepeatedAddresses(t *testing.T) {
	defer debug.SetMaxStack(debug.SetMaxStack(1 << 20))
	parsed, err := ParseAddress(strings.Repeat("<a@x.com> ", 20_000))
	require.NoError(t, err)
	assert.Equal(t, &mail.Address{Address: "a@x.com"}, parsed)
}

// TestParseAddress_NoSecondUnbracketedAddress protects the sender of a
// message: "attacker@evil.com ceo@corp.com" was read as ceo@corp.com,
// which net/mail rejects, while a check of the first address saw
// evil.com. Only an address in angle brackets is the address of a name,
// and a name with an address must be quoted, see
// TestParseAddress_AtInUnquotedName.
func TestParseAddress_NoSecondUnbracketedAddress(t *testing.T) {
	_, err := ParseAddress("attacker@evil.com ceo@corp.com")
	assert.Error(t, err)
	_, err = ParseAddress("attacker@evil.com <ceo@corp.com>")
	assert.Error(t, err)

	parsed, err := ParseAddress(`"attacker@evil.com" <ceo@corp.com>`)
	require.NoError(t, err)
	assert.Equal(t, "ceo@corp.com", parsed.Address)
	parsed, err = ParseAddress(`"\"Example\" <ar1@example.com>" <ar@example.com>`)
	require.NoError(t, err)
	assert.Equal(t, "ar@example.com", parsed.Address)
}

// FuzzSanitizeRunesOffsets checks what hasHiddenAddrSpecRune relies on to
// find the text that the parser read: for a trimmed text sanitizeRunes
// returns sanitizeAddr of it, with the offset in the text of every byte,
// so that a check of the text of an address part checks the runes that
// the parser read there.
func FuzzSanitizeRunesOffsets(f *testing.F) {
	for _, text := range []string{
		"a@example.com",
		"\"x\\\u200d\" <a@example.com>",
		"Jo\"hn <j@x.com>\\\ufeff",
		"M\u00fcller,\u00a0Hans <hans@example.com>",
		"x <\u00a0a@b.com>\x00\xff",
	} {
		f.Add(text)
	}
	f.Fuzz(func(t *testing.T, text string) {
		text = strutil.TrimSpace(text)
		var offsets []int
		sanitized := sanitizeRunes(text, &offsets)
		if want := sanitizeAddr(text); sanitized != want {
			t.Fatalf("sanitizeRunes(%q) = %q, but sanitizeAddr = %q", text, sanitized, want)
		}
		if len(offsets) != len(sanitized)+1 || offsets[len(sanitized)] != len(text) {
			t.Fatalf("sanitizeRunes(%q) returned %d offsets for %d bytes", text, len(offsets), len(sanitized))
		}
		for i := range len(sanitized) {
			if c := sanitized[i]; c != ' ' && c != text[offsets[i]] {
				t.Fatalf("sanitizeRunes(%q) byte %d %q comes from %q at %d", text, i, c, text[offsets[i]], offsets[i])
			}
			if r, _ := utf8.DecodeRuneInString(text[offsets[i]:]); sanitized[i] == ' ' && sanitizeRune(r) != ' ' {
				t.Fatalf("sanitizeRunes(%q) space %d comes from %q at %d", text, i, r, offsets[i])
			}
			if i > 0 && offsets[i] <= offsets[i-1] {
				t.Fatalf("sanitizeRunes(%q) offsets not increasing: %v", text, offsets)
			}
		}
	})
}

// FuzzParseAddressListMatchesNetMail checks that a list that the collections
// validate and store has the recipients that net/mail reads in the stored
// text, like a mail server that a message built from it is sent to: for
// every list that AddressList.Value writes and net/mail parses, the
// address parts that ParseAddressList and net/mail.ParseAddressList read,
// lower cased, are the same set. Where they differed, a recipient check of
// a validated list saw other addresses than the message was sent to, like
// for an encoded-word name with a comma and @, or a soft hyphen in an
// address. The single address of an AddressSet is checked the same way.
func FuzzParseAddressListMatchesNetMail(f *testing.F) {
	for _, list := range []string{
		`a@example.com, "Doe, John" <j@example.com>`,
		`Doe, John <john@example.com>, b@example.com (Comment, x)`,
		`"x\" <ok@example.com>, ok2@example.com (," <other@example.net>, "z)" <ok3@example.com>`,
		`=?utf-8?q?J=C3=BCrgen_M=C3=BCller?= <j@x.de>, "a b"@example.com`,
		`=?utf-8?q?spy@evil.example,?= <ceo@corp.example>`,
		`=?utf-8?q?<a@b.com>,?= <c@x.com>`,
		"ad\u00admin@corp.com",
		"Boss <ceo\u00ad@corp.com>",
		"x <\u00a0a@b.com>",
		"a@b.com,\u00a0c@d.com",
		"\"a\tb\"@x.com",
		"e@f.net..Doe",
		"undisclosed-recipients:;",
		// Found by fuzzing: the address pattern drops the apostrophe,
		// net/mail reads the address '0@0.aa
		"'0@0.AA",
		"o'brien@example.com, 'Brien' <ob@example.com>",
		// Found by the differential test against Python's email package,
		// see TestParseAddress_EncodedWordThatPythonReadsAsAddress
		`=?utf-8?q?x=40evil.com?= <ceo@corp.com>`,
		`=?utf-8?q?admin?=@corp.com, "=?utf-8?q?admin?="@corp.com`,
		"\"\u212Aate\"@corp.com, ceo@\u212Aorp.com <ceo@korp.com>",
		// Found by fuzzing: the Kelvin sign that the parser keeps, compared
		// with the address of net/mail lower cased by strings.ToLower
		"\"\u212A\"@0.AA",
		"<a@b.com> <a@b.com> ()\u00a0, c@d.com",
	} {
		f.Add(list)
	}
	f.Fuzz(func(t *testing.T, list string) {
		if v, err := AddressList(list).Value(); err == nil && v != nil {
			stdlib, err := mail.ParseAddressList(v.(string))
			if err == nil {
				parsed, err := ParseAddressList(list)
				if err != nil {
					t.Fatalf("AddressList(%q) is valid, but ParseAddressList fails: %v", list, err)
				}
				if got, want := addressPartSet(parsed), addressPartSet(stdlib); !slices.Equal(got, want) {
					t.Fatalf("AddressList(%q) is valid with the addresses %q, but net/mail reads %q", list, got, want)
				}
			}
		}
		if _, err := MakeAddressSet(Address(list)).Value(); err == nil {
			stdlib, err := mail.ParseAddress(strutil.TrimSpace(list))
			if err == nil {
				parsed, err := ParseAddress(list)
				if err != nil {
					t.Fatalf("AddressSet{%q} is valid, but ParseAddress fails: %v", list, err)
				}
				// Lower cased like the parser lower cases, so that a rune
				// that strings.ToLower turns into ASCII is the same rune
				if got, want := parsed.Address, lowerAddressPart(stdlib.Address); got != want {
					t.Fatalf("AddressSet{%q} is valid with the address %q, but net/mail reads %q", list, got, want)
				}
			}
		}
	})
}

// addressPartSet returns the sorted unique address parts of addrs lower cased.
func addressPartSet(addrs []*mail.Address) []string {
	return slices.Compact(slices.Sorted(slices.Values(addressParts(addrs))))
}

// FuzzValidListHasTheAddressesOfItsText checks the recipients of every
// list that the collections validate, also where net/mail rejects the list
// and FuzzParseAddressListMatchesNetMail checks nothing: every address
// written in the stored text outside quoted strings and comments is a
// parsed address, and every parsed address is written there. Other parsers
// read such an address as recipient, while one within a quoted string or
// comment is part of a name for them. Lists that net/mail rejects validated
// with other recipients than Python's email.utils.getaddresses reads, like
// `x@evil.com <ceo@corp.com>`, read as ceo@corp.com, and
// `"x@evil.com" "<ceo@corp.com>`, whose address is within the unclosed
// quote. The single address of an AddressSet is checked the same way.
func FuzzValidListHasTheAddressesOfItsText(f *testing.F) {
	for _, list := range []string{
		`a@example.com, "Doe, John" <j@example.com>`,
		`Doe, John <john@example.com>, b@example.com (Comment, x@y.com)`,
		`"a b"@example.com, "a\"b"@Example.COM`,
		`=?utf-8?q?spy@evil.example,?= <ceo@corp.example>`,
		`a@example.com <a@example.com>, <b@example.com> <b@example.com>`,
		// The findings that validated with other recipients
		`"x@evil.com" "<a@corp.com>`,
		`b@corp.com, "x@evil.com" "<a@corp.com>`,
		`x@evil.com <a@corp.com>,`,
		`Bob <x@evil.com> <a@corp.com>,`,
		`<x@evil.com> <a@corp.com>`,
		`x@evil.com (y) <a@corp.com>`,
		`=?utf-8?q?spy@evil.example?= <ceo@corp.example>`,
		`Support, "x\" <ceo@corp.com>, y (," <attacker@evil.com>, z@z.com`,
		`b@example.com, Doe (Sales <doe@example.com>`,
		`"a,b"@c.com <x@y.com>`,
		"a@b.com\t(Mike)",
		"a@b.com\t, c@d.com",
		// Names with a quoted @ that Python's default header parser ends
		// at a colon or reads with a backslash, see
		// TestParseAddress_QuotedAtInNameThatLenientParsersEnd
		`Team:"x@evil.com":<ceo@corp.com>`,
		`\"x@evil.com" <ceo@corp.com>`,
		// Encoded-words that Python decodes to another recipient, see
		// TestParseAddress_EncodedWordThatPythonReadsAsAddress
		`"=?utf-8?q?x=40evil.com?="; <ceo@corp.com>`,
		`=?utf-8?q?x=40evil.com?=; <ceo@corp.com>`,
		`=?utf-8?b?eEBldmlsLmNvbQ==?=) <ceo@corp.com>`,
		`Team:=?utf-8?q?x=40evil.com?=:<ceo@corp.com>`,
		`=?utf-8?q?x=40evil.com?=, <ceo@corp.com>`,
		`=?x-unknown?q?x=40evil.com?= <ceo@corp.com>`,
		`"=\?utf-8?q?x=40evil.com?=", <ceo@corp.com>`,
		`=?utf-8?q?admin?=@corp.com`,
		`=?utf-8?q?x@evil.com,?= <b@c.com>`,
		// Encoded-words that Python decodes to another recipient after
		// writing the list, see
		// TestParseAddress_EncodedWordThatPythonDecodesAfterWriting
		`"=\?utf-8?q?x" <y?=@corp.com>, "<evil@evil.com>" <c@corp.com>`,
		`"=\? x" <a@corp.com>, <z?q?z?=@corp.com>, "<evil@evil.com>" <c@corp.com>`,
		`=?utf-8?b?PT8=?=utf-8?q?x <y?=@corp.com>, "<evil@evil.com>" <c@corp.com>`,
		`"` + strings.Repeat("ä", 40) + ` =\?utf-8?q?x=40evil=2Ecom=2C?=" <ceo@corp.com>`,
		`=?utf-8?q?=3D?= =?utf-8?q?=3Futf-8=3Fq=3Fx=3D40evil.com=3F=3D?= <ceo@corp.com>`,
		strings.Repeat("a", 64) + `..=?utf-8?q?=2C_x=40evil.com=2C?=@corp.com`,
		`x.=?utf-8?q?=2C_x=40evil.com=2C?=.@corp.com`,
		// Ordinary names, see TestParseAddress_OrdinaryEncodedNames
		`=?utf-8?b?SsO8cmdlbiBNw7xsbGVyLCBCdWNoaGFsdHVuZyB1bmQgRmluYW56ZW4gKEFi?= =?utf-8?b?dGVpbHVuZyBOb3JkKQ==?= <j.mueller@example.com>`,
		`=?utf-8?q?J=C3=BCrgen_M=C3=BCller_[Extern]?= <j@example.de>`,
		`"Really=?" <a@example.com>`,
		`Müller, =?utf-8?q?J=C3=BCrgen?= <j@x.de>`,
		`Doe; =?utf-8?q?J=C3=BCrgen?= <ceo@corp.com>`,
		`[Ext] =?utf-8?q?J=C3=BCrgen?= <j@x.de>`,
	} {
		f.Add(list)
	}
	f.Fuzz(func(t *testing.T, list string) {
		if v, err := AddressList(list).Value(); err == nil && v != nil {
			parsed, err := ParseAddressList(list)
			if err != nil {
				t.Fatalf("AddressList(%q) is valid, but ParseAddressList fails: %v", list, err)
			}
			checkAddressesOfText(t, v.(string), parsed)
		}
		if _, err := MakeAddressSet(Address(list)).Value(); err == nil {
			parsed, err := ParseAddress(list)
			if err != nil {
				t.Fatalf("AddressSet{%q} is valid, but ParseAddress fails: %v", list, err)
			}
			checkAddressesOfText(t, strutil.TrimSpace(list), []*mail.Address{parsed})
		}
	})
}

// checkAddressesOfText fails t if the addresses written in the stored text
// outside quoted strings and comments, see textAddresses, are not the
// address parts of parsed, ignoring case, which the parser lower cases.
func checkAddressesOfText(t *testing.T, stored string, parsed []*mail.Address) {
	t.Helper()
	got := addressPartSet(parsed)
	written := textAddresses(stored)
	for _, addr := range written {
		if _, found := slices.BinarySearch(got, addr); !found {
			t.Fatalf("%q is valid with the addresses %q, but has the address %q outside quotes and comments", stored, got, addr)
		}
	}
	for _, addr := range got {
		if !slices.Contains(written, addr) {
			t.Fatalf("%q is valid with the address %q, which it has not outside quotes and comments: %q", stored, addr, written)
		}
	}
}

// textAddresses returns the lower cased words of text with an @ outside
// its quoted strings and comments, read like nextListSeparator reads them:
// the runs of characters that are no ASCII space and none of the
// delimiters <>,;:()[]" or the backslash, with the content of a quoted
// string right before an @ as unescaped local part, like for "a b"@x.com.
// It is no address parser, so that it finds every address that another
// parser could read, also where the address pattern of the package
// reads none.
func textAddresses(text string) []string {
	var (
		addrs []string
		state listState
		word  []byte
		local string // the quoted local part of word
		// quoted is the unescaped content of the last quoted string,
		// which ended before quoteEnd
		quoted   []byte
		quoteEnd = -1
	)
	endWord := func() {
		if slices.Contains(word, '@') {
			addrs = append(addrs, strings.ToLower(local+string(word)))
		}
		word, local = word[:0], ""
	}
	for i := 0; i < len(text); i++ {
		c := text[i]
		inQuote, inComment, escaped := state.quoted, state.comment > 0, state.escaped
		state.read(c)
		switch {
		case inQuote:
			if escaped || (c != '\\' && c != '"') {
				quoted = append(quoted, c)
			} else if c == '"' {
				quoteEnd = i + 1
			}
		case inComment:
		case c == '"':
			endWord()
			quoted = quoted[:0]
		case strings.IndexByte(" \t<>,;:()[]\\", c) >= 0:
			endWord()
		default:
			if len(word) == 0 && c == '@' && quoteEnd == i {
				local = string(quoted)
			}
			word = append(word, c)
		}
	}
	endWord()
	slices.Sort(addrs)
	return slices.Compact(addrs)
}
