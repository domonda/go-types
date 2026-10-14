package email

import (
	"fmt"
	"net/mail"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/domonda/go-errs"

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
		`@Erik <erik@domonda.com>`:                                      {Name: "@Erik", Address: "erik@domonda.com"},
		`erik.unger@domonda.com <erik@domonda.com>`:                     {Name: "erik.unger@domonda.com", Address: "erik@domonda.com"}, // Use "erik.unger" in name part vs "erik" in address part to test picking up the right part
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
		// The text of an encoded-word keeps its quotes and backslashes
		{addr: `=?utf-8?q?Say_=22Hi=22?= <a@b.com>`, wantName: `Say "Hi"`, wantAddr: "a@b.com", wantNorm: `"Say \"Hi\"" <a@b.com>`},
		{addr: `=?utf-8?b?YVxi?= <a@b.com>`, wantName: `a\b`, wantAddr: "a@b.com", wantNorm: `"a\\b" <a@b.com>`},
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
	for addr, want := range map[string]string{
		`Jo"hn <a@example.com>`:             "John",
		`John "Johnny" Doe <a@example.com>`: "John Johnny Doe",
		`"Doe <a@example.com>`:              "Doe",
	} {
		parsed, err := ParseAddress(addr)
		require.NoError(t, err, addr)
		assert.Equal(t, want, parsed.Name, addr)
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
		// An address with an unclosed comment in its name is valid
		// if no comma follows, like for ParseAddress
		{list: `b@example.com, Doe (Sales <doe@example.com>`, want: []string{"b@example.com", "doe@example.com"}},
		// The comment holds the rest of the list, so b@example.com is no
		// address: it was returned after the comment ended at "(y)"
		{list: `a@example.com (x, (y), b@example.com`, want: []string{"a@example.com"}, wantErr: true},
		{list: `a@example.com, b@example.com (x, c@example.com`, want: []string{"a@example.com", "b@example.com"}, wantErr: true},
		// The comma after doe@example.com is in the comment of its name
		{list: `Doe (Sales <doe@example.com>, b@example.com`, want: []string{"doe@example.com"}, wantErr: true},
		{list: `Doe (Sales <doe@example.com>,`, want: []string{"doe@example.com"}, wantErr: true},
		// A comment that is closed later doesn't hold the rest of the list,
		// parsing continues after the next comma like after other errors
		{list: `Doe (Sales <doe@example.com>, b), c@example.com`, want: []string{"doe@example.com", "c@example.com"}, wantErr: true},
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
		{list: `Doe (Sales \) <doe@example.com>, b@example.com`, want: []string{"doe@example.com"}, wantErr: true},
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
// name "," of the next address.
func TestParseAddressList_UnclosedQuoteOrAngle(t *testing.T) {
	tests := []struct {
		list    string
		want    []string // the addresses parts of the parsed addresses
		wantErr bool
	}{
		{list: `"Doe <a@b.cc>, z@y.zz`, want: []string{"a@b.cc"}, wantErr: true},
		{list: `x <a@b.cc, z@y.zz`, want: []string{"a@b.cc"}, wantErr: true},
		{list: `Jo"hn <a@example.com>, b@example.com`, want: []string{"a@example.com"}, wantErr: true},
		// The comma trimmed from the end of the list is in the bracket
		{list: `<a@b.cc,`, want: []string{"a@b.cc"}, wantErr: true},
		// Without a comma after them they hold no other entry,
		// so the address is valid like for ParseAddress
		{list: `"Doe <a@b.cc>`, want: []string{"a@b.cc"}},
		{list: `x <a@b.cc`, want: []string{"a@b.cc"}},
		{list: `a@b.cc, Jo"hn <c@d.ee>`, want: []string{"a@b.cc", "c@d.ee"}},
		// An empty entry is no display name
		{list: `,<a@example.com>`, want: []string{"a@example.com"}, wantErr: true},
		{list: `a@example.com,, <b@example.com>`, want: []string{"a@example.com", "b@example.com"}, wantErr: true},
		// After an earlier error, parsing recovers at every comma that
		// separates entries, up to the entry with the unclosed quote or
		// angle bracket, which holds the rest of the list
		{list: `bad, a@b.cc, x <c@d.ee, e@f.gg`, want: []string{"a@b.cc", "c@d.ee"}, wantErr: true},
		{list: `bad, a@b.cc, "e <e@f.com>, g@h.com`, want: []string{"a@b.cc", "e@f.com"}, wantErr: true},
		{list: `bad (x, y), a@b.cc, Jo"hn <c@d.ee>, e@f.gg`, want: []string{"a@b.cc", "c@d.ee"}, wantErr: true},
		// A comma that separates entries for SplitAddressList within what
		// the address pattern reads as one address, here between the
		// quotes of the second address in angle brackets after the first,
		// see the special case in parseAddress. Neither entry is valid.
		{list: `x@y.com "<"a,b"@x.com>`, want: nil, wantErr: true},
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

// TestParseAddress_NoSecondUnbracketedAddress protects the sender of a
// message: "attacker@evil.com ceo@corp.com" was read as ceo@corp.com,
// which net/mail rejects, while a check of the first address saw
// evil.com. Only an address in angle brackets is the address of a name.
func TestParseAddress_NoSecondUnbracketedAddress(t *testing.T) {
	_, err := ParseAddress("attacker@evil.com ceo@corp.com")
	assert.Error(t, err)

	parsed, err := ParseAddress("attacker@evil.com <ceo@corp.com>")
	require.NoError(t, err)
	assert.Equal(t, "ceo@corp.com", parsed.Address)
	parsed, err = ParseAddress(`"\"Example\" <ar1@example.com>" <ar@example.com>`)
	require.NoError(t, err)
	assert.Equal(t, "ar@example.com", parsed.Address)
}
