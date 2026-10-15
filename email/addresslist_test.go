package email

import (
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/mail"
	"reflect"
	"testing"

	"github.com/domonda/go-types/internal/xmlattrtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStandardComparison(t *testing.T) {
	validEmailAddresses := map[string]*mail.Address{
		`"Unger, Erik" <u.erik@domonda.com>`:        {Name: "Unger, Erik", Address: "u.erik@domonda.com"},
		`"Unger, Erik" <"Unger, Erik"@domonda.com>`: {Name: "Unger, Erik", Address: "Unger, Erik@domonda.com"},
	}

	for addr, expected := range validEmailAddresses {
		t.Run(addr, func(t *testing.T) {
			result, err := mail.ParseAddress(addr)
			assert.NoError(t, err, "valid email address")
			assert.Equal(t, expected, result, "expected: %s", expected)
		})
	}

	for addr, expected := range validEmailAddresses {
		t.Run(addr, func(t *testing.T) {
			results, err := mail.ParseAddressList(addr)
			assert.NoError(t, err, "valid email address")
			assert.Len(t, results, 1, "list of one address")
			assert.Equal(t, expected, results[0], "expected: %s", expected)
		})
	}
}

func TestAddressList_Split(t *testing.T) {
	tests := []struct {
		l       AddressList
		want    []Address
		wantErr bool
	}{
		{l: ``, want: nil},
		{l: `<hello@example.com>,`, want: []Address{`hello@example.com`}},
		{l: `<Hello@example.com>, World@example.com`, want: []Address{`hello@example.com`, `world@example.com`}},
		// Partial result: parsable addresses are returned alongside the error.
		{l: `alice@example.com, broken address, bob@example.com`, want: []Address{`alice@example.com`, `bob@example.com`}, wantErr: true},
		{l: `@broken`, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(string(tt.l), func(t *testing.T) {
			got, err := tt.l.SplitNormalized()
			if (err != nil) != tt.wantErr {
				t.Errorf("AddressList(%#v).SplitNormalized() error = %v, wantErr %v", string(tt.l), err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddressList(%#v).SplitNormalized() = %v, want %v", string(tt.l), got, tt.want)
			}
		})
	}
}

func TestAddressList_Scan(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    AddressList
		wantErr bool
	}{
		// The array branch reads a PostgreSQL text[] column. Its elements
		// are quoted and escaped, so they have to be decoded rather than
		// joined verbatim, otherwise the quotes end up inside an address.
		{
			name:  "quoted array elements",
			value: `{"a@example.com","b@example.com"}`,
			want:  "a@example.com, b@example.com",
		},
		{
			name:  "quoted element with a comma and escaped quotes",
			value: `{"\"Doe, John\" <john@example.com>","b@example.com"}`,
			want:  `"Doe, John" <john@example.com>, b@example.com`,
		},
		{
			name:  "unquoted array elements",
			value: `{a@example.com,b@example.com}`,
			want:  "a@example.com, b@example.com",
		},
		{name: "empty array", value: `{}`, want: ""},
		// A hand written literal, unlike PostgreSQL output, may pad its
		// elements. The padding is not part of the address.
		{name: "padded elements", value: `{ a@example.com , b@example.com }`, want: "a@example.com, b@example.com"},
		{name: "whitespace only array", value: `{ }`, want: ""},
		// Padding around the whole literal is trimmed like around a plain
		// string, so the array must still be decoded instead of being
		// stored as the literal text with braces and quotes
		{name: "padded array literal", value: " {\"a@example.com\",b@example.com} ", want: "a@example.com, b@example.com"},
		// Regression cases for the array codec swap (nullable.SplitArray ->
		// notnull.StringArray). Each behaves differently than before, so each
		// is pinned rather than left to be discovered in production.
		// A SQL NULL element used to become the address literally spelled NULL.
		{name: "NULL element", value: `{NULL}`, wantErr: true},
		// Quoted, it is the ordinary string "NULL", not a SQL NULL.
		{name: "quoted NULL element", value: `{"NULL"}`, want: "NULL"},
		// A trailing comma used to be accepted and yield one element.
		{name: "trailing comma", value: `{a@example.com,}`, wantErr: true},
		// Backslash escapes are unescaped now; before, both backslashes stayed.
		{name: "escaped backslash", value: `{"a\\b@example.com"}`, want: `a\b@example.com`},
		// A nested literal used to yield the inner braces as a single address.
		{name: "nested array", value: `{{a@example.com}}`, wantErr: true},
		// A non-array string is an already joined list, so it is taken as is.
		{name: "plain string", value: "a@example.com, b@example.com", want: "a@example.com, b@example.com"},
		// Like the padded array elements above and like AddressSet.Scan,
		// the padding of a plain string is not part of the list, else
		// " a@example.com " and "a@example.com" would be different lists
		// and the padding would end up in the first and last address.
		{name: "padded plain string", value: " a@example.com, b@example.com\t\n", want: "a@example.com, b@example.com"},
		{name: "padded plain bytes", value: []byte(" a@example.com "), want: "a@example.com"},
		{name: "bytes", value: []byte(`{"a@example.com"}`), want: "a@example.com"},
		// Like every non-nullable type, AddressList reads SQL NULL and a
		// blank string as its empty value instead of failing the query;
		// Value writes the empty list back as SQL NULL.
		{name: "whitespace only plain string", value: " \t ", want: ""},
		{name: "null", value: nil, want: ""},
		{name: "empty string", value: "", want: ""},
		// A plain list is kept as stored, also with other spellings or
		// invalid entries, because AddressList is a collection
		{name: "plain string not normalized", value: "A@Example.com, John <J@Example.com>", want: "A@Example.com, John <J@Example.com>"},
		{name: "invalid plain string", value: "not an address", want: "not an address"},
		// json and jsonb columns: arrays and objects are joined,
		// a JSON string is read as address list
		{name: "JSON array", value: `["b@example.com", "A@Example.com"]`, want: "b@example.com, A@Example.com"},
		{name: "JSON object", value: `{"b@example.com": {}, "a@example.com": {}}`, want: "a@example.com, b@example.com"},
		{name: "JSON string", value: `"a@example.com, b@example.com"`, want: "a@example.com, b@example.com"},
		// A quoted display name is no JSON string, but an address list
		{name: "quoted display name", value: `"Doe, John" <john@example.com>`, want: `"Doe, John" <john@example.com>`},
		{name: "JSON array of numbers", value: `[1, 2]`, wantErr: true},
		{name: "invalid array", value: `{,}`, wantErr: true},
		{name: "unsupported type", value: 123, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var l AddressList
			err := l.Scan(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Scan() error = %v, wantErr %t", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if l != tt.want {
				t.Errorf("Scan() = %q, want %q", l, tt.want)
			}
		})
	}
}

// TestNormalizeAddressList protects deduplication by mailbox: the same
// address with a different display name or case must only appear once
// (keeping the first spelling), otherwise a recipient would get a
// message multiple times.
func TestNormalizeAddressList(t *testing.T) {
	got, err := NormalizeAddressList(`John <John@Example.com>, jane@example.com, "J. Doe" <john@example.com>, JANE@example.com`)
	require.NoError(t, err)
	assert.Equal(t, []string{`"John" <john@example.com>`, `jane@example.com`}, got)

	// A strict normalization must not silently drop invalid entries.
	_, err = NormalizeAddressList(`a@example.com, not an address`)
	assert.Error(t, err)
}

// TestAddressListJoin protects that empty addresses never produce
// empty list entries like "a@example.com, , b@example.com" that
// would make the joined list invalid.
func TestAddressListJoin(t *testing.T) {
	tests := []struct {
		name  string
		addrs []Address
		want  AddressList
	}{
		{name: "none", addrs: nil, want: ""},
		{name: "only empty", addrs: []Address{"", ""}, want: ""},
		{name: "one", addrs: []Address{"a@example.com"}, want: "a@example.com"},
		{name: "empty skipped", addrs: []Address{"", "a@example.com", "", `"Doe, John" <john@example.com>`, ""}, want: `a@example.com, "Doe, John" <john@example.com>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, AddressListJoin(tt.addrs...), "AddressListJoin")

			strs := make([]string, len(tt.addrs))
			for i, a := range tt.addrs {
				strs[i] = string(a)
			}
			assert.Equal(t, tt.want, AddressListJoinStrings(strs...), "AddressListJoinStrings must behave like AddressListJoin")
		})
	}
}

func TestAddressList_Append(t *testing.T) {
	tests := []struct {
		name  string
		list  AddressList
		addrs []Address
		want  AddressList
	}{
		// Appending to an empty list must not start with a separator.
		{name: "to empty", list: "", addrs: []Address{"a@example.com"}, want: "a@example.com"},
		{name: "to non empty", list: "a@example.com", addrs: []Address{"b@example.com", "c@example.com"}, want: "a@example.com, b@example.com, c@example.com"},
		{name: "empty skipped", list: "a@example.com", addrs: []Address{"", "b@example.com", ""}, want: "a@example.com, b@example.com"},
		{name: "nothing", list: "a@example.com", addrs: nil, want: "a@example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.list.Append(tt.addrs...)
			assert.Equal(t, tt.want, got)
			assert.True(t, got.Valid(), "the appended list must stay parsable")
		})
	}
}

// TestAddressList_UniqueAddressParts protects that the set contains bare,
// lower case mailboxes so it can be used for recipient lookups regardless
// of display names and case.
func TestAddressList_UniqueAddressParts(t *testing.T) {
	set, err := AddressList(`John <John@Example.com>, jane@example.com, "J. Doe" <john@example.com>`).UniqueAddressParts()
	require.NoError(t, err)
	assert.Equal(t, MakeAddressSet("john@example.com", "jane@example.com"), set)

	// The set holds formatted address parts, so a quoted local part
	// stays quoted: unquoted, "victim@evil.com>"@example.com would
	// become a lookup key that reads as a mailbox at evil.com.
	set, err = AddressList(`"Victim@Evil.com>"@Example.com, X <"A B"@Example.com>, "a b"@example.com`).UniqueAddressParts()
	require.NoError(t, err)
	assert.Equal(t, MakeAddressSet(`"victim@evil.com>"@example.com`, `"a b"@example.com`), set)

	set, err = AddressList(`a@example.com, not an address`).UniqueAddressParts()
	assert.Error(t, err, "an invalid entry must not be silently dropped")
	assert.Nil(t, set)
}

func TestAddressList_ValidateAndNormalized(t *testing.T) {
	tests := []struct {
		list               AddressList
		valid              bool
		normalized         AddressList
		validAndNormalized bool
	}{
		{
			list:               `a@example.com, b@example.com`,
			valid:              true,
			normalized:         `a@example.com, b@example.com`,
			validAndNormalized: true,
		},
		{
			list:       `A@Example.com,b@example.com`,
			valid:      true,
			normalized: `a@example.com, b@example.com`,
		},
		{
			list:       `John <John@Example.com>, <jane@example.com>`,
			valid:      true,
			normalized: `"John" <john@example.com>, jane@example.com`,
		},
		// The quoted comma in the name must survive normalization,
		// otherwise the normalized list would split into a different
		// number of addresses.
		{
			list:               `"Doe, John" <john@example.com>, b@example.com`,
			valid:              true,
			normalized:         `"Doe, John" <john@example.com>, b@example.com`,
			validAndNormalized: true,
		},
		{list: `a@example.com, not an address`, valid: false},
		{list: `not an address`, valid: false},
	}
	for _, tt := range tests {
		t.Run(string(tt.list), func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.list.Valid(), "Valid")
			assert.Equal(t, tt.valid, tt.list.Validate() == nil, "Validate must agree with Valid")
			assert.Equal(t, tt.validAndNormalized, tt.list.ValidAndNormalized(), "ValidAndNormalized")

			normalized, err := tt.list.Normalized()
			if !tt.valid {
				assert.Error(t, err)
				// On error the input is returned unchanged so it isn't lost.
				assert.Equal(t, tt.list, normalized)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.normalized, normalized)
			assert.True(t, normalized.ValidAndNormalized(), "normalizing must be idempotent")
		})
	}
}

// AddressList is a collection: Value and MarshalJSON write the list as
// stored, because normalizing would rewrite every entry of it. Value trims
// the surrounding whitespace, writes a blank list as SQL NULL and refuses
// an invalid list, because the database is where it would stay.
// MarshalJSON writes every list unchanged, also an invalid one, so that
// it round-trips through UnmarshalJSON.
func TestAddressList_ValueAndMarshalJSONAsStored(t *testing.T) {
	tests := []struct {
		l       AddressList
		wantSQL driver.Value
	}{
		{l: `a@example.com`, wantSQL: `a@example.com`},
		{l: `A@Example.com,  Jürgen <J@Example.com>`, wantSQL: `A@Example.com,  Jürgen <J@Example.com>`},
		{l: ` a@example.com `, wantSQL: `a@example.com`},
		// A valid list without addresses needs no special case
		{l: `undisclosed-recipients:;`, wantSQL: `undisclosed-recipients:;`},
		{l: `Undisclosed recipients`, wantSQL: `Undisclosed recipients`},
		// An address that starts like a placeholder is an address
		{l: `withheld recipients <victim@example.com>`, wantSQL: `withheld recipients <victim@example.com>`},
		{l: ``, wantSQL: nil},
		{l: `  `, wantSQL: nil},
	}
	for _, tt := range tests {
		v, err := tt.l.Value()
		require.NoError(t, err)
		assert.Equal(t, tt.wantSQL, v, "AddressList(%q).Value()", string(tt.l))
		if v != nil {
			var scanned AddressList
			require.NoError(t, scanned.Scan(v))
			assert.Equal(t, AddressList(v.(string)), scanned, "Scan(Value()) of %q", string(tt.l))
		}
	}

	invalid := AddressList(`a@example.com, not an address`)
	_, err := invalid.Value()
	assert.Error(t, err, "Value() of an invalid list")

	for _, l := range []AddressList{`A@Example.com,  Jürgen <J@Example.com>`, ` a@example.com `, `  `, ``, invalid} {
		j, err := json.Marshal(l)
		require.NoError(t, err)
		var str string
		require.NoError(t, json.Unmarshal(j, &str))
		assert.Equal(t, string(l), str, "json.Marshal(AddressList(%q))", string(l))
		var back AddressList
		require.NoError(t, json.Unmarshal(j, &back))
		assert.Equal(t, l, back, "JSON round trip of %q", string(l))
	}

	// Text after a placeholder is no list without addresses: other
	// parsers read the address in it, so it must not pass Validate
	// as empty list
	for _, hidden := range []AddressList{
		"undisclosed-recipients:;, victim@example.com",
		"Undisclosed-Recipients: victim@example.com;",
	} {
		assert.Error(t, hidden.Validate(), hidden)
		_, err := hidden.Value()
		assert.Error(t, err, "Value() of %q", hidden)
	}
}

// ASCII writes a list normalized with non-ASCII names RFC 2047 encoded
// for message headers, or the list unchanged if it can't be normalized,
// while fmt prints a list as stored.
func TestAddressList_ASCII(t *testing.T) {
	l := AddressList(`A@Example.com, Jürgen <J@Example.com>`)
	assert.Equal(t, `a@example.com, =?utf-8?q?J=C3=BCrgen?= <j@example.com>`, l.ASCII())
	assert.Equal(t, string(l), fmt.Sprint(l))

	invalid := AddressList(`a@example.com, not an address`)
	assert.Equal(t, string(invalid), invalid.ASCII())
	assert.Equal(t, string(invalid), fmt.Sprint(invalid))
}

func TestAddressList_Nullable(t *testing.T) {
	const list AddressList = `"Doe, John" <john@example.com>, b@example.com`
	assert.Equal(t, NullableAddressList(list), list.Nullable(), "conversion must keep the exact string")
}

// AddressList is written and read as stored in XML, as element and as
// attribute, like in every other format.
func TestAddressList_XML(t *testing.T) {
	type doc struct {
		XMLName xml.Name            `xml:"doc"`
		To      AddressList         `xml:"to"`
		Cc      NullableAddressList `xml:"cc,attr"`
	}
	in := doc{To: `A@Example.com, "Doe, John" <john@example.com>`, Cc: `not an address`}
	data, err := xml.Marshal(in)
	require.NoError(t, err)
	var got doc
	require.NoError(t, xml.Unmarshal(data, &got))
	assert.Equal(t, in.To, got.To)
	assert.Equal(t, in.Cc, got.Cc)
}

func TestAddressList_MarshalXMLAttr(t *testing.T) {
	xmlattrtest.NilPointer(t, AddressList(""), AddressList(`A@Example.com, "Doe, John" <john@example.com>`))
}
