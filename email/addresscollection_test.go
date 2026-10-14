package email

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got, err := readAddressCollection(tt.text, "AddressSet")
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

type xmlPtrAttrDoc[T any] struct {
	XMLName xml.Name `xml:"doc"`
	Attr    *T       `xml:"attr,attr"`
}

type xmlValueAttrDoc[T any] struct {
	XMLName xml.Name `xml:"doc"`
	Attr    T        `xml:"attr,attr"`
}

// testMarshalXMLAttr checks that a nil *T attribute is left out without
// omitempty, because encoding/xml would call the value receiver
// MarshalText through the nil pointer and panic, and that a *T attribute
// is written exactly like a T attribute, with "" for an empty value.
func testMarshalXMLAttr[T any](t *testing.T, values ...T) {
	t.Helper()
	var data []byte
	var err error
	require.NotPanics(t, func() { data, err = xml.Marshal(xmlPtrAttrDoc[T]{}) })
	require.NoError(t, err)
	assert.Equal(t, `<doc></doc>`, string(data))

	for _, v := range values {
		want, err := xml.Marshal(xmlValueAttrDoc[T]{Attr: v})
		require.NoError(t, err)
		got, err := xml.Marshal(xmlPtrAttrDoc[T]{Attr: &v})
		require.NoError(t, err)
		assert.Equal(t, string(want), string(got))
	}
}
