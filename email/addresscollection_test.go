package email

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"unicode"

	"github.com/domonda/go-errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/domonda/go-types/notnull"
)

// SplitAddressList splits only where a comma separates two entries, so
// that the collection types can read an address list as written without
// parsing and normalizing it, which could merge two entries of a set.
func TestSplitAddressList(t *testing.T) {
	tests := []struct {
		list string
		want []Address
	}{
		{list: "", want: nil},
		{list: " \t ", want: nil},
		{list: "a@example.com", want: []Address{"a@example.com"}},
		{list: " A@Example.com ,b@example.com, ", want: []Address{"A@Example.com", "b@example.com"}},
		// Entries are not parsed, so invalid ones are kept as given
		{list: "not an address, b@example.com", want: []Address{"not an address", "b@example.com"}},
		// Commas in quoted names, angle brackets and comments don't split
		{list: `"Doe, John" <john@example.com>, b@example.com`, want: []Address{`"Doe, John" <john@example.com>`, "b@example.com"}},
		{list: `"a \"quoted, name\"" <a@example.com>, b@example.com`, want: []Address{`"a \"quoted, name\"" <a@example.com>`, "b@example.com"}},
		{list: `<"a,b"@example.com>, c@example.com`, want: []Address{`<"a,b"@example.com>`, "c@example.com"}},
		{list: `a@example.com (Doe, John), b@example.com`, want: []Address{`a@example.com (Doe, John)`, "b@example.com"}},
		{list: `a@example.com (outer (inner, x)), b@example.com`, want: []Address{`a@example.com (outer (inner, x))`, "b@example.com"}},
		// The commas of an unquoted display name don't split,
		// like ParseAddressList reads it as one address
		{list: `Doe, John <john@example.com>`, want: []Address{`Doe, John <john@example.com>`}},
		{list: `Doe,Jr. ,  John <john@example.com>, b@example.com`, want: []Address{`Doe,Jr. ,  John <john@example.com>`, "b@example.com"}},
		{list: `a@example.com, Doe, John <john@example.com>`, want: []Address{"a@example.com", `Doe, John <john@example.com>`}},
		// Without angle brackets after them the parts are no display name
		{list: `foo, bar@example.com`, want: []Address{"foo", "bar@example.com"}},
		{list: `foo, bar`, want: []Address{"foo", "bar"}},
		{list: `foo, bar@example.com, Doe, John <john@example.com>, baz`, want: []Address{"foo", "bar@example.com", `Doe, John <john@example.com>`, "baz"}},
		// An escaped quote doesn't end the quoted name, also when an odd
		// number of them leaves no closing quote for the display name join
		{list: `"a \" b, c" <x@example.com>, y@example.com`, want: []Address{`"a \" b, c" <x@example.com>`, "y@example.com"}},
		// A stray '>' doesn't keep the later commas from splitting
		{list: `a@example.com>, b@example.com`, want: []Address{`a@example.com>`, "b@example.com"}},
		// A list of only separators has no entries
		{list: ",", want: nil},
		{list: " , ,", want: nil},
		{list: ",,,", want: nil},
		// Angle brackets without an @ still end the display name run
		{list: `Doe, John <invalid>, b@example.com`, want: []Address{`Doe, John <invalid>`, "b@example.com"}},
		// After an unbalanced quote the rest is one entry, no error
		{list: `"unbalanced, a@example.com, b@example.com`, want: []Address{`"unbalanced, a@example.com, b@example.com`}},
		// Placeholders for hidden recipients are no entries
		{list: "undisclosed-recipients:;", want: nil},
		{list: "Undisclosed recipients", want: nil},
		// Non-ASCII names are kept
		{list: "Jürgen Müller <j@example.com>, Zoë <z@example.com>", want: []Address{"Jürgen Müller <j@example.com>", "Zoë <z@example.com>"}},
		// The commas of a display name of only encoded-words don't split,
		// like ParseAddressList reads it as one name
		{list: `=?utf-8?q?a@b,?= <a@example.com>, b@example.com`, want: []Address{`=?utf-8?q?a@b,?= <a@example.com>`, "b@example.com"}},
		// A backslash escapes the next rune that ParseAddressList reads,
		// not the zero width joiner that it drops, so the quote closes
		{list: "\"x\\\u200d\" <a@example.com>, b@example.com", want: []Address{"\"x\\\u200d\" <a@example.com>", "b@example.com"}},
		// A placeholder followed by commas, like ParseAddressList reads it
		{list: "undisclosed-recipients:;, ", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.list, func(t *testing.T) {
			assert.Equal(t, tt.want, SplitAddressList(tt.list))
			assert.Equal(t, tt.want, AddressList(tt.list).SplitRaw())
			assert.Equal(t, tt.want, NullableAddressList(tt.list).SplitRaw())
		})
	}
}

// For a valid list SplitRaw finds the same entries as SplitNormalized,
// so a list read as raw entries normalizes to the same addresses.
func TestAddressList_SplitRawMatchesSplitNormalized(t *testing.T) {
	for _, list := range []AddressList{
		"a@example.com",
		"A@Example.com, b@example.com",
		`"Doe, John" <John@Example.com>, Jürgen <J@Example.com>`,
		`a@example.com (comment, with comma), "x" <b@example.com>`,
		// Nested comments and escaped parentheses, which ParseAddressList
		// read up to the first ')' while SplitAddressList nests them
		`a@example.com (outer (inner, x)), b@example.com`,
		`a@example.com (x \), y), <b@example.com> (B (C, D)), c@example.com`,
		`John (Sales, EU) <a@example.com>, b@example.com`,
		`b@example.com, Doe (Sales <doe@example.com>`,
		// Escaped quotes in quoted names, which ParseAddressList read as
		// the end of the name, see TestParseAddressList_EscapedQuoteInName
		`"x\" <ok@example.com>, ok2@example.com (," <other@example.net>, "z)" <ok3@example.com>`,
		`"Say \"Hi\"" <a@example.com>, "back\\slash" <b@example.com>`,
		// A backslash escapes a rune that the parser drops, here a zero
		// width joiner, so it doesn't escape the quote after it
		"\"x\\\u200d\" <a@example.com>, b@example.com",
		// The commas in an encoded-word name don't separate entries,
		// also when the name is one only without the runes that the
		// parser drops, so SplitRaw reads the list like the parser
		`=?utf-8?q?a@b,?= <hidden@example.com>, =?utf-8?q?c?= <shown@example.com>`,
		"=\u200d?utf-8?q?a@b,?= <hidden@example.com>, shown@example.com",
		// Nor does a parenthesis in it open a comment
		`=?utf-8?q?(Sales?= <a@example.com>, b@example.com`,
		// The no-break space in the unquoted name is read as space
		"Müller,\u00a0Hans <hans@example.com>, b@example.com",
		// A placeholder with a trailing comma is the empty list
		"undisclosed-recipients:;, ",
	} {
		t.Run(string(list), func(t *testing.T) {
			require.True(t, list.Valid())
			normalized, err := list.SplitNormalized()
			require.NoError(t, err)
			raw := list.SplitRaw()
			require.Len(t, raw, len(normalized))
			for i, entry := range raw {
				norm, err := entry.Normalized()
				require.NoError(t, err)
				assert.Equal(t, normalized[i], norm, "entry %d", i)
			}
		})
	}

	// ParseAddressList parsed these lists without an error as more
	// addresses or other ones than SplitRaw finds entries: an unclosed
	// quote or angle bracket holds the rest of the list for SplitRaw,
	// and a leading or doubled comma became the display name ",".
	// They are invalid now, so that no validated list reads back as
	// other entries.
	for _, list := range []AddressList{
		`"Doe <a@b.cc>, z@y.zz`,
		`x <a@b.cc, z@y.zz`,
		`Jo"hn <a@example.com>, b@example.com`,
		`<a@b.cc,`,
		`,<a@example.com>`,
		`a@example.com,, <b@example.com>`,
	} {
		t.Run(string(list), func(t *testing.T) {
			assert.False(t, list.Valid())
			_, err := list.SplitNormalized()
			assert.Error(t, err)
		})
	}
}

// A comment that isn't closed holds the rest of the list for SplitRaw.
// SplitNormalized must report an error for such a list instead of
// returning the addresses in the comment as if they were entries,
// which it did after reading the comment only up to its first ')',
// so that the two don't disagree about a list without an error.
func TestAddressList_UnclosedCommentHoldsRestOfList(t *testing.T) {
	tests := []struct {
		list     AddressList
		wantRaw  []Address
		wantNorm []Address
	}{
		{
			list:     `a@example.com (x, (y), b@example.com`,
			wantRaw:  []Address{`a@example.com (x, (y), b@example.com`},
			wantNorm: []Address{`a@example.com`},
		},
		{
			list:     `a@example.com, b@example.com (x, c@example.com`,
			wantRaw:  []Address{`a@example.com`, `b@example.com (x, c@example.com`},
			wantNorm: []Address{`a@example.com`, `b@example.com`},
		},
		{
			list:     `Doe (Sales <doe@example.com>, b@example.com`,
			wantRaw:  []Address{`Doe (Sales <doe@example.com>, b@example.com`},
			wantNorm: []Address{`"Doe (Sales" <doe@example.com>`},
		},
	}
	for _, tt := range tests {
		t.Run(string(tt.list), func(t *testing.T) {
			assert.Equal(t, tt.wantRaw, tt.list.SplitRaw())
			norm, err := tt.list.SplitNormalized()
			assert.Error(t, err)
			assert.Equal(t, tt.wantNorm, norm)
			assert.False(t, tt.list.Valid())
		})
	}
}

// FuzzSplitRawMatchesParse checks that SplitAddressList finds the entries
// that ParseAddressList parses, for every list that ParseAddressList
// parses without an error: as many entries as addresses, and every entry
// normalizes to its address. AddressSet.Scan and the SplitRaw methods read
// a stored list with SplitAddressList, so where the two disagreed, a list
// that validated was read back as other recipients.
//
// For every list, also an invalid one, Validate and Value of the
// collections agree, so that no validated list fails to be stored and no
// list that Validate refuses is stored.
func FuzzSplitRawMatchesParse(f *testing.F) {
	for _, list := range []string{
		`a@example.com, b@example.com`,
		`"Doe, John" <john@example.com>, Doe, Jr., John <j@example.com>`,
		`a@example.com (outer (inner, x)), <b@example.com> (B \) C)`,
		`"x\" <ok@example.com>, ok2@example.com (," <other@example.net>, "z)" <ok3@example.com>`,
		`"Doe <a@b.cc>, z@y.zz`,
		`x <a@b.cc, z@y.zz`,
		`Jo"hn <a@example.com>, b@example.com`,
		`,<a@example.com>`,
		`=?utf-8?q?J=C3=BCrgen?= <j@example.com>, "a b"@example.com`,
		"undisclosed-recipients:;, ",
		// Found by fuzzing: a backslash escaped invalid UTF-8, which the
		// parser drops, so that it escaped the quote after it instead
		"\"\\\xa0\"\"@0.AA,",
		// Found by fuzzing: a comma in an encoded-word name
		`=?@?B?,?=<"0"@0.AA`,
		// Found by fuzzing: a backslash before invalid UTF-8 at the end,
		// which strutil.TrimSpace removes, see
		// TestCollectionsValidateAgreesWithValueAtTrailingBackslash
		"\"<0@0.AA\\\x9f",
		"Jo\"hn <j@x.com>\\\ufeff",
	} {
		f.Add(list)
	}
	f.Fuzz(func(t *testing.T, list string) {
		validateErr := AddressList(list).Validate()
		if _, valueErr := AddressList(list).Value(); (validateErr == nil) != (valueErr == nil) {
			t.Fatalf("AddressList(%q).Validate() = %v, but Value() = %v", list, validateErr, valueErr)
		}
		set := MakeAddressSet(Address(list))
		setValidateErr := set.Validate()
		if _, valueErr := set.Value(); (setValidateErr == nil) != (valueErr == nil) {
			t.Fatalf("AddressSet{%q}.Validate() = %v, but Value() = %v", list, setValidateErr, valueErr)
		}
		parsed, err := ParseAddressList(list)
		if err != nil {
			return
		}
		raw := SplitAddressList(list)
		if len(raw) != len(parsed) {
			t.Fatalf("SplitAddressList(%q) = %q, but ParseAddressList parsed %d addresses: %q", list, raw, len(parsed), parsed)
		}
		for i, entry := range raw {
			norm, err := entry.Normalized()
			if err != nil {
				t.Fatalf("entry %q of %q: %v", entry, list, err)
			}
			if want := AddressFrom(parsed[i]); norm != want {
				t.Fatalf("entry %q of %q normalizes to %q, but ParseAddressList parsed %q", entry, list, norm, want)
			}
		}
	})
}

func TestReadAddressCollection(t *testing.T) {
	tests := []struct {
		text    string
		want    addressCollection
		wantErr bool
	}{
		{text: "", want: addressCollection{}},
		{text: "  ", want: addressCollection{}},
		{text: "null", want: addressCollection{}},
		{text: `""`, want: addressCollection{}},
		{text: `["b@example.com", " a@example.com", ""]`, want: addressCollection{entries: []string{"b@example.com", " a@example.com", ""}}},
		{text: `[]`, want: addressCollection{entries: []string{}}},
		{text: `{"b@example.com": {}, "a@example.com": true}`, want: addressCollection{entries: []string{"a@example.com", "b@example.com"}}},
		// The empty object and the empty PostgreSQL array are the same
		{text: `{}`, want: addressCollection{entries: []string{}}},
		{text: `{"a@example.com", b@example.com}`, want: addressCollection{entries: []string{"a@example.com", "b@example.com"}}},
		{text: `{a@example.com,""}`, want: addressCollection{entries: []string{"a@example.com"}}},
		{text: `"a@example.com, b@example.com"`, want: addressCollection{list: "a@example.com, b@example.com"}},
		{text: `"Doe, John" <john@example.com>`, want: addressCollection{list: `"Doe, John" <john@example.com>`}},
		{text: ` a@example.com `, want: addressCollection{list: ` a@example.com `}},
		{text: `[1]`, wantErr: true},
		{text: `{NULL}`, wantErr: true},
		// Trimmed like strutil.TrimSpace, not only of JSON whitespace
		{text: "\u200b[\"a@example.com\"]\x00", want: addressCollection{entries: []string{"a@example.com"}}},
		// JSON whitespace before the first key or the closing brace
		// doesn't make a JSON object a PostgreSQL array
		{text: "{ \n\t\"a@example.com\": 1}", want: addressCollection{entries: []string{"a@example.com"}}},
		{text: `{ }`, want: addressCollection{entries: []string{}}},
		// A PostgreSQL array with quoted elements, as AddressSet.Value
		// writes it, starts like a JSON object
		{text: `{"b@example.com","a@example.com"}`, want: addressCollection{entries: []string{"b@example.com", "a@example.com"}}},
		{text: `{"a@example.com"}`, want: addressCollection{entries: []string{"a@example.com"}}},
		// Text that only starts like JSON is no JSON and is read as address
		// list without an error, like a column value that was cut off,
		// so that one bad value doesn't fail a whole query; Validate
		// reports it. A type error in it doesn't make it JSON either.
		{text: `["a@example.com"`, want: addressCollection{list: `["a@example.com"`}},
		{text: `[1, }`, want: addressCollection{list: `[1, }`}},
		{text: `["a@example.com"] x`, want: addressCollection{list: `["a@example.com"] x`}},
		{text: `{"a@example.com": 1} x`, want: addressCollection{list: `{"a@example.com": 1} x`}},
		{text: `{`, want: addressCollection{list: `{`}},
		{text: `"a\x" <a@example.com>"`, want: addressCollection{list: `"a\x" <a@example.com>"`}},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got, err := readAddressCollection(tt.text, "AddressSet")
			// Scan passes a string or a []byte and UnmarshalJSON a []byte,
			// which must read the same without being copied into a string
			gotBytes, errBytes := readAddressCollection([]byte(tt.text), "AddressSet")
			if tt.wantErr {
				assert.Error(t, err)
				assert.Error(t, errBytes, "[]byte")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			require.NoError(t, errBytes, "[]byte")
			assert.Equal(t, tt.want, gotBytes, "[]byte")
		})
	}
}

// writeCollections returns the errors of the methods that write an address
// collection after validating it: the Value methods of AddressList,
// NullableAddressList and AddressSet, and AddressSet.MarshalText.
func writeCollections(addr Address) map[string]error {
	errOf := make(map[string]error)
	_, errOf["AddressList.Value"] = AddressList(addr).Value()
	_, errOf["NullableAddressList.Value"] = NullableAddressList(addr).Value()
	_, errOf["AddressSet.Value"] = MakeAddressSet(addr).Value()
	_, errOf["AddressSet.MarshalText"] = MakeAddressSet(addr).MarshalText()
	return errOf
}

// validateCollections returns the errors of the Validate methods
// of AddressList, NullableAddressList and AddressSet for addr.
func validateCollections(addr Address) map[string]error {
	return map[string]error{
		"AddressList.Validate":         AddressList(addr).Validate(),
		"NullableAddressList.Validate": NullableAddressList(addr).Validate(),
		"AddressSet.Validate":          MakeAddressSet(addr).Validate(),
	}
}

// The parser reads an address sanitized, without the line breaks, control
// characters and invalid UTF-8 that it ignores, while the collections
// write their text as stored. A display name with a line break passed
// Validate and was written with it, so that a message header built from
// the stored list got another header injected, and a NUL passed and failed
// the write to PostgreSQL. So the Validate methods of the collections
// refuse them, and with them the methods that write the text as stored.
// Address.Validate still accepts them, because Address.Value writes the
// normalized address, which doesn't have them.
func TestCollectionsRefuseRunesTheParserIgnores(t *testing.T) {
	for _, addr := range []Address{
		"\"x\r\nBcc: evil@evil.com\" <a@example.com>",
		"\"x\rBcc: evil@evil.com\" <a@example.com>",
		"\"x\nBcc: evil@evil.com\" <a@example.com>",
		"a\x00b@example.com",
		"\"x\x7f\" <a@example.com>",
		"\"x\u0085\" <a@example.com>",
		// Bidi controls that make the name read differently
		"\"\u202eexample.com\" <a@example.com>",
		"\"\u2067example.com\u2069\" <a@example.com>",
		"\"x\xff\" <a@example.com>",
	} {
		require.NoError(t, addr.Validate(), "the parser ignores the runes of %q", addr)
		for method, err := range validateCollections(addr) {
			assert.Error(t, err, "%s of %q", method, addr)
		}
		for method, err := range writeCollections(addr) {
			assert.Error(t, err, "%s of %q", method, addr)
		}
		// The normalized address is what the parser read,
		// so it can be stored instead
		norm, err := addr.Normalized()
		require.NoError(t, err)
		for method, err := range writeCollections(norm) {
			assert.NoError(t, err, "%s of %q", method, norm)
		}
	}

	// Non-ASCII names are UTF-8 that the parser reads as written,
	// and the surrounding whitespace is trimmed and not written
	for _, addr := range []Address{
		"Jürgen Müller <j@example.com>",
		`"Zoë 李" <z@example.com>`,
		"\r\n\ta@example.com\r\n",
	} {
		for method, err := range writeCollections(addr) {
			assert.NoError(t, err, "%s of %q", method, addr)
		}
	}
}

// The collections refuse only the runes that inject a header, break the
// storage or reorder the displayed text: the control characters other
// than the tab, and the bidi embeddings, overrides and isolates. Refusing
// every rune that the parser changes or drops refused legitimate names,
// like the no-break space or the zero width non-joiner of Persian names,
// which Validate accepted, so that a validated row could not be saved.
// Validate and the methods that write the text as stored must agree for
// every rune, else a validated list fails to write or an unvalidated
// one is written.
func TestCollectionsRefuseOnlyUnsafeRunes(t *testing.T) {
	var runes []rune
	for r := rune(0); r < 0x300; r++ {
		runes = append(runes, r)
	}
	runes = append(runes,
		'\u061c', '\u1680', '\u180e', '\u2000', '\u200b', '\u200c', '\u200d', '\u200e',
		'\u2028', '\u2029', '\u202a', '\u202e', '\u205f', '\u2060', '\u2066', '\u2069',
		'\u3000', '\ue000', '\ufeff', '\ufffd', '\U0001f600', '\U000e0001', '\U0010ffff',
	)
	for _, r := range runes {
		if r == '"' || r == '\\' {
			// Not part of a valid quoted name
			continue
		}
		addr := Address(`"a` + string(r) + `b" <a@example.com>`)
		unsafe := (unicode.Is(unicode.Cc, r) && r != '\t') ||
			(r >= '\u202a' && r <= '\u202e') || (r >= '\u2066' && r <= '\u2069')
		for method, err := range validateCollections(addr) {
			if unsafe {
				assert.Error(t, err, "%s of %q", method, addr)
			} else {
				assert.NoError(t, err, "%s of %q", method, addr)
			}
		}
		for method, err := range writeCollections(addr) {
			if unsafe {
				assert.Error(t, err, "%s of %q", method, addr)
			} else {
				assert.NoError(t, err, "%s of %q", method, addr)
			}
		}
	}
}

// Names in many languages need runes that the parser reads as spaces or
// drops: the no-break space, the ideographic space between Japanese
// given and family names, the zero width non-joiner of Persian, the zero
// width joiner of emoji sequences and the soft hyphen of long German
// words. A list with such a name passed Validate, but Value refused it,
// so that a validated row could not be saved. They are written as stored.
func TestCollectionsWriteNamesWithSpecialSpacesAndJoiners(t *testing.T) {
	for _, addr := range []Address{
		"Müller,\u00a0Hans <hans@example.com>",
		"\"Müller,\u00a0Hans\" <hans@example.com>",
		"\"山田\u3000太郎\" <taro@example.jp>",
		"\"علی\u200cرضا\" <alireza@example.ir>",
		"\"Familie 👨\u200d👩\u200d👧\" <familie@example.com>",
		"\"Donau\u00addampf\u00adschiff\" <kapitaen@example.at>",
		"\"\u200fمحمد\u200e\" <m@example.com>",
		"\"Doe,\tJohn\" <john@example.com>",
	} {
		t.Run(string(addr), func(t *testing.T) {
			for method, err := range validateCollections(addr) {
				assert.NoError(t, err, method)
			}
			for method, err := range writeCollections(addr) {
				assert.NoError(t, err, method)
			}
			v, err := AddressList(addr).Value()
			require.NoError(t, err)
			assert.Equal(t, string(addr), v, "written as stored")
			v, err = NullableAddressList(addr).Value()
			require.NoError(t, err)
			assert.Equal(t, string(addr), v, "written as stored")
			text, err := MakeAddressSet(addr).MarshalText()
			require.NoError(t, err)
			assert.Equal(t, string(addr), string(text), "written as stored")
		})
	}
}

// Validate refuses the unsafe runes with the same error that Value
// returns, so that a caller that validates first sees why the write
// would fail, and both name the same text.
func TestCollectionsValidateAndValueRefuseWithSameError(t *testing.T) {
	for _, addr := range []Address{
		"\"x\r\nBcc: evil@evil.com\" <a@example.com>",
		"a\x00b@example.com",
		"\"\u202eexample.com\" <a@example.com>",
		"\"x\xff\" <a@example.com>",
	} {
		_, valueErr := AddressList(addr).Value()
		require.Error(t, valueErr)
		assert.Equal(t, errs.UnwrapCallStack(AddressList(addr).Validate()).Error(), errs.UnwrapCallStack(valueErr).Error(), "AddressList %q", addr)

		_, valueErr = NullableAddressList(addr).Value()
		require.Error(t, valueErr)
		assert.Equal(t, errs.UnwrapCallStack(NullableAddressList(addr).Validate()).Error(), errs.UnwrapCallStack(valueErr).Error(), "NullableAddressList %q", addr)

		set := MakeAddressSet(addr)
		_, valueErr = set.Value()
		require.Error(t, valueErr)
		assert.Equal(t, errs.UnwrapCallStack(set.Validate()).Error(), errs.UnwrapCallStack(valueErr).Error(), "AddressSet %q", addr)
		_, textErr := set.MarshalText()
		require.Error(t, textErr)
		assert.Equal(t, errs.UnwrapCallStack(set.Validate()).Error(), errs.UnwrapCallStack(textErr).Error(), "AddressSet.MarshalText %q", addr)
	}

	// Surrounding runes that Value trims and doesn't write are not refused
	for _, list := range []AddressList{"a@example.com\r\n", "\x00a@example.com", "\u202ea@example.com"} {
		assert.NoError(t, list.Validate(), "%q", list)
		v, err := list.Value()
		require.NoError(t, err)
		assert.Equal(t, "a@example.com", v)
	}
}

// The parser read a list sanitized: a backslash before a rune that it
// drops was dropped with it, also before the runes at the end that
// strutil.TrimSpace removes, while Value trims the list first and keeps
// the backslash with nothing left to escape. So Validate accepted these
// lists and Value refused them, and FuzzSplitRawMatchesParse failed on the
// last one. The parser trims first now, so both refuse them.
func TestCollectionsValidateAgreesWithValueAtTrailingBackslash(t *testing.T) {
	for _, addr := range []Address{
		"Jo\"hn <j@x.com>\\\ufeff",
		"Jo\"hn <j@x.com>\\\u00ad",
		"Jo\"hn <j@x.com>\\\x00",
		"\"Doe <doe@example.com>\\\u200d",
		"\"<0@0.AA\\\x9f",
	} {
		t.Run(string(addr), func(t *testing.T) {
			for method, err := range validateCollections(addr) {
				assert.Error(t, err, method)
			}
			for method, err := range writeCollections(addr) {
				assert.Error(t, err, method)
			}
			_, err := ParseAddressList(string(addr))
			assert.Error(t, err, "ParseAddressList")
		})
	}
}

// BenchmarkReadAddressCollection measures reading 1000 addresses in every
// shape that Scan and UnmarshalJSON accept, next to decoding the JSON array
// into a []string, because a JSON document or a column with a long list
// should not cost several times more to read than the JSON decoding itself.
func BenchmarkReadAddressCollection(b *testing.B) {
	addrs := make([]string, 1000)
	object := make(map[string]struct{}, len(addrs))
	for i := range addrs {
		addrs[i] = fmt.Sprintf("user%d@example.com", i)
		object[addrs[i]] = struct{}{}
	}
	list := strings.Join(addrs, ", ")
	jsonArray, err := json.Marshal(addrs)
	require.NoError(b, err)
	jsonObject, err := json.Marshal(object)
	require.NoError(b, err)
	jsonString, err := json.Marshal(list)
	require.NoError(b, err)
	pgArray, err := notnull.StringArray(addrs).Value()
	require.NoError(b, err)

	b.Run("json.Unmarshal into []string", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var s []string
			if err := json.Unmarshal(jsonArray, &s); err != nil {
				b.Fatal(err)
			}
		}
	})
	for _, bc := range []struct {
		name string
		data []byte
	}{
		{name: "UnmarshalJSON array", data: jsonArray},
		{name: "UnmarshalJSON object", data: jsonObject},
		{name: "UnmarshalJSON string", data: jsonString},
	} {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				var l AddressList
				if err := l.UnmarshalJSON(bc.data); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
	for _, bc := range []struct {
		name  string
		value any
	}{
		{name: "Scan JSON array", value: string(jsonArray)},
		{name: "Scan JSON array bytes", value: jsonArray},
		{name: "Scan PostgreSQL array", value: pgArray},
		{name: "Scan list", value: list},
		{name: "Scan one address", value: "a@example.com"},
	} {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				var l AddressList
				if err := l.Scan(bc.value); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
