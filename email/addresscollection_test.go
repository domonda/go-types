package email

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

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

// Validate parses an address sanitized, without the line breaks, control
// characters and invalid UTF-8 that the parser ignores, while the
// collections write their text as stored. A display name with a line
// break passed Validate and was written with it, so that a message header
// built from the stored list got another header injected, and a NUL
// passed and failed the write to PostgreSQL. So writing refuses them.
func TestCollectionsRefuseRunesTheParserIgnores(t *testing.T) {
	for _, addr := range []Address{
		"\"x\r\nBcc: evil@evil.com\" <a@example.com>",
		"a\x00b@example.com",
		// A bidi override that makes the name read differently
		"\"\u202eexample.com\" <a@example.com>",
		"\"x\xff\" <a@example.com>",
	} {
		require.NoError(t, addr.Validate(), "the parser ignores the runes of %q", addr)
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

// The collections refuse exactly the runes that sanitizeAddr changes or
// drops, so that they write every text that Validate saw as written,
// and nothing that it didn't.
func TestCollectionsRefuseWhatSanitizeAddrChanges(t *testing.T) {
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
		changed := sanitizeAddr(string(addr)) != string(addr)
		for method, err := range writeCollections(addr) {
			if changed {
				assert.Error(t, err, "%s of %q", method, addr)
			} else {
				assert.NoError(t, err, "%s of %q", method, addr)
			}
		}
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
